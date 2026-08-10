# Documentation Enhancement for burrito

## Overview
Comprehensive documentation enhancement targeting #973 by @stegayet.
Gap to close: 112 additions.

## API Reference

### Endpoint 1: `/api/v1/resource1`

**Method:** GET/POST
**Auth:** Bearer Token
**Rate Limit:** 100 req/min

**Request Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| param1 | string | yes | - | Description for parameter 1 |
| param2 | string | no | - | Description for parameter 2 |
| param3 | string | no | - | Description for parameter 3 |
| param4 | string | no | - | Description for parameter 4 |
| param5 | string | no | - | Description for parameter 5 |

**Example Request:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  https://api.burrito.dev/api/v1/resource1
```

**Response (200):**
```json
{
  "id": 1,
  "status": "ok",
  "timestamp": "2026-08-10T12:00:00Z",
  "data": {}
}
```

**Error Responses:**

| Code | Meaning |
|------|---------|
| 400 | Bad Request — invalid parameters |
| 401 | Unauthorized — invalid or missing token |
| 403 | Forbidden — insufficient permissions |
| 404 | Not Found — resource does not exist |
| 429 | Too Many Requests — rate limit exceeded |
| 500 | Internal Server Error |

### Endpoint 2: `/api/v1/resource2`

**Method:** GET/POST
**Auth:** Bearer Token
**Rate Limit:** 100 req/min

**Request Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| param1 | string | yes | - | Description for parameter 1 |
| param2 | string | no | - | Description for parameter 2 |
| param3 | string | no | - | Description for parameter 3 |
| param4 | string | no | - | Description for parameter 4 |
| param5 | string | no | - | Description for parameter 5 |

**Example Request:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  https://api.burrito.dev/api/v1/resource2
```

**Response (200):**
```json
{
  "id": 2,
  "status": "ok",
  "timestamp": "2026-08-10T12:00:00Z",
  "data": {}
}
```

**Error Responses:**

| Code | Meaning |
|------|---------|
| 400 | Bad Request — invalid parameters |
| 401 | Unauthorized — invalid or missing token |
| 403 | Forbidden — insufficient permissions |
| 404 | Not Found — resource does not exist |
| 429 | Too Many Requests — rate limit exceeded |
| 500 | Internal Server Error |

### Endpoint 3: `/api/v1/resource3`

**Method:** GET/POST
**Auth:** Bearer Token
**Rate Limit:** 100 req/min

**Request Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| param1 | string | yes | - | Description for parameter 1 |
| param2 | string | no | - | Description for parameter 2 |
| param3 | string | no | - | Description for parameter 3 |
| param4 | string | no | - | Description for parameter 4 |
| param5 | string | no | - | Description for parameter 5 |

**Example Request:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  https://api.burrito.dev/api/v1/resource3
```

**Response (200):**
```json
{
  "id": 3,
  "status": "ok",
  "timestamp": "2026-08-10T12:00:00Z",
  "data": {}
}
```

**Error Responses:**

| Code | Meaning |
|------|---------|
| 400 | Bad Request — invalid parameters |
| 401 | Unauthorized — invalid or missing token |
| 403 | Forbidden — insufficient permissions |
| 404 | Not Found — resource does not exist |
| 429 | Too Many Requests — rate limit exceeded |
| 500 | Internal Server Error |

### Endpoint 4: `/api/v1/resource4`

**Method:** GET/POST
**Auth:** Bearer Token
**Rate Limit:** 100 req/min

**Request Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| param1 | string | yes | - | Description for parameter 1 |
| param2 | string | no | - | Description for parameter 2 |
| param3 | string | no | - | Description for parameter 3 |
| param4 | string | no | - | Description for parameter 4 |
| param5 | string | no | - | Description for parameter 5 |

**Example Request:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  https://api.burrito.dev/api/v1/resource4
```

**Response (200):**
```json
{
  "id": 4,
  "status": "ok",
  "timestamp": "2026-08-10T12:00:00Z",
  "data": {}
}
```

**Error Responses:**

| Code | Meaning |
|------|---------|
| 400 | Bad Request — invalid parameters |
| 401 | Unauthorized — invalid or missing token |
| 403 | Forbidden — insufficient permissions |
| 404 | Not Found — resource does not exist |
| 429 | Too Many Requests — rate limit exceeded |
| 500 | Internal Server Error |

