# Gitea / Forgejo Token Authentication

Burrito supports [Gitea](https://gitea.com) and [Forgejo](https://forgejo.org) instances through the same `gitea` provider: Forgejo is a Gitea fork that exposes a Gitea-compatible API and sends Gitea-compatible webhooks. Both the `gitea` and `forgejo` provider keys are accepted in credential secrets.

## Generate an access token

Create a token with `issue` and `repository` scopes on your Gitea/Forgejo instance (Settings → Applications → Access Tokens). The token is used for API calls (PR listing, comments, statuses) and git authentication.

## Configure credentials with a Gitea/Forgejo token

Set up a credentials secret with the `giteaToken` field.

### Repository-specific credentials example

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformRepository
metadata:
  name: my-repository
  namespace: burrito-project
spec:
  repository:
    url: https://gitea.example.com/owner/repo
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
  provider: gitea
  url: https://gitea.example.com/owner/repo
  giteaToken: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  webhookSecret: "my-webhook-secret"
```

### Shared credentials example

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: gitea-shared-credentials
  namespace: burrito-system
type: credentials.burrito.tf/shared
stringData:
  provider: gitea
  url: https://gitea.example.com
  giteaToken: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  webhookSecret: "my-webhook-secret"
```

## Configure a webhook

Create a **Gitea** webhook (not Slack/Dingtalk/…) on your repository, with a secret, targeting the exposed `burrito-server` `/api/webhook` endpoint with the `application/json` content type. Trigger it on **push** and **pull request** events. Forgejo webhooks work the same way.
