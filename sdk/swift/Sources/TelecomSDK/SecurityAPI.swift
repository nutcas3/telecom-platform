import Foundation

/// Security API for fraud detection and SIM swap protection
public class SecurityAPI {
    private let client: HTTPClient
    
    public init(client: HTTPClient) {
        self.client = client
    }
    
    /// Analyze a transaction for fraud
    public func analyzeTransaction(_ transaction: [String: Any]) async throws -> FraudAlert? {
        do {
            return try await client.post(
                "/api/v1/security/fraud/analyze",
                body: transaction,
                responseType: FraudAlert.self
            )
        } catch {
            if case HTTPClient.APIError.notFound = error {
                return nil
            }
            throw error
        }
    }
    
    /// Get fraud alerts with filtering
    public func getFraudAlerts(filter: FraudAlertFilter? = nil) async throws -> [FraudAlert] {
        var body: [String: Any] = [:]
        if let filter = filter {
            if let type = filter.type { body["type"] = type }
            if let severity = filter.severity { body["severity"] = severity }
            if let status = filter.status { body["status"] = status }
            body["limit"] = filter.limit
        }
        
        return try await client.post(
            "/api/v1/security/fraud/alerts",
            body: body,
            responseType: [FraudAlert].self
        )
    }
    
    /// Update fraud alert status
    public func updateAlertStatus(
        alertId: String,
        status: String,
        actions: [String] = []
    ) async throws -> [String: Any] {
        return try await client.put(
            "/api/v1/security/fraud/alerts/\(alertId)",
            body: [
                "status": status,
                "actions": actions
            ]
        )
    }
    
    /// Get fraud detection metrics
    public func getFraudMetrics(period: String = "monthly") async throws -> FraudMetrics {
        return try await client.get(
            "/api/v1/security/fraud/metrics",
            parameters: ["period": period],
            responseType: FraudMetrics.self
        )
    }
    
    /// Get detected fraud patterns
    public func getFraudPatterns() async throws -> [String: Any] {
        return try await client.get("/api/v1/security/fraud/patterns")
    }
    
    /// Verify SIM swap request
    public func verifySIMSwap(profileId: String, msisdn: String) async throws -> [String: Any] {
        return try await client.post(
            "/api/v1/security/simswap/verify",
            body: [
                "profile_id": profileId,
                "msisdn": msisdn
            ]
        )
    }
    
    /// Get SIM swap history for a profile
    public func getSIMSwapHistory(profileId: String) async throws -> [String: Any] {
        return try await client.get("/api/v1/security/simswap/history/\(profileId)")
    }
}

public struct FraudAlert: Codable {
    public let id: String
    public let type: String
    public let severity: String
    public let profileId: String
    public let description: String
    public let riskScore: Double
    public let evidence: [String]
    public let ipAddress: String
    public let timestamp: String
    public let status: String
    public let actionsTaken: [String]
    public let metadata: [String: Any]
    
    enum CodingKeys: String, CodingKey {
        case id, type, severity, profileId, description, riskScore
        case evidence, ipAddress, timestamp, status, actionsTaken, metadata
    }
    
    public init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        id = try container.decode(String.self, forKey: .id)
        type = try container.decode(String.self, forKey: .type)
        severity = try container.decode(String.self, forKey: .severity)
        profileId = try container.decode(String.self, forKey: .profileId)
        description = try container.decode(String.self, forKey: .description)
        riskScore = try container.decode(Double.self, forKey: .riskScore)
        evidence = try container.decode([String].self, forKey: .evidence)
        ipAddress = try container.decode(String.self, forKey: .ipAddress)
        timestamp = try container.decode(String.self, forKey: .timestamp)
        status = try container.decode(String.self, forKey: .status)
        actionsTaken = try container.decode([String].self, forKey: .actionsTaken)
        metadata = try container.decode([String: Any].self, forKey: .metadata)
    }
    
    public func encode(to encoder: Encoder) throws {
        var container = encoder.container(keyedBy: CodingKeys.self)
        try container.encode(id, forKey: .id)
        try container.encode(type, forKey: .type)
        try container.encode(severity, forKey: .severity)
        try container.encode(profileId, forKey: .profileId)
        try container.encode(description, forKey: .description)
        try container.encode(riskScore, forKey: .riskScore)
        try container.encode(evidence, forKey: .evidence)
        try container.encode(ipAddress, forKey: .ipAddress)
        try container.encode(timestamp, forKey: .timestamp)
        try container.encode(status, forKey: .status)
        try container.encode(actionsTaken, forKey: .actionsTaken)
        try container.encode(metadata, forKey: .metadata)
    }
}

public struct FraudMetrics: Codable {
    public let period: String
    public let totalAlerts: Int64
    public let resolvedAlerts: Int64
    public let falsePositives: Int64
    public let resolutionRate: Double
    public let falsePositiveRate: Double
    public let byType: [String: Int64]
    public let bySeverity: [String: Int64]
    public let generatedAt: String
}

public struct FraudAlertFilter {
    public let type: String?
    public let severity: String?
    public let status: String?
    public let limit: Int
    
    public init(type: String? = nil, severity: String? = nil, status: String? = nil, limit: Int = 50) {
        self.type = type
        self.severity = severity
        self.status = status
        self.limit = limit
    }
}

public enum FraudType: String, CaseIterable {
    case accountTakeover = "account_takeover"
    case subscriptionFraud = "subscription_fraud"
    case paymentFraud = "payment_fraud"
    case usageAnomaly = "usage_anomaly"
    case simSwap = "sim_swap"
}

public enum FraudSeverity: String, CaseIterable {
    case low = "low"
    case medium = "medium"
    case high = "high"
    case critical = "critical"
}
