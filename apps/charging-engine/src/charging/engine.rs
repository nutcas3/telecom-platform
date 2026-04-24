use std::time::{Duration, SystemTime};

use tracing::info;

use super::rating_plans_repo::RatingPlansRepo;
use super::types::RatingPlan;
use crate::errors::{ChargingResult, ErrorContext, ChargingError};

/// ChargingEngine coordinates Redis-backed hot-path state (balances, sessions)
/// and Postgres-backed configuration (rating plans) via `RatingPlansRepo`.
pub struct ChargingEngine {
    pub(crate) redis_client: redis::Client,
    pub(crate) plans: RatingPlansRepo,
    pub(crate) sync_interval: Duration,
    pub(crate) startup_time: SystemTime,
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

        Ok(Self {
            redis_client,
            plans,
            sync_interval: Duration::from_secs(sync_interval_secs),
            startup_time: SystemTime::now(),
        })
    }

    /// Fetch a rating plan by id from Postgres. Returns `Ok(None)` if unknown.
    pub async fn get_rating_plan(&self, plan_id: &str) -> ChargingResult<Option<RatingPlan>> {
        self.plans.get(plan_id).await
    }

    /// Upsert a rating plan in Postgres.
    pub async fn add_rating_plan(&self, plan: RatingPlan) -> ChargingResult<()> {
        self.plans.upsert(&plan).await
    }

    /// Soft-delete a rating plan in Postgres (marks inactive).
    pub async fn remove_rating_plan(&self, plan_id: &str) -> ChargingResult<bool> {
        self.plans.deactivate(plan_id).await
    }

    /// List all active rating plans from Postgres.
    pub async fn list_rating_plans(&self) -> ChargingResult<Vec<RatingPlan>> {
        let map = self.plans.list_map().await?;
        Ok(map.into_values().collect())
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

    /// Process-uptime since construction.
    pub fn uptime(&self) -> Duration {
        self.startup_time
            .elapsed()
            .unwrap_or_else(|_| Duration::from_secs(0))
    }
}
