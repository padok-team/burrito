package github

import (
	"errors"
	nethttp "net/http"
	"strconv"

	wh "github.com/go-playground/webhooks/github"

	"github.com/labstack/gommon/log"
	utils "github.com/padok-team/burrito/internal/utils/url"
	"github.com/padok-team/burrito/internal/webhook/event"
)

type WebhookProvider struct {
	*wh.Webhook
}

func (wp *WebhookProvider) ParseWebhookPayload(r *nethttp.Request) (interface{}, bool) {
	// if the request is not a GitHub event, return false
	if r.Header.Get("X-GitHub-Event") == "" {
		return nil, false
	} else {
		// check if the request can be verified with the secret of this provider
		p, err := wp.Webhook.Parse(r, wh.PushEvent, wh.PingEvent, wh.PullRequestEvent)
		if errors.Is(err, wh.ErrHMACVerificationFailed) {
			return nil, false
		} else if err != nil {
			log.Errorf("an error occurred during request parsing : %s", err)
			return nil, false
		}
		return p, true
	}
}

func (wp *WebhookProvider) GetEventFromWebhookPayload(p interface{}) (event.Event, error) {
	var e event.Event
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
