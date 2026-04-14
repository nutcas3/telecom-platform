# Quick Start Guide

## Overview

This guide will help you get the Telecom-as-a-Service platform running in minutes.

## Prerequisites

- Docker & Docker Compose
- Make command
- Terminal access

## 1. Initial Setup

```bash
# Clone the repository (if not already done)
git clone https://github.com/nutcas3/telecom-platform.git
cd telecom-platform

# Install dependencies
make install-deps
```

## 2. Start Core Services

```bash
# Start databases and core services
docker-compose up -d mongodb redis

# Start free5GC core network
make free5gc-start

# Start application services
docker-compose up -d api-server charging-engine carrier-connector
```

## 3. Start Development Server

```bash
# Start web dashboard
make dev-ui
```

## 4. Verify Deployment

```bash
# Check all services status
make verify

# Test free5GC configuration
make free5gc-test
```

## 5. Access Services

- **Web Dashboard**: http://localhost:3000
- **API Server**: http://localhost:8000
- **Charging Engine**: http://localhost:8080

## 6. Common Commands

```bash
# View free5GC logs
make free5gc-logs

# Stop all services
docker-compose down

# Rebuild services
make docker-build

# Clean everything
make clean
```

## Service Architecture

```
Internet
    |
[Web Dashboard] - [API Server] - [Carrier Connector]
    |               |               |
    |           [Charging Engine]  |
    |               |               |
    +-----------[free5GC Core]-----+
                    |
                [MongoDB] [Redis]
```

## Troubleshooting

### Services Not Starting
```bash
# Check logs
docker-compose logs [service-name]

# Restart specific service
docker-compose restart [service-name]
```

### Port Conflicts
Ensure ports 3000, 8000, 8080, 27017, 6379 are available.

### Permission Issues
```bash
# Fix script permissions
chmod +x scripts/*.sh
```

## Next Steps

1. Explore the API documentation at `/docs/API.md`
2. Configure your first subscriber in the web dashboard
3. Test the charging engine with sample data
4. Deploy to Kubernetes using the provided manifests

## Support

- Check `make help` for all available commands
- Review `docs/DEPLOYMENT.md` for production deployment
- Check `todos.md` for development roadmap
