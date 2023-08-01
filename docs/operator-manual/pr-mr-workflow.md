# PR/MR Workflow

<p align="center"><img src="../../assets/design/pr-mr-workflow.excalidraw.png" width="1000px" /></p>

!!! info
    In this documentation all references to pull requests can be change to merge requests for GitLab. However, the resulting Kubernetes object will still be named `TerraformPullRequest`.

## Components

### The server

!!! info
    For more information about the server, see the [architectural overview](./architecture.md) documentation.

Upon receiving a Pull Request creation event, the server creates a `TerraformPullRequest` resource.

Upon receiving a Pull Request deletion event, the server deletes the related `TerraformPullRequest` resource.

### The pull request controller

The pull request controller is a Kubernetes controller which continuously monitors declared `TerraformPullRequest` resources.

It is responsible for creating temporary `TerraformLayer` resources linked to the Pull Request it was generated from. Once all the `TerraformLayer` have planned, it will send a comment containing the plan results to the pull request.

<p align="center"><img src="../../assets/demo/comment.png" width="1000px" /></p>

#### Implementation

The status of a `TerraformPulLRequest` is defined using the [conditions standards defined by the community](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties).

3 conditions ared defined for a pull request:

- `IsLastCommitDiscovered`. This condition is used to check if we received a new commit on the pull request by comparing the latest commit on the branch and the last discovered commit.
- `AreLayersStillPlanning`. This condition is used to check if all the temporary layers have finished planning. This is done by checking all the resulting `TerraformLayer` statuses.
- `IsCommentUpToDate`. This condition is used to check if the controller needs to send a comment to a pull request. This is checked by comparing the last discovered commit and the last commit for which a comment was already sent.

!!! info
    We use annotations to store information.

With those 3 conditions, we defined 3 states:

- `Idle`. This is the state of a pull request if nothing needs to be done.
- `DiscoveryNeeded`. This is the state of a pull request if the controller needs to check which layers are affected on the given pull request.
- `CommentNeeded`.  This is the state of a pull request if the controller needs to send a comment to the git provider's API.

## Configuration

|            Environment variable            |                  Description                  |
| :----------------------------------------: | :-------------------------------------------: |
| `BURRITO_CONTROLLER_GITHUBCONFIG_APITOKEN` | the API token to send comment to GitHub's API |
| `BURRITO_CONTROLLER_GITLABCONFIG_APITOKEN` | the API token to send comment to GitLab's API |
|   `BURRITO_CONTROLLER_GITLABCONFIG_URL`    |        the URL of the GitLab instance         |
