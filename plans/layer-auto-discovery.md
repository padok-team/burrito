# Layer Auto-Discovery (MVP)

> Technical strategy for [issue #894](https://github.com/padok-team/burrito/issues/894).
> Lets Burrito scan a `TerraformRepository`'s default branch, detect Terraform/Terragrunt/OpenTofu
> layers, and create/garbage-collect the matching `TerraformLayer` resources automatically —
> replacing manual/templated layer declaration.

## Scope

**In scope (MVP):** default-branch scanning, heuristic detection refined by include/exclude globs,
an optional committed config-file override, create/update/GC of owned `TerraformLayer`s, opt-in via
a new `TerraformRepository` spec field, docs.

**Out of scope (follow-up issues):** per-branch discovery (plans on feature branches), a webhook
force-refresh endpoint (Michael's comment), and any UI surfacing (badge / repo-detail view). These
are called out where a sub-agent flagged the dependency edge.

## Objectives

- A `TerraformRepository` with `layerDiscovery.enabled: true` results in one `TerraformLayer`,
  automatically, per detected layer directory on its default branch — with zero manual `TerraformLayer` YAML.
- Auto-created layers are **owned** by their `TerraformRepository` (ownerReferences) and carry
  discovery labels, so they are garbage-collected when the repo is deleted **or** when their path
  disappears from the repo — and manual layers are provably never touched.
- Detection is layered and predictable: heuristic by default, narrowable with `includes`/`excludes`
  globs, overridable by an optional committed config file for monorepo / Terragrunt-stack cases.
- The feature is strictly opt-in (`enabled` defaults to `false`); existing repositories and manual
  layers behave exactly as before.
- Detection and reconciliation are covered by tests that prove correct create, path-removal GC, and
  **non-deletion of manual layers**.

## Assumptions and Consequences

- We assume discovery must run **independently of the branch-sync loop**. Therefore `reconcileDiscovery`
  is called directly from `Reconcile` (not inside the sync state handler), because the sync path early-returns
  at `len(layers) == 0` (`states.go:65`) — a fresh, layer-less repo would otherwise never bootstrap.
- We assume auto-created layer names are **deterministic** and layers are written only on a real diff.
  Therefore the desired-set builder is a pure function of the tree and the reconciler is idempotent —
  otherwise the manager's existing `Watches(&TerraformLayer{})` turns every create into a reconcile storm.
- We assume auto-created layers carry a **minimal spec** (`Path`, `Branch`, `Repository` ref only) and
  inherit engine/config/remediation defaults through the existing `common.go` getters
  (`GetTerraformEnabled`/`GetOpenTofuEnabled`/`GetTerragruntEnabled(repo, layer)`, etc.). Therefore no
  engine config is copied into discovered layers; a Terragrunt directory relies on repo-level
  `terragrunt.enabled` inheritance for MVP.
- We assume two GC mechanisms are both required. Therefore: `ownerReferences` handle cascade on
  **repository deletion**, and an **explicit `client.Delete`** on the two-label selector handles
  **path removal** (owner refs do not cover a path vanishing from the tree).
- We assume the GC delete path must match **both** discovery labels
  (`discovery.terraform.padok.cloud/managed-by=<repo>` **AND** `discovery.terraform.padok.cloud/auto-discovered=true`).
  Therefore manual layers — which carry neither — are structurally excluded from deletion; widening the
  selector to a single label is the one bug that could cascade-delete real infrastructure.
- We assume `enabled` needs no repo→layer inheritance. Therefore it is a plain `bool` with
  `+kubebuilder:default=false` (not the `*bool`/`chooseBool` pattern used for inherited flags), and no getter is added.
- We assume default-branch resolution is a **new** capability. Therefore it is derived from the
  `HEAD` symref already returned by the existing `remote.List` ls-remote call (zero extra network),
  rather than from existing layer refs (which are circular at bootstrap).
- We assume the file-tree source is the **git tree object** of the resolved default-branch commit,
  not the on-disk working directory (which `Bundle` mutates via checkout+reset). Therefore listing is
  deterministic and decoupled from sync state.
- We assume the committed override file has a fixed conventional name **`.burrito.yaml`** at the repo
  root. Therefore no `configPath` spec knob is added for MVP; precedence is **config file > globs > heuristic**.
- We assume RBAC is already sufficient. Therefore Helm needs no RBAC change — the aggregated controller
  ClusterRole already grants `create/delete/list/watch` on `terraformlayers`; the Helm scope is only a
  chart-version bump plus the regenerated CRD (which travels with the api commit).
- We assume the CRD change is additive and opt-in. Therefore no migration guide is needed; migration
  from manual layers is a subsection on the new user-guide page.

## Technical Strategy

The dependency spine is: **api types → (repository git primitives ∥ detection) → desired-set builder →
create/update wiring → GC → config-file override → helm + docs**. Steps 1 and 2 are independent and can
land in parallel. Every step is a single reviewable PR/commit.

---

### Step 1 — API: `layerDiscovery` spec + status on `TerraformRepository`

**How To:**
- In `api/v1alpha1/terraformrepository_types.go`, add the field and struct:
  ```go
  // TerraformRepositorySpec
  LayerDiscovery LayerDiscovery `json:"layerDiscovery,omitempty"`

  type LayerDiscovery struct {
      // +kubebuilder:default=false
      Enabled  bool     `json:"enabled,omitempty"`
      Includes []string `json:"includes,omitempty"`
      Excludes []string `json:"excludes,omitempty"`
  }
  ```
- Add a minimal status sub-struct (observability without bloat — do **not** list every discovered path):
  ```go
  // TerraformRepositoryStatus
  LayerDiscovery LayerDiscoveryStatus `json:"layerDiscovery,omitempty"`

  type LayerDiscoveryStatus struct {
      DiscoveredLayersCount int          `json:"discoveredLayersCount,omitempty"`
      LastDiscoveryTime     *metav1.Time `json:"lastDiscoveryTime,omitempty"`
  }
  ```
- Reuse the existing `Status.Conditions` for discovery error reporting (no new condition machinery in this scope).
- Do **not** touch `terraformlayer_types.go` — auto-discovery marking is metadata (labels + ownerRefs), set by the controller, not a spec field.
- Run `make manifests && make generate`; commit the regenerated `zz_generated.deepcopy.go`,
  `config/crd/bases/`, `manifests/`, **and** the chart CRD template
  `deploy/charts/burrito/templates/crds/config.terraform.padok.cloud_terraformrepositories.yaml` (rewritten by the Makefile) together.
- Commit: `feat(api): add layerDiscovery config and status to TerraformRepository`.

**DoD:**
- New fields present; `enabled` defaults to `false`; existing `TerraformRepository` CRs remain valid (no migration).
- `make manifests && make generate` produces a clean, committed diff (generated files never hand-edited).
- `make test` green (envtest installs the regenerated CRD).
- Delivery gate: merged via PR with required approval(s) and a green CI build.

---

### Step 2 — Repository: default-branch + tree-listing git primitives

*(Independent of Step 1 — can land in parallel.)*

**How To:**
- Extend the `types.GitProvider` interface in `internal/repository/types/types.go` with:
  ```go
  GetDefaultBranch() (string, error)
  ListFiles(ref string) ([]string, error)   // flat list of tracked paths at ref
  GetFile(ref, path string) ([]byte, error) // for the .burrito.yaml override (Step 7)
  ```
- Implement all three on `standard.GitProvider` (`providers/standard/repository.go`) — this one type
  backs github, gitlab, and standard (they differ only by injected `AuthMethod`):
  - `GetDefaultBranch`: read the `HEAD` symref from the existing `remote.List(&git.ListOptions{Auth:…})`
    call (no extra network). Fall back path noted only if symref absent.
  - `ListFiles`/`GetFile`: ensure `p.clone()`, resolve the ref's commit, walk/read its **git tree object**
    (`commit.Tree()`), not the working dir. Follow the existing `Fetch` + hard `Reset` + `getReferenceName`
    helpers — do **not** use `Repository.Pull()` (per scope AGENTS.md).
- Add matching stubs to `providers/mock/mock.go` returning fixture trees (e.g. dirs with `main.tf`,
  `terragrunt.hcl`) so controller envtest specs compile and run deterministically.
- Commit: `feat(controller): add default-branch and tree-listing git primitives for layer auto-discovery`
  (repo has no `repository` commit scope; `controller` is the closest per AGENTS.md).

**DoD:**
- Interface + `standard` impl + `mock` stubs land together (compiles).
- Unit test in `providers/standard/repository_test.go` against a local fixture repo asserts default-branch
  resolution and tree listing.
- `make test`, `make vet`, `golangci-lint run ./...` pass.
- Delivery gate: merged via PR with required approval(s) and a green CI build.

---

### Step 3 — Controller: detection heuristic as a pure, unit-tested function

**How To:**
- New file `internal/controllers/terraformrepository/discovery.go`.
- Implement `detectLayerDirs(files []string, includes, excludes []string) []string` — TDD, **pure**,
  no k8s/git wiring: a directory is a layer if it directly contains a `*.tf` file or a `terragrunt.hcl`;
  then apply `includes`/`excludes` globs (`path/filepath.Match`, or `github.com/gobwas/glob` if already vendored).
- Backend-block detection is intentionally **not** part of MVP presence-detection (it is OR'd with `.tf`
  presence and needs HCL parsing) — globs are the escape hatch for over-detecting shared modules.
- Write table-driven unit tests first (monorepo layouts, nested modules, terragrunt dirs, glob include/exclude).
- Commit: `feat(controller): add layer detection heuristic for auto-discovery`.

**DoD:**
- `detectLayerDirs` is pure (no side effects) and covered by table-driven unit tests including a monorepo
  fixture and glob narrowing.
- `make test` / `golangci-lint` pass.
- Delivery gate: merged via PR with required approval(s) and a green CI build.

---

### Step 4 — Controller: deterministic desired-set builder (naming, ownerRefs, labels)

**How To:**
- Add label constants in `internal/annotations/annotations.go` (following the
  `<component>.terraform.padok.cloud/<key>` convention):
  ```go
  DiscoveryManagedBy      string = "discovery.terraform.padok.cloud/managed-by"      // value: <repository-name>
  DiscoveryAutoDiscovered string = "discovery.terraform.padok.cloud/auto-discovered" // value: "true"
  ```
- In `discovery.go`, implement pure builders:
  - `discoveredLayerName(repoName, path string) string` — deterministic, DNS-label-safe (sanitize path,
    append a short hash suffix for uniqueness + length bound). **Not** `GenerateName`.
  - `buildDesiredLayers(repository, defaultBranch, dirs) []TerraformLayer` — one layer per dir; sets only
    `Path`, `Branch`, `Repository{Name,Namespace}`; sets both discovery labels; sets `OwnerReferences` via
    `controllerutil.SetControllerReference` (mirror `terraformpullrequest/layer.go`). Engine/config left empty.
- Unit-test both builders (stable names across runs; both labels present; owner ref set).
- Commit: `feat(controller): build desired auto-discovered TerraformLayer set`.

**DoD:**
- Builders are pure and deterministic; unit tests assert stable naming and that every desired layer carries
  both labels + the owner reference.
- `make test` / `golangci-lint` pass.
- Delivery gate: merged via PR with required approval(s) and a green CI build.

---

### Step 5 — Controller: `reconcileDiscovery` wiring (create/update, gated, independent of sync)

**How To:**
- In `discovery.go`, implement `func (r *Reconciler) reconcileDiscovery(ctx, repository) error`:
  gate on `repository.Spec.LayerDiscovery.Enabled`; resolve the provider
  (`repo.GetGitProviderFromRepository`); call `GetDefaultBranch` + `ListFiles`; `detectLayerDirs`;
  `buildDesiredLayers`; then list actual auto-discovered layers via the **two-label** selector and
  create-or-update by deterministic name (write only on real diff).
- Call `reconcileDiscovery` from `Reconcile` in `controller.go` **directly** (not inside the sync state
  handler), so a zero-layer repo bootstraps. Requeue on the existing repository-sync cadence
  (`timers.repositorySync`, default 5m — no new timer for MVP). Never `panic`; report errors into a
  `Status.Conditions` entry and update `Status.LayerDiscovery` (count + `LastDiscoveryTime`).
- Add `+kubebuilder:rbac` markers for `create;delete` on `terraformlayers` to `controller.go` as
  self-documentation (regenerates `config/rbac`; the aggregated ClusterRole already grants this).
- Extend the ginkgo/envtest suite (`controller_test.go`) with a mock-provider file tree: assert
  create-on-discovery and update-on-path-change.
- Commit: `feat(controller): reconcile auto-discovered TerraformLayers on default branch`.

**DoD:**
- Enabling `layerDiscovery` on a repo creates one `TerraformLayer` per detected dir (envtest, via mock provider).
- Discovery runs on a fresh layer-less repo (proves independence from the sync loop).
- Reconcile is idempotent: a second pass with an unchanged tree writes nothing (no update churn).
- `make test` (envtest) green.
- Delivery gate: merged via PR with required approval(s) and a green CI build.

---

### Step 6 — Controller: garbage-collection of removed paths

**How To:**
- In `reconcileDiscovery`, after create/update, compute actual-minus-desired (matched by the two-label
  selector) and `client.Delete` each — explicit path-removal GC.
- Rely on `ownerReferences` only for the repo-deletion cascade (do not conflate the two mechanisms).
- Add envtest specs: (a) a path removed from the tree deletes its auto-created layer; (b) a **manual**
  `TerraformLayer` (no discovery labels) in the same namespace/repo is **untouched** after a discovery pass
  that deletes others.
- Commit: `feat(controller): garbage-collect auto-discovered layers on path removal`.

**DoD:**
- Removing a directory from the scanned tree deletes exactly its auto-created layer (envtest).
- **A manual layer is provably never deleted by a discovery pass** (dedicated envtest assertion).
- `make test` green.
- Delivery gate: merged via PR with required approval(s) and a green CI build.

---

### Step 7 — Controller: optional committed config-file override *(droppable under time pressure)*

**How To:**
- Support an optional `.burrito.yaml` at repo root, read via `GetFile(defaultBranch, ".burrito.yaml")`.
  Define a small schema (explicit layer list and/or per-layer engine/path overrides for monorepo &
  Terragrunt-stack cases). Precedence: **config file > includes/excludes globs > heuristic**.
- When present, it refines/overrides the heuristic result before `buildDesiredLayers`.
- Unit-test parsing + precedence; envtest a repo whose `.burrito.yaml` overrides the heuristic result.
- Commit: `feat(controller): support .burrito.yaml override for layer auto-discovery`.

> This step delivers the "overrides" half of the chosen "heuristic + overrides" approach. It is sliced
> last so it can be deferred to a follow-up if the release is time-constrained without blocking Steps 1–6.

**DoD:**
- A repo with `.burrito.yaml` produces layers per the file, overriding the heuristic (unit + envtest).
- Absent file → behavior identical to Step 6 (heuristic + globs).
- `make test` green.
- Delivery gate: merged via PR with required approval(s) and a green CI build.

---

### Step 8 — Helm chart version bump

**How To:**
- Bump `version` in `deploy/charts/burrito/Chart.yaml` (the regenerated CRD template already arrived with
  the Step 1 api commit; **do not** hand-edit `templates/crds/*.yaml`).
- No RBAC change (aggregated ClusterRole already grants create/delete on `terraformlayers`). No new
  `values.yaml` knob (enablement is per-CR; MVP reuses `timers.repositorySync`).
- Validate `helm lint` / `helm template` render clean.
- Commit: `chore(helm): bump chart version for layerDiscovery CRD field`.

**DoD:**
- `helm lint` and `helm template deploy/charts/burrito` succeed; rendered CRD contains `layerDiscovery`.
- Release note: users with `global.crds.install: false` must apply the updated CRD out-of-band.
- Delivery gate: merged via PR with required approval(s) and a green CI build.

---

### Step 9 — Documentation

**How To:**
- Add `docs/user-guide/layer-auto-discovery.md` (follow `docs/user-guide/sync-windows.md` structure):
  intro & use cases; a `| Field | Type | Description |` spec table for `layerDiscovery`; a
  `TerraformRepository` YAML example; how heuristic detection works; refining with `includes`/`excludes`;
  the `.burrito.yaml` override + precedence; identifying auto-discovered layers
  (`kubectl get terraformlayer -l discovery.terraform.padok.cloud/auto-discovered=true`); behavior on
  manual edit/delete of an auto-created layer (owned → recreated); and a "Migrating from manual layers" subsection.
- Add `docs/examples/terraform-repository-auto-discovery.yaml` (mirror `terraform-repository.yaml`).
- Register the page under `nav > User Guide` in `mkdocs.yml`.
- Add an "Automatic Layer Discovery" section to `docs/operator-manual/repository-controller.md`
  (ownership, label, heuristic, GC behavior).
- Add a one-line pointer in `docs/getting-started.md` step 3.
- Do **not** resurrect the empty `docs/reference/` stubs; spec fields are documented inline (repo convention).
- Verify with `uv run mkdocs build` and `docs/.markdownlint.jsonc`.
- Commit: `docs(docs): document layer auto-discovery`.

**DoD:**
- Field names, label string, and behavior in docs match the merged api + controller implementation (write after Steps 1–7).
- `mkdocs build` succeeds with no broken links; nav entry, page, and example are all wired.
- Delivery gate: merged via PR with required approval(s) and a green CI build.

---

## Deferred to follow-up issues

- **Per-branch discovery** — discover layers on tracked branches (`status.branches`) so feature branches get
  plans (issue comment request). Requires branch iteration in `reconcileDiscovery`.
- **Webhook force-refresh endpoint** — an endpoint in `internal/webhook` to force a `TerraformRepository`
  re-scan on demand (issue comment request).
- **UI surfacing** — an "auto-discovered" badge on layers (needs an `IsAutoDiscovered` field on the
  `/layers` server DTO) and/or a repo-detail view showing `layerDiscovery` config/status (needs a new page,
  route, and expanded `/repositories` DTO). The UI otherwise renders auto-created layers unchanged today.
