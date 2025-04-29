# Datastore

## Configuration

The Datastore storage backend can be configured using the following yaml configuration :

```yaml
config:
  burrito:
    datastore:
      storage:
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
