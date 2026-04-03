use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use redis::AsyncCommands;
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tower_http::cors::{Any, CorsLayer};
use tracing::{info, warn};

#[derive(Clone)]
struct AppState {
    redis: redis::Client,
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // Initialize tracing
    tracing_subscriber::fmt()
        .with_target(false)
        .compact()
        .init();

    // Load environment variables
    dotenv::dotenv().ok();

    // Connect to Redis
    let redis_url = std::env::var("REDIS_URI").unwrap_or_else(|_| "redis://127.0.0.1/".to_string());
    let redis_client = redis::Client::open(redis_url)?;
    
    // Test connection
    let mut conn = redis_client.get_multiplexed_async_connection().await?;
    let _: () = redis::cmd("PING").query_async(&mut conn).await?;
    info!("Connected to Redis successfully");

    let state = Arc::new(AppState {
        redis: redis_client,
    });

    // Build HTTP API with CORS
    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let app = Router::new()
        .route("/v1/credit/:ip/check", post(check_credit))
        .route("/v1/credit/:ip/deduct", post(deduct_credit))
        .route("/v1/credit/:ip/add", post(add_credit))
        .route("/v1/credit/:ip/balance", get(get_balance))
        .route("/health", get(health_check))
        .layer(cors)
        .with_state(state);

    // Start server
    let port = std::env::var("SERVER_PORT").unwrap_or_else(|_| "8080".to_string());
    let addr = format!("0.0.0.0:{}", port);
    let listener = tokio::net::TcpListener::bind(&addr).await?;
    
    info!("Charging engine listening on {}", addr);
    
    axum::serve(listener, app).await?;

    Ok(())
}

#[derive(Deserialize)]
struct CreditCheckRequest {
    bytes_requested: u64,
}

#[derive(Serialize)]
struct CreditCheckResponse {
    allowed: bool,
    remaining_bytes: i64,
}

/// POST /v1/credit/:ip/check
/// Check if user has enough credit for data usage
async fn check_credit(
    Path(ip): Path<String>,
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreditCheckRequest>,
) -> impl IntoResponse {
    let mut conn = match state.redis.get_multiplexed_async_connection().await {
        Ok(c) => c,
        Err(e) => {
            warn!("Redis connection failed: {}", e);
            return (
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(CreditCheckResponse {
                    allowed: false,
                    remaining_bytes: 0,
                }),
            );
        }
    };

    let key = format!("credit:{}", ip);

    // Get current credit balance
    let credit: i64 = conn.get(&key).await.unwrap_or(0);

    let allowed = credit >= req.bytes_requested as i64;

    info!(
        "Credit check for {}: {} bytes requested, {} bytes available, allowed: {}",
        ip, req.bytes_requested, credit, allowed
    );

    (
        StatusCode::OK,
        Json(CreditCheckResponse {
            allowed,
            remaining_bytes: credit,
        }),
    )
}

#[derive(Deserialize)]
struct DeductRequest {
    bytes_used: u64,
}

/// POST /v1/credit/:ip/deduct
/// Deduct bytes from user's credit balance
async fn deduct_credit(
    Path(ip): Path<String>,
    State(state): State<Arc<AppState>>,
    Json(req): Json<DeductRequest>,
) -> impl IntoResponse {
    let mut conn = match state.redis.get_multiplexed_async_connection().await {
        Ok(c) => c,
        Err(e) => {
            warn!("Redis connection failed: {}", e);
            return StatusCode::INTERNAL_SERVER_ERROR;
        }
    };

    let key = format!("credit:{}", ip);

    // Atomically subtract bytes
    let new_balance: i64 = conn.decr(&key, req.bytes_used).await.unwrap_or(0);

    info!(
        "User {} deducted {} bytes, remaining: {}",
        ip, req.bytes_used, new_balance
    );

    StatusCode::OK
}

#[derive(Deserialize)]
struct AddCreditRequest {
    bytes_to_add: u64,
}

/// POST /v1/credit/:ip/add
/// Add bytes to user's credit balance
async fn add_credit(
    Path(ip): Path<String>,
    State(state): State<Arc<AppState>>,
    Json(req): Json<AddCreditRequest>,
) -> impl IntoResponse {
    let mut conn = match state.redis.get_multiplexed_async_connection().await {
        Ok(c) => c,
        Err(e) => {
            warn!("Redis connection failed: {}", e);
            return StatusCode::INTERNAL_SERVER_ERROR;
        }
    };

    let key = format!("credit:{}", ip);

    // Atomically add bytes
    let new_balance: i64 = conn.incr(&key, req.bytes_to_add).await.unwrap_or(0);

    info!(
        "User {} added {} bytes, new balance: {}",
        ip, req.bytes_to_add, new_balance
    );

    StatusCode::OK
}

#[derive(Serialize)]
struct BalanceResponse {
    ip: String,
    balance_bytes: i64,
}

/// GET /v1/credit/:ip/balance
/// Get current credit balance
async fn get_balance(
    Path(ip): Path<String>,
    State(state): State<Arc<AppState>>,
) -> impl IntoResponse {
    let mut conn = match state.redis.get_multiplexed_async_connection().await {
        Ok(c) => c,
        Err(_) => {
            return (
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(BalanceResponse {
                    ip: ip.clone(),
                    balance_bytes: 0,
                }),
            );
        }
    };

    let key = format!("credit:{}", ip);
    let balance: i64 = conn.get(&key).await.unwrap_or(0);

    (
        StatusCode::OK,
        Json(BalanceResponse {
            ip,
            balance_bytes: balance,
        }),
    )
}

async fn health_check() -> &'static str {
    "OK"
}
