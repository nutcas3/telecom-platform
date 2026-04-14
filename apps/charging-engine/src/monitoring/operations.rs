use tokio::time::interval;
use tracing::{info, warn, error, debug};
use std::time::Duration;

use crate::errors::{ChargingError, ChargingResult, ErrorContext};

impl super::ChargingEngine {
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
        debug!("Performing background sync");
        
        // Sync 1: Clean up expired blocks
        self.cleanup_expired_blocks().await?;
        
        // Sync 2: Update usage statistics
        self.update_usage_statistics().await?;
        
        // Sync 3: Check for low balance alerts
        self.check_low_balance_alerts().await?;
        
        // Sync 4: Apply monthly fees (once per day)
        self.apply_monthly_fees_if_needed().await?;
        
        debug!("Background sync completed successfully");
        Ok(())
    }

    async fn cleanup_expired_blocks(&self) -> ChargingResult<()> {
        let mut conn = self.redis_client.get_async_connection().await
            .context("Failed to get Redis connection")?;

        let pattern = "block:*".to_string();
        let keys: Vec<String> = conn.keys(&pattern).await
            .context("Failed to get block keys")?;

        let mut cleaned_count = 0;

        for key in keys {
            let ttl: i64 = conn.ttl(&key).await.unwrap_or(-1);
            if ttl == -1 || ttl == -2 {
                // Key has no expiration or doesn't exist
                let _: () = conn.del(&key).await.unwrap_or(());
                cleaned_count += 1;
            }
        }

        if cleaned_count > 0 {
            info!("Cleaned up {} expired blocks", cleaned_count);
        }

        Ok(())
    }

    async fn update_usage_statistics(&self) -> ChargingResult<()> {
        let mut conn = self.redis_client.get_async_connection().await
            .context("Failed to get Redis connection")?;

        // Get total active sessions
        let pattern = "session:*".to_string();
        let keys: Vec<String> = conn.keys(&pattern).await
            .context("Failed to get session keys")?;

        let active_sessions = keys.len();
        
        // Store statistics
        let stats_key = "stats:active_sessions";
        let _: () = conn.set(stats_key, active_sessions).await.unwrap_or(());

        debug!("Updated usage statistics: {} active sessions", active_sessions);
        Ok(())
    }

    async fn check_low_balance_alerts(&self) -> ChargingResult<()> {
        let mut conn = self.redis_client.get_async_connection().await
            .context("Failed to get Redis connection")?;

        let pattern = "credit:*".to_string();
        let keys: Vec<String> = conn.keys(&pattern).await
            .context("Failed to get credit keys")?;

        let low_balance_threshold = 100_000_000; // 100MB in bytes
        let mut low_balance_count = 0;

        for key in keys {
            if let Ok(balance) = conn.get::<_, u64>(&key).await {
                if balance < low_balance_threshold {
                    low_balance_count += 1;
                    warn!("Low balance alert for {}: {} bytes remaining", key, balance);
                    
                    // Store alert
                    let alert_key = format!("alert:low_balance:{}", key);
                    let _: () = conn.setex(&alert_key, 3600, "low_balance").await.unwrap_or(());
                }
            }
        }

        if low_balance_count > 0 {
            info!("Generated {} low balance alerts", low_balance_count);
        }

        Ok(())
    }

    async fn apply_monthly_fees_if_needed(&self) -> ChargingResult<()> {
        let mut conn = self.redis_client.get_async_connection().await
            .context("Failed to get Redis connection")?;

        let last_fee_key = "last_monthly_fee";
        let last_fee: Option<String> = conn.get(last_fee_key).await.unwrap_or(None);

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
        let _: () = conn.set(last_fee_key, current_month).await.unwrap_or(());

        info!("Applied monthly fees to {} accounts for {}", processed, current_month);
        Ok(())
    }
}
