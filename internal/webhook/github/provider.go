package github

import (
	"errors"
	"net/http"
	"strconv"

	utils "github.com/padok-team/burrito/internal/utils/url"

	"github.com/go-playground/webhooks/github"
	"github.com/padok-team/burrito/internal/webhook/event"

	log "github.com/sirupsen/logrus"
)

type Github struct {
	github *github.Webhook
	Secret string
}

func (g *Github) Init() error {
	githubWebhook, err := github.New(github.Options.Secret(g.Secret))
	if err != nil {
		return err
	}
	g.github = githubWebhook
	return nil
}

func (g *Github) ParseFromProvider(r *http.Request) (interface{}, bool) {
	// if the request is not a GitHub event, return false
	if r.Header.Get("X-GitHub-Event") == "" {
		return nil, false
	} else {
		// check if the request can be verified with the secret of this provider
		p, err := g.github.Parse(r, github.PushEvent, github.PingEvent, github.PullRequestEvent)
		if errors.Is(err, github.ErrHMACVerificationFailed) {
			return nil, false
		} else if err != nil {
			log.Errorf("an error occurred during request parsing : %s", err)
			return nil, false
		}
		return p, true
	}
}

func (g *Github) GetEvent(p interface{}) (event.Event, error) {
	var e event.Event
	var err error
	switch payload := p.(type) {
	case github.PushPayload:
		log.Infof("parsing Github push event payload")
		changedFiles := []string{}
		for _, commit := range payload.Commits {
			changedFiles = append(changedFiles, commit.Added...)
			changedFiles = append(changedFiles, commit.Modified...)
			changedFiles = append(changedFiles, commit.Removed...)
		}
		e = &event.PushEvent{
			URL:      utils.NormalizeUrl(payload.Repository.HTMLURL),
			Revision: event.ParseRevision(payload.Ref),
			ChangeInfo: event.ChangeInfo{
				ShaBefore: payload.Before,
				ShaAfter:  payload.After,
			},
			Changes: changedFiles,
		}
	case github.PullRequestPayload:
		log.Infof("parsing Github pull request event payload")
		if err != nil {
			log.Warnf("could not retrieve pull request from Github API: %s", err)
			return nil, err
		}
		e = &event.PullRequestEvent{
			Provider: "github",
			ID:       strconv.FormatInt(payload.PullRequest.Number, 10),
			URL:      utils.NormalizeUrl(payload.Repository.HTMLURL),
			Revision: payload.PullRequest.Head.Ref,
			Action:   getNormalizedAction(payload.Action),
			Base:     payload.PullRequest.Base.Ref,
			Commit:   payload.PullRequest.Head.Sha,
		}
	default:
		return nil, errors.New("unsupported Event")
	}
	return e, nil
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
