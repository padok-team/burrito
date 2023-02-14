# burrito <!-- omit in toc -->

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

<p align="center"><img src="./docs/assets/icon/burrito.png" width="200px" /></p>

**Burrito** is a TACoS (**T**erraform **A**utomation **Co**llaboration **S**oftware). 

## Installation

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

### Controllers configuration

|            Environment variable             |                              Description                               |             Default              |
| :-----------------------------------------: | :--------------------------------------------------------------------: | :------------------------------: |
|         `BURRITO_CONTROLLER_TYPES`          |                      list of controllers to start                      |        `layer,repository`        |
| `BURRITO_CONTROLLER_TIMERS_DRIFTDETECTION`  |              period between two plans for drfit detection              |              `20m`               |
|     `BURRITO_CONTROLLER_TIMERS_ONERROR`     |        period between two runners launch when an error occured         |               `1m`               |
|   `BURRITO_CONTROLLER_TIMERS_WAITACTION`    |        period between two runners launch when a layer is locked        |               `1m`               |
| `BURRITO_CONTROLLER_LEADERELECTION_ENABLED` |               whether leader election is enabled or not                |              `true`              |
|   `BURRITO_CONTROLLER_LEADERELECTION_ID`    |                   lease id used for leader election                    | `6d185457.terraform.padok.cloud` |
| `BURRITO_CONTROLLER_HEALTHPROBEBINDADDRESS` |     address to bind the metrics server embedded in the controllers     |              `8080`              |
|   `BURRITO_CONTROLLER_METRICSBINDADDRESS`   |     address to bind the metrics server embedded in the controllers     |              `8081`              |
| `BURRITO_CONTROLLER_KUBERNETESWEBHOOKPORT`  | port used by the validating webhook server embedded in the controllers |              `9443`              |

### Server configuration

| Environment variable  |        Description         | Default |
| :-------------------: | :------------------------: | :-----: |
| `BURRITO_SERVER_PORT` | port the server listens on | `8080`  |

### Runner configuration

Currently, runners' configuration is not exposed.

## License

© 2022 [Padok](https://www.padok.fr/).

Licensed under the [Apache License](https://www.apache.org/licenses/LICENSE-2.0), Version 2.0 ([LICENSE](./LICENSE))
