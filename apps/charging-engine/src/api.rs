use axum::{
    extract::{Path, State},
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use tower_http::cors::{Any, CorsLayer};
use tracing::{info, warn};

use crate::errors::{ChargingError, ChargingResult, validate_ip, validate_bytes, ErrorContext};
use crate::models::*;

pub struct AppState {
    pub charging_engine: std::sync::Arc<crate::charging::ChargingEngine>,
}

pub fn create_router(state: AppState) -> Router {
    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    Router::new()
        .route("/v1/credit/:ip/check", post(check_credit))
        .route("/v1/credit/:ip/deduct", post(deduct_credit))
        .route("/v1/credit/:ip/add", post(add_credit))
        .route("/v1/credit/:ip/balance", get(get_balance))
        .route("/health", get(health_check))
        .layer(cors)
        .with_state(state)
}

/// POST /v1/credit/:ip/check
/// Check if user has enough credit for data usage
pub async fn check_credit(
    Path(ip): Path<String>,
    State(state): State<AppState>,
    Json(req): Json<CreditCheckRequest>,
) -> ChargingResult<Json<CreditCheckResponse>> {
    // Validate input
    validate_ip(&ip)?;
    validate_bytes(req.bytes_requested)?;

    let allowed = state.charging_engine.check_credit(&ip, req.bytes_requested).await
        .with_context("Failed to check credit")?;
    
    let remaining = state.charging_engine.get_balance(&ip).await
        .with_context("Failed to get balance")?;
    
    info!(
        "Credit check for {}: {} bytes requested, {} bytes available, allowed: {}",
        ip, req.bytes_requested, remaining, allowed
    );

    Ok(Json(CreditCheckResponse {
        allowed,
        remaining_bytes: remaining,
    }))
}

/// POST /v1/credit/:ip/deduct
/// Deduct bytes from user's credit balance
pub async fn deduct_credit(
    Path(ip): Path<String>,
    State(state): State<AppState>,
    Json(req): Json<DeductRequest>,
) -> ChargingResult<()> {
    // Validate input
    validate_ip(&ip)?;
    validate_bytes(req.bytes_used)?;

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
    // Validate input
    validate_ip(&ip)?;
    validate_bytes(req.bytes_to_add)?;

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
    // Validate input
    validate_ip(&ip)?;

    let balance = state.charging_engine.get_balance(&ip).await
        .with_context("Failed to get balance")?;
    
    Ok(Json(BalanceResponse {
        ip,
        balance_bytes: balance,
    }))
}

/// GET /health
/// Health check endpoint
pub async fn health_check() -> ChargingResult<Json<HealthResponse>> {
    Ok(Json(HealthResponse::default()))
}
