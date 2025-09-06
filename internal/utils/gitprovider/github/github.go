package github

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	nethttp "net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	wh "github.com/go-playground/webhooks/github"
	"github.com/google/go-github/v74/github"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	"github.com/padok-team/burrito/internal/utils/gitprovider/common"
	"github.com/padok-team/burrito/internal/utils/gitprovider/types"
	utils "github.com/padok-team/burrito/internal/utils/url"
	"github.com/padok-team/burrito/internal/webhook/event"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type Github struct {
	*github.Client
	Config           types.Config
	HttpClient       *nethttp.Client
	WebhookHandler   *wh.Webhook
	GitHubClientType string
	itr              *ghinstallation.Transport
}

type GitHubSubscription string

const (
	GitHubEnterprise GitHubSubscription = "enterprise"
	GitHubClassic    GitHubSubscription = "classic"
)

func IsAvailable(config types.Config, capabilities []string) bool {
	var allCapabilities = []string{types.Capabilities.Clone, types.Capabilities.Comment, types.Capabilities.Changes, types.Capabilities.Webhook}
	// Check that the configuration is valid
	// For webhook handling, the GitHub token is not required
	webhookOnlyRequested := len(capabilities) == 1 && capabilities[0] == types.Capabilities.Webhook
	hasGitHubToken := config.GitHubToken != ""
	hasAppCredentials := config.AppID != 0 && config.AppInstallationID != 0 && config.AppPrivateKey != ""
	hasWebhookSecret := config.WebhookSecret != ""

	if !hasGitHubToken && !hasAppCredentials && !(webhookOnlyRequested && hasWebhookSecret) {
		return false
	}
	// Check that the requested capabilities are supported
	for _, c := range capabilities {
		if !slices.Contains(allCapabilities, c) {
			return false
		}
	}
	return true
}

func (g *Github) Init() error {
	apiUrl, subscription, err := inferBaseURL(utils.NormalizeUrl(g.Config.URL))
	if err != nil {
		return err
	}

	var httpClient *nethttp.Client

	// GitHub App authentication first
	if g.Config.AppID != 0 && g.Config.AppInstallationID != 0 && g.Config.AppPrivateKey != "" {
		itr, err := ghinstallation.New(
			nethttp.DefaultTransport,
			g.Config.AppID,
			g.Config.AppInstallationID,
			[]byte(g.Config.AppPrivateKey),
		)
		if err != nil {
			return fmt.Errorf("error creating GitHub App client: %w", err)
		}
		if subscription == GitHubEnterprise {
			itr.BaseURL = apiUrl
		}
		g.GitHubClientType = "app"
		httpClient = &nethttp.Client{Transport: itr}
		g.itr = itr
	} else if g.Config.GitHubToken != "" {
		// Try GitHub Token authentication
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: g.Config.GitHubToken})
		g.GitHubClientType = "token"
		httpClient = oauth2.NewClient(context.Background(), ts)
	} else if g.Config.Username != "" && g.Config.Password != "" {
		// Try basic authentication
		ts := oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: base64.StdEncoding.EncodeToString(
				[]byte(fmt.Sprintf("%s:%s", g.Config.Username, g.Config.Password)),
			),
		})
		g.GitHubClientType = "basic"
		httpClient = oauth2.NewClient(context.Background(), ts)
	} else {
		return errors.New("no valid authentication method provided")
	}

	// Create the appropriate client based on GitHub type
	if subscription == GitHubEnterprise {
		g.Client, err = github.NewClient(httpClient).WithEnterpriseURLs(apiUrl, apiUrl)
		if err != nil {
			return fmt.Errorf("error creating GitHub Enterprise client: %w", err)
		}
	} else {
		g.Client = github.NewClient(httpClient)
	}
	return nil
}

func (g *Github) InitWebhookHandler() error {
	handler, err := wh.New(wh.Options.Secret(g.Config.WebhookSecret))
	if err != nil {
		return err
	}
	g.WebhookHandler = handler
	return nil
}

