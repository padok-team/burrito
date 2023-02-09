package webhook

import (
	"context"
	"errors"
	"fmt"
	"html"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-playground/webhooks/v6/github"
	"github.com/go-playground/webhooks/v6/gitlab"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/burrito/config"
	"k8s.io/apimachinery/pkg/runtime"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
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
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(configv1alpha1.AddToScheme(scheme))
	cl, err := client.New(ctrl.GetConfigOrDie(), client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return err
	}
	w.Client = cl
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

func (w *Webhook) GetHttpHandler() func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, r *http.Request) {
		var payload interface{}
		var err error

		switch {
		case r.Header.Get("X-GitHub-Event") != "":
			log.Println("Detected a Github event")
			payload, err = w.github.Parse(r, github.PushEvent, github.PingEvent)
			if errors.Is(err, github.ErrHMACVerificationFailed) {
				log.Println("GitHub webhook HMAC verification failed")
			}
		case r.Header.Get("X-Gitlab-Event") != "":
			payload, err = w.gitlab.Parse(r, gitlab.PushEvents, gitlab.TagEvents)
			if errors.Is(err, gitlab.ErrGitLabTokenVerificationFailed) {
				log.Println("GitLab webhook token verification failed")
			}
		default:
			log.Println("Ignoring unknown webhook event")
			http.Error(writer, "Unknown webhook event", http.StatusBadRequest)
			return
		}

		if err != nil {
			log.Printf("Webhook processing failed: %s", err)
			status := http.StatusBadRequest
			if r.Method != "POST" {
				status = http.StatusMethodNotAllowed
			}
			http.Error(writer, fmt.Sprintf("Webhook processing failed: %s", html.EscapeString(err.Error())), status)
			return
		}

		w.Handle(payload)
	}
}

func (w *Webhook) Handle(payload interface{}) {
	webUrls, revision, change, touchedHead, changedFiles := affectedRevisionInfo(payload)
	if len(webUrls) == 0 {
		log.Println("Ignoring webhook event")
		return
	}
	for _, webURL := range webUrls {
		log.Printf("Received push event repo: %s, revision: %s, touchedHead: %v", webURL, revision, touchedHead)
	}
	// The next 2 lines probably dont work, waiting for chat GPT to be up \o/
	repositories := &configv1alpha1.TerraformRepositoryList{}
	err := w.Client.List(context.TODO(), repositories)
	if err != nil {
		log.Println("could not get repositories")
	}

	for _, url := range webUrls {
		for _, repo := range repositories.Items {
			if repo.Spec.Repository.Url != url {
				continue
			}
			// The next 2 lines probably dont work, waiting for chat GPT to be up \o/
			// Should we link lkayer and repositories by label to make this list easier?
			layers := &configv1alpha1.TerraformLayerList{}
			err := w.Client.List(context.TODO(), layers, &client.ListOptions{})
			if err != nil {
				log.Println("could not get layers")
			}
			for _, layer := range layers.Items {
				log.Printf("Evaluating %s", layer.Name)
				if layerFilesHaveChanged(&layer, changedFiles) {
					ann := map[string]string{}
					ann[annotations.LastBranchCommit] = change.shaAfter
					err = annotations.Add(context.TODO(), w.Client, layer, ann)
					if err != nil {
						log.Printf("Error adding annotation to layer %s", err)
					}
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
		log.Println("event not handled")
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
