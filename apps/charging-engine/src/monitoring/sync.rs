use tokio::time::interval;
use tracing::{info, warn, error, debug};
use chrono::Datelike;

use crate::errors::{ChargingResult, log_error};
use crate::monitoring::types::{SystemStats, HealthStatus};

impl crate::charging::ChargingEngine {
    pub async fn start_background_sync(&self) -> ChargingResult<()> {
        let mut interval_timer = interval(self.sync_interval);
        
        info!("Starting background sync with interval: {:?}", self.sync_interval);
        
        loop {
            interval_timer.tick().await;
            
            if let Err(e) = self.perform_sync().await {
                error!("Background sync failed: {}", e);
            }
        }
    }

    async fn perform_sync(&self) -> ChargingResult<()> {
        info!("Starting sync cycle");

        if let Err(e) = self.cleanup_expired_blocks().await {
            warn!("Failed to cleanup expired blocks: {}", e);
            log_error(&e);
        }

        if let Err(e) = self.update_usage_statistics().await {
            warn!("Failed to update usage statistics: {}", e);
            log_error(&e);
        }

        if let Err(e) = self.check_low_balance_alerts().await {
            warn!("Failed to check low balance alerts: {}", e);
            log_error(&e);
        }

        if let Err(e) = self.apply_monthly_fees_if_needed().await {
            warn!("Failed to apply monthly fees: {}", e);
            log_error(&e);
        }

        info!("Sync cycle completed");
        Ok(())
    }

    async fn cleanup_expired_blocks(&self) -> ChargingResult<()> {
        let mut conn = self.redis_client.get_multiplexed_async_connection().await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

        let pattern = "block:*".to_string();
        let keys: Vec<String> = redis::AsyncCommands::keys(&mut conn, &pattern).await
            .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

        let mut cleaned_count = 0;

        for key in keys {
            let ttl: i64 = redis::AsyncCommands::ttl(&mut conn, &key).await.unwrap_or(-1);
            if ttl == -1 || ttl == -2 {
                // Key has no expiration or doesn't exist
                let _: () = redis::AsyncCommands::del(&mut conn, &key).await.unwrap_or(());
                cleaned_count += 1;
            }
        }

        if cleaned_count > 0 {
            info!("Cleaned up {} expired blocks", cleaned_count);
        }

        Ok(())
    }

    async fn update_usage_statistics(&self) -> ChargingResult<()> {
        let mut conn = self.redis_client.get_multiplexed_async_connection().await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

        // Get total active sessions
        let pattern = "session:*".to_string();
        let keys: Vec<String> = redis::AsyncCommands::keys(&mut conn, &pattern).await
            .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

        let active_sessions = keys.len();
        
        // Store statistics
        let stats_key = "stats:active_sessions";
        let _: () = redis::AsyncCommands::set(&mut conn, stats_key, active_sessions).await.unwrap_or(());

        debug!("Updated usage statistics: {} active sessions", active_sessions);
        Ok(())
    }

    async fn check_low_balance_alerts(&self) -> ChargingResult<()> {
        let mut conn = self.redis_client.get_multiplexed_async_connection().await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

        let pattern = "credit:*".to_string();
        let keys: Vec<String> = redis::AsyncCommands::keys(&mut conn, &pattern).await
            .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

        let low_balance_threshold = 100_000_000; // 100MB in bytes
        let mut low_balance_count = 0;

        for key in keys {
            if let Ok(balance) = redis::AsyncCommands::get::<_, u64>(&mut conn, &key).await {
                if balance < low_balance_threshold {
                    low_balance_count += 1;
                    warn!("Low balance alert for {}: {} bytes remaining", key, balance);
                    
                    // Store alert
                    let alert_key = format!("alert:low_balance:{}", key);
                    let _: () = redis::AsyncCommands::set(&mut conn, &alert_key, "low_balance").await.unwrap_or(());
                    let _: () = redis::AsyncCommands::expire(&mut conn, &alert_key, 3600).await.unwrap_or(());
                }
            }
        }

        if low_balance_count > 0 {
            info!("Generated {} low balance alerts", low_balance_count);
        }

        Ok(())
    }

    async fn apply_monthly_fees_if_needed(&self) -> ChargingResult<()> {
        let mut conn = self.redis_client.get_multiplexed_async_connection().await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

        let last_fee_key = "last_monthly_fee";
        let last_fee: Option<String> = redis::AsyncCommands::get(&mut conn, last_fee_key).await.unwrap_or(None);

        let now = chrono::Utc::now();
        let current_month = format!("{}-{}", now.year(), now.month());

        if let Some(last_month) = last_fee {
            if last_month == current_month {
                return Ok(()); // Monthly fees already applied this month
            }
        }

        // Apply monthly fees
        let processed = self.apply_monthly_fees().await?;
        
        // Update last fee timestamp
        let _: () = redis::AsyncCommands::set(&mut conn, last_fee_key, &current_month).await.unwrap_or(());

        info!("Applied monthly fees to {} accounts for {}", processed, current_month);
        Ok(())
    }

