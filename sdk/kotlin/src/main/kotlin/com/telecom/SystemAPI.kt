package com.telecom

/**
 * API for system management
 */
class SystemAPI(private val client: HTTPClient) {
    
    /**
     * Get system statistics
     */
    suspend fun getStats(): SystemStats {
        return client.get("/v1/system/stats")
    }
    
    /**
     * Get system health status
     */
    suspend fun getHealth(): HealthStatus {
        return client.get("/v1/health")
    }
}
