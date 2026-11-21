package gitea

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	wh "github.com/go-playground/webhooks/v6/gitea"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/repository/credentials"
	"github.com/padok-team/burrito/internal/repository/providers/standard"
	"github.com/padok-team/burrito/internal/repository/types"

	gitea "gitea.dev/sdk"
	utils "github.com/padok-team/burrito/internal/utils/url"
	"github.com/padok-team/burrito/internal/webhook/event"
	log "github.com/sirupsen/logrus"
)

type Gitea struct {
	Config credentials.Credential
}

func (g *Gitea) GetWebhookProvider() (types.WebhookProvider, error) {
	webhook, err := wh.New(wh.Options.Secret(g.Config.WebhookSecret))
	if err != nil {
		return nil, err
	}
	return &WebhookProvider{
		Webhook: webhook,
	}, nil
}

func (g *Gitea) GetAPIProvider() (types.APIProvider, error) {
	client, err := buildGiteaClient(g.Config)
	if err != nil {
		return nil, err
	}
	return &APIProvider{
		client: client,
	}, nil
}

func (g *Gitea) GetGitProvider(repository *configv1alpha1.TerraformRepository) (types.GitProvider, error) {
	auth, err := buildGitCredentials(g.Config)
	if err != nil {
		return nil, err
	}
	return &standard.GitProvider{
		RepoURL:    repository.Spec.Repository.Url,
		AuthMethod: auth,
	}, nil
}

// buildGiteaClient builds a Gitea API client. Forgejo exposes a
// Gitea-compatible API, so the same client works for both.
func buildGiteaClient(config credentials.Credential) (*gitea.Client, error) {
	apiURL, err := inferBaseURL(utils.NormalizeUrl(config.URL))
	if err != nil {
		return nil, err
	}

	var options []gitea.ClientOption
	if config.GiteaToken != "" {
		options = append(options, gitea.SetToken(config.GiteaToken))
	} else if config.Username != "" && config.Password != "" {
		options = append(options, gitea.SetBasicAuth(config.Username, config.Password))
	} else {
		log.Info("No authentication method provided, falling back to unauthenticated clone")
	}

	return gitea.NewClient(apiURL, options...)
}

func buildGitCredentials(config credentials.Credential) (transport.AuthMethod, error) {
	if config.GiteaToken != "" {
		return &http.BasicAuth{
			Username: "git",
			Password: config.GiteaToken,
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
	case "open", "reopen", "opened", "reopened":
		return event.PullRequestOpened
	case "close", "closed":
		return event.PullRequestClosed
	default:
		return action
	}
}

// parseGiteaUrl extracts owner and repo from a normalized repository URL.
// Gitea URLs are always "https://host/owner/repo" — no nested groups
// (unlike GitLab), so this is a simple split.
func parseGiteaUrl(url string) (string, string) {
	normalizedUrl := utils.NormalizeUrl(url)
	// normalized url is "https://gitea.com/owner/repo"
	split := strings.Split(normalizedUrl[8:], "/")
	return split[1], split[2]
}

func inferBaseURL(repoURL string) (string, error) {
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("invalid repository URL: %w", err)
	}

	host := parsedURL.Host
	host = strings.TrimPrefix(host, "www.")
	return fmt.Sprintf("https://%s", host), nil
}
