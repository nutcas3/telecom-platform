package com.telecom

/**
 * API for payment management
 */
class PaymentAPI(private val client: HTTPClient) {
    
    /**
     * Create a payment transaction
     */
    suspend fun createTransaction(request: CreatePaymentRequest): PaymentTransaction {
        return client.post("/v1/payments/transactions", request)
    }
    
    /**
     * Get a payment transaction by ID
     */
    suspend fun getTransaction(transactionId: String): PaymentTransaction {
        return client.get("/v1/payments/transactions/$transactionId")
    }
    
    /**
     * List payment transactions
     */
    suspend fun listTransactions(
        subscriberId: Long? = null,
        status: PaymentStatus? = null,
        page: Int = 1,
        pageSize: Int = 50
    ): SubscriberList {
        val params = mutableMapOf(
            "page" to page.toString(),
            "page_size" to pageSize.toString()
        )
        
        subscriberId?.let { params["subscriber_id"] = it.toString() }
        status?.let { params["status"] = it.toString() }
        
        return client.get("/v1/payments/transactions", params)
    }
}
