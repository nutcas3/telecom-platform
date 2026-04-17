import Foundation
import NIO
import AsyncAlgorithms

/// Telecom Platform Swift SDK
/// 
/// Provides async/await support with Swift Concurrency for the Telecom Platform API.
public class TelecomSDK {
    
    // MARK: - Properties
    
    private let config: TelecomConfig
    private let httpClient: HTTPClient
    private let webSocketClient: WebSocketClient
    
    // MARK: - Initialization
    
    /// Initialize a new TelecomSDK instance
    /// - Parameter config: Configuration for the SDK
    public init(config: TelecomConfig = TelecomConfig()) {
        self.config = config
        self.httpClient = HTTPClient(config: config)
        self.webSocketClient = WebSocketClient(config: config)
    }
    
    // MARK: - Subscriber Management
    
    /// Get subscriber by ID
    /// - Parameter id: Subscriber ID
    /// - Returns: Subscriber information
    public func getSubscriber(id: Int64) async throws -> Subscriber {
        return try await httpClient.get(path: "/v1/subscribers/\(id)")
    }
    
    /// List subscribers with pagination
    /// - Parameters:
    ///   - page: Page number
    ///   - pageSize: Page size
    ///   - status: Optional status filter
    /// - Returns: Paginated list of subscribers
    public func listSubscribers(
        page: Int = 1,
        pageSize: Int = 50,
        status: SubscriberStatus? = nil
    ) async throws -> SubscriberList {
        var queryParams: [String: String] = [
            "page": "\(page)",
            "page_size": "\(pageSize)"
        ]
        
        if let status = status {
            queryParams["status"] = status.rawValue
        }
        
        return try await httpClient.get(path: "/v1/subscribers", queryParams: queryParams)
    }
    
    /// Create a new subscriber
    /// - Parameter request: Create subscriber request
    /// - Returns: Created subscriber
    public func createSubscriber(request: CreateSubscriberRequest) async throws -> Subscriber {
        return try await httpClient.post(path: "/v1/subscribers", body: request)
    }
    
    /// Update an existing subscriber
    /// - Parameters:
    ///   - id: Subscriber ID
    ///   - request: Update subscriber request
    /// - Returns: Updated subscriber
    public func updateSubscriber(id: Int64, request: UpdateSubscriberRequest) async throws -> Subscriber {
        return try await httpClient.put(path: "/v1/subscribers/\(id)", body: request)
    }
    
    /// Suspend a subscriber
    /// - Parameter id: Subscriber ID
    /// - Returns: Suspended subscriber
    public func suspendSubscriber(id: Int64) async throws -> Subscriber {
        return try await httpClient.post(path: "/v1/subscribers/\(id)/suspend")
    }
    
    /// Activate a suspended subscriber
    /// - Parameter id: Subscriber ID
    /// - Returns: Activated subscriber
    public func activateSubscriber(id: Int64) async throws -> Subscriber {
        return try await httpClient.post(path: "/v1/subscribers/\(id)/activate")
    }
    
    /// Terminate a subscriber
    /// - Parameter id: Subscriber ID
    /// - Returns: Success status
    public func terminateSubscriber(id: Int64) async throws -> Bool {
        return try await httpClient.delete(path: "/v1/subscribers/\(id)")
    }
    
    // MARK: - Usage Management
    
    /// Get usage statistics for a subscriber
    /// - Parameters:
    ///   - subscriberId: Subscriber ID
    ///   - startDate: Start date
    ///   - endDate: End date
    /// - Returns: Usage statistics
    public func getUsageStats(
        subscriberId: Int64,
        startDate: Date,
        endDate: Date
    ) async throws -> UsageStats {
        let formatter = ISO8601DateFormatter()
        formatter.timeZone = TimeZone.current
        
        let queryParams: [String: String] = [
            "start_date": formatter.string(from: startDate),
            "end_date": formatter.string(from: endDate)
        ]
        
        return try await httpClient.get(
            path: "/v1/subscribers/\(subscriberId)/usage",
            queryParams: queryParams
        )
    }
    
    /// List usage events with filtering
    /// - Parameters:
    ///   - subscriberId: Optional subscriber ID filter
    ///   - usageType: Optional usage type filter
    ///   - startDate: Optional start date filter
    ///   - endDate: Optional end date filter
    ///   - page: Page number
    ///   - pageSize: Page size
    /// - Returns: Paginated list of usage events
    public func listUsageEvents(
        subscriberId: Int64? = nil,
        usageType: UsageType? = nil,
        startDate: Date? = nil,
        endDate: Date? = nil,
        page: Int = 1,
        pageSize: Int = 50
    ) async throws -> UsageEventList {
        var queryParams: [String: String] = [
            "page": "\(page)",
            "page_size": "\(pageSize)"
        ]
        
        if let subscriberId = subscriberId {
            queryParams["subscriber_id"] = "\(subscriberId)"
        }
        
        if let usageType = usageType {
            queryParams["usage_type"] = usageType.rawValue
        }
        
        if let startDate = startDate {
            let formatter = ISO8601DateFormatter()
            formatter.timeZone = TimeZone.current
            queryParams["start_date"] = formatter.string(from: startDate)
        }
        
        if let endDate = endDate {
            let formatter = ISO8601DateFormatter()
            formatter.timeZone = TimeZone.current
            queryParams["end_date"] = formatter.string(from: endDate)
        }
        
        return try await httpClient.get(path: "/v1/usage/events", queryParams: queryParams)
    }
    
