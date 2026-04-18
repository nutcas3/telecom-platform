# Telecom Platform Monitoring Stack

This directory contains the complete monitoring and observability setup for the Telecom Platform, including Prometheus metrics collection, Grafana dashboards, and InfluxDB time-series storage.

## Architecture

```
Telecom Platform Services
    |
    | Metrics (Prometheus format)
    v
+-----------------+
|   Prometheus    | <- Scrapes metrics from all services
+-----------------+
    |
    | Query API
    v
+-----------------+
|     Grafana     | <- Visualization and alerting
+-----------------+
    |
    | Long-term storage
    v
+-----------------+
|    InfluxDB     | <- Time-series data storage
+-----------------+
```

## Components

### 1. Prometheus Metrics Exporter

**Location**: `apps/api-server/internal/metrics/prometheus.go`

**Features**:
- HTTP request metrics (rate, duration, status codes)
- Subscriber metrics (total, active, suspended)
- Usage metrics (data, voice, SMS)
- Charging metrics (credit balance, transactions)
- System metrics (sessions, uptime, database connections)
- Chaos engineering metrics
- eBPF packet gateway metrics

**Key Metrics**:
- `telecom_http_requests_total` - Total HTTP requests
- `telecom_subscribers_total` - Total subscriber count
- `telecom_data_usage_bytes_total` - Data usage in bytes
- `telecom_credit_balance` - Subscriber credit balances
- `telecom_active_sessions` - Active session count
- `telecom_packets_processed_total` - eBPF packet processing

### 2. Advanced Time-Series Storage

**Location**: `apps/api-server/internal/metrics/timeseries.go`

**Features**:
- InfluxDB integration for long-term storage
- High-resolution usage metrics
- Historical data analysis
- Complex queries and aggregations
- Top user analytics

**Data Types**:
- Usage metrics (data up/down, voice, SMS)
- System metrics (CPU, memory, network)
- Charging metrics (transactions, balances)
- Packet processing metrics

### 3. Grafana Dashboards

**Location**: `monitoring/grafana/dashboards/`

#### Dashboard: Telecom Platform Overview
- Total subscribers and active sessions
- Data usage rates
- HTTP request rates
- Subscriber status breakdown
- Payment transaction rates

#### Dashboard: Telecom Subscriber Metrics
- Top 10 data users
- Subscriber credit balances
- Low balance alerts
- Data usage distribution
- Voice & SMS usage rates

### 4. Docker Compose Setup

**Location**: `monitoring/docker-compose.yml`

**Services**:
- **Prometheus** (port 9090) - Metrics collection and storage
- **Grafana** (port 3001) - Visualization and dashboards
- **InfluxDB** (port 8086) - Time-series data storage
- **Redis** (port 6379) - Caching and session storage
- **Node Exporter** (port 9100) - System metrics
- **Alertmanager** (port 9093) - Alert management

## Quick Start

### 1. Start Monitoring Stack

```bash
cd monitoring
docker-compose up -d
```

### 2. Access Services

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3001 (admin/admin)
- **InfluxDB**: http://localhost:8086

### 3. Start API Server with Metrics

```bash
cd apps/api-server
go run cmd/server.go
```

The API server will:
- Expose metrics on `http://localhost:9090/metrics`
- Store time-series data in InfluxDB (if configured)
- Update Prometheus metrics every 30 seconds

## Configuration

### Environment Variables

```bash
# API Server
METRICS_PORT=9090

# InfluxDB (optional)
INFLUXDB_URL=http://localhost:8086
INFLUXDB_TOKEN=your-token
INFLUXDB_ORG=telecom
INFLUXDB_BUCKET=telecom
```

### Prometheus Configuration

**Location**: `monitoring/prometheus/prometheus.yml`

Key scrape targets:
- `telecom-api-server` - Main API metrics
- `telecom-charging-engine` - Rust charging engine metrics
- `telecom-packet-gateway` - eBPF packet gateway metrics

### Grafana Data Sources

**Location**: `monitoring/grafana/provisioning/datasources/datasources.yml`

- **Prometheus** - Primary metrics source
- **InfluxDB** - Time-series analytics

## Metrics Collection

### Automatic Collection

