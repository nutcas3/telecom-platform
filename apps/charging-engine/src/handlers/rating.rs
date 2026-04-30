use axum::{
    extract::{Path, State},
    Json,
};

use crate::errors::{ChargingError, ChargingResult, ErrorContext};
use crate::models::AppState;

/// GET /v1/rating-plans
/// List all rating plans
pub async fn list_rating_plans(
    State(state): State<AppState>,
) -> ChargingResult<Json<serde_json::Value>> {
    let plans = state.charging_engine.list_rating_plans().await
        .with_context("Failed to list rating plans")?;
    
    // Also use list_map for internal lookups
    let _ = state.charging_engine.plans.list_map().await
        .with_context("Failed to list rating plans as map")?;
    
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
    
    let plan_id = plan.plan_id.clone();
    state.charging_engine.add_rating_plan(plan).await
        .with_context("Failed to add rating plan")?;
    
    Ok(Json(serde_json::json!({
        "status": "added",
        "plan_id": plan_id,
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
