# Telecom-as-a-Service (TaaS) Platform

> **Sovereign, Full-Stack Cellular Connectivity Platform**  
> Built with Go 1.26, Rust 1.94, TypeScript, eBPF, and 5G Core Network Technologies

## Overview

This is a production-ready Telecom-as-a-Service platform that enables developers to programmatically provision and manage cellular connectivity through a clean API. The platform implements 3GPP standards using modern open-source technologies.

**Technology Stack:**
- **Go 1.26**: Core network integration, BSS API, carrier connector
- **Rust 1.94**: eBPF packet gateway, real-time charging engine
- **TypeScript/Next.js**: Developer dashboard and SDK
- **free5GC**: Open-source 5G core network
- **MongoDB**: Subscriber data and authentication
- **Redis**: Real-time credit control
- **eBPF/Aya**: High-performance packet processing

## Quick Start

### Prerequisites

- **Go 1.26+**: [Download](https://go.dev/dl/)
- **Rust 1.94+**: `curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh`
- **Node.js 22+**: [Download](https://nodejs.org/)
- **pnpm**: `npm install -g pnpm`
- **Docker**: [Install](https://docs.docker.com/get-docker/)
- **MongoDB 7.0+**: [Install](https://www.mongodb.com/docs/manual/installation/)
- **Redis**: `sudo apt-get install redis-server`

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

