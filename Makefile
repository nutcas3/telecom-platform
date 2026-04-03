# Master Makefile - Telecom Platform
# Orchestrates builds across Go, Rust, and TypeScript

.PHONY: all build-ui build-rust build-go clean test dev help install-deps

# Colors for terminal output
CYAN := \033[0;36m
GREEN := \033[0;32m
YELLOW := \033[1;33m
RESET := \033[0m

help: ## Show this help message
	@echo "$(CYAN)Telecom Platform - Available Commands:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'

all: build-ui build-rust build-go ## Build all components

install-deps: ## Install all dependencies
	@echo "$(CYAN)Installing Go dependencies...$(RESET)"
	@cd apps/api-server && go mod download
	@cd apps/carrier-connector && go mod download
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

build-rust-dev: ## Build Rust components (debug mode)
	@cargo build

test-rust: ## Run Rust tests
	@cargo test --workspace

lint-rust: ## Run Rust linter (clippy)
	@cargo clippy --workspace -- -D warnings

# Go - BSS API & Carrier Connector
build-go: ## Build Go services
	@echo "$(CYAN)Building Go projects...$(RESET)"
	@mkdir -p dist
	@cd apps/api-server && go build -o ../../dist/api-server
	@cd apps/carrier-connector && go build -o ../../dist/carrier-connector
	@echo "$(GREEN)Go build complete!$(RESET)"

test-go: ## Run Go tests
	@go test ./apps/...

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
	@echo "$(GREEN)Docker images built!$(RESET)"

docker-up: ## Start all services with docker-compose
	@docker-compose up -d

docker-down: ## Stop all services
	@docker-compose down

# Kubernetes
k8s-deploy: ## Deploy to Kubernetes
	@kubectl apply -f deployments/kubernetes/

k8s-delete: ## Delete Kubernetes resources
	@kubectl delete -f deployments/kubernetes/

# Database setup
db-setup: ## Set up MongoDB and Redis
	@echo "$(CYAN)Setting up databases...$(RESET)"
	@bash scripts/setup-mongodb.sh
	@bash scripts/setup-redis.sh
	@echo "$(GREEN)Database setup complete!$(RESET)"

# Free5GC integration
free5gc-setup: ## Set up free5GC (requires separate installation)
	@echo "$(YELLOW)Please install free5GC separately.$(RESET)"
	@echo "Visit: https://free5gc.org/guide/"

.DEFAULT_GOAL := help
