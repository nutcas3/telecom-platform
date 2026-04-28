import Foundation
import NIO
import AsyncAlgorithms

/// Telecom Platform Swift SDK
/// 
/// Provides async/await support with Swift Concurrency for the Telecom Platform API.
public class TelecomSDK {
    
    // MARK: - Properties
    
    private let config: TelecomConfig
    private let authProvider: AuthProvider
    private let httpClient: HTTPClient
    
    // API modules
    public let subscribers: SubscriberAPI
    public let usage: UsageAPI
    public let payments: PaymentAPI
    public let ratingPlans: RatingPlanAPI
    public let system: SystemAPI
    public let graphql: GraphQLAPI
    
    // MARK: - Initialization
    
    /// Initialize a new TelecomSDK instance
    /// - Parameter config: Configuration for the SDK
    public init(config: TelecomConfig = TelecomConfig()) {
        self.config = config
        self.authProvider = AuthProvider(apiKey: config.apiKey, jwtSecret: config.jwtSecret)
        self.httpClient = HTTPClient(config: config, authProvider: authProvider)
        
        // Initialize API modules
        self.subscribers = SubscriberAPI(client: httpClient)
        self.usage = UsageAPI(client: httpClient)
        self.payments = PaymentAPI(client: httpClient)
        self.ratingPlans = RatingPlanAPI(client: httpClient)
        self.system = SystemAPI(client: httpClient)
        self.graphql = GraphQLAPI(client: httpClient)
    }
    
    // MARK: - Authentication Methods
    
    /// Generate a JWT token for authentication
    public func generateJWTToken(userID: String, expiryHours: Int, additionalClaims: [String: Any] = [:]) throws -> String {
        return try authProvider.generateJWTToken(userID: userID, expiryHours: expiryHours, additionalClaims: additionalClaims)
    }
    
    /// Validate a JWT token
    public func validateJWTToken(_ token: String) throws -> [String: Any] {
        return try authProvider.validateJWTToken(token)
    }
    
    // MARK: - Cleanup
    
    /// Close the SDK and cleanup resources
    public func close() {
        httpClient.close()
    }
}

// MARK: - Configuration

/// Configuration for Telecom SDK
public struct TelecomConfig {
    /// Base URL for the API server
    public var apiURL: String
    /// API key for authentication
    public var apiKey: String?
    /// JWT secret for token generation
    public var jwtSecret: String?
    /// Request timeout in seconds
    public var timeout: TimeInterval
    /// Maximum number of retry attempts
    public var maxRetries: Int
    /// Delay between retries in seconds
    public var retryDelay: TimeInterval
    
    public init(
        apiURL: String = "http://localhost:8000",
        apiKey: String? = nil,
        jwtSecret: String? = nil,
        timeout: TimeInterval = 30.0,
        maxRetries: Int = 3,
        retryDelay: TimeInterval = 1.0
    ) {
        self.apiURL = apiURL
        self.apiKey = apiKey
        self.jwtSecret = jwtSecret
        self.timeout = timeout
        self.maxRetries = maxRetries
        self.retryDelay = retryDelay
    }
}
