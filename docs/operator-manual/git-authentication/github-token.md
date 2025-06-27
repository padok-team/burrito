# GitHub Token Authentication

## Generate a personal access token

You need a personal access token to configure Burrito. You can generate a personal access token in your GitHub account.

Follow the instructions in the GitHub documentation for [creating a personal access token](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token):

- It should be a **fine-grained token**.
- **Permissions**: Configure the following **Repository Permissions**:
    - **Metadata:** Select Read-only.
    - **Contents:** Select Read-only.
    - **Pull requests:** Select Read & write. This is required to issue comments on pull requests.
- Under **Repository access**, select which repositories you want the token to access.

## Configure credentials with GitHub Token

Set up a credentials secret with the `githubToken` field.

### Repository-specific credentials example

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformRepository
metadata:
  name: my-repository
  namespace: burrito-project
spec:
  repository:
    url: https://github.com/owner/repo
  terraform:
    enabled: true
---
apiVersion: v1
kind: Secret
metadata:
  name: burrito-repo
  namespace: burrito-project
type: credentials.burrito.tf/repository
stringData:
  provider: github
  url: https://github.com/owner/repo
  githubToken: "github_pat_xxx"
  webhookSecret: "my-webhook-secret"
```

### Shared credentials example

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: github-token-credentials
  namespace: burrito-system
type: credentials.burrito.tf/shared
stringData:
  provider: github
  url: https://github.com/owner
  githubToken: "github_pat_xxx"
  webhookSecret: "my-webhook-secret"
```
