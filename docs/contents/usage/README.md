# Usage <!-- omit in toc -->

- [User guide](#user-guide)
  - [Override the runner pod spec](#override-the-runner-pod-spec)
  - [Choose your remediation strategy](#choose-your-remediation-strategy)
- [Operator guide](#operator-guide)
  - [Setup a git webhook](#setup-a-git-webhook)
  - [Configuration](#configuration)
    - [Controllers' configuration](#controllers-configuration)
    - [Server's configuration](#servers-configuration)
    - [Runners' configuration](#runners-configuration)

## User guide

### Override the runner pod spec

Both `TerraformRepository` and `TerraformLayer` expose a `spec.overrideRunnerSpec` map field.

If the field is specified for a given `TerraformRepository` it will be applied by default to all `TerraformLayer` linked to it.

If the field is specified for a given `TerraformLayer` it will take precedence over the `TerraformRepository` configuration.

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
  terraformVersion: "1.3.1"
  path: "internal/e2e/testdata/random-pets"
  branch: "main"
  repository:
    kind: TerraformRepository
    name: burrito
    namespace: burrito
```

But in the following case, no `tolerations` or `nodeSelector` will be used for the runner pods:

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformRepository
metadata:
  name: burrito
  namespace: burrito
spec:
  repository:
    url: https://github.com/padok-team/burrito
  overridePodSpec:
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
  terraformVersion: "1.3.1"
  path: "internal/e2e/testdata/random-pets"
  branch: "main"
  repository:
    kind: TerraformRepository
    name: burrito
    namespace: burrito
  overrideRunnerSpec:
    tolerations: []
    nodeSelector: {}
```

### Choose your remediation strategy

Currently, 2 remediation strategies are handled.

|  Strategy   |                               Effect                                |
| :---------: | :-----------------------------------------------------------------: |
|    `dry`    | The operator will only run the `plan`. This is the default strategy |
| `autoApply` |        If a `plan` is not up to date, it will run an `apply`        |

As for the [runner spec override](#override-the-runner-pod-spec), you can specify a `spec.remediationStrategy` either on the `TerraformRepository` or the `TerraformLayer`.

The configuration of the `TerraformLayer` will take precedence.

> :warning: This operator is still experimental. Use `spec.remediationStrategy: "autoApply"` at your own risk.

## Operator guide

### Setup a git webhook

Create a webhook (with a secret!) in the repository you want to receive events from.

Then create a secret:

```yaml
kind: Secret
metadata:
  name: burrito-webhook-secret
  namespace: burrito
type: Opaque
stringData:
  burrito-webhook-secret: <my-webhook-secret>
```

Add the webhook secret as an environment variable of the `burrito-server`. The variables depends on your git provider.

| Git provider |          Environment Variable          |
| :----------: | :------------------------------------: |
|    GitHub    | `BURRITO_SERVER_WEBHOOK_GITHUB_SECRET` |
|    GitLab    | `BURRITO_SERVER_WEBHOOK_GITLAB_SECRET` |

### Configuration

You can configure `burrito` with environment variables.

| Environment variable |         Description         |   Default    |
| :------------------: | :-------------------------: | :----------: |
| `BURRITO_REDIS_URL`  | the redis URL to connect to | `redis:6379` |

#### Controllers' configuration

|            Environment variable             |                              Description                               |             Default              |
| :-----------------------------------------: | :--------------------------------------------------------------------: | :------------------------------: |
|         `BURRITO_CONTROLLER_TYPES`          |                      list of controllers to start                      |        `layer,repository`        |
| `BURRITO_CONTROLLER_TIMERS_DRIFTDETECTION`  |              period between two plans for drift detection              |              `20m`               |
|     `BURRITO_CONTROLLER_TIMERS_ONERROR`     |        period between two runners launch when an error occured         |               `1m`               |
|   `BURRITO_CONTROLLER_TIMERS_WAITACTION`    |        period between two runners launch when a layer is locked        |               `1m`               |
| `BURRITO_CONTROLLER_LEADERELECTION_ENABLED` |               whether leader election is enabled or not                |              `true`              |
|   `BURRITO_CONTROLLER_LEADERELECTION_ID`    |                   lease id used for leader election                    | `6d185457.terraform.padok.cloud` |
| `BURRITO_CONTROLLER_HEALTHPROBEBINDADDRESS` |  address to bind the health probe server embedded in the controllers   |             `:8081`              |
|   `BURRITO_CONTROLLER_METRICSBINDADDRESS`   |     address to bind the metrics server embedded in the controllers     |             `:8080`              |
| `BURRITO_CONTROLLER_KUBERNETESWEBHOOKPORT`  | port used by the validating webhook server embedded in the controllers |              `9443`              |

#### Server's configuration

| Environment variable  |        Description         | Default |
| :-------------------: | :------------------------: | :-----: |
| `BURRITO_SERVER_ADDR` | addr the server listens on | `:8080` |

For webhook configuration see [Setup a git webhook](#setup-a-git-webhook).

#### Runners' configuration

Currently, runners' configuration is not exposed.
