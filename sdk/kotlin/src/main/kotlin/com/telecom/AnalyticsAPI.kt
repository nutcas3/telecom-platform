package com.telecom

import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonObject

/**
 * Analytics API for churn prediction, market analysis, and pricing optimization
 */
class AnalyticsAPI(private val httpClient: HTTPClient) {
    
    /**
     * Predict churn risk for a profile
     */
    suspend fun predictChurn(profileId: String): ChurnPrediction {
        return httpClient.post("/api/v1/analytics/churn/predict", 
            mapOf("profile_id" to profileId))
    }
    
    /**
     * Get churn metrics
     */
    suspend fun getChurnMetrics(period: String = "monthly"): ChurnMetrics {
        return httpClient.get("/api/v1/analytics/churn/metrics", 
            mapOf("period" to period))
    }
    
    /**
     * Get at-risk customers
     */
    suspend fun getAtRiskCustomers(
        riskLevel: ChurnRiskLevel, 
        limit: Int = 100
    ): List<ChurnPrediction> {
        return httpClient.post("/api/v1/analytics/churn/at-risk",
            mapOf(
                "risk_level" to riskLevel.value,
                "limit" to limit
            ))
    }
    
    /**
     * Get market metrics
     */
    suspend fun getMarketMetrics(period: String = "monthly"): MarketMetrics {
        return httpClient.get("/api/v1/analytics/market/metrics",
            mapOf("period" to period))
    }
    
    /**
     * Get competitor analysis
     */
    suspend fun getCompetitors(): JsonObject {
        return httpClient.get("/api/v1/analytics/market/competitors")
    }
    
    /**
     * Get market opportunities
     */
    suspend fun getMarketOpportunities(): JsonObject {
        return httpClient.get("/api/v1/analytics/market/opportunities")
    }
    
    /**
     * Get maintenance metrics
     */
    suspend fun getMaintenanceMetrics(period: String = "monthly"): MaintenanceMetrics {
        return httpClient.get("/api/v1/analytics/maintenance/metrics",
            mapOf("period" to period))
    }
    
    /**
     * Get assets health
     */
    suspend fun getAssetsHealth(): JsonObject {
        return httpClient.get("/api/v1/analytics/maintenance/assets")
    }
    
    /**
     * Get maintenance alerts
     */
    suspend fun getMaintenanceAlerts(): JsonObject {
        return httpClient.get("/api/v1/analytics/maintenance/alerts")
    }
    
    /**
     * Predict failure for an asset
     */
    suspend fun predictFailure(assetId: String): JsonObject {
        return httpClient.post("/api/v1/analytics/maintenance/predict/$assetId", emptyMap())
    }
    
    /**
     * Get pricing metrics
     */
    suspend fun getPricingMetrics(period: String = "monthly"): PricingMetrics {
        return httpClient.get("/api/v1/analytics/pricing/metrics",
            mapOf("period" to period))
    }
    
    /**
     * Optimize pricing for rate plans
     */
    suspend fun optimizePricing(
        ratePlanIds: List<String>,
        strategy: String = "revenue_maximization"
    ): List<PricingOptimizationResult> {
        return httpClient.post("/api/v1/analytics/pricing/optimize",
            mapOf(
                "rate_plan_ids" to ratePlanIds,
                "strategy" to strategy
            ))
    }
    
    /**
     * Get price elasticity data
     */
    suspend fun getPriceElasticity(): JsonObject {
        return httpClient.get("/api/v1/analytics/pricing/elasticity")
    }
}

@Serializable
data class ChurnPrediction(
    val profileId: String,
    val riskLevel: String,
    val riskScore: Double,
    val predictedChurnDate: String? = null,
    val reasons: List<String>,
    val recommendations: List<String>,
    val lastUpdated: String
)

@Serializable
data class ChurnMetrics(
    val period: String,
    val totalSubscribers: Long,
    val churnedSubscribers: Long,
    val churnRate: Double,
    val monthlyChurnRate: Double,
    val annualChurnRate: Double,
    val averageTenureDays: Double,
    val riskDistribution: Map<String, Long>,
    val generatedAt: String
)

@Serializable
data class MarketMetrics(
    val period: String,
    val totalMarketSize: Long,
    val ourSubscribers: Long,
    val marketShare: Double,
    val growthRate: Double,
    val byCountry: Map<String, JsonObject>,
    val generatedAt: String
)

@Serializable
data class MaintenanceMetrics(
    val period: String,
    val totalAssets: Long,
    val healthyAssets: Long,
    val assetsNeedingAttention: Long,
    val uptime: Double,
    val meanTimeToFailure: Double,
    val meanTimeToRepair: Double,
    val generatedAt: String
)

@Serializable
data class PricingMetrics(
    val period: String,
    val totalRevenue: Double,
    val arpu: Double,
    val priceElasticity: Double,
    val competitiveIndex: Double,
    val optimizationRoi: Double,
    val generatedAt: String
)

@Serializable
data class PricingOptimizationResult(
    val ratePlanId: String,
    val strategy: String,
    val currentPrice: Double,
    val optimalPrice: Double,
    val priceChangePct: Double,
    val expectedRevenue: Double,
    val expectedDemand: Double,
    val confidence: Double,
    val reasoning: List<String>,
    val risks: List<String>,
    val recommendations: List<String>
)

@Serializable
enum class ChurnRiskLevel(val value: String) {
    LOW("low"),
    MEDIUM("medium"),
    HIGH("high"),
    CRITICAL("critical")
}
