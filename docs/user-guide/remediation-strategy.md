# Choose a remediation strategy

The remediation strategy is the way to tell Burrito how it should handle the remediation of drifts on your Terraform layers.

As for the [runner spec override](./override-runner.md), you can specify a `spec.remediationStrategy` either on the `TerraformRepository` or the `TerraformLayer`.

The configuration of the `TerraformLayer` will take precedence.

## `spec.remediationStrategy` API reference

|         Field         |  Type   |                    Default                    |                                           Effect                                            |
| :-------------------: | :-----: | :-------------------------------------------: | :-----------------------------------------------------------------------------------------: |
|      `autoApply`      | Boolean |                    `false`                    |                If `true` when a `plan` shows drift, it will run an `apply`.                 |
| `nonDestructiveApply` | Boolean |                    `false`                    | If `true`, `apply` runs are skipped when the latest `plan` includes resources to destroy.   |
| `onError.maxRetries`  | Integer | `5` or value defined in Burrito configuration |          How many times Burrito should retry a `plan`/`apply` when a runner fails.          |

!!! warning
    This operator is still experimental. Use `spec.remediationStrategy.autoApply: true` at your own risk.

!!! note
    `nonDestructiveApply` only applies when `autoApply` is enabled. If the latest plan summary includes resources to delete, Burrito will keep the layer in `ApplyNeeded` and wait for the next drift detection instead of creating an `apply` run.

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
    nonDestructiveApply: true
    onError:
      maxRetries: 3
  # ... snipped ...
```
