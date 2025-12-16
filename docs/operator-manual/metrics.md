# Burrito Metrics Endpoint

This document describes the metrics exposed by Burrito at the `/metrics` endpoint for Prometheus monitoring.

## Endpoint

The metrics are available from the **controller component** at: `http://<burrito-controllers>:<metrics-port>/metrics`

The metrics endpoint is part of the controller component (not the server), as the controllers have direct access to all Terraform resources through their reconciliation loops.

## Quick Test

To quickly test the metrics endpoint:

```bash
# If running in Kubernetes (default controller metrics port is 8080)
kubectl port-forward -n burrito-system deployment/burrito-controllers 8080:8080
curl http://localhost:8080/metrics

# You can also check standard controller-runtime metrics
curl http://localhost:8080/metrics | grep controller_runtime
```

**Note**: The metrics are exposed by the `burrito-controllers` deployment, not the `burrito-server`. The server handles the Web UI and webhook events, while the controllers manage the lifecycle of Terraform resources and expose their metrics.

## Available Metrics

### Terraform Layer Metrics

#### `burrito_terraform_layer_status`

- **Type**: Gauge
- **Description**: Status of individual Terraform layers
- **Labels**: `namespace`, `layer_name`, `repository_name`, `status`
- **Value**: Always `1` when the layer exists with the given status (status is identified by the label)
- **Status label values**:
  - `disabled`: Layer has no conditions set
  - `success`: Layer is in sync and working properly
  - `warning`: Layer needs plan or apply action
  - `error`: Layer has errors (missing annotations, plan failures, etc.)
  - `running`: Layer is currently executing a run

#### `burrito_terraform_layer_state`

- **Type**: Gauge
- **Description**: Current state of individual Terraform layers (controller state)
- **Labels**: `namespace`, `layer_name`, `repository_name`, `state`
- **Values**: `1` when the layer is in the given state
- **States**: `Idle`, `PlanNeeded`, `ApplyNeeded`, `unknown`

#### `burrito_terraform_layer_runs_total`

- **Type**: Counter
- **Description**: Total number of runs for Terraform layers
- **Labels**: `namespace`, `name`, `repository`, `branch`, `path`, `action`

#### `burrito_terraform_layer_last_run_timestamp`

- **Type**: Gauge
- **Description**: Unix timestamp of the last run for Terraform layers
- **Labels**: `namespace`, `name`, `repository`, `branch`, `path`, `action`

#### `burrito_terraform_layer_run_duration_seconds`

- **Type**: Gauge
- **Description**: Duration of the last run for Terraform layers in seconds
- **Labels**: `namespace`, `name`, `repository`, `branch`, `path`, `action`, `status`

### Terraform Repository Metrics

#### `burrito_terraform_repositories_total`

- **Type**: Gauge
- **Description**: Total number of Terraform repositories

### Terraform Run Metrics

#### Current State Metrics (Gauges)

#### `burrito_runs_by_action`

- **Type**: Gauge
- **Description**: Current number of runs by action type
- **Labels**: `action`
- **Use case**: Track how many plan/apply runs currently exist in the system

#### `burrito_runs_by_status`

- **Type**: Gauge
- **Description**: Current number of runs by status
- **Labels**: `status`
- **Use case**: Monitor distribution of runs across different states (Initial, Running, Succeeded, Failed, etc.)

#### Cumulative Metrics (Counters)

#### `burrito_runs_created_total`

- **Type**: Counter
- **Description**: Total number of runs created since controller startup (cumulative)
- **Labels**: `namespace`, `action`
- **Use case**: Track run creation rate over time with `rate(burrito_runs_created_total[5m])`

#### `burrito_runs_completed_total`

- **Type**: Counter
- **Description**: Total number of runs that completed successfully (cumulative)
- **Labels**: `namespace`, `action`
- **Use case**: Calculate success rate and completion trends

#### `burrito_runs_failed_total`

