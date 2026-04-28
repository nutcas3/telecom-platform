# Operations

Deployment, monitoring, performance tuning, and production operations for the TaaS Platform.

## Table of Contents

1. [Docker Deployment](#docker-deployment)
2. [Kubernetes Deployment](#kubernetes-deployment)
3. [Production Considerations](#production-considerations)
4. [Performance & Optimization](#performance--optimization)
5. [Monitoring & Alerting](#monitoring--alerting)
6. [CI/CD](#cicd)
7. [Backup & Disaster Recovery](#backup--disaster-recovery)

## Docker Deployment

### Build

```bash
make docker-build
# Or individually
docker build -f deployments/docker/api-server.Dockerfile -t taas-api-server .
docker build -f deployments/docker/charging-engine.Dockerfile -t taas-charging-engine .
docker build -f deployments/docker/web-dashboard.Dockerfile -t taas-web-dashboard .
```

### Run

```bash
docker-compose up -d
docker-compose logs -f
docker-compose down          # Stop
docker-compose down -v       # Stop + remove volumes
```

### Troubleshooting

```bash
# MongoDB
docker-compose ps mongodb
docker-compose logs mongodb
docker exec -it taas-mongodb mongosh

# Redis
docker exec -it taas-redis redis-cli ping
```

## Kubernetes Deployment

### Prerequisites

- Kubernetes v1.28+
- `kubectl` configured
- Persistent storage provisioner
- LoadBalancer or Ingress controller

### Deploy

```bash
kubectl create namespace taas-platform
kubectl apply -f deployments/kubernetes/
kubectl get pods -n taas-platform
```

### Access

```bash
kubectl get svc api-server -n taas-platform
kubectl port-forward -n taas-platform svc/api-server 8000:8000
```

### Scale

```bash
kubectl scale deployment api-server -n taas-platform --replicas=5
kubectl autoscale deployment api-server -n taas-platform \
  --min=3 --max=10 --cpu-percent=70
```

### Rolling Updates

```bash
kubectl set image deployment/api-server \
  api-server=ghcr.io/nutcas3/telecom-platform/api-server:v2.0.0 \
  -n taas-platform
kubectl rollout status deployment/api-server -n taas-platform
kubectl rollout undo deployment/api-server -n taas-platform   # rollback
```

### Logs & Events

```bash
kubectl logs -f deployment/api-server -n taas-platform
kubectl describe pod <pod-name> -n taas-platform
kubectl get events -n taas-platform --sort-by='.lastTimestamp'
```

## Production Considerations

### High Availability

**PostgreSQL / MongoDB replica sets**:
```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: mongodb
spec:
  replicas: 3
  serviceName: mongodb
  # See MongoDB Operator docs for full config
```

**Redis Cluster**:
```bash
helm install redis bitnami/redis-cluster \
  --namespace taas-platform \
  --set cluster.nodes=6 \
  --set cluster.replicas=1
```

### Secrets Management

```bash
kubectl create secret generic db-secret \
  --from-literal=username=admin \
  --from-literal=password='STRONG_PASSWORD' \
  -n taas-platform

kubectl create secret generic api-keys \
  --from-literal=jwt-secret="$(openssl rand -base64 32)" \
  -n taas-platform
```

For production, prefer HashiCorp Vault or AWS Secrets Manager over raw K8s secrets.

### Network Policies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: api-server-policy
  namespace: taas-platform
spec:
  podSelector:
    matchLabels:
      app: api-server
  policyTypes: [Ingress, Egress]
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: web-dashboard
      ports:
        - protocol: TCP
          port: 8000
  egress:
    - to:
        - podSelector:
            matchLabels:
              app: mongodb
      ports:
        - protocol: TCP
          port: 27017
```

### Resource Limits

| Service | CPU Request | CPU Limit | Mem Request | Mem Limit |
|---------|------------:|----------:|------------:|----------:|
| API Server | 100m | 500m | 128Mi | 512Mi |
| Charging Engine | 100m | 1000m | 128Mi | 512Mi |
| Packet Gateway | 250m | 2000m | 256Mi | 1Gi |
| MongoDB | 250m | 1000m | 512Mi | 2Gi |
| PostgreSQL | 250m | 1000m | 512Mi | 2Gi |
| Redis | 100m | 500m | 256Mi | 2Gi |

## Performance & Optimization

### Rate Limiting

In-memory sliding window with per-IP tracking. Configured via `middleware.RateLimitByEndpoint()` in `@/Users/nutcase/Downloads/telecom-platform-starter/apps/api-server/internal/middleware/ratelimit.go`.

| Tier | Limit/min | Burst | Applies To |
|------|-----------|-------|------------|
| Auth | 10 | 5 | `/v1/auth/login`, `/v1/auth/register` |
| Read | 200 | 50 | `/v1/services`, `/v1/monitoring/metrics` |
| Admin | 50 | 10 | `/v1/users`, `/v1/config` |
| Default | 100 | 20 | Everything else |

### Caching

Layered in-memory LRU cache with TTL (`@/Users/nutcase/Downloads/telecom-platform-starter/apps/api-server/internal/cache/cache.go`):

| Namespace | TTL | Max Items | Use |
|-----------|-----|-----------|-----|
| `default` | 5 min | 1000 | General purpose |
| `users` | 10 min | 500 | User data |
| `services` | 2 min | 200 | Service info |
| `metrics` | 1 min | 1000 | Monitoring data |

```go
cache := cache.GetDefaultCache()
cache.Set("key", value, 5*time.Minute)
value, found := cache.Get("key")
```

The cached database wrapper (`@/Users/nutcase/Downloads/telecom-platform-starter/apps/api-server/internal/database/cached.go`) provides `CachedFind`, `CachedFirst`, `InvalidateCache`.

### Connection Pooling

Tuned in `@/Users/nutcase/Downloads/telecom-platform-starter/apps/api-server/internal/database/database.go:50-54`:

```go
sqlDB.SetMaxIdleConns(25)                   // minimum idle
sqlDB.SetMaxOpenConns(100)                  // maximum open
sqlDB.SetConnMaxLifetime(time.Hour)
sqlDB.SetConnMaxIdleTime(30 * time.Minute)
```

`db.PoolStats()` exposes live pool metrics for dashboards and alerts.

### Performance Middleware

Tracks response times, error rates, and slow requests (>100ms threshold). Adds:

- `X-Response-Time` header
- `X-Request-ID` correlation ID
- Security headers (CSP, HSTS, X-Frame-Options)
- Request size limit (10MB) and timeout (30s)

**Metrics endpoint**: `GET /metrics/performance`

### Performance Targets

| Metric | Target |
|--------|--------|
| API response time (p50) | <50ms |
| API response time (p99) | <100ms |
| Uptime | 99.9% |
| Error rate | <1% |
| Test coverage | >80% |

### Load Testing

```bash
brew install k6
k6 run scripts/loadtest.js
```

Sample script:
```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '2m', target: 100 },
    { duration: '5m', target: 100 },
    { duration: '2m', target: 200 },
    { duration: '5m', target: 200 },
    { duration: '2m', target: 0 },
  ],
};

export default function () {
  const res = http.get('http://api-server:8000/health');
  check(res, { 'status is 200': (r) => r.status === 200 });
  sleep(1);
}
```

## Monitoring & Alerting

### Health Checks

Three probe endpoints (`@/Users/nutcase/Downloads/telecom-platform-starter/apps/api-server/internal/monitoring/health.go`):

| Endpoint | Purpose |
|----------|---------|
| `/health` | Aggregated component health |
| `/ready` | Kubernetes readiness probe |
| `/live` | Kubernetes liveness probe |

Monitored components: **database**, **Prometheus**, **Kubernetes**, **system** (memory, goroutines, GC).

### Alert Rules

Default rules in `@/Users/nutcase/Downloads/telecom-platform-starter/apps/api-server/internal/monitoring/alerting.go`:

| Rule | Threshold | Severity |
|------|-----------|----------|
| System Health | Degraded/Unhealthy | Warning |
| High Memory | >80% | Warning |
| High Goroutines | >1000 | Warning |
| DB Connection Utilization | >80% | Warning |

### Alert Notifiers

- **WebSocket** — real-time dashboard
- **Log** — structured logs
- **Composite** — multiple channels simultaneously

### Prometheus Metrics

Key exported metrics:
- `api_requests_total` — total request count
- `esim_creations_total` — eSIMs created
- `data_usage_bytes_total` — cumulative data usage
- `charging_checks_duration_seconds` — credit check latency histogram
- `packet_gateway_packets_total` — packets processed

### Prometheus + Grafana Install

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack \
  -n monitoring --create-namespace
```

### WebSocket Real-Time Events

`GET /ws` streams:
- Service status updates
- System metrics
- Alert notifications

## CI/CD

### GitHub Actions

Two workflows in `.github/workflows/`:

**`ci.yml`** — runs on every push/PR:
- Go tests with PostgreSQL service container
- `go vet` and `golangci-lint`
- Web dashboard lint + typecheck + build
- Docker image builds (on main)
- Trivy security scan

**`deploy.yml`** — runs on main pushes and version tags:
- Build + push images to GHCR
- Deploy to staging (main branch)
- Deploy to production (version tags)

Manual trigger via `workflow_dispatch` with environment input.

### Local Testing

```bash
# Go tests with coverage
cd apps/api-server
go test ./... -v -race -coverprofile=coverage.out
go tool cover -html=coverage.out

# Benchmarks
go test -bench=. ./internal/middleware/
go test -bench=. ./internal/monitoring/

# Linting
golangci-lint run --timeout 5m
go vet ./...
go fix ./...
```

## Backup & Disaster Recovery

### MongoDB Backup (CronJob)

```bash
kubectl create cronjob mongodb-backup \
  --image=mongo:7.0 \
  --schedule="0 2 * * *" \
  -- mongodump --uri=$MONGODB_URI --out=/backup
```

### Redis Persistence

Enable AOF for durability:
```bash
kubectl edit configmap redis-config -n taas-platform
# Set: appendonly yes
```

### Recovery Objectives

- **RTO** (Recovery Time Objective): <1 hour
- **RPO** (Recovery Point Objective): <15 minutes

### Recovery Procedure

1. Spin up new Kubernetes cluster
2. Restore database from latest backup (cross-region replicated)
3. Restore Redis from RDB snapshot or rebuild from DB
4. Deploy all services from Git (`kubectl apply -f deployments/kubernetes/`)
5. Update DNS to new cluster
6. Run `make verify` to confirm functionality

### Backup Locations

- **MongoDB/PostgreSQL**: S3 bucket with cross-region replication
- **Redis**: S3 RDB snapshots
- **Configuration**: Git repository
- **Secrets**: HashiCorp Vault / AWS Secrets Manager

## Cost Optimization

- Use spot/preemptible instances for non-critical services
- Enable autoscaling to match load
- Reserved instances for predictable workloads
- Store old backups in S3 Glacier

**Sample monthly costs (AWS, base infrastructure)**:
- 3× t3.medium (API Server): $90
- 2× t3.medium (Charging Engine): $60
- 1× t3.large (MongoDB): $75
- 1× t3.small (Redis): $19
- Load Balancer: $25
- **Total**: ~$270/month + data transfer

## Related Docs

- [`GETTING_STARTED.md`](./GETTING_STARTED.md) — Setup and API reference
- [`ARCHITECTURE.md`](./ARCHITECTURE.md) — System design, RBAC, security
