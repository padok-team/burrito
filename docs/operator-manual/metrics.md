# Burrito Metrics Endpoint

This document describes the metrics exposed by Burrito at the `/metrics` endpoint for Prometheus monitoring.

## Endpoint

The metrics are available at: `http://<burrito-server>/metrics`

## Quick Test

To quickly test the metrics endpoint:

```bash
# If running locally (default port 8080)
curl http://localhost:8080/metrics

# Or using kubectl port-forward if running in Kubernetes
kubectl port-forward svc/burrito-server 8080:8080
curl http://localhost:8080/metrics
```

## Available Metrics

### Terraform Layer Metrics

#### `burrito_terraform_layer_status`
- **Type**: Gauge
- **Description**: Status of Terraform layers based on UI representation
- **Labels**: `namespace`, `name`, `repository`, `branch`, `path`, `status`
- **Values**:
  - `0`: disabled - Layer has no conditions set
  - `1`: success - Layer is in sync and working properly
  - `2`: warning - Layer needs plan or apply action
  - `3`: error - Layer has errors (missing annotations, plan failures, etc.)
  - `4`: running - Layer is currently executing a run

#### `burrito_terraform_layer_state`
- **Type**: Gauge
- **Description**: Current state of Terraform layers (controller state)
- **Labels**: `namespace`, `name`, `repository`, `branch`, `path`, `state`
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

#### `burrito_terraform_layer_condition`
- **Type**: Gauge
- **Description**: Condition status for Terraform layers
- **Labels**: `namespace`, `name`, `repository`, `branch`, `path`, `condition`, `status`, `reason`
- **Values**:
  - `1`: True
  - `0`: False
  - `-1`: Unknown

### Terraform Repository Metrics

#### `burrito_terraform_repositories_total`
- **Type**: Gauge
- **Description**: Total number of Terraform repositories

#### `burrito_terraform_repository_status`
- **Type**: Gauge
- **Description**: Status of Terraform repositories
- **Labels**: `namespace`, `name`, `url`, `branch`, `status`
- **Values**: Same as layer status values

### Terraform Run Metrics

#### `burrito_terraform_run_status`
- **Type**: Gauge
- **Description**: Status of Terraform runs
- **Labels**: `namespace`, `name`, `layer_name`, `action`, `state`
- **Values**: `1` when the run is in the given state

#### `burrito_terraform_run_duration_seconds`
- **Type**: Gauge
- **Description**: Duration of Terraform runs in seconds
- **Labels**: `namespace`, `name`, `layer_name`, `action`, `state`

#### `burrito_terraform_run_retries`
- **Type**: Gauge
- **Description**: Number of retries for Terraform runs
- **Labels**: `namespace`, `name`, `layer_name`, `action`, `state`

### Terraform Pull Request Metrics

#### `burrito_terraform_pullrequests_total`
- **Type**: Gauge
- **Description**: Total number of Terraform pull requests

#### `burrito_terraform_pullrequest_status`
- **Type**: Gauge
- **Description**: Status of Terraform pull requests
- **Labels**: `namespace`, `name`, `repository`, `pr_id`, `state`
- **Values**: `1` when the pull request is in the given state

## Usage Examples

### Prometheus Configuration

Add the following to your Prometheus configuration:

```yaml
scrape_configs:
  - job_name: 'burrito'
    static_configs:
      - targets: ['burrito-server:8080']
    metrics_path: '/metrics'
    scrape_interval: 30s
```

### Example Queries

#### Layers by Status
```promql
# Count layers by status
sum by (status) (burrito_terraform_layer_status)

# Get all layers in error state
burrito_terraform_layer_status{status="error"}

# Get layers that need attention (warning or error)
burrito_terraform_layer_status{status=~"warning|error"}
```

#### Layer States
```promql
# Count layers by state
sum by (state) (burrito_terraform_layer_state)

# Get layers that need planning
burrito_terraform_layer_state{state="PlanNeeded"}
```

#### Run Metrics
```promql
# Average run duration by action
avg by (action) (burrito_terraform_run_duration_seconds)

# Runs with retries
burrito_terraform_run_retries > 0

# Failed runs
burrito_terraform_run_status{state="Failed"}
```

#### Layer Conditions
```promql
# Layers that are running
burrito_terraform_layer_condition{condition="IsRunning", status="True"}

# Layers with failed conditions
burrito_terraform_layer_condition{status="False"}
```

### Alerting Rules

The most important alert for monitoring Terraform layers is when they are not OK but not currently running:

```yaml
# Alert when a layer is not OK (error or warning state) but not currently running
- alert: BurritoLayerNotOK
  expr: |
    burrito_terraform_layer_status{status=~"error|warning"} == 1
    and
    burrito_terraform_layer_status{status="running"} == 0
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Terraform layer {{ $labels.name }} in {{ $labels.namespace }} is not OK"
    description: |
      Layer "{{ $labels.name }}" has status "{{ $labels.status }}" and needs attention.
      Repository: {{ $labels.repository }}, Path: {{ $labels.path }}
```

For a complete set of alerting rules, see: [examples/prometheus-alerts.yaml](../../examples/prometheus-alerts.yaml)

#### Key Alert Categories

1. **Layer Health Alerts**
   - `BurritoLayerNotOK`: Layers in error/warning state (not running)
   - `BurritoLayerInErrorStateTooLong`: Persistent error states
   - `BurritoLayerStuckRunning`: Runs taking too long

2. **Run Alerts**  
   - `BurritoRunFailed`: Failed terraform runs
   - `BurritoRunTooManyRetries`: Runs with excessive retries

3. **Overview Alerts**
   - `BurritoTooManyLayersInError`: High error rate across layers
   - `BurritoNoLayersFound`: Monitoring health check

## Grafana Dashboard

A ready-to-use Grafana dashboard is available in the `/examples/` directory of this repository.

### Quick Import

1. Download `examples/grafana-dashboard.json`
2. In Grafana, go to "Dashboards" → "Import"
3. Upload the JSON file
4. Select your Prometheus datasource
5. Click "Import"

The dashboard includes:

1. **Layer Status Overview**: Bar chart showing distribution of layer statuses
2. **Layer States Distribution**: Pie chart showing controller states
3. **Layers Requiring Attention**: Table of layers needing intervention
4. **Active Runs**: Real-time count of running operations
5. **Run Performance**: Average duration by action type
6. **Failed Runs**: Table of failed operations
7. **Status Timeline**: Historical view of layer status changes
8. **Summary Statistics**: Total counts of layers, repositories, and PRs

### Dashboard Features

- **Template Variables**: Filter by namespace and repository
- **Auto-refresh**: 30-second refresh rate for real-time monitoring
- **Color Coding**: Consistent colors (green=success, yellow=warning, red=error, blue=running)
- **Responsive Design**: Works on desktop and mobile devices

For detailed setup instructions and customization options, see `examples/README-grafana.md`.

## Security Considerations

The metrics endpoint is exposed without authentication by default to allow Prometheus scraping. Consider:

1. Restricting access to the `/metrics` endpoint via network policies
2. Using service mesh/ingress controller authentication if needed
3. Being aware that metrics may expose layer names, namespaces, and repository information

## Performance Notes

- Metrics are collected on each scrape request by querying the Kubernetes API
- For large clusters, consider adjusting the scrape interval to balance freshness vs. load
- The metrics collection time depends on the number of layers, runs, and repositories in your cluster