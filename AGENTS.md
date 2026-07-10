# AGENTS.md — Burrito

Monorepo for **Burrito**: a Kubernetes operator that automates Terraform (TACoS).
This file is the canonical guidance for AI agents. `CLAUDE.md` is a symlink to it.
Nested `AGENTS.md` files add directory-specific rules — read them when working in their scope.

## Working Principles

- **Think first.** State assumptions; if multiple interpretations exist, surface them instead of picking silently. If something is unclear, ask.
- **Simplicity.** Minimum code that solves the problem. No speculative features, abstractions, or config that wasn't requested.
- **Surgical changes.** Touch only what the task requires. Match existing style. Don't refactor or reformat unrelated code; flag dead code rather than deleting it.
- **Verify.** Turn tasks into checkable goals and run the relevant build/test/lint command before claiming done.

## Monorepo Map

- `api/v1alpha1/` — CRD Go types (see nested AGENTS.md). **Codegen-sensitive.**
- `internal/controllers/` — reconciliation logic, one package per resource (see nested AGENTS.md).
- `internal/repository/` — git provider access (see nested AGENTS.md). `internal/runner/` — the Terraform runner binary (see nested AGENTS.md). `internal/datastore/` — artifact storage service (see nested AGENTS.md). `internal/server/` — dashboard server + API (see nested AGENTS.md). `internal/webhook/` — VCS webhook receiver (see nested AGENTS.md).
- `ui/` — React/Vite/TS dashboard (see nested AGENTS.md).
- `deploy/charts/burrito/` — Helm chart (see nested AGENTS.md).
- `cmd/` — binary entrypoints. `hack/` — dev/build scripts. `manifests/` & `config/crd/bases/` — generated manifests. `testdata/` — fixtures. `docs/` — documentation.

## Never Touch (generated / vendored)

Do not read, edit, or use as reference:

- `*zz_generated.deepcopy.go` (produced by `controller-gen`).
- Lock files: `go.sum`, `ui/yarn.lock`.
- Generated manifests in `config/crd/bases/` and `manifests/`.

To change CRDs: edit `api/v1alpha1/*_types.go`, then run `make manifests` (and `make generate`).

## Build & Validation

- **Go:** `make build` · `make test` (spins up envtest + docker-compose — heavy) · `make vet`
- **Lint (Go):** `golangci-lint run ./...` — there is no `make` target; it runs in CI ([.github/workflows/ci.yaml](.github/workflows/ci.yaml)).
- **After API changes:** `make manifests && make generate`
- **UI:** `yarn --cwd ui lint` · `yarn --cwd ui build` · `yarn --cwd ui format-check`

## Go Style

- Always check errors explicitly. Never `_ = err` or silently ignored returns.
- No `panic()` in reconcilers — see `internal/controllers/AGENTS.md`.
- **Tests:** add new cases to the existing test suite for a package rather than creating a new suite from scratch. Controllers use BDD (ginkgo/gomega): each `controller_test.go` reads `Describe("When X …")` → `It("should Y …")` (run via `RunSpecs`) — add cases to that suite.

## Commits — Conventional Commits

Format: `<type>(<scope>): <description>`.

- **Types:** `feat`, `fix`, `chore`, `docs`, `test`, `refactor`.
- **Scopes** — suggested, not enforced. Prefer the closest fit (`controller`, `api`, `ui`, `helm`, `ci`, `deps`, `docker`); otherwise use another short, relevant scope or omit it.
- CI gates each commit with commitlint (`@commitlint/config-conventional`) — self-check a message with `npx commitlint`.
