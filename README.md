# burrito <!-- omit in toc -->

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

<p align="center"><img src="./docs/assets/icon/burrito.png" width="200px" /></p>

**Burrito** is a TACoS (**T**erraform **A**utomation **Co**llaboration **S**oftware) implemented as Kubernetes Operator. 

- [Why does this exists?](#why-does-this-exists)
- [Installation](#installation)
- [Usage](#usage)
  - [Connect to a public repository](#connect-to-a-public-repository)
  - [Connect to a private repository using username/password (or access token) authentication](#connect-to-a-private-repository-using-usernamepassword-or-access-token-authentication)
  - [Connect to a private repository using SSH authentication](#connect-to-a-private-repository-using-ssh-authentication)
  - [Synchronize a terraform layer](#synchronize-a-terraform-layer)
  - [Setup a git webhook](#setup-a-git-webhook)
  - [Override the runner pod spec](#override-the-runner-pod-spec)
  - [Choose your remediation strategy](#choose-your-remediation-strategy)
- [Configuration](#configuration)
  - [Controllers' configuration](#controllers-configuration)
  - [Server's configuration](#servers-configuration)
  - [Runners' configuration](#runners-configuration)
- [How it works](#how-it-works)
- [License](#license)


## Why does this exists?

[`terraform`](https://www.terraform.io/) is a tremendous tool to manage your infrastructure in IaC.
But, it does not come up with an out-of the box solution for managing [state drift](https://developer.hashicorp.com/terraform/tutorials/state/resource-drift).

Also, writing a CI/CD pipeline for terraform can be painful and depends on the tool you are using.

Finally, currently, there is no easy way to navigate your terraform state to truly understand the modifications your state undergoes when running `terraform apply`.

`burrito` aims to tackle those issues by:
- Planning continuously your terraform code and run applies if needed
- Offering an out of the box PR/MR integration so you do not have to write CI/CD pipelines for terraform ever again (not implemented yet)
- Showing your state's modifications in a simple Web UI (not implemented UI)

## Installation

You can just run the following:

```bash
kubectl apply -f https://raw.githubusercontent.com/padok-team/burrito/main/manifests/install.yaml
```

## Usage

### Connect to a public repository

Create a `TerraformRepository` Kubernetes ressource which looks like:

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformRepository
metadata:
  name: burrito
  namespace: burrito
spec:
  repository:
    url: https://github.com/padok-team/burrito
```

### Connect to a private repository using username/password (or access token) authentication

Create a Kubernetes `Secret` which looks like:

```yaml
kind: Secret
metadata:
  name: burrito-repo
  namespace: burrito
type: Opaque
stringData:
  username: <my-username>
  password: <my-password | my-access-token>
```

Then, create a `TerraformRepository` Kubernetes resource:

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformRepository
metadata:
  name: burrito
  namespace: burrito
spec:
  repository:
    url: https://github.com/padok-team/burrito
    secretName: burrito-repo
```

### Connect to a private repository using SSH authentication

Create a Kubernetes `Secret` which looks like:

```yaml
kind: Secret
metadata:
  name: burrito-repo
  namespace: burrito
type: Opaque
stringData:
  sshPrivateKey: |
    -----BEGIN OPENSSH PRIVATE KEY-----
    ...
    -----END OPENSSH PRIVATE KEY-----
```

Add the public key as a deploy key of your repository.

Then, create a `TerraformRepository` Kubernetes resource:

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformRepository
metadata:
  name: burrito
  namespace: burrito
spec:
  repository:
    url: git@github.com:padok-team/burrito.git
    secretName: burrito-repo
```

### Synchronize a terraform layer

First, you need to create a `TerraformRepository`.

Then, create a `TerraformLayer` ressource which looks like:

```yaml
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
|    GitLab    | `BURRITO_SERVER_WEBHOOK_GITHUB_SECRET` |


### Override the runner pod spec

Both `TerraformRepository` and `TerraformLayer` expose a `spec.OverrideRunnerSpec` map field.

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
  overridePodSpec:
    imagePullSecrets:
    - name: ghcr-creds
    tolerations:
    - effect: NoScehdule
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
    - effect: NoScehdule
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

## Configuration

You can configure `burrito` with environment variables.

| Environment variable |         Description         |   Default    |
| :------------------: | :-------------------------: | :----------: |
| `BURRITO_REDIS_URL`  | the redis URL to connect to | `redis:6379` |

### Controllers' configuration

|            Environment variable             |                              Description                               |             Default              |
| :-----------------------------------------: | :--------------------------------------------------------------------: | :------------------------------: |
|         `BURRITO_CONTROLLER_TYPES`          |                      list of controllers to start                      |        `layer,repository`        |
| `BURRITO_CONTROLLER_TIMERS_DRIFTDETECTION`  |              period between two plans for drfit detection              |              `20m`               |
|     `BURRITO_CONTROLLER_TIMERS_ONERROR`     |        period between two runners launch when an error occured         |               `1m`               |
|   `BURRITO_CONTROLLER_TIMERS_WAITACTION`    |        period between two runners launch when a layer is locked        |               `1m`               |
| `BURRITO_CONTROLLER_LEADERELECTION_ENABLED` |               whether leader election is enabled or not                |              `true`              |
|   `BURRITO_CONTROLLER_LEADERELECTION_ID`    |                   lease id used for leader election                    | `6d185457.terraform.padok.cloud` |
| `BURRITO_CONTROLLER_HEALTHPROBEBINDADDRESS` |     address to bind the metrics server embedded in the controllers     |             `:8081`              |
|   `BURRITO_CONTROLLER_METRICSBINDADDRESS`   |     address to bind the metrics server embedded in the controllers     |             `:8080`              |
| `BURRITO_CONTROLLER_KUBERNETESWEBHOOKPORT`  | port used by the validating webhook server embedded in the controllers |              `9443`              |

### Server's configuration

| Environment variable  |        Description         | Default |
| :-------------------: | :------------------------: | :-----: |
| `BURRITO_SERVER_ADDR` | addr the server listens on | `:8080` |

For webhook configuration see [Setup a git webhook](#setup-a-git-webhook).

### Runners' configuration

Currently, runners' configuration is not exposed.

## How it works

See [Designer](docs/contents/design/README.md) for details ong how `burrito` works under the hood.

## License

© 2022 [Padok](https://www.padok.fr/).

Licensed under the [Apache License](https://www.apache.org/licenses/LICENSE-2.0), Version 2.0 ([LICENSE](./LICENSE))
