package gitlab

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-playground/webhooks/gitlab"
	utils "github.com/padok-team/burrito/internal/utils/url"
	"github.com/padok-team/burrito/internal/webhook/event"
	log "github.com/sirupsen/logrus"
)

type Gitlab struct {
	gitlab *gitlab.Webhook
	Secret string
}

func (g *Gitlab) Init() error {
	gitlabWebhook, err := gitlab.New(gitlab.Options.Secret(g.Secret))
	if err != nil {
		return err
	}
	g.gitlab = gitlabWebhook
	return nil
}

func (g *Gitlab) ParseFromProvider(r *http.Request) (interface{}, bool) {
	// if the request is not a GitLab event, return false
	if r.Header.Get("X-Gitlab-Event") == "" {
		return nil, false
	} else {
		// check if the request can be verified with the secret of this provider
		p, err := g.gitlab.Parse(r, gitlab.PushEvents, gitlab.TagEvents, gitlab.MergeRequestEvents)
		if errors.Is(err, gitlab.ErrGitLabTokenVerificationFailed) {
			return nil, false
		} else if err != nil {
			log.Errorf("an error occurred during request parsing: %s", err)
			return nil, false
		}
		return p, true
	}
}

func (g *Gitlab) GetEvent(p interface{}) (event.Event, error) {
	var e event.Event

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
			URL:      utils.NormalizeUrl(payload.Project.WebURL),
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
