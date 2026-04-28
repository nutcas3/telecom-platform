# Getting Started

This guide helps you get the Telecom-as-a-Service (TaaS) platform running locally and make your first API call.

## Prerequisites

- Docker & Docker Compose
- Go 1.26+
- Rust 1.94+
- Node.js 22+
- MongoDB 7.0+ / PostgreSQL 15+
- Redis 7+
- pnpm
- Make

## Quick Start (Docker Compose)

The fastest path to a running platform:

```bash
git clone https://github.com/nutcas3/telecom-platform.git
cd telecom-platform

# Install deps and start core services
make install-deps
docker-compose up -d mongodb redis
make free5gc-start
docker-compose up -d api-server charging-engine carrier-connector

# Start the dashboard
make dev-ui

# Verify everything is running
make verify
```

### Service URLs

| Service | URL |
|---------|-----|
| Web Dashboard | http://localhost:3000 |
| API Server | http://localhost:8000 |
| Charging Engine | http://localhost:8080 |
| MongoDB | mongodb://localhost:27017 |
| Redis | redis://localhost:6379 |

## Local Development (Without Docker)

```bash
# 1. Setup
chmod +x scripts/dev-setup.sh
./scripts/dev-setup.sh
make db-setup
make all

# 2. Start services (separate terminals)
./dist/api-server
./dist/carrier-connector
./target/release/charging-engine
cd apps/web-dashboard && pnpm dev
```

### Environment Variables

**`apps/api-server/.env`:**
```env
MONGODB_URI=mongodb://localhost:27017/free5gc
REDIS_URI=redis://localhost:6379
API_PORT=8000
GIN_MODE=debug
JWT_SECRET=your-dev-secret
```

**`apps/charging-engine/.env`:**
```env
REDIS_URI=redis://127.0.0.1/
SERVER_PORT=8080
RUST_LOG=info
```

**`apps/web-dashboard/.env.local`:**
```env
NEXT_PUBLIC_API_URL=http://localhost:8000
```

## Your First API Call

### 1. Authenticate

```bash
# Register
curl -X POST http://localhost:8000/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "password123",
    "first_name": "Admin",
    "last_name": "User"
  }'

# Login
TOKEN=$(curl -s -X POST http://localhost:8000/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123"}' \
  | jq -r .token)
```

### 2. Create an eSIM

```bash
curl -X POST http://localhost:8000/v1/esims \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "data_plan": "1GB",
    "country_code": "US"
  }'
```

Response:
```json
{
  "esim_id": "550e8400-e29b-41d4-a716-446655440000",
  "imsi": "208930000000001",
  "iccid": "8933123456789012345",
  "activation_code": "LPA:1$smdp.example.com$MATCHING_ID",
  "status": "provisioned"
}
```

### 3. Check Usage

```bash
curl http://localhost:8000/v1/esims/$ESIM_ID/usage \
  -H "Authorization: Bearer $TOKEN"
```

## API Reference

### Base URL

- Production: `https://api.taas-platform.com/v1`
- Development: `http://localhost:8000/v1`

### Authentication

All protected endpoints require a Bearer token:

```bash
curl -H "Authorization: Bearer $TOKEN" ...
```

### Core Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | System health |
| GET | `/ready` | Readiness probe |
| GET | `/live` | Liveness probe |
| POST | `/v1/auth/register` | Create user |
| POST | `/v1/auth/login` | Obtain tokens |
| POST | `/v1/auth/refresh` | Refresh access token |
| POST | `/v1/esims` | Create eSIM |
| GET | `/v1/esims/:id` | Get eSIM |
| GET | `/v1/esims/:id/usage` | Get usage |
| DELETE | `/v1/esims/:id` | Terminate eSIM |
| GET | `/v1/services` | List services |
| GET | `/v1/monitoring/metrics` | System metrics |
| GET | `/v1/monitoring/alerts` | Active alerts |
| GET | `/v1/billing/invoices` | List invoices |
| GET | `/v1/config` | Get config |
| GET | `/metrics/performance` | Performance stats |
| GET | `/ws` | WebSocket real-time updates |

### Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 201 | Created |
| 400 | Bad request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not found |
| 429 | Rate limited |
| 500 | Server error |
| 503 | Service unavailable |

### Rate Limits

| Tier | Limit | Burst |
|------|-------|-------|
| Auth endpoints | 10/min | 5 |
| Read endpoints | 200/min | 50 |
| Admin endpoints | 50/min | 10 |
| Default | 100/min | 20 |

Response headers include `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`.

## SDKs

### TypeScript

```bash
npm install @taas-platform/sdk
```

```typescript
import { TaasClient } from '@taas-platform/sdk';

const client = new TaasClient({ apiKey: process.env.TAAS_API_KEY });
const esim = await client.esims.create({ dataPlan: '1GB', countryCode: 'US' });
const usage = await client.esims.getUsage(esim.esimId);
```

