package com.telecom

import java.time.format.DateTimeFormatter

/**
 * API for usage management
 */
class UsageAPI(private val client: HTTPClient) {
    
    private val formatter = DateTimeFormatter.ISO_DATE_TIME
    
    /**
     * Get usage statistics for a subscriber
     */
    suspend fun getStats(subscriberId: Long, startDate: java.time.LocalDateTime, endDate: java.time.LocalDateTime): UsageStats {
        val params = mapOf(
            "start_date" to formatter.format(startDate),
            "end_date" to formatter.format(endDate)
        )
        
        return client.get("/v1/subscribers/$subscriberId/usage", params)
    }
    
    /**
     * List usage events
     */
    suspend fun listEvents(
        subscriberId: Long? = null,
        usageType: UsageType? = null,
        startDate: java.time.LocalDateTime? = null,
        endDate: java.time.LocalDateTime? = null,
        page: Int = 1,
        pageSize: Int = 50
    ): UsageEventList {
        val params = mutableMapOf(
            "page" to page.toString(),
            "page_size" to pageSize.toString()
        )
        
        subscriberId?.let { params["subscriber_id"] = it.toString() }
        usageType?.let { params["usage_type"] = it.toString() }
        startDate?.let { params["start_date"] = formatter.format(it) }
        endDate?.let { params["end_date"] = formatter.format(it) }
        
        return client.get("/v1/usage/events", params)
    }
    
    /**
     * Get real-time usage for a subscriber
     */
    suspend fun getRealtime(subscriberId: Long): RealTimeUsage {
        return client.get("/v1/subscribers/$subscriberId/realtime")
    }
}
