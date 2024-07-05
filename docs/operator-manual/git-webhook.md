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

## Reference the webhook secret in the controller

Create a secret called `burrito-webhook-secret` in the controller namespace with the webhook secret.
```yaml
kind: Secret
apiVersion: v1
metadata:
  name: burrito-webhook-secret
  namespace: burrito-system
type: Opaque
stringData:
  burrito-webhook-secret: <my-webhook-secret>
```

Add the webhook secret as an environment variable of the `burrito-server`. The variables depends on your git provider.

| Git provider |          Environment Variable          |
| :----------: | :------------------------------------: |
|    GitHub    | `BURRITO_SERVER_WEBHOOK_GITHUB_SECRET` |
|    GitLab    | `BURRITO_SERVER_WEBHOOK_GITLAB_SECRET` |