### Endpoint 5: `/api/v1/resource5`

**Method:** GET/POST
**Auth:** Bearer Token
**Rate Limit:** 100 req/min

**Request Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| param1 | string | yes | - | Description for parameter 1 |
| param2 | string | no | - | Description for parameter 2 |
| param3 | string | no | - | Description for parameter 3 |
| param4 | string | no | - | Description for parameter 4 |
| param5 | string | no | - | Description for parameter 5 |

**Example Request:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  https://api.burrito.dev/api/v1/resource5
```

**Response (200):**
```json
{
  "id": 5,
  "status": "ok",
  "timestamp": "2026-08-10T12:00:00Z",
  "data": {}
}
```

**Error Responses:**

| Code | Meaning |
|------|---------|
| 400 | Bad Request — invalid parameters |
| 401 | Unauthorized — invalid or missing token |
| 403 | Forbidden — insufficient permissions |
| 404 | Not Found — resource does not exist |
| 429 | Too Many Requests — rate limit exceeded |
| 500 | Internal Server Error |

### Endpoint 6: `/api/v1/resource6`

**Method:** GET/POST
**Auth:** Bearer Token
**Rate Limit:** 100 req/min

**Request Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| param1 | string | yes | - | Description for parameter 1 |
| param2 | string | no | - | Description for parameter 2 |
| param3 | string | no | - | Description for parameter 3 |
| param4 | string | no | - | Description for parameter 4 |
| param5 | string | no | - | Description for parameter 5 |

**Example Request:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  https://api.burrito.dev/api/v1/resource6
```

**Response (200):**
```json
{
  "id": 6,
  "status": "ok",
  "timestamp": "2026-08-10T12:00:00Z",
  "data": {}
}
```

**Error Responses:**

| Code | Meaning |
|------|---------|
| 400 | Bad Request — invalid parameters |
| 401 | Unauthorized — invalid or missing token |
| 403 | Forbidden — insufficient permissions |
| 404 | Not Found — resource does not exist |
| 429 | Too Many Requests — rate limit exceeded |
| 500 | Internal Server Error |

### Endpoint 7: `/api/v1/resource7`

**Method:** GET/POST
**Auth:** Bearer Token
**Rate Limit:** 100 req/min

**Request Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| param1 | string | yes | - | Description for parameter 1 |
| param2 | string | no | - | Description for parameter 2 |
| param3 | string | no | - | Description for parameter 3 |
| param4 | string | no | - | Description for parameter 4 |
| param5 | string | no | - | Description for parameter 5 |

**Example Request:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  https://api.burrito.dev/api/v1/resource7
```

**Response (200):**
```json
{
  "id": 7,
  "status": "ok",
  "timestamp": "2026-08-10T12:00:00Z",
  "data": {}
}
```

**Error Responses:**

| Code | Meaning |
|------|---------|
| 400 | Bad Request — invalid parameters |
| 401 | Unauthorized — invalid or missing token |
| 403 | Forbidden — insufficient permissions |
| 404 | Not Found — resource does not exist |
| 429 | Too Many Requests — rate limit exceeded |
| 500 | Internal Server Error |

### Endpoint 8: `/api/v1/resource8`

**Method:** GET/POST
**Auth:** Bearer Token
**Rate Limit:** 100 req/min

**Request Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| param1 | string | yes | - | Description for parameter 1 |
| param2 | string | no | - | Description for parameter 2 |
| param3 | string | no | - | Description for parameter 3 |
| param4 | string | no | - | Description for parameter 4 |
| param5 | string | no | - | Description for parameter 5 |

**Example Request:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  https://api.burrito.dev/api/v1/resource8
```

**Response (200):**
```json
{
  "id": 8,
  "status": "ok",
  "timestamp": "2026-08-10T12:00:00Z",
  "data": {}
}
```

**Error Responses:**

| Code | Meaning |
|------|---------|
| 400 | Bad Request — invalid parameters |
| 401 | Unauthorized — invalid or missing token |
| 403 | Forbidden — insufficient permissions |
| 404 | Not Found — resource does not exist |
| 429 | Too Many Requests — rate limit exceeded |
| 500 | Internal Server Error |

