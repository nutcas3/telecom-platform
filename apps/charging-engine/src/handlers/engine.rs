use axum::{
    extract::{State},
    Json,
};

use crate::errors::{ChargingResult, ErrorContext};
use crate::models::AppState;

/// POST /v1/engine/start
/// Start the engine
pub async fn engine_start(
    State(state): State<AppState>,
) -> ChargingResult<Json<serde_json::Value>> {
    state.charging_engine.start().await
        .with_context("Failed to start engine")?;
    
    Ok(Json(serde_json::json!({
        "status": "started",
    })))
}

/// POST /v1/engine/stop
/// Stop the engine
pub async fn engine_stop(
    State(state): State<AppState>,
) -> ChargingResult<Json<serde_json::Value>> {
    state.charging_engine.stop().await
        .with_context("Failed to stop engine")?;
    
    Ok(Json(serde_json::json!({
        "status": "stopped",
    })))
}

/// GET /v1/engine/uptime
/// Get engine uptime
pub async fn engine_uptime(
    State(state): State<AppState>,
) -> ChargingResult<Json<serde_json::Value>> {
    let uptime = state.charging_engine.uptime();
    
    Ok(Json(serde_json::json!({
        "uptime_seconds": uptime.as_secs(),
    })))
}
