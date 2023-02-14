# burrito

<p align="center"><img src="./assets/icon/burrito.png" width="200px" /></p>

**Burrito** is a TACoS (**T**erraform **A**utomation **Co**llaboration **S**oftware). 

## Installation

## Usage

### Create a TerraformRepository

Create a `TerraformRepository` ressource which looks like:

```
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformRepository
metadata:
  name: burrito
  namespace: burrito
spec:
  repository:
    url: https://github.com/padok-team/burrito
```

You can now declare layers inside this repository.

### Create a TerraformLayer

Create a `TerraformLayer` ressource which looks like:

```
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformLayer
metadata:
  name: random-pets
  namespace: burrito
spec:
  terraformVersion: "1.3.1"
  remediationStrategy: dry
  path: "internal/e2e/testdata/random-pets"
  branch: "main"
  repository:
    kind: TerraformRepository
    name: burrito
    namespace: burrito
```

## Configuration

You can configure `burrito` with environment variables.

### Controllers configuration

|            Environment variable             |                              Description                               |             Default              |
| :-----------------------------------------: | :--------------------------------------------------------------------: | :------------------------------: |
|         `BURRITO_CONTROLLER_TYPES`          |                      list of controllers to start                      |        `layer,repository`        |
| `BURRITO_CONTROLLER_TIMERS_DRIFTDETECTION`  |              period between two plans for drfit detection              |              `20m`               |
|     `BURRITO_CONTROLLER_TIMERS_ONERROR`     |        period between two runners launch when an error occured         |               `1m`               |
|   `BURRITO_CONTROLLER_TIMERS_WAITACTION`    |        period between two runners launch when a layer is locked        |               `1m`               |
| `BURRITO_CONTROLLER_LEADERELECTION_ENABLED` |               whether leader election is enabled or not                |              `true`              |
|   `BURRITO_CONTROLLER_LEADERELECTION_ID`    |                   lease id used for leader election                    | `6d185457.terraform.padok.cloud` |
| `BURRITO_CONTROLLER_HEALTHPROBEBINDADDRESS` |     address to bind the metrics server embedded in the controllers     |             `:8080`              |
|   `BURRITO_CONTROLLER_METRICSBINDADDRESS`   |     address to bind the metrics server embedded in the controllers     |             `:8081`              |
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
