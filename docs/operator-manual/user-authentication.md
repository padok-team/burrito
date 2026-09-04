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
        requiredClaims:
          groups:
            - "burrito-admins"
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
| `requiredClaims`          | Map of claim name to allowed values, used to authorize users (see below) |

## Disabling Authentication

If both Basic Authentication and OIDC are disabled, the Burrito server will be publicly accessible. This may be suitable for development environments or if you have other means of securing access (authentication proxy, VPN, etc.)...

### Authorization

By default, any user able to authenticate with the configured OIDC provider is authorized to
access the Burrito UI. To restrict access, set `requiredClaims` to a map of claim name to the
list of values that satisfy it:

```yaml
server:
  oidc:
    requiredClaims:
      groups:
        - "burrito-admins"
        - "platform-team"
```

A user is authorized only if **every** configured claim has a matching value on their ID
token — a claim is satisfied if its value (a plain string, or an array such as `groups`)
contains at least one of the allowed values. Users whose token doesn't satisfy the required
claims are redirected back to the login page with an error.

This is an all-or-nothing gate on the whole Burrito instance — there is currently no
per-tenant/per-namespace authorization based on claims.
