# Scope: VCS Webhook Receiver (`internal/webhook/`)

Handles inbound GitHub/GitLab webhooks, mounted (unauthenticated) at `POST /api/webhook` by
`internal/server`. `webhook.go` is the dispatch; `event/` the parsed events that mutate k8s.

## Rules & Gotchas

- This package is **provider-agnostic dispatch**, not VCS-specific parsing.
  `tryGetEventFromPayload` iterates over every known credential (shared + per-repo) and asks
  each provider's `WebhookProvider.ParseWebhookPayload` to claim the request. The first match
  wins; no match → `nil` event → HTTP 400.
- **Signature/secret verification lives in each provider's `ParseWebhookPayload`**
  (`internal/repository/providers/*`), not here. To add a VCS, implement `WebhookProvider`
  there and reuse `repo.GetProviderFromCredentials` — do not special-case a provider in this
  package.
- The endpoint is unauthenticated by design (secret-based validation is the gate). Keep error
  responses HTML-escaped (`html.EscapeString`) — the payload is untrusted input reflected back.
- Parsed events implement `event.Handle(client)`, which updates resources via the
  controller-runtime client; keep handling idempotent (a webhook can be redelivered).

## Validate

`make test`, `make vet`, `golangci-lint run ./...`.
