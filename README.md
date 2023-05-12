# burrito <!-- omit in toc -->

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/padok-team/burrito)](https://goreportcard.com/report/github.com/padok-team/burrito)
[![codecov](https://codecov.io/gh/padok-team/burrito/branch/main/graph/badge.svg)](https://codecov.io/gh/padok-team/burrito)

<p align="center"><img src="./docs/assets/icon/burrito.png" width="200px" /></p>

**Burrito** is a TACoS (**T**erraform **A**utomation **Co**llaboration **S**oftware) Kubernetes Operator.

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

## Documenation

To learn more about burrito [go to the complete documentation](https://padok-team.github.io/burrito/).

## Community

### Contibution, Discussion and Support

You can reach burrito's maintainers on Twitter:

- [@spoukke](https://twitter.com/spoukke)
- [@LonguetAlan](https://twitter.com/LonguetAlan)

### Blogs and Presentations

1. [Our burrito is a TACoS](https://www.padok.fr/en/blog/burrito-tacos)

## License

© 2022 [Padok](https://www.padok.fr/).

Licensed under the [Apache License](https://www.apache.org/licenses/LICENSE-2.0), Version 2.0 ([LICENSE](./LICENSE))
