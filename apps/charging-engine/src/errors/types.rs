use axum::response::{IntoResponse, Response};
use axum::http::StatusCode;
use serde_json::json;
use std::fmt;

#[derive(Debug, Clone)]
pub enum ChargingError {
    RedisConnection(String),
    RedisOperation(String),
    DatabaseError(String),
    SubscriberNotFound(String),
    RatingPlanNotFound(String),
    InsufficientCredit { available: u64, requested: u64 },
    UsageBlocked(String),
    InvalidInput(String),
    SerializationError(String),
    InternalError(String),
}

impl fmt::Display for ChargingError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            ChargingError::RedisConnection(msg) => write!(f, "Redis connection error: {}", msg),
            ChargingError::RedisOperation(msg) => write!(f, "Redis operation error: {}", msg),
            ChargingError::DatabaseError(msg) => write!(f, "Database error: {}", msg),
            ChargingError::SubscriberNotFound(imsi) => write!(f, "Subscriber not found: {}", imsi),
            ChargingError::RatingPlanNotFound(plan_id) => write!(f, "Rating plan not found: {}", plan_id),
            ChargingError::InsufficientCredit { available, requested } => {
                write!(f, "Insufficient credit: available={}, requested={}", available, requested)
            }
            ChargingError::UsageBlocked(reason) => write!(f, "Usage blocked: {}", reason),
            ChargingError::InvalidInput(msg) => write!(f, "Invalid input: {}", msg),
            ChargingError::SerializationError(msg) => write!(f, "Serialization error: {}", msg),
            ChargingError::InternalError(msg) => write!(f, "Internal error: {}", msg),
        }
    }
}

impl std::error::Error for ChargingError {}

impl IntoResponse for ChargingError {
    fn into_response(self) -> Response {
        let (status, message) = match &self {
            ChargingError::RedisConnection(_) => (StatusCode::SERVICE_UNAVAILABLE, "Redis connection error"),
            ChargingError::RedisOperation(_) => (StatusCode::INTERNAL_SERVER_ERROR, "Redis operation error"),
            ChargingError::DatabaseError(_) => (StatusCode::INTERNAL_SERVER_ERROR, "Database error"),
            ChargingError::SubscriberNotFound(_) => (StatusCode::NOT_FOUND, "Subscriber not found"),
            ChargingError::RatingPlanNotFound(_) => (StatusCode::NOT_FOUND, "Rating plan not found"),
            ChargingError::InsufficientCredit { .. } => (StatusCode::PAYMENT_REQUIRED, "Insufficient credit"),
            ChargingError::UsageBlocked(_) => (StatusCode::FORBIDDEN, "Usage blocked"),
            ChargingError::InvalidInput(_) => (StatusCode::BAD_REQUEST, "Invalid input"),
            ChargingError::SerializationError(_) => (StatusCode::INTERNAL_SERVER_ERROR, "Serialization error"),
            ChargingError::InternalError(_) => (StatusCode::INTERNAL_SERVER_ERROR, "Internal error"),
        };

        let body = json!({
            "error": message,
            "detail": self.to_string(),
        });

        (status, axum::Json(body)).into_response()
    }
}
