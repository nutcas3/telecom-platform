use std::time::{Duration, SystemTime};

use tracing::info;

use super::rating_plans_repo::RatingPlansRepo;
use super::types::RatingPlan;
use crate::circuit_breaker::CircuitBreaker;
use crate::errors::ChargingResult;

/// ChargingEngine coordinates Redis-backed hot-path state (balances, sessions)
/// and Postgres-backed configuration (rating plans) via `RatingPlansRepo`.
pub struct ChargingEngine {
    pub(crate) redis_client: redis::Client,
    pub(crate) plans: RatingPlansRepo,
    pub(crate) sync_interval: Duration,
    pub(crate) startup_time: SystemTime,
    pub(crate) redis_circuit_breaker: CircuitBreaker,
    pub(crate) postgres_circuit_breaker: CircuitBreaker,
}

impl ChargingEngine {
    /// Construct the engine with a Redis hot store and a Postgres-backed plan repo.
    /// The plan repo migration and default seeding must be done by the caller via
    /// [`RatingPlansRepo::connect`] + [`RatingPlansRepo::seed_defaults`].
    pub fn new(
        redis_url: &str,
        plans: RatingPlansRepo,
        sync_interval_secs: u64,
    ) -> ChargingResult<Self> {
        let redis_client = redis::Client::open(redis_url)
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;
        let sync_interval = Duration::from_secs(sync_interval_secs);

        // Initialize circuit breakers with environment variable configuration
        let redis_failure_threshold: u32 = std::env::var("REDIS_FAILURE_THRESHOLD")
            .ok()
            .and_then(|s| s.parse().ok())
            .unwrap_or(5);
        let redis_timeout: u64 = std::env::var("REDIS_TIMEOUT")
            .ok()
            .and_then(|s| s.parse().ok())
            .unwrap_or(60);
        
        let postgres_failure_threshold: u32 = std::env::var("POSTGRES_FAILURE_THRESHOLD")
            .ok()
            .and_then(|s| s.parse().ok())
            .unwrap_or(5);
        let postgres_timeout: u64 = std::env::var("POSTGRES_TIMEOUT")
            .ok()
            .and_then(|s| s.parse().ok())
            .unwrap_or(60);

        let redis_circuit_breaker = CircuitBreaker::new(
            redis_failure_threshold,
            Duration::from_secs(redis_timeout),
        );
        
        let postgres_circuit_breaker = CircuitBreaker::new(
            postgres_failure_threshold,
            Duration::from_secs(postgres_timeout),
        );

        Ok(Self {
            redis_client,
            plans,
            sync_interval,
            startup_time: SystemTime::now(),
            redis_circuit_breaker,
            postgres_circuit_breaker,
        })
    }

    /// Confirm Redis connectivity at startup.
    pub async fn test_connection(&self) -> ChargingResult<()> {
        let mut conn = self
            .redis_client
            .get_multiplexed_async_connection()
            .await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;
        let _: String = redis::AsyncCommands::get(&mut conn, "startup_probe")
            .await
            .unwrap_or_else(|_| "ok".to_string());
        Ok(())
    }

    pub async fn start(&self) -> ChargingResult<()> {
        info!(
            "Starting charging engine with sync interval: {:?}",
            self.sync_interval
        );
        self.test_connection().await?;
        info!("Charging engine started successfully");
        Ok(())
    }

    pub async fn stop(&self) -> ChargingResult<()> {
        info!("Stopping charging engine");
        Ok(())
    }

    pub fn uptime(&self) -> Duration {
        self.startup_time
            .elapsed()
            .unwrap_or_else(|_| Duration::from_secs(0))
    }

    pub async fn get_rating_plan(&self, plan_id: &str) -> ChargingResult<Option<RatingPlan>> {
        self.plans.get(plan_id).await
    }

    pub async fn add_rating_plan(&self, plan: RatingPlan) -> ChargingResult<()> {
        self.plans.upsert(&plan).await
    }

    pub async fn remove_rating_plan(&self, plan_id: &str) -> ChargingResult<bool> {
        self.plans.deactivate(plan_id).await
    }

    pub async fn list_rating_plans(&self) -> ChargingResult<Vec<RatingPlan>> {
        self.plans.list().await
            .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))
    }
}
