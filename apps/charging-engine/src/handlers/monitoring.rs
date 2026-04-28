use axum::{
    extract::{State},
    Json,
};

use crate::errors::{ChargingResult, ErrorContext};
use crate::models::AppState;

/// GET /v1/stats
/// Get system statistics
pub async fn get_system_stats(
    State(state): State<AppState>,
) -> ChargingResult<Json<serde_json::Value>> {
    let stats = state.charging_engine.get_system_statistics().await
        .with_context("Failed to get system statistics")?;
    
    Ok(Json(serde_json::json!({
        "total_accounts": stats.total_accounts,
        "active_sessions": stats.active_sessions,
        "blocked_users": stats.blocked_users,
        "low_balance_alerts": stats.low_balance_alerts,
        "uptime": stats.uptime,
    })))
}

/// GET /v1/metrics
/// Get performance metrics
pub async fn get_performance_metrics(
    State(state): State<AppState>,
) -> ChargingResult<Json<serde_json::Value>> {
    let metrics = state.charging_engine.get_performance_metrics().await
        .with_context("Failed to get performance metrics")?;
    
    Ok(Json(serde_json::json!({
        "connected_clients": metrics.connected_clients,
        "used_memory": metrics.used_memory,
        "total_commands_processed": metrics.total_commands_processed,
        "requests_per_second": metrics.requests_per_second,
        "average_response_time": metrics.average_response_time,
    })))
}

/// GET /v1/errors
/// Get error statistics
pub async fn get_error_stats(
    State(state): State<AppState>,
) -> ChargingResult<Json<serde_json::Value>> {
    let errors = state.charging_engine.get_error_statistics().await
        .with_context("Failed to get error statistics")?;
    
    Ok(Json(serde_json::json!({
        "total_errors": errors.total_errors,
        "error_types": errors.error_types,
        "last_error": errors.last_error,
    })))
}

/// POST /v1/sync/start
/// Start background sync
pub async fn start_sync(
    State(state): State<AppState>,
) -> ChargingResult<Json<serde_json::Value>> {
    state.charging_engine.start_background_sync().await
        .with_context("Failed to start background sync")?;
    
    Ok(Json(serde_json::json!({
        "status": "sync_started",
    })))
}

/// GET /v1/health/detailed
/// Get detailed health check
pub async fn detailed_health_check(
    State(state): State<AppState>,
) -> ChargingResult<Json<serde_json::Value>> {
    let health = state.charging_engine.health_check().await
        .with_context("Failed to get health check")?;
    
    Ok(Json(serde_json::json!({
        "redis_connected": health.redis_connected,
        "active_sync": health.active_sync,
        "last_sync": health.last_sync,
        "memory_usage": health.memory_usage,
    })))
}
