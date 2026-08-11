# Burrito Deployment Guide

This guide covers deploying Burrito in a Kubernetes cluster.

## Prerequisites

- Kubernetes cluster (v1.25+)
- `kubectl` configured with cluster access
- Helm v3.8+ (for Helm-based installation)
- Container registry access (for custom images)

## Installation Methods

### Helm (Recommended)

```bash
# Add the Burrito Helm repository
helm repo add burrito https://charts.burrito.tf
helm repo update

# Install Burrito
helm install burrito burrito/burrito \
  --namespace burrito-system \
  --create-namespace \
  --values values.yaml
```

### Static Manifests

For environments without Helm, use the static manifests:

```bash
kubectl apply -f https://github.com/padok-team/burrito/releases/latest/download/install.yaml
```

## Configuration

### Required: Git Authentication

Burrito needs access to your Git repositories. Create a secret with your credentials:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: git-credentials
  namespace: burrito-system
type: Opaque
stringData:
  token: "ghp_xxxxxxxxxxxx"  # GitHub personal access token
```

Reference this secret in your `TerraformRepository` resources.

For GitHub App authentication, see [Git Authentication](docs/operator-manual/git-authentication.md).

### Required: Datastore Backend

Configure cloud storage for plan artifacts. Supported backends:

- **Azure Blob Storage**
- **Google Cloud Storage (GCS)**
- **S3-compatible storage** (AWS S3, MinIO, etc.)

Example for S3:

```yaml
datastore:
  type: s3
  s3:
    bucket: burrito-artifacts
    region: us-east-1
    endpoint: ""  # Leave empty for AWS S3
    accessKeyId: "AKIAxxx"
    secretAccessKey: "xxx"
```

### Namespace Configuration

By default, Burrito watches all namespaces. To limit to specific namespaces:

```yaml
config:
  watchedNamespaces:
    - team-a
    - team-b
```

Or watch only its own namespace:

```yaml
config:
  watchedNamespaces: []
```

### Runner Configuration

The Terraform runner can be customized:

```yaml
runner:
  image: ghcr.io/padok-team/burrito-runner:v0.13.0
  resources:
    requests:
      cpu: 500m
      memory: 512Mi
    limits:
      cpu: 2
      memory: 2Gi
  maxConcurrentReconciles: 5
  serviceAccount:
    create: true
    annotations: {}
```

### Webhook Configuration

To enable PR/MR workflows, configure the webhook receiver:

```yaml
webhook:
  enabled: true
  ingress:
    enabled: true
    host: burrito-webhook.example.com
    tls:
      enabled: true
      secretName: burrito-webhook-tls
```

Then add the webhook URL to your Git provider (GitHub/GitLab).

## Helm Values Reference

| Parameter                        | Description                                | Default                       |
|----------------------------------|--------------------------------------------|-------------------------------|
| `image.repository`               | Controller image                           | `ghcr.io/padok-team/burrito`  |
| `image.tag`                      | Image tag                                  | Chart `appVersion`            |
| `runner.image`                   | Runner image                               | `ghcr.io/padok-team/burrito`  |
| `runner.maxConcurrentReconciles` | Max concurrent runner pods                 | `5`                           |
| `datastore.type`                 | Storage backend (`s3`, `gcs`, `azure`)     | Required                      |
| `webhook.enabled`                | Enable webhook receiver                    | `false`                       |
| `config.watchedNamespaces`       | Namespaces to watch (empty = all)          | `[]`                          |
| `config.logLevel`                | Log level (`debug`, `info`, `warn`, `error`) | `info`                      |

Full reference: see `deploy/charts/burrito/values.yaml`.

## Upgrade

```bash
helm repo update
helm upgrade burrito burrito/burrito \
  --namespace burrito-system \
  --values values.yaml
```

Check the [migration guides](docs/migration-guides/) for breaking changes between versions.

## Uninstall

```bash
helm uninstall burrito --namespace burrito-system
```

> **Note:** CRDs are not deleted automatically. Remove them manually if needed:
> ```bash
> kubectl delete crd terraformlayers.config.burrito.tf
> kubectl delete crd terraformrepositories.config.burrito.tf
> ```

## Multi-Tenant Deployment

Burrito supports multi-tenant deployments. See [Multi-Tenant Architecture](docs/operator-manual/multi-tenant-architecture.md) for details on:

- Namespace-scoped controllers
- Per-tenant runner configurations
- Resource isolation strategies

## Monitoring

Burrito exposes Prometheus metrics at `:8080/metrics`. See [Metrics](docs/operator-manual/metrics.md) for available metrics and alerting recommendations.

## Production Checklist

- [ ] Git authentication configured (not using plaintext tokens in values)
- [ ] Datastore backend provisioned with proper IAM
- [ ] Webhook receiver behind TLS with ingress
- [ ] Resource limits set for controller and runner
- [ ] Metrics endpoint scraped by Prometheus
- [ ] Backup strategy for datastore artifacts
- [ ] Reviewed [advanced configuration](docs/operator-manual/advanced-configuration.md)
