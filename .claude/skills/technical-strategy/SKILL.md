---
name: technical-strategy
description: Turn product requirements into a repo-scoped technical implementation strategy for Burrito. Discovers the affected parts of the codebase, dispatches a sub-agent per relevant scope to investigate and plan, synthesizes their reports into one dependency-ordered plan, writes it to ./plans/, and opens it as a PR for review. Use this whenever someone hands over a feature ask, a ticket, or product requirements and wants a technical strategy / implementation plan / "how would we build this in Burrito" — especially when the work plausibly spans more than one part of the repo (api, controllers, ui, helm, docs). Do NOT use for a single-file bugfix, a question about how existing code works, or a request to actually implement code.
---

# Technical Strategy

Turn product requirements into a reviewable technical implementation plan for this repo. You orchestrate; scope sub-agents investigate. You write no product code — the deliverable is a strategy document, committed and opened as a PR so the team reviews the *plan* before anyone implements.

## Why this shape

Burrito is split into clearly-marked scopes (each carries an `AGENTS.md` with its own rules). A change to one feature usually ripples across several — a new CRD field means `api/v1alpha1` types, then codegen, then controller wiring, then maybe UI and Helm. No single context holds all those rules well. So you fan out: each scope's sub-agent reads that scope's `AGENTS.md`, investigates only its own code, and reports back in a fixed shape. You then merge those reports into one dependency-ordered plan. Reviewing the strategy as a PR — before code exists — is cheaper than reviewing a wrong implementation.

## Workflow

### 1. Absorb the requirements — and gate on clarity

Read the requirements from whatever form they arrive in: inline text, a file path, a pasted ticket/issue. Restate them in your own words.

**Do not fan out on vague requirements.** A confident-but-wrong plan multiplied across 4 sub-agents is expensive and misleading. If the ask is ambiguous, thin, or you can't tell what "done" means, ask clarifying questions first and wait for answers. Only proceed once you could explain the feature to someone else without guessing.

### 2. Discover scopes

Find the scope boundaries dynamically — do not hardcode a list, it rots:

```bash
find . -name AGENTS.md -not -path './.claude/*'
```

The repo-root `AGENTS.md` is the overall map; each **nested** `AGENTS.md` marks one scope (e.g. `api/v1alpha1/`, `internal/controllers/`, `ui/`, `deploy/`). Read the root one for the monorepo map and build/verify commands.

**`AGENTS.md` files are the primary map, not the whole map.** A scope can be a directory *without* an `AGENTS.md` — the root map's "Monorepo Map" lists more parts (`cmd/`, `internal/` packages like `internal/server`, `docs/reference`, etc.) than have their own file. A sub-agent dispatched to such a scope simply has no `AGENTS.md` to read; that's fine. Use the root map plus the nested files together to build the candidate scope list, and don't assume "no `AGENTS.md`" means "not a scope."

### 3. Triage — which scopes are relevant

Decide which scopes the feature touches. Read each scope's `AGENTS.md` purpose line if you're unsure. **Bias toward inclusion**: a sub-agent that reports "not affected" is cheap; a missed scope means an incomplete plan. State your triage explicitly — affected *and* not-affected, each with a one-line reason.

