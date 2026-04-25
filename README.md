# Telecom-as-a-Service (TaaS) Platform

> **Sovereign, Full-Stack Cellular Connectivity Platform**  
> Built with Go 1.26, Rust 1.94, TypeScript, eBPF, and 5G Core Network Technologies

## Overview

The Telecom-as-a-Service (TaaS) Platform is a comprehensive, sovereign cellular connectivity solution that enables organizations to deploy and manage their own private 5G/LTE networks. This full-stack platform provides end-to-end capabilities from core network integration to subscriber management, charging, and developer APIs.

### What It Does

The TaaS Platform allows enterprises, telecom operators, and system integrators to:

- **Deploy Private Cellular Networks**: Set up and manage private 5G/LTE networks with full control over data sovereignty and security
- **Manage Subscribers**: Provision, authenticate, and manage cellular subscribers with role-based access control
- **Real-time Charging**: Implement flexible credit control and billing with real-time usage monitoring
- **Developer APIs**: Expose cellular network capabilities through REST and GraphQL APIs for custom application development
- **Network Orchestration**: Automate network operations, scaling, and health monitoring across distributed infrastructure
- **Service Discovery**: Enable dynamic service registration and discovery for microservices architecture

### Architecture

The platform is built as a microservices architecture with the following core components:

**Core Network Services:**
- **API Server**: Central BSS (Business Support System) API providing authentication, subscriber management, automation, and plugin system
- **Carrier Connector**: ES2+ interface for eSIM profile management and carrier integration
- **Charging Engine**: Real-time credit control, usage tracking, and billing with Redis-backed rate limiting
- **Packet Gateway**: High-performance eBPF-based packet processing for network traffic routing and QoS enforcement

**Supporting Services:**
- **Service Discovery (Consul)**: Dynamic service registration and health checking
- **Message Queue (RabbitMQ)**: Asynchronous event-driven communication between services
- **Redis**: Distributed caching, rate limiting, and session management
- **PostgreSQL**: Persistent data storage for subscribers, automations, and configuration
- **Vault**: Secure secret management for sensitive credentials and keys

**Developer Tools:**
- **CLI**: Command-line interface for service orchestration, configuration, and health checks
- **Web Dashboard**: Next.js-based management interface for network operations
- **TypeScript SDK**: Client library for integrating with TaaS APIs
- **Kubernetes Operators**: Custom resources for deploying and managing TaaS services

### Key Features

**Sovereignty & Security:**
- Full data sovereignty with on-premise deployment
- End-to-end encryption for subscriber data
- Role-based access control (RBAC) with fine-grained permissions
- Vault-based secret management for credential security

**Performance & Scalability:**
- eBPF-accelerated packet processing for line-rate throughput
- Redis-backed distributed rate limiting and caching
- Horizontal scaling with Kubernetes orchestration
- Circuit breakers and retry logic for external service resilience

**Developer Experience:**
- REST and GraphQL APIs with comprehensive documentation
- TypeScript SDK for type-safe client integration
- Plugin system for extending platform capabilities
- Automation framework for network operations

**Operations:**
- Real-time health monitoring with Prometheus integration
- Automated scaling and service discovery
- Centralized logging with structured logs
- Chaos engineering capabilities for testing resilience


**Technology Stack:**

**Core Languages & Runtimes:**
- **Go 1.26**: Core network integration, BSS API, carrier connector
- **Rust 1.94**: eBPF packet gateway, real-time charging engine
- **TypeScript/Next.js**: Developer dashboard and SDK

**Databases:**
- **PostgreSQL**: Primary database for subscribers, automations, plugins, and configuration
- **Redis**: Real-time credit control, caching, and rate limiting
- **MongoDB**: Used exclusively by free5GC 5G core network for UDR/UDM subscription data and authentication

**5G Core Network:**
- **free5GC**: Open-source 5G core network (AMF, SMF, UDM, UDR, etc.)

**Infrastructure & Orchestration:**
- **Kubernetes**: Container orchestration and deployment
- **Helm**: Package management for Kubernetes deployments
- **Istio**: Service mesh for traffic management and security
- **eBPF/Aya**: High-performance packet processing in kernel space

**Message Queuing & Service Discovery:**
- **RabbitMQ**: Asynchronous message queue for event-driven communication
- **Consul**: Service discovery, health checking, and configuration

**Security & Secrets:**
- **Vault**: Secure secret management for credentials and keys
- **cert-manager**: Automated TLS certificate management

