# Install burrito with Helm

## Requirements

- Installed [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) command-line tool
- Installed [helm](https://helm.sh/docs/intro/install/) command-line tool **(version v3.8.0 and further)**
- Have access to a Kubernetes cluster

## 1. Basic installation

!!! info
    Our Helm chart is published in an [OCI-based registry](https://helm.sh/docs/topics/registries/) (ghcr.io). You must use Helm v3.8.0 or above.

```bash
helm install burrito oci://ghcr.io/padok-team/charts/burrito -n burrito-system --create-namespace
```

This will create a new namespace, `burrito-system`, where the burrito core components will live.

You can change the chart's version with any version available on [our Chart registry](https://github.com/padok-team/burrito/pkgs/container/charts%2Fburrito).

## 2. Burrito Helm configuration

The Burrito configuration is managed through Helm values files, which can be overridden at installation time.

You can find the default values of the Burrito Helm chart by running:

```bash
helm show values oci://ghcr.io/padok-team/charts/burrito
```

The [source code](https://github.com/padok-team/burrito/tree/main/deploy/charts/burrito) and [values file](https://github.com/padok-team/burrito/blob/main/deploy/charts/burrito/values.yaml) of the chart is available on [burrito GitHub repository](https://github.com/padok-team/burrito).


Here is an example of a simple burrito Helm values file that you can use to bootstrap your installation:

```yaml
config:
  burrito:
    controller:
      timers:
        driftDetection: 10m # run drift detection every 10 minutes
        onError: 10s # wait 10 seconds before retrying on error
        waitAction: 1m # wait 1 minute before retrying on locked layer
        failureGracePeriod: 30s # set a grace period of 30 seconds before retrying on failure (increases exponentially with the amount of failed retries)
    datastore:
      storage:
        mock: true # use a mock storage for the datastore (useful for testing, not recommended for production)
tenants:
  - namespace:
      create: true
      name: "burrito-project"
    serviceAccounts:
    - name: burrito-runner
      annotations:
        iam.gke.io/gcp-service-account: burrito@company-project.iam.gserviceaccount.com # example: use GKE Workload Identity to have access to GCP infrastructure
```

!!! info
    Learn more about these values in the chart's [README file](https://github.com/padok-team/burrito/tree/main/deploy/charts/burrito/README.md) and [Multi-tenant architecture](../operator-manual/multi-tenant-architecture.md).