- **Type**: Counter
- **Description**: Total number of runs that failed (cumulative)
- **Labels**: `namespace`, `action`
- **Use case**: Track failure rate with `rate(burrito_runs_failed_total[5m])`

### Aggregate Summary Metrics

#### `burrito_layers_status_total`

- **Type**: Gauge
- **Description**: Total count of layers by status across all namespaces
- **Labels**: `status`

#### `burrito_layers_namespace_total`

- **Type**: Gauge
- **Description**: Total count of layers per namespace
- **Labels**: `namespace`

#### `burrito_layers_total`

- **Type**: Gauge
- **Description**: Total number of Terraform layers across all namespaces

#### `burrito_repositories_total`

- **Type**: Gauge
- **Description**: Total number of Terraform repositories

#### `burrito_runs_total`

- **Type**: Gauge
- **Description**: Current total number of Terraform run resources

#### `burrito_pullrequests_total`

- **Type**: Gauge
- **Description**: Total number of Terraform pull requests

### Controller Performance Metrics

#### `burrito_controller_reconcile_duration_seconds`

- **Type**: Histogram
- **Description**: Time spent in controller reconciliation (includes buckets for percentile calculations)
- **Labels**: `controller`
- **Use case**: Monitor reconciliation performance with `histogram_quantile(0.95, rate(burrito_controller_reconcile_duration_seconds_bucket[5m]))`

#### `burrito_controller_reconcile_total`

- **Type**: Counter
- **Description**: Total number of reconciliations per controller
- **Labels**: `controller`, `result`
- **Use case**: Track reconciliation frequency with `rate(burrito_controller_reconcile_total[5m])`

#### `burrito_controller_reconcile_errors_total`

- **Type**: Counter
- **Description**: Total reconciliation errors per controller
- **Labels**: `controller`, `error_type`
- **Use case**: Monitor error rates with `rate(burrito_controller_reconcile_errors_total[5m])`

## Usage Examples

### Prometheus Configuration

Add the following to your Prometheus configuration:

```yaml
scrape_configs:
  - job_name: 'burrito-controllers'
    static_configs:
      - targets: ['burrito-controllers:8080']
    metrics_path: '/metrics'
    scrape_interval: 30s
```

**Note**: Make sure to scrape the `burrito-controllers` service/deployment (port 8080 by default, configurable via `config.burrito.controller.metricsBindAddress`), not the `burrito-server`.

### Example Queries

#### Layers by Status

```promql
# Count layers by status
sum by (status) (burrito_terraform_layer_status)

# Get all layers in error state
burrito_terraform_layer_status{status="error"}

# Get layers that need attention (warning or error)
burrito_terraform_layer_status{status=~"warning|error"}

# Get layers by repository
sum by (repository_name) (burrito_terraform_layer_status)

# Count total layers across all statuses
count(burrito_terraform_layer_status)
```

#### Layer States

```promql
# Count layers by state
sum by (state) (burrito_terraform_layer_state)

# Get layers that need planning
burrito_terraform_layer_state{state="PlanNeeded"}

# Get specific layer status
burrito_terraform_layer_status{namespace="production", layer_name="app-infrastructure"}
```

#### Run Metrics

```promql
# Current number of runs by action
sum by (action) (burrito_runs_by_action)

# Current number of runs by status
sum by (status) (burrito_runs_by_status)

# Run creation rate (runs per second over last 5 minutes)
rate(burrito_runs_created_total[5m])

# Run failure rate
rate(burrito_runs_failed_total[5m])

# Success rate (percentage of completed vs failed)
(
  sum(rate(burrito_runs_completed_total[1h])) /
  (sum(rate(burrito_runs_completed_total[1h])) + sum(rate(burrito_runs_failed_total[1h])))
) * 100

# Total runs created in last 24 hours
sum(increase(burrito_runs_created_total[24h]))

# Failed runs in specific namespace
sum by (action) (increase(burrito_runs_failed_total{namespace="production"}[1h]))
```

