package bitbucket

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	utils "github.com/padok-team/burrito/internal/utils/url"
	"github.com/padok-team/burrito/internal/webhook/event"
	log "github.com/sirupsen/logrus"
)

type WebhookProvider struct{}

func (wp *WebhookProvider) ParseWebhookPayload(r *http.Request) (interface{}, bool) {
	if r.Header.Get("X-Event-Key") == "" {
		return nil, false
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errorf("failed to read Bitbucket webhook body: %s", err)
		return nil, false
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, false
	}
	return payload, true
}

func (wp *WebhookProvider) GetEventFromWebhookPayload(p interface{}) (event.Event, error) {
	payload, ok := p.(map[string]interface{})
	if !ok {
		return nil, nil
	}
	eventKey, _ := payload["eventKey"].(string)

	switch {
	case eventKey == "repo:push" || eventKey == "push":
		push, _ := payload["push"].(map[string]interface{})
		changes := []string{}
		if push != nil {
			if ch, ok := push["changes"].([]interface{}); ok {
				for _, c := range ch {
					if cm, ok := c.(map[string]interface{}); ok {
						if old, ok := cm["old"].(map[string]interface{}); ok {
							if path, ok := old["path"].(string); ok {
								changes = append(changes, path)
							}
						}
					}
				}
			}
		}
		repo, _ := payload["repository"].(map[string]interface{})
		links, _ := repo["links"].(map[string]interface{})
		selfLink, _ := links["self"].(map[string]interface{})
		href, _ := selfLink["href"].(string)
		return &event.PushEvent{
			URL:       utils.NormalizeUrl(href),
			Reference: event.ParseReference("refs/heads/main"),
			Changes:   changes,
		}, nil

	case strings.Contains(eventKey, "pullrequest"):
		pr, _ := payload["pullrequest"].(map[string]interface{})
		if pr == nil {
			return nil, nil
		}
		id := fmt.Sprintf("%.0f", pr["id"].(float64))
		source, _ := pr["source"].(map[string]interface{})
		dest, _ := pr["destination"].(map[string]interface{})
		sourceBranch, _ := source["branch"].(map[string]interface{})
		destBranch, _ := dest["branch"].(map[string]interface{})
		
		action := event.PullRequestOpened
		if strings.Contains(eventKey, "merged") || strings.Contains(eventKey, "declined") {
			action = event.PullRequestClosed
		}
		
		return &event.PullRequestEvent{
			ID:        id,
			URL:       utils.NormalizeUrl(""),
			Reference: sourceBranch["name"].(string),
			Action:    action,
			Base:      destBranch["name"].(string),
		}, nil
	}
	return nil, nil
}
