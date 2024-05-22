# Install burrito with Helm

## Requirements

- Installed [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) command-line tool
- Installed [helm](https://helm.sh/docs/intro/install/) command-line tool **(version v3.8.0 and further)**
- Have a [kubeconfig](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/) file (default location is `~/.kube/config`)

## 1. Install burrito

!!! info
    Our Helm chart is published in an [OCI-based registry](https://helm.sh/docs/topics/registries/) (ghcr.io). You must use Helm v3.8.0 or above.

```bash
helm install burrito oci://ghcr.io/padok-team/charts/burrito -n burrito-system --create-namespace
```

This will create a new namespace, `burrito-system`, where the burrito operator will live.

You can change the chart's version with any version available on [our Chart registry](https://github.com/padok-team/burrito/pkgs/container/charts%2Fburrito).

## 2. Override values

You can inspect the chart's values with Helm.

```bash
helm show values oci://ghcr.io/padok-team/charts/burrito
```

The chart's source code is available on [burrito GitHub repository](https://github.com/padok-team/burrito).

Here is an example of values file overriding some default values of burrito:

```yaml
tenants:
  # Example tenant with 1 service account having additional role bindings
  - namespace:
      create: true
      name: "burrito-project-1"
      labels: {}
      annotations: {}
    serviceAccounts:
      - name: runner-project-1
        additionalRoleBindings:
          - name: custom
            role:
              kind: ClusterRole
              name: custom-role
        annotations: {}
        labels: {}
  # Example tenant with multiple service accounts using GKE Workload Identity
  - namespace:
      create: true
      name: "burrito-project-1"
    serviceAccounts:
      - name: runner-frontend
        annotations:
          iam.gke.io/gcp-service-account: burrito-frontend@company-project.iam.gserviceaccount.com
      - name: runner-backend
        annotations:
          iam.gke.io/gcp-service-account: burrito-backend@company-project.iam.gserviceaccount.com
      - name: runner-network
        annotations:
          iam.gke.io/gcp-service-account: burrito-network@company-project.iam.gserviceaccount.com
```

!!! info
    Learn more about these values in [Advanced Configuration](../advanced-configuration.md) and [Multi-tenant architecture](../multi-tenant-architecture.md).
