use tracing::{info, warn, debug};

use super::types::{SubscriberAccount, UsageEvent};
use crate::circuit_breaker::CircuitBreakerError;
use crate::errors::{ChargingError, ChargingResult};

impl crate::charging::ChargingEngine {
    pub async fn get_balance(&self, ip: &str) -> ChargingResult<u64> {
        self.redis_circuit_breaker.execute(|| async {
            let mut conn = self.redis_client.get_multiplexed_async_connection().await
                .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

            let key = format!("credit:{}", ip);
            let balance: u64 = redis::AsyncCommands::get(&mut conn, &key).await
                .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

            info!("Retrieved balance for IP: {}, balance: {} bytes", ip, balance);
            Ok(balance)
        }).await.map_err(|e| match e {
            CircuitBreakerError::Open => ChargingError::RedisConnection("Circuit breaker is open".to_string()),
            CircuitBreakerError::Inner(e) => e,
        })
    }

    pub async fn add_credit(&self, ip: &str, bytes_to_add: u64) -> ChargingResult<u64> {
        self.redis_circuit_breaker.execute(|| async {
            let mut conn = self.redis_client.get_multiplexed_async_connection().await
                .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

            let key = format!("credit:{}", ip);
            
            // Check for potential overflow before incrementing
            let current_balance: Option<u64> = redis::AsyncCommands::get(&mut conn, &key).await
                .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;
            
            if let Some(existing_balance) = current_balance {
                if existing_balance.checked_add(bytes_to_add).is_none() {
                    return Err(ChargingError::InvalidInput("Credit addition would overflow".to_string()));
                }
            }
            
            let new_balance: u64 = redis::AsyncCommands::incr(&mut conn, &key, bytes_to_add).await
                .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

            // Set expiration if this is a new key (configurable via CREDIT_EXPIRATION_SECONDS, defaults to 24 hours)
            let credit_expiration: u64 = std::env::var("CREDIT_EXPIRATION_SECONDS")
                .ok()
                .and_then(|s| s.parse().ok())
                .unwrap_or(86400);
            let _: () = redis::AsyncCommands::expire(&mut conn, &key, credit_expiration as i64).await.unwrap_or(());

            info!("Added {} bytes credit to IP: {}, new balance: {}", bytes_to_add, ip, new_balance);
            Ok(new_balance)
        }).await.map_err(|e| match e {
            CircuitBreakerError::Open => ChargingError::RedisConnection("Circuit breaker is open".to_string()),
            CircuitBreakerError::Inner(e) => e,
        })
    }

    pub async fn deduct_credit(&self, ip: &str, bytes_used: u64) -> ChargingResult<u64> {
        self.redis_circuit_breaker.execute(|| async {
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
        }).await.map_err(|e| match e {
            CircuitBreakerError::Open => ChargingError::RedisConnection("Circuit breaker is open".to_string()),
            CircuitBreakerError::Inner(e) => e,
        })
    }

    pub async fn check_credit(&self, ip: &str, bytes_requested: u64) -> ChargingResult<bool> {
        let current_balance = self.get_balance(ip).await?;
        Ok(current_balance >= bytes_requested)
    }

    pub async fn block_user(&self, ip: &str) -> ChargingResult<()> {
        self.redis_circuit_breaker.execute(|| async {
            let mut conn = self.redis_client.get_multiplexed_async_connection().await
                .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

            let key = format!("block:{}", ip);
            let blocked: Option<String> = redis::AsyncCommands::get(&mut conn, &key).await
                .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

            if let Some(_blocked) = blocked {
                return Err(ChargingError::UsageBlocked(format!("User IP {} is already blocked", ip)));
            }

            let _: () = redis::AsyncCommands::set(&mut conn, &key, "blocked").await
                .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

            // Set expiration for block (configurable via BLOCK_EXPIRATION_SECONDS, defaults to 24 hours)
            let block_expiration: u64 = std::env::var("BLOCK_EXPIRATION_SECONDS")
                .ok()
                .and_then(|s| s.parse().ok())
                .unwrap_or(86400);
            let _: () = redis::AsyncCommands::expire(&mut conn, &key, block_expiration as i64).await.unwrap_or(());

            warn!("User IP: {} has been blocked for {} seconds", ip, block_expiration);
            Ok(())
        }).await.map_err(|e| match e {
            CircuitBreakerError::Open => ChargingError::RedisConnection("Circuit breaker is open".to_string()),
            CircuitBreakerError::Inner(e) => e,
        })
    }

