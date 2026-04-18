package com.telecom

import io.ktor.client.*
import io.ktor.client.call.*
import io.ktor.client.plugins.*
import io.ktor.client.plugins.contentnegotiation.*
import io.ktor.client.request.*
import io.ktor.client.statement.*
import io.ktor.http.*
import io.ktor.serialization.kotlinx.json.*
import io.ktor.websocket.*
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.*
import kotlinx.serialization.*
import kotlinx.serialization.json.*
import java.time.*
import java.time.format.DateTimeFormatter

/**
 * Telecom Platform Kotlin SDK
 * 
 * Provides coroutine-based access to the Telecom Platform API with full Kotlin serialization support.
 */
class TelecomSDK private constructor(
    private val httpClient: HttpClient,
    private val config: TelecomConfig,
    private val webSocketClient: WebSocketClient
) {
    
    companion object {
        /**
         * Create a new TelecomSDK instance
         */
        fun create(config: TelecomConfig = TelecomConfig()): TelecomSDK {
            val httpClient = HttpClient {
                defaultRequest {
                    url(config.apiURL)
                    header("User-Agent", "Telecom-Kotlin-SDK/1.0.0")
                    config.apiKey?.let { header("Authorization", "Bearer $it") }
                }
                
                install(ContentNegotiation) {
                    json(Json {
                        ignoreUnknownKeys = true
                        isLenient = true
                    })
                }
                
                install(HttpTimeout) {
                    requestTimeoutMillis = config.timeout.toMillis()
                }
                
                install(Retry) {
                    maxRetries = config.maxRetries
                    retryOnExceptionIf { request, throwable ->
                        throwable is HttpRequestTimeoutException || 
                        throwable is ConnectTimeoutException
                    }
                    exponentialDelay()
                }
                
                HttpResponseValidator {
                    handleResponseExceptionWithRequest { exception, request ->
                        when (exception) {
                            is ClientRequestException -> {
                                when (exception.response.status) {
                                    HttpStatusCode.Unauthorized -> throw TelecomException.AuthenticationError("Authentication failed")
                                    HttpStatusCode.TooManyRequests -> throw TelecomException.RateLimitError("Rate limit exceeded")
                                    in HttpStatusCode.BadRequest..HttpStatusCode.UnprocessableEntity -> {
                                        val errorBody = exception.response.body<String>()
                                        throw TelecomException.APIError("API error: $errorBody")
                                    }
                                    in HttpStatusCode.InternalServerError..HttpStatusCode.NetworkAuthenticationRequired -> {
                                        throw TelecomException.ServerError("Server error: ${exception.response.status}")
                                    }
                                }
                            }
                        }
                    }
                }
            }
            
            val webSocketClient = WebSocketClient()
            
            return TelecomSDK(httpClient, config, webSocketClient)
        }
    }
    
    // Subscriber Management
    
    /**
     * Get subscriber by ID
     */
    suspend fun getSubscriber(id: Long): Subscriber {
        return httpClient.get("/v1/subscribers/$id").body()
    }
    
    /**
     * List subscribers with pagination
     */
    suspend fun listSubscribers(
        page: Int = 1,
        pageSize: Int = 50,
        status: SubscriberStatus? = null
    ): SubscriberList {
        val params = mutableListOf(
            "page" to page.toString(),
            "page_size" to pageSize.toString()
        )
        status?.let { params.add("status" to it.value) }
        
        return httpClient.get("/v1/subscribers") {
            parameter(params)
        }.body()
    }
    
    /**
     * Create a new subscriber
     */
    suspend fun createSubscriber(request: CreateSubscriberRequest): Subscriber {
        return httpClient.post("/v1/subscribers") {
            setBody(request)
        }.body()
    }
    
    /**
     * Update an existing subscriber
     */
    suspend fun updateSubscriber(id: Long, request: UpdateSubscriberRequest): Subscriber {
        return httpClient.put("/v1/subscribers/$id") {
            setBody(request)
        }.body()
    }
    
    /**
     * Suspend a subscriber
     */
    suspend fun suspendSubscriber(id: Long): Subscriber {
        return httpClient.post("/v1/subscribers/$id/suspend").body()
    }
    
    /**
     * Activate a suspended subscriber
     */
    suspend fun activateSubscriber(id: Long): Subscriber {
        return httpClient.post("/v1/subscribers/$id/activate").body()
    }
    
    /**
     * Terminate a subscriber
     */
    suspend fun terminateSubscriber(id: Long): Boolean {
        return httpClient.delete("/v1/subscribers/$id").body()
    }
    
    // Usage Management
    
    /**
     * Get usage statistics for a subscriber
     */
    suspend fun getUsageStats(
        subscriberId: Long,
        startDate: LocalDateTime,
        endDate: LocalDateTime
    ): UsageStats {
        return httpClient.get("/v1/subscribers/$subscriberId/usage") {
            parameter("start_date", startDate.format(DateTimeFormatter.ISO_DATE_TIME))
            parameter("end_date", endDate.format(DateTimeFormatter.ISO_DATE_TIME))
        }.body()
    }
    
    /**
     * List usage events with filtering
     */
    suspend fun listUsageEvents(
        subscriberId: Long? = null,
        usageType: UsageType? = null,
        startDate: LocalDateTime? = null,
        endDate: LocalDateTime? = null,
        page: Int = 1,
        pageSize: Int = 50
    ): UsageEventList {
        val params = mutableListOf(
            "page" to page.toString(),
            "page_size" to pageSize.toString()
        )
        
        subscriberId?.let { params.add("subscriber_id" to it.toString()) }
        usageType?.let { params.add("usage_type" to it.value) }
        startDate?.let { params.add("start_date", it.format(DateTimeFormatter.ISO_DATE_TIME)) }
        endDate?.let { params.add("end_date", it.format(DateTimeFormatter.ISO_DATE_TIME)) }
        
        return httpClient.get("/v1/usage/events") {
            parameter(params)
        }.body()
    }
    
    /**
     * Get real-time usage for a subscriber
     */
    suspend fun getRealTimeUsage(subscriberId: Long): RealTimeUsage {
        return httpClient.get("/v1/subscribers/$subscriberId/realtime").body()
    }
    
    // Payment Management
    
    /**
     * Create a payment transaction
     */
    suspend fun createPaymentTransaction(request: CreatePaymentRequest): PaymentTransaction {
        return httpClient.post("/v1/payments/transactions") {
            setBody(request)
        }.body()
    }
    
    /**
     * Get payment transaction by ID
     */
    suspend fun getPaymentTransaction(transactionId: String): PaymentTransaction {
        return httpClient.get("/v1/payments/transactions/$transactionId").body()
    }
    
    /**
     * List payment transactions with filtering
     */
    suspend fun listPaymentTransactions(
        subscriberId: Long? = null,
        status: PaymentStatus? = null,
        page: Int = 1,
        pageSize: Int = 50
    ): PaymentTransactionList {
        val params = mutableListOf(
            "page" to page.toString(),
            "page_size" to pageSize.toString()
        )
        
        subscriberId?.let { params.add("subscriber_id" to it.toString()) }
        status?.let { params.add("status" to it.value) }
        
        return httpClient.get("/v1/payments/transactions") {
            parameter(params)
        }.body()
    }
    
    // Rating Plans
    
    /**
     * List all available rating plans
     */
    suspend fun listRatingPlans(): List<RatingPlan> {
        return httpClient.get("/v1/rating-plans").body()
    }
    
    /**
     * Get rating plan by ID
     */
    suspend fun getRatingPlan(planId: String): RatingPlan {
        return httpClient.get("/v1/rating-plans/$planId").body()
    }
    
    // System Management
    
    /**
     * Get system statistics
     */
    suspend fun getSystemStats(): SystemStats {
        return httpClient.get("/v1/system/stats").body()
    }
    
    /**
     * Get system health status
     */
    suspend fun getHealthStatus(): HealthStatus {
        return httpClient.get("/v1/health").body()
    }
    
    // WebSocket Support
    
    /**
     * Connect to WebSocket for real-time updates
     */
    fun connectWebSocket(
        messageHandler: suspend (WebSocketMessage) -> Unit
    ): Flow<WebSocketMessage> = flow {
        val wsUrl = config.apiURL.replace("http://", "ws://") + "/ws"
        
        webSocketClient.webSocket(
            urlString = wsUrl,
            request = {
                config.apiKey?.let { header("Authorization", "Bearer $it") }
            }
        ) { session ->
            session.incoming.consumeAsFlow().collect { frame ->
                when (frame) {
                    is Frame.Text -> {
                        try {
                            val message = Json.decodeFromString<WebSocketMessage>(frame.readText())
                            emit(message)
                            messageHandler(message)
                        } catch (e: Exception) {
                            throw TelecomException.WebSocketError("Failed to parse WebSocket message: $e")
                        }
                    }
                    else -> {
                        // Handle other frame types if needed
                    }
                }
            }
        }
    }.catch { e ->
        throw TelecomException.WebSocketError("WebSocket connection failed: $e")
    }
    
    // GraphQL Support
    
    /**
     * Execute a GraphQL query
     */
    suspend fun executeGraphQLQuery(
        query: String,
        variables: Map<String, Any> = emptyMap()
    ): GraphQLResponse {
        val request = GraphQLRequest(query, variables)
        return httpClient.post("/graphql") {
            setBody(request)
        }.body()
    }
    
    /**
     * Close the SDK and cleanup resources
     */
    fun close() {
        httpClient.close()
    }
}

/**
 * Configuration for Telecom SDK
 */
data class TelecomConfig(
    val apiURL: String = "http://localhost:8000",
    val apiKey: String? = null,
    val timeout: Duration = Duration.ofSeconds(30),
    val maxRetries: Int = 3
)