#### Layer Detailed Analysis

```promql
# Layers currently running
burrito_terraform_layer_status{status="running"}

# Layers with errors
burrito_terraform_layer_status{status="error"}

# Percentage of layers in error state
(sum(burrito_terraform_layer_status{status="error"}) / sum(burrito_terraform_layer_status)) * 100
```

### Alerting Rules

Here's a comprehensive set of alerting rules for monitoring Burrito:

```yaml
groups:
  - name: burrito_layer_health
    interval: 30s
    rules:
      # Critical: Layers in error state (not currently running)
      - alert: BurritoLayerInError
        expr: |
          burrito_terraform_layer_status{status="error"}
          unless
          burrito_terraform_layer_status{status="running"}
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Terraform layer {{ $labels.layer_name }} in {{ $labels.namespace }} is in error state"
          description: |
            Layer "{{ $labels.layer_name }}" in namespace "{{ $labels.namespace }}" has been in error state for 5 minutes.
            Repository: {{ $labels.repository_name }}
            This typically indicates plan or apply failures that require immediate attention.

      # Warning: Layers need planning or applying
      - alert: BurritoLayerNeedsAttention
        expr: |
          burrito_terraform_layer_status{status="warning"}
          unless
          burrito_terraform_layer_status{status="running"}
        for: 15m
        labels:
          severity: warning
        annotations:
          summary: "Terraform layer {{ $labels.layer_name }} needs attention"
          description: |
            Layer "{{ $labels.layer_name }}" in namespace "{{ $labels.namespace }}" has been in warning state for 15 minutes.
            Repository: {{ $labels.repository_name }}
            This may indicate drift detection or pending changes that need to be applied.

      # Warning: Layers stuck in running state
      - alert: BurritoLayerStuckRunning
        expr: |
          burrito_terraform_layer_status{status="running"}
        for: 30m
        labels:
          severity: warning
        annotations:
          summary: "Terraform layer {{ $labels.layer_name }} stuck in running state"
          description: |
            Layer "{{ $labels.layer_name }}" has been running for over 30 minutes.
            Repository: {{ $labels.repository_name }}
            Check if the run is actually progressing or if it's stuck.

      # Critical: High percentage of layers in error
      - alert: BurritoHighErrorRate
        expr: |
          (sum(burrito_terraform_layer_status{status="error"}) / sum(burrito_terraform_layer_status)) * 100 > 50
        for: 10m
        labels:
          severity: critical
        annotations:
          summary: "More than 50% of Terraform layers are in error state"
          description: |
            {{ printf "%.1f" $value }}% of layers are currently in error state.
            This may indicate a systemic issue affecting multiple layers.

      # Warning: No layers found (monitoring health check)
      - alert: BurritoNoLayersFound
        expr: |
          sum(burrito_terraform_layer_status) == 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "No Terraform layers found in Burrito"
          description: |
            No layers are being reported by Burrito metrics.
            This may indicate an issue with the controller or metrics collection.

  - name: burrito_run_failures
    interval: 30s
    rules:
      # Critical: Run failures increasing
      - alert: BurritoRunFailureRate
        expr: |
          rate(burrito_runs_failed_total[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High rate of Terraform run failures in {{ $labels.namespace }}"
          description: |
            Namespace "{{ $labels.namespace }}" is experiencing {{ printf "%.2f" $value }} run failures per second for action "{{ $labels.action }}".
            Current failure rate indicates {{ printf "%.0f" (mul $value 300) }} failures in the last 5 minutes.

      # Warning: Runs created but not completing
      - alert: BurritoRunsStalling
        expr: |
          (
            rate(burrito_runs_created_total[10m]) - 
            rate(burrito_runs_completed_total[10m]) - 
            rate(burrito_runs_failed_total[10m])
          ) > 0.5
        for: 15m
        labels:
          severity: warning
        annotations:
          summary: "Terraform runs may be stalling in {{ $labels.namespace }}"
          description: |
            More runs are being created than completed/failed in namespace "{{ $labels.namespace }}" for action "{{ $labels.action }}".
            This could indicate runs getting stuck or taking too long.

      # Info: High run volume
      - alert: BurritoHighRunVolume
        expr: |
          rate(burrito_runs_created_total[5m]) > 1
        for: 10m
        labels:
          severity: info
        annotations:
          summary: "High volume of Terraform runs in {{ $labels.namespace }}"
          description: |
            Namespace "{{ $labels.namespace }}" is creating {{ printf "%.2f" $value }} runs per second for action "{{ $labels.action }}".
            This may be expected behavior or could indicate excessive reconciliation.

  - name: burrito_repository_health
    interval: 30s
    rules:
      # Warning: Repository in error state
      - alert: BurritoRepositoryError
        expr: |
          burrito_terraform_repository_status{status="error"}
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Terraform repository {{ $labels.repository_name }} is in error state"
          description: |
            Repository "{{ $labels.repository_name }}" in namespace "{{ $labels.namespace }}" is in error state.
            URL: {{ $labels.url }}
            This may prevent layers from being updated. Check repository connectivity and credentials.

      # Critical: No repositories found
      - alert: BurritoNoRepositoriesFound
        expr: |
          burrito_repositories_total == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "No Terraform repositories found in Burrito"
          description: |
            No repositories are being reported by Burrito metrics.
            This will prevent any layers from functioning.

  - name: burrito_controller_health
    interval: 30s
    rules:
      # Warning: High reconciliation error rate
      - alert: BurritoControllerErrors
        expr: |
          rate(burrito_controller_reconcile_errors_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate in {{ $labels.controller }} controller"
          description: |
            The {{ $labels.controller }} controller is experiencing {{ printf "%.2f" $value }} errors per second.
            Error type: {{ $labels.error_type }}
            Check controller logs for details.

      # Warning: Slow reconciliation
      - alert: BurritoSlowReconciliation
        expr: |
          histogram_quantile(0.95, rate(burrito_controller_reconcile_duration_seconds_bucket[5m])) > 30
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "{{ $labels.controller }} controller reconciliation is slow"
          description: |
            95th percentile reconciliation time for {{ $labels.controller }} is {{ printf "%.2f" $value }} seconds.
            This may indicate performance issues or resource constraints.

  - name: burrito_aggregate_health
    interval: 1m
    rules:
      # Info: Overall system health score
      - record: burrito:layer_health_score
        expr: |
          (
            sum(burrito_terraform_layer_status{status="success"}) /
            sum(burrito_terraform_layer_status)
          ) * 100

      # Info: Success rate of runs
      - record: burrito:run_success_rate
        expr: |
          (
            sum(rate(burrito_runs_completed_total[1h])) /
            (sum(rate(burrito_runs_completed_total[1h])) + sum(rate(burrito_runs_failed_total[1h])))
          ) * 100
```

#### Alert Severity Guidelines

- **Critical**: Immediate action required - layers in error, high failure rates, missing resources
- **Warning**: Attention needed soon - layers needing updates, slow performance, stalling runs
- **Info**: Informational - high volume, health scores, trends

#### Customization Tips

1. **Adjust Thresholds**: Modify `for:` durations and numeric thresholds based on your environment
2. **Add Routing**: Use `severity` labels to route alerts to appropriate channels (PagerDuty, Slack, email)
3. **Namespace Filtering**: Add namespace filters to alerts if you have different SLAs per environment
4. **Team Assignment**: Add custom labels like `team` or `service` to route alerts to the right owners

## Security Considerations

The metrics endpoint is exposed without authentication by default to allow Prometheus scraping. Consider:

1. Restricting access to the `/metrics` endpoint via network policies
2. Using service mesh/ingress controller authentication if needed
3. Being aware that metrics may expose layer names, namespaces, and repository information
