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
  BURRITO_DATASTORE_STORAGE_ENCRYPTION_KEY: <your-encryption-key>
```

!!! warning
    Losing the encryption key will make all encrypted data unrecoverable. Make sure to back up your encryption key securely.

You can generate a key with:

```bash
openssl rand -hex 32
```

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

The different cloud provider implementations rely on the default credentials chain of the cloud provider SDKs. Use annotations and labels on the service account associated to the datastore by updating the `datastore.serviceAccount.metadata` field to specify the credentials to use. (for example the `eks.amazonaws.com/role-arn` annotation with IRSA on EKS, see [Cloud permissions](#cloud-permissions) for the exact configuration of each provider)

## Cloud permissions

The datastore only reads, writes, deletes and lists objects inside the bucket or container that is
configured as the storage backend. It never creates, reconfigures or deletes the bucket itself, so
the permissions below are everything the datastore needs and can be copy-pasted as-is.

The datastore runs with the `burrito-datastore` service account, in the namespace where Burrito is
installed (`burrito-system` by default).

| Backend | Object operations | Bucket / container operations |
| --- | --- | --- |
| AWS S3 | `s3:GetObject`, `s3:PutObject`, `s3:DeleteObject` | `s3:ListBucket` |
| Azure Blob Storage | read, write and delete blobs | list blobs in the container |
| Google Cloud Storage | `storage.objects.get`, `storage.objects.create`, `storage.objects.delete` | `storage.objects.list` |

### AWS S3

On EKS, the recommended way to grant the datastore access to the bucket is
[IRSA](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html):

```hcl
data "aws_iam_policy_document" "burrito_datastore" {
  statement {
    sid       = "ListDatastoreBucket"
    actions   = ["s3:ListBucket"]
    resources = [aws_s3_bucket.datastore.arn]
  }

  statement {
    sid       = "ManageDatastoreObjects"
    actions   = ["s3:GetObject", "s3:PutObject", "s3:DeleteObject"]
    resources = ["${aws_s3_bucket.datastore.arn}/*"]
  }
}

resource "aws_iam_policy" "burrito_datastore" {
  name   = "burrito-datastore"
  policy = data.aws_iam_policy_document.burrito_datastore.json
}

data "aws_eks_cluster" "this" {
  name = "<your-cluster-name>"
}

data "aws_iam_openid_connect_provider" "this" {
  url = data.aws_eks_cluster.this.identity[0].oidc[0].issuer
}

data "aws_iam_policy_document" "burrito_datastore_assume_role" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [data.aws_iam_openid_connect_provider.this.arn]
    }

    condition {
      test     = "StringEquals"
      variable = "${replace(data.aws_eks_cluster.this.identity[0].oidc[0].issuer, "https://", "")}:sub"
      values   = ["system:serviceaccount:burrito-system:burrito-datastore"]
    }

    condition {
      test     = "StringEquals"
      variable = "${replace(data.aws_eks_cluster.this.identity[0].oidc[0].issuer, "https://", "")}:aud"
      values   = ["sts.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "burrito_datastore" {
  name               = "burrito-datastore"
  assume_role_policy = data.aws_iam_policy_document.burrito_datastore_assume_role.json
}

resource "aws_iam_role_policy_attachment" "burrito_datastore" {
  role       = aws_iam_role.burrito_datastore.name
  policy_arn = aws_iam_policy.burrito_datastore.arn
}
```

Then annotate the datastore service account with the role ARN:

```yaml
datastore:
  serviceAccount:
    metadata:
      annotations:
        eks.amazonaws.com/role-arn: arn:aws:iam::<account-id>:role/burrito-datastore
```

!!! note
    If the bucket is encrypted with a customer managed KMS key (SSE-KMS), the policy also needs `kms:Decrypt` and `kms:GenerateDataKey` on that key.

!!! note
    On clusters without IRSA support, provide static credentials instead, as shown in the [Private S3 endpoint](#private-s3-endpoint) section.

### Azure Blob Storage

The datastore authenticates with `DefaultAzureCredential`, which supports managed identities
(including [AKS Workload Identity](https://learn.microsoft.com/en-us/azure/aks/workload-identity-overview)).
The built-in `Storage Blob Data Contributor` role, scoped to the container, contains exactly the
permissions the datastore needs (read, write, delete and list blobs, no control-plane or
account-level permission):

```hcl
resource "azurerm_role_assignment" "burrito_datastore" {
  scope                = azurerm_storage_container.datastore.resource_manager_id
  role_definition_name = "Storage Blob Data Contributor"
  principal_id         = azurerm_user_assigned_identity.burrito_datastore.principal_id
}

resource "azurerm_federated_identity_credential" "burrito_datastore" {
  name                = "burrito-datastore"
  resource_group_name = azurerm_resource_group.this.name
  parent_id           = azurerm_user_assigned_identity.burrito_datastore.id
  audience            = ["api://AzureADTokenExchange"]
  issuer              = azurerm_kubernetes_cluster.this.oidc_issuer_url
  subject             = "system:serviceaccount:burrito-system:burrito-datastore"
}
```

Then reference the managed identity from the datastore pod. The workload identity label is set on
the datastore metadata, which is applied to both the deployment selector and the pod labels:

```yaml
datastore:
  metadata:
    labels:
      azure.workload.identity/use: "true"
  serviceAccount:
    metadata:
      annotations:
        azure.workload.identity/client-id: <managed-identity-client-id>
```

### Google Cloud Storage

The datastore authenticates with [Application Default Credentials](https://cloud.google.com/docs/authentication/application-default-credentials),
which supports [GKE Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity).
The `roles/storage.objectAdmin` role, scoped to the bucket, allows the datastore to get, create,
delete and list objects:

```hcl
resource "google_service_account" "burrito_datastore" {
  account_id   = "burrito-datastore"
  display_name = "Burrito datastore"
}

# Allow the burrito-datastore Kubernetes service account to impersonate the GCP service account
resource "google_service_account_iam_member" "burrito_datastore_workload_identity" {
  service_account_id = google_service_account.burrito_datastore.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${google_container_cluster.this.project}.svc.id.goog[burrito-system/burrito-datastore]"
}

resource "google_storage_bucket_iam_member" "burrito_datastore" {
  bucket = google_storage_bucket.datastore.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.burrito_datastore.email}"
}
```

Then annotate the datastore service account with the GCP service account email:

```yaml
datastore:
  serviceAccount:
    metadata:
      annotations:
        iam.gke.io/gcp-service-account: burrito-datastore@<project-id>.iam.gserviceaccount.com
```

!!! note
    On clusters using Workload Identity Federation for GKE (instead of the service account based binding above), the bucket binding can target the Kubernetes service account directly with `member = "principal://iam.googleapis.com/projects/<project-number>/locations/global/workloadIdentityPools/<project-id>.svc.id.goog/subject/system:serviceaccount:burrito-system:burrito-datastore"`, and the `iam.gke.io/gcp-service-account` annotation is not needed.

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
