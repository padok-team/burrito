# Scope: Dashboard Server (`internal/server/`)

Serves the `ui/` dashboard and its JSON API. Echo v4 (`server.go`), one binary
(`cmd/server`). `api/` holds the handlers backing the UI, `auth/` the auth backends.

## Rules & Gotchas

- The UI is embedded with `//go:embed all:dist` — the `ui/` build output must exist
  (`yarn --cwd ui build`) before the server binary is built, or the embed fails.
- `/api/*` handlers in `api/` are the contract the `ui/` TypeScript consumes (`/layers`,
  `/repositories`, `/logs/...`, `/run/.../attempts`, sync). Treat their response shapes like
  the CRD↔UI contract: **flag breaking changes** and update `ui/` in step.
- Reads live k8s state via a controller-runtime `client` and run artifacts via the
  `datastore/client` — the server owns no storage itself.
- Auth is OIDC (`auth/oauth`) or Basic (`auth/basic`), selected by config; with neither the
  server is public (it warns). Auth backends implement `auth.AuthHandlers`; sessions are
  cookie-based (`burrito_session`). Keep new `/api` routes behind `authMiddleware`.
- `POST /api/webhook` is intentionally unauthenticated (handled by `internal/webhook`);
  don't move it behind auth.

## Validate

`make test`, `make vet`, `golangci-lint run ./...`. Build needs the UI: `yarn --cwd ui build`
then `make build`.
