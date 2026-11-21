package gitea_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/padok-team/burrito/internal/repository/credentials"
	"github.com/padok-team/burrito/internal/repository/providers/gitea"
	"github.com/padok-team/burrito/internal/webhook/event"

	webhook "github.com/go-playground/webhooks/v6/gitea"
	"github.com/stretchr/testify/assert"
)

func TestGitea_GetEventFromWebhookPayload_PushEvent(t *testing.T) {
	payloadFile, err := os.Open("testdata/gitea-push-main-event.json")
	if err != nil {
		t.Fatalf("failed to open payload file: %v", err)
	}
	defer func() {
		_ = payloadFile.Close()
	}()

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
	gitea := &gitea.Gitea{
		Config: credentials.Credential{
			WebhookSecret: secret,
		},
	}
	webhookProvider, err := gitea.GetWebhookProvider()
	assert.NoError(t, err)

	req.Header.Set("X-Gitea-Event", "push")

	mac := hmac.New(sha256.New, []byte(secret))
	_, err = mac.Write(payloadBytes)
	assert.NoError(t, err)
	expectedMac := hex.EncodeToString(mac.Sum(nil))
	req.Header.Set("X-Gitea-Signature", expectedMac)

	parsed, ok := webhookProvider.ParseWebhookPayload(req)
	assert.True(t, ok)
	evt, err := webhookProvider.GetEventFromWebhookPayload(parsed)
	assert.NoError(t, err)
	assert.IsType(t, &event.PushEvent{}, evt)

	pushEvt := evt.(*event.PushEvent)
	assert.Equal(t, "https://gitea.example.com/padok-team/burrito-examples", pushEvt.URL)
	assert.Equal(t, "main", pushEvt.Reference)
	assert.Equal(t, "0000000000000000000000000000000000000000", pushEvt.ShaBefore)
	assert.Equal(t, "2ebd0cab3bddb8c4c664b0b2fe8b1a0c6c1c1e1b", pushEvt.ShaAfter)
	assert.ElementsMatch(t, []string{"modules/random-pets/main.tf", "terragrunt/random-pets/test/inputs.hcl", "modules/random-pets/variables.tf"}, pushEvt.Changes)
}

func TestGitea_GetEventFromWebhookPayload_PullRequestEvent(t *testing.T) {
	payloadFile, err := os.Open("testdata/gitea-open-pull-request-event.json")
	if err != nil {
		t.Fatalf("failed to open payload file: %v", err)
	}
	defer func() {
		_ = payloadFile.Close()
	}()

	payloadBytes, err := io.ReadAll(payloadFile)
	if err != nil {
		t.Fatalf("failed to read payload file: %v", err)
	}

	var payload webhook.PullRequestPayload
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	testWithGivenAction := func(action string, expected string) {
		payload.Action = webhook.HookIssueAction(action)
		payloadBytes, err := json.Marshal(payload)
		assert.NoError(t, err)

		req, err := http.NewRequest("POST", "/", bytes.NewBuffer(payloadBytes))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		secret := "test-secret"
		gitea := &gitea.Gitea{
			Config: credentials.Credential{
				WebhookSecret: secret,
			},
		}
		webhookProvider, err := gitea.GetWebhookProvider()
		assert.NoError(t, err)

		req.Header.Set("X-Gitea-Event", "pull_request")

		mac := hmac.New(sha256.New, []byte(secret))
		_, err = mac.Write(payloadBytes)
		assert.NoError(t, err)
		expectedMac := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Gitea-Signature", expectedMac)

		parsed, ok := webhookProvider.ParseWebhookPayload(req)
		assert.True(t, ok)
		evt, err := webhookProvider.GetEventFromWebhookPayload(parsed)
		assert.NoError(t, err)
		assert.IsType(t, &event.PullRequestEvent{}, evt)

		pullRequestEvt := evt.(*event.PullRequestEvent)
		assert.Equal(t, "20", pullRequestEvt.ID)
		assert.Equal(t, "https://gitea.example.com/padok-team/burrito-examples", pullRequestEvt.URL)
		assert.Equal(t, "demo", pullRequestEvt.Reference)
		assert.Equal(t, "main", pullRequestEvt.Base)
		assert.Equal(t, "faf5e25402a9bd10f7318c8a2cd984af576c687f", pullRequestEvt.Commit)
		assert.Equal(t, expected, pullRequestEvt.Action)
	}

	testWithGivenAction("opened", event.PullRequestOpened)
	testWithGivenAction("reopened", event.PullRequestOpened)
	testWithGivenAction("closed", event.PullRequestClosed)
}

func TestGitea_GetEventFromWebhookPayload_RejectsMalformedPullRequestPayload(t *testing.T) {
	// The PR payload is missing the nested pull_request object entirely; the
	// HMAC signature below is computed over exactly these bytes, so signature
	// verification succeeds and only the nil-guard can prevent a panic.
	payloadBytes := []byte(`{"action":"opened","number":20}`)

	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	secret := "test-secret"
	provider := &gitea.Gitea{
		Config: credentials.Credential{
			WebhookSecret: secret,
		},
	}
	webhookProvider, err := provider.GetWebhookProvider()
	assert.NoError(t, err)

	req.Header.Set("X-Gitea-Event", "pull_request")

	mac := hmac.New(sha256.New, []byte(secret))
	_, err = mac.Write(payloadBytes)
	assert.NoError(t, err)
	expectedMac := hex.EncodeToString(mac.Sum(nil))
	req.Header.Set("X-Gitea-Signature", expectedMac)

	parsed, ok := webhookProvider.ParseWebhookPayload(req)
	assert.True(t, ok)
	_, err = webhookProvider.GetEventFromWebhookPayload(parsed)
	assert.Error(t, err, "malformed payload must be rejected, not panic")
}

func TestGitea_GetEventFromWebhookPayload_RejectsBadSignature(t *testing.T) {
	payloadBytes, err := os.ReadFile("testdata/gitea-push-main-event.json")
	if err != nil {
		t.Fatalf("failed to read payload file: %v", err)
	}

	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	provider := &gitea.Gitea{
		Config: credentials.Credential{
			WebhookSecret: "test-secret",
		},
	}
	webhookProvider, err := provider.GetWebhookProvider()
	assert.NoError(t, err)

	req.Header.Set("X-Gitea-Event", "push")
	req.Header.Set("X-Gitea-Signature", "deadbeef")

	parsed, ok := webhookProvider.ParseWebhookPayload(req)
	assert.False(t, ok)
	assert.Nil(t, parsed)
}
