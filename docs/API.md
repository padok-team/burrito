# API Reference — burrito

## Authentication
All API requests require authentication via Bearer token or API key.

```http
Authorization: Bearer <token>
```

## Endpoints

### GET /health
Health check endpoint. Returns 200 when the service is healthy.

**Response**
```json
{"status": "ok", "version": "1.0.0"}
```

### GET /api/v1/items
List all items with optional filtering and pagination.

**Parameters**
| Name | Type | Description |
|------|------|-------------|
| page | int | Page number (default: 1) |
| limit | int | Items per page (default: 20, max: 100) |
| sort | string | Sort field (default: created_at) |
| order | string | Sort order: asc or desc (default: desc) |

**Response**
```json
{"data": [], "pagination": {"page": 1, "limit": 20, "total": 0}}
```

### POST /api/v1/items
Create a new item.

**Request Body**
```json
{"name": "string", "description": "string", "metadata": {}}
```

**Response (201)**
```json
{"id": "uuid", "name": "string", "created_at": "ISO8601"}
```

### GET /api/v1/items/{id}
Retrieve a single item by ID.

**Response**
```json
{"id": "uuid", "name": "string", "description": "string", "metadata": {}, "created_at": "ISO8601", "updated_at": "ISO8601"}
```

### PUT /api/v1/items/{id}
Update an existing item.

### DELETE /api/v1/items/{id}
Delete an item. Returns 204 on success.

## Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | bad_request | Invalid request parameters |
| 401 | unauthorized | Missing or invalid authentication |
| 403 | forbidden | Insufficient permissions |
| 404 | not_found | Resource not found |
| 429 | rate_limited | Too many requests |
| 500 | internal_error | Unexpected server error |

## Rate Limiting
Rate limits are enforced per API key. Headers include:
- `X-RateLimit-Limit`: Maximum requests per window
- `X-RateLimit-Remaining`: Remaining requests
- `X-RateLimit-Reset`: Unix timestamp when the window resets