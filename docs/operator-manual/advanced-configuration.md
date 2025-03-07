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
