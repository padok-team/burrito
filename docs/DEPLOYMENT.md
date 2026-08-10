# Deployment Guide — burrito

## Environments

| Environment | Purpose | URL |
|-------------|---------|-----|
| Development | Local development | http://localhost:3000 |
| Staging | Pre-production testing | https://staging.example.com |
| Production | Live environment | https://example.com |

## Prerequisites
- Docker and Docker Compose
- Access to the deployment platform
- Required environment variables configured

## Docker Deployment

### Build
```bash
docker build -t burrito:latest .
```

### Run
```bash
docker-compose up -d
```

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| PORT | No | 3000 | Server port |
| NODE_ENV | No | development | Environment mode |
| DATABASE_URL | Yes | - | Database connection string |
| REDIS_URL | No | - | Redis connection for caching |
| LOG_LEVEL | No | info | Logging verbosity |

## Health Checks
- `/health`: Returns service health status
- `/ready`: Returns readiness status (checks dependencies)

## Rollback
1. Identify the last stable version tag
2. Redeploy that version
3. Verify service health

## Monitoring
- Metrics exported via `/metrics` (Prometheus format)
- Structured JSON logging to stdout/stderr
- Alert on: error rate > 1%, latency p99 > 5s, health check failures

## Security
- Secrets managed via environment variables or a secrets manager
- Never commit secrets to version control
- Regular dependency updates for security patches
- Network policies restrict service-to-service communication