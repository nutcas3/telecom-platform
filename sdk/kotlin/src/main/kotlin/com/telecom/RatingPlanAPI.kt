package com.telecom

/**
 * API for rating plan management
 */
class RatingPlanAPI(private val client: HTTPClient) {
    
    /**
     * List all available rating plans
     */
    suspend fun list(): List<RatingPlan> {
        return client.get("/v1/rating-plans")
    }
    
    /**
     * Get a rating plan by ID
     */
    suspend fun get(planId: String): RatingPlan {
        return client.get("/v1/rating-plans/$planId")
    }
}
