# Additionnal Trigger Paths

By default, when you are creating a layer, you must specify a repository and a path. This path is used to trigger the layer changes which means that when a change occurs in this path, the layer will be plan / apply accordingly.

Sometimes, you need to trigger changes on a layer where the changes are not in the same path (e.g. update made on an internal terraform module hosted on the same repository).

That's where the additional trigger paths feature comes!

Let's take the following `TerraformLayer`:

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformLayer
metadata:
  name: random-pets-terragrunt
spec:
  terraform:
    enabled: true
    version: "1.3.1"
  terragrunt:
    enabled: true
    version: "0.45.4"
  remediationStrategy:
    autoApply: true
  path: "terragrunt/random-pets/test"
  branch: "main"
  repository:
    name: burrito
    namespace: burrito
```

The repository's path of my `TerraformLayer` is set to `terragrunt/random-pets/test`. But I want to trigger the layer plan / apply when a change occurs on my module which is in the `modules/random-pets` directory of my repository.

To do so, I just have to add the `config.terraform.padok.cloud/additionnal-trigger-paths` annotation to my `TerraformLayer` as below.

!!! warning
    Additional trigger paths are **relative to the layer `spec.path`**, not to the repository root. Here the layer path is `terragrunt/random-pets/test`, so `modules/random-pets` is reached with `../../../modules/random-pets`. A path written from the repository root (e.g. `modules/random-pets`) would resolve to `terragrunt/random-pets/test/modules/random-pets` and never match.

You can set several paths separated with a comma (spaces around commas are ignored), e.g. `"../../../modules/random-pets, ../../../modules/common"`.

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformLayer
metadata:
  name: random-pets-terragrunt
  annotations:
    config.terraform.padok.cloud/additionnal-trigger-paths: "../../../modules/random-pets"
spec:
  terraform:
    enabled: true
    version: "1.3.1"
  terragrunt:
    enabled: true
    version: "0.45.4"
  remediationStrategy:
    autoApply: true
  path: "terragrunt/random-pets/test"
  branch: "main"
  repository:
    name: burrito
    namespace: burrito
```

Now, when a change occurs in the `modules/random-pets` directory, the layer will be plan / apply.
