# GitLab Token Authentication

## Generate a private token

You need a private token for your GitLab app to configure Burrito. You can generate a private token in your GitLab account.

Follow the instructions in the GitLab documentation for [generating a private token](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html#creating-a-personal-access-token).

## Configure credentials with GitLab Token

Set up a credentials secret with the `gitlabToken` field.

### Repository-specific credentials example

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformRepository
metadata:
  name: my-repository
  namespace: burrito-project
spec:
  repository:
    url: https://gitlab.com/owner/repo
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
  provider: gitlab
  url: https://gitlab.com/owner/repo
  gitlabToken: "glpat-xxxx"
  webhookSecret: "my-webhook-secret"
```

### Shared credentials example

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: gitlab-token-credentials
  namespace: burrito-system
type: credentials.burrito.tf/shared
stringData:
  provider: gitlab
  url: https://gitlab.example.com/owner
  gitlabToken: "glpat-xxxx"
  webhookSecret: "my-webhook-secret"
```
