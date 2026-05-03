package com.telecom

/**
 * Telecom Platform Kotlin SDK
 * 
 * Provides coroutine-based access to the Telecom Platform API with full Kotlin serialization support.
 */
class TelecomSDK private constructor(
    private val config: TelecomConfig,
    private val authProvider: AuthProvider,
    private val httpClient: HTTPClient
) {
    
    // API modules
    val subscribers: SubscriberAPI
    val usage: UsageAPI
    val payments: PaymentAPI
    val ratingPlans: RatingPlanAPI
    val system: SystemAPI
    val analytics: AnalyticsAPI
    val security: SecurityAPI
    val currency: CurrencyAPI
    
    companion object {
        /**
         * Create a new TelecomSDK instance
         */
        fun create(config: TelecomConfig = TelecomConfig()): TelecomSDK {
            val authProvider = AuthProvider(apiKey = config.apiKey, jwtSecret = config.jwtSecret)
            val httpClient = HTTPClient(config = config, authProvider = authProvider)
            
            // Initialize API modules
            val subscribers = SubscriberAPI(httpClient)
            val usage = UsageAPI(httpClient)
            val payments = PaymentAPI(httpClient)
            val ratingPlans = RatingPlanAPI(httpClient)
            val system = SystemAPI(httpClient)
            val analytics = AnalyticsAPI(httpClient)
            val security = SecurityAPI(httpClient)
            val currency = CurrencyAPI(httpClient)
            
            return TelecomSDK(config, authProvider, httpClient, subscribers, usage, payments, ratingPlans, system, analytics, security, currency)
        }
    }
    
    private constructor(
        config: TelecomConfig,
        authProvider: AuthProvider,
        httpClient: HTTPClient,
        subscribers: SubscriberAPI,
        usage: UsageAPI,
        payments: PaymentAPI,
        ratingPlans: RatingPlanAPI,
        system: SystemAPI,
        analytics: AnalyticsAPI,
        security: SecurityAPI,
        currency: CurrencyAPI
    ) {
        this.config = config
        this.authProvider = authProvider
        this.httpClient = httpClient
        this.subscribers = subscribers
        this.usage = usage
        this.payments = payments
        this.ratingPlans = ratingPlans
        this.system = system
        this.analytics = analytics
        this.security = security
        this.currency = currency
    }
    
    // Authentication methods
    
    /**
     * Generate a JWT token for authentication
     */
    fun generateJWTToken(userId: String, expiryHours: Int, additionalClaims: Map<String, Any> = emptyMap()): String {
        return authProvider.generateJWTToken(userId, expiryHours, additionalClaims)
    }
    
    /**
     * Validate a JWT token
     */
    fun validateJWTToken(token: String): Map<String, Any> {
        return authProvider.validateJWTToken(token)
    }
    
    // Cleanup
    
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
    val jwtSecret: String? = null,
    val timeout: java.time.Duration = java.time.Duration.ofSeconds(30),
    val maxRetries: Int = 3,
    val retryDelay: Long = 1 // in seconds
)
