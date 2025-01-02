# Git Authentication

## Overview

This section will guide you through the different ways to authenticate to a git repository.
Authentication is required for:
- Cloning a private repository
- Implementing the [PR/MR workflow](./pr-mr-workflow.md)

Available authentication methods:
- Username and password (only supports cloning)
- SSH private key (only supports cloning)
- GitHub App
- GitHub API token
- GitLab API token

## Repository Secret

The `TerraformRepository` spec allows you to specify a secret that contains the credentials to authenticate to a git repository.
The secret must be created in the same namespace as the `TerraformRepository` and be referenced in `spec.repository.secretName`.

### Expected keys

To add an authentication method for a repository, the secret must contain the following keys:

Username and password:
- `username`
- `password`

SSH private key:
- `sshPrivateKey`

GitHub App:
- `githubAppId`
- `githubAppInstallationId`
- `githubAppPrivateKey`

GitHub API token:
- `githubToken`

GitLab API token:
- `gitlabToken`

For the PR/MR workflow, the Kubernetes secret must also contain the webhook secret:
- `webhookSecret`

Example of a Kubernetes secret for a GitHub repository, using authentication with a GitHub App and implementing the PR workflow:

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

### Behavior

If multiple authentication methods are provided, the runner will try them all until one succeeds to clone the repository.
