package gitlab

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-playground/webhooks/gitlab"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/webhook/event"
	log "github.com/sirupsen/logrus"
)

type Gitlab struct {
	gitlab *gitlab.Webhook
}

func (g *Gitlab) Init(c *config.Config) error {
	gitlabWebhook, err := gitlab.New(gitlab.Options.Secret(c.Server.Webhook.Gitlab.Secret))
	if err != nil {
		return err
	}
	g.gitlab = gitlabWebhook
	return nil
}

func (g *Gitlab) IsFromProvider(r *http.Request) bool {
	return r.Header.Get("X-Gitlab-Event") != ""
}

func (g *Gitlab) GetEvent(r *http.Request) (event.Event, error) {
	var e event.Event
	p, err := g.gitlab.Parse(r, gitlab.PushEvents, gitlab.TagEvents, gitlab.MergeRequestEvents)
	if errors.Is(err, gitlab.ErrGitLabTokenVerificationFailed) {
		log.Errorf("GitLab webhook token verification failed: %s", err)
		return nil, err
	}
	if err != nil {
		log.Errorf("an error occurred during event parsing: %s", err)
		return nil, err
	}
	switch payload := p.(type) {
	case gitlab.PushEventPayload:
		log.Infof("parsing Gitlab push event payload")
		changedFiles := []string{}
		for _, commit := range payload.Commits {
			changedFiles = append(changedFiles, commit.Added...)
			changedFiles = append(changedFiles, commit.Modified...)
			changedFiles = append(changedFiles, commit.Removed...)
		}
		e = &event.PushEvent{
			URL:      event.NormalizeUrl(payload.Project.WebURL),
			Revision: event.ParseRevision(payload.Ref),
			ChangeInfo: event.ChangeInfo{
				ShaBefore: payload.Before,
				ShaAfter:  payload.After,
			},
			Changes: changedFiles,
		}
	case gitlab.MergeRequestEventPayload:
		log.Infof("parsing Gitlab merge request event payload")
		e = &event.PullRequestEvent{
			Provider: "gitlab",
			ID:       strconv.Itoa(int(payload.ObjectAttributes.ID)),
			URL:      event.NormalizeUrl(payload.Project.WebURL),
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

func getNormalizedAction(action string) string {
	switch action {
	case "open":
		return event.PullRequestOpened
	case "close":
		return event.PullRequestClosed
	default:
		return action
	}
}
