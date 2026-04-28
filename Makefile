.PHONY: all build-ui build-rust build-go build-cli build-packet-gateway clean test dev help install-deps db-setup db-migrate

# Colors for terminal output
CYAN := \033[0;36m
GREEN := \033[0;32m
YELLOW := \033[1;33m
RESET := \033[0m

help: ## Show this help message
	@echo "$(CYAN)Telecom Platform - Available Commands:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'

all: build-ui build-rust build-go build-cli build-packet-gateway ## Build all components

install-deps: ## Install all dependencies
	@echo "$(CYAN)Installing Go dependencies...$(RESET)"
	@cd apps/api-server && go mod download
	@cd apps/carrier-connector && go mod download
	@cd apps/cli && go mod download
	@echo "$(CYAN)Installing Rust dependencies...$(RESET)"
	@cargo fetch
	@echo "$(CYAN)Installing Node.js dependencies...$(RESET)"
	@pnpm install
	@echo "$(GREEN)All dependencies installed!$(RESET)"

# TypeScript - Frontend & SDK
build-ui: ## Build Next.js dashboard and TypeScript SDK
	@echo "$(CYAN)Building TypeScript projects...$(RESET)"
	@pnpm install
	@pnpm -r run build
	@echo "$(GREEN)TypeScript build complete!$(RESET)"

dev-ui: ## Start Next.js development server
	@pnpm --filter web-dashboard dev

lint-ui: ## Lint TypeScript code
	@pnpm -r run lint

# Rust - Data Plane & Charging Engine
build-rust: ## Build Rust components (release mode)
	@echo "$(CYAN)Building Rust projects...$(RESET)"
	@cargo build --release
	@echo "$(GREEN)Rust build complete!$(RESET)"

build-packet-gateway: ## Build eBPF packet gateway
	@echo "$(CYAN)Building packet gateway...$(RESET)"
	@cd apps/packet-gateway && cargo build --release
	@echo "$(GREEN)Packet gateway build complete!$(RESET)"

build-rust-dev: ## Build Rust components (debug mode)
	@cargo build

test-rust: ## Run Rust tests
	@cargo test --workspace

test-rust-release: ## Run Rust tests in release mode
	@cargo test --workspace --release

lint-rust: ## Run Rust linter (clippy)
	@cargo clippy --workspace -- -D warnings

# Go - BSS API & Carrier Connector
build-go: ## Build Go applications
	@echo "$(GREEN)Building Go applications...$(RESET)"
	@cd apps/api-server && go build -o ../../dist/api-server ./cmd/
	@cd apps/carrier-connector && go build -o ../../dist/carrier-connector .
	@echo "$(GREEN)Go build complete!$(RESET)"

build-cli: ## Build CLI tool
	@echo "$(CYAN)Building CLI...$(RESET)"
	@mkdir -p dist
	@cd apps/cli && go build -o ../../dist/taas-cli .
	@echo "$(GREEN)CLI build complete!$(RESET)"

test-go: ## Run Go tests
	@go test -v -race -coverprofile=coverage.out ./apps/api-server/...
	@go test -v ./apps/carrier-connector/...
	@go test -v ./apps/cli/...

lint-go: ## Run Go linter
	@golangci-lint run ./apps/...

fmt-go: ## Format Go code
	@go fmt ./apps/...

# Development
dev: ## Start development environment (info only)
	@echo "$(CYAN)To start the development environment:$(RESET)"
	@echo ""
	@echo "$(YELLOW)Terminal 1 - API Server:$(RESET)"
	@echo "  ./dist/api-server"
	@echo ""
	@echo "$(YELLOW)Terminal 2 - Carrier Connector:$(RESET)"
	@echo "  ./dist/carrier-connector"
	@echo ""
	@echo "$(YELLOW)Terminal 3 - Charging Engine:$(RESET)"
	@echo "  ./target/release/charging-engine"
	@echo ""
	@echo "$(YELLOW)Terminal 4 - Web Dashboard:$(RESET)"
	@echo "  make dev-ui"
	@echo ""

# Testing
test: test-go test-rust ## Run all tests
	@echo "$(GREEN)All tests passed!$(RESET)"

# Linting
lint: lint-go lint-rust lint-ui ## Run all linters

# Formatting
fmt: fmt-go ## Format all code

# Cleaning
clean: ## Remove all build artifacts
	@echo "$(CYAN)Cleaning build artifacts...$(RESET)"
	@rm -rf dist/
	@rm -rf target/
	@pnpm -r exec rm -rf .next dist node_modules
	@find . -name "*.o" -type f -delete
	@find . -name "*.a" -type f -delete
	@echo "$(GREEN)Clean complete!$(RESET)"

clean-all: clean ## Remove all artifacts including dependencies
	@rm -rf node_modules
	@rm -rf apps/*/node_modules
	@rm -rf libs/*/node_modules
	@cd apps/api-server && rm -rf vendor/
	@cd apps/carrier-connector && rm -rf vendor/

# Docker
docker-build: ## Build Docker images for all services
	@echo "$(CYAN)Building Docker images...$(RESET)"
	@docker build -f deployments/docker/api-server.Dockerfile -t taas-api-server .
	@docker build -f deployments/docker/carrier-connector.Dockerfile -t taas-carrier-connector .
	@docker build -f deployments/docker/charging-engine.Dockerfile -t taas-charging-engine .
	@docker build -f deployments/docker/web-dashboard.Dockerfile -t taas-web-dashboard .
	@docker build -f deployments/docker/packet-gateway.Dockerfile -t taas-packet-gateway .
	@echo "$(GREEN)Docker images built!$(RESET)"

docker-push: ## Push Docker images to registry
	@echo "$(CYAN)Pushing Docker images...$(RESET)"
	@docker push taas-api-server
	@docker push taas-carrier-connector
	@docker push taas-charging-engine
	@docker push taas-web-dashboard
	@docker push taas-packet-gateway
	@echo "$(GREEN)Docker images pushed!$(RESET)"

docker-up: ## Start all services with docker-compose
	@docker-compose up -d

docker-down: ## Stop all services
	@docker-compose down

# Kubernetes
k8s-deploy: ## Deploy to Kubernetes
	@echo "$(CYAN)Deploying to Kubernetes...$(RESET)"
	@kubectl apply -f deployments/kubernetes/
	@echo "$(GREEN)Deployment complete!$(RESET)"

k8s-deploy-helm: ## Deploy to Kubernetes using Helm
	@echo "$(CYAN)Deploying with Helm...$(RESET)"
	@helm upgrade --install telecom-platform deployments/helm/telecom-platform
	@echo "$(GREEN)Helm deployment complete!$(RESET)"

k8s-delete: ## Delete Kubernetes resources
	@kubectl delete -f deployments/kubernetes/

k8s-logs: ## Show logs for all services
	@kubectl logs -l app=telecom-platform --all-containers=true -f --tail=100

k8s-status: ## Show status of all pods
	@kubectl get pods -l app=telecom-platform

# Database setup
db-setup: ## Set up PostgreSQL, MongoDB and Redis
	@echo "$(CYAN)Setting up databases...$(RESET)"
	@bash scripts/setup-postgres.sh
	@bash scripts/setup-mongodb.sh
	@bash scripts/setup-redis.sh
	@echo "$(GREEN)Database setup complete!$(RESET)"

db-migrate: ## Run database migrations
	@echo "$(CYAN)Running database migrations...$(RESET)"
	@cd apps/api-server && go run cmd/migrate/main.go
	@echo "$(GREEN)Migrations complete!$(RESET)"

# Free5GC integration
free5gc-setup: ## Set up free5GC configuration files
	@echo "$(CYAN)Setting up free5GC configuration...$(RESET)"
	@mkdir -p deployments/free5gc/config
	@echo "$(GREEN)free5GC configuration ready!$(RESET)"
	@echo "$(YELLOW)To start free5GC: docker-compose up -d db free5gc-nrf free5gc-amf free5gc-smf free5gc-upf$(RESET)"

free5gc-start: ## Start free5GC core network services
	@echo "$(CYAN)Starting free5GC core network...$(RESET)"
	@docker-compose up -d db free5gc-nrf free5gc-amf free5gc-smf free5gc-upf free5gc-udr free5gc-udm free5gc-ausf free5gc-nssf free5gc-pcf
	@echo "$(GREEN)free5GC services started!$(RESET)"

free5gc-stop: ## Stop free5GC services
	@echo "$(CYAN)Stopping free5GC services...$(RESET)"
	@docker-compose stop db free5gc-nrf free5gc-amf free5gc-smf free5gc-upf free5gc-udr free5gc-udm free5gc-ausf free5gc-nssf free5gc-pcf
	@echo "$(GREEN)free5GC services stopped!$(RESET)"

free5gc-logs: ## Show free5GC logs
	@docker-compose logs -f free5gc-nrf free5gc-amf free5gc-smf

free5gc-test: ## Test free5GC configuration
	@echo "$(CYAN)Testing free5GC configuration...$(RESET)"
	@bash scripts/test-free5gc.sh

verify: ## Verify complete deployment status
	@echo "$(CYAN)Verifying deployment...$(RESET)"
	@bash scripts/verify-deployment.sh

.DEFAULT_GOAL := help
