# Telecom Platform Go SDK

Go SDK for the Telecom Platform with both HTTP and gRPC support.

## Features

- **Dual Protocol Support**: Both HTTP REST and gRPC APIs
- **Type Safety**: Strong typing with Go structs
- **Context Support**: Full context.Context support
- **Error Handling**: Comprehensive error handling
- **Retry Logic**: Built-in retry mechanism
- **Async Operations**: Non-blocking operations

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
    subscriber, err := client.GetSubscriber(ctx, 1)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Subscriber: %s %s\n", subscriber.FirstName, subscriber.LastName)
    
    // List subscribers
    list, err := client.ListSubscribers(ctx, 1, 50)
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
subscriber, err := client.GetSubscriber(ctx, 1)

// List subscribers
list, err := client.ListSubscribers(ctx, 1, 50)

// Create subscriber
req := &telecom.CreateSubscriberRequest{
    IMSI:      "208930000000001",
    MSISDN:    "+1234567890",
    FirstName: "John",
    LastName:  "Doe",
    Email:     "john.doe@example.com",
    PlanID:    1,
}
subscriber, err := client.CreateSubscriber(ctx, req)

// Update subscriber
updateReq := &telecom.UpdateSubscriberRequest{
    FirstName: stringPtr("Jane"),
}
subscriber, err := client.UpdateSubscriber(ctx, 1, updateReq)

// Delete subscriber
err := client.DeleteSubscriber(ctx, 1)
```

### Usage Management

```go
// Get usage stats
stats, err := client.GetUsageStats(ctx, 1, time.Now().AddDate(0, -1, 0), time.Now())
fmt.Printf("Data usage: %d bytes\n", stats.DataUp + stats.DataDown)
```

### Payment Management

```go
// Create payment
req := &telecom.CreatePaymentRequest{
    SubscriberID: "1",
    Amount:       25.00,
    Currency:     "USD",
    Gateway:      "stripe",
}
transaction, err := client.CreatePaymentTransaction(ctx, req)
```

### System Management

```go
// Get system stats
stats, err := client.GetSystemStats(ctx)
fmt.Printf("Active sessions: %d\n", stats.ActiveSessions)
```

## Error Handling

```go
subscriber, err := client.GetSubscriber(ctx, 1)
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
