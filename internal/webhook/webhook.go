package webhook

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-playground/webhooks/v6/github"
	"github.com/go-playground/webhooks/v6/gitlab"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/burrito/config"

	"sigs.k8s.io/controller-runtime/pkg/client"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
)

type Handler interface {
	Handle()
}

type Webhook struct {
	client.Client
	config *config.Config
	github *github.Webhook
	gitlab *gitlab.Webhook
}

func New(c *config.Config) *Webhook {
	return &Webhook{
		config: c,
	}
}

func (w *Webhook) Init() error {
	githubWebhook, err := github.New(github.Options.Secret(w.config.Webhook.Github.Secret))
	if err != nil {
		return err
	}
	w.github = githubWebhook
	gitlabWebhook, err := gitlab.New(gitlab.Options.Secret(w.config.Webhook.Gitlab.Secret))
	if err != nil {
		return err
	}
	w.gitlab = gitlabWebhook
	return nil
}

func (w *Webhook) Handle(payload interface{}) {
	webUrls, revision, change, touchedHead, changedFiles := affectedRevisionInfo(payload)
	if len(webUrls) == 0 {
		fmt.Println("Ignoring webhook event")
		return
	}
	for _, webURL := range webUrls {
		fmt.Println("Received push event repo: %s, revision: %s, touchedHead: %v", webURL, revision, touchedHead)
	}
	// The next 2 lines probably dont work, waiting for chat GPT to be up \o/
	repositories := &configv1alpha1.TerraformRepositoryList{}
	err := w.Client.List(context.TODO(), repositories)
	if err != nil {
		fmt.Println("could not get repositories")
	}

	for _, url := range webUrls {
		for _, repo := range repositories.Items {
			if repo.Spec.Repository.Url != url {
				continue
			}
			// The next 2 lines probably dont work, waiting for chat GPT to be up \o/
			// Should we link lkayer and repositories by label to make this list easier?
			layers := &configv1alpha1.TerraformLayerList{}
			err := w.Client.List(context.TODO(), layers)
			if err != nil {
				fmt.Println("could not get layers")
			}
			for _, layer := range layers.Items {
				if layerFilesHaveChanged(&layer, changedFiles) {
					ann := map[string]string{}
					ann[annotations.LastBranchCommit] = change.shaAfter
					err = annotations.AddAnnotations(context.TODO(), w.Client, layer, ann)
				}
			}
		}
	}
	return
}

type changeInfo struct {
	shaBefore string
	shaAfter  string
}

func parseRevision(ref string) string {
	refParts := strings.SplitN(ref, "/", 3)
	return refParts[len(refParts)-1]
}

func affectedRevisionInfo(payloadIf interface{}) (webUrls []string, revision string, change changeInfo, touchedHead bool, changedFiles []string) {
	switch payload := payloadIf.(type) {
	case github.PushPayload:
		webUrls = append(webUrls, payload.Repository.HTMLURL)
		revision = parseRevision(payload.Ref)
		change.shaAfter = parseRevision(payload.After)
		change.shaBefore = parseRevision(payload.Before)
		touchedHead = bool(payload.Repository.DefaultBranch == revision)
		for _, commit := range payload.Commits {
			changedFiles = append(changedFiles, commit.Added...)
			changedFiles = append(changedFiles, commit.Modified...)
			changedFiles = append(changedFiles, commit.Removed...)
		}
	case gitlab.PushEventPayload:
		webUrls = append(webUrls, payload.Project.WebURL)
		revision = parseRevision(payload.Ref)
		change.shaAfter = parseRevision(payload.After)
		change.shaBefore = parseRevision(payload.Before)
		touchedHead = bool(payload.Project.DefaultBranch == revision)
		for _, commit := range payload.Commits {
			changedFiles = append(changedFiles, commit.Added...)
			changedFiles = append(changedFiles, commit.Modified...)
			changedFiles = append(changedFiles, commit.Removed...)
		}
	default:
		fmt.Println("event not handled")
	}
	return webUrls, revision, change, touchedHead, changedFiles
}

func layerFilesHaveChanged(layer *configv1alpha1.TerraformLayer, changedFiles []string) bool {
	// an empty slice of changed files means that the payload didn't include a list
	// of changed files and we have to assume that a refresh is required
	if len(changedFiles) == 0 {
		return true
	}

	// At last one changed file must be under refresh path
	for _, f := range changedFiles {
		f = ensureAbsPath(f)
		if strings.Contains(f, layer.Spec.Path) {
			return true
		}
	}

	return false
}

func ensureAbsPath(input string) string {
	if !filepath.IsAbs(input) {
		return string(filepath.Separator) + input
	}
	return input
}
