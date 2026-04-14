use std::collections::HashMap;
use std::time::Duration;

use super::types::{RatingPlan, SubscriberAccount};
use crate::errors::{ChargingError, ChargingResult, ErrorContext};

pub struct ChargingEngine {
    redis_client: redis::Client,
    rating_plans: HashMap<String, RatingPlan>,
    sync_interval: Duration,
}

impl ChargingEngine {
    pub fn new(redis_url: &str, sync_interval_secs: u64) -> ChargingResult<Self> {
        let redis_client = redis::Client::open(redis_url)
            .context("Failed to create Redis client")?;

        let mut rating_plans = HashMap::new();
        
        // Default rating plans
        rating_plans.insert("basic".to_string(), RatingPlan {
            plan_id: "basic".to_string(),
            name: "Basic Plan".to_string(),
            data_rate: 0.01,      // $0.01 per MB
            voice_rate: 0.05,     // $0.05 per minute
            sms_rate: 0.10,       // $0.10 per SMS
            monthly_fee: 10.0,    // $10 monthly
            data_limit: 1_000_000_000,    // 1GB
            voice_limit: 300,             // 300 minutes
            sms_limit: 100,                // 100 SMS
        });

        rating_plans.insert("premium".to_string(), RatingPlan {
            plan_id: "premium".to_string(),
            name: "Premium Plan".to_string(),
            data_rate: 0.005,     // $0.005 per MB
            voice_rate: 0.02,     // $0.02 per minute
            sms_rate: 0.05,       // $0.05 per SMS
            monthly_fee: 25.0,    // $25 monthly
            data_limit: 5_000_000_000,    // 5GB
            voice_limit: 1000,            // 1000 minutes
            sms_limit: 500,                // 500 SMS
        });

        rating_plans.insert("enterprise".to_string(), RatingPlan {
            plan_id: "enterprise".to_string(),
            name: "Enterprise Plan".to_string(),
            data_rate: 0.001,    // $0.001 per MB
            voice_rate: 0.01,    // $0.01 per minute
            sms_rate: 0.02,      // $0.02 per SMS
            monthly_fee: 100.0,   // $100 monthly
            data_limit: 50_000_000_000,  // 50GB
            voice_limit: 5000,            // 5000 minutes
            sms_limit: 2000,              // 2000 SMS
        });

        Ok(Self {
            redis_client,
            rating_plans,
            sync_interval: Duration::from_secs(sync_interval_secs),
        })
    }

    pub fn get_rating_plan(&self, plan_id: &str) -> Option<&RatingPlan> {
        self.rating_plans.get(plan_id)
    }

    pub fn add_rating_plan(&mut self, plan: RatingPlan) {
        self.rating_plans.insert(plan.plan_id.clone(), plan);
    }

    pub fn remove_rating_plan(&mut self, plan_id: &str) -> Option<RatingPlan> {
        self.rating_plans.remove(plan_id)
    }

    pub fn list_rating_plans(&self) -> Vec<&RatingPlan> {
        self.rating_plans.values().collect()
    }

    pub async fn start(&self) -> ChargingResult<()> {
        info!("Starting charging engine with sync interval: {:?}", self.sync_interval);
        
        // Initialize Redis connection
        let mut conn = self.redis_client.get_async_connection().await
            .context("Failed to establish Redis connection")?;

        // Test Redis connection
        let _: String = conn.get("test").await.unwrap_or_else(|_| "ok".to_string());
        
        info!("Charging engine started successfully");
        Ok(())
    }

    pub async fn stop(&self) -> ChargingResult<()> {
        info!("Stopping charging engine");
        // Cleanup operations
        Ok(())
    }
}
