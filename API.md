# Burrito API Reference

The Burrito API server exposes a REST API for querying repository and layer status, plan results, and logs.

## Base URL

```
http://<burrito-server>:8080/api/v1
```

## Authentication

The API supports configurable authentication methods including:
- No authentication (development)
- Static token-based authentication
- OIDC / OAuth2 integration (see [user authentication docs](docs/operator-manual/user-authentication.md))

Include an `Authorization: Bearer <token>` header when authentication is enabled.

## Endpoints

### Repositories

#### List Repositories

```
GET /api/v1/repositories
```

Returns all `TerraformRepository` resources across all watched namespaces.

**Query Parameters:**
| Parameter  | Type    | Description                                |
|------------|---------|--------------------------------------------|
| `namespace`| string  | Filter by namespace                        |
| `limit`    | integer | Max results per page (default: 50)         |
| `offset`   | integer | Pagination offset                          |

**Response:**
```json
{
  "items": [
    {
      "name": "my-infra-repo",
      "namespace": "default",
      "status": "Ready",
      "lastSync": "2025-01-15T10:30:00Z",
      "layers": 12
    }
  ],
  "total": 1
}
```

#### Get Repository

```
GET /api/v1/repositories/{namespace}/{name}
```

Returns details for a specific repository, including its layers.

**Response (200):**
```json
{
  "name": "my-infra-repo",
  "namespace": "default",
  "url": "https://github.com/org/infra",
  "branch": "main",
  "status": "Ready",
  "layers": [
    {
      "name": "production",
      "path": "environments/prod",
      "status": "UpToDate",
      "lastPlan": "2025-01-15T10:30:00Z"
    }
  ]
}
```

### Layers

#### List Layers

```
GET /api/v1/layers
```

Returns all `TerraformLayer` resources.

**Query Parameters:**
| Parameter      | Type    | Description                              |
|----------------|---------|------------------------------------------|
| `namespace`    | string  | Filter by namespace                      |
| `repository`   | string  | Filter by repository name                |
| `status`       | string  | Filter by layer status                   |
| `limit`        | integer | Max results per page (default: 50)       |
| `offset`       | integer | Pagination offset                        |

**Layer Status Values:**
- `Unknown` — layer hasn't been reconciled yet
- `PlanInProgress` — a plan is currently running
- `ApplyInProgress` — an apply is currently running
- `PlanFailed` — the last plan failed
- `ApplyFailed` — the last apply failed
- `Drifted` — a plan detected drift between code and state
- `UpToDate` — infrastructure matches the desired state

#### Get Layer

```
GET /api/v1/layers/{namespace}/{name}
```

Returns full details for a specific layer.

#### Get Layer Plan

```
GET /api/v1/layers/{namespace}/{name}/plan
```

Returns the latest plan artifact for a layer.

**Response (200):**
```json
{
  "commit": "abc123def456",
  "timestamp": "2025-01-15T10:30:00Z",
  "add": 3,
  "change": 1,
  "destroy": 0,
  "summary": "Plan: 3 to add, 1 to change, 0 to destroy.",
  "logs_url": "/api/v1/layers/default/production/plan/logs"
}
```

#### Get Layer Logs

```
GET /api/v1/layers/{namespace}/{name}/plan/logs
```

Returns the raw Terraform plan output as text/plain.

### Sync (Manual Trigger)

#### Trigger Layer Sync

```
POST /api/v1/layers/{namespace}/{name}/sync
```

Triggers an immediate reconciliation of a layer. The layer must be associated with a repository that has sync enabled.

**Response (202):**
```json
{
  "message": "Sync triggered for layer default/production"
}
```

### Health

#### Health Check

```
GET /healthz
```

Returns `200 OK` when the server is healthy.

```
GET /readyz
```

Returns `200 OK` when the server is ready to serve traffic.

## Error Responses

All errors follow a consistent format:

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Layer 'default/missing' not found"
  }
}
```

**HTTP Status Codes:**
| Status | Meaning                        |
|--------|--------------------------------|
| 200    | Success                        |
| 201    | Created                        |
| 202    | Accepted (async operation)     |
| 400    | Bad Request (invalid input)    |
| 401    | Unauthorized                   |
| 403    | Forbidden                      |
| 404    | Resource not found             |
| 409    | Conflict (resource locked)     |
| 500    | Internal server error          |
| 503    | Service unavailable            |

## Pagination

List endpoints support offset-based pagination:

```
GET /api/v1/layers?limit=20&offset=40
```

The response includes a `total` field for total count:

```json
{
  "items": [...],
  "total": 157,
  "limit": 20,
  "offset": 40
}
```

## Rate Limiting

The API implements token-bucket rate limiting. When exceeded, requests receive `429 Too Many Requests` with a `Retry-After` header.

## Versioning

The API version is embedded in the URL path (`/api/v1/`). Breaking changes will introduce a new version (`/api/v2/`). The current v1 API is considered stable.
