# Telecom Platform Go SDK

Go SDK for the Telecom Platform with modular architecture for HTTP and gRPC support.

## Features

- **Modular Architecture**: Separate modules for authentication, HTTP client, and API services
- **Dual Protocol Support**: Both HTTP REST and gRPC APIs
- **Type Safety**: Strong typing with Go structs
- **Context Support**: Full context.Context support
- **Error Handling**: Comprehensive error handling
- **Retry Logic**: Built-in retry mechanism
- **Async Operations**: Non-blocking operations

## Architecture

The SDK is organized into modular components:

- `auth.go` - Authentication provider handling API keys and JWT tokens
- `client.go` - HTTP client with retry logic and authentication
- `subscribers_api.go` - Subscriber management API
- `usage_api.go` - Usage tracking API
- `payments_api.go` - Payment processing API
- `rating_plans_api.go` - Rating plan management API
- `system_api.go` - System monitoring API
- `graphql_api.go` - GraphQL query execution
- `telecom.go` - Main client integrating all modules

## Installation

```bash
go get github.com/nutcas3/telecom-sdk-go
```

## Quick Start

### HTTP Client

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/nutcas3/telecom-sdk-go"
)

func main() {
    config := telecom.DefaultConfig()
    config.APIKey = "your-api-key"
    
    client, err := telecom.NewClient(config)
    if err != nil {
        panic(err)
    }
    defer client.Close()
    
    ctx := context.Background()
    
    // Get subscriber
    subscriber, err := client.Subscribers.Get(ctx, 1)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Subscriber: %s %s\n", subscriber.FirstName, subscriber.LastName)
    
    // List subscribers
    list, err := client.Subscribers.List(ctx, 1, 50)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Total subscribers: %d\n", list.Total)
}
```

### gRPC Client

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/nutcas3/telecom-sdk-go"
)

func main() {
    config := telecom.DefaultConfig()
    config.APIKey = "your-api-key"
    config.EnableGRPC = true
    config.EnableHTTP = false // Only gRPC
    
    client, err := telecom.NewClient(config)
    if err != nil {
        panic(err)
    }
    defer client.Close()
    
    ctx := context.Background()
    
    // Get subscriber via gRPC
    req := &telecom.GetSubscriberRequest{Id: 1}
    subscriber, err := client.GetSubscriberGRPC(ctx, req)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Subscriber: %s %s\n", subscriber.FirstName, subscriber.LastName)
}
```

## API Reference

### Configuration

```go
config := &telecom.Config{
    APIURL:     "http://localhost:8000",
    GRPCURL:    "localhost:50051",
    APIKey:     "your-api-key",
    Timeout:    30 * time.Second,
    Retries:    3,
    EnableGRPC: true,
    EnableHTTP: true,
}
```

### Subscriber Management

```go
// Get subscriber
subscriber, err := client.Subscribers.Get(ctx, 1)

// List subscribers
list, err := client.Subscribers.List(ctx, 1, 50, "")

// Create subscriber
req := &telecom.CreateSubscriberRequest{
    IMSI:      "208930000000001",
    MSISDN:    "+1234567890",
    FirstName: "John",
    LastName:  "Doe",
    Email:     "john.doe@example.com",
    PlanID:    1,
}
subscriber, err := client.Subscribers.Create(ctx, req)

// Update subscriber
updateReq := &telecom.UpdateSubscriberRequest{
    FirstName: stringPtr("Jane"),
}
subscriber, err := client.Subscribers.Update(ctx, 1, updateReq)

// Delete subscriber
err := client.Subscribers.Delete(ctx, 1)

// Suspend subscriber
subscriber, err := client.Subscribers.Suspend(ctx, 1)

// Activate subscriber
subscriber, err := client.Subscribers.Activate(ctx, 1)
```

### Usage Management

```go
// Get usage stats
stats, err := client.Usage.GetStats(ctx, 1, time.Now().AddDate(0, -1, 0), time.Now())
fmt.Printf("Data usage: %d bytes\n", stats.DataUp + stats.DataDown)

// List usage events
events, err := client.Usage.ListEvents(ctx, 1, "", 1, 50)

// Get real-time usage
realtime, err := client.Usage.GetRealTime(ctx, 1)
```

### Payment Management

```go
// Create payment transaction
req := &telecom.CreatePaymentRequest{
    SubscriberID: "1",
    Amount:       25.00,
    Currency:     "USD",
    Gateway:      "stripe",
}
transaction, err := client.Payments.CreateTransaction(ctx, req)

// Get transaction
transaction, err := client.Payments.GetTransaction(ctx, "txn-123")

// List transactions
transactions, err := client.Payments.ListTransactions(ctx, 1, "", 1, 50)
```

### Rating Plans

```go
// List rating plans
plans, err := client.RatingPlans.List(ctx)

// Get specific plan
plan, err := client.RatingPlans.Get(ctx, "plan-123")
```

### System Management

```go
// Get system stats
stats, err := client.System.GetStats(ctx)
fmt.Printf("Active sessions: %d\n", stats.ActiveSessions)

// Get health status
health, err := client.System.GetHealth(ctx)
```

## Error Handling

```go
subscriber, err := client.Subscribers.Get(ctx, 1)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "HTTP error: 401"):
        fmt.Println("Authentication failed")
    case strings.Contains(err.Error(), "HTTP error: 404"):
        fmt.Println("Subscriber not found")
    case strings.Contains(err.Error(), "HTTP error: 500"):
        fmt.Println("Server error")
    default:
        fmt.Printf("Error: %v\n", err)
    }
}
```

## gRPC Service Interfaces

The SDK provides interfaces for all gRPC services:

- `SubscriberServiceClient` - Subscriber management
- `PaymentServiceClient` - Payment processing  
- `UsageServiceClient` - Usage tracking
- `SystemServiceClient` - System monitoring

## Development

### Running Tests

```bash
go test ./...
```

### Code Coverage

```bash
go test -cover ./...
```

### Building

```bash
go build ./...
```

## License

This SDK is licensed under the MIT License.
