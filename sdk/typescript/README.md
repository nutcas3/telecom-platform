# Telecom Platform TypeScript SDK

A comprehensive TypeScript SDK for integrating with the Telecom Platform API. This SDK provides a clean, type-safe interface for managing subscribers, monitoring usage, handling payments, and receiving real-time updates.

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
const subscribers = await sdk.subscribersService.listSubscribers({
  page: 1,
  pageSize: 20,
  status: 'active',
  organizationId: 'org-123',
  search: 'John Doe',
});

// Get subscriber by ID
const subscriber = await sdk.subscribersService.getSubscriber(123);

// Get subscriber by IMSI
const subscriber = await sdk.subscribersService.getSubscriberByImsi('123456789012345');

// Create new subscriber
const newSubscriber = await sdk.subscribersService.createSubscriber({
  msisdn: '+1234567890',
  firstName: 'John',
  lastName: 'Doe',
  email: 'john.doe@example.com',
  planId: 1,
});

// Update subscriber
const updated = await sdk.subscribersService.updateSubscriber(123, {
  firstName: 'Jane',
  email: 'jane.doe@example.com',
});

// Delete subscriber
await sdk.subscribersService.deleteSubscriber(123);

// Suspend subscriber
await sdk.subscribersService.suspendSubscriber(123);

// Activate subscriber
await sdk.subscribersService.activateSubscriber(123);

// Terminate subscriber
await sdk.subscribersService.terminateSubscriber(123);
```

### Account Management

```typescript
// Get subscriber account
const account = await sdk.subscribersService.getSubscriberAccount('123456789012345');

// Top up balance
const updatedAccount = await sdk.subscribersService.topUpBalance('123456789012345', {
  amount: 1000, // $10.00 in cents
  paymentMethodId: 'pm_123',
});

// Get usage statistics
const stats = await sdk.subscribersService.getUsageStats('123456789012345', 'monthly');

// Get real-time usage
const realtime = await sdk.subscribersService.getRealTimeUsage('123456789012345');
```

### eSIM Management

```typescript
// Provision eSIM profile
const esim = await sdk.subscribersService.provisionESIM('123456789012345');

// Activate eSIM profile
await sdk.subscribersService.activateESIM('123456789012345', esim.profileId);

// Deactivate eSIM profile
await sdk.subscribersService.deactivateESIM('123456789012345', esim.profileId);

// Get eSIM status
const status = await sdk.subscribersService.getESIMStatus('123456789012345');
```

### Billing & Invoices

```typescript
// Get subscriber invoices
const invoices = await sdk.subscribersService.getInvoices('123456789012345');

// Get specific invoice
const invoice = await sdk.subscribersService.getInvoice('inv_123');

// Download invoice PDF
const pdfBlob = await sdk.subscribersService.downloadInvoicePDF('inv_123');
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
const stats = await sdk.getSystemStats();
console.log('Active sessions:', stats.activeSessions);
console.log('Total accounts:', stats.totalAccounts);

// Get health status
const health = await sdk.getHealthStatus();
console.log('Redis connected:', health.redisConnected);

// Test connection
const test = await sdk.testConnection();
console.log('API status:', test.status);
```

## Error Handling

The SDK provides comprehensive error handling with detailed error information:

```typescript
try {
  const subscriber = await sdk.subscribersService.getSubscriber(123);
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
const subscriber: Subscriber = await sdk.subscribersService.getSubscriber(123);
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
