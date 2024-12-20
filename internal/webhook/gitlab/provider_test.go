package gitlab_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"

	"net/http"
	"testing"

	"github.com/padok-team/burrito/internal/webhook/event"
	"github.com/padok-team/burrito/internal/webhook/gitlab"

	webhook "github.com/go-playground/webhooks/gitlab"
	"github.com/stretchr/testify/assert"
)

func TestGitlab_GetEvent_PushEvent(t *testing.T) {
	payloadFile, err := os.Open("testdata/gitlab-push-main-event.json")
	if err != nil {
		t.Fatalf("failed to open payload file: %v", err)
	}
	defer payloadFile.Close()

	payloadBytes, err := io.ReadAll(payloadFile)
	if err != nil {
		t.Fatalf("failed to read payload file: %v", err)
	}

	var payload webhook.PushEventPayload
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	secret := "test-secret"
	gitlab := gitlab.Gitlab{
		Secret: secret,
	}
	err = gitlab.Init()
	assert.NoError(t, err)

	req.Header.Set("X-GitLab-Event", "Push Hook")
	req.Header.Set("X-Gitlab-Token", secret)

	parsed, ok := gitlab.ParseFromProvider(req)
	assert.True(t, ok)
	evt, err := gitlab.GetEvent(parsed)
	assert.NoError(t, err)
	assert.IsType(t, &event.PushEvent{}, evt)

	pushEvt := evt.(*event.PushEvent)
	assert.Equal(t, "https://gitlab.com/burrito/examples", pushEvt.URL)
	assert.Equal(t, "main", pushEvt.Revision)
	assert.Equal(t, "95790bf891e76fee5e1747ab589903a6a1f80f22", pushEvt.ShaBefore)
	assert.Equal(t, "da1560886d4f094c3e6c9ef40349f7d38b5d27d7", pushEvt.ShaAfter)
	assert.ElementsMatch(t, []string{"test.hcl", "layer-1/prod.hcl", "layer-2/staging.hcl"}, pushEvt.Changes)
}

func TestGitlab_GetEvent_MergeRequestEvent(t *testing.T) {
	// Test GitLab initialization
	secret := "test-secret"
	gitlab := gitlab.Gitlab{
		Secret: secret,
	}
	err := gitlab.Init()
	assert.NoError(t, err)

	// Test event handling
	payloadFile, err := os.Open("testdata/gitlab-open-merge-request-event.json")
	if err != nil {
		t.Fatalf("failed to open payload file: %v", err)
	}
	defer payloadFile.Close()

	payloadBytes, err := io.ReadAll(payloadFile)
	if err != nil {
		t.Fatalf("failed to read payload file: %v", err)
	}

	var payload webhook.MergeRequestEventPayload
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	testWithGivenAction := func(action string, expected string) {
		payload.ObjectAttributes.Action = action
		payloadBytes, err := json.Marshal(payload)
		assert.NoError(t, err)

		req, err := http.NewRequest("POST", "/", bytes.NewBuffer(payloadBytes))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		req.Header.Set("X-GitLab-Event", "Merge Request Hook")
		req.Header.Set("X-Gitlab-Token", secret)

		parsed, ok := gitlab.ParseFromProvider(req)
		assert.True(t, ok)
		evt, err := gitlab.GetEvent(parsed)
		assert.NoError(t, err)
		assert.IsType(t, &event.PullRequestEvent{}, evt)

		pullRequestEvt := evt.(*event.PullRequestEvent)
		assert.Equal(t, "1", pullRequestEvt.ID)
		assert.Equal(t, "https://example.com/gitlabhq/gitlab-test", pullRequestEvt.URL)
		assert.Equal(t, "demo", pullRequestEvt.Revision)
		assert.Equal(t, "main", pullRequestEvt.Base)
		assert.Equal(t, "da1560886d4f094c3e6c9ef40349f7d38b5d27d7", pullRequestEvt.Commit)
		assert.Equal(t, expected, pullRequestEvt.Action)
	}

	testWithGivenAction("open", event.PullRequestOpened)
	testWithGivenAction("reopen", event.PullRequestOpened)
	testWithGivenAction("close", event.PullRequestClosed)
	testWithGivenAction("merge", event.PullRequestClosed)
}
