# Setup a Git Webhook

Create a webhook (with a secret!) in the repository you want to receive events from.

Then create a secret:

```yaml
kind: Secret
metadata:
  name: burrito-webhook-secret
  namespace: burrito
type: Opaque
stringData:
  burrito-webhook-secret: <my-webhook-secret>
```

Add the webhook secret as an environment variable of the `burrito-server`. The variables depends on your git provider.

| Git provider |          Environment Variable          |
| :----------: | :------------------------------------: |
|    GitHub    | `BURRITO_SERVER_WEBHOOK_GITHUB_SECRET` |
|    GitLab    | `BURRITO_SERVER_WEBHOOK_GITLAB_SECRET` |
