# TaaS Platform Architecture

## System Overview

The TaaS Platform is a multi-layer telecom infrastructure that provides cellular connectivity as a service. It implements 3GPP standards while exposing a modern, developer-friendly API.

## Architecture Layers

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

**Purpose**: Developer-facing RESTful API and business logic

**Technology**:
- Language: Go 1.26 (Green Tea GC, new(expr) syntax)
- Framework: Gin
- Database: MongoDB (subscriber data)
- Cache: Redis (session management)

**Responsibilities**:
- eSIM creation and management
- Authentication and authorization
- Rate limiting and quotas
- Webhook delivery
- API documentation

**Endpoints**:
- `POST /v1/esims` - Create eSIM
- `GET /v1/esims/:id` - Get eSIM details
- `GET /v1/esims/:id/usage` - Get usage stats
- `DELETE /v1/esims/:id` - Terminate eSIM

### 2. Carrier Connector (Go 1.26)

**Purpose**: Integration with GSMA SM-DP+ servers for eSIM provisioning

**Technology**:
- Language: Go 1.26
- Protocol: HTTPS (ES2+ API)
- Vendor: Thales, G+D, or Idemia

**Responsibilities**:
- Profile ordering via ES2+ protocol
- Activation code generation (LPA strings)
- Profile status tracking
- Vendor-specific adaptations

**Flow**:
1. Receive eSIM order from API Server
2. Call SM-DP+ vendor API with IMSI, K, OPc keys
3. Receive activation code (LPA:1$...)
4. Return to API Server for user display

### 3. Charging Engine (Rust 1.94)

**Purpose**: Real-time credit control and billing

**Technology**:
- Language: Rust 1.94 (array_windows, async Tokio)
- Framework: Axum
- Database: Redis (in-memory credit balances)

**Responsibilities**:
- Credit balance management
- Real-time authorization checks
- Usage deduction (atomic operations)
- Credit top-up processing

**API**:
- `POST /v1/credit/:ip/check` - Check if user has credit
- `POST /v1/credit/:ip/deduct` - Deduct bytes used
- `POST /v1/credit/:ip/add` - Add credit
- `GET /v1/credit/:ip/balance` - Get balance

**Performance**:
- Sub-millisecond response times
- Atomic Redis operations (INCR, DECR)
- Horizontal scaling via Redis Cluster

### 4. Packet Gateway (Rust 1.94 + eBPF)

**Purpose**: High-performance packet processing and traffic control

**Technology**:
- Language: Rust 1.94
- Framework: Aya (pure Rust eBPF)
- Attach Point: XDP (eXpress Data Path)
- Protocol: GTP-U decapsulation (UDP port 2152)

**Responsibilities**:
- Parse GTP-U tunneled packets
- Extract user IP addresses
- Count bytes per subscriber
- Block traffic for users without credit
- Sync stats to Redis

**Architecture**:
```
┌─────────────────────────────────────────────┐
│           Network Interface (eth0)          │
└─────────────────────────────────────────────┘
                    │
          ┌─────────▼─────────┐
          │   XDP Hook        │ ← eBPF program runs here
          │   (Kernel)        │
          └─────────┬─────────┘
                    │
          ┌─────────▼─────────┐
          │   eBPF Maps       │
          │  - Packet Stats   │
          │  - Credit Status  │
          └─────────┬─────────┘
                    │
          ┌─────────▼─────────┐
          │  Userspace Daemon │ ← Rust application
          │  (Reads maps,     │
          │   syncs to Redis) │
          └───────────────────┘
```

**Packet Processing**:
1. Packet arrives at NIC
2. XDP hook inspects Ethernet → IP → UDP → GTP-U headers
3. Extract inner IP packet (actual user data)
4. Look up credit status in eBPF map
5. If credit > 0: `XDP_PASS`, else: `XDP_DROP`
6. Update packet statistics in eBPF map
7. Userspace daemon reads stats, syncs to Redis

### 5. 5G Core Network (free5GC)

**Purpose**: 3GPP-compliant 5G standalone core

**Technology**:
- Project: free5GC v4.2.1
- Language: Go
- Database: MongoDB
- Deployment: Kubernetes or bare metal

**Network Functions**:
- **AMF** (Access and Mobility Management)
- **SMF** (Session Management)
- **UPF** (User Plane Function)
- **PCF** (Policy Control)
- **UDM** (Unified Data Management)
- **UDR** (Unified Data Repository)
- **AUSF** (Authentication Server)
- **NRF** (NF Repository)

**Integration Points**:
- API Server writes subscriber data to MongoDB
- UPF forwards GTP-U traffic to Packet Gateway
- AMF handles UE authentication using stored keys (K, OPc)

### 6. Web Dashboard (Next.js 15)

**Purpose**: Developer portal and management interface

**Technology**:
- Framework: Next.js 15 (App Router)
- Language: TypeScript
- Styling: Tailwind CSS
- Data Fetching: TanStack Query

**Features**:
- eSIM creation and management
- Real-time usage charts (Recharts)
- API key management
- Webhook configuration
- Billing and invoices

## Data Flow

### eSIM Creation Flow

