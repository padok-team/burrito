# User Authentication

Burrito supports two modes of authentication for its server component:

1. **Basic Authentication (default)**
2. **OpenID Connect (OIDC)**

## SAML authentication is not supported at this time but will be added in the future.

## Basic Authentication (Default)

When OIDC is disabled (`server.oidc.enabled: false`), Burrito falls back to a built-in basic authentication scheme. This mode is **not recommended for production**.

### Configuration

```yaml
server:
  oidc:
    enabled: false
```

### Credentials

- **Username:** `admin`
- **Password:** Stored in the Kubernetes Secret `burrito-admin-credentials`.

Retrieve the password with:

```bash
kubectl -n <burrito-namespace> get secret burrito-admin-credentials \
  -o jsonpath="{.data.password}" | base64 --decode
```

Use `admin` and the decoded password to log in to the Burrito server.

---

## OpenID Connect (OIDC) Authentication

Enable OIDC to integrate Burrito with your identity provider. This is the recommended approach for production environments.

### Configuration

```yaml
server:
  oidc:
    enabled: true # Enable OIDC
    issuerUrl: <OIDC_ISSUER> # e.g. https://accounts.example.com
    clientId: <CLIENT_ID>
    clientSecret:
      secretName: "burrito-oidc-secret"
      secretKey: "clientSecret"
    redirectUrl: "https://<your-domain>/auth/callback"
    scopes:
      - "openid"
      - "profile"
      - "email"
```

| Field                     | Description                                                              |
| ------------------------- | ------------------------------------------------------------------------ |
| `enabled`                 | Turn OIDC on or off                                                      |
| `issuerUrl`               | Base URL of your OIDC provider                                           |
| `clientId`                | Registered client ID                                                     |
| `clientSecret.secretName` | Kubernetes Secret containing your OIDC client secret                     |
| `clientSecret.secretKey`  | Key within the Secret that holds the client secret value                 |
| `redirectUrl`             | Callback URL for OIDC (must match the one registered with your provider) |
| `scopes`                  | OIDC scopes to request                                                   |

### Creating the Client Secret

```bash
kubectl -n <burrito-namespace> create secret generic burrito-oidc-secret \
  --from-literal=clientSecret=<YOUR_SECRET>
```

### Authorization

For the moment, Burrito does not implement authorization mechanisms. All users that are able to authenticate with the configured OIDC provider will be able to access the Burrito UI.
