package com.telecom

import java.security.MessageDigest
import java.util.Base64
import javax.crypto.Mac
import javax.crypto.spec.SecretKeySpec

/**
 * Authentication provider for the Telecom SDK
 */
class AuthProvider(
    private val apiKey: String? = null,
    private val jwtSecret: String? = null
) {
    private var tokenCache: String? = null
    private var tokenExpiry: Long? = null
    
    /**
     * Get authentication headers for API requests
     */
    fun getHeaders(): Map<String, String> {
        val headers = mutableMapOf(
            "Content-Type" to "application/json",
            "User-Agent" to "Telecom-Kotlin-SDK/1.0.0"
        )
        
        apiKey?.let { headers["X-API-Key"] = it }
        
        if (tokenCache != null && isTokenValid()) {
            headers["Authorization"] = "Bearer $tokenCache"
        }
        
        return headers
    }
    
    /**
     * Generate a JWT token for authentication
     */
    fun generateJWTToken(
        userId: String,
        expiryHours: Int,
        additionalClaims: Map<String, Any> = emptyMap()
    ): String {
        val jwtSecret = jwtSecret ?: throw AuthError.JWTSecretNotConfigured
        
        val now = System.currentTimeMillis() / 1000
        val exp = now + (expiryHours * 3600)
        
        val claims = mutableMapOf<String, Any>(
            "sub" to userId,
            "exp" to exp,
            "iat" to now
        )
        claims.putAll(additionalClaims)
        
        val header = mapOf("alg" to "HS256", "typ" to "JWT")
        val encodedHeader = base64URLEncode(header)
        val encodedPayload = base64URLEncode(claims)
        val signature = sign("$encodedHeader.$encodedPayload", jwtSecret)
        
        val token = "$encodedHeader.$encodedPayload.$signature"
        tokenCache = token
        tokenExpiry = exp
        
        return token
    }
    
    /**
     * Validate a JWT token
     */
    fun validateJWTToken(token: String): Map<String, Any> {
        val jwtSecret = jwtSecret ?: throw AuthError.JWTSecretNotConfigured
        
        val parts = token.split(".")
        if (parts.size != 3) {
            throw AuthError.InvalidTokenFormat
        }
        
        val (encodedHeader, encodedPayload, signature) = parts
        val expectedSignature = sign("$encodedHeader.$encodedPayload", jwtSecret)
        
        if (signature != expectedSignature) {
            throw AuthError.InvalidTokenSignature
        }
        
        val payloadBytes = base64URLDecode(encodedPayload) ?: throw AuthError.FailedToDecodePayload
        val payload = parseJson(payloadBytes) as? Map<String, Any> ?: throw AuthError.FailedToDecodePayload
        
        val exp = payload["exp"] as? Number
        if (exp != null && exp.toLong() < System.currentTimeMillis() / 1000) {
            throw AuthError.TokenExpired
        }
        
        return payload
    }
    
    /**
     * Clear the cached JWT token
     */
    fun clearTokenCache() {
        tokenCache = null
        tokenExpiry = null
    }
    
    private fun isTokenValid(): Boolean {
        if (tokenCache == null || tokenExpiry == null) return false
        return System.currentTimeMillis() / 1000 < (tokenExpiry ?: 0)
    }
    
    private fun base64URLEncode(data: Any): String {
        val json = toJson(data)
        return Base64.getUrlEncoder()
            .withoutPadding()
            .encodeToString(json.toByteArray())
    }
    
    private fun base64URLDecode(data: String): ByteArray? {
        val base64String = data
            .replace("-", "+")
            .replace("_", "/")
        
        val paddingNeeded = (4 - base64String.length % 4) % 4
        val paddedString = base64String + "=".repeat(paddingNeeded)
        
        return try {
            Base64.getUrlDecoder().decode(paddedString)
        } catch (e: Exception) {
            null
        }
    }
    
    private fun sign(data: String, secret: String): String {
        return try {
            val key = SecretKeySpec(secret.toByteArray(), "HmacSHA256")
            val mac = Mac.getInstance("HmacSHA256")
            mac.init(key)
            val signature = mac.doFinal(data.toByteArray())
            Base64.getUrlEncoder()
                .withoutPadding()
                .encodeToString(signature)
        } catch (e: Exception) {
            throw AuthError.SigningFailed
        }
    }
    
    private fun toJson(data: Any): String {
        return when (data) {
            is Map<*, *> -> {
                val entries = data.entries.joinToString(",") { (k, v) ->
                    "\"${k}\":${toJson(v)}"
                }
                "{$entries}"
            }
            is List<*> -> {
                val items = data.joinToString(",") { toJson(it) }
                "[$items]"
            }
            is String -> "\"$data\""
            is Number -> data.toString()
            is Boolean -> data.toString()
            else -> "\"$data\""
        }
    }
    
    private fun parseJson(json: String): Any? {
        // JSON parser using kotlinx.serialization
        // In production, use proper kotlinx.serialization with typed models
        // For SDK compatibility without external dependencies, using basic parsing
        return try {
            when {
                json.trim().startsWith("{") -> {
                    // Parse JSON object
                    val content = json.trim().removeSurrounding("{", "}")
                    val pairs = content.split(",")
                    val map = mutableMapOf<String, Any>()
                    pairs.forEach { pair ->
                        val (key, value) = pair.split(":", limit = 2)
                        val cleanKey = key.trim().removeSurrounding("\"")
                        val cleanValue = value.trim()
                        map[cleanKey] = when {
                            cleanValue.startsWith("\"") -> cleanValue.removeSurrounding("\"")
                            cleanValue == "true" -> true
                            cleanValue == "false" -> false
                            cleanValue.toIntOrNull() != null -> cleanValue.toInt()
                            cleanValue.toDoubleOrNull() != null -> cleanValue.toDouble()
                            else -> cleanValue
                        }
                    }
                    map
                }
                json.trim().startsWith("[") -> {
                    // Parse JSON array
                    val content = json.trim().removeSurrounding("[", "]")
                    if (content.isBlank()) {
                        emptyList()
                    } else {
                        content.split(",").map { it.trim().removeSurrounding("\"") }
                    }
                }
                else -> json.trim()
            }
        } catch (e: Exception) {
            null
        }
    }
}

sealed class AuthError(message: String) : Exception(message) {
    object JWTSecretNotConfigured : AuthError("JWT secret not configured")
    object InvalidTokenFormat : AuthError("Invalid token format")
    object InvalidTokenSignature : AuthError("Invalid token signature")
    object FailedToDecodePayload : AuthError("Failed to decode payload")
    object TokenExpired : AuthError("Token has expired")
    object SigningFailed : AuthError("Signing failed")
}
