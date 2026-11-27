# Choose a remediation strategy

The remediation strategy is the way to tell Burrito how it should handle the remediation of drifts on your Terraform layers.

As for the [runner spec override](./override-runner.md), you can specify a `spec.remediationStrategy` either on the `TerraformRepository` or the `TerraformLayer`.

The configuration of the `TerraformLayer` will take precedence.

## `spec.remediationStrategy` API reference

|         Field         |  Type   |                    Default                    |                                             Effect                                             |
| :-------------------: | :-----: | :-------------------------------------------: | :--------------------------------------------------------------------------------------------: |
|      `autoApply`      | Boolean |                    `false`                    |                     If `true` when a `plan` shows drift, it will run an `apply`.                |
| `nonDestructiveApply` | Boolean |                    `false`                    | If `true`, Burrito will not run an `apply` when the latest `plan` would destroy any resources. |
| `onError.maxRetries`  | Integer | `5` or value defined in Burrito configuration |            How many times Burrito should retry a `plan`/`apply` when a runner fails.            |

!!! warning
    This operator is still experimental. Use `spec.remediationStrategy.autoApply: true` at your own risk.

!!! tip
    On critical layers, combine `autoApply: true` with `nonDestructiveApply: true` to keep drift
    remediation automated while leaving any `plan` that would delete resources for a human to apply.

## Example

With this example configuration, Burrito will create `apply` runs for this layer, with a maximum of 3 retries.

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformLayer
metadata:
  name: random-pets-terragrunt
spec:
  remediationStrategy:
    autoApply: true
    onError:
      maxRetries: 3
  # ... snipped ...
```
