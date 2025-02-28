# Sync Windows

Sync windows are a way to control when Burrito can run `apply` operations on Terraform layers. This is useful to prevent changes during specific timeframes, like business hours or maintenance windows. A sync window is defined by a kind (`allow` or `deny`), a schedule in cron format, a duration and a selector for layers in which wildcard are supported.

## Spec & Example

| Field                    | Type   | Description                                                               |
| ------------------------ | ------ | ------------------------------------------------------------------------- |
| `syncWindows`            | Array  | The list of sync windows.                                                 |
| `syncWindows[].kind`     | String | The kind of the sync window, either `allow` or `deny`.                    |
| `syncWindows[].schedule` | String | The schedule of the sync window in cron format.                           |
| `syncWindows[].duration` | String | The duration of the sync window.                                          |
| `syncWindows[].layers`   | Array  | The list of layers to which the sync window applies (supports wildcards). |

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
    - kind: deny
        schedule: "30 1 * * *"
        duration: "30m"
        layers:
          - "layer*"
```

## Behavior

Sync Windows work as follows:

- If no sync window is defined for a layer, the layer is always allowed to be applied.
- If a deny sync window is defined for a layer, the layer is not allowed to be applied during the sync window.
- If an allow sync window is defined for a layer, the layer is only allowed to be applied during the sync window.
- If multiple sync windows are defined for a layer and they overlap, the deny sync window takes precedence over the allow sync window.
