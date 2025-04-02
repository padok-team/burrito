# Configure a Terraform/Terragrunt/OpenTofu version

By leveraging [`tenv`](https://github.com/tofuutils/tenv), Burrito auto-detects the Terraform, Terragrunt or OpenTofu version used in your repository, with version constraints set in your code (see [`tenv`'s README](https://github.com/tofuutils/tenv/blob/main/README.md)).

Additionally, you can to specify version constraints in the `TerraformRepository` or `TerraformLayer` resource as described below.

## Choose Terraform version

Both `TerraformRepository` and `TerraformLayer` expose a `spec.terraform.version` map field that support version constraints as described in the [Terraform documentation](https://www.terraform.io/docs/language/expressions/version-constraints.html).

If the field is specified for a given `TerraformRepository` it will be applied by default to all `TerraformLayer` linked to it.

If the field is specified for a given `TerraformLayer` it will take precedence over the `TerraformRepository` configuration.

## Enable Terragrunt

You can specify usage of Terragrunt with the `spec.terraform.terragrunt` map as follow:

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformLayer
metadata:
  name: random-pets-terragrunt
spec:
  terraform:
    version: "~> 1.3.0"
    enabled: true
  terragrunt:
    enabled: true
    version: "0.44.5"
  remediationStrategy:
    autoApply: false
  path: "internal/e2e/testdata/terragrunt/random-pets/prod"
  branch: "feat/handle-terragrunt"
  repository:
    name: burrito
    namespace: burrito
```

!!! info
    This configuration can be specified at the `TerraformRepository` level to be enabled by default in each of its layers.

## Use OpenTofu instead of Terraform

To leverage OpenTofu simply use the `opentofu` block in place of the `terraform` block described above:

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformLayer
metadata:
  name: random-pets-opentofu
spec:
  opentofu:
    version: "~> 1.9.0"
    enabled: true
  path: "internal/e2e/testdata/terraform/random-pets"
  branch: "main"
  repository:
    name: burrito
    namespace: burrito
```
