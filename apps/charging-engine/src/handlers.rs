use axum::{
    extract::{Path, State},
    Json,
};
use tracing::info;

use crate::errors::{ChargingError, ChargingResult, ErrorContext, log_error, validate_ip, validate_bytes, validate_amount, validate_session_id};
use crate::models::*;

/// POST /v1/credit/:ip/check
/// Check if user has enough credit for data usage
pub async fn check_credit(
    Path(ip): Path<String>,
    State(state): State<AppState>,
    Json(req): Json<CreditCheckRequest>,
) -> ChargingResult<Json<CreditCheckResponse>> {
    validate_ip(&ip)?;
    validate_bytes(req.bytes_requested)?;

    let allowed = state.charging_engine.check_credit(&ip, req.bytes_requested).await
        .with_context("Failed to check credit")
        .map_err(|e| {
            log_error(&e);
            e
        })?;
    
    let remaining = state.charging_engine.get_balance(&ip).await
        .with_context("Failed to get balance")?;
    
    info!(
        "Credit check for {}: {} bytes requested, {} bytes available, allowed: {}",
        ip, req.bytes_requested, remaining, allowed
    );

    Ok(Json(CreditCheckResponse {
        allowed,
        remaining_bytes: remaining as i64,
    }))
}

/// POST /v1/credit/:ip/deduct
/// Deduct bytes from user's credit balance
pub async fn deduct_credit(
    Path(ip): Path<String>,
    State(state): State<AppState>,
    Json(req): Json<DeductRequest>,
) -> ChargingResult<()> {
    validate_ip(&ip)?;
    validate_bytes(req.bytes_used)?;
    validate_amount(req.bytes_used as f64)?;

    let new_balance = state.charging_engine.deduct_credit(&ip, req.bytes_used).await
        .with_context("Failed to deduct credit")?;
    
    info!(
        "User {} deducted {} bytes, remaining: {}",
        ip, req.bytes_used, new_balance
    );

    Ok(())
}

/// POST /v1/credit/:ip/add
/// Add bytes to user's credit balance
pub async fn add_credit(
    Path(ip): Path<String>,
    State(state): State<AppState>,
    Json(req): Json<AddCreditRequest>,
) -> ChargingResult<()> {
    validate_ip(&ip)?;
    validate_bytes(req.bytes_to_add)?;
    validate_amount(req.bytes_to_add as f64)?;

    let new_balance = state.charging_engine.add_credit(&ip, req.bytes_to_add).await
        .with_context("Failed to add credit")?;
    
    info!(
        "User {} added {} bytes, new balance: {}",
        ip, req.bytes_to_add, new_balance
    );

    Ok(())
}

/// GET /v1/credit/:ip/balance
/// Get current credit balance
pub async fn get_balance(
    Path(ip): Path<String>,
    State(state): State<AppState>,
) -> ChargingResult<Json<BalanceResponse>> {
    validate_ip(&ip)?;

    let balance = state.charging_engine.get_balance(&ip).await
        .with_context("Failed to get balance")?;

    info!("Retrieved balance for IP: {}", ip);

    Ok(Json(BalanceResponse {
        ip: ip.clone(),
        balance_bytes: balance as i64,
    }))
}

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

/// GET /health
/// Health check endpoint
pub async fn health_check() -> ChargingResult<Json<HealthResponse>> {
    Ok(Json(HealthResponse::default()))
}

/// GET /v1/rating-plans
/// List all rating plans
pub async fn list_rating_plans(
    State(state): State<AppState>,
) -> ChargingResult<Json<serde_json::Value>> {
    let plans = state.charging_engine.list_rating_plans().await
        .with_context("Failed to list rating plans")?;
    
    Ok(Json(serde_json::json!({
        "plans": plans
    })))
}

/// GET /v1/rating-plans/:id
/// Get a specific rating plan
pub async fn get_rating_plan(
    Path(id): Path<String>,
    State(state): State<AppState>,
) -> ChargingResult<Json<serde_json::Value>> {
    let plan = state.charging_engine.get_rating_plan(&id).await
        .with_context("Failed to get rating plan")?;
    
    match plan {
        Some(p) => Ok(Json(serde_json::json!({
            "id": p.plan_id,
            "name": p.name,
            "data_rate": p.data_rate,
            "voice_rate": p.voice_rate,
            "sms_rate": p.sms_rate,
            "monthly_fee": p.monthly_fee,
            "data_limit": p.data_limit,
            "voice_limit": p.voice_limit,
            "sms_limit": p.sms_limit,
        }))),
        None => Err(ChargingError::RatingPlanNotFound(id)),
    }
}

/// POST /v1/rating-plans
/// Add a new rating plan
pub async fn add_rating_plan(
    State(state): State<AppState>,
    Json(req): Json<serde_json::Value>,
) -> ChargingResult<Json<serde_json::Value>> {
    let plan = crate::charging::types::RatingPlan {
        plan_id: req.get("plan_id").and_then(|v| v.as_str()).unwrap_or("default").to_string(),
        name: req.get("name").and_then(|v| v.as_str()).unwrap_or("Default").to_string(),
        data_rate: req.get("data_rate").and_then(|v| v.as_f64()).unwrap_or(0.001),
        voice_rate: req.get("voice_rate").and_then(|v| v.as_f64()).unwrap_or(0.01),
        sms_rate: req.get("sms_rate").and_then(|v| v.as_f64()).unwrap_or(0.1),
        monthly_fee: req.get("monthly_fee").and_then(|v| v.as_f64()).unwrap_or(10.0),
        data_limit: req.get("data_limit").and_then(|v| v.as_u64()).unwrap_or(1_000_000_000),
        voice_limit: req.get("voice_limit").and_then(|v| v.as_u64()).unwrap_or(1000),
        sms_limit: req.get("sms_limit").and_then(|v| v.as_u64()).unwrap_or(100),
    };
    
    state.charging_engine.add_rating_plan(plan.clone()).await
        .with_context("Failed to add rating plan")?;
    
    Ok(Json(serde_json::json!({
        "status": "added",
        "plan_id": plan.plan_id,
    })))
}

/// DELETE /v1/rating-plans/:id
/// Remove a rating plan
pub async fn remove_rating_plan(
    Path(id): Path<String>,
    State(state): State<AppState>,
) -> ChargingResult<Json<serde_json::Value>> {
    let removed = state.charging_engine.remove_rating_plan(&id).await
        .with_context("Failed to remove rating plan")?;
    
    Ok(Json(serde_json::json!({
        "status": if removed { "removed" } else { "not_found" },
        "plan_id": id,
    })))
}

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
