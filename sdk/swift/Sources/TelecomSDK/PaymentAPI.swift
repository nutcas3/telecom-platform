import Foundation

/// API for payment management
public class PaymentAPI {
    private let client: HTTPClient
    
    public init(client: HTTPClient) {
        self.client = client
    }
    
    /// Create a payment transaction
    public func createTransaction(_ request: CreatePaymentRequest) async throws -> PaymentTransaction {
        return try await client.post(path: "/v1/payments/transactions", body: request)
    }
    
    /// Get a payment transaction by ID
    public func getTransaction(transactionID: String) async throws -> PaymentTransaction {
        return try await client.get(path: "/v1/payments/transactions/\(transactionID)")
    }
    
    /// List payment transactions
    public func listTransactions(
        subscriberID: Int64? = nil,
        status: PaymentStatus? = nil,
        page: Int = 1,
        pageSize: Int = 50
    ) async throws -> SubscriberList {
        var params: [String: String] = [
            "page": "\(page)",
            "page_size": "\(pageSize)"
        ]
        
        if let subscriberID = subscriberID {
            params["subscriber_id"] = "\(subscriberID)"
        }
        if let status = status {
            params["status"] = "\(status)"
        }
        
        return try await client.get(path: "/v1/payments/transactions", params: params)
    }
}