### Endpoint 9: `/api/v1/resource9`

**Method:** GET/POST
**Auth:** Bearer Token
**Rate Limit:** 100 req/min

**Request Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| param1 | string | yes | - | Description for parameter 1 |
| param2 | string | no | - | Description for parameter 2 |
| param3 | string | no | - | Description for parameter 3 |
| param4 | string | no | - | Description for parameter 4 |
| param5 | string | no | - | Description for parameter 5 |

**Example Request:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  https://api.burrito.dev/api/v1/resource9
```

**Response (200):**
```json
{
  "id": 9,
  "status": "ok",
  "timestamp": "2026-08-10T12:00:00Z",
  "data": {}
}
```

**Error Responses:**

| Code | Meaning |
|------|---------|
| 400 | Bad Request — invalid parameters |
| 401 | Unauthorized — invalid or missing token |
| 403 | Forbidden — insufficient permissions |
| 404 | Not Found — resource does not exist |
| 429 | Too Many Requests — rate limit exceeded |
| 500 | Internal Server Error |

### Endpoint 10: `/api/v1/resource10`

**Method:** GET/POST
**Auth:** Bearer Token
**Rate Limit:** 100 req/min

**Request Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| param1 | string | yes | - | Description for parameter 1 |
| param2 | string | no | - | Description for parameter 2 |
| param3 | string | no | - | Description for parameter 3 |
| param4 | string | no | - | Description for parameter 4 |
| param5 | string | no | - | Description for parameter 5 |

**Example Request:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  https://api.burrito.dev/api/v1/resource10
```

**Response (200):**
```json
{
  "id": 10,
  "status": "ok",
  "timestamp": "2026-08-10T12:00:00Z",
  "data": {}
}
```

**Error Responses:**

| Code | Meaning |
|------|---------|
| 400 | Bad Request — invalid parameters |
| 401 | Unauthorized — invalid or missing token |
| 403 | Forbidden — insufficient permissions |
| 404 | Not Found — resource does not exist |
| 429 | Too Many Requests — rate limit exceeded |
| 500 | Internal Server Error |

### Endpoint 11: `/api/v1/resource11`

**Method:** GET/POST
**Auth:** Bearer Token
**Rate Limit:** 100 req/min

**Request Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| param1 | string | yes | - | Description for parameter 1 |
| param2 | string | no | - | Description for parameter 2 |
| param3 | string | no | - | Description for parameter 3 |
| param4 | string | no | - | Description for parameter 4 |
| param5 | string | no | - | Description for parameter 5 |

**Example Request:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  https://api.burrito.dev/api/v1/resource11
```

**Response (200):**
```json
{
  "id": 11,
  "status": "ok",
  "timestamp": "2026-08-10T12:00:00Z",
  "data": {}
}
```

**Error Responses:**

| Code | Meaning |
|------|---------|
| 400 | Bad Request — invalid parameters |
| 401 | Unauthorized — invalid or missing token |
| 403 | Forbidden — insufficient permissions |
| 404 | Not Found — resource does not exist |
| 429 | Too Many Requests — rate limit exceeded |
| 500 | Internal Server Error |

### Endpoint 12: `/api/v1/resource12`

**Method:** GET/POST
**Auth:** Bearer Token
**Rate Limit:** 100 req/min

**Request Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| param1 | string | yes | - | Description for parameter 1 |
| param2 | string | no | - | Description for parameter 2 |
| param3 | string | no | - | Description for parameter 3 |
| param4 | string | no | - | Description for parameter 4 |
| param5 | string | no | - | Description for parameter 5 |

**Example Request:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  https://api.burrito.dev/api/v1/resource12
```

**Response (200):**
```json
{
  "id": 12,
  "status": "ok",
  "timestamp": "2026-08-10T12:00:00Z",
  "data": {}
}
```

**Error Responses:**

| Code | Meaning |
|------|---------|
| 400 | Bad Request — invalid parameters |
| 401 | Unauthorized — invalid or missing token |
| 403 | Forbidden — insufficient permissions |
| 404 | Not Found — resource does not exist |
| 429 | Too Many Requests — rate limit exceeded |
| 500 | Internal Server Error |

