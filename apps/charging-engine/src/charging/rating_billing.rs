use chrono::{Utc, Datelike};
use tracing::{info, debug, warn};

use super::types::{UsageEvent, UsageType, ChargingRule};
use crate::errors::{ChargingError, ChargingResult, validate_amount};

impl super::ChargingEngine {
    pub async fn calculate_usage_cost(&self, event: &UsageEvent) -> ChargingResult<f64> {
        // Get subscriber account to determine rating plan
        let account = self.get_subscriber_account(&event.imsi).await?;
        let _account = account.ok_or_else(|| ChargingError::SubscriberNotFound(event.imsi.clone()))?;

        // Get rating plan from Postgres (simplified - defaults to basic).
        let plan = self
            .get_rating_plan("basic")
            .await?
            .ok_or_else(|| ChargingError::RatingPlanNotFound("basic".to_string()))?;

        let cost = match event.usage_type {
            UsageType::Data => {
                // Use integer arithmetic: bytes to MB = bytes / 1,048,576
                let data_mb = event.volume / 1_048_576;
                let remainder_bytes = event.volume % 1_048_576;
                // Calculate cost: full MBs + partial MB
                let cost_full_mb = data_mb as f64 * plan.data_rate;
                let cost_partial = (remainder_bytes as f64 / 1_048_576.0) * plan.data_rate;
                cost_full_mb + cost_partial
            }
            UsageType::Voice => {
                // Use integer arithmetic: seconds to minutes = seconds / 60
                let voice_minutes = event.volume / 60;
                let remainder_seconds = event.volume % 60;
                // Calculate cost: full minutes + partial minutes
                let cost_full_min = voice_minutes as f64 * plan.voice_rate;
                let cost_partial = (remainder_seconds as f64 / 60.0) * plan.voice_rate;
                cost_full_min + cost_partial
            }
            UsageType::SMS => {
                // SMS rate is per message, no conversion needed
                event.volume as f64 * plan.sms_rate
            }
        };

        validate_amount(cost)?;

        debug!("Calculated cost for IMSI {}: ${:.4}", event.imsi, cost);
        Ok(cost)
    }

    pub async fn rate_usage(&self, mut event: UsageEvent) -> ChargingResult<UsageEvent> {
        event.cost = self.calculate_usage_cost(&event).await?;
        
        // Update rate field
        if event.volume > 0 {
            event.rate = event.cost / event.volume as f64;
        }

        // Record the rated event
        self.record_usage_event(&event).await?;
        
        info!("Rated usage event for IMSI {}: ${:.4}", event.imsi, event.cost);
        Ok(event)
    }

    pub async fn apply_charging_rules(&self, event: &UsageEvent) -> ChargingResult<Vec<ChargingRule>> {
        let mut applied_rules = Vec::new();

        // Rule 1: Check if user has sufficient credit
        if !self.check_credit(&event.imsi, event.volume).await? {
            applied_rules.push(ChargingRule::InsufficientCredit);
            warn!("Insufficient credit for IMSI: {}", event.imsi);
        }

        // Rule 2: Check usage limits
        let account = self.get_subscriber_account(&event.imsi).await?;
        if let Some(account) = account {
            match event.usage_type {
                UsageType::Data => {
                    let new_usage = account.data_used.checked_add(event.volume)
                        .ok_or_else(|| ChargingError::InvalidInput("Data usage would overflow".to_string()))?;
                    if new_usage > account.data_limit {
                        applied_rules.push(ChargingRule::DataLimitExceeded);
                        warn!("Data limit exceeded for IMSI: {}", event.imsi);
                    }
                }
                UsageType::Voice => {
                    let new_usage = account.voice_used.checked_add(event.volume)
                        .ok_or_else(|| ChargingError::InvalidInput("Voice usage would overflow".to_string()))?;
                    if new_usage > account.voice_limit {
                        applied_rules.push(ChargingRule::VoiceLimitExceeded);
                        warn!("Voice limit exceeded for IMSI: {}", event.imsi);
                    }
                }
                UsageType::SMS => {
                    let new_usage = account.sms_used.checked_add(event.volume)
                        .ok_or_else(|| ChargingError::InvalidInput("SMS usage would overflow".to_string()))?;
                    if new_usage > account.sms_limit {
                        applied_rules.push(ChargingRule::SmsLimitExceeded);
                        warn!("SMS limit exceeded for IMSI: {}", event.imsi);
                    }
                }
            }
        }

        // Rule 3: Check if user is blocked
        if self.is_user_blocked(&event.imsi).await? {
            applied_rules.push(ChargingRule::UserBlocked);
            warn!("Blocked user attempted usage: {}", event.imsi);
        }

        if applied_rules.is_empty() {
            applied_rules.push(ChargingRule::Allowed);
        }

        Ok(applied_rules)
    }

