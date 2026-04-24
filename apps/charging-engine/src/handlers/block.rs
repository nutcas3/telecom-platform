use axum::{
    extract::{Path, State},
    Json,
};

use crate::errors::{ChargingResult, ErrorContext};
use crate::models::AppState;

/// POST /v1/block/:ip
/// Block a user
pub async fn block_user(
    Path(ip): Path<String>,
    State(state): State<AppState>,
) -> ChargingResult<Json<serde_json::Value>> {
    state.charging_engine.block_user(&ip).await
        .with_context("Failed to block user")?;
    
    Ok(Json(serde_json::json!({
        "status": "blocked",
        "ip": ip,
    })))
}

/// POST /v1/unblock/:ip
/// Unblock a user
pub async fn unblock_user(
    Path(ip): Path<String>,
    State(state): State<AppState>,
) -> ChargingResult<Json<serde_json::Value>> {
    state.charging_engine.unblock_user(&ip).await
        .with_context("Failed to unblock user")?;
    
    Ok(Json(serde_json::json!({
        "status": "unblocked",
        "ip": ip,
    })))
}

/// GET /v1/blocked/:ip
/// Check if user is blocked
pub async fn is_user_blocked(
    Path(ip): Path<String>,
    State(state): State<AppState>,
) -> ChargingResult<Json<serde_json::Value>> {
    let blocked = state.charging_engine.is_user_blocked(&ip).await
        .with_context("Failed to check blocked status")?;
    
    Ok(Json(serde_json::json!({
        "ip": ip,
        "blocked": blocked,
    })))
}
