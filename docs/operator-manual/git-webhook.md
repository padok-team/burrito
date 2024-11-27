# Setup a Git Webhook

## Expose Burrito server to the internet

Expose the `burrito-server` service to the internet using the method of your choice. (e.g. ingress, port-forward & ngrok for local testing...). Accessing the URL on the browser should display the Burrito UI.

## Configure a webhook on GitHub or GitLab

Create a webhook (with a secret!) in the repository you want to receive events from.  
The target URL must point to the exposed `burrito-server` on the `/api/webhook` path.

GitHub triggers:  
The webhook should be triggered on `push` and `pull_request` events.

GitLab triggers:  
The webhook should be triggered on `Push events` from all branches and `Merge request events`.

## Reference the webhook secret in the repository secret

Add the webhook secret to the secret used to authenticate to the repository. If the repository is public, create a secret in the same namespace as the `TerraformRepository` and reference it in the `spec.repository.secretName`.
Reference the webhook secret in the webhookSecret key of the Kubernetes secret.

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
```

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: burrito-repo
  namespace: burrito-project
type: Opaque
stringData:
  githubAppId: "123456"
  githubAppInstallationId: "12345678"
  githubAppPrivateKey: |
    -----BEGIN RSA PRIVATE KEY-----
    my-private-key
    -----END RSA PRIVATE KEY-----
  webhookSecret: "my-webhook-secret" 
```
