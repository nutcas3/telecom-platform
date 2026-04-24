use axum::{
    extract::{Path, State},
    Json,
};
use tracing::info;

use crate::errors::{ChargingResult, ErrorContext, log_error, validate_ip, validate_bytes, validate_amount};
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
