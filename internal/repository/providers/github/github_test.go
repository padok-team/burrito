package github_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/padok-team/burrito/internal/repository/credentials"
	"github.com/padok-team/burrito/internal/repository/providers/github"
	"github.com/padok-team/burrito/internal/webhook/event"

	webhook "github.com/go-playground/webhooks/github"
	"github.com/stretchr/testify/assert"
)

func TestGithub_GetEventFromWebhookPayload_PushEvent(t *testing.T) {
	payloadFile, err := os.Open("testdata/github-push-main-event.json")
	if err != nil {
		t.Fatalf("failed to open payload file: %v", err)
	}
	defer payloadFile.Close()

	payloadBytes, err := io.ReadAll(payloadFile)
	if err != nil {
		t.Fatalf("failed to read payload file: %v", err)
	}

	var payload webhook.PushPayload
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	secret := "test-secret"
	github := &github.Github{
		Config: credentials.Credential{
			WebhookSecret: secret,
		},
	}
	webhookProvider, err := github.GetWebhookProvider()
	assert.NoError(t, err)

	req.Header.Set("X-GitHub-Event", "push")

	mac := hmac.New(sha1.New, []byte(secret))
	_, err = mac.Write(payloadBytes)
	assert.NoError(t, err)
	expectedMac := hex.EncodeToString(mac.Sum(nil))
	req.Header.Set("X-Hub-Signature", fmt.Sprintf("sha1=%s", expectedMac))

	parsed, ok := webhookProvider.ParseWebhookPayload(req)
	assert.True(t, ok)
	evt, err := webhookProvider.GetEventFromWebhookPayload(parsed)
	assert.NoError(t, err)
	assert.IsType(t, &event.PushEvent{}, evt)

	pushEvt := evt.(*event.PushEvent)
	assert.Equal(t, "https://github.com/padok-team/burrito-examples", pushEvt.URL)
	assert.Equal(t, "main", pushEvt.Reference)
	assert.Equal(t, "6f51b4ffd5e3adadfc3ee649d5ea2499472ea33b", pushEvt.ShaBefore)
	assert.Equal(t, "ca9b6c80ac8fb5cd837ae9b374b79ff33f472558", pushEvt.ShaAfter)
	assert.ElementsMatch(t, []string{"modules/random-pets/main.tf", "terragrunt/random-pets/test/inputs.hcl", "modules/random-pets/variables.tf"}, pushEvt.Changes)
}

func TestGithub_GetEventFromWebhookPayload_PullRequestEvent(t *testing.T) {
	payloadFile, err := os.Open("testdata/github-open-pull-request-event.json")
	if err != nil {
		t.Fatalf("failed to open payload file: %v", err)
	}
	defer payloadFile.Close()

	payloadBytes, err := io.ReadAll(payloadFile)
	if err != nil {
		t.Fatalf("failed to read payload file: %v", err)
	}

	var payload webhook.PullRequestPayload
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	secret := "test-secret"
	github := &github.Github{
		Config: credentials.Credential{
			WebhookSecret: secret,
		},
	}
	webhookProvider, err := github.GetWebhookProvider()
	assert.NoError(t, err)

	req.Header.Set("X-GitHub-Event", "pull_request")

	mac := hmac.New(sha1.New, []byte(secret))
	_, err = mac.Write(payloadBytes)
	assert.NoError(t, err)
	expectedMac := hex.EncodeToString(mac.Sum(nil))
	req.Header.Set("X-Hub-Signature", fmt.Sprintf("sha1=%s", expectedMac))

	parsed, ok := webhookProvider.ParseWebhookPayload(req)
	assert.True(t, ok)
	evt, err := webhookProvider.GetEventFromWebhookPayload(parsed)
	assert.NoError(t, err)
	assert.IsType(t, &event.PullRequestEvent{}, evt)

	pullRequestEvt := evt.(*event.PullRequestEvent)
	assert.Equal(t, "20", pullRequestEvt.ID)
	assert.Equal(t, "https://github.com/padok-team/burrito-examples", pullRequestEvt.URL)
	assert.Equal(t, "demo", pullRequestEvt.Reference)
	assert.Equal(t, "main", pullRequestEvt.Base)
	assert.Equal(t, "faf5e25402a9bd10f7318c8a2cd984af576c687f", pullRequestEvt.Commit)
	assert.Equal(t, "opened", pullRequestEvt.Action)
}
