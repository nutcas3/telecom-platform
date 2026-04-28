use axum::{
    extract::State,
    routing::{get, post},
    Json, Router,
};
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tokio::sync::RwLock;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PacketGatewayConfig {
    pub interface: String,
    pub redis_url: String,
    pub charging_engine_url: String,
    pub sync_interval: u64,
    pub health_port: u16,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConfigUpdateRequest {
    pub sync_interval: Option<u64>,
    pub charging_engine_url: Option<String>,
}

#[derive(Debug, Clone, Serialize)]
pub struct ConfigResponse {
    pub config: PacketGatewayConfig,
    pub status: String,
}

#[derive(Clone)]
pub struct ConfigState {
    pub config: Arc<RwLock<PacketGatewayConfig>>,
}

pub fn create_config_router(config_state: ConfigState) -> Router {
    Router::new()
        .route("/config", get(get_config).post(update_config))
        .route("/config/sync", post(trigger_sync))
        .with_state(config_state)
}

pub async fn get_config(State(state): State<ConfigState>) -> Json<ConfigResponse> {
    let config = state.config.read().await;
    Json(ConfigResponse {
        config: config.clone(),
        status: "active".to_string(),
    })
}

pub async fn update_config(
    State(state): State<ConfigState>,
    Json(req): Json<ConfigUpdateRequest>,
) -> Json<ConfigResponse> {
    let mut config = state.config.write().await;
    
    if let Some(interval) = req.sync_interval {
        config.sync_interval = interval;
    }
    
    if let Some(url) = req.charging_engine_url {
        config.charging_engine_url = url;
    }
    
    Json(ConfigResponse {
        config: config.clone(),
        status: "updated".to_string(),
    })
}

pub async fn trigger_sync(State(state): State<ConfigState>) -> Json<serde_json::Value> {
    // This would trigger a manual sync from Redis to eBPF maps
    // For now, return a success response
    Json(serde_json::json!({
        "status": "sync_triggered",
        "message": "Manual sync triggered"
    }))
}
