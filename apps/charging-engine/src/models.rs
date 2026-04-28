use serde::{Deserialize, Serialize};

use crate::auth::AuthConfig;

#[derive(Clone)]
pub struct AppState {
    pub charging_engine: std::sync::Arc<crate::charging::ChargingEngine>,
    pub auth_config: AuthConfig,
}

#[derive(Deserialize)]
pub struct CreditCheckRequest {
    pub bytes_requested: u64,
}

#[derive(Serialize)]
pub struct CreditCheckResponse {
    pub allowed: bool,
    pub remaining_bytes: i64,
}

#[derive(Deserialize)]
pub struct DeductRequest {
    pub bytes_used: u64,
}

#[derive(Deserialize)]
pub struct AddCreditRequest {
    pub bytes_to_add: u64,
}

#[derive(Serialize)]
pub struct BalanceResponse {
    pub ip: String,
    pub balance_bytes: i64,
}

#[derive(Serialize)]
pub struct HealthResponse {
    pub status: String,
    pub timestamp: chrono::DateTime<chrono::Utc>,
}

impl Default for HealthResponse {
    fn default() -> Self {
        Self {
            status: "OK".to_string(),
            timestamp: chrono::Utc::now(),
        }
    }
}
