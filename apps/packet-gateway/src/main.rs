use anyhow::Result;
use clap::Parser;
use redis::AsyncCommands;
use tokio::time::{interval, Duration};
use tracing::info;
use axum::{
    routing::get,
    Router,
};
use std::sync::Arc;
use std::sync::atomic::{AtomicBool, Ordering};
use tokio::sync::RwLock;

#[cfg(feature = "ebpf")]
mod ebpf;
mod health;
mod charging_client;
mod config;

#[cfg(feature = "ebpf")]
use ebpf::EbpfManager;
use health::{health_handler, liveness_handler, readiness_handler, metrics_handler};
use charging_client::ChargingEngineClient;
use config::{ConfigState, PacketGatewayConfig, create_config_router};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    /// Network interface to attach XDP program
    #[arg(short, long, default_value = "eth0")]
    interface: String,

    /// Redis connection URL
    #[arg(short, long, default_value = "redis://127.0.0.1/")]
    redis_url: String,

    /// Charging Engine URL for usage reporting
    #[arg(long, default_value = "http://localhost:3001")]
    charging_engine_url: String,

    /// Stats sync interval in seconds
    #[arg(short, long, default_value = "1")]
    sync_interval: u64,

    /// Health check HTTP port
    #[arg(long, default_value = "8081")]
    health_port: u16,
}

