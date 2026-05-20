# Advanced configuration

Here are some important configuration options that can be set to customize Burrito's behavior.
They can be set in the Helm chart [values](https://github.com/padok-team/burrito/blob/main/deploy/charts/burrito/values.yaml) or as environment variables.

## Controllers' configuration

|              Environment variable              |                                 Description                                 |              Default               |
| :--------------------------------------------: | :-------------------------------------------------------------------------: | :--------------------------------: |
|        `BURRITO_CONTROLLER_NAMESPACES`         |                list of namespaces to watch (comma-separated)                |          `burrito-system`          |
|           `BURRITO_CONTROLLER_TYPES`           |                        list of controllers to start                         | `layer,repository,run,pullrequest` |
|   `BURRITO_CONTROLLER_TIMERS_DRIFTDETECTION`   |                period between two plans for drift detection                 |               `20m`                |
|      `BURRITO_CONTROLLER_TIMERS_ONERROR`       | period between two runners launch when an error occurred in the controllers |                `1m`                |
|     `BURRITO_CONTROLLER_TIMERS_WAITACTION`     |          period between two runners launch when a layer is locked           |                `1m`                |
| `BURRITO_CONTROLLER_TIMERS_FAILUREGRACEPERIOD` |   initial time before retry, goes exponential function of number failure    |               `15s`                |
|    `BURRITO_CONTROLLER_TERRAFORMMAXRETRIES`    |   default number of retries for terraform runs (can be overriden in CRDs)   |                `5`                 |
|  `BURRITO_CONTROLLER_LEADERELECTION_ENABLED`   |                  whether leader election is enabled or not                  |               `true`               |
|     `BURRITO_CONTROLLER_LEADERELECTION_ID`     |                      lease id used for leader election                      |  `6d185457.terraform.padok.cloud`  |
|  `BURRITO_CONTROLLER_HEALTHPROBEBINDADDRESS`   |     address to bind the health probe server embedded in the controllers     |              `:8081`               |
|    `BURRITO_CONTROLLER_METRICSBINDADDRESS`     |       address to bind the metrics server embedded in the controllers        |              `:8080`               |
|   `BURRITO_CONTROLLER_KUBERNETESWEBHOOKPORT`   |   port used by the validating webhook server embedded in the controllers    |               `9443`               |
|   `BURRITO_CONTROLLER_MAXCONCURRENTRECONCILES` |    number of parallel resource reconciliation performed by the contoller    |                `0`                 |
|   `BURRITO_CONTROLLER_MAXCONCURRENTRUNNERPODS` | maximum number for pods that run in parallel to perform plan/apply (0=inf)  |                `0`                 |

## Server's configuration

| Environment variable  |          Description          | Default |
| :-------------------: | :---------------------------: | :-----: |
| `BURRITO_SERVER_ADDR` | address the server listens on | `:8080` |

!!! info
    For webhook configuration see [Setup a git webhook](./git-webhook.md).

## Runners' configuration

Currently, runners' configuration is not exposed.

!!! info
    You can override some of the runner's pod spec. See [override the runner pod spec](../user-guide/override-runner.md) documentation.

## Recommended environment variables for runner pods

Burrito's runner pods use [`tenv`](https://github.com/tofuutils/tenv) to download Terraform, Terragrunt or OpenTofu binaries on demand. By default, `tenv` reaches GitHub's public API to discover versions, which is subject to a low unauthenticated rate limit (about 60 requests per hour per source IP).

For production or high-scale deployments where many runner pods may launch in a short window, it is **strongly recommended** to set `TENV_GITHUB_TOKEN` on the runner pods so that `tenv` makes authenticated requests to GitHub and is subject to the much higher authenticated rate limit.

Create a GitHub token with no specific permissions (a fine-grained personal access token with no repository or account access is sufficient — `tenv` only reads public release metadata), store it in a Kubernetes `Secret`, and reference it from the `overrideRunnerSpec` on your `TerraformRepository` (or `TerraformLayer`):

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformRepository
metadata:
  name: my-repo
  namespace: burrito
spec:
  repository:
    url: https://github.com/<org>/<repo>
  terraform:
    enabled: true
  overrideRunnerSpec:
    env:
      - name: TENV_GITHUB_TOKEN
        valueFrom:
          secretKeyRef:
            name: tenv-github-token
            key: token
```

The same variable is already documented for local development setups in [Contributing → Advanced Settings](../contributing.md#advanced-settings); the snippet above is the production-shaped equivalent (Secret reference rather than an inline value).

!!! tip
    Apply the override on the `TerraformRepository` so that every linked `TerraformLayer` inherits it, instead of repeating the setting on each layer. See [override the runner pod spec](../user-guide/override-runner.md) for the merge rules.
