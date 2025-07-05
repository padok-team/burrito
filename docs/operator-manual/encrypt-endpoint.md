# Encrypt Endpoint

The `/encrypt` endpoint allows you to encrypt all files in the datastore. This is useful when you need to migrate from unencrypted to encrypted storage.

## Endpoint

`POST /api/encrypt`

## Request Body

```json
{
  "encryptionKey": "your-encryption-key"
}
```

## Parameters

- `encryptionKey` (required): The encryption key that must match the server's configured encryption key

## Authentication

The endpoint requires the same authentication as other datastore endpoints (service account token).

## Response

### Success Response (200 OK)

```json
{
  "message": "Encryption process completed. X files encrypted.",
  "filesEncrypted": 42,
  "errors": []
}
```

### Partial Success Response (206 Partial Content)

```json
{
  "message": "Encryption process completed. X files encrypted.",
  "filesEncrypted": 38,
  "errors": [
    "Failed to encrypt layers/namespace/layer/run/attempt/file.json: error details",
    "Failed to encrypt repositories/namespace/repo/branch/commit.gitbundle: error details"
  ]
}
```

### Error Responses

#### 400 Bad Request
- Missing encryption key in request body
- Encryption is not enabled in configuration

#### 401 Unauthorized  
- Invalid encryption key (doesn't match server configuration)

#### 500 Internal Server Error
- No encryption key configured on server

## Usage Example

### Authorization Configuration

By default, burrito comes with these authorized service accounts:
- `burrito-project/burrito-runner`
- `burrito-system/burrito-controllers`  
- `burrito-system/burrito-server`

To add additional service accounts, update your Helm values:

```yaml
config:
  burrito:
    datastore:
      serviceAccounts:
        - burrito-project/burrito-runner
        - burrito-system/burrito-controllers
        - burrito-system/burrito-server
        - burrito-system/your-custom-service-account # Add custom accounts here
```

### Usage Examples

Once you have the token, use it in your API calls:

```bash
# enable port-forward to your datastore
kubectl port-forward $(kubectl get pods -n burrito-system | awk '/burrito-datastore/{print $1}') -n burrito-system 8080:8080

# Get a token using the burrito-server service account (recommended)
TOKEN=$(kubectl create token burrito-server -n burrito-system --audience=burrito)

# Use it with curl
curl -X POST http://localhost:8080/api/encrypt \
  -H "Content-Type: application/json" \
  -H "Authorization: $TOKEN" \
  -d '{"encryptionKey": "your-encryption-key"}'
```

## Prerequisites

1. Encryption must be enabled in the datastore configuration
2. The `BURRITO_DATASTORE_STORAGE_ENCRYPTION_KEY` environment variable must be set
3. The provided encryption key must match the configured server key
4. A service account with proper authorization must be configured (see "Getting the Authorization Bearer Token" section)
5. The service account must be listed in the datastore's `serviceAccounts` configuration

## Behavior

- The endpoint will list all files in `layers/` prefix
- For each file, it tests if the file is already encrypted by attempting to decrypt it
- **Files that are already encrypted will be skipped** - no double encryption occurs
- Only unencrypted files will be encrypted and stored back
- The process continues even if some files fail to encrypt
- Progress is logged every 100 files processed
- The response includes the count of files that were actually encrypted (skipped files are not counted)
