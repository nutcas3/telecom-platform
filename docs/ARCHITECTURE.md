# Architecture

Comprehensive overview of the TaaS Platform system design, components, data flow, security, and RBAC.

## Table of Contents

1. [System Overview](#system-overview)
2. [Component Details](#component-details)
3. [Data Flow](#data-flow)
4. [Scaling](#scaling)
5. [Security & RBAC](#security--rbac)
6. [Future Enhancements](#future-enhancements)

## System Overview

The TaaS Platform is a multi-layer telecom infrastructure that provides cellular connectivity as a service, implementing 3GPP standards while exposing a modern, developer-friendly API.

### Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                     Developer Interface                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Web UI      │  │  TypeScript  │  │  REST API    │      │
│  │  Dashboard   │  │  SDK         │  │  Webhooks    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                   Business Support System (BSS)              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  API Server  │  │  Carrier     │  │  Subscriber  │      │
│  │  (Go 1.26)   │  │  Connector   │  │  Management  │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│              Operations Support System (OSS)                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Charging    │  │  Usage       │  │  eSIM        │      │
│  │  Engine      │  │  Tracking    │  │  Provisioning│      │
│  │  (Rust)      │  │  (Redis)     │  │  (ES2+)      │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                  Telecom Core Network                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  5G Core     │  │  Packet      │  │  Data Plane  │      │
│  │  (free5GC)   │  │  Gateway     │  │  (eBPF)      │      │
│  │  (Go)        │  │  (Rust)      │  │              │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

## Component Details

### 1. API Server (Go 1.26)

**Purpose**: Developer-facing RESTful API and business logic.

**Stack**:
- Gin framework
- GORM with PostgreSQL (primary) / MongoDB (free5GC subscriber data)
- Redis (session/cache)
- Casbin for RBAC
- WebSocket for real-time updates

**Responsibilities**:
- eSIM lifecycle management
- Authentication (JWT) and authorization (Casbin)
- Rate limiting, quotas, and performance monitoring
- Webhook delivery
- Health checks and alerting

**Key Endpoints**:
- `/v1/auth/*` — authentication
- `/v1/esims/*` — eSIM management
- `/v1/services/*` — service operations
- `/v1/monitoring/*` — metrics & alerts
- `/v1/billing/*` — invoicing
- `/ws` — WebSocket stream
- `/health`, `/ready`, `/live` — probes

### 2. Carrier Connector (Go 1.26)

**Purpose**: Integration with GSMA SM-DP+ servers for eSIM provisioning.

**Flow**:
1. Receive eSIM order from API Server
2. Call SM-DP+ vendor API with IMSI, K, OPc keys (ES2+ over HTTPS)
3. Receive activation code (LPA:1$...)
4. Return to API Server

Supports multiple vendors: Thales, G+D, Idemia.

### 3. Charging Engine (Rust 1.94)

**Purpose**: Real-time credit control and billing.

**Stack**:
- Axum (async web framework)
- Redis for in-memory credit balances
- Atomic INCR/DECR operations

**Endpoints**:
- `POST /v1/credit/:ip/check` — authorization
- `POST /v1/credit/:ip/deduct` — usage deduction
- `POST /v1/credit/:ip/add` — top-up
- `GET /v1/credit/:ip/balance` — balance query

**Performance**: Sub-millisecond p99, horizontally scalable via Redis Cluster.

### 4. Packet Gateway (Rust 1.94 + eBPF)

**Purpose**: High-performance packet processing at line rate.

**Stack**:
- Aya (pure Rust eBPF)
- XDP (eXpress Data Path) hook
- GTP-U decapsulation (UDP 2152)

**Architecture**:
```
NIC → XDP Hook (kernel eBPF) → eBPF Maps → Userspace Daemon → Redis
```

**Packet flow**:
1. Packet arrives at NIC
2. XDP inspects Ethernet → IP → UDP → GTP-U headers
3. Extract inner IP (user data), look up credit in eBPF map
4. If credit > 0: `XDP_PASS`; else: `XDP_DROP`
5. Userspace daemon syncs byte counters to Redis every 1s

### 5. 5G Core Network (free5GC v4.2.1)

3GPP-compliant 5G standalone core with all network functions (AMF, SMF, UPF, PCF, UDM, UDR, AUSF, NRF).

**Integration**:
- API Server writes subscriber data to MongoDB
- UPF forwards GTP-U to Packet Gateway
- AMF authenticates using stored K/OPc keys

### 6. Web Dashboard (Next.js 15)

- App Router, TypeScript, Tailwind CSS
- TanStack Query for data fetching
- Real-time updates via WebSocket
- Error boundaries and loading states

**Pages**: eSIM management, services, monitoring, billing, config, chaos engineering, plugins, automation.

## Data Flow

### eSIM Creation

```
1. Developer → API Server: POST /v1/esims { data_plan: "1GB" }
2. API Server → IMSI Allocator: generate 15-digit IMSI
3. API Server → Key Generator: K + OPc
4. API Server → MongoDB: insert subscriptionData
5. API Server → Carrier Connector: order via ES2+
6. Carrier Connector → SM-DP+: HTTPS POST (IMSI, K, OPc)
7. SM-DP+ → Carrier Connector: activation code (LPA:1$...)
8. API Server → Developer: eSIM + activation code
9. End User scans QR → SM-DP+: profile download
10. Device → free5GC AMF: 5G AKA registration
11. free5GC → Packet Gateway: GTP-U tunnel established
```

### Real-Time Usage

```
1. User device sends data
2. free5GC UPF encapsulates in GTP-U
3. Packet Gateway (eBPF XDP):
   - Decapsulate, extract user IP
   - Check credit in eBPF map
4. If credit > 0: XDP_PASS + increment byte counter
5. Userspace daemon: read eBPF maps → Redis (1s interval)
6. Charging Engine: deduct from credit balance
7. Packet Gateway: reload credit status from Redis → eBPF map
8. If credit = 0: XDP_DROP (user traffic stops)
```

## Scaling

### Horizontal (Stateless)

- API Server: 3+ replicas behind load balancer
- Carrier Connector: 2+ replicas
- Charging Engine: 2+ replicas
- Web Dashboard: CDN + multiple origins

### Stateful

- MongoDB: replica set (1 primary, 2 secondaries)
- PostgreSQL: primary + streaming replicas
- Redis: Cluster or Sentinel
- Packet Gateway: one per physical server (eBPF kernel binding)
- free5GC: K8s deployment with dedicated UPF nodes

### Performance Targets

| Component | Throughput | Latency (p99) |
|-----------|-----------|----------------|
| API Server | 10,000 req/s | <100ms |
| Charging Engine | 50,000 checks/s | <1ms |
| Packet Gateway | 10+ Gbps/server | <1µs (kernel) |
| free5GC UPF | 5 Gbps/instance | — |
| eSIM provisioning | — | <2s end-to-end |

## Security & RBAC

### Authentication

- **JWT** tokens for web dashboard and API
- **API keys** with HMAC signatures for programmatic access
- **OAuth 2.0** for third-party integrations
- **Refresh tokens** for long-lived sessions

JWT middleware (`@/Users/nutcase/Downloads/telecom-platform-starter/apps/api-server/internal/middleware/auth.go`) validates tokens on all protected routes.

### RBAC with Casbin

Sophisticated policy-based access control using Casbin with a GORM adapter for persistent policies.

**Request flow**:
```
Request → AuthMiddleware (JWT) → CasbinRBAC (policy check) → Handler
```

**Model** (in `@/Users/nutcase/Downloads/telecom-platform-starter/apps/api-server/internal/rbac/casbin.go`):

```
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch(r.obj, p.obj) && (r.act == p.act || p.act == "*")
```

**Default policies**:

| Role | Permissions |
|------|-------------|
| `admin` | All resources, all actions |
| `operator` | Services, monitoring, deployments (GET/POST) |
| `viewer` | Read-only on all resources |

**Usage in routes**:
```go
users := protected.Group("/users")
if casbinSvc != nil {
    users.Use(middleware.RequireCasbinPermission(casbinSvc, "/v1/users", "GET"))
} else {
    users.Use(middleware.RequireRole("admin")) // fallback
}
```

**Policy management**:
```go
casbinSvc.AddPolicy("operator", "/v1/services/*", "GET")
casbinSvc.AddRoleForUser("user123", "operator")
allowed, err := casbinSvc.CheckPermission(userID, "/v1/services", "GET")
```

### Data Encryption

- TLS 1.3 for all API endpoints
- MongoDB/PostgreSQL encryption at rest
- Redis TLS connections
- Secrets via Kubernetes Secrets or HashiCorp Vault
- No hardcoded credentials

### Security Headers

Applied globally via middleware:

```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
```

Plus request size limit (10MB), request timeout (30s), and CORS configuration.

### Enhanced Error Handling

Comprehensive error handling system with standardized error responses:

**Error Codes**:
- `VALIDATION_FAILED` - Input validation errors with field-specific details
- `UNAUTHORIZED` - Authentication failures
- `FORBIDDEN` - Authorization/permission errors
- `NOT_FOUND` - Resource not found
- `ALREADY_EXISTS` - Duplicate resource attempts
- `INTERNAL_ERROR` - Server-side errors
- `RATE_LIMITED` - Rate limiting violations

**Error Response Format**:
```json
{
  "error": "Human-readable error message",
  "code": "ERROR_CODE",
  "details": "Detailed error context",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

**Validation Errors**:
```json
{
  "error": "Validation failed",
  "code": "VALIDATION_FAILED",
  "errors": [
    {
      "field": "email",
      "message": "must be a valid email address",
      "value": "invalid-email"
    }
  ]
}
```

### Testing Strategy

**Integration Tests**:
- Full API endpoint testing with in-memory SQLite
- Authentication and RBAC testing with mock users
- Service layer testing for all major services
- Error handling validation across all endpoints

**Test Coverage**:
- Authentication flows (login, token validation, refresh)
- RBAC permissions (admin, operator, viewer roles)
- Service operations (CRUD, validation, error handling)
- Error scenarios (invalid input, missing resources, permissions)

**Test Structure**:
```go
func TestAuthenticationEndpoints(t *testing.T) {
    ts := setupTest(t)
    defer ts.DB.Close()
    
    t.Run("Login with valid credentials", func(t *testing.T) {
        // Test implementation
    })
}
```

### Monitoring & Observability

**Metrics Collection**:
- Request latency and throughput
- Error rates by endpoint and error type
- Authentication success/failure rates
- Database query performance
- JWT token validation metrics

**Health Checks**:
- `/health` - Basic health status
- `/ready` - Readiness probe (dependencies check)
- `/live` - Liveness probe (service health)

**Logging**:
- Structured JSON logging
- Request correlation IDs
- Error stack traces in debug mode
- Security event logging

### Rate Limiting

Tiered limits protect against abuse (see `@/Users/nutcase/Downloads/telecom-platform-starter/apps/api-server/internal/middleware/ratelimit.go`):

| Tier | Limit/min | Burst |
|------|-----------|-------|
| Auth | 10 | 5 |
| Read | 200 | 50 |
| Admin | 50 | 10 |
| Default | 100 | 20 |

## Future Enhancements

1. **Multi-region deployment** — AWS multi-region with Route 53 global LB and data residency compliance.
2. **Advanced analytics** — predictive usage forecasting, anomaly detection.
3. **Additional carriers** — multi-vendor SM-DP+ with fallback.
4. **IoT optimization** — SGP.32 (IoT eSIM), NB-IoT/LTE-M, power-saving features.
5. **Developer tools** — expanded CLI, Terraform provider, GitHub Actions integration.

## Related Docs

- [`GETTING_STARTED.md`](./GETTING_STARTED.md) — Setup, API reference, SDKs
- [`OPERATIONS.md`](./OPERATIONS.md) — Deployment, monitoring, performance tuning
