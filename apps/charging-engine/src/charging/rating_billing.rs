use chrono::Utc;
use tracing::{info, debug, warn};

use super::types::{UsageEvent, UsageType, RatingPlan};
use crate::errors::{ChargingError, ChargingResult, ErrorContext};

impl super::ChargingEngine {
    pub async fn calculate_usage_cost(&self, event: &UsageEvent) -> ChargingResult<f64> {
        // Get subscriber account to determine rating plan
        let account = self.get_subscriber_account(&event.imsi).await?;
        let account = account.ok_or_else(|| ChargingError::SubscriberNotFound(event.imsi.clone()))?;

        // Get rating plan (simplified - in real implementation this would be more complex)
        let plan = self.get_rating_plan("basic") // Default to basic plan for now
            .ok_or_else(|| ChargingError::RatingPlanNotFound("basic".to_string()))?;

        let cost = match event.usage_type {
            UsageType::Data => {
                let data_mb = event.volume as f64 / 1_048_576.0; // Convert bytes to MB
                data_mb * plan.data_rate
            }
            UsageType::Voice => {
                let voice_minutes = event.volume as f64 / 60.0; // Convert seconds to minutes
                voice_minutes * plan.voice_rate
            }
            UsageType::SMS => {
                event.volume as f64 * plan.sms_rate
            }
        };

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

    pub async fn apply_charging_rules(&self, event: &UsageEvent) -> ChargingResult<Vec<String>> {
        let mut applied_rules = Vec::new();

        // Rule 1: Check if user has sufficient credit
        if !self.check_credit(&event.imsi, event.volume).await? {
            applied_rules.push("INSUFFICIENT_CREDIT".to_string());
            warn!("Insufficient credit for IMSI: {}", event.imsi);
        }

        // Rule 2: Check usage limits
        let account = self.get_subscriber_account(&event.imsi).await?;
        if let Some(account) = account {
            match event.usage_type {
                UsageType::Data => {
                    if account.data_used + event.volume > account.data_limit {
                        applied_rules.push("DATA_LIMIT_EXCEEDED".to_string());
                        warn!("Data limit exceeded for IMSI: {}", event.imsi);
                    }
                }
                UsageType::Voice => {
                    if account.voice_used + event.volume > account.voice_limit {
                        applied_rules.push("VOICE_LIMIT_EXCEEDED".to_string());
                        warn!("Voice limit exceeded for IMSI: {}", event.imsi);
                    }
                }
                UsageType::SMS => {
                    if account.sms_used + event.volume > account.sms_limit {
                        applied_rules.push("SMS_LIMIT_EXCEEDED".to_string());
                        warn!("SMS limit exceeded for IMSI: {}", event.imsi);
                    }
                }
            }
        }

        // Rule 3: Check if user is blocked
        if self.is_user_blocked(&event.imsi).await? {
            applied_rules.push("USER_BLOCKED".to_string());
            warn!("Blocked user attempted usage: {}", event.imsi);
        }

        if applied_rules.is_empty() {
            applied_rules.push("ALLOWED".to_string());
        }

        Ok(applied_rules)
    }

    pub async fn process_usage_event(&self, event: UsageEvent) -> ChargingResult<()> {
        // Apply charging rules
        let rules = self.apply_charging_rules(&event).await?;
        
        // Check if usage is allowed
        if rules.contains(&"BLOCKED".to_string()) || 
           rules.contains(&"INSUFFICIENT_CREDIT".to_string()) ||
           rules.contains(&"USER_BLOCKED".to_string()) {
            return Err(ChargingError::UsageBlocked(rules.join(", ")));
        }

        // Rate the usage
        let rated_event = self.rate_usage(event).await?;

        // Update subscriber account
        if let Some(mut account) = self.get_subscriber_account(&rated_event.imsi).await? {
            match rated_event.usage_type {
                UsageType::Data => account.data_used += rated_event.volume,
                UsageType::Voice => account.voice_used += rated_event.volume,
                UsageType::SMS => account.sms_used += rated_event.volume,
            }
            account.last_updated = Utc::now();
            
            self.update_subscriber_account(&account).await?;
        }

        info!("Processed usage event for IMSI: {}", rated_event.imsi);
        Ok(())
    }

    pub async fn generate_invoice(&self, imsi: &str, billing_period: &str) -> ChargingResult<f64> {
        // Get all usage events for the billing period
        let mut conn = self.redis_client.get_async_connection().await
            .context("Failed to get Redis connection")?;

        let pattern = format!("usage:{}:*", imsi);
        let keys: Vec<String> = conn.keys(&pattern).await
            .context("Failed to get usage keys")?;

        let mut total_cost = 0.0;
        let mut usage_count = 0;

        for key in keys {
            if let Ok(Some(event)) = conn.get::<_, Option<UsageEvent>>(&key).await {
                // Filter by billing period (simplified)
                total_cost += event.cost;
                usage_count += 1;
            }
        }

        info!("Generated invoice for IMSI {}: ${:.4} from {} usage events", 
              imsi, total_cost, usage_count);
        
        Ok(total_cost)
    }

    pub async fn apply_monthly_fees(&self) -> ChargingResult<u32> {
        let mut conn = self.redis_client.get_async_connection().await
            .context("Failed to get Redis connection")?;

        let pattern = "account:*".to_string();
        let keys: Vec<String> = conn.keys(&pattern).await
            .context("Failed to get account keys")?;

        let mut processed_accounts = 0;

        for key in keys {
            if let Ok(Some(mut account)) = conn.get::<_, Option<crate::charging_types::SubscriberAccount>>(&key).await {
                // Get rating plan and apply monthly fee
                if let Some(plan) = self.get_rating_plan("basic") { // Simplified
                    account.balance -= (plan.monthly_fee * 100.0) as i64; // Convert to cents
                    account.last_updated = Utc::now();
                    
                    let _: () = conn.set(&key, &account).await.unwrap_or(());
                    processed_accounts += 1;
                }
            }
        }

        info!("Applied monthly fees to {} accounts", processed_accounts);
        Ok(processed_accounts)
    }
}