    pub async fn process_usage_event(&self, event: UsageEvent) -> ChargingResult<()> {
        // Apply charging rules
        let rules = self.apply_charging_rules(&event).await?;
        
        // Check if usage is allowed
        if rules.contains(&ChargingRule::Blocked) || 
           rules.contains(&ChargingRule::InsufficientCredit) ||
           rules.contains(&ChargingRule::UserBlocked) {
            let rule_strs: Vec<&str> = rules.iter().map(|r| r.as_str()).collect();
            return Err(ChargingError::UsageBlocked(rule_strs.join(", ")));
        }

        // Rate the usage
        let rated_event = self.rate_usage(event).await?;

        // Update subscriber account
        if let Some(mut account) = self.get_subscriber_account(&rated_event.imsi).await? {
            match rated_event.usage_type {
                UsageType::Data => {
                    account.data_used = account.data_used.checked_add(rated_event.volume)
                        .ok_or_else(|| ChargingError::InvalidInput("Data usage would overflow".to_string()))?;
                }
                UsageType::Voice => {
                    account.voice_used = account.voice_used.checked_add(rated_event.volume)
                        .ok_or_else(|| ChargingError::InvalidInput("Voice usage would overflow".to_string()))?;
                }
                UsageType::SMS => {
                    account.sms_used = account.sms_used.checked_add(rated_event.volume)
                        .ok_or_else(|| ChargingError::InvalidInput("SMS usage would overflow".to_string()))?;
                }
            }
            account.last_updated = Utc::now();
            
            self.update_subscriber_account(&account).await?;
        }

        info!("Processed usage event for IMSI: {}", rated_event.imsi);
        Ok(())
    }

    pub async fn generate_invoice(&self, imsi: &str, billing_period: &str) -> ChargingResult<f64> {
        let mut conn = self.redis_client.get_multiplexed_async_connection().await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

        let pattern = format!("usage:{}:*", imsi);
        let mut keys: Vec<String> = Vec::new();
        
        // Use SCAN instead of KEYS to avoid blocking Redis
        let mut cursor = 0u64;
        loop {
            let (next_cursor, batch_keys): (u64, Vec<String>) = redis::cmd("SCAN")
                .arg(cursor)
                .arg("MATCH")
                .arg(&pattern)
                .arg("COUNT")
                .arg(100)
                .query_async(&mut conn)
                .await
                .map_err(|e| ChargingError::RedisOperation(e.to_string()))?;
            
            keys.extend(batch_keys);
            cursor = next_cursor;
            
            if cursor == 0 {
                break;
            }
        }

        // Parse billing period (format: YYYY-MM)
        let (year, month) = if billing_period.len() >= 7 {
            let parts: Vec<&str> = billing_period.split('-').collect();
            let year = parts[0].parse::<i32>().unwrap_or(0);
            let month = parts.get(1).and_then(|m| m.parse::<u32>().ok()).unwrap_or(0);
            (year, month)
        } else {
            return Err(ChargingError::InvalidInput("Invalid billing period format. Use YYYY-MM".to_string()));
        };

        let mut total_cost = 0.0;
        let mut usage_count = 0;

        for key in keys {
            if let Ok(Some(event)) = redis::AsyncCommands::get::<_, Option<UsageEvent>>(&mut conn, &key).await {
                // Filter by billing period
                if event.timestamp.year() == year && event.timestamp.month() == month {
                    total_cost += event.cost;
                    usage_count += 1;
                }
            }
        }

        info!("Generated invoice for IMSI {}: ${:.4} from {} usage events (period: {})", 
              imsi, total_cost, usage_count, billing_period);
        
        Ok(total_cost)
    }

    pub async fn apply_monthly_fees(&self) -> ChargingResult<u32> {
        let mut conn = self.redis_client.get_multiplexed_async_connection().await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

        let pattern = "account:*".to_string();
        let mut keys: Vec<String> = Vec::new();
        
        // Use SCAN instead of KEYS to avoid blocking Redis
        let mut cursor = 0u64;
        loop {
            let (next_cursor, batch_keys): (u64, Vec<String>) = redis::cmd("SCAN")
                .arg(cursor)
                .arg("MATCH")
                .arg(&pattern)
                .arg("COUNT")
                .arg(100)
                .query_async(&mut conn)
                .await
                .map_err(|e| ChargingError::RedisOperation(e.to_string()))?;
            
            keys.extend(batch_keys);
            cursor = next_cursor;
            
            if cursor == 0 {
                break;
            }
        }

        let mut processed_accounts = 0;

        for key in keys {
            if let Ok(Some(mut account)) = redis::AsyncCommands::get::<_, Option<crate::charging::types::SubscriberAccount>>(&mut conn, &key).await {
                // Get rating plan from Postgres and apply monthly fee.
                if let Ok(Some(plan)) = self.get_rating_plan("basic").await {
                    // Use rounding to avoid precision loss when converting to cents
                    let fee_cents = (plan.monthly_fee * 100.0).round() as i64;
                    account.balance = account.balance.saturating_sub(fee_cents);
                    account.last_updated = Utc::now();

                    let _: () = redis::AsyncCommands::set(&mut conn, &key, &account).await.unwrap_or(());
                    processed_accounts += 1;
                }
            }
        }

        info!("Applied monthly fees to {} accounts", processed_accounts);
        Ok(processed_accounts)
    }
}
