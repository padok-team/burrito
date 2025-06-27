# Git Authentication

## Overview

This section will guide you through the different ways to authenticate to a git repository.
Authentication is required for:

- Cloning a private repository
- Implementing the [PR/MR workflow](./pr-mr-workflow.md)
- Setting up the webhook for listening to push / pull requests events

Available authentication methods:

- Username and password (only supports cloning)
- SSH private key (only supports cloning)
- [GitHub App](./git-authentication/github-app.md) (recommended for GitHub)
- [GitHub API token](./git-authentication/github-token.md)
- [GitLab API token](./git-authentication/gitlab-token.md)

## Repository Credentials

Burrito uses Kubernetes Secrets to store credentials for repositories. There are two types of credential secrets:

1. **Repository-specific credentials** - Used for a specific repository in a specific namespace
2. **Shared credentials** - Can be used across multiple repositories and namespaces

### Repository-specific Credentials

Repository-specific credentials are tied to a single repository URL in a single namespace. These secrets use the type `credentials.burrito.tf/repository`.

Example of a repository-specific credential secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: burrito-repo
  namespace: default
type: credentials.burrito.tf/repository
stringData:
  provider: github
  url: https://github.com/padok-team/burrito-examples
  githubToken: "github_pat_xxx"
  webhookSecret: "my-webhook-secret"
```

### Shared Credentials

Shared credentials can be used across multiple repositories and optionally restricted to specific namespaces. These secrets use the type `credentials.burrito.tf/shared`.

To restrict shared credentials to specific namespaces, add the annotation `burrito.tf/allowed-tenants` with a comma-separated list of namespace names.

Example of a shared credential secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: github-shared-credentials
  namespace: burrito-system
  annotations:
    burrito.tf/allowed-tenants: "team-a,team-b"
type: credentials.burrito.tf/shared
stringData:
  provider: github
  url: github.com
  githubToken: "github_pat_xxx"
  webhookSecret: "my-webhook-secret"
```

### Credentials Fields

The following fields are available in credential secrets:

Authentication methods:

- Basic authentication:
    - `username` - Username for basic authentication
    - `password` - Password for basic authentication

- SSH authentication:
    - `sshPrivateKey` - SSH private key for authentication

- GitHub App authentication:
    - `githubAppID` - GitHub App ID
    - `githubAppInstallationID` - GitHub App Installation ID
    - `githubAppPrivateKey` - GitHub App Private Key

- Token-based authentication:
    - `githubToken` - GitHub API token
    - `gitlabToken` - GitLab API token

Additional fields:

- `provider` - The Git provider (`github`, `gitlab` or `standard`)
- `url` - The repository URL or domain for matching
- `webhookSecret` - Secret used for webhook validation

### Credential Resolution

When Burrito needs to authenticate to a repository, it follows these steps:

1. First, it looks for a repository-specific credential in the same namespace with an exact URL match.
2. If no repository-specific credential is found, it looks for shared credentials that match the repository URL.
3. For shared credentials, the most specific URL match is selected (e.g., `github.com/org` is more specific than `github.com`).
4. For shared credentials with namespace restrictions, only credentials that allow the repository's namespace will be considered.

### Examples

#### Repository-specific credential using username/password standard git authentication

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: burrito-repo
  namespace: burrito-project
type: credentials.burrito.tf/repository
stringData:
  provider: standard
  url: https://git.example.com/my-org/my-repo
  username: "my-username"
  password: "my-password"
```

#### Repository-specific credential using SSH private key

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: burrito-repo
  namespace: burrito-project
type: credentials.burrito.tf/repository
stringData:
  provider: standard
  url: git@git.example.com:my-org/my-repo
  sshPrivateKey: |
    -----BEGIN OPENSSH PRIVATE KEY-----
    my-private-key
    -----END OPENSSH PRIVATE KEY-----
  webhookSecret: "my-webhook-secret"
```

#### Repository-specific credential for GitHub using Personal Access Token

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: burrito-repo
  namespace: burrito-project
type: credentials.burrito.tf/repository
stringData:
  provider: github
  url: https://github.com/my-org/my-repo
  githubToken: "github_pat_xxx"
  webhookSecret: "my-webhook-secret"
```

#### Repository-specific credential for GitHub using GitHub App

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: burrito-repo
  namespace: burrito-project
type: credentials.burrito.tf/repository
stringData:
  provider: github
  url: https://github.com/my-org/my-repo
  githubAppID: "123456"
  githubAppInstallationID: "12345678"
  githubAppPrivateKey: |
    -----BEGIN RSA PRIVATE KEY-----
    my-private-key
    -----END RSA PRIVATE KEY-----
  webhookSecret: "my-webhook-secret"
```

#### Shared credential for all repositories on gitlab.example.com

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: gitlab-shared-credentials
  namespace: burrito-system
# 👇 by default this secret is shared with all tenants
type: credentials.burrito.tf/shared
stringData:
  provider: gitlab
  url: https://gitlab.example.com
  gitlabToken: "glpat-xxxx"
  webhookSecret: "my-webhook-secret"
```

#### Shared credential for all repos in an organization repositories with namespace restrictions

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: github-org-credentials
  namespace: burrito-system
  annotations:
    # 👇 only repositories with those 3 tenants can use these credentials
    burrito.tf/allowed-tenants: "team-a,team-b,team-c"
type: credentials.burrito.tf/shared
stringData:
  provider: github
  url: https://github.com/my-org
  githubAppID: "123456"
  githubAppInstallationID: "12345678"
  githubAppPrivateKey: |
    -----BEGIN RSA PRIVATE KEY-----
    my-private-key
    -----END RSA PRIVATE KEY-----
  webhookSecret: "my-webhook-secret"
```