    pub async fn unblock_user(&self, ip: &str) -> ChargingResult<()> {
        self.redis_circuit_breaker.execute(|| async {
            let mut conn = self.redis_client.get_multiplexed_async_connection().await
                .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

            // Check if user is blocked
            let blocked_key = format!("blocked:{}", ip);
            let blocked: Option<String> = redis::AsyncCommands::get(&mut conn, &blocked_key).await
                .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;
            
            if let Some(_blocked) = blocked {

                let _: () = redis::AsyncCommands::del(&mut conn, &blocked_key).await
                    .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

                info!("User IP: {} has been unblocked", ip);
                Ok(())
            } else {
                Ok(())
            }
        }).await.map_err(|e| match e {
            CircuitBreakerError::Open => ChargingError::RedisConnection("Circuit breaker is open".to_string()),
            CircuitBreakerError::Inner(e) => e,
        })
    }

    pub async fn is_user_blocked(&self, ip: &str) -> ChargingResult<bool> {
        self.redis_circuit_breaker.execute(|| async {
            let mut conn = self.redis_client.get_multiplexed_async_connection().await
                .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

            let key = format!("block:{}", ip);
            let blocked: Option<String> = redis::AsyncCommands::get(&mut conn, &key).await
                .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

            Ok(blocked.is_some())
        }).await.map_err(|e| match e {
            CircuitBreakerError::Open => ChargingError::RedisConnection("Circuit breaker is open".to_string()),
            CircuitBreakerError::Inner(e) => e,
        })
    }

    pub async fn get_subscriber_account(&self, imsi: &str) -> ChargingResult<Option<SubscriberAccount>> {
        self.redis_circuit_breaker.execute(|| async {
            let mut conn = self.redis_client.get_multiplexed_async_connection().await
                .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

            let key = format!("account:{}", imsi);
            let account: Option<SubscriberAccount> = redis::AsyncCommands::get(&mut conn, &key).await
                .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

            info!("Retrieved subscriber account for IMSI: {}", imsi);
            Ok(account)
        }).await.map_err(|e| match e {
            CircuitBreakerError::Open => ChargingError::RedisConnection("Circuit breaker is open".to_string()),
            CircuitBreakerError::Inner(e) => e,
        })
    }

    pub async fn update_subscriber_account(&self, account: &SubscriberAccount) -> ChargingResult<()> {
        self.redis_circuit_breaker.execute(|| async {
            let mut conn = self.redis_client.get_multiplexed_async_connection().await
                .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

            let key = format!("account:{}", account.imsi);
            let _: () = redis::AsyncCommands::set(&mut conn, &key, account).await
                .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

            debug!("Updated subscriber account for IMSI: {}", account.imsi);
            Ok(())
        }).await.map_err(|e| match e {
            CircuitBreakerError::Open => ChargingError::RedisConnection("Circuit breaker is open".to_string()),
            CircuitBreakerError::Inner(e) => e,
        })
    }

    pub async fn record_usage_event(&self, event: &UsageEvent) -> ChargingResult<()> {
        self.redis_circuit_breaker.execute(|| async {
            let mut conn = self.redis_client.get_multiplexed_async_connection().await
                .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

            let key = format!("usage:{}:{}", event.imsi, event.session_id);
            let _: () = redis::AsyncCommands::set(&mut conn, &key, event).await
                .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

            info!("Recorded usage event for IMSI: {}, session: {}", event.imsi, event.session_id);
            Ok(())
        }).await.map_err(|e| match e {
            CircuitBreakerError::Open => ChargingError::RedisConnection("Circuit breaker is open".to_string()),
            CircuitBreakerError::Inner(e) => e,
        })
    }
}
