# Scope: Helm Chart (`deploy/`)

The Burrito Helm chart lives in `deploy/charts/burrito/` (templates under `templates/`).

## Rules

- Keep `values.yaml` the single source of defaults; expose new behavior through values rather
  than hardcoding it in templates. Document every new value.
- Templates must stay in sync with the CRDs in `api/v1alpha1` and the controller's expected
  config. If a chart change requires updated CRD manifests, regenerate them with `make manifests`
  (don't hand-edit generated CRD YAML).
- Bump `Chart.yaml` `version` (chart) / `appVersion` (image) appropriately when changing the chart.
- Validate before declaring done: `helm lint deploy/charts/burrito` and
  `helm template deploy/charts/burrito` render without errors.
- Use scope `helm` for commits touching this directory.