#[tokio::main]
async fn main() -> Result<()> {
    // Initialize tracing with centralized logging support
    let log_format = std::env::var("LOG_FORMAT").unwrap_or_else(|_| "compact".to_string());
    let log_json = std::env::var("LOG_JSON").unwrap_or_else(|_| "false".to_string()) == "true";
    
    let subscriber = tracing_subscriber::fmt()
        .with_target(true)
        .with_env_filter(
            tracing_subscriber::EnvFilter::from_default_env()
                .add_directive(tracing::Level::INFO.into())
        );
    
    if log_json {
        subscriber.json().init();
    } else if log_format == "pretty" {
        subscriber.pretty().init();
    } else {
        subscriber.compact().init();
    }
    
    info!("Packet gateway logging initialized (format: {}, json: {})", log_format, log_json);

    dotenv::dotenv().ok();
    let args = Args::parse();

    info!("Starting Packet Gateway");
    info!("Interface: {}", args.interface);
    info!("Redis URL: {}", args.redis_url);
    info!("Charging Engine URL: {}", args.charging_engine_url);
    info!("Health Port: {}", args.health_port);

    // Connect to Redis
    let redis_client = redis::Client::open(args.redis_url.as_str())?;
    let mut redis_conn = redis_client.get_multiplexed_async_connection().await?;
    
    // Test connection
    let _: () = redis::cmd("PING").query_async(&mut redis_conn).await?;
    info!("Connected to Redis successfully");
    
    // Initialize charging engine client
    let charging_client = ChargingEngineClient::new(args.charging_engine_url.clone());
    info!("Charging engine client initialized");
    
    #[cfg(feature = "ebpf")]
    let (ebpf_manager, ebpf_attached) = {
        // Initialize eBPF manager
        let mut ebpf_manager = EbpfManager::new(&args.interface).await
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
        
        (Arc::new(ebpf_manager), ebpf_attached)
    };
    
    #[cfg(not(feature = "ebpf"))]
    let ebpf_attached = Arc::new(AtomicBool::new(false));

    // Set up periodic synchronization loop
    let mut ticker = interval(Duration::from_secs(args.sync_interval));
    let charging_client_arc = Arc::new(charging_client);
    
    info!("Packet gateway running. Press Ctrl+C to exit.");
    
    // Start HTTP health check server
    let redis_client_arc = Arc::new(redis_client);
    let ebpf_attached_clone = Arc::clone(&ebpf_attached);
    
    #[cfg(feature = "ebpf")]
    let ebpf_manager_clone = Arc::clone(&ebpf_manager);
    
    let health_port = args.health_port;

    // Create config state
    let config = PacketGatewayConfig {
        interface: args.interface,
        redis_url: args.redis_url,
        charging_engine_url: args.charging_engine_url,
        sync_interval: args.sync_interval,
        health_port: args.health_port,
    };
    let config_state = ConfigState {
        config: Arc::new(RwLock::new(config)),
    };

    tokio::spawn(async move {
        #[cfg(feature = "ebpf")]
        let app = {
            Router::new()
                .route("/health", get(health_handler))
                .route("/health/ready", get(readiness_handler))
                .route("/health/live", get(liveness_handler))
                .route("/metrics", get(metrics_handler))
                .with_state((redis_client_arc, ebpf_attached_clone, ebpf_manager_clone))
        };
        
        #[cfg(not(feature = "ebpf"))]
        let app = {
            Router::new()
                .route("/health", get(health_handler))
                .route("/health/live", get(liveness_handler))
                .route("/metrics", get(metrics_handler))
                .with_state((redis_client_arc, ebpf_attached_clone))
        };

        let addr = format!("0.0.0.0:{}", health_port);
        info!("Health check server listening on {}", addr);

        let listener = tokio::net::TcpListener::bind(&addr).await
            .expect("Failed to bind health check server");
        axum::serve(listener, app).await
            .expect("Failed to start health check server");
    });
    
    // Main synchronization loop with graceful shutdown
    let mut counter = 0u64;
    let mut last_usage_report = 0u64;
    let usage_report_interval = 60; // Report usage every 60 seconds
    
    #[cfg(feature = "ebpf")]
    let ebpf_manager_ref = Arc::clone(&ebpf_manager);
    
    loop {
        tokio::select! {
            _ = ticker.tick() => {
                #[cfg(feature = "ebpf")]
                {
                    // Use batch sync from eBPF to Redis
                    if let Err(e) = ebpf_manager_ref.sync_batch_to_redis(&mut redis_conn).await {
                        error!("Failed to batch sync eBPF maps to Redis: {}", e);
                        continue;
                    }
                    
                    // Sync Redis data to eBPF maps (for credit updates, etc.)
                    if let Err(e) = ebpf_manager_ref.sync_from_redis(&mut redis_conn).await {
                        error!("Failed to sync from Redis to eBPF maps: {}", e);
                    }
                }
                
                counter += 1;
                
                #[cfg(feature = "ebpf")]
                {
                    // Report usage to charging-engine periodically
                    let elapsed = counter * args.sync_interval;
                    if elapsed - last_usage_report >= usage_report_interval {
                        if let Ok(stats) = ebpf_manager_ref.get_packet_stats() {
                            let usage_data: Vec<(String, String, u64)> = stats.iter()
                                .filter(|s| s.bytes > 0)
                                .map(|s| {
                                    let ip_str = format!("{}.{}.{}.{}", 
                                        (s.ip >> 24) & 0xFF,
                                        (s.ip >> 16) & 0xFF,
                                        (s.ip >> 8) & 0xFF,
                                        s.ip & 0xFF
                                    );
                                    let session_id = format!("session-{}", ip_str);
                                    (ip_str, session_id, s.bytes)
                                })
                                .collect();
                            
                            if !usage_data.is_empty() {
                                let charging_client = Arc::clone(&charging_client_arc);
                                tokio::spawn(async move {
                                    if let Err(e) = charging_client.report_batch_usage(usage_data).await {
                                        warn!("Failed to report batch usage to charging engine: {}", e);
                                    }
                                });
                                last_usage_report = elapsed;
                            }
                        }
                    }
                    
                    // Log current stats (every 10 iterations to avoid spam)
                    if counter % 10 == 0 {
                        if let Ok(stats) = ebpf_manager_ref.get_packet_stats() {
                            let total_bytes: u64 = stats.iter().map(|s| s.bytes).sum();
                            info!("Current tracked IPs: {}, Total bytes processed: {}", stats.len(), total_bytes);
                        }
                        
                        if let Ok(credits) = ebpf_manager_ref.get_credit_info() {
                            let low_credit_users = credits.iter().filter(|c| c.credit < 1000).count();
                            if low_credit_users > 0 {
                                warn!("{} users with low credit balance", low_credit_users);
                            }
                        }
                    }
                }
                
                #[cfg(not(feature = "ebpf"))]
                {
                    // Without eBPF, just log that we're running
                    if counter % 10 == 0 {
                        info!("Packet gateway running without eBPF support (counter: {})", counter);
                    }
                }
            }
            _ = tokio::signal::ctrl_c() => {
                info!("Received shutdown signal, cleaning up...");
                
                // Mark eBPF as detached
                ebpf_attached.store(false, Ordering::Relaxed);
                
                #[cfg(feature = "ebpf")]
                {
                    // Final sync before shutdown
                    if let Err(e) = ebpf_manager_ref.sync_batch_to_redis(&mut redis_conn).await {
                        error!("Failed to sync final stats to Redis: {}", e);
                    }
                    
                    // Clean up eBPF maps
                    if let Err(e) = ebpf_manager_ref.cleanup_maps() {
                        error!("Failed to clean up eBPF maps: {}", e);
                    }
                }
                
                info!("Packet gateway shutdown complete");
                return Ok(());
            }
        }
    }
}
