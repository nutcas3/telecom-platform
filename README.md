# Telecom-as-a-Service (TaaS) Platform

> **Sovereign, Full-Stack Cellular Connectivity Platform**  
> Built with Go 1.26, Rust 1.95, TypeScript, eBPF, and 5G Core Network Technologies

## Overview

The Telecom-as-a-Service (TaaS) Platform is a comprehensive, sovereign cellular connectivity solution that enables organizations to deploy and manage their own private 5G/LTE networks. This full-stack platform provides end-to-end capabilities from core network integration to subscriber management, charging, and developer APIs.

### API Gateway Integration

The platform now includes **Traefik API Gateway** for centralized request routing, security, and monitoring. All services are accessible through a unified HTTPS endpoint with built-in rate limiting, authentication, and SSL termination.

### What It Does

The TaaS Platform allows enterprises, telecom operators, and system integrators to:

- **Deploy Private Cellular Networks**: Set up and manage private 5G/LTE networks with full control over data sovereignty and security
- **Manage Subscribers**: Provision, authenticate, and manage cellular subscribers with role-based access control
- **Real-time Charging**: Implement flexible credit control and billing with real-time usage monitoring
- **Developer APIs**: Expose cellular network capabilities through REST and GraphQL APIs for custom application development
- **Network Orchestration**: Automate network operations, scaling, and health monitoring across distributed infrastructure
- **Service Discovery**: Enable dynamic service registration and discovery for microservices architecture

### Commercial Applications

The platform is designed for various commercial use cases:

**eSIM Operators (Airalo-style)**
- Multi-carrier aggregation across 400+ global carriers
- Real-time eSIM provisioning via GSMA ES2+ standards
- Usage-based billing with global rate plans
- B2B2C model for MVNO partnerships

**Enterprise Private Networks**
- Industrial IoT and manufacturing connectivity
- Campus networks for universities and hospitals
- Critical infrastructure communications
- Secure data sovereignty deployments

**Telecom Service Providers**
- MVNO enablement platform
- Network slicing as a service
- Edge computing integration
- 5G core network hosting

### Architecture

The platform is built as a microservices architecture with the following core components:

**API Gateway Layer:**
- **Traefik API Gateway**: Centralized entry point providing SSL termination, rate limiting, authentication, and request routing
- **Unified HTTPS Endpoint**: All services accessible via `https://api.telecom.com`
- **Security Middleware**: JWT authentication, security headers, compression, and retry logic
- **Monitoring Dashboard**: Real-time metrics and service health visualization

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
- **Multi-Language SDKs**: Client libraries for Go, Python, TypeScript, Kotlin, Ruby, Swift, Rust, and Elixir
- **Kubernetes Operators**: Custom resources for deploying and managing TaaS services

**Analytics & Intelligence:**
- **Churn Analysis**: ML-powered customer churn prediction with risk scoring and retention recommendations
- **Fraud Detection**: Real-time fraud detection for account takeover, subscription fraud, payment fraud, and SIM swap attacks
- **Market Analytics**: Market penetration analysis, competitor tracking, and growth opportunity identification
- **Predictive Maintenance**: Infrastructure health monitoring with failure prediction and maintenance scheduling
- **Pricing Optimization**: Dynamic pricing strategies for revenue maximization, market share, and churn reduction

### Key Features

**API Gateway & Security:**
- **Unified HTTPS Endpoint**: All services accessible via `https://api.telecom.com`
- **Centralized Authentication**: JWT validation with rate limiting per service
- **SSL Termination**: Automatic HTTPS with security headers enforcement
- **Request Routing**: Intelligent routing with circuit breakers and retry logic

**Sovereignty & Security:**
- Full data sovereignty with on-premise deployment
- End-to-end encryption for subscriber data
- Role-based access control (RBAC) with fine-grained permissions
- Vault-based secret management for credential security

**Performance & Scalability:**
- eBPF-accelerated packet processing for line-rate throughput
- Redis-backed distributed rate limiting and caching
- Horizontal scaling with Kubernetes orchestration
- Gateway-level load balancing and connection pooling

**Developer Experience:**
- **Single API Endpoint**: Simplified client integration through gateway
- REST and GraphQL APIs with comprehensive documentation
- TypeScript SDK for type-safe client integration
- Plugin system for extending platform capabilities
- Automation framework for network operations