**If no scope is relevant** (the requirements don't map to Burrito, or are still too vague after clarifying), stop. Say so plainly and ask for direction. Do not invent work or fan out anyway.

### 4. Fan out — one sub-agent per relevant scope

Dispatch `general-purpose` sub-agents **in parallel, in a single message** — one per relevant scope. Each is read-only: it investigates and plans, it writes no code.

Use this prompt for each (fill in the bracketed parts):

```
You are investigating one scope of the Burrito repo to plan an implementation. Read-only: do NOT write code.

Scope: <scope-path>
First, read <scope-path>/AGENTS.md and follow its rules. Investigate ONLY within this scope.

Product requirements:
<the absorbed requirements>

Return EXACTLY these 7 sections, in this order:

1. Verdict — affected or not affected, one-line why.
2. Changes — concrete files / types / functions to touch.
3. Steps — ordered implementation steps within this scope.
4. Depends on — what you need FROM other scopes (e.g. "new field `Foo` in api/v1alpha1"). Empty if none.
5. Triggers — what your change FORCES elsewhere (e.g. "requires `make manifests && make generate`", "controller must set the new status field"). Empty if none.
6. Risks — codegen sensitivity, envtest, migrations, breaking changes, gotchas.
7. Commit/PR slice — how to group these changes into commit(s), with a Conventional Commit message (feat/fix/chore/docs/…, scope from api|controller|ui|helm|ci|deps|docker).

Be concrete and specific to the actual code you find. If your investigation shows another scope is also involved, say so in section 5.
```

The **Depends on** and **Triggers** sections are the load-bearing part: they are the dependency edges you need in the next step. A sub-agent that returns freeform prose instead of these sections can't be merged — hold it to the format.

### 4b. Second pass — catch scopes triage missed

Read every report's **Depends on** and **Triggers**. If any names a scope you did *not* dispatch (a common case: `internal/server` must expose a new field for the UI, but it has no `AGENTS.md` so triage skipped it), dispatch a sub-agent for it now, using the same prompt. This is why triage bias toward inclusion isn't enough on its own — sub-agents surface real cross-scope edges that a from-the-requirements triage can't see. One extra round is plenty; stop once no new scopes appear.

### 5. Synthesize — one dependency-ordered plan

This is your core job. Merge the reports into a single global sequence that honors every dependency edge. The usual spine in this repo:

```
api/v1alpha1 types → make manifests && make generate → internal/controllers → internal/server → ui → deploy (helm) → docs
```

Use each report's **Depends on / Triggers** to order steps across scopes — a step can't come before the thing it depends on. Fold each scope's per-scope `Steps` into the right place in the global order.

### 6. Write the plan

Write to `./plans/<feature-slug>.md` (create `./plans/` if missing; `<feature-slug>` is a short kebab-case name for the feature).

Follow this template exactly — don't invent a structure. The plan should read like a deliberate, reviewable strategy, not a brain dump.

```markdown
# [TITLE OF THE STRATEGY]

## Objectives

*List 3–5 clear, measurable goals for this strategy.*

- Goal 1
- Goal 2

## Assumptions and Consequences

*List each technical or environmental assumption followed immediately by its direct consequence.*

- We assume [assumption]. Therefore, [consequence].
- We assume [assumption]. Therefore, [consequence].

## Technical Strategy

### [Step Name 1]

**How To:**
- Actionable instruction 1
- Actionable instruction 2

**DoD:**
- Criteria for completion
- Delivery gate: merged via PR with required approval(s) and a green CI build

---

### [Step Name 2]

**How To:**
- Actionable instruction 1

**DoD:**
- Criteria for completion
- Delivery gate: merged via PR with required approval(s) and a green CI build
```

**Guidelines:**

- Each `### Step` must be completable in half a day or less and map to a single sprint task.
- Each `### Step` must produce a concrete, reviewable deliverable: a merged Git artifact (chart, values, manifest, code, or documentation change) or a verified operational change. Do not write steps that only "define", "document", or "decide" something with no artifact to review — that content belongs in `## Objectives` or `## Assumptions and Consequences`. If a step is purely operational (no Git change, e.g. seeding a secret value), state its operational verification in the `**DoD:**` and note the deviation from the PR delivery gate.
- Each `**DoD:**` must include the delivery gate: the change is merged via PR with the required approval(s) and a green CI build.
- Add code examples directly inside the relevant step when they help clarify implementation.
- Test-Driven Development: begin with a testing strategy — interface tests first (API endpoints, UI interactions mirroring real user behavior), unit tests for complex business logic or critical utilities.
- Data-modeling first: start from robust domain modeling (CRD/API contracts, data structures) and expand outward.
- Incremental development: decompose into logical, independently testable phases.
- Path-specific: use exact file paths and follow the established package structure.
- Maintainability: follow existing patterns and conventions in the codebase.

**Mapping — where the fan-out reports land:**

- **`## Objectives`** — 3–5 measurable goals distilled from the requirements.
- **`## Assumptions and Consequences`** — one "We assume X. Therefore Y." line per load-bearing assumption. Product ambiguities a sub-agent flagged (an unclear field's meaning) and cross-cutting decisions (a shared contract, a chosen default) belong here — not buried in a step.
- **`## Technical Strategy`** — your ordered cross-scope sequence, one `### Step` per coherent unit of work. Put each scope's `Steps` into `**How To:**`, naming the Conventional Commit (`feat(api): …`) so the artifact is unambiguous, and express cross-scope ordering there too ("after the api step lands, …"). Put completion criteria + the verify commands the scope's `Triggers`/`Risks` require (`make manifests && make generate`, `make test`, `yarn --cwd ui build`) into `**DoD:**`.

### 7. Confirm, then commit and open the PR

Show the finished plan (or its path + summary) and **ask for a go-ahead before opening the PR.** Opening a PR notifies the team — don't fire it on a plan the user hasn't glanced at.

Once confirmed:

```bash
git switch -c plan/<feature-slug> main
git add plans/<feature-slug>.md
git commit -m "docs: add technical strategy for <feature>"
git push -u origin plan/<feature-slug>
gh pr create --title "Technical strategy: <feature>" --body "<short summary + scopes touched>"
```

Branch off `main`, not whatever is currently checked out. Report the PR URL. Done — this skill stops at the plan; implementation is a separate session.
