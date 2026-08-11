# Contributing to Burrito

Thank you for your interest in contributing to Burrito! This guide covers everything you need to know to start contributing.

## Code of Conduct

Be respectful, collaborative, and constructive. We follow the [Contributor Covenant](https://www.contributor-covenant.org/).

## Ways to Contribute

- **Bug reports** — Open an issue with reproduction steps
- **Feature requests** — Describe the use case and proposed solution
- **Documentation** — Improve docs, fix typos, add examples
- **Code** — Fix bugs, add features, improve tests
- **Reviews** — Review open pull requests

## Development Setup

### Prerequisites

- Go 1.22+
- Docker & Docker Compose (for integration tests)
- kubectl
- A local Kubernetes cluster (kind, minikube, k3s)
- Node.js 20+ and pnpm (for UI development)
- Python 3.10+ (for build tooling)

### Clone and Build

```bash
git clone https://github.com/padok-team/burrito.git
cd burrito

# Build all binaries
make build

# Run unit tests
make test

# Run linter
golangci-lint run ./...
```

### Local Development

```bash
# Create a local kind cluster
kind create cluster --name burrito-dev

# Install CRDs
make install

# Run the controller locally
make run
```

### UI Development

```bash
cd ui
pnpm install
pnpm dev        # Start dev server
pnpm build      # Production build
pnpm lint       # Run linter
```

## Project Structure

See [ARCHITECTURE.md](ARCHITECTURE.md) for a detailed architecture overview and directory map.

## Commit Convention

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>
```

**Types:** `feat`, `fix`, `chore`, `docs`, `test`, `refactor`

**Examples:**
- `feat(controller): add exponential backoff for layer reconciliation`
- `fix(api): correct pagination offset calculation`
- `docs(contributing): add local development guide`
- `chore(deps): bump controller-runtime to v0.18.0`

CI validates commit messages with commitlint. Self-check with `npx commitlint`.

## Developer Certificate of Origin (DCO)

All commits must include a `Signed-off-by` line:

```
Signed-off-by: Your Name <your.email@example.com>
```

Use `git commit -s` to add it automatically.

## Pull Request Process

1. **Fork** the repository and create a feature branch from `main`
2. **Write code** following existing patterns and style
3. **Add tests** — add cases to existing test suites, don't create new ones
4. **Run tests** — `make test` (requires Docker for envtest)
5. **Run linter** — `golangci-lint run ./...`
6. **Sign your commits** — `git commit -s`
7. **Open a PR** against `main` with a clear description
8. **Wait for CI** — all checks must pass
9. **Address review feedback**

### PR Best Practices

- Keep PRs focused — one concern per PR
- Reference related issues with `Closes #123`
- Include screenshots for UI changes
- Update documentation if changing behavior
- Don't include generated files (`zz_generated.deepcopy.go`, manifests)

## Coding Standards

### Go

- Follow standard Go conventions (gofmt, go vet)
- Always check errors explicitly — never `_ = err`
- No `panic()` in reconcilers — use proper error returns
- Use existing patterns from the codebase
- See `internal/controllers/AGENTS.md` for controller-specific rules

### API Changes

When modifying CRD types (`api/v1alpha1/*_types.go`):

```bash
make manifests && make generate
```

### Testing

- Controllers use BDD-style tests with ginkgo/gomega
- Add test cases to existing suites, don't create new ones from scratch
- See [TESTING.md](TESTING.md) for detailed testing guidance

## Documentation

Documentation lives in `docs/` and is built with [MkDocs](https://www.mkdocs.org/):

```bash
# Install Python deps
pip install -r docs/requirements.txt

# Serve locally
mkdocs serve
```

## Release Process

Releases are automated via GitHub Actions with GoReleaser:

1. Update `VERSION` file
2. Push a tag matching the version: `git tag v0.14.0 && git push origin v0.14.0`
3. CI builds binaries, Docker images, and Helm chart
4. Release is published on GitHub Releases

## Getting Help

- [GitHub Issues](https://github.com/padok-team/burrito/issues)
- [Documentation](https://docs.burrito.tf/)
- [Burrito Community](https://github.com/padok-team/burrito/discussions)
