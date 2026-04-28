# Telecom Platform Elixir SDK

Elixir SDK for the Telecom Platform with GenServer-based async access and full pattern matching.

## Features

- **Modular Architecture**: Separate modules for authentication, HTTP client, and API services
- **GenServer Support**: Full GenServer-based async access
- **Pattern Matching**: Full Elixir pattern matching
- **Type Safety**: Struct-based data structures with typespecs
- **Error Handling**: Comprehensive error handling
- **Retry Logic**: Built-in retry mechanism with exponential backoff
- **OTP Integration**: Full OTP application support

## Installation

Add to your `mix.exs`:

```elixir
defp deps do
  [
    {:telecom_sdk, path: "../sdk/elixir"}
  ]
end
```

Or install from source:

```bash
cd sdk/elixir
mix deps.get
mix compile
```

## Architecture

The SDK is organized into modular components:

- `auth.ex` - Authentication provider handling API keys and JWT tokens
- `http_client.ex` - HTTP client with retry logic and authentication
- `subscribers_api.ex` - Subscriber management API
- `usage_api.ex` - Usage tracking API
- `payments_api.ex` - Payment processing API
- `rating_plans_api.ex` - Rating plan management API
- `system_api.ex` - System monitoring API
- `graphql_api.ex` - GraphQL query execution
- `telecom_sdk.ex` - Main SDK client integrating all modules

## Quick Start

```elixir
# Start the SDK
{:ok, _pid} = TelecomSDK.start_link(
  api_url: "http://localhost:8000",
  api_key: "your-api-key",
  enable_logging: true
)

# Get subscriber
{:ok, subscriber} = TelecomSDK.get_subscriber(1)
IO.puts("Subscriber: #{subscriber.first_name} #{subscriber.last_name}")

# List subscribers
{:ok, list} = TelecomSDK.list_subscribers(1, 50, "active")
IO.puts("Total subscribers: #{list.total}")
```

## API Reference

### Configuration

```elixir
config = %TelecomSDK.Config{
  api_url: "http://localhost:8000",
  api_key: "your-api-key",
  jwt_secret: "your-jwt-secret",
  timeout: 30_000,
  max_retries: 3,
  retry_delay: 1_000,
  enable_logging: false
}
```

### Subscriber Management

```elixir
# Get subscriber
{:ok, subscriber} = TelecomSDK.get_subscriber(1)

# List subscribers
{:ok, list} = TelecomSDK.list_subscribers(1, 50, "active")

# Create subscriber
{:ok, subscriber} = TelecomSDK.create_subscriber(%{
  imsi: "208930000000001",
  msisdn: "+1234567890",
  first_name: "John",
  last_name: "Doe",
  email: "john.doe@example.com",
  plan_id: 1
})

# Update subscriber
{:ok, subscriber} = TelecomSDK.update_subscriber(1, %{
  first_name: "Jane"
})

# Delete subscriber
{:ok, success} = TelecomSDK.delete_subscriber(1)

# Suspend subscriber
{:ok, subscriber} = TelecomSDK.suspend_subscriber(1)

# Activate subscriber
{:ok, subscriber} = TelecomSDK.activate_subscriber(1)
```

### Usage Management

```elixir
# Get usage stats
start_date = DateTime.add(DateTime.utc_now(), -30, :day)
end_date = DateTime.utc_now()
{:ok, stats} = TelecomSDK.get_usage_stats(1, start_date, end_date)

# List usage events
{:ok, events} = TelecomSDK.list_usage_events(%{
  subscriber_id: 1,
  usage_type: "data",
  page: 1,
  page_size: 50
})

# Get real-time usage
{:ok, realtime} = TelecomSDK.get_real_time_usage(1)
```

### Payment Management

```elixir
# Create payment transaction
{:ok, transaction} = TelecomSDK.create_payment_transaction(%{
  subscriber_id: "1",
  amount: 25.00,
  currency: "USD",
  gateway: "stripe",
  metadata: %{description: "Monthly plan"}
})

# Get transaction
{:ok, transaction} = TelecomSDK.get_payment_transaction("txn-123")

# List transactions
{:ok, transactions} = TelecomSDK.list_payment_transactions(%{
  subscriber_id: 1,
  status: "completed",
  page: 1,
  page_size: 50
})
```

### Rating Plans

```elixir
# List rating plans
{:ok, plans} = TelecomSDK.list_rating_plans()

# Get specific plan
{:ok, plan} = TelecomSDK.get_rating_plan("plan-123")
```

### System Management

```elixir
# Get system stats
{:ok, stats} = TelecomSDK.get_system_stats()

# Get health status
{:ok, health} = TelecomSDK.get_health_status()
```

### GraphQL

```elixir
# Execute GraphQL query
query = """
query GetSubscribers($first: Int) {
  subscribers(first: $first) {
    id
    firstName
    lastName
  }
}
"""

variables = %{first: 10}
{:ok, result} = TelecomSDK.execute_graphql_query(query, variables)
```

## Authentication

### JWT Token Generation

```elixir
# Generate JWT token
{:ok, token} = TelecomSDK.generate_jwt_token("user-123", 24, %{custom_claim: "value"})

# Validate JWT token
{:ok, claims} = TelecomSDK.validate_jwt_token(token)
```

## Error Handling

```elixir
case TelecomSDK.get_subscriber(1) do
  {:ok, subscriber} ->
    IO.puts("Subscriber: #{subscriber.first_name}")
  {:error, :authentication_error} ->
    IO.puts("Authentication failed")
  {:error, :api_error} ->
    IO.puts("API error")
  {:error, :network_error} ->
    IO.puts("Network error")
  {:error, reason} ->
    IO.puts("Error: #{inspect(reason)}")
end
```

## Development

### Running Tests

```bash
mix test
```

### Code Style

```bash
mix format
mix compile --warnings-as-errors
```

### Building

```bash
mix compile
```

## Requirements

- Elixir 1.13+
- OTP 26+

## License

This SDK is licensed under the MIT License.
