use axum::{
    extract::{Path, State},
    Json,
};

use crate::errors::{ChargingError, ChargingResult, ErrorContext, validate_session_id};
use crate::models::AppState;

/// POST /v1/usage
/// Record usage event
pub async fn record_usage(
    State(state): State<AppState>,
    Json(req): Json<serde_json::Value>,
) -> ChargingResult<Json<serde_json::Value>> {
    let imsi = req.get("imsi").and_then(|v| v.as_str()).ok_or_else(|| {
        ChargingError::InvalidInput("Missing IMSI".to_string())
    })?;
    let session_id = req.get("session_id").and_then(|v| v.as_str()).ok_or_else(|| {
        ChargingError::InvalidInput("Missing session_id".to_string())
    })?;
    
    validate_session_id(session_id)?;

    let event = crate::charging::types::UsageEvent {
        imsi: imsi.to_string(),
        session_id: session_id.to_string(),
        volume: req.get("volume").and_then(|v| v.as_u64()).unwrap_or(0),
        cost: req.get("cost").and_then(|v| v.as_f64()).unwrap_or(0.0),
        rate: req.get("rate").and_then(|v| v.as_f64()).unwrap_or(0.0),
        usage_type: crate::charging::types::UsageType::Data,
        timestamp: chrono::Utc::now(),
    };

    state.charging_engine.record_usage_event(&event).await
        .with_context("Failed to record usage event")?;

    Ok(Json(serde_json::json!({
        "status": "recorded",
        "imsi": imsi,
        "session_id": session_id,
    })))
}

/// POST /v1/usage/calculate-cost
/// Calculate usage cost for an event
pub async fn calculate_usage_cost(
    State(state): State<AppState>,
    Json(req): Json<serde_json::Value>,
) -> ChargingResult<Json<serde_json::Value>> {
    let imsi = req.get("imsi").and_then(|v| v.as_str()).ok_or_else(|| {
        ChargingError::InvalidInput("Missing IMSI".to_string())
    })?;
    let session_id = req.get("session_id").and_then(|v| v.as_str()).ok_or_else(|| {
        ChargingError::InvalidInput("Missing session_id".to_string())
    })?;
    
    validate_session_id(session_id)?;

    let event = crate::charging::types::UsageEvent {
        imsi: imsi.to_string(),
        session_id: session_id.to_string(),
        volume: req.get("volume").and_then(|v| v.as_u64()).unwrap_or(0),
        cost: req.get("cost").and_then(|v| v.as_f64()).unwrap_or(0.0),
        rate: req.get("rate").and_then(|v| v.as_f64()).unwrap_or(0.0),
        usage_type: crate::charging::types::UsageType::Data,
        timestamp: chrono::Utc::now(),
    };

    let cost = state.charging_engine.calculate_usage_cost(&event).await
        .with_context("Failed to calculate usage cost")?;
    
    Ok(Json(serde_json::json!({
        "cost": cost,
        "imsi": imsi,
        "session_id": session_id,
    })))
}

/// POST /v1/usage/rate
/// Rate usage event
pub async fn rate_usage(
    State(state): State<AppState>,
    Json(req): Json<serde_json::Value>,
) -> ChargingResult<Json<serde_json::Value>> {
    let imsi = req.get("imsi").and_then(|v| v.as_str()).ok_or_else(|| {
        ChargingError::InvalidInput("Missing IMSI".to_string())
    })?;
    let session_id = req.get("session_id").and_then(|v| v.as_str()).ok_or_else(|| {
        ChargingError::InvalidInput("Missing session_id".to_string())
    })?;
    
    validate_session_id(session_id)?;

    let event = crate::charging::types::UsageEvent {
        imsi: imsi.to_string(),
        session_id: session_id.to_string(),
        volume: req.get("volume").and_then(|v| v.as_u64()).unwrap_or(0),
        cost: req.get("cost").and_then(|v| v.as_f64()).unwrap_or(0.0),
        rate: req.get("rate").and_then(|v| v.as_f64()).unwrap_or(0.0),
        usage_type: crate::charging::types::UsageType::Data,
        timestamp: chrono::Utc::now(),
    };

    let rated_event = state.charging_engine.rate_usage(event).await
        .with_context("Failed to rate usage")?;
    
    Ok(Json(serde_json::json!({
        "rated_event": rated_event,
    })))
}

/// POST /v1/usage/process
/// Process usage event with full rating and billing
pub async fn process_usage(
    State(state): State<AppState>,
    Json(req): Json<serde_json::Value>,
) -> ChargingResult<Json<serde_json::Value>> {
    let imsi = req.get("imsi").and_then(|v| v.as_str()).ok_or_else(|| {
        ChargingError::InvalidInput("Missing IMSI".to_string())
    })?;
    let session_id = req.get("session_id").and_then(|v| v.as_str()).ok_or_else(|| {
        ChargingError::InvalidInput("Missing session_id".to_string())
    })?;
    
    validate_session_id(session_id)?;

    let event = crate::charging::types::UsageEvent {
        imsi: imsi.to_string(),
        session_id: session_id.to_string(),
        volume: req.get("volume").and_then(|v| v.as_u64()).unwrap_or(0),
        cost: req.get("cost").and_then(|v| v.as_f64()).unwrap_or(0.0),
        rate: req.get("rate").and_then(|v| v.as_f64()).unwrap_or(0.0),
        usage_type: crate::charging::types::UsageType::Data,
        timestamp: chrono::Utc::now(),
    };

    state.charging_engine.process_usage_event(event).await
        .with_context("Failed to process usage event")?;
    
    Ok(Json(serde_json::json!({
        "status": "processed",
        "imsi": imsi,
    })))
}

/// GET /v1/invoice/:imsi/:period
/// Generate invoice for billing period
pub async fn generate_invoice(
    Path((imsi, period)): Path<(String, String)>,
    State(state): State<AppState>,
) -> ChargingResult<Json<serde_json::Value>> {
    let invoice = state.charging_engine.generate_invoice(&imsi, &period).await
        .with_context("Failed to generate invoice")?;
    
    Ok(Json(serde_json::json!({
        "imsi": imsi,
        "period": period,
        "total_amount": invoice,
    })))
}