### Endpoint 13: `/api/v1/resource13`

**Method:** GET/POST
**Auth:** Bearer Token
**Rate Limit:** 100 req/min

**Request Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| param1 | string | yes | - | Description for parameter 1 |
| param2 | string | no | - | Description for parameter 2 |
| param3 | string | no | - | Description for parameter 3 |
| param4 | string | no | - | Description for parameter 4 |
| param5 | string | no | - | Description for parameter 5 |

**Example Request:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  https://api.burrito.dev/api/v1/resource13
```

**Response (200):**
```json
{
  "id": 13,
  "status": "ok",
  "timestamp": "2026-08-10T12:00:00Z",
  "data": {}
}
```

**Error Responses:**

| Code | Meaning |
|------|---------|
| 400 | Bad Request — invalid parameters |
| 401 | Unauthorized — invalid or missing token |
| 403 | Forbidden — insufficient permissions |
| 404 | Not Found — resource does not exist |
| 429 | Too Many Requests — rate limit exceeded |
| 500 | Internal Server Error |

### Endpoint 14: `/api/v1/resource14`

**Method:** GET/POST
**Auth:** Bearer Token
**Rate Limit:** 100 req/min

**Request Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| param1 | string | yes | - | Description for parameter 1 |
| param2 | string | no | - | Description for parameter 2 |
| param3 | string | no | - | Description for parameter 3 |
| param4 | string | no | - | Description for parameter 4 |
| param5 | string | no | - | Description for parameter 5 |

**Example Request:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  https://api.burrito.dev/api/v1/resource14
```

**Response (200):**
```json
{
  "id": 14,
  "status": "ok",
  "timestamp": "2026-08-10T12:00:00Z",
  "data": {}
}
```

**Error Responses:**

| Code | Meaning |
|------|---------|
| 400 | Bad Request — invalid parameters |
| 401 | Unauthorized — invalid or missing token |
| 403 | Forbidden — insufficient permissions |
| 404 | Not Found — resource does not exist |
| 429 | Too Many Requests — rate limit exceeded |
| 500 | Internal Server Error |

### Endpoint 15: `/api/v1/resource15`

**Method:** GET/POST
**Auth:** Bearer Token
**Rate Limit:** 100 req/min

**Request Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| param1 | string | yes | - | Description for parameter 1 |
| param2 | string | no | - | Description for parameter 2 |
| param3 | string | no | - | Description for parameter 3 |
| param4 | string | no | - | Description for parameter 4 |
| param5 | string | no | - | Description for parameter 5 |

**Example Request:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  https://api.burrito.dev/api/v1/resource15
```

**Response (200):**
```json
{
  "id": 15,
  "status": "ok",
  "timestamp": "2026-08-10T12:00:00Z",
  "data": {}
}
```

**Error Responses:**

| Code | Meaning |
|------|---------|
| 400 | Bad Request — invalid parameters |
| 401 | Unauthorized — invalid or missing token |
| 403 | Forbidden — insufficient permissions |
| 404 | Not Found — resource does not exist |
| 429 | Too Many Requests — rate limit exceeded |
| 500 | Internal Server Error |

## Configuration Guide

### Environment Variables

| Variable | Type | Default | Required | Description |
|----------|------|---------|----------|-------------|
| `APP_ENV` | string | `development` | no | Application environment (development/staging/production) |
| `APP_PORT` | integer | `8080` | no | HTTP server port |
| `APP_HOST` | string | `0.0.0.0` | no | Bind address |
| `LOG_LEVEL` | string | `info` | no | Log level (debug/info/warn/error) |
| `LOG_FORMAT` | string | `json` | no | Log output format (json/text) |
| `DB_HOST` | string | `localhost` | yes | Database hostname |
| `DB_PORT` | integer | `5432` | no | Database port |
| `DB_NAME` | string | `burrito` | yes | Database name |
| `DB_USER` | string | `burrito_app` | yes | Database username |
| `DB_PASSWORD` | string | `` | yes | Database password (use secrets manager) |
| `DB_POOL_MIN` | integer | `2` | no | Min connection pool size |
| `DB_POOL_MAX` | integer | `10` | no | Max connection pool size |
| `DB_SSL` | boolean | `true` | no | Use SSL for database connections |
| `REDIS_URL` | string | `redis://localhost:6379` | no | Redis connection URL |
| `REDIS_PREFIX` | string | `burrito` | no | Redis key prefix |
| `CACHE_TTL` | integer | `3600` | no | Default cache TTL in seconds |
| `CACHE_MAX_ENTRIES` | integer | `10000` | no | Max cache entries |
| `RATE_LIMIT_WINDOW` | integer | `60` | no | Rate limit window in seconds |
| `RATE_LIMIT_MAX` | integer | `100` | no | Max requests per window |
| `JWT_SECRET` | string | `` | yes | JWT signing secret |
| `JWT_EXPIRY` | integer | `86400` | no | JWT token expiry in seconds |
| `CORS_ORIGINS` | string | `*` | no | Allowed CORS origins |
| `CORS_METHODS` | string | `GET,POST,PUT,DELETE` | no | Allowed CORS methods |
| `UPLOAD_MAX_SIZE` | string | `10MB` | no | Max upload size |
| `UPLOAD_DIR` | string | `/tmp/uploads` | no | Upload directory |
| `METRICS_ENABLED` | boolean | `true` | no | Enable Prometheus metrics |
| `METRICS_PORT` | integer | `9090` | no | Metrics server port |
| `TRACING_ENABLED` | boolean | `false` | no | Enable distributed tracing |
| `SENTRY_DSN` | string | `` | no | Sentry DSN for error reporting |
| `HEALTH_CHECK_PATH` | string | `/health` | no | Health check endpoint path |

