# Telecom Platform Python SDK

Async Python SDK for the Telecom Platform with full type hints and modern async/await support.

## Features

- **Modular Architecture**: Separate modules for authentication, API clients, and WebSocket
- **Async/Await Support**: Full async/await support with aiohttp
- **Type Safety**: Complete type hints with Pydantic models
- **WebSocket Support**: Real-time updates via WebSocket
- **GraphQL Support**: Execute GraphQL queries
- **Retry Logic**: Built-in retry mechanism with exponential backoff
- **Error Handling**: Comprehensive error handling with custom exceptions
- **Logging**: Optional debug logging
- **Context Manager**: Async context manager support

## Installation

```bash
pip install telecom-sdk
```

Or install from source:

```bash
git clone https://github.com/nutcas3/telecom-platform.git
cd telecom-platform/sdk/python
# Use uv for faster installation (recommended)
uv pip install -r requirements.txt
uv pip install -e .
# Or traditional pip
pip install -r requirements.txt
pip install -e .
```

## Architecture

The SDK is organized into modular components:

- `auth.py` - Authentication provider handling API keys and JWT tokens
- `api.py` - Main API client with HTTP request handling and retry logic
- `websocket.py` - WebSocket client for real-time updates
- `types.py` - Type definitions and Pydantic models

## Quick Start

```python
import asyncio
from telecom_sdk import TelecomSDK

async def main():
    async with TelecomSDK(
        api_url="http://localhost:8000",
        api_key="your-api-key",
        enable_logging=True
    ) as sdk:
        
        # Get subscriber
        subscriber = await sdk.get_subscriber(1)
        print(f"Subscriber: {subscriber.first_name} {subscriber.last_name}")
        
        # Get usage stats
        from datetime import datetime, timedelta
        end_date = datetime.now()
        start_date = end_date - timedelta(days=30)
        
        usage = await sdk.get_usage_stats(1, start_date, end_date)
        print(f"Data usage: {usage.data_up + usage.data_down} bytes")
        
        # Create payment transaction
        transaction = await sdk.create_payment_transaction(
            subscriber_id=1,
            amount=25.00,
            currency="USD"
        )
        print(f"Transaction ID: {transaction.id}")

if __name__ == "__main__":
    asyncio.run(main())
```

## WebSocket Example

```python
import asyncio
from telecom_sdk import TelecomSDK, WebSocketMessage

async def handle_message(message: WebSocketMessage):
    """Handle WebSocket messages."""
    print(f"Received message type: {message.type}")
    print(f"Message data: {message.data}")

async def main():
    sdk = TelecomSDK(api_url="http://localhost:8000")
    
    # Connect to WebSocket
    await sdk.connect_websocket(handle_message)
    
    # Keep running
    try:
        while True:
            await asyncio.sleep(1)
    except KeyboardInterrupt:
        await sdk.disconnect_websocket()

if __name__ == "__main__":
    asyncio.run(main())
```

## GraphQL Example

```python
import asyncio
from telecom_sdk import TelecomSDK

async def main():
    async with TelecomSDK(api_url="http://localhost:8000") as sdk:
        
        query = """
        query GetSubscribers($first: Int, $after: String) {
            subscribers(first: $first, after: $after) {
                edges {
                    node {
                        id
                        firstName
                        lastName
                        email
                        status
                    }
                }
                pageInfo {
                    hasNextPage
                    endCursor
                }
            }
        }
        """
        
        variables = {"first": 10}
        result = await sdk.execute_graphql_query(query, variables)
        
        for edge in result["data"]["subscribers"]["edges"]:
            subscriber = edge["node"]
            print(f"{subscriber['firstName']} {subscriber['lastName']}")

if __name__ == "__main__":
    asyncio.run(main())
```

## API Reference

### Subscriber Management

```python
# Get subscriber
subscriber = await sdk.get_subscriber(subscriber_id)

# List subscribers
subscribers = await sdk.list_subscribers(page=1, page_size=50, status="active")

# Create subscriber
subscriber = await sdk.create_subscriber(
    imsi="208930000000001",
    msisdn="+1234567890",
    first_name="John",
    last_name="Doe",
    email="john.doe@example.com",
    plan_id=1
)

# Update subscriber
subscriber = await sdk.update_subscriber(subscriber_id, first_name="Jane")

# Suspend subscriber
subscriber = await sdk.suspend_subscriber(subscriber_id)

# Activate subscriber
subscriber = await sdk.activate_subscriber(subscriber_id)

# Terminate subscriber
success = await sdk.terminate_subscriber(subscriber_id)
```

### Usage Management

```python
from datetime import datetime, timedelta

# Get usage stats
usage = await sdk.get_usage_stats(
    subscriber_id=1,
    start_date=datetime.now() - timedelta(days=30),
    end_date=datetime.now()
)

# List usage events
events = await sdk.list_usage_events(
    subscriber_id=1,
    usage_type="data",
    page=1,
    page_size=50
)

# Get real-time usage
realtime = await sdk.get_real_time_usage(subscriber_id)
```

### Payment Management

```python
# Create payment transaction
transaction = await sdk.create_payment_transaction(
    subscriber_id=1,
    amount=25.00,
    currency="USD",
    gateway="stripe",
    metadata={"description": "Monthly plan"}
)

# Get transaction
transaction = await sdk.get_payment_transaction(transaction_id)

# List transactions
transactions = await sdk.list_payment_transactions(
    subscriber_id=1,
    status="completed"
)
```

### System Management

```python
# Get system stats
stats = await sdk.get_system_stats()

# Get health status
health = await sdk.get_health_status()
```

### Rating Plans

```python
# List rating plans
plans = await sdk.list_rating_plans()

# Get specific plan
plan = await sdk.get_rating_plan(plan_id)
```

## Error Handling

```python
from telecom_sdk import TelecomSDK, AuthenticationError, APIError, NetworkError

async def main():
    try:
        async with TelecomSDK(api_url="http://localhost:8000") as sdk:
            subscriber = await sdk.get_subscriber(1)
    except AuthenticationError:
        print("Authentication failed")
    except APIError as e:
        print(f"API error: {e}")
    except NetworkError as e:
        print(f"Network error: {e}")
    except TelecomError as e:
        print(f"Telecom error: {e}")
```

## Configuration

The SDK can be configured with the following options:

- `api_url`: Base URL for the API server (default: "http://localhost:8000")
- `api_key`: API key for authentication
- `timeout`: Request timeout in seconds (default: 30)
- `max_retries`: Maximum number of retry attempts (default: 3)
- `retry_delay`: Delay between retries in seconds (default: 1.0)
- `enable_logging`: Enable debug logging (default: False)

## Development

### Running Tests

```bash
# Install development dependencies
pip install -r requirements-dev.txt

# Run tests
pytest tests/

# Run tests with coverage
pytest tests/ --cov=telecom_sdk
```

### Code Style

```bash
# Format code
black telecom_sdk/

# Lint code
flake8 telecom_sdk/

# Type checking
mypy telecom_sdk/
```

## License

This SDK is licensed under the MIT License.
