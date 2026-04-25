use anyhow::{Context, Result};
use clap::Parser;
use redis::AsyncCommands;
use tokio::time::{interval, Duration};
use tracing::{info, warn, error, debug};
use axum::{
    routing::get,
    Router,
    Json,
};
use serde::Serialize;
use std::sync::Arc;
use std::sync::atomic::{AtomicBool, Ordering};

mod ebpf;
use ebpf::EbpfManager;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    /// Network interface to attach XDP program
    #[arg(short, long, default_value = "eth0")]
    interface: String,

    /// Redis connection URL
    #[arg(short, long, default_value = "redis://127.0.0.1/")]
    redis_url: String,

    /// Stats sync interval in seconds
    #[arg(short, long, default_value = "1")]
    sync_interval: u64,

    /// Health check HTTP port
    #[arg(long, default_value = "8081")]
    health_port: u16,
}

#[derive(Serialize)]
struct HealthResponse {
    status: String,
    service: String,
    timestamp: String,
}

#[derive(Serialize)]
struct ReadinessResponse {
    status: String,
    service: String,
    timestamp: String,
    checks: serde_json::Value,
}

async fn health_handler() -> Json<HealthResponse> {
    Json(HealthResponse {
        status: "healthy".to_string(),
        service: "packet-gateway".to_string(),
        timestamp: chrono::Utc::now().to_rfc3339(),
    })
}

async fn liveness_handler() -> Json<HealthResponse> {
    Json(HealthResponse {
        status: "alive".to_string(),
        service: "packet-gateway".to_string(),
        timestamp: chrono::Utc::now().to_rfc3339(),
    })
}

#[derive(Serialize)]
struct MetricsResponse {
    status: String,
    service: String,
    timestamp: String,
    tracked_ips: usize,
    total_bytes_processed: u64,
    low_credit_users: usize,
}

async fn metrics_handler(
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

async fn readiness_handler(
    State(redis_client): State<Arc<redis::Client>>,
    State(ebpf_attached): State<Arc<AtomicBool>>,
) -> Result<Json<ReadinessResponse>, axum::http::StatusCode> {
    let mut checks = serde_json::Map::new();
    
    // Check Redis connectivity
    let redis_ok = match redis_client.get_multiplexed_async_connection().await {
        Ok(mut conn) => {
            match redis::cmd("PING").query_async::<_, String>(&mut conn).await {
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
    
    // Check eBPF attachment
    let ebpf_ok = ebpf_attached.load(Ordering::Relaxed);
    checks.insert("ebpf".to_string(), serde_json::Value::String(if ebpf_ok { "ok".to_string() } else { "not attached".to_string() }));
    
    if redis_ok && ebpf_ok {
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

#[tokio::main]
async fn main() -> Result<()> {
    // Initialize tracing
    tracing_subscriber::fmt()
        .with_target(false)
        .compact()
        .init();

    dotenv::dotenv().ok();
    let args = Args::parse();

    info!("Starting Packet Gateway");
    info!("Interface: {}", args.interface);
    info!("Redis URL: {}", args.redis_url);
    info!("Health Port: {}", args.health_port);

    // Connect to Redis
    let redis_client = redis::Client::open(args.redis_url.as_str())?;
    let mut redis_conn = redis_client.get_multiplexed_async_connection().await?;
    
    // Test connection
    let _: () = redis::cmd("PING").query_async(&mut redis_conn).await?;
    info!("Connected to Redis successfully");
    
    // Initialize eBPF manager
    let mut ebpf_manager = EbpfManager::new(args.interface.clone()).await
        .context("Failed to initialize eBPF manager")?;
    
    // Load and attach XDP program
    ebpf_manager.attach()
        .context("Failed to attach eBPF program")?;
    
    // Track eBPF attachment status
    let ebpf_attached = Arc::new(AtomicBool::new(true));
    
    // Initial sync from Redis to eBPF maps
    info!("Performing initial sync from Redis to eBPF maps...");
    ebpf_manager.sync_from_redis(&mut redis_conn).await
        .context("Failed to sync from Redis to eBPF maps")?;
    info!("Initial sync completed");

    // Set up periodic synchronization loop
    let mut ticker = interval(Duration::from_secs(args.sync_interval));
    
    info!("Packet gateway running. Press Ctrl+C to exit.");
    
    // Start HTTP health check server
    let redis_client_arc = Arc::new(redis_client.clone());
    let ebpf_attached_clone = Arc::clone(&ebpf_attached);
    let ebpf_manager_clone = Arc::clone(&ebpf_manager);
    let health_port = args.health_port;
    
    tokio::spawn(async move {
        let app = Router::new()
            .route("/health", get(health_handler))
            .route("/health/ready", get(readiness_handler))
            .route("/health/live", get(liveness_handler))
            .route("/metrics", get(metrics_handler))
            .with_state(redis_client_arc)
            .with_state(ebpf_attached_clone)
            .with_state(ebpf_manager_clone);
        
        let addr = format!("0.0.0.0:{}", health_port);
        info!("Health check server listening on {}", addr);
        
        let listener = tokio::net::TcpListener::bind(&addr).await
            .expect("Failed to bind health check server");
        axum::serve(listener, app).await
            .expect("Failed to start health check server");
    });
    
    // Main synchronization loop with graceful shutdown
    let mut counter = 0u64;
    loop {
        tokio::select! {
            _ = ticker.tick() => {
                // Use batch sync from eBPF to Redis
                if let Err(e) = ebpf_manager.sync_batch_to_redis(&mut redis_conn).await {
                    error!("Failed to batch sync eBPF maps to Redis: {}", e);
                    continue;
                }
                
                // Sync Redis data to eBPF maps (for credit updates, etc.)
                if let Err(e) = ebpf_manager.sync_from_redis(&mut redis_conn).await {
                    error!("Failed to sync from Redis to eBPF maps: {}", e);
                }
                
                counter += 1;
                
                // Log current stats (every 10 iterations to avoid spam)
                if counter % 10 == 0 {
                    if let Ok(stats) = ebpf_manager.get_packet_stats() {
                        let total_bytes: u64 = stats.iter().map(|s| s.bytes).sum();
                        info!("Current tracked IPs: {}, Total bytes processed: {}", stats.len(), total_bytes);
                    }
                    
                    if let Ok(credits) = ebpf_manager.get_credit_info() {
                        let low_credit_users = credits.iter().filter(|c| c.credit < 1000).count();
                        if low_credit_users > 0 {
                            warn!("{} users with low credit balance", low_credit_users);
                        }
                    }
                }
            }
            _ = tokio::signal::ctrl_c() => {
                info!("Received shutdown signal, cleaning up...");
                
                // Mark eBPF as detached
                ebpf_attached.store(false, Ordering::Relaxed);
                
                // Final sync before shutdown
                if let Err(e) = ebpf_manager.sync_batch_to_redis(&mut redis_conn).await {
                    error!("Failed to sync final stats to Redis: {}", e);
                }
                
                // Clean up eBPF maps
                if let Err(e) = ebpf_manager.cleanup_maps() {
                    error!("Failed to clean up eBPF maps: {}", e);
                }
                
                info!("Packet gateway shutdown complete");
                return Ok(());
            }
        }
    }
}