```
1. Developer → API Server
   POST /v1/esims { data_plan: "1GB" }

2. API Server → IMSI Allocator
   Generate unique 15-digit IMSI

3. API Server → Key Generator
   Generate K (subscriber key)
   Generate OPc (operator key)

4. API Server → MongoDB (free5GC)
   Insert subscriber data:
   - subscriptionData.authenticationData (K, OPc)
   - subscriptionData.provisionedData (IMSI, AMBR)

5. API Server → Carrier Connector
   Order eSIM profile via ES2+

6. Carrier Connector → SM-DP+ Vendor
   HTTPS POST with IMSI, K, OPc

7. SM-DP+ → Carrier Connector
   Return activation code (LPA string)

8. API Server → Developer
   Return eSIM details + activation code

9. Developer → End User
   Display QR code with LPA string

10. End User's Device → SM-DP+
    Scan QR, download eSIM profile

11. Device → free5GC AMF
    Registration procedure (5G AKA)

12. free5GC → Packet Gateway
    Establish GTP-U tunnel for data
```

### Real-Time Usage Flow

```
1. User Device sends data
   ↓
2. free5GC UPF encapsulates in GTP-U
   ↓
3. Packet Gateway (eBPF XDP)
   - Decapsulate GTP-U
   - Extract user IP
   - Check credit status in eBPF map
   ↓
4. If credit > 0:
   - Pass packet (XDP_PASS)
   - Increment byte counter in eBPF map
   ↓
5. Userspace daemon (Rust)
   - Reads eBPF maps every 1 second
   - Writes stats to Redis
   ↓
6. Charging Engine
   - Reads usage from Redis
   - Deducts from credit balance
   - Updates credit status
   ↓
7. Packet Gateway
   - Reads new credit status from Redis
   - Updates eBPF map
   ↓
8. If credit reaches 0:
   - eBPF drops packets (XDP_DROP)
   - User's internet stops
```

## Scaling Architecture

### Horizontal Scaling

**Stateless Components** (can be replicated):
- API Server: 3+ replicas behind load balancer
- Carrier Connector: 2+ replicas
- Charging Engine: 2+ replicas
- Web Dashboard: CDN + multiple origins

**Stateful Components** (require special handling):
- MongoDB: Replica set (1 primary, 2 secondaries)
- Redis: Cluster mode or Sentinel
- Packet Gateway: One instance per physical server (eBPF)
- free5GC: Kubernetes deployment with dedicated UPF nodes

### Performance Metrics

**Throughput**:
- API Server: 10,000 requests/second
- Charging Engine: 50,000 credit checks/second
- Packet Gateway: 10 Gbps+ per server (eBPF)
- free5GC UPF: 5 Gbps per instance

**Latency**:
- API Server: <50ms p99
- Charging Engine: <1ms p99 (Redis)
- Packet Gateway: <1µs (eBPF in kernel)
- eSIM Provisioning: <2 seconds end-to-end

## Security

### Authentication & Authorization
- API keys with HMAC signatures
- JWT tokens for web dashboard
- OAuth 2.0 for third-party integrations
- Rate limiting per API key

### Data Encryption
- TLS 1.3 for all API endpoints
- MongoDB encryption at rest
- Redis TLS connections
- eBPF packet inspection (no decryption)

### Secrets Management
- Kubernetes Secrets for production
- HashiCorp Vault for sensitive keys
- Environment variables for local dev
- No hardcoded credentials

## Deployment

### Development
```bash
make all           # Build all components
make db-setup      # Set up MongoDB + Redis
make dev           # Start all services
```

### Docker Compose
```bash
docker-compose up -d
```

### Kubernetes
```bash
kubectl apply -f deployments/kubernetes/
```

### Production Checklist
- [ ] MongoDB replica set configured
- [ ] Redis Cluster or Sentinel enabled
- [ ] TLS certificates installed
- [ ] Monitoring (Prometheus, Grafana)
- [ ] Logging (ELK Stack or Loki)
- [ ] Backup strategy implemented
- [ ] Disaster recovery plan documented
- [ ] Load testing completed (10x expected traffic)

## Monitoring

### Metrics (Prometheus)
- `api_requests_total` - Total API requests
- `esim_creations_total` - eSIMs created
- `data_usage_bytes_total` - Total data usage
- `charging_checks_duration_seconds` - Credit check latency
- `packet_gateway_packets_total` - Packets processed

### Alerts
- High error rate (>1%)
- Low credit balances (<10% of users)
- MongoDB replica lag (>10 seconds)
- Redis memory usage (>90%)
- Packet gateway packet loss (>0.1%)

## Future Enhancements

1. **Multi-Region Deployment**
   - Deploy in multiple AWS regions
   - Global load balancing with Route 53
   - Data residency compliance

2. **Advanced Analytics**
   - Real-time dashboards with WebSockets
   - Predictive usage forecasting
   - Anomaly detection (unusual data patterns)

3. **Additional Carriers**
   - Support for multiple SM-DP+ vendors
   - Dynamic carrier selection based on region
   - Fallback mechanisms

4. **IoT Optimization**
   - SGP.32 support (IoT eSIM standard)
   - LPWAN protocol support (NB-IoT, LTE-M)
   - Power-saving features

5. **Developer Tools**
   - CLI tool for API interaction
   - Terraform provider
   - GitHub Actions integration