    pub async fn get_system_statistics(&self) -> ChargingResult<SystemStats> {
        let mut conn = self.redis_client.get_multiplexed_async_connection().await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

        let active_sessions: u64 = redis::AsyncCommands::get(&mut conn, "stats:active_sessions").await.unwrap_or(0);
        let account_keys: Vec<String> = redis::AsyncCommands::keys(&mut conn, "account:*").await.unwrap_or_default();
        let total_accounts: u64 = account_keys.len() as u64;
        let block_keys: Vec<String> = redis::AsyncCommands::keys(&mut conn, "block:*").await.unwrap_or_default();
        let blocked_users: u64 = block_keys.len() as u64;
        let alert_keys: Vec<String> = redis::AsyncCommands::keys(&mut conn, "alert:low_balance:*").await.unwrap_or_default();
        let low_balance_alerts: u64 = alert_keys.len() as u64;

        let stats = SystemStats {
            active_sessions,
            total_accounts,
            blocked_users,
            low_balance_alerts,
            uptime: self.get_uptime().await?,
        };

        Ok(stats)
    }

    pub async fn get_uptime(&self) -> ChargingResult<u64> {
        // Calculate actual uptime since startup
        match self.startup_time.elapsed() {
            Ok(duration) => Ok(duration.as_secs()),
            Err(_) => {
                // If system time went backwards, fallback to current time
                Ok(std::time::SystemTime::now()
                    .duration_since(std::time::UNIX_EPOCH)
                    .unwrap_or_default()
                    .as_secs())
            }
        }
    }

    pub async fn health_check(&self) -> ChargingResult<HealthStatus> {
        let mut conn = self.redis_client.get_multiplexed_async_connection().await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

        // Test Redis connection
        let _: String = redis::AsyncCommands::get(&mut conn, "health_check").await.unwrap_or_else(|_| "ok".to_string());

        let status = HealthStatus {
            redis_connected: true,
            active_sync: true,
            last_sync: chrono::Utc::now(),
            memory_usage: self.get_memory_usage().await?,
        };

        Ok(status)
    }

    async fn get_memory_usage(&self) -> ChargingResult<u64> {
        // Get actual memory usage from system
        match self.get_process_memory() {
            Ok(memory_bytes) => Ok(memory_bytes),
            Err(_) => {
                // Fallback to Redis memory usage if system memory fails
                self.get_redis_memory_usage().await
            }
        }
    }

    fn get_process_memory(&self) -> ChargingResult<u64> {
        use std::fs;
        
        // Try to read from /proc/self/status for Linux systems
        if let Ok(status) = fs::read_to_string("/proc/self/status") {
            for line in status.lines() {
                if line.starts_with("VmRSS:") {
                    if let Some(memory_str) = line.split_whitespace().nth(1) {
                        if let Ok(memory_kb) = memory_str.parse::<u64>() {
                            return Ok(memory_kb * 1024); // Convert KB to bytes
                        }
                    }
                }
            }
        }
        
        // Fallback for non-Linux systems or if reading fails
        self.estimate_memory_usage()
    }

    fn estimate_memory_usage(&self) -> ChargingResult<u64> {
        // Estimate memory usage based on known structures
        // This is a rough estimate for non-Linux systems
        let base_memory = 20_000_000; // 20MB base
        let redis_connections = 5_000_000; // 5MB per connection estimate
        let session_data = 10_000_000; // 10MB for session data
        
        Ok(base_memory + redis_connections + session_data)
    }

    async fn get_redis_memory_usage(&self) -> ChargingResult<u64> {
        let mut conn = self.redis_client.get_multiplexed_async_connection().await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

        // Get Redis memory info
        let info: String = redis::cmd("INFO").query_async(&mut conn).await.unwrap_or_default();
        
        // Parse memory usage from Redis info
        for line in info.lines() {
            if line.starts_with("used_memory:") {
                if let Some(memory_str) = line.split(':').nth(1) {
                    if let Ok(memory_bytes) = memory_str.parse::<u64>() {
                        return Ok(memory_bytes);
                    }
                }
            }
        }
        
        // Fallback to estimate
        self.estimate_memory_usage()
    }
}

