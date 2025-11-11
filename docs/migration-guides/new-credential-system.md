# Migrating to the New Credential System

## Overview

Burrito has introduced a new credential system that provides more flexibility and security for managing Git authentication. This guide will help you migrate from the old repository secrets approach to the new credential system.

The new credential system introduces two types of credentials:

1. **Repository-specific credentials** (`credentials.burrito.tf/repository`): For authenticating to a specific repository in a specific namespace
2. **Shared credentials** (`credentials.burrito.tf/shared`): For sharing authentication across multiple repositories and namespaces

## Key Changes

- Secret type has changed from `Opaque` to `credentials.burrito.tf/repository` or `credentials.burrito.tf/shared`
- Field names follow a consistent camelCase format (e.g., `githubAppID` instead of `githubAppId`)
- The `provider` field is now required to specify the Git provider type
- The `url` field is now required to specify the repository URL
- Shared credentials can be restricted to specific namespaces using annotations

## Migration Steps

### Step 1: Identify Existing Secret References

First, identify all your `TerraformRepository` resources that reference secrets in their configuration:

```bash
kubectl get terraformrepositories --all-namespaces -o custom-columns=NAME:.metadata.name,NAMESPACE:.metadata.namespace,SECRET:.spec.repository.secretName
```

### Step 2: Export and Convert Existing Secrets

For each identified secret, export, convert, and reapply it in the new format. Below are conversion examples for each authentication method.

### Step 3: Convert Repository Secrets

#### GitHub App Authentication

**Old format:**

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

**New format:**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: burrito-repo
  namespace: burrito-project
type: credentials.burrito.tf/repository
stringData:
  provider: github
  url: https://github.com/owner/repo  # Add the actual repository URL here
  githubAppID: "123456"  # Note the change from githubAppId to githubAppID
  githubAppInstallationID: "12345678"  # Note the change from githubAppInstallationId to githubAppInstallationID
  githubAppPrivateKey: |
    -----BEGIN RSA PRIVATE KEY-----
    my-private-key
    -----END RSA PRIVATE KEY-----
  webhookSecret: "my-webhook-secret"
```

#### GitHub Token Authentication

**Old format:**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: burrito-repo
  namespace: burrito-project
type: Opaque
stringData:
  githubToken: "github_pat_xxx"
  webhookSecret: "my-webhook-secret"
```

**New format:**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: burrito-repo
  namespace: burrito-project
type: credentials.burrito.tf/repository
stringData:
  provider: github
  url: https://github.com/owner/repo  # Add the actual repository URL here
  githubToken: "github_pat_xxx"
  webhookSecret: "my-webhook-secret"
```

#### GitLab Token Authentication

**Old format:**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: burrito-repo
  namespace: burrito-project
type: Opaque
stringData:
  gitlabToken: "glpat-xxxx"
  webhookSecret: "my-webhook-secret"
```

**New format:**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: burrito-repo
  namespace: burrito-project
type: credentials.burrito.tf/repository
stringData:
  provider: gitlab
  url: https://gitlab.com/owner/repo  # Add the actual repository URL here
  gitlabToken: "glpat-xxxx"
  webhookSecret: "my-webhook-secret"
```

#### Basic Authentication

**Old format:**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: burrito-repo
  namespace: burrito-project
type: Opaque
stringData:
  username: "my-username"
  password: "my-password"
  webhookSecret: "my-webhook-secret"
```

**New format:**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: burrito-repo
  namespace: burrito-project
type: credentials.burrito.tf/repository
stringData:
  provider: standard
  url: https://git.example.com/owner/repo  # Add the actual repository URL here
  username: "my-username"
  password: "my-password"
  webhookSecret: "my-webhook-secret"
```

#### SSH Authentication

**Old format:**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: burrito-repo
  namespace: burrito-project
type: Opaque
stringData:
  sshPrivateKey: |
    -----BEGIN OPENSSH PRIVATE KEY-----
    my-private-key
    -----END OPENSSH PRIVATE KEY-----
  webhookSecret: "my-webhook-secret"
```

**New format:**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: burrito-repo
  namespace: burrito-project
type: credentials.burrito.tf/repository
stringData:
  provider: standard
  url: git@git.example.com:owner/repo  # Add the actual repository URL here
  sshPrivateKey: |
    -----BEGIN OPENSSH PRIVATE KEY-----
    my-private-key
    -----END OPENSSH PRIVATE KEY-----
  webhookSecret: "my-webhook-secret"
```

### Step 4: Update TerraformRepository Resources

In previous versions of Burrito, you needed to explicitly reference the secret name in the `TerraformRepository` resource. With the new credential system, this is no longer necessary as credentials are matched automatically based on URL and namespace.

**Old format:**

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformRepository
metadata:
  name: my-repository
  namespace: burrito-project
spec:
  repository:
    url: https://github.com/owner/repo
    secretName: burrito-repo  # This reference is no longer needed
  terraform:
    enabled: true
```

**New format:**

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

The `secretName` field is no longer required. The credential system will automatically find the appropriate credentials based on:

1. Matching URL in repository-specific credentials in the same namespace
2. Most specific URL match in shared credentials (if no repository-specific credentials are found)

### Step 5: Consider Using Shared Credentials

If you have multiple repositories that use the same credentials, consider migrating to shared credentials for easier management:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: github-shared-credentials
  namespace: burrito-system  # Typically stored in a system namespace
  annotations:
    burrito.tf/allowed-tenants: "team-a,team-b,burrito-project"  # Optional: restrict to specific namespaces
type: credentials.burrito.tf/shared
stringData:
  provider: github
  url: github.com  # Use a common base URL to match multiple repositories
  githubToken: "github_pat_xxx"
  webhookSecret: "my-webhook-secret"
```

## Verification

After migrating your credentials, you can verify that they're correctly configured by:

1. Checking that your secrets have the correct type:

    ```bash
    kubectl get secrets -n your-namespace -o custom-columns=NAME:.metadata.name,TYPE:.type | grep credentials.burrito.tf
    ```

1. Verifying that the credentials are properly detected by Burrito by checking the logs:

    ```bash
    kubectl logs -n your-namespace deployment/burrito-controller -c controller
    ```

1. Ensuring that Git operations such as clone, pull, and webhook events still work properly.

## Troubleshooting

If you encounter issues after migration:

1. **Credentials not found**: Ensure the `url` field in your credential secret matches the repository URL in your `TerraformRepository` resource.

2. **Permissions issues**: If using shared credentials with namespace restrictions, verify that the repository's namespace is included in the `burrito.tf/allowed-tenants` annotation.

3. **Authentication failures**: Double-check that all credential fields have been correctly migrated (pay attention to ID/Id case differences).

4. **Webhook issues**: Ensure the `webhookSecret` is correctly set in your credentials secret and that the webhook is configured with the same secret in your Git provider.

## Additional Resources

For more information on the new credential system, refer to:

- [Git Authentication Documentation](../operator-manual/git-authentication.md)
- [GitHub App Authentication](../operator-manual/git-authentication/github-app.md)
- [GitHub Token Authentication](../operator-manual/git-authentication/github-token.md)
- [GitLab Token Authentication](../operator-manual/git-authentication/gitlab-token.md)
