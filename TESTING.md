# Burrito Testing Guide

This document describes the testing strategy, tools, and patterns used in Burrito.

## Testing Philosophy

- **Test behavior, not implementation.** Tests should validate expected outcomes, not internal details.
- **Add to existing suites.** Don't create new test suites from scratch — extend the existing ones.
- **Cover edge cases.** Think about error paths, empty states, and boundary conditions.
- **Keep tests fast.** Unit tests should run in milliseconds; integration tests should be targeted.

## Test Categories

### Unit Tests

Run with the standard Go test runner:

```bash
go test ./...
# or
make test
```

Unit tests cover:
- Individual function logic
- Data transformations
- Utility functions
- Controller logic (with mocked dependencies)

### Controller Tests (BDD)

Controller tests use [ginkgo](https://onsi.github.io/ginkgo/) v2 and [gomega](https://onsi.github.io/gomega/) with `envtest`:

```bash
# Run all controller tests
go test ./internal/controllers/...

# Run a specific suite
go test ./internal/controllers/ -run "TerraformLayer" -v
```

**Test structure:**
```go
var _ = Describe("TerraformLayer Controller", func() {
    When("a layer is created with a valid repository reference", func() {
        It("should transition to PlanInProgress status", func() {
            // Setup
            // Exercise
            // Assert
        })
    })
})
```

**Key patterns:**
- Each `Describe` block corresponds to a controller
- Each `When` block describes a scenario
- Each `It` block asserts a specific outcome
- Use `BeforeEach` for setup, `AfterEach` for cleanup
- Use gomega matchers: `Expect(err).NotTo(HaveOccurred())`, `Expect(status).To(Equal("Ready"))`

### Integration Tests

Integration tests require a running Kubernetes environment (kind/minikube) and Docker:

```bash
# Requires Docker daemon running
make test
```

These tests spin up `envtest` (a real API server + etcd) and validate full reconciliation loops.

### End-to-End Tests

Located in `testdata/` and CI workflows. E2E tests:

- Deploy Burrito to a real cluster
- Create TerraformRepository and TerraformLayer resources
- Validate plan/apply workflows
- Test webhook integrations

## Running Tests

### Quick Start

```bash
# Run all tests
make test

# Run specific package
go test ./internal/controllers/ -v

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Linting

```bash
# Go lint
golangci-lint run ./...

# UI lint
cd ui && pnpm lint

# UI type checking
cd ui && pnpm typecheck
```

## Writing Tests

### Adding a Controller Test

1. Find the existing test file: `internal/controllers/<resource>_controller_test.go`
2. Identify the right `Describe`/`When` block to add your case
3. Follow the existing pattern for setup and assertions
4. Use the helper functions available in the test file

### Example: Layer Controller Test

```go
var _ = Describe("TerraformLayer Controller", func() {
    When("a layer references a non-existent repository", func() {
        It("should set status to Error with a descriptive message", func() {
            ctx := context.Background()

            layer := &configv1alpha1.TerraformLayer{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "test-layer",
                    Namespace: "default",
                },
                Spec: configv1alpha1.TerraformLayerSpec{
                    Repository: configv1alpha1.TerraformLayerRepository{
                        Name: "non-existent-repo",
                    },
                    Path: "environments/dev",
                },
            }

            Expect(k8sClient.Create(ctx, layer)).To(Succeed())

            Eventually(func() string {
                _ = k8sClient.Get(ctx, types.NamespacedName{
                    Name: "test-layer", Namespace: "default",
                }, layer)
                return layer.Status.State
            }, timeout, interval).Should(Equal("Error"))
        })
    })
})
```

### Mocking

For unit tests outside controllers, use standard Go interfaces and mocks:

```go
type mockGitProvider struct {
    files map[string]string
}

func (m *mockGitProvider) GetFile(path string) (string, error) {
    content, ok := m.files[path]
    if !ok {
        return "", fmt.Errorf("file not found: %s", path)
    }
    return content, nil
}
```

### Test Fixtures

Test fixtures live in `testdata/`:
- Sample Terraform configurations
- Repository manifests
- Layer manifests

## CI Pipeline

Tests run automatically on every PR via GitHub Actions (`.github/workflows/ci.yaml`):

1. **Lint** — `golangci-lint run ./...`
2. **Unit tests** — `go test ./...`
3. **Integration tests** — `make test` with envtest
4. **UI checks** — `pnpm lint`, `pnpm typecheck`
5. **Build** — `make build`

Code coverage is tracked via [Codecov](https://codecov.io/) with the configuration in `codecov.yml`.

## Test Dependencies

| Dependency           | Purpose                          | Required For        |
|----------------------|----------------------------------|---------------------|
| Docker               | envtest binary + containers      | Integration tests   |
| kubebuilder assets   | envtest API server               | Controller tests    |
| kind / minikube      | Local cluster                    | E2E tests           |
| golangci-lint        | Static analysis                  | Lint checks         |

## Debugging Tests

### Common Issues

**"exec: "gcc": executable file not found"**
→ Install gcc (required for CGO): `apt install build-essential` or equivalent

**envtest not found**
→ Run `make envtest` to download kubebuilder test binaries

**Docker not running**
→ Start Docker daemon before running integration tests

**Timeout errors**
→ Increase timeout in test setup or check if required services are running

### Verbose Output

```bash
go test -v ./internal/controllers/...
```

### Running a Single Test

```bash
go test ./internal/controllers/ -run "should transition to PlanInProgress" -v
```
