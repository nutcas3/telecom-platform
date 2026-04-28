# Telecom Platform Kotlin SDK

Kotlin SDK for the Telecom Platform with coroutines support and type safety.

## Features

- **Modular Architecture**: Separate modules for authentication, HTTP client, and API services
- **Coroutines Support**: Full async/await support with Kotlin coroutines
- **Type Safety**: Strong typing with Kotlin data classes
- **Error Handling**: Comprehensive error handling with sealed classes
- **Retry Logic**: Built-in retry mechanism with exponential backoff
- **Multiplatform Support**: Works on JVM, Android, and more

## Installation

### Gradle (Kotlin DSL)

```kotlin
dependencies {
    implementation("com.telecom:telecom-sdk:1.0.0")
}
```

### Gradle (Groovy DSL)

```groovy
dependencies {
    implementation 'com.telecom:telecom-sdk:1.0.0'
}
```

### Maven

```xml
<dependency>
    <groupId>com.telecom</groupId>
    <artifactId>telecom-sdk</artifactId>
    <version>1.0.0</version>
</dependency>
```

## Architecture

The SDK is organized into modular components:

- `AuthProvider.kt` - Authentication provider handling API keys and JWT tokens
- `HTTPClient.kt` - HTTP client with retry logic and authentication
- `SubscriberAPI.kt` - Subscriber management API
- `UsageAPI.kt` - Usage tracking API
- `PaymentAPI.kt` - Payment processing API
- `RatingPlanAPI.kt` - Rating plan management API
- `SystemAPI.kt` - System monitoring API
- `GraphQLAPI.kt` - GraphQL query execution
- `TelecomSDK.kt` - Main SDK client integrating all modules

## Quick Start

```kotlin
import com.telecom.TelecomSDK
import kotlinx.coroutines.runBlocking

fun main() = runBlocking {
    val sdk = TelecomSDK(
        baseURL = "http://localhost:8000",
        apiKey = "your-api-key"
    )
    
    // Get subscriber
    val subscriber = sdk.subscribers.get(1)
    println("Subscriber: ${subscriber.firstName} ${subscriber.lastName}")
    
    // List subscribers
    val subscribers = sdk.subscribers.list(1, 50, "active")
    println("Total subscribers: ${subscribers.total}")
}
```

## API Reference

### Configuration

```kotlin
val config = TelecomConfig(
    apiURL = "http://localhost:8000",
    apiKey = "your-api-key",
    jwtSecret = "your-jwt-secret",
    timeout = Duration.ofSeconds(30),
    maxRetries = 3,
    retryDelay = Duration.ofSeconds(1),
    enableLogging = false
)
```

### Subscriber Management

```kotlin
// Get subscriber
val subscriber = sdk.subscribers.get(1)

// List subscribers
val subscribers = sdk.subscribers.list(
    page = 1,
    pageSize = 50,
    status = "active"
)

// Create subscriber
val newSubscriber = sdk.subscribers.create(
    msisdn = "+1234567890",
    firstName = "John",
    lastName = "Doe",
    email = "john.doe@example.com",
    planId = 1
)

// Update subscriber
val updated = sdk.subscribers.update(
    id = 1,
    firstName = "Jane",
    email = "jane.doe@example.com"
)

// Delete subscriber
sdk.subscribers.delete(1)

// Suspend subscriber
sdk.subscribers.suspend(1)

// Activate subscriber
sdk.subscribers.activate(1)
```

### Usage Management

```kotlin
import java.time.Duration
import java.time.Instant

// Get usage stats
val start = Instant.now().minus(Duration.ofDays(30))
val end = Instant.now()
val stats = sdk.usage.getStats(1, start, end)

// List usage events
val events = sdk.usage.listEvents(
    subscriberId = 1,
    usageType = "data",
    page = 1,
    pageSize = 50
)

// Get real-time usage
val realtime = sdk.usage.getRealTime(1)
```

### Payment Management

```kotlin
// Create payment transaction
val transaction = sdk.payments.createTransaction(
    subscriberId = "1",
    amount = 25.00,
    currency = "USD",
    gateway = "stripe",
    metadata = mapOf("description" to "Monthly plan")
)

// Get transaction
val transaction = sdk.payments.getTransaction("txn-123")

// List transactions
val transactions = sdk.payments.listTransactions(
    subscriberId = 1,
    status = "completed",
    page = 1,
    pageSize = 50
)
```

### Rating Plans

```kotlin
// List rating plans
val plans = sdk.ratingPlans.list()

// Get specific plan
val plan = sdk.ratingPlans.get("plan-123")
```

### System Management

```kotlin
// Get system stats
val stats = sdk.system.getStats()

// Get health status
val health = sdk.system.getHealth()
```

### GraphQL

```kotlin
// Execute GraphQL query
val query = """
query GetSubscribers($first: Int) {
    subscribers(first: $first) {
        id
        firstName
        lastName
    }
}
"""

val variables = mapOf("first" to 10)
val result = sdk.graphql.execute(query, variables)
```

## Error Handling

```kotlin
try {
    val subscriber = sdk.subscribers.get(1)
} catch (e: AuthError.JWTSecretNotConfigured) {
    println("JWT secret not configured")
} catch (e: AuthError.InvalidTokenFormat) {
    println("Invalid token format")
} catch (e: AuthError.InvalidTokenSignature) {
    println("Invalid token signature")
} catch (e: AuthError.TokenExpired) {
    println("Token expired")
} catch (e: AuthError.SigningFailed) {
    println("Signing failed")
} catch (e: Exception) {
    println("Error: ${e.message}")
}
```

## Development

### Building

```bash
./gradlew build
```

### Testing

```bash
./gradlew test
```

## Requirements

- Kotlin 1.9+
- Java 11+

## License

This SDK is licensed under the MIT License.
