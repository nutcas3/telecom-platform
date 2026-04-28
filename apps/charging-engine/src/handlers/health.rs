use axum::Json;

use crate::errors::ChargingResult;
use crate::models::HealthResponse;

/// GET /health
/// Health check endpoint
pub async fn health_check() -> ChargingResult<Json<HealthResponse>> {
    Ok(Json(HealthResponse::default()))
}
