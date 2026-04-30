//! Postgres-backed rating plan repository.
//! `ChargingEngine`. Rating plans are shared config: the same rows are
//! consumed by `apps/api-server` via the `rating_plans` table, so both
//! services read a single source of truth.

use std::collections::HashMap;
use std::process::Command;

use sqlx::postgres::{PgPool, PgPoolOptions};
use sqlx::Row;

use super::types::RatingPlan;
use crate::circuit_breaker::{CircuitBreaker, CircuitBreakerError};
use crate::errors::{ChargingError, ChargingResult};

/// Repository for rating plans backed by Postgres.
#[derive(Clone)]
pub struct RatingPlansRepo {
    pool: PgPool,
    circuit_breaker: CircuitBreaker,
}

impl RatingPlansRepo {
    /// Connect to Postgres using the given DSN and run migrations.
    /// Note: Database schema is managed by centralized migration system using Goose.
    /// Migrations are automatically run on startup.
    pub async fn connect(database_url: &str) -> ChargingResult<Self> {
        let pool = PgPoolOptions::new()
            .max_connections(10)
            .connect(database_url)
            .await
            .map_err(|e| crate::errors::ChargingError::InternalError(e.to_string()))?;

        let circuit_breaker = CircuitBreaker::new(5, std::time::Duration::from_secs(60));

        // Run Goose migrations on startup
        if let Err(e) = run_migrations(database_url).await {
            return Err(crate::errors::ChargingError::InternalError(format!("Migration failed: {}", e)));
        }

        Ok(Self { pool, circuit_breaker })
    }

    /// Seed the canonical basic/premium/enterprise plans if the table is empty.
    /// Idempotent — existing rows are preserved.
    pub async fn seed_defaults(&self) -> ChargingResult<()> {
        let defaults = vec![
            RatingPlan {
                plan_id: "basic".into(),
                name: "Basic Plan".into(),
                data_rate: 0.01,
                voice_rate: 0.05,
                sms_rate: 0.10,
                monthly_fee: 10.0,
                data_limit: 1_000_000_000,
                voice_limit: 300,
                sms_limit: 100,
            },
            RatingPlan {
                plan_id: "premium".into(),
                name: "Premium Plan".into(),
                data_rate: 0.005,
                voice_rate: 0.02,
                sms_rate: 0.05,
                monthly_fee: 25.0,
                data_limit: 5_000_000_000,
                voice_limit: 1000,
                sms_limit: 500,
            },
            RatingPlan {
                plan_id: "enterprise".into(),
                name: "Enterprise Plan".into(),
                data_rate: 0.001,
                voice_rate: 0.01,
                sms_rate: 0.02,
                monthly_fee: 100.0,
                data_limit: 50_000_000_000,
                voice_limit: 5000,
                sms_limit: 2000,
            },
        ];

        for plan in defaults {
            self.upsert(&plan).await?;
        }
        Ok(())
    }

    /// Fetch a single plan by id.
    pub async fn get(&self, plan_id: &str) -> ChargingResult<Option<RatingPlan>> {
        let plan_id = plan_id.to_string();
        self.circuit_breaker.execute(|| async move {
            let row = sqlx::query(
                r#"
                SELECT plan_id, name, data_rate, voice_rate, sms_rate,
                       monthly_fee, data_limit, voice_limit, sms_limit
                  FROM rating_plans
                 WHERE plan_id = $1 AND is_active = TRUE
                "#,
            )
            .bind(&plan_id)
            .fetch_optional(&self.pool)
            .await
            .map_err(|e| crate::errors::ChargingError::DatabaseError(e.to_string()))?;

            Ok(row.map(row_to_plan))
        }).await.map_err(|e| match e {
            CircuitBreakerError::Open => ChargingError::DatabaseError("Circuit breaker is open".to_string()),
            CircuitBreakerError::Inner(e) => e,
        })
    }

    /// Fetch all active plans as a HashMap for fast lookup.
    pub async fn list_map(&self) -> ChargingResult<HashMap<String, RatingPlan>> {
        self.circuit_breaker.execute(|| async {
            let rows = sqlx::query(
                r#"
                SELECT plan_id, name, data_rate, voice_rate, sms_rate,
                       monthly_fee, data_limit, voice_limit, sms_limit
                  FROM rating_plans
                 WHERE is_active = TRUE
                "#,
            )
            .fetch_all(&self.pool)
            .await
            .map_err(|e| crate::errors::ChargingError::DatabaseError(e.to_string()))?;

            let plans = rows
                .into_iter()
                .map(|row| row_to_plan(row))
                .collect::<Vec<_>>();

            let mut out = HashMap::new();
            for plan in plans {
                out.insert(plan.plan_id.clone(), plan);
            }
            Ok(out)
        }).await.map_err(|e| match e {
            CircuitBreakerError::Open => ChargingError::DatabaseError("Circuit breaker is open".to_string()),
            CircuitBreakerError::Inner(e) => e,
        })
    }

