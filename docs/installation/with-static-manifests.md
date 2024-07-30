# Install burrito with static manifests

## Requirements

- Installed [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) command-line tool.
- Have a [kubeconfig](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/) file (default location is `~/.kube/config`).

## Install burrito

!!! info
    This will install a mono-tenant version of burrito. See the [Helm installation method](./with-helm.md) for a [multi-tenant-architecture](../multi-tenant-architecture.md).

```bash
kubectl create namespace burrito
kubectl apply -n burrito -f https://raw.githubusercontent.com/padok-team/burrito/main/manifests/install.yaml
```

This will create a new namespace, `burrito`, where burrito services will live.

!!! warning
    The installation manifests include `ClusterRoleBinding` resources that reference `burrito` namespace. If you are installing burrito into a different namespace then make sure to update the namespace reference.
