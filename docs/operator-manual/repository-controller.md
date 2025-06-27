# Repository Controller

The Repository Controller is responsible for managing the lifecycle of `TerraformRepository` resources in Burrito. This document explains how the controller works with Git repositories, branches, revisions, and bundles.

## Overview

The Repository Controller continuously monitors `TerraformRepository` resources in the Kubernetes cluster. For each repository, it:

1. Authenticates to the Git provider using the appropriate credentials
2. Retrieves information about branches and their latest revisions (commits)
3. Creates Git bundles for efficient code distribution
4. Stores these bundles in the datastore for runners to access
5. Updates the status of the repository with the latest branch information

## Repository States

The Repository Controller implements a state machine for each `TerraformRepository`. The possible states are:

- **SyncNeeded**: The repository needs to be synchronized with the Git provider
- **Synced**: The repository has been synced recently and bundles are up-to-date

## Git Bundles

### What are Git Bundles?

Git bundles are a convenient way to transfer Git repository data efficiently. A bundle is a single file that contains all the necessary Git objects (commits, trees, blobs) needed to recreate a specific branch or set of commits.

### Why Bundles?

Burrito uses Git bundles for several advantages:

1. **Efficiency**: Bundles contain only the necessary objects, reducing data transfer
2. **Portability**: A bundle can be passed around as a single file
3. **Offline Access**: Runners don't need direct access to the original Git repository
4. **Performance**: Bundles can be unpacked faster than cloning a repository
5. **Security**: Minimizes direct Git repository access from multiple pods

### Bundle Creation Process

When a `TerraformRepository` is reconciled:

1. The controller connects to the Git repository using the appropriate credentials
2. For each branch specified (or detected), it checks the latest revision (commit SHA)
3. It creates a bundle containing all objects needed for that branch
4. The bundle is stored in the datastore with metadata including:
      - Repository namespace and name
      - Branch name
      - Revision (commit SHA)

### Bundle Usage by Runners

When a runner needs to access repository code:

1. The runner requests the bundle from the datastore instead of cloning the repository
2. The bundle is retrieved using the repository, branch, and revision information
3. The runner unpacks the bundle locally to access the Git objects
4. The code is then available for Terraform operations

## Revision Handling

### What is a Revision?

In Burrito, a revision refers to the specific commit SHA that represents the latest state of a branch. Each revision uniquely identifies a specific state of the repository's files.

### Revision Tracking

The Repository Controller:

1. Queries the Git provider for the latest revision of each branch
2. Compares it with the previously stored revision
3. If a new revision is detected:
      - Creates a new Git bundle for the updated code
      - Stores it in the datastore
      - Updates the repository status with the new revision information

### Revision and Bundle Synchronization

The controller ensures that the datastore always has the latest bundle for each branch's current revision. This is accomplished through:

1. Regular reconciliation checks (periodic polling)
2. Webhook events that trigger immediate reconciliation

## How to Monitor Repository Status

You can check the status of a repository using kubectl:

```bash
kubectl get terraformrepository -n <namespace> <repository-name> -o yaml
```

Example status output:

```yaml
status:
  branches:
  - lastSyncDate: Fri Jun 13 13:48:48 UTC 2025
    lastSyncStatus: success
    latestRev: 7a07490d90fdf6ffae19a6d00420871415a11ae5
    name: main
  conditions:
  - status: "False"
    type: IsLastSyncTooOld
    # snip ...
  - status: "False"
    type: HasLastSyncFailed
    # snip ...
  state: Synced
```

## Troubleshooting

If you encounter issues with the Repository Controller or TerraformRepository resources:

1. Check the controller logs:

   ```bash
   kubectl logs -n <burrito-namespace> deployment/burrito-controller -c controller
   ```

2. Describe the TerraformRepository resource:

   ```bash
   kubectl describe terraformrepository -n <namespace> <repository-name>
   ```

3. Check the events related to the resource:

   ```bash
   kubectl get events -n <namespace> --field-selector involvedObject.name=<repository-name>
   ```
