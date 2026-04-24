use axum::{
    extract::{Path, State},
    Json,
};

use crate::errors::{ChargingError, ChargingResult, ErrorContext};
use crate::models::AppState;

/// GET /v1/subscriber/:imsi
/// Get subscriber account by IMSI
pub async fn get_subscriber(
    Path(imsi): Path<String>,
    State(state): State<AppState>,
) -> ChargingResult<Json<serde_json::Value>> {
    let account = state.charging_engine.get_subscriber_account(&imsi).await
        .with_context("Failed to get subscriber account")?;

    match account {
        Some(acc) => Ok(Json(serde_json::json!({
            "imsi": acc.imsi,
            "balance": acc.balance,
            "data_used": acc.data_used,
            "data_limit": acc.data_limit,
            "voice_used": acc.voice_used,
            "voice_limit": acc.voice_limit,
            "sms_used": acc.sms_used,
            "sms_limit": acc.sms_limit,
        }))),
        None => Err(ChargingError::SubscriberNotFound(imsi)),
    }
}

/// PUT /v1/subscriber/:imsi
/// Update subscriber account
pub async fn update_subscriber(
    Path(imsi): Path<String>,
    State(state): State<AppState>,
    Json(req): Json<serde_json::Value>,
) -> ChargingResult<Json<serde_json::Value>> {
    let account = crate::charging::types::SubscriberAccount {
        imsi: imsi.clone(),
        balance: req.get("balance").and_then(|v| v.as_i64()).unwrap_or(0),
        data_used: req.get("data_used").and_then(|v| v.as_u64()).unwrap_or(0),
        data_limit: req.get("data_limit").and_then(|v| v.as_u64()).unwrap_or(1_000_000_000),
        voice_used: req.get("voice_used").and_then(|v| v.as_u64()).unwrap_or(0),
        voice_limit: req.get("voice_limit").and_then(|v| v.as_u64()).unwrap_or(1000),
        sms_used: req.get("sms_used").and_then(|v| v.as_u64()).unwrap_or(0),
        sms_limit: req.get("sms_limit").and_then(|v| v.as_u64()).unwrap_or(100),
        status: crate::charging::types::AccountStatus::Active,
        last_updated: chrono::Utc::now(),
    };
    
    state.charging_engine.update_subscriber_account(&account).await
        .with_context("Failed to update subscriber account")?;
    
    Ok(Json(serde_json::json!({
        "status": "updated",
        "imsi": imsi,
    })))
}
