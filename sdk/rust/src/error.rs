use reqwest::StatusCode;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum TelecomError {
    #[error("Authentication failed")]
    AuthenticationError,
    #[error("API error: {0}")]
    APIError(String),
    #[error("Network error: {0}")]
    NetworkError(#[from] reqwest::Error),
    #[error("Validation error: {0}")]
    ValidationError(String),
    #[error("Rate limit exceeded")]
    RateLimitError,
    #[error("Server error: {0}")]
    ServerError(StatusCode),
    #[error("WebSocket error: {0}")]
    WebSocketError(String),
    #[error("JSON error: {0}")]
    JsonError(#[from] serde_json::Error),
}