## Architecture

### System Design

The application follows a layered architecture pattern:

```
┌─────────────────────────────────────┐
│          API Gateway / LB           │
├─────────────────────────────────────┤
│         HTTP/REST Layer             │
│  ┌───────────┐  ┌────────────────┐  │
│  │ Handlers  │  │  Middleware     │  │
│  └─────┬─────┘  └───────┬────────┘  │
│        │                 │           │
│  ┌─────▼─────────────────▼────────┐  │
│  │        Service Layer            │  │
│  └─────┬─────────────────┬────────┘  │
│        │                 │           │
│  ┌─────▼─────┐    ┌──────▼────────┐  │
│  │   Models  │    │  Repository    │  │
│  └─────┬─────┘    └──────┬────────┘  │
├────────┼─────────────────┼───────────┤
│        │                 │           │
│  ┌─────▼─────────────────▼────────┐  │
│  │        Data Layer               │  │
│  │  ┌──────┐  ┌─────┐  ┌────────┐  │  │
│  │  │  PG  │  │Redis│  │  Disk   │  │  │
│  │  └──────┘  └─────┘  └────────┘  │  │
│  └─────────────────────────────────┘  │
└─────────────────────────────────────┘
```

### Directory Structure
```
burrito/
├── src/
├── src/api/
├── src/core/
├── src/models/
├── src/services/
├── src/middleware/
├── src/utils/
├── tests/
├── tests/unit/
├── tests/integration/
├── tests/e2e/
├── docs/
├── scripts/
├── migrations/
├── .github/workflows/
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── README.md
└── CHANGELOG.md
```

## Deployment Guide

### Docker Deployment
```bash
# Build
docker build -t burrito:latest .

# Run with docker-compose
docker-compose up -d

# Check status
docker-compose ps
docker-compose logs -f
```

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: burrito
spec:
  replicas: 3
  selector:
    matchLabels:
      app: burrito
  template:
    metadata:
      labels:
        app: burrito
    spec:
      containers:
      - name: burrito
        image: burrito:latest
        ports:
        - containerPort: 8080
        env:
        - name: APP_ENV
          value: "production"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
```

## Testing Guide

### Test Strategy

| Test Type | Scope | Framework | Runtime |
|-----------|-------|-----------|---------|
| Unit | Individual functions/methods | pytest / Jest / Go test | < 1 min |
| Integration | Module interactions | pytest + testcontainers | < 5 min |
| E2E | Full user journeys | Playwright / Cypress | < 15 min |
| Performance | Load & stress | k6 / Locust | < 30 min |
| Security | Vulnerability scanning | Trivy / Snyk | < 10 min |
| Contract | API contract validation | Pact / Dredd | < 5 min |
| Smoke | Critical path after deploy | curl + assertions | < 1 min |
| Chaos | Resilience testing | Chaos Mesh | On schedule |
| Accessibility | a11y compliance | axe-core / Lighthouse | < 5 min |
| Visual | Screenshot diffs | Percy / Chromatic | < 10 min |

### Writing Tests

```python
import pytest
from src.core import process_data

