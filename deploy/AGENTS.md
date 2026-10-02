# Scope: Helm Chart (`deploy/`)

The Burrito Helm chart lives in `deploy/charts/burrito/` (templates under `templates/`).

## Rules

- Keep `values.yaml` the single source of defaults; expose new behavior through values rather
  than hardcoding it in templates. Document every new value with a `# --` helm-docs comment.
- `deploy/charts/burrito/README.md` is regenerated automatically by a `PostToolUse` hook
  (`.claude/settings.json`) whenever `values.yaml` is edited; commit the README diff with it.
  Outside Claude Code, run `mise run helm-docs` by hand. CI (`check-helm-docs` in `.github/workflows/helm.yaml`) fails the build if the README drifts
  from `values.yaml`.
- Templates must stay in sync with the CRDs in `api/v1alpha1` and the controller's expected
  config. If a chart change requires updated CRD manifests, regenerate them with `make manifests`
  (don't hand-edit generated CRD YAML).
- Bump `Chart.yaml` `version` (chart) / `appVersion` (image) appropriately when changing the chart.
- Validate before declaring done: `helm lint deploy/charts/burrito` and
  `helm template deploy/charts/burrito` render without errors.
- Use scope `helm` for commits touching this directory.
