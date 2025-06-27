# Setup a Git Webhook

Using the webhook feature, Burrito can automatically trigger actions based on events in your Git repository, such as pull requests or pushes. This allows for real-time updates and interactions with your infrastructure as code.

Otherwise, Burrito will poll repositories for changes, which may not be as efficient or timely as using webhooks.

!!! warning "Burrito polling limitations"
    Burrito's automatic polling of repositories only works for changes on referenced branches in TerraformLayers.
    **Automatic polling of Pull Requests is not implemented yet.**

## Expose Burrito server to the internet

Expose the `burrito-server` service to the internet using the method of your choice. (e.g. ingress, port-forward & ngrok for local testing...). Accessing the URL on the browser should display the Burrito UI.

## Configure a webhook on GitHub or GitLab

Create a webhook (with a secret!) in the repository you want to receive events from.
The target URL must point to the exposed `burrito-server` on the `/api/webhook` path.
Only `application/json` content type is supported.

**GitHub triggers:** The webhook should be triggered on `push` and `pull_request` events.

**GitLab triggers:** The webhook should be triggered on `Push events` from all branches and `Merge request events`.

## Configure the webhook secret in credentials

Add the webhook secret to the repository or shared credentials used to authenticate to the repository. The webhook secret is used to validate the authenticity of webhook payloads from your Git provider.

### Using repository-specific credentials

Create a credential secret in the same namespace as the `TerraformRepository` using the type `credentials.burrito.tf/repository`:

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
```

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: burrito-repo
  namespace: burrito-project
type: credentials.burrito.tf/repository
stringData:
  provider: github
  url: https://github.com/owner/repo
  githubAppID: "123456"
  githubAppInstallationID: "12345678"
  githubAppPrivateKey: |
    -----BEGIN RSA PRIVATE KEY-----
    my-private-key
    -----END RSA PRIVATE KEY-----
  webhookSecret: "my-webhook-secret" # 👈 Used to validate webhook payloads
```

### Using shared credentials

Alternatively, you can use shared credentials that apply to multiple repositories:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: github-shared-credentials
  namespace: burrito-system
  annotations:
    burrito.tf/allowed-tenants: "burrito-project"
type: credentials.burrito.tf/shared
stringData:
  provider: github
  url: https://github.com/owner
  githubToken: "github_pat_xxx"
  webhookSecret: "my-webhook-secret" # 👈 Used to validate webhook payloads
```
