# Scope: Datastore (`internal/datastore/`)

Stores run artifacts (logs, plans, git bundles). Runs as its own binary
(`cmd/datastore`): an Echo HTTP service (`datastore.go`) fronting a pluggable object store.
`storage/` holds the backends, `api/` the HTTP handlers, `client/` the client the
runner/server/controllers use to reach it.

## Rules & Gotchas

- Backends implement `storage.StorageBackend` (`storage/common.go`) — `s3/`, `gcs/`,
  `azure/`, `mock/`. `storage.New` selects one via a `switch` on config. **Adding a backend =
  implement the interface AND add a `case`.**
- Always go through the `Storage` methods (`GetPlan`/`PutPlan`/`GetLogs`/`GetGitBundle`…),
  **never call `Backend` directly**: `Storage` wraps every read/write with the per-namespace
  `EncryptionManager` (Put encrypts, Get decrypts). Bypassing it breaks the encryption contract.
- Object keys are built centrally (`computeLogsKey`/`computePlanKey`/`computeGitBundleKey`) —
  `layers/ns/layer/run/attempt/<file>`, `repositories/ns/repo/branch/rev.gitbundle`. Don't
  hand-build keys; keep the format constants and the client in sync.
- The API is reached over HTTP through `datastore/client`, authorized by Kubernetes
  ServiceAccount token (audience `burrito`). A change to a handler in `api/` must be mirrored
  in `client/` (and vice-versa) — they are one contract.

## Validate

`make test` (see `storage/storage_test.go`, `api/api_test.go`), `make vet`,
`golangci-lint run ./...`.
