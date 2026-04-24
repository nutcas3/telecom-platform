mod api;
mod charging;
mod monitoring;
mod errors;
mod models;

use anyhow::Result;
use std::sync::Arc;
use tracing::info;

use api::create_router;
use charging::{ChargingEngine, RatingPlansRepo};

#[tokio::main]
async fn main() -> Result<()> {
    // Initialize tracing
    tracing_subscriber::fmt()
        .with_target(false)
        .compact()
        .init();

    // Load environment variables
    dotenv::dotenv().ok();

    // Initialize charging engine
    let redis_url = std::env::var("REDIS_URI").unwrap_or_else(|_| "redis://127.0.0.1/".to_string());
    let database_url = std::env::var("DATABASE_URL")
        .expect("DATABASE_URL is required (for rating plans)");
    let sync_interval = std::env::var("SYNC_INTERVAL")
        .unwrap_or_else(|_| "1".to_string())
        .parse::<u64>()
        .unwrap_or(1);

    // Connect to Postgres and seed defaults if the table is empty.
    let plans_repo = RatingPlansRepo::connect(&database_url).await?;
    plans_repo.seed_defaults().await?;
    info!("Connected to Postgres and ensured default rating plans");

    let charging_engine = Arc::new(ChargingEngine::new(&redis_url, plans_repo, sync_interval)?);

    // Test Redis connection
    charging_engine.test_connection().await?;
    info!("Connected to Redis successfully");

    // Create application state
    let state = api::AppState {
        charging_engine,
    };

    // Create router
    let app = create_router(state);

    // Start server
    let port = std::env::var("SERVER_PORT").unwrap_or_else(|_| "8080".to_string());
    let addr = format!("0.0.0.0:{}", port);
    let listener = tokio::net::TcpListener::bind(&addr).await?;
    
    info!("Charging engine listening on {}", addr);
    
    axum::serve(listener, app).await?;

    Ok(())
}
