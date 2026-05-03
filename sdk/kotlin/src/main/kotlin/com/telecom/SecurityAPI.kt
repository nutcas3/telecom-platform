package com.telecom

import kotlinx.serialization.Serializable

/**
 * Security API for fraud detection and SIM swap protection
 */
class SecurityAPI(private val httpClient: HTTPClient) {
    
    /**
     * Analyze a transaction for fraud
     */
    suspend fun analyzeTransaction(transaction: Map<String, Any>): FraudAlert? {
        return httpClient.post("/api/v1/security/fraud/analyze", transaction)
    }
    
    /**
     * Get fraud alerts with filtering
     */
    suspend fun getFraudAlerts(filter: FraudAlertFilter? = null): List<FraudAlert> {
        val payload = filter?.let {
            mapOf(
                "type" to it.type,
                "severity" to it.severity,
                "status" to it.status,
                "limit" to it.limit
            ).filterValues { it != null }
        } ?: emptyMap()
        
        return httpClient.post("/api/v1/security/fraud/alerts", payload)
    }
    
    /**
     * Update fraud alert status
     */
    suspend fun updateAlertStatus(
        alertId: String,
        status: String,
        actions: List<String> = emptyList()
    ): JsonObject {
        return httpClient.put("/api/v1/security/fraud/alerts/$alertId",
            mapOf(
                "status" to status,
                "actions" to actions
            ))
    }
    
    /**
     * Get fraud detection metrics
     */
    suspend fun getFraudMetrics(period: String = "monthly"): FraudMetrics {
        return httpClient.get("/api/v1/security/fraud/metrics",
            mapOf("period" to period))
    }
    
    /**
     * Get detected fraud patterns
     */
    suspend fun getFraudPatterns(): JsonObject {
        return httpClient.get("/api/v1/security/fraud/patterns")
    }
    
    /**
     * Verify SIM swap request
     */
    suspend fun verifySIMSwap(profileId: String, msisdn: String): JsonObject {
        return httpClient.post("/api/v1/security/simswap/verify",
            mapOf(
                "profile_id" to profileId,
                "msisdn" to msisdn
            ))
    }
    
    /**
     * Get SIM swap history for a profile
     */
    suspend fun getSIMSwapHistory(profileId: String): JsonObject {
        return httpClient.get("/api/v1/security/simswap/history/$profileId")
    }
}

@Serializable
data class FraudAlert(
    val id: String,
    val type: String,
    val severity: String,
    val profileId: String,
    val description: String,
    val riskScore: Double,
    val evidence: List<String>,
    val ipAddress: String,
    val timestamp: String,
    val status: String,
    val actionsTaken: List<String>,
    val metadata: Map<String, Any>
)

@Serializable
data class FraudMetrics(
    val period: String,
    val totalAlerts: Long,
    val resolvedAlerts: Long,
    val falsePositives: Long,
    val resolutionRate: Double,
    val falsePositiveRate: Double,
    val byType: Map<String, Long>,
    val bySeverity: Map<String, Long>,
    val generatedAt: String
)

@Serializable
data class FraudAlertFilter(
    val type: String? = null,
    val severity: String? = null,
    val status: String? = null,
    val limit: Int = 50
)

@Serializable
enum class FraudType(val value: String) {
    ACCOUNT_TAKEOVER("account_takeover"),
    SUBSCRIPTION_FRAUD("subscription_fraud"),
    PAYMENT_FRAUD("payment_fraud"),
    USAGE_ANOMALY("usage_anomaly"),
    SIM_SWAP("sim_swap")
}

@Serializable
enum class FraudSeverity(val value: String) {
    LOW("low"),
    MEDIUM("medium"),
    HIGH("high"),
    CRITICAL("critical")
}
