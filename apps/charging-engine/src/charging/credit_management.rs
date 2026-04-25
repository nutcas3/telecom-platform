use tracing::{info, warn, debug};

use super::types::{SubscriberAccount, UsageEvent};
use crate::errors::{ChargingError, ChargingResult};

impl crate::charging::ChargingEngine {
    pub async fn get_balance(&self, ip: &str) -> ChargingResult<u64> {
        let mut conn = self.redis_client.get_multiplexed_async_connection().await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

        let key = format!("credit:{}", ip);
        let balance: u64 = redis::AsyncCommands::get(&mut conn, &key).await
            .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

        debug!("Retrieved balance {} for IP: {}", balance, ip);
        Ok(balance)
    }

    pub async fn add_credit(&self, ip: &str, bytes_to_add: u64) -> ChargingResult<u64> {
        let mut conn = self.redis_client.get_multiplexed_async_connection().await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

        let key = format!("credit:{}", ip);
        let new_balance: u64 = redis::AsyncCommands::incr(&mut conn, &key, bytes_to_add).await
            .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

        // Set expiration if this is a new key
        let _: () = redis::AsyncCommands::expire(&mut conn, &key, 86400).await.unwrap_or(());

        info!("Added {} bytes credit to IP: {}, new balance: {}", bytes_to_add, ip, new_balance);
        Ok(new_balance)
    }

    pub async fn deduct_credit(&self, ip: &str, bytes_used: u64) -> ChargingResult<u64> {
        let mut conn = self.redis_client.get_multiplexed_async_connection().await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

        let key = format!("credit:{}", ip);
        
        // Check current balance first
        let current_balance: u64 = redis::AsyncCommands::get(&mut conn, &key).await
            .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;
        
        if current_balance < bytes_used {
            return Err(ChargingError::InsufficientCredit {
                available: current_balance,
                requested: bytes_used,
            });
        }

        let new_balance: u64 = redis::AsyncCommands::decr(&mut conn, &key, bytes_used).await
            .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

        info!("Deducted {} bytes credit from IP: {}, new balance: {}", bytes_used, ip, new_balance);
        Ok(new_balance)
    }

    pub async fn check_credit(&self, ip: &str, bytes_requested: u64) -> ChargingResult<bool> {
        let current_balance = self.get_balance(ip).await?;
        Ok(current_balance >= bytes_requested)
    }

    pub async fn block_user(&self, ip: &str) -> ChargingResult<()> {
        let mut conn = self.redis_client.get_multiplexed_async_connection().await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

        let key = format!("block:{}", ip);
        let blocked: Option<String> = redis::AsyncCommands::get(&mut conn, &key).await
            .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

        if blocked.is_some() {
            return Err(ChargingError::UsageBlocked(format!("User IP {} is already blocked", ip)));
        }

        let _: () = redis::AsyncCommands::set(&mut conn, &key, "blocked").await
            .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

        // Set expiration for block (24 hours)
        let _: () = redis::AsyncCommands::expire(&mut conn, &key, 86400).await.unwrap_or(());

        warn!("User IP: {} has been blocked", ip);
        Ok(())
    }

    pub async fn unblock_user(&self, ip: &str) -> ChargingResult<()> {
        let mut conn = self.redis_client.get_multiplexed_async_connection().await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

        let key = format!("block:{}", ip);
        let _: () = redis::AsyncCommands::del(&mut conn, &key).await
            .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

        info!("User IP: {} has been unblocked", ip);
        Ok(())
    }

    pub async fn is_user_blocked(&self, ip: &str) -> ChargingResult<bool> {
        let mut conn = self.redis_client.get_multiplexed_async_connection().await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

        let key = format!("block:{}", ip);
        let blocked: Option<String> = redis::AsyncCommands::get(&mut conn, &key).await
            .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

        Ok(blocked.is_some())
    }

    pub async fn get_subscriber_account(&self, imsi: &str) -> ChargingResult<Option<SubscriberAccount>> {
        let mut conn = self.redis_client.get_multiplexed_async_connection().await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

        let key = format!("account:{}", imsi);
        let account: Option<SubscriberAccount> = redis::AsyncCommands::get(&mut conn, &key).await
            .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

        info!("Retrieved subscriber account for IMSI: {}", imsi);
        Ok(account)
    }

    pub async fn update_subscriber_account(&self, account: &SubscriberAccount) -> ChargingResult<()> {
        let mut conn = self.redis_client.get_multiplexed_async_connection().await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

        let key = format!("account:{}", account.imsi);
        let _: () = redis::AsyncCommands::set(&mut conn, &key, account).await
            .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

        debug!("Updated subscriber account for IMSI: {}", account.imsi);
        Ok(())
    }

    pub async fn record_usage_event(&self, event: &UsageEvent) -> ChargingResult<()> {
        let mut conn = self.redis_client.get_multiplexed_async_connection().await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

        let key = format!("usage:{}:{}", event.imsi, event.session_id);
        let _: () = redis::AsyncCommands::set(&mut conn, &key, event).await
            .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

        info!("Recorded usage event for IMSI: {}, session: {}", event.imsi, event.session_id);
        Ok(())
    }
}
