# TaaS Platform SDK Documentation

Multi-language SDK documentation for integrating with the Telecom-as-a-Service Platform.

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Authentication](#authentication)
- [Go SDK](#go-sdk)
- [Python SDK](#python-sdk)
- [TypeScript SDK](#typescript-sdk)
- [Analytics API](#analytics-api)
- [Security API](#security-api)
- [Currency & Billing API](#currency--billing-api)

---

## Overview

The TaaS Platform provides SDKs for multiple programming languages to simplify integration with platform services:

| Language | Package | Status |
|----------|---------|--------|
| Go | `github.com/nutcas3/telecom-platform/sdk/go` | ✅ Stable |
| Python | `telecom-sdk` | ✅ Stable |
| TypeScript | `@taas/sdk` | ✅ Stable |
| Kotlin | `com.taas:sdk` | 🚧 Beta |
| Ruby | `taas-sdk` | 🚧 Beta |
| Swift | `TaaSSDK` | 🚧 Beta |
| Rust | `taas-sdk` | 🚧 Beta |
| Elixir | `taas_sdk` | 🚧 Beta |

---

## Installation

### Go

```bash
go get github.com/nutcas3/telecom-platform/sdk/go
```

### Python

```bash
pip install telecom-sdk
```

### TypeScript/JavaScript

```bash
npm install @taas/sdk
# or
pnpm add @taas/sdk
```

---

## Authentication

All SDKs use JWT-based authentication. Obtain an API key from the dashboard or use username/password authentication.

### API Key Authentication

```go
// Go
client := taas.NewClient(taas.WithAPIKey("your-api-key"))
```

```python
# Python
from telecom_sdk import TelecomClient
client = TelecomClient(api_key="your-api-key")
```

```typescript
// TypeScript
import { TelecomClient } from '@taas/sdk';
const client = new TelecomClient({ apiKey: 'your-api-key' });
```

---

## Go SDK

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    taas "github.com/nutcas3/telecom-platform/sdk/go"
)

func main() {
    // Initialize client
    client := taas.NewClient(
        taas.WithBaseURL("https://api.telecom.com"),
        taas.WithAPIKey("your-api-key"),
    )

    // Create a subscriber
    subscriber, err := client.Subscribers.Create(context.Background(), &taas.CreateSubscriberRequest{
        MSISDN:    "+1234567890",
        FirstName: "John",
        LastName:  "Doe",
        Email:     "john@example.com",
        PlanID:    1,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created subscriber: %s\n", subscriber.ID)
}
```

### Analytics API

```go
// Churn Prediction
analyticsAPI := taas.NewAnalyticsAPI(client.HTTPClient)

// Predict churn for a profile
prediction, err := analyticsAPI.PredictChurn(ctx, "profile-123")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Churn Risk: %s (Score: %.2f)\n", prediction.RiskLevel, prediction.RiskScore)

// Get churn metrics
metrics, err := analyticsAPI.GetChurnMetrics(ctx, "monthly")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Monthly Churn Rate: %.2f%%\n", metrics.MonthlyChurnRate)

// Get at-risk customers
atRisk, err := analyticsAPI.GetAtRiskCustomers(ctx, taas.ChurnRiskHigh, 100)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Found %d high-risk customers\n", len(atRisk))

// Market metrics
marketMetrics, err := analyticsAPI.GetMarketMetrics(ctx, "quarterly")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Market Share: %.2f%%\n", marketMetrics.MarketShare)
```

### Security API

```go
// Fraud Detection
securityAPI := taas.NewSecurityAPI(client.HTTPClient)

// Analyze a transaction
alert, err := securityAPI.AnalyzeTransaction(ctx, map[string]interface{}{
    "profile_id":  "profile-123",
    "amount":      99.99,
    "ip_address":  "192.168.1.1",
    "device_id":   "device-456",
    "transaction": "payment",
})
if err != nil {
    log.Fatal(err)
}
if alert != nil {
    fmt.Printf("Fraud Alert: %s (Severity: %s)\n", alert.Type, alert.Severity)
}

// Get fraud alerts
alerts, err := securityAPI.GetFraudAlerts(ctx, taas.FraudAlertFilter{
    Severity: taas.FraudSeverityHigh,
    Status:   "new",
    Limit:    50,
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Found %d fraud alerts\n", len(alerts))

// Get fraud metrics
fraudMetrics, err := securityAPI.GetFraudMetrics(ctx, "monthly")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Resolution Rate: %.2f%%\n", fraudMetrics.ResolutionRate)
```

---

## Python SDK

### Basic Usage

```python
from telecom_sdk import TelecomClient
from telecom_sdk.types import CreateSubscriberRequest

# Initialize client
client = TelecomClient(
    base_url="https://api.telecom.com",
    api_key="your-api-key"
)

# Create a subscriber
subscriber = client.subscribers.create(CreateSubscriberRequest(
    msisdn="+1234567890",
    first_name="John",
    last_name="Doe",
    email="john@example.com",
    plan_id=1
))
print(f"Created subscriber: {subscriber.id}")
```

### Analytics API

```python
from telecom_sdk import TelecomClient
from telecom_sdk.types import ChurnRiskLevel

client = TelecomClient(api_key="your-api-key")

# Churn Prediction
prediction = client.analytics.predict_churn("profile-123")
print(f"Churn Risk: {prediction.risk_level} (Score: {prediction.risk_score:.2f})")

# Get churn metrics
metrics = client.analytics.get_churn_metrics("monthly")
print(f"Monthly Churn Rate: {metrics.monthly_churn_rate:.2f}%")

# Get at-risk customers
at_risk = client.analytics.get_at_risk_customers(
    risk_level=ChurnRiskLevel.HIGH,
    limit=100
)
print(f"Found {len(at_risk)} high-risk customers")

# Market metrics
market = client.analytics.get_market_metrics("quarterly")
print(f"Market Share: {market.market_share_pct:.2f}%")
```

### Security API

```python
from telecom_sdk import TelecomClient
from telecom_sdk.types import FraudType, FraudSeverity, FraudAlertFilter

client = TelecomClient(api_key="your-api-key")

# Analyze transaction
alert = client.security.analyze_transaction({
    "profile_id": "profile-123",
    "amount": 99.99,
    "ip_address": "192.168.1.1",
    "device_id": "device-456",
    "transaction": "payment"
})
if alert:
    print(f"Fraud Alert: {alert.type} (Severity: {alert.severity})")

# Get fraud alerts
alerts = client.security.get_fraud_alerts(FraudAlertFilter(
    severity=FraudSeverity.HIGH,
    status="new",
    limit=50
))
print(f"Found {len(alerts)} fraud alerts")

# Get fraud metrics
fraud_metrics = client.security.get_fraud_metrics("monthly")
print(f"Resolution Rate: {fraud_metrics.resolution_rate_pct:.2f}%")
```

---

## TypeScript SDK

### Basic Usage

```typescript
import { TelecomClient, CreateSubscriberRequest } from '@taas/sdk';

// Initialize client
const client = new TelecomClient({
  baseUrl: 'https://api.telecom.com',
  apiKey: 'your-api-key'
});

// Create a subscriber
const subscriber = await client.subscribers.create({
  msisdn: '+1234567890',
  firstName: 'John',
  lastName: 'Doe',
  email: 'john@example.com',
  planId: 1
});
console.log(`Created subscriber: ${subscriber.id}`);
```

### Analytics API

```typescript
import { TelecomClient, ChurnRiskLevel } from '@taas/sdk';
import { AnalyticsAPI } from '@taas/sdk/analytics';

const client = new TelecomClient({ apiKey: 'your-api-key' });
const analytics = new AnalyticsAPI(client.httpClient);

// Churn Prediction
const prediction = await analytics.predictChurn('profile-123');
console.log(`Churn Risk: ${prediction.riskLevel} (Score: ${prediction.riskScore.toFixed(2)})`);

// Get churn metrics
const metrics = await analytics.getChurnMetrics('monthly');
console.log(`Monthly Churn Rate: ${metrics.monthlyChurnRate.toFixed(2)}%`);

// Get at-risk customers
const atRisk = await analytics.getAtRiskCustomers(ChurnRiskLevel.High, 100);
console.log(`Found ${atRisk.length} high-risk customers`);

// Market metrics
const market = await analytics.getMarketMetrics('quarterly');
console.log(`Market Share: ${market.marketSharePct.toFixed(2)}%`);
```

### Security API

```typescript
import { TelecomClient, FraudSeverity, FraudAlertFilter } from '@taas/sdk';
import { SecurityAPI } from '@taas/sdk/security';

const client = new TelecomClient({ apiKey: 'your-api-key' });
const security = new SecurityAPI(client.httpClient);

// Analyze transaction
const alert = await security.analyzeTransaction({
  profileId: 'profile-123',
  amount: 99.99,
  ipAddress: '192.168.1.1',
  deviceId: 'device-456',
  transaction: 'payment'
});
if (alert) {
  console.log(`Fraud Alert: ${alert.type} (Severity: ${alert.severity})`);
}

// Get fraud alerts
const alerts = await security.getFraudAlerts({
  severity: FraudSeverity.High,
  status: 'new',
  limit: 50
});
console.log(`Found ${alerts.length} fraud alerts`);

// Get fraud metrics
const fraudMetrics = await security.getFraudMetrics('monthly');
console.log(`Resolution Rate: ${fraudMetrics.resolutionRatePct.toFixed(2)}%`);
```

---

## Analytics API

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/analytics/churn/predict` | Predict churn for a profile |
| GET | `/api/v1/analytics/churn/metrics` | Get churn metrics |
| GET | `/api/v1/analytics/churn/at-risk` | Get at-risk customers |
| GET | `/api/v1/analytics/market/metrics` | Get market metrics |
| GET | `/api/v1/analytics/maintenance/metrics` | Get predictive maintenance metrics |
| GET | `/api/v1/analytics/pricing/metrics` | Get pricing metrics |
| POST | `/api/v1/analytics/pricing/optimize` | Optimize pricing for rate plans |

### Churn Risk Levels

- `low` - Low risk of churn (< 25% probability)
- `medium` - Medium risk of churn (25-50% probability)
- `high` - High risk of churn (50-75% probability)
- `critical` - Critical risk of churn (> 75% probability)

---

## Security API

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/security/fraud/analyze` | Analyze transaction for fraud |
| POST | `/api/v1/security/fraud/alerts` | Get fraud alerts with filters |
| PUT | `/api/v1/security/fraud/alerts/:id` | Update alert status |
| GET | `/api/v1/security/fraud/metrics` | Get fraud metrics |

### Fraud Types

- `account_takeover` - Unauthorized account access
- `subscription_fraud` - Fraudulent subscription creation
- `payment_fraud` - Fraudulent payment transactions
- `usage_anomaly` - Abnormal usage patterns
- `sim_swap` - Unauthorized SIM swap attempts

### Fraud Severity Levels

- `low` - Low severity, monitor only
- `medium` - Medium severity, review required
- `high` - High severity, immediate action needed
- `critical` - Critical severity, block transaction

---

## Currency & Billing API

### Endpoints

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

### Example: Currency Conversion

```typescript
// TypeScript
const result = await client.currency.convert({
  from: 'USD',
  to: 'EUR',
  amount: 100.00
});
console.log(`${result.amount} ${result.to} = ${result.converted} ${result.from}`);
```

```python
# Python
result = client.currency.convert(
    from_currency="USD",
    to_currency="EUR",
    amount=100.00
)
print(f"{result.amount} {result.to_currency} = {result.converted} {result.from_currency}")
```

```go
// Go
result, err := client.Currency.Convert(ctx, &taas.ConvertRequest{
    From:   "USD",
    To:     "EUR",
    Amount: 100.00,
})
fmt.Printf("%f %s = %f %s\n", result.Amount, result.To, result.Converted, result.From)
```

---

## Error Handling

All SDKs provide consistent error handling:

### Go

```go
subscriber, err := client.Subscribers.Get(ctx, "invalid-id")
if err != nil {
    if apiErr, ok := err.(*taas.APIError); ok {
        fmt.Printf("API Error: %s (Code: %d)\n", apiErr.Message, apiErr.StatusCode)
    } else {
        fmt.Printf("Network Error: %v\n", err)
    }
}
```

### Python

```python
from telecom_sdk.exceptions import APIError, NetworkError

try:
    subscriber = client.subscribers.get("invalid-id")
except APIError as e:
    print(f"API Error: {e.message} (Code: {e.status_code})")
except NetworkError as e:
    print(f"Network Error: {e}")
```

### TypeScript

```typescript
import { APIError, NetworkError } from '@taas/sdk';

try {
  const subscriber = await client.subscribers.get('invalid-id');
} catch (error) {
  if (error instanceof APIError) {
    console.log(`API Error: ${error.message} (Code: ${error.statusCode})`);
  } else if (error instanceof NetworkError) {
    console.log(`Network Error: ${error.message}`);
  }
}
```

---

## Rate Limiting

The API enforces rate limits per user:
- **Default**: 100 requests per minute
- **Burst**: Up to 200 requests in short bursts

SDKs automatically handle rate limiting with exponential backoff:

```typescript
const client = new TelecomClient({
  apiKey: 'your-api-key',
  retryConfig: {
    maxRetries: 3,
    retryDelay: 1000, // 1 second
    retryOnRateLimit: true
  }
});
```

---

## Support

- **Documentation**: [https://docs.telecom-platform.com](https://docs.telecom-platform.com)
- **GitHub Issues**: [https://github.com/nutcas3/telecom-platform/issues](https://github.com/nutcas3/telecom-platform/issues)
- **Email**: sdk-support@telecom-platform.com
