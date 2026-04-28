# Telecom Platform Swift SDK

Swift SDK for the Telecom Platform with async/await support and type safety.

## Features

- **Modular Architecture**: Separate modules for authentication, HTTP client, and API services
- **Async/Await Support**: Full async/await support with Swift Concurrency
- **Type Safety**: Strong typing with Swift structs and enums
- **Error Handling**: Comprehensive error handling with custom error types
- **Retry Logic**: Built-in retry mechanism with exponential backoff
- **iOS & macOS Support**: Universal framework for Apple platforms

## Installation

### Swift Package Manager

Add to your `Package.swift`:

```swift
dependencies: [
    .package(
        url: "https://github.com/nutcas3/telecom-sdk-swift",
        from: "1.0.0"
    )
]
```

### Manual Installation

```bash
cd sdk/swift
swift build
```

## Architecture

The SDK is organized into modular components:

- `AuthProvider.swift` - Authentication provider handling API keys and JWT tokens
- `HTTPClient.swift` - HTTP client with retry logic and authentication
- `SubscriberAPI.swift` - Subscriber management API
- `UsageAPI.swift` - Usage tracking API
- `PaymentAPI.swift` - Payment processing API
- `RatingPlanAPI.swift` - Rating plan management API
- `SystemAPI.swift` - System monitoring API
- `GraphQLAPI.swift` - GraphQL query execution
- `TelecomSDK.swift` - Main SDK client integrating all modules

## Quick Start

```swift
import TelecomSDK

let sdk = try TelecomSDK(
    baseURL: "http://localhost:8000",
    apiKey: "your-api-key"
)

// Get subscriber
let subscriber = try await sdk.subscribers.get(id: 1)
print("Subscriber: \(subscriber.firstName) \(subscriber.lastName)")

// List subscribers
let subscribers = try await sdk.subscribers.list(page: 1, pageSize: 50, status: "active")
print("Total subscribers: \(subscribers.total)")
```

## API Reference

### Configuration

```swift
let config = TelecomConfig(
    baseURL: "http://localhost:8000",
    apiKey: "your-api-key",
    jwtSecret: "your-jwt-secret",
    timeout: 30.0,
    maxRetries: 3,
    retryDelay: 1.0,
    enableLogging: false
)
```

### Subscriber Management

```swift
// Get subscriber
let subscriber = try await sdk.subscribers.get(id: 1)

// List subscribers
let subscribers = try await sdk.subscribers.list(
    page: 1,
    pageSize: 50,
    status: "active"
)

// Create subscriber
let newSubscriber = try await sdk.subscribers.create(
    msisdn: "+1234567890",
    firstName: "John",
    lastName: "Doe",
    email: "john.doe@example.com",
    planId: 1
)

// Update subscriber
let updated = try await sdk.subscribers.update(
    id: 1,
    firstName: "Jane",
    email: "jane.doe@example.com"
)

// Delete subscriber
try await sdk.subscribers.delete(id: 1)

// Suspend subscriber
try await sdk.subscribers.suspend(id: 1)

// Activate subscriber
try await sdk.subscribers.activate(id: 1)
```

### Usage Management

```swift
import Foundation

// Get usage stats
let start = Date().addingTimeInterval(-30 * 24 * 60 * 60)
let end = Date()
let stats = try await sdk.usage.getStats(
    subscriberId: 1,
    startDate: start,
    endDate: end
)

// List usage events
let events = try await sdk.usage.listEvents(
    subscriberId: 1,
    usageType: "data",
    page: 1,
    pageSize: 50
)

// Get real-time usage
let realtime = try await sdk.usage.getRealTime(subscriberId: 1)
```

### Payment Management

```swift
// Create payment transaction
let transaction = try await sdk.payments.createTransaction(
    subscriberId: 1,
    amount: 25.00,
    currency: "USD",
    gateway: "stripe",
    metadata: ["description": "Monthly plan"]
)

// Get transaction
let transaction = try await sdk.payments.getTransaction(id: "txn-123")

// List transactions
let transactions = try await sdk.payments.listTransactions(
    subscriberId: 1,
    status: "completed",
    page: 1,
    pageSize: 50
)
```

### Rating Plans

```swift
// List rating plans
let plans = try await sdk.ratingPlans.list()

// Get specific plan
let plan = try await sdk.ratingPlans.get(id: "plan-123")
```

### System Management

```swift
// Get system stats
let stats = try await sdk.system.getStats()

// Get health status
let health = try await sdk.system.getHealth()
```

### GraphQL

```swift
// Execute GraphQL query
let query = """
query GetSubscribers($first: Int) {
    subscribers(first: $first) {
        id
        firstName
        lastName
    }
}
"""

let variables = ["first": 10]
let result = try await sdk.graphql.execute(query, variables: variables)
```

## Error Handling

```swift
do {
    let subscriber = try await sdk.subscribers.get(id: 1)
} catch TelecomError.authenticationError {
    print("Authentication failed")
} catch TelecomError.apiError(let message) {
    print("API error: \(message)")
} catch TelecomError.networkError(let message) {
    print("Network error: \(message)")
} catch {
    print("Error: \(error)")
}
```

## Development

### Building

```bash
swift build
```

### Testing

```bash
swift test
```

## Requirements

- iOS 13.0+ / macOS 10.15+
- Swift 5.5+

## License

This SDK is licensed under the MIT License.
