package com.telecom

import io.ktor.client.*
import io.ktor.client.call.*
import io.ktor.client.plugins.*
import io.ktor.client.plugins.contentnegotiation.*
import io.ktor.client.request.*
import io.ktor.http.*
import io.ktor.serialization.kotlinx.json.*
import kotlinx.serialization.json.Json

/**
 * HTTP client for making API requests
 */
class HTTPClient(
    private val config: TelecomConfig,
    private val authProvider: AuthProvider
) {
    private val client: HttpClient = HttpClient {
        defaultRequest {
            url(config.apiURL)
            header("User-Agent", "Telecom-Kotlin-SDK/1.0.0")
        }
        
        install(ContentNegotiation) {
            json(Json {
                ignoreUnknownKeys = true
                isLenient = true
            })
        }
        
        install(HttpTimeout) {
            requestTimeoutMillis = config.timeout.toMillis()
        }
    }
    
    /**
     * Make an HTTP GET request
     */
    suspend fun <T> get(path: String, params: Map<String, String>? = null): T {
        return request<T>("GET", path, null, params)
    }
    
    /**
     * Make an HTTP POST request
     */
    suspend fun <T> post(path: String, body: Any? = null): T {
        return request<T>("POST", path, body, null)
    }
    
    /**
     * Make an HTTP PUT request
     */
    suspend fun <T> put(path: String, body: Any? = null): T {
        return request<T>("PUT", path, body, null)
    }
    
    /**
     * Make an HTTP DELETE request
     */
    suspend fun delete(path: String) {
        request<Unit>("DELETE", path, null, null)
    }
    
    private suspend fun <T> request(
        method: String,
        path: String,
        body: Any?,
        params: Map<String, String>?
    ): T {
        var lastError: Throwable? = null
        
        for (attempt in 0..config.maxRetries) {
            try {
                val response = client.request(path) {
                    this.method = HttpMethod.parse(method)
                    
                    // Set headers
                    val headers = authProvider.getHeaders()
                    for ((key, value) in headers) {
                        header(key, value)
                    }
                    
                    // Add query parameters
                    params?.let {
                        it.forEach { (k, v) ->
                            parameter(k, v)
                        }
                    }
                    
                    // Add body
                    body?.let {
                        setBody(it)
                        contentType(ContentType.Application.Json)
                    }
                }
                
                handleResponseErrors(response.status)
                
                return response.body()
            } catch (e: Exception) {
                lastError = e
                if (attempt < config.maxRetries) {
                    kotlinx.coroutines.delay(config.retryDelay * (1 shl attempt))
                }
            }
        }
        
        throw lastError ?: NetworkError.RequestFailed
    }
    
    private fun handleResponseErrors(status: HttpStatusCode) {
        when (status) {
            HttpStatusCode.Unauthorized -> throw AuthError.AuthenticationFailed
            HttpStatusCode.TooManyRequests -> throw AuthError.RateLimitExceeded
            in 400..499 -> throw NetworkError.ClientError(status.value)
            in 500..599 -> throw NetworkError.ServerError(status.value)
            else -> {}
        }
    }
    
    fun close() {
        client.close()
    }
}

sealed class NetworkError(message: String) : Exception(message) {
    object RequestFailed : NetworkError("Request failed")
    class ClientError(code: Int) : NetworkError("Client error: $code")
    class ServerError(code: Int) : NetworkError("Server error: $code")
}

// Extend AuthError with additional errors
object AuthError {
    object AuthenticationFailed : AuthError("Authentication failed")
    object RateLimitExceeded : AuthError("Rate limit exceeded")
}
