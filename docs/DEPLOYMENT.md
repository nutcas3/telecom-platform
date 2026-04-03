# Deployment Guide

This guide covers deploying the TaaS Platform in different environments.

## Table of Contents

1. [Local Development](#local-development)
2. [Docker Deployment](#docker-deployment)
3. [Kubernetes Deployment](#kubernetes-deployment)
4. [Production Considerations](#production-considerations)

---

## Local Development

### Prerequisites

- Go 1.26+
- Rust 1.94+
- Node.js 22+
- MongoDB 7.0+
- Redis 7+
- pnpm

### Quick Start

```bash
# 1. Clone repository
git clone https://github.com/nutcas3/telecom-platform.git
cd telecom-platform

# 2. Run setup script
chmod +x scripts/dev-setup.sh
./scripts/dev-setup.sh

# 3. Set up databases
make db-setup

# 4. Build all components
make all

# 5. Start services (in separate terminals)
# Terminal 1: API Server
./dist/api-server

# Terminal 2: Carrier Connector
./dist/carrier-connector

# Terminal 3: Charging Engine
./target/release/charging-engine

# Terminal 4: Web Dashboard
cd apps/web-dashboard && pnpm dev
```

### Environment Variables

Create `.env` files in each service directory:

**apps/api-server/.env:**
```env
MONGODB_URI=mongodb://localhost:27017/free5gc
REDIS_URI=redis://localhost:6379
API_PORT=8000
GIN_MODE=debug
```

**apps/charging-engine/.env:**
```env
REDIS_URI=redis://127.0.0.1/
SERVER_PORT=8080
RUST_LOG=info
```

**apps/web-dashboard/.env.local:**
```env
NEXT_PUBLIC_API_URL=http://localhost:8000
```

---

## Docker Deployment

### Build Images

```bash
# Build all Docker images
make docker-build

# Or build individually
docker build -f deployments/docker/api-server.Dockerfile -t taas-api-server .
docker build -f deployments/docker/charging-engine.Dockerfile -t taas-charging-engine .
docker build -f deployments/docker/web-dashboard.Dockerfile -t taas-web-dashboard .
```

### Run with Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

### Service URLs

After starting with docker-compose:

- **API Server**: http://localhost:8000
- **Charging Engine**: http://localhost:8080
- **Web Dashboard**: http://localhost:3000
- **MongoDB**: mongodb://localhost:27017
- **Redis**: redis://localhost:6379

### Troubleshooting

**MongoDB connection issues:**
```bash
# Check if MongoDB is running
docker-compose ps mongodb

# View MongoDB logs
docker-compose logs mongodb

# Connect to MongoDB shell
docker exec -it taas-mongodb mongosh
```

**Redis connection issues:**
```bash
# Test Redis connection
docker exec -it taas-redis redis-cli ping
```

---

## Kubernetes Deployment

### Prerequisites

- Kubernetes cluster (v1.28+)
- kubectl configured
- Persistent storage provisioner
- LoadBalancer service support (or Ingress controller)

### Deploy to Kubernetes

```bash
# 1. Create namespace
kubectl create namespace taas-platform

# 2. Deploy all manifests
kubectl apply -f deployments/kubernetes/

# 3. Verify deployments
kubectl get pods -n taas-platform

# Expected output:
# NAME                              READY   STATUS    RESTARTS   AGE
# mongodb-xxx                       1/1     Running   0          2m
# redis-xxx                         1/1     Running   0          2m
# api-server-xxx                    1/1     Running   0          1m
# charging-engine-xxx               1/1     Running   0          1m
```

### Access Services

```bash
# Get API Server external IP
kubectl get svc api-server -n taas-platform

# Port forward for local access
kubectl port-forward -n taas-platform svc/api-server 8000:8000
kubectl port-forward -n taas-platform svc/charging-engine 8080:8080
```

### Scaling

```bash
# Scale API Server
kubectl scale deployment api-server -n taas-platform --replicas=5

# Scale Charging Engine
kubectl scale deployment charging-engine -n taas-platform --replicas=3

# Enable autoscaling
kubectl autoscale deployment api-server -n taas-platform \
  --min=3 --max=10 --cpu-percent=70
```

### Monitoring

```bash
# View logs
kubectl logs -f deployment/api-server -n taas-platform

# Describe pod for issues
kubectl describe pod <pod-name> -n taas-platform

# Get events
kubectl get events -n taas-platform --sort-by='.lastTimestamp'
```

### Update Deployment

```bash
# Update image
kubectl set image deployment/api-server \
  api-server=taas-api-server:v2.0.0 \
  -n taas-platform

# Check rollout status
kubectl rollout status deployment/api-server -n taas-platform

# Rollback if needed
kubectl rollout undo deployment/api-server -n taas-platform
```

---

## Production Considerations

### High Availability

**MongoDB Replica Set:**
```yaml
# Use MongoDB Operator or StatefulSet
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: mongodb
spec:
  replicas: 3
  serviceName: mongodb
  # ... (see MongoDB Operator docs)
```

**Redis Cluster:**
```bash
# Deploy Redis in cluster mode
helm install redis bitnami/redis-cluster \
  --namespace taas-platform \
  --set cluster.nodes=6 \
  --set cluster.replicas=1
```

### Security

**Secrets Management:**
```bash
# Create Kubernetes secrets
kubectl create secret generic mongodb-secret \
  --from-literal=username=admin \
  --from-literal=password='STRONG_PASSWORD' \
  -n taas-platform

kubectl create secret generic api-keys \
  --from-literal=jwt-secret='JWT_SECRET' \
  -n taas-platform
```

**Network Policies:**
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
  policyTypes:
  - Ingress
  - Egress
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

### Monitoring Stack

**Prometheus & Grafana:**
```bash
# Install using Helm
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack \
  -n monitoring --create-namespace
```

**Application Metrics:**
- Add `/metrics` endpoint to each service
- Use Prometheus client libraries
- Create Grafana dashboards

### Backup Strategy

**MongoDB Backups:**
```bash
# Daily backups with mongodump
kubectl create cronjob mongodb-backup \
  --image=mongo:7.0 \
  --schedule="0 2 * * *" \
  -- mongodump --uri=$MONGODB_URI --out=/backup
```

**Redis Persistence:**
```bash
# Enable AOF in Redis config
kubectl edit configmap redis-config -n taas-platform
# Set: appendonly yes
```

### Resource Limits

**Recommended limits per service:**

| Service | CPU Request | CPU Limit | Memory Request | Memory Limit |
|---------|-------------|-----------|----------------|--------------|
| API Server | 100m | 500m | 128Mi | 512Mi |
| Charging Engine | 100m | 1000m | 128Mi | 512Mi |
| Packet Gateway | 250m | 2000m | 256Mi | 1Gi |
| MongoDB | 250m | 1000m | 512Mi | 2Gi |
| Redis | 100m | 500m | 256Mi | 2Gi |

### Load Testing

```bash
# Install k6
brew install k6  # macOS
# or: sudo apt-get install k6

# Run load test
k6 run scripts/loadtest.js
```

**Sample load test script:**
```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '2m', target: 100 },  // Ramp up
    { duration: '5m', target: 100 },  // Stay at 100 users
    { duration: '2m', target: 200 },  // Ramp to 200
    { duration: '5m', target: 200 },  // Stay at 200
    { duration: '2m', target: 0 },    // Ramp down
  ],
};

export default function () {
  const res = http.get('http://api-server:8000/health');
  check(res, { 'status is 200': (r) => r.status === 200 });
  sleep(1);
}
```

### Disaster Recovery

**Backup Locations:**
- MongoDB backups: S3 bucket (cross-region replication)
- Redis RDB snapshots: S3 bucket
- Configuration: Git repository
- Secrets: HashiCorp Vault or AWS Secrets Manager

**Recovery Time Objective (RTO):** <1 hour
**Recovery Point Objective (RPO):** <15 minutes

**Recovery Procedure:**
1. Spin up new Kubernetes cluster
2. Restore MongoDB from latest backup
3. Restore Redis from snapshot
4. Deploy all services from Git
5. Update DNS to point to new cluster
6. Verify functionality

### CI/CD Pipeline

**GitHub Actions example:**
```yaml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Build Docker images
        run: make docker-build
      
      - name: Push to registry
        run: |
          docker tag taas-api-server registry.example.com/taas-api-server:${{ github.sha }}
          docker push registry.example.com/taas-api-server:${{ github.sha }}
      
      - name: Deploy to Kubernetes
        run: |
          kubectl set image deployment/api-server \
            api-server=registry.example.com/taas-api-server:${{ github.sha }} \
            -n taas-platform
```

### Cost Optimization

**Cloud Provider Recommendations:**
- Use spot/preemptible instances for non-critical services
- Enable autoscaling to match load
- Use reserved instances for predictable workloads
- Store backups in cheaper storage classes (S3 Glacier)

**Estimated Monthly Costs (AWS):**
- 3x t3.medium (API Server): $90
- 2x t3.medium (Charging Engine): $60
- 1x t3.large (MongoDB): $75
- 1x t3.small (Redis): $19
- Load Balancer: $25
- Data transfer: Variable
- **Total**: ~$270/month (base infrastructure)

---

## Support

For deployment issues:
- Check logs: `kubectl logs -f <pod-name>`
- Review events: `kubectl get events`
- Join community: [Discussions](https://github.com/taas-platform/discussions)
- Email: devops@taas-platform.com
