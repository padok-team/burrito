# User Authentication

Burrito supports two modes of authentication for its server component:

1. **Basic Authentication (default)**
2. **OpenID Connect (OIDC)**

SAML authentication is not supported at this time but will be added in the future

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

OIDC configuration requires setting up a client in your OIDC provider. You will need the following details:

- **Issuer URL**
- **Client ID**
- **Client Secret**
- **Redirect URL** (should be `https://<your-domain>/auth/callback`)
- **Scopes** (typically `openid`, `profile`, and `email`)

The client secret must be stored in a Kubernetes Secret and referenced in the deployment environment variables.
The environment variable name for the client secret must be `BURRITO_SERVER_OIDC_CLIENTSECRET`.

```yaml
config:
  burrito:
    server:
      oidc:
        enabled: true # Enable OIDC
        issuerUrl: <OIDC_ISSUER> # e.g. https://accounts.example.com
        clientId: <CLIENT_ID>
        redirectUrl: "https://<your-domain>/auth/callback"
        scopes:
          - "openid"
          - "profile"
          - "email"
...
server:
  deployment:
    envFrom:
      - secretRef:
          name: burrito-oidc-client-secret
```

| Field                     | Description                                                              |
| ------------------------- | ------------------------------------------------------------------------ |
| `enabled`                 | Turn OIDC on or off                                                      |
| `issuerUrl`               | Base URL of your OIDC provider                                           |
| `clientId`                | Registered client ID                                                     |
| `redirectUrl`             | Callback URL for OIDC (must match the one registered with your provider) |
| `scopes`                  | OIDC scopes to request                                                   |

## Disabling Authentication

If both Basic Authentication and OIDC are disabled, the Burrito server will be publicly accessible. This may be suitable for development environments or if you have other means of securing access (authentication proxy, VPN, etc.)...

### Authorization

For the moment, Burrito does not implement authorization mechanisms. All users that are able to authenticate with the configured OIDC provider will be able to access the Burrito UI.
