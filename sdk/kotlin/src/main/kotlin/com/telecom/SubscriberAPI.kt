package com.telecom

/**
 * API for subscriber management
 */
class SubscriberAPI(private val client: HTTPClient) {
    
    /**
     * Get subscriber by ID
     */
    suspend fun get(id: Long): Subscriber {
        return client.get("/v1/subscribers/$id")
    }
    
    /**
     * List subscribers with pagination
     */
    suspend fun list(page: Int = 1, pageSize: Int = 50, status: SubscriberStatus? = null): SubscriberList {
        val params = mutableMapOf(
            "page" to page.toString(),
            "page_size" to pageSize.toString()
        )
        
        status?.let { params["status"] = it.toString() }
        
        return client.get("/v1/subscribers", params)
    }
    
    /**
     * Create a new subscriber
     */
    suspend fun create(request: CreateSubscriberRequest): Subscriber {
        return client.post("/v1/subscribers", request)
    }
    
    /**
     * Update an existing subscriber
     */
    suspend fun update(id: Long, request: UpdateSubscriberRequest): Subscriber {
        return client.put("/v1/subscribers/$id", request)
    }
    
    /**
     * Delete a subscriber
     */
    suspend fun delete(id: Long) {
        client.delete("/v1/subscribers/$id")
    }
    
    /**
     * Suspend a subscriber
     */
    suspend fun suspend(id: Long): Subscriber {
        return client.post("/v1/subscribers/$id/suspend", null)
    }
    
    /**
     * Activate a suspended subscriber
     */
    suspend fun activate(id: Long): Subscriber {
        return client.post("/v1/subscribers/$id/activate", null)
    }
}
