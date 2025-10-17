package gitlab

import (
	"errors"
	nethttp "net/http"
	"strconv"

	wh "github.com/go-playground/webhooks/gitlab"
	utils "github.com/padok-team/burrito/internal/utils/url"
	"github.com/padok-team/burrito/internal/webhook/event"
	log "github.com/sirupsen/logrus"
)

type WebhookProvider struct {
	*wh.Webhook
}

func (wp *WebhookProvider) ParseWebhookPayload(r *nethttp.Request) (interface{}, bool) {
	// if the request is not a GitLab event, return false
	if r.Header.Get("X-Gitlab-Event") == "" {
		return nil, false
	} else {
		// check if the request can be verified with the secret of this provider
		p, err := wp.Webhook.Parse(r, wh.PushEvents, wh.TagEvents, wh.MergeRequestEvents)
		if errors.Is(err, wh.ErrGitLabTokenVerificationFailed) {
			return nil, false
		} else if err != nil {
			log.Errorf("an error occurred during request parsing: %s", err)
			return nil, false
		}
		return p, true
	}
}

func (wp *WebhookProvider) GetEventFromWebhookPayload(p interface{}) (event.Event, error) {
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
			URL:       utils.NormalizeUrl(payload.Project.WebURL),
			Reference: event.ParseReference(payload.Ref),
			ChangeInfo: event.ChangeInfo{
				ShaBefore: payload.Before,
				ShaAfter:  payload.After,
			},
			Changes: changedFiles,
		}
	case wh.MergeRequestEventPayload:
		log.Infof("parsing Gitlab merge request event payload")
		e = &event.PullRequestEvent{
			ID:        strconv.Itoa(int(payload.ObjectAttributes.IID)),
			URL:       utils.NormalizeUrl(payload.Project.WebURL),
			Reference: payload.ObjectAttributes.SourceBranch,
			Action:    getNormalizedAction(payload.ObjectAttributes.Action),
			Base:      payload.ObjectAttributes.TargetBranch,
			Commit:    payload.ObjectAttributes.LastCommit.ID,
		}
	default:
		return nil, errors.New("unsupported event")
	}
	return e, nil
}
