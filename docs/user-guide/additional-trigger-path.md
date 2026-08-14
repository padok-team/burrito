# Additional Trigger Paths

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

To do so, I just have to set `spec.additionalTriggerPaths` on my `TerraformLayer` as below. Each entry is resolved relative to `spec.path`, so `../` is used to go up to the repository root before reaching `modules/random-pets`.

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
  additionalTriggerPaths:
    - "../../../modules/random-pets"
  branch: "main"
  repository:
    name: burrito
    namespace: burrito
```

Now, when a change occurs in the `modules/random-pets` directory, the layer will be plan / apply.

## Glob patterns

Each entry in `additionalTriggerPaths` can also be a glob pattern, including the recursive `**` wildcard. This is useful when you want to trigger the layer for a given file type anywhere under a path, without having to list every subdirectory:

```yaml
spec:
  path: "terragrunt/random-pets/test"
  additionalTriggerPaths:
    - "../**/*.yaml"
```

This matches any `.yaml` file anywhere under `terragrunt/random-pets` (the resolved path of `../` from `spec.path`), at any depth.

Entries without any glob character (`*`, `?`, `[`) keep the historical exact-path matching shown in the first example.

## Deprecated: annotation

Prior to `spec.additionalTriggerPaths`, this feature was configured via the `config.terraform.padok.cloud/additionnal-trigger-paths` annotation (comma-separated paths, same relative resolution and glob support as above):

```yaml
metadata:
  annotations:
    config.terraform.padok.cloud/additionnal-trigger-paths: "../../../modules/random-pets"
```

This annotation is deprecated and only used when `spec.additionalTriggerPaths` is not set. **If `spec.additionalTriggerPaths` is set, the annotation is ignored entirely** (no merging between the two). Using the annotation logs a `DEPRECATED` warning; migrate to the spec field.

> **Note:** the annotation name contains a typo — it is `additionnal` (with two `n`), not `additional`. This is kept as-is for backward compatibility; do not "fix" it in your manifests, it must match exactly.
