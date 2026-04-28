use axum::{
    Json,
    extract::State,
};
use redis::AsyncCommands;
use serde::Serialize;
use std::sync::atomic::AtomicBool;
use std::sync::Arc;

#[cfg(feature = "ebpf")]
use crate::ebpf::EbpfManager;

#[derive(Serialize)]
pub struct HealthResponse {
    pub status: String,
    pub service: String,
    pub timestamp: String,
}

#[derive(Serialize)]
pub struct ReadinessResponse {
    pub status: String,
    pub service: String,
    pub timestamp: String,
    pub checks: serde_json::Value,
}

#[derive(Serialize)]
pub struct MetricsResponse {
    pub status: String,
    pub service: String,
    pub timestamp: String,
    pub tracked_ips: usize,
    pub total_bytes_processed: u64,
    pub low_credit_users: usize,
}

pub async fn health_handler() -> Json<HealthResponse> {
    Json(HealthResponse {
        status: "healthy".to_string(),
        service: "packet-gateway".to_string(),
        timestamp: chrono::Utc::now().to_rfc3339(),
    })
}

pub async fn liveness_handler() -> Json<HealthResponse> {
    Json(HealthResponse {
        status: "alive".to_string(),
        service: "packet-gateway".to_string(),
        timestamp: chrono::Utc::now().to_rfc3339(),
    })
}

pub async fn readiness_handler(
    State(redis_client): State<Arc<redis::Client>>,
    State(ebpf_attached): State<Arc<AtomicBool>>,
) -> Result<Json<ReadinessResponse>, axum::http::StatusCode> {
    let mut checks = serde_json::Map::new();
    
    let redis_ok = match redis_client.get_multiplexed_async_connection().await {
        Ok(mut conn) => {
            match redis::AsyncCommands::ping::<String>(&mut conn).await {
                Ok(_) => {
                    checks.insert("redis".to_string(), serde_json::Value::String("ok".to_string()));
                    true
                }
                Err(e) => {
                    checks.insert("redis".to_string(), serde_json::Value::String(format!("error: {}", e)));
                    false
                }
            }
        }
        Err(e) => {
            checks.insert("redis".to_string(), serde_json::Value::String(format!("connection error: {}", e)));
            false
        }
    };
    
    let ebpf_ok = ebpf_attached.load(std::sync::atomic::Ordering::Relaxed);
    checks.insert("ebpf".to_string(), serde_json::Value::String(if ebpf_ok { "ok".to_string() } else { "not attached".to_string() }));
    
    #[cfg(feature = "ebpf")]
    let ready = redis_ok && ebpf_ok;
    #[cfg(not(feature = "ebpf"))]
    let ready = redis_ok;
    
    if ready {
        Ok(Json(ReadinessResponse {
            status: "ready".to_string(),
            service: "packet-gateway".to_string(),
            timestamp: chrono::Utc::now().to_rfc3339(),
            checks: serde_json::Value::Object(checks),
        }))
    } else {
        Err(axum::http::StatusCode::SERVICE_UNAVAILABLE)
    }
}

#[cfg(feature = "ebpf")]
pub async fn metrics_handler(
    State(ebpf_manager): State<Arc<EbpfManager>>,
) -> Json<MetricsResponse> {
    let tracked_ips = ebpf_manager.get_packet_stats()
        .map(|stats| stats.len())
        .unwrap_or(0);
    
    let total_bytes_processed = ebpf_manager.get_packet_stats()
        .map(|stats| stats.iter().map(|s| s.bytes).sum())
        .unwrap_or(0);
    
    let low_credit_users = ebpf_manager.get_credit_info()
        .map(|credits| credits.iter().filter(|c| c.credit < 1000).count())
        .unwrap_or(0);
    
    Json(MetricsResponse {
        status: "ok".to_string(),
        service: "packet-gateway".to_string(),
        timestamp: chrono::Utc::now().to_rfc3339(),
        tracked_ips,
        total_bytes_processed,
        low_credit_users,
    })
}

#[cfg(not(feature = "ebpf"))]
pub async fn metrics_handler() -> Json<MetricsResponse> {
    Json(MetricsResponse {
        status: "ok".to_string(),
        service: "packet-gateway".to_string(),
        timestamp: chrono::Utc::now().to_rfc3339(),
        tracked_ips: 0,
        total_bytes_processed: 0,
        low_credit_users: 0,
    })
}
