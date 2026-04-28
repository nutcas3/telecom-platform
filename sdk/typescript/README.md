# Telecom Platform TypeScript SDK

A comprehensive TypeScript SDK for integrating with the Telecom Platform API. This SDK provides a clean, type-safe interface for managing subscribers, monitoring usage, handling payments, and receiving real-time updates.

## Architecture

The SDK is organized into modular components:

- `auth.ts` - Authentication provider handling API keys and JWT tokens
- `api/` - Directory containing API modules:
  - `http-client.ts` - HTTP client with retry logic
  - `subscribers.ts` - Subscriber management API
  - `usage.ts` - Usage tracking API
  - `payments.ts` - Payment processing API
  - `rating-plans.ts` - Rating plan management API
  - `system.ts` - System monitoring API
  - `graphql.ts` - GraphQL query execution
- `websocket.ts` - WebSocket client for real-time updates
- `types.ts` - Type definitions and interfaces

## Installation

```bash
npm install @telecom-platform/sdk
# or
yarn add @telecom-platform/sdk
# or
pnpm add @telecom-platform/sdk
```

## Quick Start

```typescript
import { TelecomSDK } from '@telecom-platform/sdk';

// Initialize the SDK
const sdk = TelecomSDK.initialize({
  baseURL: 'https://api.telecom-platform.com',
  apiKey: 'your-api-key',
  websocketURL: 'wss://ws.telecom-platform.com',
  enableLogging: true,
});

// Connect to WebSocket for real-time updates
await sdk.connectWebSocket();

// Subscribe to usage updates for a subscriber
sdk.subscribeToUsage('123456789012345');

// Listen for usage updates
sdk.onUsageUpdate((update) => {
  console.log('Usage update:', update);
});

// Subscribe to alerts
sdk.subscribeToAlerts();

// Listen for alerts
sdk.onAlert((alert) => {
  console.log('Alert:', alert);
});

// List subscribers
const subscribers = await sdk.subscribersService.listSubscribers({
  page: 1,
  pageSize: 20,
  status: 'active',
});

console.log('Subscribers:', subscribers);
```

## API Reference

### Initialization

```typescript
const sdk = TelecomSDK.initialize({
  baseURL: string,           // API base URL (required)
  apiKey?: string,           // API key (optional)
  timeout?: number,          // Request timeout in ms (default: 30000)
  retryAttempts?: number,    // Number of retry attempts (default: 3)
  retryDelay?: number,       // Retry delay in ms (default: 1000)
  enableLogging?: boolean,   // Enable debug logging (default: false)
  websocketURL?: string,     // WebSocket URL for real-time updates
});
```

### Subscriber Management

```typescript
// List subscribers
const subscribers = await sdk.subscribers.listSubscribers({
  page: 1,
  pageSize: 20,
  status: 'active',
});

// Get subscriber by ID
const subscriber = await sdk.subscribers.getSubscriber(123);

// Create new subscriber
const newSubscriber = await sdk.subscribers.createSubscriber({
  msisdn: '+1234567890',
  firstName: 'John',
  lastName: 'Doe',
  email: 'john.doe@example.com',
  planId: 1,
});

// Update subscriber
const updated = await sdk.subscribers.updateSubscriber(123, {
  firstName: 'Jane',
  email: 'jane.doe@example.com',
});

// Delete subscriber
await sdk.subscribers.deleteSubscriber(123);

// Suspend subscriber
await sdk.subscribers.suspendSubscriber(123);

// Activate subscriber
await sdk.subscribers.activateSubscriber(123);
```

### Usage Management

```typescript
// Get usage stats
const stats = await sdk.usage.getStats(1, '2024-01-01', '2024-01-31');

// List usage events
const events = await sdk.usage.listEvents({
  subscriberId: 1,
  usageType: 'data',
  page: 1,
  pageSize: 50,
});

// Get real-time usage
const realtime = await sdk.usage.getRealTimeUsage(1);
```

### Payment Management

