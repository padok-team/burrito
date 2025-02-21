package gitlab

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	wh "github.com/go-playground/webhooks/gitlab"
	"github.com/padok-team/burrito/internal/repository/credentials"
	"github.com/padok-team/burrito/internal/repository/providers/standard"
	"github.com/padok-team/burrito/internal/repository/types"

	utils "github.com/padok-team/burrito/internal/utils/url"
	"github.com/padok-team/burrito/internal/webhook/event"
	log "github.com/sirupsen/logrus"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type Gitlab struct {
	Config credentials.Credential
}

func (g *Gitlab) GetWebhookProvider() (types.WebhookProvider, error) {
	webhook, err := wh.New(wh.Options.Secret(g.Config.WebhookSecret))
	if err != nil {
		return nil, err
	}
	return &WebhookProvider{
		Webhook: webhook,
	}, nil
}

func (g *Gitlab) GetAPIProvider() (types.APIProvider, error) {
	client, err := buildGitlabClient(g.Config)
	if err != nil {
		return nil, err
	}
	return &APIProvider{
		client: client,
	}, nil
}

func (g *Gitlab) GetGitProvider() (types.GitProvider, error) {
	auth, err := buildGitCredentials(g.Config)
	if err != nil {
		return nil, err
	}
	return &standard.GitProvider{
		URL:        g.Config.URL,
		AuthMethod: auth,
	}, nil
}

func buildGitlabClient(config credentials.Credential) (*gitlab.Client, error) {
	apiUrl, err := inferBaseURL(utils.NormalizeUrl(config.URL))
	if err != nil {
		return nil, err
	}

	var token string
	if config.GitLabToken != "" {
		token = config.GitLabToken
	} else if config.Username != "" && config.Password != "" {
		token = config.Password
	} else {
		log.Info("No authentication method provided, falling back to unauthenticated clone")
	}

	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(apiUrl))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	return client, nil
}

func buildGitCredentials(config credentials.Credential) (transport.AuthMethod, error) {
	if config.GitLabToken != "" {
		return &http.BasicAuth{
			Username: "oauth2",
			Password: config.GitLabToken,
		}, nil
	} else if config.Username != "" && config.Password != "" {
		return &http.BasicAuth{
			Username: config.Username,
			Password: config.Password,
		}, nil
	}
	log.Info("No authentication method provided, falling back to unauthenticated clone")
	return nil, nil
}

func getNormalizedAction(action string) string {
	switch action {
	case "open", "reopen":
		return event.PullRequestOpened
	case "close", "merge":
		return event.PullRequestClosed
	default:
		return action
	}
}

func getGitlabNamespacedName(url string) string {
	normalizedUrl := utils.NormalizeUrl(url)
	return strings.Join(strings.Split(normalizedUrl[8:], "/")[1:], "/")
}

func inferBaseURL(repoURL string) (string, error) {
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("invalid repository URL: %w", err)
	}

	host := parsedURL.Host
	host = strings.TrimPrefix(host, "www.")
	return fmt.Sprintf("https://%s/api/v4", host), nil
}
