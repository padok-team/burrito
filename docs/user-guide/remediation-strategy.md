# Choose a remediation strategy

Currently, 2 remediation strategies are handled.

|  Strategy   |                               Effect                                |
| :---------: | :-----------------------------------------------------------------: |
|    `dry`    | The operator will only run the `plan`. This is the default strategy |
| `autoApply` |        If a `plan` is not up to date, it will run an `apply`        |

As for the [runner spec override](./override-runner.md), you can specify a `spec.remediationStrategy` either on the `TerraformRepository` or the `TerraformLayer`.

The configuration of the `TerraformLayer` will take precedence.

!!! warning
    This operator is still experimental. Use `spec.remediationStrategy: "autoApply"` at your own risk.
