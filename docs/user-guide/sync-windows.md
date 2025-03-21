# Sync Windows

Sync windows are a way to control when Burrito can run `apply` operations on Terraform layers. This is useful to prevent changes during specific timeframes, like business hours or maintenance windows. A sync window is defined by a kind (`allow` or `deny`), a schedule in cron format, a duration and a selector for layers in which wildcard are supported. Sync window can be defined at the repository level or at global level (in the Burrito configuration). The sync window can be applied to `plan`, `apply` or both actions.

## Use Cases

- Blocking all Burrito operations outside of business hours to reduce cloud costs.
- Preventing Burrito to apply unwanted changes outside of business hours, while keeping drift detection enabled.
- Allowing only `apply` operations during specific maintenance windows to ensure that changes are applied at a specific time.

## Spec & Example

| Field                    | Type   | Description                                                                                         |
| ------------------------ | ------ | --------------------------------------------------------------------------------------------------- |
| `syncWindows`            | Array  | The list of sync windows.                                                                           |
| `syncWindows[].kind`     | String | The kind of the sync window, either `allow` or `deny`.                                              |
| `syncWindows[].schedule` | String | The schedule of the sync window in cron format.                                                     |
| `syncWindows[].duration` | String | The duration of the sync window.                                                                    |
| `syncWindows[].layers`   | Array  | The list of layers to which the sync window applies (supports wildcards).                           |
| `syncWindows[].actions`  | Array  | List of actions that are affected by the sync window. `["plan"]`, `["apply"]` or `["plan","apply"]` |

The following example shows how to define sync windows in a Terraform repository, it is purely to demonstrate the syntax and is not representative of a real-world use case.

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformRepository
metadata:
  name: my-repository
  namespace: burrito-project
spec:
  repository:
    url: https://github.com/padok-team/burrito-examples.git
  terraform:
    enabled: true
  syncWindows:
    - kind: allow
        schedule: "0 8 * * *"
        duration: "12h"
        layers:
          - "layer1"
          - "layer2"
        actions:
          - "plan"
          - "apply"
    - kind: deny
        schedule: "30 1 * * *"
        duration: "30m"
        layers:
          - "layer*"
        actions:
          - "apply"
```

## Behavior

Sync Windows work as follows:

- If no sync window is defined for a layer, the layer is always allowed to be applied.
- If a deny sync window is defined for a layer, the layer is not allowed to be applied during the sync window.
- If an allow sync window is defined for a layer, the layer is only allowed to be applied during the sync window.
- If multiple sync windows are defined for a layer and they overlap, the deny sync window takes precedence over the allow sync window.

The sync window will apply only for the actions defined in the `actions` field. If the `actions` field is not defined, the sync window will not apply to any action.

## Global Sync Windows

Default sync windows are defined in the Burrito configuration and apply to all Burrito reconciliation runs. They are useful to define sync windows that apply to all layers.
The default sync windows are defined in the `burrito.controller.defaultSyncWindows` field of the Burrito configuration.
If using helm, you can define the default sync windows in the [values file](https://github.com/padok-team/burrito/blob/main/deploy/charts/burrito/values.yaml).

```yaml
config:
  burrito:
    controller:
      # -- Default sync windows for layer reconciliation
      defaultSyncWindows:
        - kind: allow
          schedule: "0 8 * * *"
          duration: "12h"
          layers:
            - "layer1"
            - "layer2"
          actions:
            - "plan"
            - "apply"
        - kind: deny
          schedule: "30 1 * * *"
          duration: "30m"
          layers:
            - "layer*"
          actions:
            - "apply"
```
