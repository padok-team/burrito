# Scope: Git Repository Access (`internal/repository/`)

Resolves a `TerraformRepository` to a provider and performs git/API operations against it.
`repository.go` dispatches on credentials; `providers/` holds the implementations
(`github/`, `gitlab/`, `standard/`, `mock/`), `credentials/` the `CredentialStore`,
`types/` the interfaces.

## Rules & Gotchas

- Providers are selected in `GetProviderFromCredentials` (`repository.go`) by
  `Credential.Provider` (`github`/`gitlab`/`standard`/`mock`). No credentials → public
  `standard` no-auth. **Adding a provider = add a `case` here AND implement the
  `types.Provider` interface** (`GetWebhookProvider` / `GetAPIProvider` / `GetGitProvider`).
- The `standard` provider drives go-git. **Do not use `Repository.Pull()`** — it has a
  non-fast-forward bug on non-default branches (see `providers/standard/repository_test.go`).
  Use `Fetch` + hard `Reset`, and always pass `ReferenceName` so go-git targets the right ref.
- `types.go` is a contract shared with `internal/webhook` (`WebhookProvider`) and
  `internal/controllers/terraformpullrequest` (`APIProvider`) — changing it ripples there.

## Validate

`make test` (unit tests live beside the code, e.g. `providers/standard/repository_test.go`),
`make vet`, `golangci-lint run ./...`.
