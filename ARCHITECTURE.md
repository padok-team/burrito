# Burrito Architecture

## Overview

Burrito is a Kubernetes-native TACoS (**T**erraform **A**utomation **Co**llaboration **S**oftware) operator built on the [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) framework. It extends the Kubernetes API with custom resources to manage Terraform/OpenTofu/Terragrunt workflows.

## High-Level Components

```
┌─────────────────────────────────────────────────────────────┐
│                     Kubernetes Cluster                       │
│  ┌─────────────┐  ┌──────────────┐  ┌───────────────────┐  │
│  │  Terraform   │  │  Terraform   │  │     Burrito       │  │
│  │  Repository  │  │    Layer     │  │     Server        │  │
│  │  Controller  │  │  Controller  │  │    (API + UI)     │  │
│  └──────┬───────┘  └──────┬───────┘  └────────┬──────────┘  │
│         │                 │                    │            │
│  ┌──────┴─────────────────┴────────────────────┴──────────┐ │
│  │                    Burrito Datastore                    │ │
│  │              (Plan / State Artifacts)                   │ │
│  └────────────────────────────────────────────────────────┘ │
│         │                 │                                 │
│  ┌──────┴───────┐  ┌──────┴────────┐                       │
│  │   Runner     │  │   Webhook     │                       │
│  │  (Terraform  │  │   Receiver    │                       │
│  │   binary)    │  │  (GitHub/GL)  │                       │
│  └──────────────┘  └───────────────┘                       │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### Custom Resource Definitions (CRDs)

Burrito defines two primary CRDs under `api/v1alpha1/`:

- **TerraformRepository** — references a Git repository containing Terraform code. It defines:
  - Repository URL and authentication
  - Default branch and path filters
  - Terraform/Terragrunt/OpenTofu version constraints
  - Pull request workflow configuration
  - Remediation strategies

- **TerraformLayer** — represents a single Terraform workspace within a repository. It defines:
  - Path within the repository
  - Terraform backend configuration
  - Override settings for the runner
  - Sync windows
  - Additional trigger paths

### Controllers (`internal/controllers/`)

Controllers implement the reconciliation loop pattern. Each CRD has a dedicated controller:

- **TerraformRepositoryController** — watches `TerraformRepository` resources.
  - Discovers layers from the repository
  - Manages runner scheduling via Hermitcrab
  - Handles commit status updates
  - Manages Git webhook event processing

- **TerraformLayerController** — watches `TerraformLayer` resources.
  - Executes `terraform plan` on changes
  - Executes `terraform apply` when remediation is triggered
  - Manages layer lifecycle (plan → apply → drift detection)
  - Stores plan artifacts in the datastore

Both controllers use exponential backoff for retry logic and respect sync windows.

### Runner (`internal/runner/`)

The runner is a standalone Go binary (`cmd/runner/`) that executes Terraform operations:

- Downloads repository at a specific commit
- Installs required Terraform/OpenTofu/Terragrunt version via `tenv`
- Supports provider caching
- Executes plan/apply in isolated workspaces
- Streams logs and uploads artifacts to the datastore

### Datastore (`internal/datastore/`)

A storage service for Terraform plan artifacts and logs:

- Pluggable storage backends (Azure Blob, GCS, S3-compatible)
- Serves plan results to the API server
- TTL-based artifact expiration
- Hostname-aware storage paths for multi-node deployments

### Server (`internal/server/`)

The API and UI backend server:

- REST API for the Burrito dashboard
- Serves the React/Vite UI (`ui/`)
- Provides layer/repository status queries
- Plan diff viewing

### Webhook (`internal/webhook/`)

Receives Git webhook events (GitHub, GitLab):

- Validates webhook signatures
- Triggers repository reconciliation on push/PR events
- Supports multi-namespace event routing

### UI (`ui/`)

A React + TypeScript + Vite dashboard providing:

- Repository and layer status overview
- Plan result inspection and diff viewing
- Layer pagination and filtering
- Manual sync trigger (sync button)

## Data Flow

### Plan/Apply Flow

```
VCS Push/PR → Webhook → Repository Controller
    → Runner Pod Created → Clone Repo → terraform init
    → terraform plan → Upload Artifact → Datastore
    → Commit Status Posted → UI Shows Plan
    → (On merge/auto-remediate) → Runner Pod → terraform apply
```

### Reconciliation Loop

```
Kubernetes Watch → Controller Queue → Exponential Backoff
    → Reconcile() → Compare Spec vs State
    → Create/Update Runner Job → Update Status Subresource
```

## Technology Stack

| Component     | Technology                         |
|---------------|------------------------------------|
| Operator      | Go, Kubebuilder, controller-runtime |
| API           | Go, net/http                      |
| UI            | React, TypeScript, Vite           |
| Storage       | Azure Blob / GCS / S3 (pluggable) |
| IaC Tools     | Terraform, OpenTofu, Terragrunt   |
| Version Mgmt  | tenv                              |
| CI/CD         | GitHub Actions, GoReleaser        |
| Deployment    | Helm chart                        |

## Directory Map

```
burrito/
├── api/v1alpha1/        # CRD Go types (codegen-sensitive)
├── cmd/                 # Binary entrypoints
│   ├── controller/      # Main controller binary
│   ├── runner/          # Terraform runner binary
│   └── datastore/       # Datastore server binary
├── internal/
│   ├── controllers/     # Reconciliation logic
│   ├── repository/      # Git provider abstraction
│   ├── runner/          # Runner implementation
│   ├── datastore/       # Storage backend
│   ├── server/          # API server
│   └── webhook/         # VCS webhook receiver
├── ui/                  # React dashboard
├── deploy/charts/       # Helm chart
├── docs/                # Documentation (mkdocs)
├── hack/                # Development scripts
└── manifests/           # Generated manifests
```

## Key Design Decisions

1. **Kubernetes-Native**: All state is stored in Kubernetes CRDs and status subresources. No external database required.
2. **GitOps Principles**: Terraform state is reconciled against the desired state defined in Git repositories.
3. **Pluggable Storage**: The datastore supports multiple cloud storage backends for portability.
4. **Exponential Backoff**: All reconciliation uses exponential backoff to handle transient failures gracefully.
5. **DCO Required**: All commits must be signed off (`Signed-off-by`).
