use redis::AsyncCommands;
use tracing::{info, warn, error, debug};

use super::types::{SubscriberAccount, UsageEvent};
use crate::errors::{ChargingError, ChargingResult, ErrorContext};

impl super::ChargingEngine {
    pub async fn get_balance(&self, ip: &str) -> ChargingResult<u64> {
        let mut conn = self.redis_client.get_async_connection().await
            .context("Failed to get Redis connection")?;

        let key = format!("credit:{}", ip);
        let balance: u64 = conn.get(&key).await
            .context("Failed to get balance from Redis")?;

        debug!("Retrieved balance {} for IP: {}", balance, ip);
        Ok(balance)
    }

    pub async fn add_credit(&self, ip: &str, bytes_to_add: u64) -> ChargingResult<u64> {
        let mut conn = self.redis_client.get_async_connection().await
            .context("Failed to get Redis connection")?;

        let key = format!("credit:{}", ip);
        let new_balance: u64 = conn.incr(&key, bytes_to_add).await
            .context("Failed to add credit in Redis")?;

        // Set expiration if this is a new key
        let _: () = conn.expire(&key, 86400).await.unwrap_or(());

        info!("Added {} bytes credit to IP: {}, new balance: {}", bytes_to_add, ip, new_balance);
        Ok(new_balance)
    }

    pub async fn deduct_credit(&self, ip: &str, bytes_used: u64) -> ChargingResult<u64> {
        let mut conn = self.redis_client.get_async_connection().await
            .context("Failed to get Redis connection")?;

        let key = format!("credit:{}", ip);
        
        // Check current balance first
        let current_balance: u64 = conn.get(&key).await.unwrap_or(0);
        
        if current_balance < bytes_used {
            return Err(ChargingError::InsufficientCredit {
                available: current_balance,
                requested: bytes_used,
            });
        }

        let new_balance: u64 = conn.decr(&key, bytes_used).await
            .context("Failed to deduct credit in Redis")?;

        info!("Deducted {} bytes credit from IP: {}, new balance: {}", bytes_used, ip, new_balance);
        Ok(new_balance)
    }

    pub async fn check_credit(&self, ip: &str, bytes_requested: u64) -> ChargingResult<bool> {
        let current_balance = self.get_balance(ip).await.unwrap_or(0);
        Ok(current_balance >= bytes_requested)
    }

    pub async fn block_user(&self, ip: &str) -> ChargingResult<()> {
        let mut conn = self.redis_client.get_async_connection().await
            .context("Failed to get Redis connection")?;

        let key = format!("block:{}", ip);
        let _: () = conn.set(&key, "blocked").await
            .context("Failed to block user in Redis")?;

        // Set expiration for block (24 hours)
        let _: () = conn.expire(&key, 86400).await.unwrap_or(());

        warn!("User IP: {} has been blocked", ip);
        Ok(())
    }

    pub async fn unblock_user(&self, ip: &str) -> ChargingResult<()> {
        let mut conn = self.redis_client.get_async_connection().await
            .context("Failed to get Redis connection")?;

        let key = format!("block:{}", ip);
        let _: () = conn.del(&key).await
            .context("Failed to unblock user in Redis")?;

        info!("User IP: {} has been unblocked", ip);
        Ok(())
    }

    pub async fn is_user_blocked(&self, ip: &str) -> ChargingResult<bool> {
        let mut conn = self.redis_client.get_async_connection().await
            .context("Failed to get Redis connection")?;

        let key = format!("block:{}", ip);
        let blocked: Option<String> = conn.get(&key).await
            .context("Failed to check block status in Redis")?;

        Ok(blocked.is_some())
    }

    pub async fn get_subscriber_account(&self, imsi: &str) -> ChargingResult<Option<SubscriberAccount>> {
        let mut conn = self.redis_client.get_async_connection().await
            .context("Failed to get Redis connection")?;

        let key = format!("account:{}", imsi);
        let account: Option<SubscriberAccount> = conn.get(&key).await
            .context("Failed to get subscriber account from Redis")?;

        Ok(account)
    }

    pub async fn update_subscriber_account(&self, account: &SubscriberAccount) -> ChargingResult<()> {
        let mut conn = self.redis_client.get_async_connection().await
            .context("Failed to get Redis connection")?;

        let key = format!("account:{}", account.imsi);
        let _: () = conn.set(&key, account).await
            .context("Failed to update subscriber account in Redis")?;

        debug!("Updated subscriber account for IMSI: {}", account.imsi);
        Ok(())
    }

    pub async fn record_usage_event(&self, event: &UsageEvent) -> ChargingResult<()> {
        let mut conn = self.redis_client.get_async_connection().await
            .context("Failed to get Redis connection")?;

        let key = format!("usage:{}:{}", event.imsi, event.session_id);
        let _: () = conn.set(&key, event).await
            .context("Failed to record usage event in Redis")?;

        // Set expiration for usage events (7 days)
        let _: () = conn.expire(&key, 604800).await.unwrap_or(());

        info!("Recorded usage event for IMSI: {}, cost: ${:.4}", event.imsi, event.cost);
        Ok(())
    }
}