**Operations:**
- **Gateway Dashboard**: Real-time monitoring of all services
- **Unified Metrics**: Prometheus integration with gateway-level insights
- Automated scaling and service discovery
- Centralized logging with structured logs
- Health checks and failover automation


**Technology Stack:**

**Core Languages & Runtimes:**
- **Go 1.26**: Core network integration, BSS API, carrier connector
- **Rust 1.95**: eBPF packet gateway, real-time charging engine
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
- **Rust 1.95+**: `curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh`
- **Node.js 22+**: [Download](https://nodejs.org/)
- **pnpm**: `npm install -g pnpm`
- **Docker**: [Install](https://docs.docker.com/get-docker/)
- **PostgreSQL 15+**: [Install](https://www.postgresql.org/download/)
- **Redis**: [Install](https://redis.io/download)
- **Kubernetes**: [Install](https://kubernetes.io/docs/tasks/tools/) (for production deployment)
- **Helm 3+**: [Install](https://helm.sh/docs/intro/install/) (for production deployment)
- **MongoDB 7.0+**: [Install](https://www.mongodb.com/docs/manual/installation/) (required for free5GC only)

### Installation

#### Quick Start with API Gateway (Recommended)

```bash
# Clone the repository
git clone https://github.com/nutcas3/telecom-platform.git
cd telecom-platform

# Start with API Gateway (includes all services)
./scripts/start-gateway.sh

# Add domain to hosts file
echo "127.0.0.1 api.telecom.com" | sudo tee -a /etc/hosts

# Access services
# API Gateway Dashboard: http://localhost:8080
# API Documentation: https://api.telecom.com/api/v1/swagger
# Web Dashboard: http://localhost:3000
```

#### Manual Installation

```bash
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
|-- apps/
|   |-- api-server/          # Go: Developer BSS API
|   |-- carrier-connector/   # Go: eSIM ES2+ Provisioning
|   |-- charging-engine/     # Rust: OCS Real-time Credit Control
|   |-- packet-gateway/      # Rust: eBPF UPF Data Plane
|   |-- web-dashboard/       # TypeScript: Next.js Frontend
|-- sdk/
|   |-- go/                  # Go SDK
|   |-- python/              # Python SDK
|   |-- typescript/          # TypeScript SDK
|   |-- kotlin/              # Kotlin SDK
|   |-- ruby/                # Ruby SDK
|   |-- swift/               # Swift SDK
|   |-- rust/                # Rust SDK
|   |-- elixir/              # Elixir SDK
|-- libs/
|   |-- shared-ts-sdk/       # TypeScript: Drop-in Widget SDK
|   |-- proto/               # Shared Protobufs
|-- deployments/
|   |-- kubernetes/          # K8s manifests
|   |-- docker/              # Dockerfiles
|-- traefik/                 # API Gateway configuration
|   |-- traefik.yml          # Static configuration
|   |-- dynamic/             # Dynamic middleware config
|-- docs/                    # Architecture & API docs
|   |-- sdk-usage.md         # Multi-language SDK documentation
|   |-- gateway-quickstart.md # API Gateway guide
|   |-- api-gateway.md       # Gateway implementation details
|-- scripts/                 # Automation scripts
|   |-- start-gateway.sh     # Gateway startup script
|-- docker-compose.yml       # Container orchestration
|-- Makefile                 # Master build orchestrator
|-- Cargo.toml              # Rust workspace config
|-- go.work                 # Go workspace config
|-- pnpm-workspace.yaml     # TypeScript workspace config
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
- **[SDK Documentation](./docs/sdk-usage.md)**: Multi-language SDK usage guide
- **[Building 5G Networks](./docs/building-5g-with-taas.md)**: Complete 5G deployment guide
- **[eSIM Operator Analysis](./docs/esim-operator-analysis.md)**: Commercial use case analysis
- **[Gateway Quickstart](./docs/gateway-quickstart.md)**: API Gateway setup and configuration
- **[API Gateway Guide](./docs/api-gateway.md)**: Implementation details and patterns
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

### Docker with API Gateway (Recommended)

```bash
# Start with API Gateway
./scripts/start-gateway.sh

# Or manually
docker-compose up -d
```

### Docker (Legacy)

```bash
make docker-build
docker-compose up -d
```

### Kubernetes

```bash
kubectl apply -f deployments/kubernetes/
```

### API Gateway Configuration

The API Gateway provides:

- **Unified Endpoint**: `https://api.telecom.com`
- **Rate Limiting**: Per-service rate limits
- **SSL Termination**: Automatic HTTPS
- **Authentication**: JWT validation
- **Monitoring**: Real-time metrics dashboard

For detailed setup, see [Gateway Quickstart Guide](./docs/gateway-quickstart.md)

## Platform Architecture & Components

### **API Gateway Layer**
- **Traefik API Gateway**: Centralized entry point providing SSL termination, rate limiting, authentication, and request routing
- **Unified HTTPS Endpoint**: All services accessible via `https://api.telecom.com`
- **Security Middleware**: JWT authentication, security headers, compression, and retry logic
- **Monitoring Dashboard**: Real-time metrics and service health visualization

### **Core Network Services**

#### **API Server (Go/Gin)**
- **Purpose**: Central BSS (Business Support System) API
- **Features**: Authentication, subscriber management, automation, plugin system
- **Architecture**: Microservices with Gin framework, PostgreSQL, Redis caching
- **Key Modules**: Handlers for analytics, payments, monitoring, RBAC, websockets

#### **Carrier Connector (Go/Gin)**
- **Purpose**: ES2+ interface for eSIM profile management and carrier integration
- **Features**: Multi-carrier aggregation, GSMA ES2+ standards compliance, real-time eSIM provisioning
- **Architecture**: GORM for database, ES2+ client, message queue integration
- **Key Modules**: Pricing optimization, security (fraud detection), rate plans, MVNO support

#### **Charging Engine (Rust/Axum)**
- **Purpose**: Real-time credit control, usage tracking, and billing
- **Features**: Redis-backed rate limiting, PostgreSQL for rate plans, circuit breakers
- **Architecture**: High-performance Rust with tokio async runtime
- **Key Modules**: Charging handlers, authentication, monitoring, rating plans

#### **Packet Gateway (Rust/eBPF)**
- **Purpose**: High-performance packet processing for network traffic routing and QoS enforcement
- **Features**: eBPF-accelerated packet processing for line-rate throughput

### **Supporting Infrastructure**
- **PostgreSQL**: Persistent data storage for subscribers, automations, configuration
- **Redis**: Distributed caching, rate limiting, session management
- **MongoDB**: Document storage for 5G core network data
- **RabbitMQ**: Asynchronous event-driven communication
- **Consul**: Service discovery and health checking
- **Vault**: Secure secret management

### **Frontend Applications**

#### **Web Dashboard (Next.js/TypeScript)**
- **Purpose**: Management interface for network operations
- **Features**: Real-time dashboard, subscriber management, analytics, pricing optimization
- **Architecture**: React components, Tailwind CSS, API integration
- **Key Pages**: Dashboard, analytics, pricing, subscribers, system health

### **SDK Ecosystem**

Multi-language SDKs for developer integration:
- **Swift**: iOS/macOS applications with async/await support
- **Python**: Backend integration and automation
- **TypeScript**: Web applications and Node.js backends
- **Go**: Microservices and CLI tools
- **Kotlin**: Android applications
- **Rust**: High-performance systems
- **Elixir**: Phoenix applications
- **Ruby**: Rails integration

### **Analytics & Intelligence**

#### **Advanced Analytics Modules**
1. **Churn Analysis**: ML-powered customer churn prediction with risk scoring
2. **Fraud Detection**: Real-time fraud detection (account takeover, subscription fraud, SIM swap attacks)
3. **Market Analytics**: Market penetration analysis, competitor tracking
4. **Predictive Maintenance**: Infrastructure health monitoring with failure prediction
5. **Pricing Optimization**: Dynamic pricing strategies with elasticity calculations

#### **Pricing Optimization System**
- **Strategies**: Revenue maximization, market share, profit margin, competitive positioning, churn reduction
- **Advanced Calculations**: 
  - Dynamic elasticity based on rate plan characteristics
  - Competitive index with seasonal market analysis
  - ROI calculation with period-based adjustments
- **Implementation**: Go services with mathematical modeling and bounded realistic values

### **Commercial Applications**

#### **eSIM Operators (Airalo-style)**
- Multi-carrier aggregation across 400+ global carriers
- Real-time eSIM provisioning via GSMA ES2+ standards
- Usage-based billing with global rate plans
- B2B2C model for MVNO partnerships

#### **Enterprise Private Networks**
- Industrial IoT and manufacturing connectivity
- Campus networks for universities and hospitals
- Critical infrastructure communications
- Secure data sovereignty deployments

#### **Telecom Service Providers**
- MVNO enablement platform
- Network slicing as a service
- Edge computing integration
- 5G core network hosting

### **Data Flow Architecture**

```
Client Applications → Traefik Gateway → API Services → Backend Services
                              ↓
                        Authentication & Rate Limiting
                              ↓
                    Message Queue (RabbitMQ) for Async Events
                              ↓
              Database Layer (PostgreSQL, Redis, MongoDB)
```

### **Key Features Summary**

- **Sovereignty & Security**: Full data sovereignty, end-to-end encryption, RBAC
- **Performance**: eBPF-accelerated packet processing, Redis-backed caching
- **Scalability**: Microservices architecture, horizontal scaling
- **Developer Experience**: Multi-language SDKs, comprehensive documentation
- **Enterprise Ready**: Monitoring, backup, security, compliance features

The platform represents a **complete telecom stack** for modern cellular network operations, combining carrier-grade reliability with cloud-native architecture and advanced analytics capabilities.

## API Endpoints

### Analytics API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/analytics/churn/predict` | Predict churn for a profile |
| GET | `/api/v1/analytics/churn/metrics` | Get churn metrics |
| GET | `/api/v1/analytics/churn/at-risk` | Get at-risk customers |
| GET | `/api/v1/analytics/market/metrics` | Get market metrics |
| GET | `/api/v1/analytics/market/competitors` | Get competitor analysis |
| GET | `/api/v1/analytics/market/opportunities` | Get market opportunities |
| GET | `/api/v1/analytics/maintenance/metrics` | Get maintenance metrics |
| GET | `/api/v1/analytics/maintenance/assets` | Get assets health |
| GET | `/api/v1/analytics/maintenance/alerts` | Get maintenance alerts |
| POST | `/api/v1/analytics/maintenance/predict/:asset_id` | Predict asset failure |
| GET | `/api/v1/analytics/pricing/metrics` | Get pricing metrics |
| POST | `/api/v1/analytics/pricing/optimize` | Optimize pricing |
| GET | `/api/v1/analytics/pricing/elasticity` | Get price elasticity |

### Security API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/security/fraud/analyze` | Analyze transaction for fraud |
| POST | `/api/v1/security/fraud/alerts` | Get fraud alerts |
| PUT | `/api/v1/security/fraud/alerts/:id` | Update alert status |
| GET | `/api/v1/security/fraud/metrics` | Get fraud metrics |
| GET | `/api/v1/security/fraud/patterns` | Get fraud patterns |
| POST | `/api/v1/security/simswap/verify` | Verify SIM swap |
| GET | `/api/v1/security/simswap/history/:profile_id` | Get SIM swap history |

### Currency & Billing API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/currency/convert` | Convert currency |
| GET | `/api/v1/currency/exchange/:from/:to` | Get exchange rate |
| GET | `/api/v1/currency/exchange/:from/:to/history` | Get exchange rate history |
| GET | `/api/v1/currency/currencies` | List supported currencies |
| POST | `/api/v1/currency/exchange/refresh` | Refresh exchange rates |
| POST | `/api/v1/currency/billing` | Process billing |
| GET | `/api/v1/currency/billing/history/:profile_id` | Get billing history |
| GET | `/api/v1/currency/billing/summary/:profile_id` | Get billing summary |
| POST | `/api/v1/currency/billing/refund/:transaction_id` | Process refund |
| GET | `/api/v1/currency/billing/analytics` | Get billing analytics |

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
- **Rust 1.95 Blog**: [https://blog.rust-lang.org/2026/06/05/Rust-1.95.0/](https://blog.rust-lang.org/2026/06/05/Rust-1.95.0/)
- **Traefik**: [https://traefik.io/](https://traefik.io/)

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/nutcas3/telecom-platform/issues)
- **Discussions**: [GitHub Discussions](https://github.com/nutcas3/telecom-platform/discussions)
- **Email**: support@yourcompany.com

## Acknowledgments

- **free5GC Team**: For the open-source 5G core implementation
- **Aya Community**: For the pure-Rust eBPF framework
- **GSMA**: For standardizing eSIM technology
- **Go Team**: For the amazing Go 1.26 release
- **Rust Team**: For continuous language improvements