```typescript
// Create payment transaction
const transaction = await sdk.payments.createTransaction({
  subscriberId: 1,
  amount: 25.00,
  currency: 'USD',
  gateway: 'stripe',
  metadata: { description: 'Monthly plan' }
});

// Get transaction
const transaction = await sdk.payments.getTransaction('txn-123');

// List transactions
const transactions = await sdk.payments.listTransactions({
  subscriberId: 1,
  status: 'completed',
  page: 1,
  pageSize: 50,
});
```

### Rating Plans

```typescript
// List rating plans
const plans = await sdk.ratingPlans.list();

// Get specific plan
const plan = await sdk.ratingPlans.get('plan-123');
```

### System Management

```typescript
// Get system stats
const stats = await sdk.system.getStats();

// Get health status
const health = await sdk.system.getHealth();
```

### GraphQL

```typescript
// Execute GraphQL query
const result = await sdk.graphql.execute(`
  query GetSubscribers($first: Int) {
    subscribers(first: $first) {
      id
      firstName
      lastName
    }
  }
`, { first: 10 });
```

### Real-time Updates

```typescript
// Connect to WebSocket
await sdk.connectWebSocket();

// Subscribe to usage updates
sdk.subscribeToUsage('123456789012345');

// Subscribe to alerts
sdk.subscribeToAlerts();

// Listen for usage updates
sdk.onUsageUpdate((update) => {
  console.log(`Data used: ${update.dataUsed} bytes`);
  console.log(`Cost: $${update.cost / 100}`);
});

// Listen for alerts
sdk.onAlert((alert) => {
  console.log(`Alert: ${alert.message}`);
  if (alert.severity === 'critical') {
    // Handle critical alert
  }
});

// Custom event handlers
sdk.onWebSocketEvent('custom_event', (message) => {
  console.log('Custom event:', message);
});

// Check connection status
if (sdk.isWebSocketConnected) {
  console.log('WebSocket is connected');
}

// Disconnect
sdk.disconnectWebSocket();
```

### System Monitoring

```typescript
// Get system statistics
const stats = await sdk.system.getStats();
console.log('Active sessions:', stats.activeSessions);
console.log('Total accounts:', stats.totalAccounts);

// Get health status
const health = await sdk.system.getHealth();
console.log('Redis connected:', health.redisConnected);
```

## Error Handling

The SDK provides comprehensive error handling with detailed error information:

```typescript
try {
  const subscriber = await sdk.subscribers.getSubscriber(123);
} catch (error) {
  console.error('Error:', error.code);
  console.error('Message:', error.message);
  console.error('Details:', error.details);
  console.error('Timestamp:', error.timestamp);
}
```

## Configuration Updates

You can update the SDK configuration at runtime:

```typescript
// Update API key
sdk.setApiKey('new-api-key');

// Update base URL
sdk.setBaseURL('https://new-api.telecom-platform.com');

// Enable/disable logging
sdk.setLogging(true);

// Get current configuration
const config = sdk.getConfig();
console.log('Current config:', config);
```

## TypeScript Support

The SDK is fully typed with TypeScript. All methods have proper type definitions and IntelliSense support:

```typescript
import { Subscriber, SubscriberStatus, UsageType } from '@telecom-platform/sdk';

// Types are automatically inferred
const subscriber: Subscriber = await sdk.subscribers.getSubscriber(123);
const status: SubscriberStatus = subscriber.status;
```

## Browser Support

The SDK works in modern browsers that support:
- ES2020 features
- WebSocket API
- Fetch API

For older browsers, you may need to include appropriate polyfills.

## Node.js Usage

The SDK also works in Node.js environments:

```typescript
import { TelecomSDK } from '@telecom-platform/sdk';

// Node.js WebSocket polyfill may be required
global.WebSocket = require('ws');

const sdk = TelecomSDK.initialize({
  baseURL: 'https://api.telecom-platform.com',
  apiKey: process.env.API_KEY,
});
```

## Examples

See the `examples/` directory for complete examples of:
- Basic subscriber management
- Real-time usage monitoring
- eSIM provisioning
- Billing integration
- Custom dashboard implementation

## Support

- Documentation: [https://docs.telecom-platform.com/sdk](https://docs.telecom-platform.com/sdk)
- Issues: [GitHub Issues](https://github.com/telecom-platform/sdk/issues)
- Support: support@telecom-platform.com

## License

MIT License - see LICENSE file for details.
