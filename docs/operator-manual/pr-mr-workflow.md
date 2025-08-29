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
- `CommentNeeded`. This is the state of a pull request if the controller needs to send a comment to the git provider's API.

## Configuration

### Webhook

Follow the instructions in [Setting up a Git Webhook](./git-webhook.md) to configure a webhook in your repository.
The webhook will be used to trigger:

- Drift detection when a push event is received.
- The PR/MR workflow when a pull request event is received.

### GitHub with a dedicated GitHub App

#### Create the GitHub App

You can create and register GitHub Apps in your personal GitHub account or in any GitHub organization where you have administrative access.

Follow the instructions in the GitHub documentation on [Creating a GitHub App](https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/registering-a-github-app). Populate the settings as follows:

- **GitHub App Name**: Choose a name for your GitHub App. For example, something featuring `burrito`.
- **Homepage URL**: Enter <https://docs.burrito.tf>.
- **Webhook**: Unselect Active. The app doesn't use this webhook events mechanism at the moment.
- **Permissions**: Configure the following **Repository Permissions**:
  - **Metadata**: Select Read-only.
  - **Pull requests**: Select Read & write. This is required to issue comments on pull requests.
- Where can this GitHub App be installed: Select **Any account**.

#### Creating a custom badge for your GitHub App

You can create a custom badge for your GitHub App to display on your GitHub repository. Follow the instructions in the GitHub documentation on [Creating a custom badge for your GitHub App](https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/creating-a-custom-badge-for-your-github-app).

We suggest using the following one:

<p align="center"><img src="../../assets/icon/burrito.png" width="200px" /></p>

#### Install the GitHub App

Follow the instructions in the GitHub documentation on [Installing your own GitHub App](https://docs.github.com/en/apps/using-github-apps/installing-your-own-github-app), and note the following:

- For Repository access, select **Only select repositories**, and then select the repos you want to connect with Burrito.

#### Get the Installation ID and App ID

You need the **Installation ID** and **App ID** to configure Burrito.

<!-- markdownlint-disable MD032 -->
1. Get the **Installation ID** from the URL of the installed app, such as:
  <p align="center"><img src="../../assets/pr-mr-workflow/github_installation_id.png" /></p>
2. Get the **App ID** from the app's General tab.
  <p align="center"><img src="../../assets/pr-mr-workflow/github_app_id.png" /></p>
<!-- markdownlint-enable MD032 -->

#### Generate a private key

You need a private key for your GitHub app to configure Burrito.

- Follow the instructions in the GitHub documentation for [generating private keys for GitHub Apps](https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/managing-private-keys-for-github-apps#generating-private-keys)

- Save the private key file to your local machine. GitHub only stores the public portion of the key.

#### Configure a repository secret with the GitHub App credentials

Add the credentials of your newly created app to the secret associated to your `TerraformRepository` resource. If the repository is public, create a secret in the same namespace as the `TerraformRepository` and reference it in the `spec.repository.secretName`.

For example:

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformRepository
metadata:
  name: my-repository
  namespace: burrito-project
spec:
  repository:
    url: https://github.com/owner/repo
    secretName: burrito-repo
  terraform:
    enabled: true
---
apiVersion: v1
kind: Secret
metadata:
  name: burrito-repo
  namespace: burrito-project
type: Opaque
stringData:
  webhookSecret: "my-webhook-secret"
  githubAppId: "123456"
  githubAppInstallationId: "12345678"
  githubAppPrivateKey: |
    -----BEGIN RSA PRIVATE KEY-----
    my-private-key
    -----END RSA PRIVATE KEY-----
```

### GitHub with a personal access token

#### Generate a personal access token

You need a personal access token to configure Burrito. You can generate a personal access token in your GitHub account.

Follow the instructions in the GitHub documentation for [creating a personal access token](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token):

- It should be a **fine-grained token**.
- **Permissions**: Configure the following **Repository Permissions**:
  - **Metadata**: Select Read-only.
  - **Pull requests**: Select Read & write. This is required to issue comments on pull requests.
- Under **Repository access**, select which repositories you want the token to access.

#### Configure Burrito

Set the `githubToken` key in the secret associated to your `TerraformRepository` resource.

For example:

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformRepository
metadata:
  name: my-repository
  namespace: burrito-project
spec:
  repository:
    url: https://github.com/owner/repo
    secretName: burrito-repo
  terraform:
    enabled: true
---
apiVersion: v1
kind: Secret
metadata:
  name: burrito-repo
  namespace: burrito-project
type: Opaque
stringData:
  githubToken: "123456"
```

### GitLab

#### Generate a private token

You need a private token for your GitLab app to configure Burrito. You can generate a private token in your GitLab account. Follow the instructions in the GitLab documentation for [generating a private token](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html#creating-a-personal-access-token).

#### Configure Burrito

Set the `gitlabToken` key in the secret associated to your `TerraformRepository` resource.

For example:

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformRepository
metadata:
  name: my-repository
  namespace: burrito-project
spec:
  repository:
    url: https://gitlab.com/owner/repo
    secretName: burrito-repo
  terraform:
    enabled: true
---
apiVersion: v1
kind: Secret
metadata:
  name: burrito-repo
  namespace: burrito-project
type: Opaque
stringData:
  gitlabToken: "123456"
```
