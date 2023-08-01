# Getting started

## Requirements

- Installed [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) command-line tool.
- Have a [kubeconfig](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/) file (default location is `~/.kube/config`).

## 1. Install burrito

```bash
kubectl create namespace burrito
kubectl apply -n burrito -f https://raw.githubusercontent.com/padok-team/burrito/main/manifests/install.yaml
```

This will create a new namespace, `burrito`, where burrito services will live.

!!! warning
    The installation manifests include `ClusterRoleBinding` resources that reference `burrito` namespace. If you are installing burrito into a different namespace then make sure to update the namespace reference.

!!! info
    You might be interested by our [Helm chart](./operator-manual/install/with-helm.md), that provides more control over burrito's configuration as well as a [multi-tenant architecture](./operator-manual/multi-tenant-architecture.md).

## 2. Create a connection to a private repository

Create a Kubernetes `Secret` which looks like:

```yaml
kind: Secret
metadata:
  name: burrito-repo
  namespace: burrito
type: Opaque
stringData:
  username: <my-username>
  password: <my-password | my-access-token>
  sshPrivateKey: |
    -----BEGIN OPENSSH PRIVATE KEY-----
    ...
    -----END OPENSSH PRIVATE KEY-----
```

Then, create a `TerraformRepository` Kubernetes resource:

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformRepository
metadata:
  name: burrito
  namespace: burrito
spec:
  repository:
    url: <https_or_ssh_repository_url>
    secretName: burrito-repo
```

!!! info
    You can also connect to a public repository by omitting `spec.repository.secretName` in your `TerraformLayer` definition.

## 3. Synchronize a terraform layer

After creating a `TerraformRepository` you can create a `TerraformLayer` ressource which looks like:

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformLayer
metadata:
  name: random-pets
  namespace: burrito
spec:
  terraform:
    version: "1.3.1"
  path: "internal/e2e/testdata/random-pets"
  branch: "main"
  repository:
    name: burrito
    namespace: burrito
```
