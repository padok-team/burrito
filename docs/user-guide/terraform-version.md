# Choose a terraform/terragrunt version

## Choose terraform version

Both `TerraformRepository` and `TerraformLayer` expose a `spec.terrafrom.version` map field.

If the field is specified for a given `TerraformRepository` it will be applied by default to all `TerraformLayer` linked to it.

If the field is specified for a given `TerraformLayer` it will take precedence over the `TerraformRepository` configuration.

## Enable Terragrunt

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

!!! info
    This configuration can be specified at the `TerraformRepository` level to be enabled by default in each of its layers.
