# Getting started

## Requirements

- Installed [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) command-line tool.
- Have a [kubeconfig](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/) file (default location is `~/.kube/config`).

1. Install burrito

```bash
kubectl create namespace burrito
kubectl apply -n burrito -f https://raw.githubusercontent.com/padok-team/burrito/main/manifests/install.yaml
```

This will create a new namespace, `burrito`, where burrito services will live.

!!! warning
    The installation manifests include `ClusterRoleBinding` resources that reference `burrito` namespace. If you are installing burrito into a different namespace then make sure to update the namespace reference.

2. Create a layer from a Git Repository

> TODO: reference burrito-examples repository
