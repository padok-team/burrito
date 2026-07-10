# Scope: Terraform Runner (`internal/runner/`)

The short-lived binary that actually runs Terraform/Terragrunt. Launched as a Kubernetes Job
per plan/apply — **not** a controller, there is no reconcile loop. `runner.go` is the
lifecycle (`Exec`: init clients → fetch code → install binaries → init → action),
`actions.go` the `plan`/`apply` logic, `tools/` the binary abstraction.

## Rules & Gotchas

- Entry is `Runner.Exec()` and runs **once** then exits. It uses `logrus` (not the
  controllers' `logr`/`ctx` logging) — match the surrounding style here.
- Never shell out to `terraform`/`terragrunt` directly: go through `tools.BaseExec`
  (`Init`/`Plan`/`Apply`/`Show`, `TenvName()`), which resolves versions via tenv. Add tool
  support under `tools/`.
- Source code is **not** cloned from the remote: the runner pulls a git bundle from the
  datastore (`Datastore.GetGitBundle`) and `git clone`s it locally.
- Artifacts flow through the datastore client keyed by `namespace/name/run/attempt/format`.
  `plan` writes formats `pretty`/`json`/`short`/`bin`; `apply` reuses the `bin` plan artifact
  (unless `GetApplyWithoutPlanArtifactEnabled`). Results are surfaced to the layer via
  `annotations.Add`, not by mutating status directly.
- `Action` is `plan` or `apply`; an unknown action signals a controller/runner version
  mismatch — keep the two in sync.

## Validate

`make test` (see `runner_test.go`), `make vet`, `golangci-lint run ./...`.
