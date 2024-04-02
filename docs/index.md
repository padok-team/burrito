# Overview

## What is burrito?

**Burrito** is a TACoS (**T**erraform **A**utomation **Co**llaboration **S**oftware) Kubernetes Operator.

![demo](assets/demo/demo.gif)

## Why burrito?

[`terraform`](https://www.terraform.io/) is a tremendous tool to manage your infrastructure in IaC.
However, it lacks built-in solutions for managing [state drift](https://developer.hashicorp.com/terraform/tutorials/state/resource-drift).

Additionally, configuring a CI/CD pipeline for Terraform can be challenging and often varies depending on the selected tools

Finally, currently, there is no easy way to navigate your Terraform state to truly understand the modifications it undergoes when running `terraform apply`.

`burrito` aims to tackle those issues by:

- Planning continuously your Terraform code and run applies if needed
- Offering an out of the box PR/MR integration so you do not have to write CI/CD pipelines for Terraform ever again
- Showing your state's modifications in a simple Web UI (not implemented yet)

## Getting started

### Quick start

```bash
kubectl create namespace burrito
kubectl apply -n burrito -f https://raw.githubusercontent.com/padok-team/burrito/main/manifests/install.yaml
```

Follow our [getting started guide](./getting-started.md). Further user oriented [documentation](./user-guide/) is provided for additional features.
