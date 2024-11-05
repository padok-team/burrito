# Datastore

## Configuration

The Datastore storage backend can be configured using the following yaml configuration :

```yaml
config:
  burrito:
    datastore:
      skipLeadingSlashInKey: <false|true> # default: false
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
