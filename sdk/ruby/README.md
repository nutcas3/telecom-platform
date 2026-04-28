# Telecom Platform Ruby SDK

Ruby SDK for the Telecom Platform with full Ruby patterns and HTTP support.

## Features

- **Modular Architecture**: Separate modules for authentication, HTTP client, and API services
- **HTTParty Integration**: Built on HTTParty for robust HTTP requests
- **Type Safety**: Hash-based data structures with validation
- **Error Handling**: Comprehensive error handling with custom exceptions
- **Retry Logic**: Built-in retry mechanism with exponential backoff
- **WebSocket Support**: Real-time updates via WebSocket

## Installation

Add to your `Gemfile`:

```ruby
gem 'telecom-sdk'
```

Or install manually:

```bash
cd sdk/ruby
gem build telecom_sdk.gemspec
gem install telecom_sdk-1.0.0.gem
```

## Architecture

The SDK is organized into modular components:

- `auth.rb` - Authentication provider handling API keys and JWT tokens
- `http_client.rb` - HTTP client with retry logic and authentication
- `subscribers_api.rb` - Subscriber management API
- `usage_api.rb` - Usage tracking API
- `payments_api.rb` - Payment processing API
- `rating_plans_api.rb` - Rating plan management API
- `system_api.rb` - System monitoring API
- `graphql_api.rb` - GraphQL query execution
- `telecom_sdk.rb` - Main SDK client integrating all modules

## Quick Start

```ruby
require 'telecom_sdk'

sdk = TelecomSDK::Client.new(
  api_url: "http://localhost:8000",
  api_key: "your-api-key",
  enable_logging: true
)

# Get subscriber
subscriber = sdk.subscribers.get(1)
puts "Subscriber: #{subscriber['first_name']} #{subscriber['last_name']}"

# List subscribers
subscribers = sdk.subscribers.list(page: 1, page_size: 50, status: "active")
puts "Total subscribers: #{subscribers['total']}"
```

## API Reference

### Configuration

```ruby
client = TelecomSDK::Client.new(
  api_url: "http://localhost:8000",
  api_key: "your-api-key",
  jwt_secret: "your-jwt-secret",
  timeout: 30,
  max_retries: 3,
  retry_delay: 1,
  enable_logging: false
)
```

### Subscriber Management

```ruby
# Get subscriber
subscriber = sdk.subscribers.get(1)

# List subscribers
subscribers = sdk.subscribers.list(page: 1, page_size: 50, status: "active")

# Create subscriber
subscriber = sdk.subscribers.create(
  imsi: "208930000000001",
  msisdn: "+1234567890",
  first_name: "John",
  last_name: "Doe",
  email: "john.doe@example.com",
  plan_id: 1
)

# Update subscriber
subscriber = sdk.subscribers.update(1, first_name: "Jane")

# Delete subscriber
sdk.subscribers.delete(1)

# Suspend subscriber
subscriber = sdk.subscribers.suspend(1)

# Activate subscriber
subscriber = sdk.subscribers.activate(1)
```

### Usage Management

```ruby
require 'time'

# Get usage stats
start_date = Time.now - 30 * 24 * 60 * 60
end_date = Time.now
stats = sdk.usage.get_stats(1, start_date, end_date)

# List usage events
events = sdk.usage.list_events(
  subscriber_id: 1,
  usage_type: "data",
  page: 1,
  page_size: 50
)

# Get real-time usage
realtime = sdk.usage.get_real_time(1)
```

### Payment Management

```ruby
# Create payment transaction
transaction = sdk.payments.create_transaction(
  subscriber_id: 1,
  amount: 25.00,
  currency: "USD",
  gateway: "stripe",
  metadata: { description: "Monthly plan" }
)

# Get transaction
transaction = sdk.payments.get_transaction("txn-123")

# List transactions
transactions = sdk.payments.list_transactions(
  subscriber_id: 1,
  status: "completed",
  page: 1,
  page_size: 50
)
```

### Rating Plans

```ruby
# List rating plans
plans = sdk.rating_plans.list

# Get specific plan
plan = sdk.rating_plans.get("plan-123")
```

### System Management

```ruby
# Get system stats
stats = sdk.system.get_stats

# Get health status
health = sdk.system.get_health
```

### GraphQL

```ruby
# Execute GraphQL query
query = <<-GRAPHQL
  query GetSubscribers($first: Int) {
    subscribers(first: $first) {
      id
      firstName
      lastName
    }
  }
GRAPHQL

variables = { first: 10 }
result = sdk.graphql.execute(query, variables)
```

## Error Handling

```ruby
begin
  subscriber = sdk.subscribers.get(1)
rescue TelecomSDK::AuthenticationError
  puts "Authentication failed"
rescue TelecomSDK::APIError => e
  puts "API error: #{e.message}"
rescue TelecomSDK::NetworkError => e
  puts "Network error: #{e.message}"
rescue TelecomSDK::ValidationError => e
  puts "Validation error: #{e.message}"
rescue TelecomSDK::RateLimitError => e
  puts "Rate limit error: #{e.message}"
rescue TelecomSDK::ServerError => e
  puts "Server error: #{e.message}"
rescue TelecomSDK::Error => e
  puts "Telecom error: #{e.message}"
end
```

## Authentication

### JWT Token Generation

```ruby
# Generate JWT token
token = sdk.generate_jwt_token("user-123", 24, { custom_claim: "value" })

# Validate JWT token
claims = sdk.validate_jwt_token(token)
```

## Development

### Running Tests

```bash
bundle install
bundle exec rspec
```

### Code Style

```bash
bundle exec rubocop
```

### Building Gem

```bash
gem build telecom_sdk.gemspec
```

## Requirements

- Ruby 2.7+
- HTTParty
- WebSocket Client Simple

## License

This SDK is licensed under the MIT License.
