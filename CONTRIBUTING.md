# Contributing to TaaS Platform

Thank you for your interest in contributing to the TaaS Platform! This document provides guidelines for contributing to the project.

## Table of Contents

- [Getting Started](#getting-started)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Code Style](#code-style)
- [Testing](#testing)
- [Commit Message Format](#commit-message-format)
- [Pull Request Process](#pull-request-process)
- [CI/CD](#cicd)
- [Release Process](#release-process)
- [Reporting Issues](#reporting-issues)
- [License](#license)

## Getting Started

### Prerequisites

- **Go 1.26+**: [Download](https://go.dev/dl/)
- **Rust 1.94+**: `curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh`
- **Node.js 22+**: [Download](https://nodejs.org/)
- **pnpm**: `npm install -g pnpm`
- **Docker**: [Install](https://docs.docker.com/get-docker/)
- **PostgreSQL 15+**: [Install](https://www.postgresql.org/download/)
- **Redis**: [Install](https://redis.io/download)
- **Kubernetes**: [Install](https://kubernetes.io/docs/tasks/tools/) (for production deployment)
- **Helm 3+**: [Install](https://helm.sh/docs/intro/install/) (for production deployment)
- **MongoDB 7.0+**: [Install](https://www.mongodb.com/docs/manual/installation/) (required for free5GC only)

### Setup

1. **Fork the repository**
2. **Clone your fork:**
   ```bash
   git clone https://github.com/nutcas3/telecom-platform.git
   cd telecom-platform
   ```
3. **Install dependencies:**
   ```bash
   make install-deps
   ```
4. **Set up databases:**
   ```bash
   make db-setup
   make db-migrate
   ```
5. **Build all components:**
   ```bash
   make all
   ```

## Project Structure

```
telecom-platform/
├── apps/
│   ├── api-server/          # Go BSS API server
│   ├── carrier-connector/   # Go ES2+ carrier integration
│   ├── charging-engine/     # Rust real-time charging
│   ├── packet-gateway/      # Rust eBPF packet processing
│   ├── cli/                 # Go CLI tool
│   └── web-dashboard/       # Next.js management UI
├── sdk/
│   └── typescript/          # TypeScript SDK
├── deployments/
│   ├── docker/              # Dockerfiles
│   ├── kubernetes/          # K8s manifests
│   ├── helm/                # Helm charts
│   └── free5gc/             # 5G core network config
├── monitoring/             # Prometheus/Grafana configs
└── scripts/                # Setup and utility scripts
```

## Development Workflow

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes:**
   - Write clear, documented code
   - Follow language-specific style guides
   - Add tests for new functionality
   - Update documentation as needed

3. **Run linters:**
   ```bash
   make lint
   ```

4. **Run tests:**
   ```bash
   make test
   ```

5. **Format code:**
   ```bash
   make fmt
   ```

6. **Commit your changes:**
   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```

7. **Push to your fork:**
   ```bash
   git push origin feature/your-feature-name
   ```

8. **Create a Pull Request**

## Code Style

### Go

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` for formatting: `make fmt-go`
- Run `golangci-lint` before committing: `make lint-go`
- Use descriptive variable and function names
- Add godoc comments for exported functions
- Handle errors explicitly, don't ignore them

### Rust

- Follow [Rust API Guidelines](https://rust-lang.github.io/api-guidelines/)
- Use `cargo fmt` for formatting: `cargo fmt`
- Run `cargo clippy` before committing: `make lint-rust`
- Use `Result<T, E>` for error handling
- Prefer `unwrap()` only in tests
- Document unsafe blocks with safety comments

### TypeScript

- Follow [Airbnb Style Guide](https://github.com/airbnb/javascript)
- Use Prettier for formatting: `make lint-ui`
- Use ESLint for linting
- Use TypeScript strict mode
- Prefer functional programming patterns
- Add JSDoc comments for public APIs

## Testing

### Testing Strategy

- **Unit Tests**: Test individual functions and methods in isolation
- **Integration Tests**: Test service interactions with real dependencies
- **End-to-End Tests**: Test complete workflows across services
- **Race Detection**: Use `-race` flag for Go tests to detect data races

### Running Tests

```bash
# Run all tests
make test

# Run Go tests with race detection
make test-go

# Run Rust tests in release mode
make test-rust-release

# Run TypeScript tests
pnpm -r test
```

### Test Coverage

- Aim for >80% code coverage on new code
- Run coverage reports before submitting PRs
- Add tests for bug fixes to prevent regressions

## Commit Message Format

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks
- `perf`: Performance improvements
- `ci`: CI/CD changes

**Scopes:**
- `api`: API server changes
- `charging`: Charging engine changes
- `gateway`: Packet gateway changes
- `connector`: Carrier connector changes
- `cli`: CLI tool changes
- `dashboard`: Web dashboard changes
- `sdk`: TypeScript SDK changes
- `infra`: Infrastructure changes
- `docs`: Documentation changes

**Examples:**
```
feat(api): add eSIM deletion endpoint
fix(charging): correct credit deduction logic
docs: update deployment guide
perf(gateway): optimize eBPF packet processing
ci: add race detection to tests
```

## Pull Request Process

1. **Update documentation** if needed
2. **Add tests** for your changes
3. **Ensure all tests pass**: `make test`
4. **Run linters**: `make lint`
5. **Format code**: `make fmt`
6. **Update CHANGELOG.md** with your changes
7. **Request review** from maintainers
8. **Address review feedback** promptly
9. **Wait for approval** and merge

### PR Checklist

- [ ] Code follows project style guidelines
- [ ] Tests added/updated and passing
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] No merge conflicts
- [ ] Commit messages follow conventional commits

## CI/CD

### GitHub Actions

The project uses GitHub Actions for:
- **Linting**: golangci-lint, clippy, ESLint
- **Testing**: Go tests with race detection, Rust tests, TypeScript tests
- **Building**: Docker images for all services
- **Security**: Trivy vulnerability scanning

### Branch Protection

- `main` and `develop` branches are protected
- PRs require at least one approval
- CI must pass before merging
- No direct pushes to protected branches

## Release Process

1. Update version numbers in relevant files
2. Update CHANGELOG.md with release notes
3. Create release branch: `git checkout -b release/v1.0.0`
4. Run full test suite: `make test`
5. Build Docker images: `make docker-build`
6. Tag release: `git tag -a v1.0.0 -m "Release v1.0.0"`
7. Push tag: `git push origin v1.0.0`
8. Create GitHub release with notes
9. Deploy to production: `make k8s-deploy-helm`

## Reporting Issues

When reporting bugs or requesting features:

1. **Use GitHub Issues** with appropriate templates
2. **Provide clear description** of the problem
3. **Include reproduction steps**
4. **Add relevant logs** or screenshots
5. **Specify environment details**:
   - OS and version
   - Go/Rust/Node.js versions
   - Database versions
   - Kubernetes version (if applicable)

## Code Review

We review all contributions. Expect:
- **Feedback on code quality** and best practices
- **Suggestions for improvements** and optimizations
- **Questions about implementation** decisions
- **Requests for tests** or documentation
- **Security considerations** for sensitive code

## Security

- Report security vulnerabilities privately to maintainers
- Never commit secrets, API keys, or passwords
- Use Vault for secret management in production
- Follow OWASP guidelines for security best practices

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
