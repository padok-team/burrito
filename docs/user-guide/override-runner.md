# Override the runner pod spec

Both `TerraformRepository` and `TerraformLayer` expose a `spec.overrideRunnerSpec` map field.

If the field is specified for a given `TerraformRepository` it will be applied by default to all `TerraformLayer` linked to it.

If the field is specified for a given `TerraformLayer` it will take precedence over the `TerraformRepository` configuration.

Available overrides are:

|         Fields         |
| :--------------------: |
|   `ImagePullSecrets`   |
|        `Image`         |
|     `Tolerations`      |
|     `NodeSelector`     |
|  `ServiceAccountName`  |
|      `Resources`       |
|         `Env`          |
|       `EnvFrom`        |
|       `Volumes`        |
|     `VolumeMounts`     |
| `Metadata.Annotations` |
|   `Metadata.Labels`    |

For instance with the following configuration, all the runner pods will have the specifications described inside the `TerraformRepository`:

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformRepository
metadata:
  name: burrito
  namespace: burrito
spec:
  repository:
    url: https://github.com/padok-team/burrito
  terraform:
    enabled: true
  overrideRunnerSpec:
    imagePullSecrets:
    - name: ghcr-creds
    tolerations:
    - effect: NoSchedule
      key: burrito.io/production
      operator: Exists
    nodeSelector:
      production: "true"
    serviceAccountName: "production"
---
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformLayer
metadata:
  name: random-pets
  namespace: burrito
spec:
  terraform:
    version: "1.3.1"
  path: "internal/e2e/testdata/random-pets"
  branch: "main"
  repository:
    name: burrito
    namespace: burrito
```

In the following case, `tolerations` and `nodeSelector` will be merged:

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformRepository
metadata:
  name: burrito
  namespace: burrito
spec:
  repository:
    url: https://github.com/padok-team/burrito
  terraform:
    enabled: true
  overrideRunnerSpec:
    imagePullSecrets:
    - name: ghcr-creds
    tolerations:
    - effect: NoExecute
      key: burrito.io/production
      operator: Exists
    nodeSelector:
      production: "true"
    serviceAccountName: "production"
---
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformLayer
metadata:
  name: random-pets
  namespace: burrito
spec:
  terraform:
    version: "1.3.1"
  path: "internal/e2e/testdata/random-pets"
  branch: "main"
  repository:
    name: burrito
    namespace: burrito
  overrideRunnerSpec:
    tolerations:
    - effect: NoSchedule
      key: burrito.io/production
      operator: Exists
    nodeSelector: {}
```

Resulting in the following `podSpec`:

```yaml
tolerations:
- effect: NoSchedule
  key: burrito.io/production
  operator: Exists
nodeSelector:
  production: "true"
```
