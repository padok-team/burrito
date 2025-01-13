package gitlab

import (
	"errors"
	"fmt"
	nethttp "net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	wh "github.com/go-playground/webhooks/gitlab"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	"github.com/padok-team/burrito/internal/utils/gitprovider/types"
	utils "github.com/padok-team/burrito/internal/utils/url"
	"github.com/padok-team/burrito/internal/webhook/event"
	log "github.com/sirupsen/logrus"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type Gitlab struct {
	*gitlab.Client
	WebhookHandler *wh.Webhook
	Config         types.Config
}

func IsAvailable(config types.Config, capabilities []string) bool {
	var allCapabilities = []string{types.Capabilities.Clone, types.Capabilities.Comment, types.Capabilities.Changes, types.Capabilities.Webhook}
	// Check that the configuration is valid
	// For webhook handling, the GitLab token is not required
	webhookOnlyRequested := len(capabilities) == 1 && capabilities[0] == types.Capabilities.Webhook
	hasGitLabToken := config.GitLabToken != ""
	hasWebhookSecret := config.WebhookSecret != ""

	if !(hasGitLabToken || (webhookOnlyRequested && hasWebhookSecret)) {
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

func (g *Gitlab) Init() error {
	apiUrl, err := inferBaseURL(utils.NormalizeUrl(g.Config.URL))
	if err != nil {
		return err
	}

	var token string
	if g.Config.GitLabToken != "" {
		token = g.Config.GitLabToken
	} else if g.Config.Username != "" && g.Config.Password != "" {
		token = g.Config.Password
	} else {
		log.Info("No authentication method provided, falling back to unauthenticated clone")
	}

	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(apiUrl))
	if err != nil {
		return fmt.Errorf("failed to create GitLab client: %w", err)
	}

	g.Client = client
	return nil
}

func (g *Gitlab) InitWebhookHandler() error {
	handler, err := wh.New(wh.Options.Secret(g.Config.WebhookSecret))
	if err != nil {
		return err
	}
	g.WebhookHandler = handler
	return nil
}

func (g *Gitlab) GetChanges(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) ([]string, error) {
	id, err := strconv.Atoi(pr.Spec.ID)
	if err != nil {
		log.Errorf("Error while parsing Gitlab merge request ID: %s", err)
		return []string{}, err
	}
	listOpts := gitlab.ListMergeRequestDiffsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 20,
		},
	}
	var changes []string
	for {
		diffs, resp, err := g.Client.MergeRequests.ListMergeRequestDiffs(getGitlabNamespacedName(repository.Spec.Repository.Url), id, &listOpts)
		if err != nil {
			log.Errorf("Error while getting merge request changes: %s", err)
			return []string{}, err
		}
		for _, change := range diffs {
			changes = append(changes, change.NewPath)
		}
		if resp.NextPage == 0 {
			break
		}
		listOpts.Page = resp.NextPage
	}
	return changes, nil
}

func (g *Gitlab) Comment(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, comment comment.Comment) error {
	body, err := comment.Generate(pr.Annotations[annotations.LastBranchCommit])
	if err != nil {
		log.Errorf("Error while generating comment: %s", err)
		return err
	}
	id, err := strconv.Atoi(pr.Spec.ID)
	if err != nil {
		log.Errorf("Error while parsing Gitlab merge request ID: %s", err)
		return err
	}
	_, _, err = g.Client.Notes.CreateMergeRequestNote(getGitlabNamespacedName(repository.Spec.Repository.Url), id, &gitlab.CreateMergeRequestNoteOptions{
		Body: gitlab.Ptr(body),
	})
	if err != nil {
		log.Errorf("Error while creating merge request note: %s", err)
		return err
	}
	return nil
}

func (g *Gitlab) Clone(repository *configv1alpha1.TerraformRepository, branch string, repositoryPath string) (*git.Repository, error) {
	auth, err := g.GetGitAuth()
	if err != nil {
		return nil, err
	}

	cloneOptions := &git.CloneOptions{
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		URL:           repository.Spec.Repository.Url,
		Auth:          auth,
	}

	if auth == nil {
		return nil, errors.New("no valid authentication method provided")
	}

	log.Infof("Cloning gitlab repository %s on %s branch", repository.Spec.Repository.Url, branch)
	repo, err := git.PlainClone(repositoryPath, false, cloneOptions)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (g *Gitlab) ParseWebhookPayload(r *nethttp.Request) (interface{}, bool) {
	// if the request is not a GitLab event, return false
	if r.Header.Get("X-Gitlab-Event") == "" {
		return nil, false
	} else {
		// check if the request can be verified with the secret of this provider
		p, err := g.WebhookHandler.Parse(r, wh.PushEvents, wh.TagEvents, wh.MergeRequestEvents)
		if errors.Is(err, wh.ErrGitLabTokenVerificationFailed) {
			return nil, false
		} else if err != nil {
			log.Errorf("an error occurred during request parsing: %s", err)
			return nil, false
		}
		return p, true
	}
}

func (g *Gitlab) GetEventFromWebhookPayload(p interface{}) (event.Event, error) {
	var e event.Event

	switch payload := p.(type) {
	case wh.PushEventPayload:
		log.Infof("parsing Gitlab push event payload")
		changedFiles := []string{}
		for _, commit := range payload.Commits {
			changedFiles = append(changedFiles, commit.Added...)
			changedFiles = append(changedFiles, commit.Modified...)
			changedFiles = append(changedFiles, commit.Removed...)
		}
		e = &event.PushEvent{
			URL:      utils.NormalizeUrl(payload.Project.WebURL),
			Revision: event.ParseRevision(payload.Ref),
			ChangeInfo: event.ChangeInfo{
				ShaBefore: payload.Before,
				ShaAfter:  payload.After,
			},
			Changes: changedFiles,
		}
	case wh.MergeRequestEventPayload:
		log.Infof("parsing Gitlab merge request event payload")
		e = &event.PullRequestEvent{
			ID:       strconv.Itoa(int(payload.ObjectAttributes.IID)),
			URL:      utils.NormalizeUrl(payload.Project.WebURL),
			Revision: payload.ObjectAttributes.SourceBranch,
			Action:   getNormalizedAction(payload.ObjectAttributes.Action),
			Base:     payload.ObjectAttributes.TargetBranch,
			Commit:   payload.ObjectAttributes.LastCommit.ID,
		}
	default:
		return nil, errors.New("unsupported event")
	}
	return e, nil
}

// Required API scope: api read_api
func (g *Gitlab) GetLatestRevisionForRef(repository *configv1alpha1.TerraformRepository, ref string) (string, error) {
	projectID := getGitlabNamespacedName(repository.Spec.Repository.Url)
	b, _, err := g.Client.Branches.GetBranch(projectID, ref)
	if err == nil {
		return b.Commit.ID, nil
	}
	t, _, err := g.Client.Tags.GetTag(projectID, ref)
	if err == nil {
		return t.Commit.ID, nil
	}
	return "", fmt.Errorf("could not find revision for ref %s: %w", ref, err)
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

// GetGitAuth returns the appropriate authentication method for GitLab
func (g *Gitlab) GetGitAuth() (transport.AuthMethod, error) {
	if g.Config.GitLabToken != "" {
		return &http.BasicAuth{
			Username: "oauth2",
			Password: g.Config.GitLabToken,
		}, nil
	} else if g.Config.Username != "" && g.Config.Password != "" {
		return &http.BasicAuth{
			Username: g.Config.Username,
			Password: g.Config.Password,
		}, nil
	}
	log.Info("No authentication method provided, falling back to unauthenticated clone")
	return nil, nil
}