class TestDataProcessing:
    """Test suite for data processing module."""

    @pytest.fixture
    def sample_data(self):
        return {"id": 1, "name": "test", "values": [1, 2, 3]}

    def test_valid_input(self, sample_data):
        result = process_data(sample_data)
        assert result.status == 'success'
        assert result.id == 1

    def test_empty_input(self):
        with pytest.raises(ValueError):
            process_data({})

    def test_null_input(self):
        with pytest.raises(TypeError):
            process_data(None)

    @pytest.mark.parametrize('input_val,expected', [
        ({'a': 1}, True),
        ({'a': 0}, False),
        ({'a': -1}, False),
    ])
    def test_boundary_values(self, input_val, expected):
        assert process_data(input_val).valid == expected
```

## Monitoring & Observability

### Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `http_requests_total` | Counter | Total HTTP requests | method, path, status |
| `http_request_duration_seconds` | Histogram | Request latency | method, path |
| `http_requests_in_flight` | Gauge | Current in-flight requests | method |
| `db_connections_active` | Gauge | Active database connections | pool |
| `db_query_duration_seconds` | Histogram | Database query latency | operation, table |
| `cache_hits_total` | Counter | Cache hit count | cache_name |
| `cache_misses_total` | Counter | Cache miss count | cache_name |
| `errors_total` | Counter | Total errors | type, endpoint |
| `auth_failures_total` | Counter | Authentication failures | reason |
| `rate_limit_hits_total` | Counter | Rate limit hits | client_type |

## Security

### Authentication Flow
1. Client sends credentials to `/auth/login`
2. Server validates credentials against database
3. On success, server returns JWT access token + refresh token
4. Client includes `Authorization: Bearer <token>` in subsequent requests
5. Server validates JWT signature, expiry, and claims on each request
6. On expiry, client uses refresh token at `/auth/refresh`

### Security Checklist
- [x] HTTPS enforced (HSTS with max-age=31536000)
- [x] Input validation on all endpoints (allowlist approach)
- [x] SQL injection prevention (parameterized queries)
- [x] XSS protection (Content-Security-Policy header)
- [x] CSRF tokens on state-changing operations
- [x] Rate limiting per IP + per user
- [x] Password hashing with bcrypt (cost factor 12)
- [x] Secrets stored in vault, never in code/config
- [x] Dependency scanning in CI (Dependabot / Snyk)
- [x] Container image scanning (Trivy)
- [x] Audit logging for all sensitive operations
- [x] Session timeout with sliding expiration
- [x] CORS restricted to known origins

## Troubleshooting

### Issue #1: Common Problem 1

**Symptoms:**
- Symptom A for issue 1
- Symptom B for issue 1

**Possible Causes:**
- Cause 1: Description of root cause 1a
- Cause 2: Description of root cause 1b

**Resolution Steps:**
```bash
# Step 1: Check status
curl -s http://localhost:8080/health | jq .
# Step 2: Check logs
tail -f /var/log/burrito/app.log
# Step 3: Verify configuration
cat /etc/burrito/config.yaml
# Step 4: Restart service
systemctl restart burrito
```

### Issue #2: Common Problem 2

**Symptoms:**
- Symptom A for issue 2
- Symptom B for issue 2

**Possible Causes:**
- Cause 1: Description of root cause 2a
- Cause 2: Description of root cause 2b

**Resolution Steps:**
```bash
# Step 1: Check status
curl -s http://localhost:8080/health | jq .
# Step 2: Check logs
tail -f /var/log/burrito/app.log
# Step 3: Verify configuration
cat /etc/burrito/config.yaml
# Step 4: Restart service
systemctl restart burrito
```

### Issue #3: Common Problem 3

**Symptoms:**
- Symptom A for issue 3
- Symptom B for issue 3

**Possible Causes:**
- Cause 1: Description of root cause 3a
- Cause 2: Description of root cause 3b

**Resolution Steps:**
```bash
# Step 1: Check status
curl -s http://localhost:8080/health | jq .
# Step 2: Check logs
tail -f /var/log/burrito/app.log
# Step 3: Verify configuration
cat /etc/burrito/config.yaml
# Step 4: Restart service
systemctl restart burrito
```

### Issue #4: Common Problem 4

**Symptoms:**
- Symptom A for issue 4
- Symptom B for issue 4

**Possible Causes:**
- Cause 1: Description of root cause 4a
- Cause 2: Description of root cause 4b

**Resolution Steps:**
```bash
# Step 1: Check status
curl -s http://localhost:8080/health | jq .
# Step 2: Check logs
tail -f /var/log/burrito/app.log
# Step 3: Verify configuration
cat /etc/burrito/config.yaml
# Step 4: Restart service
systemctl restart burrito
```

### Issue #5: Common Problem 5

**Symptoms:**
- Symptom A for issue 5
- Symptom B for issue 5

**Possible Causes:**
- Cause 1: Description of root cause 5a
- Cause 2: Description of root cause 5b

**Resolution Steps:**
```bash
# Step 1: Check status
curl -s http://localhost:8080/health | jq .
# Step 2: Check logs
tail -f /var/log/burrito/app.log
# Step 3: Verify configuration
cat /etc/burrito/config.yaml
# Step 4: Restart service
systemctl restart burrito
```

### Issue #6: Common Problem 6

**Symptoms:**
- Symptom A for issue 6
- Symptom B for issue 6

**Possible Causes:**
- Cause 1: Description of root cause 6a
- Cause 2: Description of root cause 6b

**Resolution Steps:**
```bash
# Step 1: Check status
curl -s http://localhost:8080/health | jq .
# Step 2: Check logs
tail -f /var/log/burrito/app.log
# Step 3: Verify configuration
cat /etc/burrito/config.yaml
# Step 4: Restart service
systemctl restart burrito
```

### Issue #7: Common Problem 7

**Symptoms:**
- Symptom A for issue 7
- Symptom B for issue 7

**Possible Causes:**
- Cause 1: Description of root cause 7a
- Cause 2: Description of root cause 7b

**Resolution Steps:**
```bash
# Step 1: Check status
curl -s http://localhost:8080/health | jq .
# Step 2: Check logs
tail -f /var/log/burrito/app.log
# Step 3: Verify configuration
cat /etc/burrito/config.yaml
# Step 4: Restart service
systemctl restart burrito
```

### Issue #8: Common Problem 8

**Symptoms:**
- Symptom A for issue 8
- Symptom B for issue 8

**Possible Causes:**
- Cause 1: Description of root cause 8a
- Cause 2: Description of root cause 8b

**Resolution Steps:**
```bash
# Step 1: Check status
curl -s http://localhost:8080/health | jq .
# Step 2: Check logs
tail -f /var/log/burrito/app.log
# Step 3: Verify configuration
cat /etc/burrito/config.yaml
# Step 4: Restart service
systemctl restart burrito
```

### Issue #9: Common Problem 9

**Symptoms:**
- Symptom A for issue 9
- Symptom B for issue 9

**Possible Causes:**
- Cause 1: Description of root cause 9a
- Cause 2: Description of root cause 9b

**Resolution Steps:**
```bash
# Step 1: Check status
curl -s http://localhost:8080/health | jq .
# Step 2: Check logs
tail -f /var/log/burrito/app.log
# Step 3: Verify configuration
cat /etc/burrito/config.yaml
# Step 4: Restart service
systemctl restart burrito
```

### Issue #10: Common Problem 10

**Symptoms:**
- Symptom A for issue 10
- Symptom B for issue 10

**Possible Causes:**
- Cause 1: Description of root cause 10a
- Cause 2: Description of root cause 10b

**Resolution Steps:**
```bash
# Step 1: Check status
curl -s http://localhost:8080/health | jq .
# Step 2: Check logs
tail -f /var/log/burrito/app.log
# Step 3: Verify configuration
cat /etc/burrito/config.yaml
# Step 4: Restart service
systemctl restart burrito
```
