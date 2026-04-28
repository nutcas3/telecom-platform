# Telecom Platform Rust SDK

Async Rust SDK for the Telecom Platform with full type safety and modern async/await support.

## Features

- **Modular Architecture**: Separate modules for authentication, HTTP client, and API services
- **Async/Await Support**: Full async/await support with tokio
- **Type Safety**: Strong typing with Rust structs and enums
- **Error Handling**: Comprehensive error handling with custom error types
- **Retry Logic**: Built-in retry mechanism with exponential backoff
- **Context Support**: Full context.Context support for cancellation

## Installation

Add to your `Cargo.toml`:

```toml
[dependencies]
telecom-sdk = { path = "../sdk/rust" }
```

Or install from source:

```bash
cd sdk/rust
cargo build
```

## Architecture

The SDK is organized into modular components:

- `auth.rs` - Authentication provider handling API keys and JWT tokens
- `client.rs` - HTTP client with retry logic and authentication
- `error.rs` - Error types for the SDK
- `types.rs` - Data structures and type definitions
- `api/` - Directory containing API modules:
  - `subscribers.rs` - Subscriber management API
  - `usage.rs` - Usage tracking API
  - `payments.rs` - Payment processing API
  - `rating_plans.rs` - Rating plan management API
  - `system.rs` - System monitoring API
  - `graphql.rs` - GraphQL query execution
  - `mod.rs` - Module exports

## Quick Start

```rust
use telecom_sdk::{TelecomClient, Config};
use tokio;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let config = Config {
        api_url: "http://localhost:8000".to_string(),
        api_key: Some("your-api-key".to_string()),
        ..Default::default()
    };

    let client = TelecomClient::new(config).await?;
    
    // Get subscriber
    let subscriber = client.subscribers.get(1).await?;
    println!("Subscriber: {} {}", subscriber.first_name, subscriber.last_name);
    
    // List subscribers
    let list = client.subscribers.list(1, 50, None).await?;
    println!("Total subscribers: {}", list.total);
    
    Ok(())
}
```

## API Reference

### Configuration

```rust
let config = Config {
    api_url: "http://localhost:8000".to_string(),
    api_key: Some("your-api-key".to_string()),
    jwt_secret: Some("your-jwt-secret".to_string()),
    timeout: Some(Duration::from_secs(30)),
    max_retries: Some(3),
    retry_delay: Some(Duration::from_secs(1)),
};
```

### Subscriber Management

```rust
// Get subscriber
let subscriber = client.subscribers.get(1).await?;

// List subscribers
let list = client.subscribers.list(1, 50, Some("active".to_string())).await?;

// Create subscriber
let request = CreateSubscriberRequest {
    imsi: "208930000000001".to_string(),
    msisdn: "+1234567890".to_string(),
    first_name: "John".to_string(),
    last_name: "Doe".to_string(),
    email: "john.doe@example.com".to_string(),
    plan_id: 1,
};
let subscriber = client.subscribers.create(request).await?;

// Update subscriber
let update = UpdateSubscriberRequest {
    first_name: Some("Jane".to_string()),
    ..Default::default()
};
let subscriber = client.subscribers.update(1, update).await?;

// Delete subscriber
client.subscribers.delete(1).await?;

// Suspend subscriber
client.subscribers.suspend(1).await?;

// Activate subscriber
client.subscribers.activate(1).await?;
```

### Usage Management

```rust
use std::time::{Duration, SystemTime};

// Get usage stats
let start_date = SystemTime::now() - Duration::from_secs(30 * 24 * 60 * 60);
let end_date = SystemTime::now();
let stats = client.usage.get_stats(1, start_date, end_date).await?;

// List usage events
let events = client.usage.list_events(1, Some("data".to_string()), 1, 50).await?;

// Get real-time usage
let realtime = client.usage.get_real_time(1).await?;
```

### Payment Management

```rust
// Create payment transaction
let transaction = client.payments.create_transaction(CreatePaymentTransactionRequest {
    subscriber_id: "1".to_string(),
    amount: 25.00,
    currency: "USD".to_string(),
    gateway: "stripe".to_string(),
    metadata: None,
}).await?;

// Get transaction
let transaction = client.payments.get_transaction("txn-123".to_string()).await?;

// List transactions
let transactions = client.payments.list_transactions(1, Some("completed".to_string()), 1, 50).await?;
```

### Rating Plans

```rust
// List rating plans
let plans = client.rating_plans.list().await?;

// Get specific plan
let plan = client.rating_plans.get("plan-123".to_string()).await?;
```

### System Management

```rust
// Get system stats
let stats = client.system.get_stats().await?;

// Get health status
let health = client.system.get_health().await?;
```

### GraphQL

```rust
// Execute GraphQL query
let query = r#"
    query GetSubscribers($first: Int) {
        subscribers(first: $first) {
            id
            firstName
            lastName
        }
    }
"#;

let variables = serde_json::json!({"first": 10});
let result = client.graphql.execute(query, Some(variables)).await?;
```

## Error Handling

```rust
use telecom_sdk::TelecomError;

match client.subscribers.get(1).await {
    Ok(subscriber) => println!("Subscriber: {}", subscriber.first_name),
    Err(TelecomError::AuthenticationError) => println!("Authentication failed"),
    Err(TelecomError::APIError(msg)) => println!("API error: {}", msg),
    Err(TelecomError::NetworkError(msg)) => println!("Network error: {}", msg),
    Err(e) => println!("Error: {}", e),
}
```

## Development

### Running Tests

```bash
cargo test
```

### Building

```bash
cargo build --release
```

### Code Style

```bash
cargo fmt
cargo clippy
```

## License

This SDK is licensed under the MIT License.
