# Usage <!-- omit in toc -->

- [User guide](#user-guide)
  - [Override the runner pod spec](#override-the-runner-pod-spec)
  - [Configure the TerraformLayer to use private modules' repositories](#configure-the-terraformlayer-to-use-private-modules-repositories)
  - [Choose your remediation strategy](#choose-your-remediation-strategy)
  - [Choose your terraform version](#choose-your-terraform-version)
  - [Use Terragrunt](#use-terragrunt)
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
    kind: TerraformRepository
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
  overridePodSpec:
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
    kind: TerraformRepository
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

### Configure the TerraformLayer to use private modules' repositories

If your stack use Terraform modules that are hosted on private repositories, you can configure the TerraformLayer to use a specific SSH key to clone the repository and a specific SSH known hosts file to verify the host.

To do so, you'll first need to create a **Secret** containing the SSH key and, if necessary, edit the `burrito-ssh-known-hosts` **ConfigMap** to add your known host key:

```yaml
apiVersion: v1
data:
  key: <YOUR_PRIVATE_SSH_KEY_BASE64_ENCODED>
immutable: false
kind: Secret
metadata:
  name: git-private-key–modules
  namespace: burrito
type: Opaque
---
apiVersion: v1
data:
  known_hosts: |-
    bitbucket.org ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAubiN81eDcafrgMeLzaFPsw2kNvEcqTKl/VqLat/MaB33pZy0y3rJZtnqwR2qOOvbwKZYKiEO1O6VqNEBxKvJJelCq0dTXWT5pbO2gDXC6h6QDXCaHo6pOHGPUy+YBaGQRGuSusMEASYiWunYN0vCAI8QaXnWMXNMdFP3jHAJH0eDsoiGnLPBlBp4TNm6rYI74nMzgz3B9IikW4WVK+dc8KZJZWYjAuORU3jc1c/NPskD2ASinf8v3xnfXeukU0sJ5N6m5E8VLjObPEO+mN2t/FZTMZLiFqPWc/ALSqnMnnhwrNi2rbfg/rd/IpL8Le3pSBne8+seeFVBoGqzHM9yXw==
    github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCj7ndNxQowgcQnjshcLrqPEiiphnt+VTTvDP6mHBL9j1aNUkY4Ue1gvwnGLVlOhGeYrnZaMgRK6+PKCUXaDbC7qtbW8gIkhL7aGCsOr/C56SJMy/BCZfxd1nWzAOxSDPgVsmerOBYfNqltV9/hWCqBywINIR+5dIg6JTJ72pcEpEjcYgXkE2YEFXV1JHnsKgbLWNlhScqb2UmyRkQyytRLtL+38TGxkxCflmO+5Z8CSSNY7GidjMIZ7Q4zMjA2n1nGrlTDkzwDCsw+wqFPGQA179cnfGWOWRVruj16z6XyvxvjJwbz0wQZ75XK5tKSb7FNyeIEs4TT4jk+S4dhPeAUC5y+bDYirYgM4GC7uEnztnZyaVWQ7B381AK4Qdrwt51ZqExKbQpTUNn+EjqoTwvqNj4kqx5QUCI0ThS/YkOxJCXmPUWZbhjpCg56i+2aB6CmK2JGhn57K5mj0MNdBXA4/WnwH6XoPWJzK5Nyu2zB3nAZp+S5hpQs+p1vN1/wsjk=
    gitlab.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFSMqzJeV9rUzU4kWitGjeR4PWSa29SPqJ1fVkhtj3Hw9xjLVXVYrU9QlYWrOLXBpQ6KWjbjTDTdDkoohFzgbEY=
    gitlab.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAfuCHKVTjquxvt6CM6tdG4SLp1Btn/nOeHHE5UOzRdf
    gitlab.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCsj2bNKTBSpIYDEGk9KxsGh3mySTRgMtXL583qmBpzeQ+jqCMRgBqB98u3z++J1sKlXHWfM9dyhSevkMwSbhoR8XIq/U0tCNyokEi/ueaBMCvbcTHhO7FcwzY92WK4Yt0aGROY5qX2UKSeOvuP4D6TPqKF1onrSzH9bx9XUf2lEdWT/ia1NEKjunUqu1xOB/StKDHMoX4/OKyIzuS0q/T1zOATthvasJFoPrAjkohTyaDUz2LN5JoH839hViyEG82yB+MjcFV5MU3N1l1QL3cVUCh93xSaua1N85qivl+siMkPGbO5xR/En4iEY6K2XPASUEMaieWVNTRCtJ4S8H+9
    github.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=
    github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
    git.yourcompany.ai <HOST_KEY>
kind: ConfigMap
metadata:
  labels:
    app.kubernetes.io/instance: in-cluster-burrito
    app.kubernetes.io/name: burrito-ssh-known-hosts
    app.kubernetes.io/part-of: burrito
  name: burrito-ssh-known-hosts
  namespace: burrito
```

**Hint**: You can find your host's key in your own `~/.ssh/known_hosts` file.

Once the **Secret** created and the **ConfigMap** edited, you can update your **TerraformLayer** to use the new SSH key:

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformLayer
metadata:
  name: terragrunt-private-module-ssh
spec:
  terraform:
    version: "1.3.1"
    terragrunt:
      enabled: true
      version: "0.45.4"
  remediationStrategy: autoApply
  path: "terragrunt/random-pets-private-module-ssh/test"
  branch: main
  repository:
    name: burrito
    namespace: burrito
  overrideRunnerSpec:
    env:
    - name: GIT_SSH_COMMAND
      value: ssh -i /home/burrito/.ssh/key
    volumes:
    - name: private-key
      secret:
        secretName: private-key-ssh-module
    volumeMounts:
    - name: private-key
      mountPath: /home/burrito/.ssh/key
      subPath: key
      readOnly: true
```

As you can see, we added a new `overrideRunnerSpec` field to the `TerraformLayer` spec. This field allows you to override the default runner pod spec.
In this case, we added a new volume and a new environment variable to the runner pod spec:
- The volume is a secret volume that contains the SSH key we created earlier
- The environment variable is used to tell git to use the SSH key we added to the runner pod

### Choose your remediation strategy

Currently, 2 remediation strategies are handled.

|  Strategy   |                               Effect                                |
| :---------: | :-----------------------------------------------------------------: |
|    `dry`    | The operator will only run the `plan`. This is the default strategy |
| `autoApply` |        If a `plan` is not up to date, it will run an `apply`        |

As for the [runner spec override](#override-the-runner-pod-spec), you can specify a `spec.remediationStrategy` either on the `TerraformRepository` or the `TerraformLayer`.

The configuration of the `TerraformLayer` will take precedence.

> :warning: This operator is still experimental. Use `spec.remediationStrategy: "autoApply"` at your own risk.

### Choose your terraform version

Both `TerraformRepository` and `TerraformLayer` expose a `spec.terrafrom.version` map field.

If the field is specified for a given `TerraformRepository` it will be applied by default to all `TerraformLayer` linked to it.

If the field is specified for a given `TerraformLayer` it will take precedence over the `TerraformRepository` configuration.

### Use Terragrunt

You can specify usage of terragrunt as follow:

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformLayer
metadata:
  name: random-pets-terragrunt
spec:
  terraform:
    version: "1.3.1"
    terragrunt:
      enabled: true
      version: "0.44.5"
  remediationStrategy: dry
  path: "internal/e2e/testdata/terragrunt/random-pets/prod"
  branch: "feat/handle-terragrunt"
  repository:
    kind: TerraformRepository
    name: burrito
    namespace: burrito
```

> This configuration can be specified at the `TerraformRepository` level to be enabled by default in each of its layers.

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

|            Environment variable                      |                                       Description                                            |             Default              |
| :--------------------------------------------------: | :------------------------------------------------------------------------------------------: | :------------------------------: |
|         `BURRITO_CONTROLLER_TYPES`                   |                                list of controllers to start                                  |        `layer,repository`        |
| `BURRITO_CONTROLLER_TIMERS_DRIFTDETECTION`           |                         period between two plans for drift detection                         |              `20m`               |
|     `BURRITO_CONTROLLER_TIMERS_ONERROR`              |           period between two runners launch when an error occurred in the controllers        |               `1m`               |
|   `BURRITO_CONTROLLER_TIMERS_WAITACTION`             |                   period between two runners launch when a layer is locked                   |               `1m`               |
|   `BURRITO_CONTROLLER_TIMERS_FAILUREGRACEPERIOD`     |           initial time before retry, goes exponential function of number failure             |              `15s`               |
| `BURRITO_CONTROLLER_LEADERELECTION_ENABLED`          |                           whether leader election is enabled or not                          |              `true`              |
|   `BURRITO_CONTROLLER_LEADERELECTION_ID`             |                               lease id used for leader election                              | `6d185457.terraform.padok.cloud` |
| `BURRITO_CONTROLLER_HEALTHPROBEBINDADDRESS`          |             address to bind the health probe server embedded in the controllers              |             `:8081`              |
|   `BURRITO_CONTROLLER_METRICSBINDADDRESS`            |               address to bind the metrics server embedded in the controllers                 |             `:8080`              |
| `BURRITO_CONTROLLER_KUBERNETESWEBHOOKPORT`           |             port used by the validating webhook server embedded in the controllers           |              `9443`              |

#### Server's configuration

| Environment variable  |        Description         | Default |
| :-------------------: | :------------------------: | :-----: |
| `BURRITO_SERVER_ADDR` | addr the server listens on | `:8080` |

For webhook configuration see [Setup a git webhook](#setup-a-git-webhook).

#### Runners' configuration

Currently, runners' configuration is not exposed.
