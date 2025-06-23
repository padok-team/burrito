# Datastore

## Configuration

The Datastore storage backend can be configured using the following yaml configuration :

```yaml
config:
  burrito:
    datastore:
      storage:
        encryption:
          enabled: <false|true> # default: false
        mock: <false|true> # default: false
        s3:
          bucket: <bucket-name>
          usePathStyle: <false|true> # default: false
        gcs:
          bucket: <bucket-name>
        azure:
          storageAccount: <storage-account>
          container: <container-name>
```

!!! info
    Only one storage backend can be configured at a time.

!!! warning
    The `mock` storage backend is only for testing purposes and should not be used in production. If enabled, Burrito will store the data in memory and will lose it when the pod is restarted. It also might fill up the memory of the pod if too much data is stored.

## Encryption

### Configuration

Burrito supports encryption of data at rest in the datastore. When encryption is enabled, all data stored in the backend storage will be encrypted using AES-256-CBC. This allows you to decrypt with external tools such as `openssl`

To enable encryption, you need to:

1. Set `encryption.enabled: true` in the configuration
2. Provide an encryption key via the `BURRITO_DATASTORE_STORAGE_ENCRYPTION_KEY` environment variable, through a Kubernetes secret like below

```yaml
config:
  burrito:
    datastore:
      storage:
        encryption:
          enabled: true

datastore:
  deployment:
    envFrom:
      - secretRef:
          name: burrito-datastore-encryption-key
```

You'll need to create a secret containing the encryption key:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: burrito-datastore-encryption-key
  namespace: <datastoreNamespace>
type: Opaque
stringData:
  BURRITO_DATASTORE_STORAGE_ENCRYPTION_KEY: <your-base64-encoded-encryption-key>
```

!!! warning
    Losing the encryption key will make all encrypted data unrecoverable. Make sure to back up your encryption key securely.

### Security Notes

- Always keep your encryption keys secure
- The IV is stored in plaintext at the beginning of each encrypted file (this is standard practice)
- Each encryption operation uses a random IV, ensuring the same plaintext produces different ciphertext

### Files format

The encrypted files use the following format:
- First 16 bytes: Initialization Vector (IV)
- Remaining bytes: AES-256-CBC encrypted data with PKCS#7 padding

The encryption key is derived by taking the SHA-256 hash of the provided key string.

### Decrypting with OpenSSL

You have downloaded the encrypted file, you can now start to decrypt it:

```bash
# Extract the first 16 bytes (IV) as hex
IV_HEX=$(xxd -l 16 -p plan.json)

# derivate your key with sha256
KEY_HASH=$(echo -n "your-encryption-key" | sha256sum | cut -d' ' -f1)

# decrypt the file
openssl enc -aes-256-cbc -d -in plan.json -K "${KEY_HASH}" -iv "${IV_HEX}"
```

### Encrypting existing files

If you enable encryption on an existing datastore with unencrypted files, you can use the `/encrypt` endpoint to encrypt all existing files. See the [Encrypt Endpoint documentation](encrypt-endpoint.md) for detailed usage instructions.

## Authentication

The different cloud provider implementations rely on the default credentials chain of the cloud provider SDKs. Use annotations and labels on the service account associated to the datastore by updating the `datastore.serviceAccount.metadata` field to specify the credentials to use. (e.g. `iam.amazonaws.com/role` for AWS)

## Authorization

The Datastore relies on TokenReview and mounted volumes for authorization. We rely on a custom audience for the TokenReview to ensure that the token can only be used for the Datastore.

## Object expiration

For now the datastore doesn't delete any object it puts into the storage backend. This is a feature that will be implemented in the future.

## Private S3 endpoint

You can use a private endpoint for S3, like Ceph or Minio. To do so, you'll need to create a secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: s3-secrets
  namespace: <datastoreNamespace>
type: Opaque
stringData:
  AWS_ACCESS_KEY_ID: xxx
  AWS_SECRET_ACCESS_KEY: xxx
  AWS_ENDPOINT_URL_S3: https://s3.domain.com
  AWS_REGION: yourRegion
```

where `<datastoreNamespace>` is the namespace on which datastore is installed (`burrito-system` by default)

In your Helm chart values, you'll also need to tell datastore to use this secret as environment variables:

```yaml
config:
  burrito:
    datastore:
      storage:
        mock: false
        s3:
          bucket: <bucketName>
          usePathStyle: true

datastore:
  deployment:
    envFrom:
      - secretRef:
          name: s3-secrets
```
