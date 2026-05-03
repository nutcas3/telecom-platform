import Foundation

/// Currency and Billing API
public class CurrencyAPI {
    private let client: HTTPClient
    
    public init(client: HTTPClient) {
        self.client = client
    }
    
    // MARK: - Currency Conversion
    
    /// Convert currency
    public func convert(from: String, to: String, amount: Double) async throws -> ConvertResponse {
        return try await client.post(
            "/api/v1/currency/convert",
            body: [
                "from": from,
                "to": to,
                "amount": amount
            ],
            responseType: ConvertResponse.self
        )
    }
    
    /// Get exchange rate between currencies
    public func getExchangeRate(from: String, to: String) async throws -> ExchangeRate {
        return try await client.get(
            "/api/v1/currency/exchange/\(from)/\(to)",
            responseType: ExchangeRate.self
        )
    }
    
    /// Get exchange rate history
    public func getExchangeRateHistory(from: String, to: String, days: Int = 30) async throws -> [ExchangeRate] {
        return try await client.get(
            "/api/v1/currency/exchange/\(from)/\(to)/history",
            parameters: ["days": days],
            responseType: [ExchangeRate].self
        )
    }
    
    /// Get supported currencies
    public func getSupportedCurrencies() async throws -> [Currency] {
        return try await client.get(
            "/api/v1/currency/currencies",
            responseType: [Currency].self
        )
    }
    
    /// Refresh exchange rates
    public func refreshExchangeRates() async throws -> [String: Any] {
        return try await client.post("/api/v1/currency/exchange/refresh", body: [:])
    }
    
    // MARK: - Billing
    
    /// Process billing transaction
    public func processBilling(_ billingData: [String: Any]) async throws -> BillingTransaction {
        return try await client.post(
            "/api/v1/currency/billing",
            body: billingData,
            responseType: BillingTransaction.self
        )
    }
    
    /// Get billing history for a profile
    public func getBillingHistory(profileId: String, limit: Int = 50) async throws -> [BillingTransaction] {
        return try await client.get(
            "/api/v1/currency/billing/history/\(profileId)",
            parameters: ["limit": limit],
            responseType: [BillingTransaction].self
        )
    }
    
    /// Get billing summary for a profile
    public func getBillingSummary(profileId: String, period: String = "monthly") async throws -> BillingSummary {
        return try await client.get(
            "/api/v1/currency/billing/summary/\(profileId)",
            parameters: ["period": period],
            responseType: BillingSummary.self
        )
    }
    
    /// Process refund
    public func processRefund(transactionId: String, reason: String) async throws -> BillingTransaction {
        return try await client.post(
            "/api/v1/currency/billing/refund/\(transactionId)",
            body: ["reason": reason],
            responseType: BillingTransaction.self
        )
    }
    
    /// Get billing analytics
    public func getBillingAnalytics(period: String = "monthly") async throws -> [String: Any] {
        return try await client.get(
            "/api/v1/currency/billing/analytics",
            parameters: ["period": period]
        )
    }
}

// MARK: - Currency Types

public struct ConvertRequest: Codable {
    public let from: String
    public let to: String
    public let amount: Double
}

public struct ConvertResponse: Codable {
    public let from: String
    public let to: String
    public let amount: Double
    public let converted: Double
    public let rate: Double
    public let timestamp: String
}

public struct ExchangeRate: Codable {
    public let from: String
    public let to: String
    public let rate: Double
    public let timestamp: String
}

public struct Currency: Codable {
    public let code: String
    public let name: String
    public let symbol: String
}

public struct BillingTransaction: Codable {
    public let id: String
    public let profileId: String
    public let amount: Double
    public let currency: String
    public let type: String
    public let status: String
    public let description: String
    public let createdAt: String
}

public struct BillingSummary: Codable {
    public let profileId: String
    public let period: String
    public let totalAmount: Double
    public let currency: String
    public let transactionCount: Int
    public let breakdown: [String: Double]
}
