package com.telecom

import kotlinx.serialization.Serializable

/**
 * Currency and Billing API
 */
class CurrencyAPI(private val httpClient: HTTPClient) {
    
    /**
     * Convert currency
     */
    suspend fun convert(from: String, to: String, amount: Double): ConvertResponse {
        return httpClient.post("/api/v1/currency/convert",
            mapOf(
                "from" to from,
                "to" to to,
                "amount" to amount
            ))
    }
    
    /**
     * Get exchange rate between currencies
     */
    suspend fun getExchangeRate(from: String, to: String): ExchangeRate {
        return httpClient.get("/api/v1/currency/exchange/$from/$to")
    }
    
    /**
     * Get exchange rate history
     */
    suspend fun getExchangeRateHistory(from: String, to: String, days: Int = 30): List<ExchangeRate> {
        return httpClient.get("/api/v1/currency/exchange/$from/$to/history",
            mapOf("days" to days))
    }
    
    /**
     * Get supported currencies
     */
    suspend fun getSupportedCurrencies(): List<Currency> {
        return httpClient.get("/api/v1/currency/currencies")
    }
    
    /**
     * Refresh exchange rates
     */
    suspend fun refreshExchangeRates(): JsonObject {
        return httpClient.post("/api/v1/currency/exchange/refresh", emptyMap())
    }
    
    /**
     * Process billing transaction
     */
    suspend fun processBilling(billingData: Map<String, Any>): BillingTransaction {
        return httpClient.post("/api/v1/currency/billing", billingData)
    }
    
    /**
     * Get billing history for a profile
     */
    suspend fun getBillingHistory(profileId: String, limit: Int = 50): List<BillingTransaction> {
        return httpClient.get("/api/v1/currency/billing/history/$profileId",
            mapOf("limit" to limit))
    }
    
    /**
     * Get billing summary for a profile
     */
    suspend fun getBillingSummary(profileId: String, period: String = "monthly"): BillingSummary {
        return httpClient.get("/api/v1/currency/billing/summary/$profileId",
            mapOf("period" to period))
    }
    
    /**
     * Process refund
     */
    suspend fun processRefund(transactionId: String, reason: String): BillingTransaction {
        return httpClient.post("/api/v1/currency/billing/refund/$transactionId",
            mapOf("reason" to reason))
    }
    
    /**
     * Get billing analytics
     */
    suspend fun getBillingAnalytics(period: String = "monthly"): JsonObject {
        return httpClient.get("/api/v1/currency/billing/analytics",
            mapOf("period" to period))
    }
}

@Serializable
data class ConvertRequest(
    val from: String,
    val to: String,
    val amount: Double
)

@Serializable
data class ConvertResponse(
    val from: String,
    val to: String,
    val amount: Double,
    val converted: Double,
    val rate: Double,
    val timestamp: String
)

@Serializable
data class ExchangeRate(
    val from: String,
    val to: String,
    val rate: Double,
    val timestamp: String
)

@Serializable
data class Currency(
    val code: String,
    val name: String,
    val symbol: String
)

@Serializable
data class BillingTransaction(
    val id: String,
    val profileId: String,
    val amount: Double,
    val currency: String,
    val type: String,
    val status: String,
    val description: String,
    val createdAt: String
)

@Serializable
data class BillingSummary(
    val profileId: String,
    val period: String,
    val totalAmount: Double,
    val currency: String,
    val transactionCount: Int,
    val breakdown: Map<String, Double>
)
