package github

import (
	"context"
	"encoding/base64"
	"fmt"
	nethttp "net/http"
	"net/url"
	"strings"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	wh "github.com/go-playground/webhooks/github"
	"github.com/google/go-github/v68/github"
	"github.com/padok-team/burrito/internal/repository/credentials"
	"github.com/padok-team/burrito/internal/repository/providers/standard"

	"github.com/padok-team/burrito/internal/repository/types"
	utils "github.com/padok-team/burrito/internal/utils/url"
	"github.com/padok-team/burrito/internal/webhook/event"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type Github struct {
	Config credentials.Credential
}

func (g *Github) GetWebhookProvider() (types.WebhookProvider, error) {
	webhook, err := wh.New(wh.Options.Secret(g.Config.WebhookSecret))
	if err != nil {
		return nil, err
	}
	return &WebhookProvider{
		Webhook: webhook,
	}, nil
}

func (g *Github) GetAPIProvider() (types.APIProvider, error) {
	githubClient, err := buildGithubClient(g.Config, detectClientType(g.Config))
	if err != nil {
		return nil, err
	}
	return &APIProvider{
		client: githubClient,
	}, nil
}

func (g *Github) GetGitProvider() (types.GitProvider, error) {
	githubClient, err := buildGithubClient(g.Config, detectClientType(g.Config))
	if err != nil {
		return nil, err
	}
	auth, err := buildGitCredentials(g.Config, detectClientType(g.Config), githubClient)
	if err != nil {
		return nil, err
	}
	return &standard.GitProvider{
		URL:        g.Config.URL,
		AuthMethod: auth,
	}, nil
}

type GitHubSubscription string

const (
	GitHubEnterprise GitHubSubscription = "enterprise"
	GitHubClassic    GitHubSubscription = "classic"
)

func detectClientType(config credentials.Credential) string {
	clientType := ""
	// GitHub App authentication first
	if config.AppID != 0 && config.AppInstallationID != 0 && config.AppPrivateKey != "" {
		clientType = "app"
	} else if config.GitHubToken != "" {
		// Try GitHub Token authentication
		clientType = "token"
	} else if config.Username != "" && config.Password != "" {
		// Try basic authentication
		clientType = "basic"
	} else {
		clientType = "none"
	}
	return clientType
}

func buildGithubClient(config credentials.Credential, clientType string) (*github.Client, error) {
	apiUrl, subscription, err := inferBaseURL(utils.NormalizeUrl(config.URL))
	if err != nil {
		return nil, err
	}
	var client *github.Client
	var httpClient *nethttp.Client
	switch clientType {
	case "app":
		itr, err := ghinstallation.New(
			nethttp.DefaultTransport,
			config.AppID,
			config.AppInstallationID,
			[]byte(config.AppPrivateKey),
		)
		if err != nil {
			return nil, fmt.Errorf("error creating GitHub App client: %w", err)
		}
		httpClient = &nethttp.Client{Transport: itr}
	case "token":
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.GitHubToken})
		httpClient = oauth2.NewClient(context.Background(), ts)
	case "basic":
		ts := oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: base64.StdEncoding.EncodeToString(
				[]byte(fmt.Sprintf("%s:%s", config.Username, config.Password)),
			)})
		httpClient = oauth2.NewClient(context.Background(), ts)
	default:
		return nil, fmt.Errorf("unsupported GitHub client type: %s", clientType)
	}
	if subscription == GitHubEnterprise {
		client, err = github.NewClient(httpClient).WithEnterpriseURLs(apiUrl, apiUrl)
		if err != nil {
			return nil, fmt.Errorf("error creating GitHub Enterprise client: %w", err)
		}
		return client, nil
	}
	client = github.NewClient(httpClient)
	return client, nil
}

func buildGitCredentials(config credentials.Credential, clientType string, client *github.Client) (transport.AuthMethod, error) {
	switch clientType {
	case "app":
		itr := client.Client().Transport.(*ghinstallation.Transport)
		token, err := itr.Token(context.Background())
		if err != nil {
			return nil, fmt.Errorf("error getting GitHub App token: %w", err)
		}
		return &http.BasicAuth{
			Username: "x-access-token",
			Password: token,
		}, nil
	case "token":
		return &http.BasicAuth{
			Username: "x-access-token",
			Password: config.GitHubToken,
		}, nil
	case "basic":
		return &http.BasicAuth{
			Username: config.Username,
			Password: config.Password,
		}, nil
	default:
		log.Info("No authentication method provided, falling back to unauthenticated clone")
		return nil, nil
	}
}

func getNormalizedAction(action string) string {
	switch action {
	case "opened", "reopened":
		return event.PullRequestOpened
	case "closed":
		return event.PullRequestClosed
	default:
		return action
	}
}

func parseGithubUrl(url string) (string, string) {
	normalizedUrl := utils.NormalizeUrl(url)
	// nomalized url are "https://padok.github.com/owner/repo"
	// we remove "https://" then split on "/"
	split := strings.Split(normalizedUrl[8:], "/")
	return split[1], split[2]
}

func inferBaseURL(repoURL string) (string, GitHubSubscription, error) {
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid repository URL: %w", err)
	}

	host := parsedURL.Host
	host = strings.TrimPrefix(host, "www.")

	if host != "github.com" {
		return fmt.Sprintf("https://%s/api/v3", host), GitHubEnterprise, nil
	} else {
		return "", GitHubClassic, nil
	}
}
