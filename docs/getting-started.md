# Getting started

## Prerequisites

- A Kubernetes cluster
- [Optional for testing, necessary for production use] A storage bucket in a cloud provider (AWS, GCP, Azure)
- [Optional, recommended for production use] cert-manager installed in your cluster (for internal encryption of plans and logs & provider cache)

## Requirements

- [helm](https://helm.sh/docs/intro/install/) CLI
- Have a [kubeconfig](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/) file (default location is `~/.kube/config`) to access your Kubernetes cluster

## 1. Install Burrito

Copy and modify the [default values](https://github.com/padok-team/burrito/blob/main/deploy/charts/burrito/values.yaml) to match your requirements.

Make sure to configure a tenant by updating the `tenant` field in the `values.yaml` file. The associated namespace will be created automatically and used to deploy Burrito resources in step 3.

For example, here is a default `values.yaml` file:

```yaml
config:
  datastore:
    storage:
      mock: true

tenants:
  - namespace:
      create: true
      name: "burrito-project-1"
    serviceAccounts:
      - name: "runner-project-1"
```

!!! info
    To try Burrito without setting up a remote storage, set the `config.burrito.datastore.storage.mock` field to `true` in the `values.yaml` file. To persist data such as Terraform logs, you must configure a storage bucket field. Make sure to specify a service account that has the necessary permissions to read/write to your remote bucket.

Then, install Burrito using the following command:

```bash
helm install burrito oci://ghcr.io/padok-team/charts/burrito --create-namespace -n burrito-system -f ./values.yaml
```

This will create a new namespace, `burrito-system`, where Burrito services will be deployed.

## 2. Create a connection to a private repository

Create a Kubernetes `Secret` to reference the necessary credentials to clone your IaC repository (GitHub or GitLab)

<!-- markdownlint-disable MD046 -->
!!! info
    Supported authentication methods are:

    - Username and password
    - SSH private key
    - GitHub App
    - GitHub API token
    - GitLab API token

    More information on how to create a secret can be found in the [Git Authentication](./operator-manual/git-authentication.md) section.
<!-- markdownlint-enable MD046 -->

```yaml
kind: Secret
metadata:
  name: burrito-repo
  namespace: <tenant-namespace>
type: Opaque
stringData:
  username: <my-username>
  password: <my-password | my-access-token>
  sshPrivateKey: |
    -----BEGIN OPENSSH PRIVATE KEY-----
    ...
    -----END OPENSSH PRIVATE KEY-----
```

Then, create a `TerraformRepository` Kubernetes resource. The `spec.terraform.enabled` sets the repository as a Terraform repository (as opposed to an OpenTofu repository). This setting will propagate to all layers linked to this repository by default, but can be overridden at the layer level.

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformRepository
metadata:
  name: burrito
  namespace: <tenant-namespace>
spec:
  repository:
    url: <https-or-ssh-repository-url>
    secretName: burrito-repo
  terraform:
    enabled: true
```

!!! info
    You can also connect to a public repository by omitting `spec.repository.secretName` in your `TerraformRepository` definition.

## 3. Synchronize a Terraform layer

After creating a `TerraformRepository` you can create a `TerraformLayer` resource which looks like:

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformLayer
metadata:
  name: random-pets
  namespace: burrito
spec:
  terraform:
    version: "1.3.1"
  path: "internal/e2e/testdata/terraform/random-pets"
  branch: "main"
  repository:
    name: burrito
    namespace: burrito
```

The controller will create a runner pod in your tenant namespace to synchronize the repository and apply the Terraform code.

## Guides

- For detailed guides on how to use Burrito, see the [Guides](./guides/index.md) section.
- To learn more about advanced configuration and features, see the [Operator Manual](./operator-manual/index.md) section.