    /// Get real-time usage for a subscriber
    /// - Parameter subscriberId: Subscriber ID
    /// - Returns: Real-time usage data
    public func getRealTimeUsage(subscriberId: Int64) async throws -> RealTimeUsage {
        return try await httpClient.get(path: "/v1/subscribers/\(subscriberId)/realtime")
    }
    
    // MARK: - Payment Management
    
    /// Create a payment transaction
    /// - Parameter request: Create payment request
    /// - Returns: Created payment transaction
    public func createPaymentTransaction(request: CreatePaymentRequest) async throws -> PaymentTransaction {
        return try await httpClient.post(path: "/v1/payments/transactions", body: request)
    }
    
    /// Get payment transaction by ID
    /// - Parameter transactionId: Transaction ID
    /// - Returns: Payment transaction
    public func getPaymentTransaction(transactionId: String) async throws -> PaymentTransaction {
        return try await httpClient.get(path: "/v1/payments/transactions/\(transactionId)")
    }
    
    /// List payment transactions with filtering
    /// - Parameters:
    ///   - subscriberId: Optional subscriber ID filter
    ///   - status: Optional status filter
    ///   - page: Page number
    ///   - pageSize: Page size
    /// - Returns: Paginated list of payment transactions
    public func listPaymentTransactions(
        subscriberId: Int64? = nil,
        status: PaymentStatus? = nil,
        page: Int = 1,
        pageSize: Int = 50
    ) async throws -> PaymentTransactionList {
        var queryParams: [String: String] = [
            "page": "\(page)",
            "page_size": "\(pageSize)"
        ]
        
        if let subscriberId = subscriberId {
            queryParams["subscriber_id"] = "\(subscriberId)"
        }
        
        if let status = status {
            queryParams["status"] = status.rawValue
        }
        
        return try await httpClient.get(path: "/v1/payments/transactions", queryParams: queryParams)
    }
    
    // MARK: - Rating Plans
    
    /// List all available rating plans
    /// - Returns: List of rating plans
    public func listRatingPlans() async throws -> [RatingPlan] {
        return try await httpClient.get(path: "/v1/rating-plans")
    }
    
    /// Get rating plan by ID
    /// - Parameter planId: Plan ID
    /// - Returns: Rating plan
    public func getRatingPlan(planId: String) async throws -> RatingPlan {
        return try await httpClient.get(path: "/v1/rating-plans/\(planId)")
    }
    
    // MARK: - System Management
    
    /// Get system statistics
    /// - Returns: System statistics
    public func getSystemStats() async throws -> SystemStats {
        return try await httpClient.get(path: "/v1/system/stats")
    }
    
    /// Get system health status
    /// - Returns: Health status
    public func getHealthStatus() async throws -> HealthStatus {
        return try await httpClient.get(path: "/v1/health")
    }
    
    // MARK: - WebSocket Support
    
    /// Connect to WebSocket for real-time updates
    /// - Parameter messageHandler: Closure to handle WebSocket messages
    /// - Returns: Async stream of WebSocket messages
    public func connectWebSocket(
        messageHandler: @escaping (WebSocketMessage) async -> Void
    ) -> AsyncThrowingStream<WebSocketMessage, Error> {
        return try await webSocketClient.connect(messageHandler: messageHandler)
    }
    
    // MARK: - GraphQL Support
    
    /// Execute a GraphQL query
    /// - Parameters:
    ///   - query: GraphQL query string
    ///   - variables: Optional query variables
    /// - Returns: GraphQL response
    public func executeGraphQLQuery(
        query: String,
        variables: [String: Any]? = nil
    ) async throws -> GraphQLResponse {
        let request = GraphQLRequest(query: query, variables: variables ?? [:])
        return try await httpClient.post(path: "/graphql", body: request)
    }
}

// MARK: - Configuration

/// Configuration for Telecom SDK
public struct TelecomConfig {
    /// Base URL for the API server
    public var apiURL: String
    /// API key for authentication
    public var apiKey: String?
    /// Request timeout in seconds
    public var timeout: TimeInterval
    /// Maximum number of retry attempts
    public var maxRetries: Int
    
    public init(
        apiURL: String = "http://localhost:8000",
        apiKey: String? = nil,
        timeout: TimeInterval = 30.0,
        maxRetries: Int = 3
    ) {
        self.apiURL = apiURL
        self.apiKey = apiKey
        self.timeout = timeout
        self.maxRetries = maxRetries
    }
}