func (g *Github) GetChanges(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) ([]string, error) {
	owner, repoName := parseGithubUrl(repository.Spec.Repository.Url)
	id, err := strconv.Atoi(pr.Spec.ID)
	if err != nil {
		log.Errorf("Error while parsing Github pull request ID: %s", err)
		return []string{}, err
	}
	// Per page is 30 by default, max is 100
	opts := &github.ListOptions{
		PerPage: 100,
	}
	// Get all pull request files from Github
	var allChangedFiles []string
	for {
		changedFiles, resp, err := g.Client.PullRequests.ListFiles(context.TODO(), owner, repoName, id, opts)
		if err != nil {
			return []string{}, err
		}
		for _, file := range changedFiles {
			if *file.Status != "unchanged" {
				allChangedFiles = append(allChangedFiles, *file.Filename)
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allChangedFiles, nil
}

func (g *Github) Comment(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, comment comment.Comment) error {
	body, err := comment.Generate(pr.Annotations[annotations.LastBranchCommit])
	if err != nil {
		log.Errorf("Error while generating comment: %s", err)
		return err
	}
	owner, repoName := parseGithubUrl(repository.Spec.Repository.Url)
	id, err := strconv.Atoi(pr.Spec.ID)
	if err != nil {
		log.Errorf("Error while parsing Github pull request ID: %s", err)
		return err
	}
	_, _, err = g.Client.Issues.CreateComment(context.TODO(), owner, repoName, id, &github.IssueComment{
		Body: &body,
	})
	return err
}

func (g *Github) Clone(repository *configv1alpha1.TerraformRepository, branch string, repositoryPath string) (*git.Repository, error) {
	auth, err := g.GetGitAuth()
	if err != nil {
		return nil, err
	}

	cloneOptions := &git.CloneOptions{
		ReferenceName: common.ReferenceName(branch),
		URL:           repository.Spec.Repository.Url,
		Auth:          auth,
	}

	log.Infof("Cloning github repository %s on ref %s with github %s authentication", repository.Spec.Repository.Url, branch, g.GitHubClientType)
	repo, err := git.PlainClone(repositoryPath, false, cloneOptions)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (g *Github) ParseWebhookPayload(r *nethttp.Request) (interface{}, bool) {
	// if the request is not a GitHub event, return false
	if r.Header.Get("X-GitHub-Event") == "" {
		return nil, false
	} else {
		// check if the request can be verified with the secret of this provider
		p, err := g.WebhookHandler.Parse(r, wh.PushEvent, wh.PingEvent, wh.PullRequestEvent)
		if errors.Is(err, wh.ErrHMACVerificationFailed) {
			return nil, false
		} else if err != nil {
			log.Errorf("an error occurred during request parsing : %s", err)
			return nil, false
		}
		return p, true
	}
}

func (g *Github) GetEventFromWebhookPayload(p interface{}) (event.Event, error) {
	var e event.Event
	var err error
	switch payload := p.(type) {
	case wh.PushPayload:
		log.Infof("parsing Github push event payload")
		changedFiles := []string{}
		for _, commit := range payload.Commits {
			changedFiles = append(changedFiles, commit.Added...)
			changedFiles = append(changedFiles, commit.Modified...)
			changedFiles = append(changedFiles, commit.Removed...)
		}
		e = &event.PushEvent{
			URL:       utils.NormalizeUrl(payload.Repository.HTMLURL),
			Reference: event.ParseReference(payload.Ref),
			ChangeInfo: event.ChangeInfo{
				ShaBefore: payload.Before,
				ShaAfter:  payload.After,
			},
			Changes: changedFiles,
		}
	case wh.PullRequestPayload:
		log.Infof("parsing Github pull request event payload")
		if err != nil {
			log.Warnf("could not retrieve pull request from Github API: %s", err)
			return nil, err
		}
		e = &event.PullRequestEvent{
			ID:        strconv.FormatInt(payload.PullRequest.Number, 10),
			URL:       utils.NormalizeUrl(payload.Repository.HTMLURL),
			Reference: payload.PullRequest.Head.Ref,
			Action:    getNormalizedAction(payload.Action),
			Base:      payload.PullRequest.Base.Ref,
			Commit:    payload.PullRequest.Head.Sha,
		}
	default:
		return nil, errors.New("unsupported Event")
	}
	return e, nil
}

func (g *Github) GetLatestRevisionForRef(repository *configv1alpha1.TerraformRepository, ref string) (string, error) {
	owner, repoName := parseGithubUrl(repository.Spec.Repository.Url)
	b, _, err := g.Client.Repositories.GetBranch(context.TODO(), owner, repoName, ref, 10)
	if err == nil {
		return b.Commit.GetSHA(), nil
	}
	t, _, err := g.Client.Git.GetRef(context.TODO(), owner, repoName, fmt.Sprintf("refs/tags/%s", ref))
	if err == nil {
		return t.Object.GetSHA(), nil
	}
	return "", fmt.Errorf("could not find revision for ref %s: %w", ref, err)
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

// GetGitAuth returns the appropriate authentication method based on the GitHub client type
func (g *Github) GetGitAuth() (transport.AuthMethod, error) {
	switch g.GitHubClientType {
	case "app":
		token, err := g.itr.Token(context.Background())
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
			Password: g.Config.GitHubToken,
		}, nil
	case "basic":
		return &http.BasicAuth{
			Username: g.Config.Username,
			Password: g.Config.Password,
		}, nil
	default:
		log.Info("No authentication method provided, falling back to unauthenticated clone")
		return nil, nil
	}
}