    /// List all active plans.
    pub async fn list(&self) -> Result<Vec<RatingPlan>, sqlx::Error> {
        let rows = sqlx::query(
            r#"
            SELECT plan_id, name, data_rate, voice_rate, sms_rate,
                   monthly_fee, data_limit, voice_limit, sms_limit
              FROM rating_plans
             WHERE is_active = TRUE
            "#,
        )
        .fetch_all(&self.pool)
        .await?;

        let plans = rows
            .into_iter()
            .map(|row| row_to_plan(row))
            .collect();

        Ok(plans)
    }

    /// Insert or update a plan.
    pub async fn upsert(&self, plan: &RatingPlan) -> ChargingResult<()> {
        let plan = plan.clone();
        self.circuit_breaker.execute(|| async move {
            sqlx::query(
                r#"
                INSERT INTO rating_plans
                    (plan_id, name, data_rate, voice_rate, sms_rate,
                     monthly_fee, data_limit, voice_limit, sms_limit, is_active, updated_at)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, TRUE, NOW())
                ON CONFLICT (plan_id) DO UPDATE SET
                    name        = EXCLUDED.name,
                    data_rate   = EXCLUDED.data_rate,
                    voice_rate  = EXCLUDED.voice_rate,
                    sms_rate    = EXCLUDED.sms_rate,
                    monthly_fee = EXCLUDED.monthly_fee,
                    data_limit  = EXCLUDED.data_limit,
                    voice_limit = EXCLUDED.voice_limit,
                    sms_limit   = EXCLUDED.sms_limit,
                    is_active   = TRUE,
                    updated_at  = NOW()
                "#,
            )
            .bind(&plan.plan_id)
            .bind(&plan.name)
            .bind(plan.data_rate)
            .bind(plan.voice_rate)
            .bind(plan.sms_rate)
            .bind(plan.monthly_fee)
            .bind(plan.data_limit as i64)
            .bind(plan.voice_limit as i64)
            .bind(plan.sms_limit as i64)
            .execute(&self.pool)
            .await
            .map_err(|e| crate::errors::ChargingError::DatabaseError(e.to_string()))?;
            Ok(())
        }).await.map_err(|e| match e {
            CircuitBreakerError::Open => ChargingError::DatabaseError("Circuit breaker is open".to_string()),
            CircuitBreakerError::Inner(e) => e,
        })
    }

    /// Soft-delete a plan by marking it inactive.
    pub async fn deactivate(&self, plan_id: &str) -> ChargingResult<bool> {
        let plan_id = plan_id.to_string();
        self.circuit_breaker.execute(|| async move {
            let result = sqlx::query(
                r#"UPDATE rating_plans SET is_active = FALSE, updated_at = NOW() WHERE plan_id = $1"#,
            )
            .bind(&plan_id)
            .execute(&self.pool)
            .await
            .map_err(|e| crate::errors::ChargingError::DatabaseError(e.to_string()))?;
            Ok(result.rows_affected() > 0)
        }).await.map_err(|e| match e {
            CircuitBreakerError::Open => ChargingError::DatabaseError("Circuit breaker is open".to_string()),
            CircuitBreakerError::Inner(e) => e,
        })
    }
}

/// Map a Postgres row to a RatingPlan. Column order must match the SELECT.
fn row_to_plan(row: sqlx::postgres::PgRow) -> RatingPlan {
    RatingPlan {
        plan_id: row.get("plan_id"),
        name: row.get("name"),
        data_rate: row.get("data_rate"),
        voice_rate: row.get("voice_rate"),
        sms_rate: row.get("sms_rate"),
        monthly_fee: row.get("monthly_fee"),
        data_limit: row.get::<i64, _>("data_limit") as u64,
        voice_limit: row.get::<i64, _>("voice_limit") as u64,
        sms_limit: row.get::<i64, _>("sms_limit") as u64,
    }
}

/// Run Goose migrations by calling the goose binary.
/// This ensures both Go and Rust services use the same migration system.
/// Changes to the repo root directory to find the migrations folder.
async fn run_migrations(database_url: &str) -> Result<(), Box<dyn std::error::Error>> {
    use std::process::Command;
    
    // Change to repo root to find migrations directory
    let manifest_dir = std::path::PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    let repo_root = manifest_dir
        .parent()
        .and_then(|p| p.parent())
        .ok_or("Failed to determine repo root")?;
    
    let output = Command::new("goose")
        .current_dir(&repo_root)
        .args(["postgres", database_url, "up", "migrations"])
        .output()?;

    if !output.status.success() {
        let stderr = String::from_utf8_lossy(&output.stderr);
        return Err(format!("Goose migration failed: {}", stderr).into());
    }

    tracing::info!("Database migrations completed successfully");
    Ok(())
}
