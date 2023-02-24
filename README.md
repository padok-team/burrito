# burrito <!-- omit in toc -->
asdf
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

<p align="center"><img src="./docs/assets/icon/burrito.png" width="200px" /></p>

**Burrito** is a TACoS (**T**erraform **A**utomation **Co**llaboration **S**oftware) Kubernetes Operator.

- [Why does this exists?](#why-does-this-exists)
- [Demo](#demo)
- [Installation](#installation)
- [Getting started](#getting-started)
  - [Connect to a private repository](#connect-to-a-private-repository)
  - [Synchronize a terraform layer](#synchronize-a-terraform-layer)
- [Advanced usage](#advanced-usage)
- [How it works](#how-it-works)
- [License](#license)

## Why does this exists?

[`terraform`](https://www.terraform.io/) is a tremendous tool to manage your infrastructure in IaC.
But, it does not come up with an out-of the box solution for managing [state drift](https://developer.hashicorp.com/terraform/tutorials/state/resource-drift).

Also, writing a CI/CD pipeline for terraform can be painful and depends on the tool you are using.

Finally, currently, there is no easy way to navigate your terraform state to truly understand the modifications it undergoes when running `terraform apply`.

`burrito` aims to tackle those issues by:

- Planning continuously your terraform code and run applies if needed
- Offering an out of the box PR/MR integration so you do not have to write CI/CD pipelines for terraform ever again (not implemented yet)
- Showing your state's modifications in a simple Web UI (not implemented yet)

## Demo 

![demo](./docs/assets/demo/demo.gif)

## Installation

You can just run the following:

```bash
kubectl apply -f https://raw.githubusercontent.com/padok-team/burrito/main/manifests/install.yaml
```

## Getting started

`burrito` uses two [CRDs](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/):
- `TerraformRepository` which defines the way to connect to a public or private repository containing some `terraform` code
- `TerraformLayer` which defines a `terraform` layer insidea given `TerraformRepository`

### Connect to a private repository

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

> N.B. You can also connect to a public repository by omitting `spec.repository.secretName` in your `TerraformLayer` definition.

### Synchronize a terraform layer

First, you need to create a `TerraformRepository`.

Then, create a `TerraformLayer` ressource which looks like:

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformLayer
metadata:
  name: random-pets
  namespace: burrito
spec:
  terraformVersion: "1.3.1"
  path: "internal/e2e/testdata/random-pets"
  branch: "main"
  repository:
    kind: TerraformRepository
    name: burrito
    namespace: burrito
```

## Advanced usage

See [Usage](docs/contents/usage) for more advanced configuration options.

## How it works

See [Design](docs/contents/design) for details on how `burrito` works under the hood.

## License

© 2022 [Padok](https://www.padok.fr/).

Licensed under the [Apache License](https://www.apache.org/licenses/LICENSE-2.0), Version 2.0 ([LICENSE](./LICENSE))
