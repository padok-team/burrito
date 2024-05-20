# Datastore

## Configuration

The Datastore storage backend can be configured using the following yaml configuration :

```yaml
config:
  burrito:
    datastore:
      storage:
        s3:
          bucket: XXX
        gcs:
          bucket: XXX
        azure:
          storageAccount: XXX
          container: XXX
```

!!! info
    Only one storage backend can be configured at a time.

## Authentication

The different cloud provider implementations rely on the default credentials chain of the cloud provider SDKs.

## Authorization

The Datastore relies on TokenReview and mounted volumes for authorization. We rely on a custom audience for the TokenReview to ensure that the token can only be used for the Datastore.

## Object expiration

For now the datastore doesn't delete any object it puts into the storage backend. This is a feature that will be implemented in the future.
