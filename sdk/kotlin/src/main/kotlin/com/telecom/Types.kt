package com.telecom

import kotlinx.serialization.*
import java.time.*

@Serializable
enum class SubscriberStatus(val value: String) {
    @SerialName("active")
    ACTIVE("active"),
    @SerialName("suspended")
    SUSPENDED("suspended"),
    @SerialName("terminated")
    TERMINATED("terminated")
}

@Serializable
enum class UsageType(val value: String) {
    @SerialName("data")
    DATA("data"),
    @SerialName("voice")
    VOICE("voice"),
    @SerialName("sms")
    SMS("sms")
}

@Serializable
enum class PaymentStatus(val value: String) {
    @SerialName("pending")
    PENDING("pending"),
    @SerialName("completed")
    COMPLETED("completed"),
    @SerialName("failed")
    FAILED("failed"),
    @SerialName("refunded")
    REFUNDED("refunded")
}

@Serializable
data class Subscriber(
    val id: Long,
    val imsi: String,
    val msisdn: String,
    val firstName: String,
    val lastName: String,
    val email: String,
    val organizationId: String? = null,
    val status: SubscriberStatus,
    val planId: Long,
    val balance: Double,
    val createdAt: LocalDateTime,
    val updatedAt: LocalDateTime
)

@Serializable
data class SubscriberList(
    val subscribers: List<Subscriber>,
    val total: Long,
    val page: Int,
    val pageSize: Int,
    val hasNext: Boolean,
    val hasPrev: Boolean
)

@Serializable
data class CreateSubscriberRequest(
    val imsi: String,
    val msisdn: String,
    val firstName: String,
    val lastName: String,
    val email: String,
    val planId: Long,
    val organizationId: String? = null
)

@Serializable
data class UpdateSubscriberRequest(
    val firstName: String? = null,
    val lastName: String? = null,
    val email: String? = null,
    val planId: Long? = null,
    val status: SubscriberStatus? = null
)

@Serializable
data class UsageStats(
    val subscriberId: String,
    val dataUp: Long,
    val dataDown: Long,
    val voiceSeconds: Long,
    val smsCount: Long,
    val periodStart: LocalDateTime,
    val periodEnd: LocalDateTime,
    val cost: Double
)

@Serializable
data class UsageEvent(
    val id: String,
    val subscriberId: String,
    val usageType: UsageType,
    val amount: Long,
    val cost: Double,
    val timestamp: LocalDateTime,
    val metadata: Map<String, @Serializable(with = AnySerializer::class) Any>? = null
)

@Serializable
data class UsageEventList(
    val events: List<UsageEvent>,
    val total: Long,
    val page: Int,
    val pageSize: Int,
    val hasNext: Boolean,
    val hasPrev: Boolean
)

@Serializable
data class CurrentSession(
    val sessionId: String,
    val startTime: LocalDateTime,
    val dataUp: Long,
    val dataDown: Long,
    val voiceSeconds: Long,
    val smsCount: Long
)

@Serializable
data class RealTimeUsage(
    val currentSession: CurrentSession? = null,
    val todayUsage: Map<String, Long>? = null
)

@Serializable
data class PaymentTransaction(
    val id: String,
    val subscriberId: String,
    val amount: Double,
    val currency: String,
    val status: PaymentStatus,
    val gateway: String,
    val transactionId: String? = null,
    val createdAt: LocalDateTime,
    val completedAt: LocalDateTime? = null,
    val metadata: Map<String, @Serializable(with = AnySerializer::class) Any>? = null
)

@Serializable
data class PaymentTransactionList(
    val transactions: List<PaymentTransaction>,
    val total: Long,
    val page: Int,
    val pageSize: Int,
    val hasNext: Boolean,
    val hasPrev: Boolean
)

@Serializable
data class CreatePaymentRequest(
    val subscriberId: String,
    val amount: Double,
    val currency: String = "USD",
    val gateway: String = "stripe",
    val metadata: Map<String, @Serializable(with = AnySerializer::class) Any>? = null
)

@Serializable
data class RatingPlan(
    val planId: String,
    val name: String,
    val dataRate: Double,
    val voiceRate: Double,
    val smsRate: Double,
    val monthlyFee: Double,
    val dataLimit: Long,
    val voiceLimit: Long,
    val smsLimit: Long
)

@Serializable
data class SystemStats(
    val activeSessions: Long,
    val totalAccounts: Long,
    val blockedUsers: Long,
    val lowBalanceAlerts: Long,
    val uptime: Double,
    val cpuUsage: Double,
    val memoryUsage: Double
)

@Serializable
data class HealthStatus(
    val status: String,
    val timestamp: LocalDateTime,
    val checks: Map<String, @Serializable(with = AnySerializer::class) Any>,
    val uptime: Double
)

@Serializable
data class WebSocketMessage(
    val type: String,
    val data: Map<String, @Serializable(with = AnySerializer::class) Any>,
    val timestamp: LocalDateTime
)

@Serializable
data class GraphQLRequest(
    val query: String,
    val variables: Map<String, @Serializable(with = AnySerializer::class) Any> = emptyMap()
)

@Serializable
data class GraphQLResponse(
    val data: Map<String, @Serializable(with = AnySerializer::class) Any>? = null,
    val errors: List<GraphQLError>? = null
)

@Serializable
data class GraphQLError(
    val message: String,
    val locations: List<GraphQLErrorLocation>? = null,
    val path: List<String>? = null,
    val extensions: Map<String, @Serializable(with = AnySerializer::class) Any>? = null
)

@Serializable
data class GraphQLErrorLocation(
    val line: Int,
    val column: Int
)

object AnySerializer : KSerializer<Any> {
    override val descriptor: SerialDescriptor = buildClassSerialDescriptor("Any") {
        element("value")
    }

    override fun serialize(encoder: Encoder, value: Any) {
        when (value) {
            is String -> encoder.encodeString(value)
            is Int -> encoder.encodeInt(value)
            is Long -> encoder.encodeLong(value)
            is Double -> encoder.encodeDouble(value)
            is Float -> encoder.encodeDouble(value.toDouble())
            is Boolean -> encoder.encodeBoolean(value)
            else -> throw IllegalArgumentException("Unsupported type: ${value::class}")
        }
    }

    override fun deserialize(decoder: Decoder): Any {
        return when (val elementValue = decoder.decodeSerializableValue(JsonElement.serializer())) {
            is JsonPrimitive -> {
                when {
                    elementValue.isString -> elementValue.content
                    elementValue.content.contains('.') -> elementValue.content.toDouble()
                    else -> elementValue.content.toLong()
                }
            }
            is JsonArray -> elementValue.map { AnySerializer.deserialize(JsonDecoder(it)) }
            is JsonObject -> elementValue.mapValues { AnySerializer.deserialize(JsonDecoder(it.value)) }
            else -> throw IllegalArgumentException("Unsupported JSON element type")
        }
    }
}

sealed class TelecomException(message: String, cause: Throwable? = null) : Exception(message, cause) {
    
    class AuthenticationError(message: String) : TelecomException(message)
    
    class APIError(message: String) : TelecomException(message)
    
    class NetworkError(message: String) : TelecomException(message)
    
    class ValidationError(message: String) : TelecomException(message)
    
    class RateLimitError(message: String) : TelecomException(message)
    
    class ServerError(message: String) : TelecomException(message)
    
    class WebSocketError(message: String) : TelecomException(message)
}
