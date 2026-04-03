# TaaS Platform API Documentation

## Overview

The TaaS Platform provides a RESTful API for managing cellular connectivity programmatically. All API endpoints are available at `https://api.taas-platform.com/v1` (production) or `http://localhost:8000/v1` (development).

## Authentication

All API requests require authentication using Bearer tokens:

```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
  https://api.taas-platform.com/v1/esims
```

Get your API key from the [Developer Dashboard](http://localhost:3000).

## Endpoints

### Health Check

**GET** `/health`

Check API server status.

**Response:**
```json
{
  "status": "healthy",
  "service": "api-server",
  "timestamp": 1711929600
}
```

### Create eSIM

**POST** `/v1/esims`

Provision a new eSIM with cellular connectivity.

**Request:**
```json
{
  "data_plan": "1GB",
  "country_code": "US",
  "carrier": "optional-carrier-name"
}
```

**Response:**
```json
{
  "esim_id": "550e8400-e29b-41d4-a716-446655440000",
  "imsi": "208930000000001",
  "iccid": "8933123456789012345",
  "activation_code": "LPA:1$smdp.example.com$MATCHING_ID",
  "status": "provisioned",
  "data_plan": "1GB",
  "country_code": "US",
  "created_at": "2026-04-01T12:00:00Z"
}
```

### Get eSIM Details

**GET** `/v1/esims/:id`

Retrieve information about a specific eSIM.

**Response:**
```json
{
  "esim_id": "550e8400-e29b-41d4-a716-446655440000",
  "imsi": "208930000000001",
  "iccid": "8933123456789012345",
  "status": "active",
  "data_plan": "1GB",
  "created_at": "2026-04-01T12:00:00Z"
}
```

### Get eSIM Usage

**GET** `/v1/esims/:id/usage`

Get data usage statistics for an eSIM.

**Response:**
```json
{
  "esim_id": "550e8400-e29b-41d4-a716-446655440000",
  "data_used": 262144000,
  "data_limit": 1073741824,
  "sessions": [
    {
      "started_at": "2026-04-01T10:00:00Z",
      "ended_at": "2026-04-01T11:00:00Z",
      "bytes_up": 10485760,
      "bytes_down": 52428800
    }
  ],
  "updated_at": "2026-04-01T12:00:00Z"
}
```

### Terminate eSIM

**DELETE** `/v1/esims/:id`

Permanently terminate an eSIM.

**Response:**
```json
{
  "message": "eSIM terminated successfully",
  "esim_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

## Status Codes

- `200 OK` - Request successful
- `201 Created` - Resource created
- `400 Bad Request` - Invalid request parameters
- `401 Unauthorized` - Missing or invalid authentication
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

## Rate Limits

- **Free Tier**: 100 requests/hour
- **Pro Tier**: 10,000 requests/hour
- **Enterprise**: Unlimited

Rate limit headers are included in all responses:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 99
X-RateLimit-Reset: 1711933200
```

## SDKs

### TypeScript/JavaScript

```bash
npm install @taas-platform/sdk
```

```typescript
import { TaasClient } from '@taas-platform/sdk';

const client = new TaasClient({ apiKey: 'YOUR_API_KEY' });

// Create eSIM
const esim = await client.esims.create({
  dataPlan: '1GB',
  countryCode: 'US'
});

// Get usage
const usage = await client.esims.getUsage(esim.esimId);
```

### Go

```bash
go get github.com/taas-platform/sdk-go
```

```go
import "github.com/taas-platform/sdk-go"

client := taas.NewClient("YOUR_API_KEY")

// Create eSIM
esim, err := client.ESims.Create(&taas.CreateESIMRequest{
    DataPlan:    "1GB",
    CountryCode: "US",
})
```

## Webhooks

Configure webhooks to receive real-time notifications:

### Events

- `esim.created` - New eSIM provisioned
- `esim.activated` - eSIM activated on device
- `usage.threshold` - Data usage threshold reached (50%, 75%, 90%)
- `usage.depleted` - Data plan exhausted
- `esim.terminated` - eSIM terminated

### Webhook Payload

```json
{
  "event": "usage.threshold",
  "timestamp": "2026-04-01T12:00:00Z",
  "data": {
    "esim_id": "550e8400-e29b-41d4-a716-446655440000",
    "threshold": 75,
    "data_used": 805306368,
    "data_limit": 1073741824
  }
}
```

## Examples

### Create and Activate eSIM

```bash
# 1. Create eSIM
curl -X POST https://api.taas-platform.com/v1/esims \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "data_plan": "1GB",
    "country_code": "US"
  }'

# 2. Show QR code to user (activation_code)
# User scans QR code on their device

# 3. Monitor usage
curl https://api.taas-platform.com/v1/esims/ESIM_ID/usage \
  -H "Authorization: Bearer YOUR_API_KEY"
```

## Support

- **Documentation**: https://docs.taas-platform.com
- **API Status**: https://status.taas-platform.com
- **Support Email**: support@taas-platform.com
- **GitHub Issues**: https://github.com/nutcas3/telecom-platform/issues
