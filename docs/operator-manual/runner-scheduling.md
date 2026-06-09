# Fine-tuning the scheduling of Burrito pods

Burrito creates runner pods to execute plans and apply changes on your infrastructure, and it also runs shared components such as the controller, the datastore, and optionally Hermitcrab. Scheduling those workloads carefully can reduce inter-zone traffic, isolate noisy jobs, and keep costs under control.

## Limit the number of runner pods in parallel

By default, Burrito does not limit the number of runner pods that can run in parallel. This can lead to a high number of pods running at the same time, which can be costly or can overload your infrastructure.

It is possible to limit the number of runner pods that can run in parallel by setting the `BURRITO_CONTROLLER_MAXCONCURRENTRUNNERPODS` environment variable in the controller, or by setting the `config.burrito.controller.maxConcurrentRunnerPods` value in the [Helm chart values file](https://github.com/padok-team/burrito/blob/main/deploy/charts/burrito/values.yaml).

You can also set this value in the TerraformRepository CRD by setting the `spec.maxConcurrentRunnerPods` field.

If the value of this parameter is set to `0`, there is no limit to the number of runner pods that can run in parallel.

When Burrito creates a pod, if the setting is both set in the controller and in the TerraformRepository, the TerraformRepository value will take precedence.

## Spread shared components across failure domains

When you run Burrito across multiple availability zones, spread the shared components that serve every reconciliation path:

- `datastore`
- `hermitcrab` when provider caching is enabled
- the Burrito server or controllers when you run multiple replicas

The Helm chart already exposes `deployment.nodeSelector`, `deployment.tolerations`, `deployment.affinity`, and `deployment.topologySpreadConstraints` for these components. A practical baseline is:

- use `topologySpreadConstraints` to keep replicas distributed across zones or nodes
- use anti-affinity to avoid placing every replica on the same node
- keep at least one `datastore` and `hermitcrab` pod per zone when cross-zone traffic is expensive

For example, `hermitcrab.deployment.topologySpreadConstraints` and `hermitcrab.deployment.affinity` can be used together to keep the provider cache available close to runner pods while avoiding a single-node concentration.

### Prefer zone-local traffic with `trafficDistribution`

The Helm chart exposes `service.trafficDistribution` for the `datastore`, `hermitcrab`, and `server` components. This value maps to the Kubernetes Service spec field [`.spec.trafficDistribution`](https://kubernetes.io/docs/concepts/services-networking/service/#traffic-distribution). Setting this to `PreferClose` (or `Close` on Kubernetes 1.27+) instructs the Service to route traffic to endpoints in the same topological domain (zone, node, etc.) as the client, reducing cross-zone traffic latency and cost. This complements the pod-level scheduling controls above by shaping where established connections land.

Example:

```yaml
hermitcrab:
  service:
    trafficDistribution: PreferClose
datastore:
  service:
    trafficDistribution: PreferClose
server:
  service:
    trafficDistribution: PreferClose
```

## Isolate runner pods onto dedicated nodes

Runner pods are usually the burstiest Burrito workload because they download code, contact providers, and execute Terraform plans or applies. If you want to keep that activity away from shared services, target a dedicated node pool for runners.

The recommended approach is to configure `spec.overrideRunnerSpec` with scheduling fields such as:

- `nodeSelector`
- `tolerations`
- `affinity`

That lets you direct runners to nodes reserved for infrastructure jobs while keeping the control plane components on a different pool. The full override mechanism is documented in [Override the runner pod spec](../user-guide/override-runner.md).

A common pattern is:

- label a dedicated node pool for Burrito runners
- taint those nodes so only runner pods land there
- add matching `nodeSelector` and `tolerations` in `overrideRunnerSpec`
- optionally add pod anti-affinity if you want to avoid stacking too many heavy runs on the same node

## Choose the knob that matches the goal

Use different scheduling controls depending on what you are optimizing for:

- **Cost control:** lower `maxConcurrentRunnerPods`, then place runners on cheaper or autoscaled nodes.
- **Lower latency:** spread `datastore` and `hermitcrab` close to the zones where runners execute; set `trafficDistribution: PreferClose` on their Services.
- **Isolation:** separate runners from controllers, server, and datastore with dedicated node pools.
- **Resilience:** add topology spread constraints and anti-affinity for shared multi-replica components.

Start with a simple policy, observe runner throughput and inter-zone traffic, then tighten the scheduling rules only where they solve a concrete bottleneck.
