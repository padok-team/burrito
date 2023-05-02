package webhook_test

// import (
// 	"testing"

// 	"github.com/padok-team/burrito/internal/burrito/config"
// 	"github.com/padok-team/burrito/internal/webhook"
// 	"github.com/stretchr/testify/assert"
// )

// func TestWebhook_Init(t *testing.T) {
// 	secret := "test-secret"
// 	config := &config.Config{
// 		Server: config.Server{
// 			Webhook: config.WebhookConfig{
// 				Github: config.WebhookGithubConfig{
// 					Secret: secret,
// 				},
// 			},
// 		},
// 	}

// 	w := webhook.New(config)
// 	err := w.Init()
// 	assert.NoError(t, err)
// }
