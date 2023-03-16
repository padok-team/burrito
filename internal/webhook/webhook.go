package webhook

import (
	"context"
	"errors"
	"fmt"
	"html"
	"net/http"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

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
	githubWebhook, err := github.New(github.Options.Secret(w.config.Server.Webhook.Github.Secret))
	if err != nil {
		return err
	}
	w.github = githubWebhook
	gitlabWebhook, err := gitlab.New(gitlab.Options.Secret(w.config.Server.Webhook.Gitlab.Secret))
	if err != nil {
		return err
	}
	w.gitlab = gitlabWebhook
	return nil
}

func (w *Webhook) GetHttpHandler() func(http.ResponseWriter, *http.Request) {
	log.Infof("webhook event received...")
	return func(writer http.ResponseWriter, r *http.Request) {
		var payload interface{}
		var err error

		switch {
		case r.Header.Get("X-GitHub-Event") != "":
			log.Infof("webhook has detected a GitHub event")
			payload, err = w.github.Parse(r, github.PushEvent, github.PingEvent)
			if errors.Is(err, github.ErrHMACVerificationFailed) {
				log.Errorf("GitHub webhook HMAC verification failed: %s", err)
			}
		case r.Header.Get("X-Gitlab-Event") != "":
			log.Infof("webhook has detected a GitLab event")
			payload, err = w.gitlab.Parse(r, gitlab.PushEvents, gitlab.TagEvents)
			if errors.Is(err, gitlab.ErrGitLabTokenVerificationFailed) {
				log.Errorf("GitLab webhook token verification failed: %s", err)
			}
		default:
			log.Infof("ignoring unknown webhook event")
			http.Error(writer, "Unknown webhook event", http.StatusBadRequest)
			return
		}

		if err != nil {
			log.Errorf("webhook processing failed: %s", err)
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
	webUrls, sshUrls, revision, change, touchedHead, changedFiles := affectedRevisionInfo(payload)
	allUrls := append(webUrls, sshUrls...)

	if len(allUrls) == 0 {
		log.Infof("ignoring webhook event")
		return
	}
	for _, url := range allUrls {
		log.Infof("received event repo: %s, revision: %s, touchedHead: %v", url, revision, touchedHead)
	}

	repositories := &configv1alpha1.TerraformRepositoryList{}
	err := w.Client.List(context.TODO(), repositories)
	if err != nil {
		log.Errorf("could not get terraform repositories: %s", err)
	}

	for _, url := range allUrls {
		for _, repo := range repositories.Items {
			log.Infof("evaluating terraform repository %s for url %s", repo.Name, url)
			if repo.Spec.Repository.Url != url {
				log.Infof("evaluating terraform repository %s url %s not matching %s", repo.Name, repo.Spec.Repository.Url, url)
				continue
			}
			layers := &configv1alpha1.TerraformLayerList{}
			err := w.Client.List(context.TODO(), layers, &client.ListOptions{})
			if err != nil {
				log.Errorf("could not get terraform layers: %s", err)
			}
			for _, layer := range layers.Items {
				ann := map[string]string{}
				log.Printf("evaluating terraform layer %s for revision %s", layer.Name, revision)
				if layer.Spec.Branch != revision {
					log.Infof("branch %s for terraform layer %s not matching revision %s", layer.Spec.Branch, layer.Name, revision)
					continue
				}
				ann[annotations.LastBranchCommit] = change.shaAfter
				if layerFilesHaveChanged(&layer, changedFiles) {
					ann[annotations.LastRelevantCommit] = change.shaAfter
				}
				err = annotations.Add(context.TODO(), w.Client, layer, ann)
				if err != nil {
					log.Errorf("could not add annotation to terraform layer %s", err)
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

func affectedRevisionInfo(payloadIf interface{}) (webUrls []string, sshUrls []string, revision string, change changeInfo, touchedHead bool, changedFiles []string) {
	switch payload := payloadIf.(type) {
	case github.PushPayload:
		log.Infof("parsing GitHub push event payload")
		webUrls = append(webUrls, payload.Repository.HTMLURL)
		sshUrls = append(sshUrls, payload.Repository.SSHURL)
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
		log.Infof("parsing GitLab push event payload")
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
		log.Infof("event not handled")
	}
	return webUrls, sshUrls, revision, change, touchedHead, changedFiles
}

func layerFilesHaveChanged(layer *configv1alpha1.TerraformLayer, changedFiles []string) bool {
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
