package gitea

import (
	"errors"
	nethttp "net/http"
	"strconv"

	wh "github.com/go-playground/webhooks/v6/gitea"
	utils "github.com/padok-team/burrito/internal/utils/url"
	"github.com/padok-team/burrito/internal/webhook/event"
	log "github.com/sirupsen/logrus"
)

type WebhookProvider struct {
	*wh.Webhook
}

func (wp *WebhookProvider) ParseWebhookPayload(r *nethttp.Request) (interface{}, bool) {
	// if the request is not a Gitea/Forgejo event, return false. Forgejo
	// instances send both X-Forgejo-Event and X-Gitea-Event headers.
	if r.Header.Get("X-Gitea-Event") == "" {
		return nil, false
	} else {
		// check if the request can be verified with the secret of this provider
		p, err := wp.Webhook.Parse(r, wh.PushEvent, wh.PullRequestEvent, wh.PullRequestSyncEvent)
		if errors.Is(err, wh.ErrHMACVerificationFailed) {
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
	case wh.PushPayload:
		log.Infof("parsing Gitea push event payload")
		changedFiles := []string{}
		for _, commit := range payload.Commits {
			changedFiles = append(changedFiles, commit.Added...)
			changedFiles = append(changedFiles, commit.Modified...)
			changedFiles = append(changedFiles, commit.Removed...)
		}
		e = &event.PushEvent{
			URL:       utils.NormalizeUrl(payload.Repo.HTMLURL),
			Reference: event.ParseReference(payload.Ref),
			ChangeInfo: event.ChangeInfo{
				ShaBefore: payload.Before,
				ShaAfter:  payload.After,
			},
			Changes: changedFiles,
		}
	case wh.PullRequestPayload:
		log.Infof("parsing Gitea pull request event payload")
		if payload.PullRequest == nil || payload.PullRequest.Head == nil || payload.PullRequest.Base == nil {
			return nil, errors.New("malformed pull request payload: missing pull request, head or base")
		}
		e = &event.PullRequestEvent{
			ID:        strconv.FormatInt(payload.Index, 10),
			URL:       utils.NormalizeUrl(payload.Repository.HTMLURL),
			Reference: payload.PullRequest.Head.Ref,
			Action:    getNormalizedAction(string(payload.Action)),
			Base:      payload.PullRequest.Base.Ref,
			Commit:    payload.PullRequest.Head.Sha,
		}
	default:
		return nil, errors.New("unsupported event")
	}
	return e, nil
}
