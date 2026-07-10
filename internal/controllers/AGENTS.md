# Scope: Kubernetes Controllers (`internal/controllers/`)

Go core engine, built on `controller-runtime` / Kubebuilder. One reconciler package per
custom resource: `terraformlayer/`, `terraformrun/`, `terraformrepository/`,
`terraformpullrequest/`. `metrics/` holds Prometheus instrumentation. `manager.go` wires them.

## Rules & Gotchas

- Follow standard `controller-runtime` reconciliation patterns and the Kubebuilder layout.
- Reconcilers must be **idempotent** — they can run on the same object repeatedly.
- Never block the reconcile loop. For retries/transient failures, return a requeue
  (`ctrl.Result{RequeueAfter: ...}`), do not sleep.
- **Never `panic()`.** Return an error or a requeue.
- Always check errors explicitly — no `_ = err`.
- Use structured logging via the `logr.Logger` carried in `ctx` (`log.FromContext(ctx)`).
- Propagate `ctx` and set timeouts on every external call (GitHub, GitLab, Terraform).
- CRD shapes live in `api/v1alpha1`; never edit generated deepcopy code. After API changes run `make manifests && make generate`.
