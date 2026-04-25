use anyhow::{Context, Result};
use redis::AsyncCommands;
use tracing::{debug, warn};
use super::ip_utils::{u32_to_ipv4_string, ipv4_string_to_u32};

/// Handles synchronization between eBPF maps and Redis
pub struct RedisSyncer;

impl RedisSyncer {
    /// Syncs packet stats from eBPF to Redis
    pub async fn sync_packet_stats(
        stats: &[(u32, u64)],
        redis_conn: &mut redis::aio::MultiplexedConnection,
    ) -> Result<()> {
        for (ip, bytes) in stats {
            let ip_str = u32_to_ipv4_string(*ip);
            let key = format!("packet_stats:{}", ip_str);
            let _: () = redis_conn.set(&key, bytes).await
                .context("Failed to sync packet stats to Redis")?;
        }
        Ok(())
    }

    /// Syncs credit info from eBPF to Redis
    pub async fn sync_credit_info(
        credits: &[(u32, i64)],
        redis_conn: &mut redis::aio::MultiplexedConnection,
    ) -> Result<()> {
        for (ip, credit) in credits {
            let ip_str = u32_to_ipv4_string(*ip);
            let key = format!("user_credit:{}", ip_str);
            let _: () = redis_conn.set(&key, credit).await
                .context("Failed to sync credit info to Redis")?;
        }
        Ok(())
    }

    /// Syncs credit data from Redis to eBPF
    pub async fn sync_credits_from_redis<F>(
        redis_conn: &mut redis::aio::MultiplexedConnection,
        update_credit: F,
    ) -> Result<()>
    where
        F: Fn(u32, i64) -> Result<()>,
    {
        let keys: Vec<String> = redis_conn.keys("user_credit:*").await
            .context("Failed to get credit keys from Redis")?;

        for key in keys {
            if let Some(ip_part) = key.strip_prefix("user_credit:") {
                let credit: i64 = redis_conn.get(&key).await
                    .context("Failed to get credit from Redis")?;

                if let Ok(ip) = ipv4_string_to_u32(ip_part) {
                    update_credit(ip, credit)?;
                    debug!("Synced credit from Redis for IP {}: {}", ip_part, credit);
                } else {
                    warn!("Invalid IP format in Redis key: {}", key);
                }
            }
        }
        Ok(())
    }

    /// Syncs blocked users from Redis to eBPF
    pub async fn sync_blocked_from_redis<F>(
        redis_conn: &mut redis::aio::MultiplexedConnection,
        block_user: F,
    ) -> Result<()>
    where
        F: Fn(u32) -> Result<()>,
    {
        let blocked_keys: Vec<String> = redis_conn.keys("blocked_user:*").await
            .context("Failed to get blocked user keys from Redis")?;

        for key in blocked_keys {
            if let Some(ip_part) = key.strip_prefix("blocked_user:") {
                if let Ok(ip) = ipv4_string_to_u32(ip_part) {
                    block_user(ip)?;
                    debug!("Synced block status from Redis for IP: {}", ip_part);
                } else {
                    warn!("Invalid IP format in Redis block key: {}", key);
                }
            }
        }
        Ok(())
    }

    /// Syncs a batch entry to Redis
    pub async fn sync_batch_entry(
        ip: u32,
        bytes: u64,
        credit: i64,
        blocked: u8,
        redis_conn: &mut redis::aio::MultiplexedConnection,
    ) -> Result<()> {
        let ip_str = u32_to_ipv4_string(ip);

        // Sync packet stats
        let stats_key = format!("packet_stats:{}", ip_str);
        let _: () = redis_conn.set(&stats_key, bytes).await
            .context("Failed to sync packet stats to Redis")?;

        // Sync credit info
        let credit_key = format!("user_credit:{}", ip_str);
        let _: () = redis_conn.set(&credit_key, credit).await
            .context("Failed to sync credit info to Redis")?;

        // Sync block status
        if blocked == 1 {
            let block_key = format!("blocked_user:{}", ip_str);
            let _: () = redis_conn.set(&block_key, 1u8).await
                .context("Failed to sync block status to Redis")?;
        }

        debug!("Batch synced IP {}: {} bytes, credit {}, blocked {}", 
               ip_str, bytes, credit, blocked);
        Ok(())
    }
}