### Go

```bash
go get github.com/taas-platform/sdk-go
```

```go
import "github.com/taas-platform/sdk-go"

client := taas.NewClient(os.Getenv("TAAS_API_KEY"))
esim, err := client.ESims.Create(&taas.CreateESIMRequest{
    DataPlan:    "1GB",
    CountryCode: "US",
})
```

## Webhooks

Configure webhooks to receive real-time events:

| Event | Trigger |
|-------|---------|
| `esim.created` | New eSIM provisioned |
| `esim.activated` | eSIM activated on device |
| `usage.threshold` | 50%/75%/90% of data used |
| `usage.depleted` | Data plan exhausted |
| `esim.terminated` | eSIM terminated |

**Payload example:**
```json
{
  "event": "usage.threshold",
  "timestamp": "2026-04-01T12:00:00Z",
  "data": {
    "esim_id": "550e8400-...",
    "threshold": 75,
    "data_used": 805306368,
    "data_limit": 1073741824
  }
}
```

## Development Workflow

### Running Tests

```bash
# Run all tests
make test

# Run API server tests
cd apps/api-server && go test ./...

# Run integration tests
cd apps/api-server && go test ./internal/handlers -v

# Run service layer tests
cd apps/api-server && go test ./internal/services -v

# Run with coverage
cd apps/api-server && go test -cover ./...

# Run specific test
cd apps/api-server && go test ./internal/handlers -run TestAuthenticationEndpoints -v
```

### Code Quality

```bash
# Format code
cd apps/api-server && go fmt ./...
cd apps/web-dashboard && pnpm format

# Lint code
cd apps/api-server && golangci-lint run
cd apps/web-dashboard && pnpm lint

# Run static analysis
cd apps/api-server && go vet ./...
```

### Database Operations

```bash
# Run database migrations
make db-migrate

# Reset database (development only)
make db-reset

# Create test data
make db-seed

# View database logs
docker-compose logs mongodb
docker-compose logs postgres
```

### Monitoring & Debugging

```bash
# View service logs
make logs
make logs-api-server
make logs-charging-engine

# Monitor metrics
curl http://localhost:8000/metrics
curl http://localhost:8080/metrics

# Check health endpoints
curl http://localhost:8000/health
curl http://localhost:8000/ready
curl http://localhost:8000/live
```

## Common Commands

```bash
make help              # List all targets
make all               # Build everything
make verify            # Check services
make free5gc-logs      # View free5GC logs
make docker-build      # Build Docker images
make clean             # Remove build artifacts
make test              # Run all tests
make logs              # View all service logs
make dev-ui            # Start web dashboard in dev mode
make dev-api           # Start API server in dev mode
```

## Troubleshooting

### Services won't start
```bash
docker-compose logs [service-name]
docker-compose restart [service-name]
```

### Port conflicts
Ensure ports 3000, 8000, 8080, 27017, 6379 are free.

### Permission issues
```bash
chmod +x scripts/*.sh
```

### Database connection errors
- Verify MongoDB/PostgreSQL is running: `docker-compose ps`
- Check connection strings in `.env`
- For PostgreSQL: confirm `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` are set

## Recent Improvements

### Enhanced Error Handling
The platform now includes comprehensive error handling with standardized error responses:

- **Structured error codes** for different error types
- **Field-specific validation errors** with detailed messages
- **Consistent error response format** across all endpoints
- **Request correlation IDs** for debugging

Example error response:
```json
{
  "error": "Validation failed",
  "code": "VALIDATION_FAILED",
  "details": "Invalid input data",
  "errors": [
    {
      "field": "email",
      "message": "must be a valid email address",
      "value": "invalid-email"
    }
  ]
}
```

### Comprehensive Testing
The platform now has extensive test coverage:

- **Integration tests** for all API endpoints
- **Service layer tests** for business logic
- **Authentication and RBAC testing** with mock users
- **Error handling validation** across all scenarios

Run tests with:
```bash
cd apps/api-server && go test ./internal/handlers -v
cd apps/api-server && go test ./internal/services -v
```

### Role-Based Access Control (RBAC)
Enhanced RBAC implementation using Casbin:

- **Policy-based permissions** with fine-grained control
- **Default roles**: admin, operator, viewer
- **Dynamic policy management** via API
- **Fallback role middleware** for compatibility

Example policy check:
```go
allowed, err := casbinSvc.CheckPermission(userID, "/v1/services", "GET")
```

## Next Steps

- **Architecture**: See [`ARCHITECTURE.md`](./ARCHITECTURE.md) for system design, RBAC, and security.
- **Operations**: See [`OPERATIONS.md`](./OPERATIONS.md) for deployment, monitoring, and performance tuning.

## Support

- Issues: https://github.com/nutcas3/telecom-platform/issues
- Discussions: https://github.com/nutcas3/telecom-platform/discussions