**Monitoring & Observability:**
- **Prometheus**: Metrics collection and alerting
- **Grafana**: Visualization dashboards for metrics
- **ELK Stack**: Elasticsearch, Logstash, Kibana for centralized logging

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

### Installation

```bash
# Clone the repository
git clone https://github.com/nutcas3/telecom-platform.git
cd telecom-platform

# Build all components
make all

# Start services (requires separate terminals)
# Terminal 1: API Server
./dist/api-server

# Terminal 2: Carrier Connector
./dist/carrier-connector

# Terminal 3: Charging Engine
./target/release/charging-engine

# Terminal 4: Web Dashboard
cd apps/web-dashboard && pnpm dev
```

## Project Structure

```
telecom-platform/
├── apps/
│   ├── api-server/          # Go: Developer BSS API
│   ├── carrier-connector/   # Go: eSIM ES2+ Provisioning
│   ├── charging-engine/     # Rust: OCS Real-time Credit Control
│   ├── packet-gateway/      # Rust: eBPF UPF Data Plane
│   └── web-dashboard/       # TypeScript: Next.js Frontend
├── libs/
│   ├── shared-ts-sdk/       # TypeScript: Drop-in Widget SDK
│   └── proto/               # Shared Protobufs
├── deployments/
│   ├── kubernetes/          # K8s manifests
│   └── docker/              # Dockerfiles
├── docs/                    # Architecture & API docs
├── scripts/                 # Automation scripts
├── Makefile                 # Master build orchestrator
├── Cargo.toml              # Rust workspace config
├── go.work                 # Go workspace config
└── pnpm-workspace.yaml     # TypeScript workspace config
```

## Development

### Build Commands

```bash
# Build everything
make all

# Build specific language
make build-go        # Go services
make build-rust      # Rust components
make build-ui        # TypeScript dashboard

# Run tests
make test

# Clean artifacts
make clean
```

### Working with Specific Components

**Go Services:**
```bash
cd apps/api-server
go run main.go
```

**Rust Components:**
```bash
cd apps/charging-engine
cargo run --release
```

**TypeScript Dashboard:**
```bash
cd apps/web-dashboard
pnpm dev
```

## Documentation

- **[API Documentation](./docs/api-spec.yaml)**: OpenAPI 3.0 specification
- **[Architecture Guide](./docs/architecture.md)**: System design and data flows
- **[Deployment Guide](./docs/deployment.md)**: Kubernetes and Docker setup

## Testing

```bash
# Unit tests
make test

# Go tests
go test ./apps/...

# Rust tests
cargo test --workspace

# TypeScript tests
pnpm -r test
```

## Deployment

### Docker

```bash
make docker-build
docker-compose up -d
```

### Kubernetes

```bash
kubectl apply -f deployments/kubernetes/
```

## Environment Variables

Create `.env` files in each service directory:

**API Server** (`apps/api-server/.env`):
```env
MONGODB_URI=mongodb://localhost:27017/free5gc
REDIS_URI=redis://localhost:6379
API_PORT=8000
JWT_SECRET=your-secret-key
```

**Charging Engine** (`apps/charging-engine/.env`):
```env
REDIS_URI=redis://localhost:6379
SERVER_PORT=8080
```

**Web Dashboard** (`apps/web-dashboard/.env.local`):
```env
NEXT_PUBLIC_API_URL=http://localhost:8000
```

## 📖 Key Resources

- **free5GC**: [https://free5gc.org/](https://free5gc.org/)
- **Aya eBPF**: [https://aya-rs.dev/](https://aya-rs.dev/)
- **GSMA eSIM Specs**: [https://www.gsma.com/esim/](https://www.gsma.com/esim/)
- **Go 1.26 Docs**: [https://go.dev/doc/go1.26](https://go.dev/doc/go1.26)
- **Rust 1.94 Blog**: [https://blog.rust-lang.org/2026/03/05/Rust-1.94.0/](https://blog.rust-lang.org/2026/03/05/Rust-1.94.0/)

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/yourorg/telecom-platform/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourorg/telecom-platform/discussions)
- **Email**: support@yourcompany.com

## Acknowledgments

- **free5GC Team**: For the open-source 5G core implementation
- **Aya Community**: For the pure-Rust eBPF framework
- **GSMA**: For standardizing eSIM technology
- **Go Team**: For the amazing Go 1.26 release
- **Rust Team**: For continuous language improvements

