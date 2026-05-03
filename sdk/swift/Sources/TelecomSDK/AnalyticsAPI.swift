import Foundation

/// Analytics API for churn prediction, market analysis, and pricing optimization
public class AnalyticsAPI {
    private let client: HTTPClient
    
    public init(client: HTTPClient) {
        self.client = client
    }
    
    // MARK: - Churn Analysis
    
    /// Predict churn risk for a profile
    public func predictChurn(profileId: String) async throws -> ChurnPrediction {
        return try await client.post(
            "/api/v1/analytics/churn/predict",
            body: ["profile_id": profileId],
            responseType: ChurnPrediction.self
        )
    }
    
    /// Get churn metrics
    public func getChurnMetrics(period: String = "monthly") async throws -> ChurnMetrics {
        return try await client.get(
            "/api/v1/analytics/churn/metrics",
            parameters: ["period": period],
            responseType: ChurnMetrics.self
        )
    }
    
    /// Get at-risk customers
    public func getAtRiskCustomers(
        riskLevel: ChurnRiskLevel,
        limit: Int = 100
    ) async throws -> [ChurnPrediction] {
        return try await client.post(
            "/api/v1/analytics/churn/at-risk",
            body: [
                "risk_level": riskLevel.rawValue,
                "limit": limit
            ],
            responseType: [ChurnPrediction].self
        )
    }
    
    // MARK: - Market Analytics
    
    /// Get market metrics
    public func getMarketMetrics(period: String = "monthly") async throws -> MarketMetrics {
        return try await client.get(
            "/api/v1/analytics/market/metrics",
            parameters: ["period": period],
            responseType: MarketMetrics.self
        )
    }
    
    /// Get competitor analysis
    public func getCompetitors() async throws -> [String: Any] {
        return try await client.get("/api/v1/analytics/market/competitors")
    }
    
    /// Get market opportunities
    public func getMarketOpportunities() async throws -> [String: Any] {
        return try await client.get("/api/v1/analytics/market/opportunities")
    }
    
    // MARK: - Predictive Maintenance
    
    /// Get maintenance metrics
    public func getMaintenanceMetrics(period: String = "monthly") async throws -> MaintenanceMetrics {
        return try await client.get(
            "/api/v1/analytics/maintenance/metrics",
            parameters: ["period": period],
            responseType: MaintenanceMetrics.self
        )
    }
    
    /// Get assets health
    public func getAssetsHealth() async throws -> [String: Any] {
        return try await client.get("/api/v1/analytics/maintenance/assets")
    }
    
    /// Get maintenance alerts
    public func getMaintenanceAlerts() async throws -> [String: Any] {
        return try await client.get("/api/v1/analytics/maintenance/alerts")
    }
    
    /// Predict failure for an asset
    public func predictFailure(assetId: String) async throws -> [String: Any] {
        return try await client.post("/api/v1/analytics/maintenance/predict/\(assetId)", body: [:])
    }
    
    // MARK: - Pricing Optimization
    
    /// Get pricing metrics
    public func getPricingMetrics(period: String = "monthly") async throws -> PricingMetrics {
        return try await client.get(
            "/api/v1/analytics/pricing/metrics",
            parameters: ["period": period],
            responseType: PricingMetrics.self
        )
    }
    
    /// Optimize pricing for rate plans
    public func optimizePricing(
        ratePlanIds: [String],
        strategy: String = "revenue_maximization"
    ) async throws -> [PricingOptimizationResult] {
        return try await client.post(
            "/api/v1/analytics/pricing/optimize",
            body: [
                "rate_plan_ids": ratePlanIds,
                "strategy": strategy
            ],
            responseType: [PricingOptimizationResult].self
        )
    }
    
    /// Get price elasticity data
    public func getPriceElasticity() async throws -> [String: Any] {
        return try await client.get("/api/v1/analytics/pricing/elasticity")
    }
}

// MARK: - Analytics Types

public struct ChurnPrediction: Codable {
    public let profileId: String
    public let riskLevel: String
    public let riskScore: Double
    public let predictedChurnDate: String?
    public let reasons: [String]
    public let recommendations: [String]
    public let lastUpdated: String
}

public struct ChurnMetrics: Codable {
    public let period: String
    public let totalSubscribers: Int64
    public let churnedSubscribers: Int64
    public let churnRate: Double
    public let monthlyChurnRate: Double
    public let annualChurnRate: Double
    public let averageTenureDays: Double
    public let riskDistribution: [String: Int64]
    public let generatedAt: String
}

public struct MarketMetrics: Codable {
    public let period: String
    public let totalMarketSize: Int64
    public let ourSubscribers: Int64
    public let marketShare: Double
    public let growthRate: Double
    public let byCountry: [String: [String: Any]]
    public let generatedAt: String
}

public struct MaintenanceMetrics: Codable {
    public let period: String
    public let totalAssets: Int64
    public let healthyAssets: Int64
    public let assetsNeedingAttention: Int64
    public let uptime: Double
    public let meanTimeToFailure: Double
    public let meanTimeToRepair: Double
    public let generatedAt: String
}

public struct PricingMetrics: Codable {
    public let period: String
    public let totalRevenue: Double
    public let arpu: Double
    public let priceElasticity: Double
    public let competitiveIndex: Double
    public let optimizationRoi: Double
    public let generatedAt: String
}

public struct PricingOptimizationResult: Codable {
    public let ratePlanId: String
    public let strategy: String
    public let currentPrice: Double
    public let optimalPrice: Double
    public let priceChangePct: Double
    public let expectedRevenue: Double
    public let expectedDemand: Double
    public let confidence: Double
    public let reasoning: [String]
    public let risks: [String]
    public let recommendations: [String]
}

public enum ChurnRiskLevel: String, CaseIterable {
    case low = "low"
    case medium = "medium"
    case high = "high"
    case critical = "critical"
}
