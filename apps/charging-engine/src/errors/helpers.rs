use anyhow::Context;

use super::types::ChargingError;

pub type ChargingResult<T> = Result<T, ChargingError>;

pub trait ErrorContext<T> {
    fn with_context(self, msg: &str) -> ChargingResult<T>;
}

impl<T> ErrorContext<T> for anyhow::Result<T> {
    fn with_context(self, msg: &str) -> ChargingResult<T> {
        self.map_err(|e| ChargingError::InternalError(format!("{}: {}", msg, e)))
            .map_err(|e| e.into())
    }
}

impl<T, E> ErrorContext<T> for Result<T, E>
where
    E: Into<ChargingError>,
{
    fn with_context(self, msg: &str) -> ChargingResult<T> {
        self.map_err(|e| {
            let charging_err = e.into();
            match charging_err {
                ChargingError::InternalError(internal_msg) => {
                    ChargingError::InternalError(format!("{}: {}", msg, internal_msg))
                }
                _ => charging_err,
            }
        })
    }
}

// Validation helper functions
pub fn validate_ip(ip: &str) -> ChargingResult<()> {
    if ip.is_empty() {
        return Err(ChargingError::InvalidInput("IP address cannot be empty".to_string()));
    }

    // Basic IP validation (simplified)
    let parts: Vec<&str> = ip.split('.').collect();
    if parts.len() != 4 {
        return Err(ChargingError::InvalidInput("Invalid IP address format".to_string()));
    }

    for part in parts {
        if let Ok(num) = part.parse::<u8>() {
            if num == 0 {
                return Err(ChargingError::InvalidInput("IP address cannot be 0.0.0.0".to_string()));
            }
        } else {
            return Err(ChargingError::InvalidInput("Invalid IP address octet".to_string()));
        }
    }

    Ok(())
}

pub fn validate_bytes(bytes: u64) -> ChargingResult<()> {
    if bytes == 0 {
        return Err(ChargingError::InvalidInput("Bytes cannot be zero".to_string()));
    }

    if bytes > 1_000_000_000_000 { // 1TB limit
        return Err(ChargingError::InvalidInput("Bytes exceed maximum limit".to_string()));
    }

    Ok(())
}

pub fn validate_imsi(imsi: &str) -> ChargingResult<()> {
    if imsi.is_empty() {
        return Err(ChargingError::InvalidInput("IMSI cannot be empty".to_string()));
    }

    if imsi.len() < 15 || imsi.len() > 15 {
        return Err(ChargingError::InvalidInput("IMSI must be exactly 15 digits".to_string()));
    }

    if !imsi.chars().all(|c| c.is_ascii_digit()) {
        return Err(ChargingError::InvalidInput("IMSI must contain only digits".to_string()));
    }

    Ok(())
}

pub fn validate_session_id(session_id: &str) -> ChargingResult<()> {
    if session_id.is_empty() {
        return Err(ChargingError::InvalidInput("Session ID cannot be empty".to_string()));
    }

    if session_id.len() > 64 {
        return Err(ChargingError::InvalidInput("Session ID too long".to_string()));
    }

    Ok(())
}

pub fn validate_amount(amount: f64) -> ChargingResult<()> {
    if amount <= 0.0 {
        return Err(ChargingError::InvalidInput("Amount must be positive".to_string()));
    }

    if amount > 10_000.0 {
        return Err(ChargingError::InvalidInput("Amount exceeds maximum limit".to_string()));
    }

    Ok(())
}

// Error conversion helpers
impl From<redis::RedisError> for ChargingError {
    fn from(err: redis::RedisError) -> Self {
        ChargingError::RedisConnection(err.to_string())
    }
}

impl From<serde_json::Error> for ChargingError {
    fn from(err: serde_json::Error) -> Self {
        ChargingError::SerializationError(err.to_string())
    }
}

impl From<std::num::ParseIntError> for ChargingError {
    fn from(err: std::num::ParseIntError) -> Self {
        ChargingError::InvalidInput(format!("Parse error: {}", err))
    }
}

impl From<std::num::ParseFloatError> for ChargingError {
    fn from(err: std::num::ParseFloatError) -> Self {
        ChargingError::InvalidInput(format!("Parse error: {}", err))
    }
}

// Error logging helpers
pub fn log_error(error: &ChargingError) {
    match error {
        ChargingError::RedisConnection(msg) => {
            tracing::error!("Redis connection error: {}", msg);
        }
        ChargingError::RedisOperation(msg) => {
            tracing::error!("Redis operation error: {}", msg);
        }
        ChargingError::SubscriberNotFound(imsi) => {
            tracing::warn!("Subscriber not found: {}", imsi);
        }
        ChargingError::RatingPlanNotFound(plan_id) => {
            tracing::warn!("Rating plan not found: {}", plan_id);
        }
        ChargingError::InsufficientCredit { available, requested } => {
            tracing::warn!("Insufficient credit: available={}, requested={}", available, requested);
        }
        ChargingError::UsageBlocked(reason) => {
            tracing::warn!("Usage blocked: {}", reason);
        }
        ChargingError::InvalidInput(msg) => {
            tracing::warn!("Invalid input: {}", msg);
        }
        ChargingError::SerializationError(msg) => {
            tracing::error!("Serialization error: {}", msg);
        }
        ChargingError::ConfigurationError(msg) => {
            tracing::error!("Configuration error: {}", msg);
        }
        ChargingError::InternalError(msg) => {
            tracing::error!("Internal error: {}", msg);
        }
    }
}