The metrics collector automatically gathers:
1. **HTTP Metrics** - Request rate, duration, status codes
2. **Business Metrics** - Subscriber counts, usage data
3. **System Metrics** - Database connections, Redis operations
4. **Chaos Metrics** - Experiment results and status

### Custom Metrics

Add custom metrics in your service code:

```go
// Record custom usage metrics
metricsCollector.RecordDataUsage("subscriber123", "up", 1024)
metricsCollector.UpdateCreditBalance("subscriber123", 25.50)

// Record chaos experiment
metricsCollector.RecordChaosExperiment("latency", "completed")
```

### Time-Series Storage

Store detailed metrics in InfluxDB:

```go
usageMetrics := metrics.UsageMetrics{
    SubscriberID: "subscriber123",
    IMSI:         "208930000000001",
    DataUp:       1024 * 1024,  // 1MB
    DataDown:     2 * 1024 * 1024, // 2MB
    VoiceSeconds: 300,
    SMSCount:     5,
    Timestamp:    time.Now(),
}

err := timeSeriesStorage.StoreUsageMetrics(ctx, usageMetrics)
```

## Alerting

### Prometheus Alert Rules

Create alert rules in `monitoring/prometheus/alert_rules.yml`:

```yaml
groups:
  - name: telecom_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(telecom_http_requests_total{status_code=~"5.."}[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
      
      - alert: LowCreditBalance
        expr: telecom_credit_balance < 5
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Subscriber has low credit balance"
```

### Grafana Alerting

Configure alerts in Grafana dashboards:
- Low balance alerts
- High error rate alerts
- System resource alerts
- Usage threshold alerts

## Advanced Features

### 1. Distributed Tracing

Integration with Jaeger for request tracing:
- Trace HTTP requests across services
- Track GraphQL query performance
- Monitor database query performance

### 2. Log Aggregation

ELK stack integration for log aggregation:
- Centralized log collection
- Log parsing and indexing
- Log-based alerting

### 3. Business Intelligence

Advanced analytics with InfluxDB:
- Customer usage patterns
- Revenue analytics
- Network performance trends
- Capacity planning

## Troubleshooting

### Common Issues

1. **Metrics not appearing in Prometheus**
   - Check if metrics endpoint is accessible: `curl http://localhost:9090/metrics`
   - Verify Prometheus configuration targets
   - Check network connectivity

2. **Grafana dashboards not loading**
   - Verify Grafana can reach Prometheus: `curl http://prometheus:9090/api/v1/query?query=up`
   - Check data source configuration
   - Verify dashboard JSON syntax

3. **InfluxDB connection issues**
   - Verify InfluxDB is running: `curl http://localhost:8086/ping`
   - Check authentication token
   - Verify bucket and organization names

### Debug Commands

```bash
# Check Prometheus targets
curl http://localhost:9090/api/v1/targets

# Query specific metric
curl "http://localhost:9090/api/v1/query?query=telecom_subscribers_total"

# Check InfluxDB health
curl http://localhost:8086/health

# View Grafana data sources
curl -u admin:admin http://localhost:3001/api/datasources
```

## Performance Considerations

### Metrics Collection Overhead
- Prometheus scraping adds minimal overhead (~1-2ms per request)
- InfluxDB writes are batched for efficiency
- Time-series storage is optional and can be disabled

### Storage Requirements
- Prometheus: ~1-2GB per month for medium load
- InfluxDB: Depends on retention period and data volume
- Grafana: Minimal storage for dashboards

### Scaling
- Prometheus can handle ~10k metrics per second
- InfluxDB scales horizontally with clustering
- Grafana supports multiple data sources

## Security

### Authentication
- Grafana: Configure LDAP/OAuth integration
- InfluxDB: Use authentication tokens
- Prometheus: Use basic auth or TLS

### Network Security
- Run monitoring stack in isolated network
- Use TLS for all communications
- Implement firewall rules for metric endpoints

## Next Steps

1. **Custom Dashboards** - Create domain-specific dashboards
2. **Advanced Alerting** - Implement predictive alerting
3. **ML Integration** - Add anomaly detection
4. **Automation** - Auto-scaling based on metrics
5. **Compliance** - Audit logging and retention policies

## Support

For monitoring-related issues:
1. Check service logs: `docker-compose logs [service]`
2. Verify configuration files
3. Test connectivity between services
4. Check resource usage and limits
