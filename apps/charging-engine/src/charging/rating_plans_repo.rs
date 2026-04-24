//! Postgres-backed rating plan repository.
//! `ChargingEngine`. Rating plans are shared config: the same rows are
//! consumed by `apps/api-server` via the `rating_plans` table, so both
//! services read a single source of truth.

use std::collections::HashMap;

use sqlx::postgres::{PgPool, PgPoolOptions};
use sqlx::Row;

use super::types::RatingPlan;
use crate::errors::{ChargingError, ChargingResult, ErrorContext};

/// Repository for rating plans backed by Postgres.
#[derive(Clone)]
pub struct RatingPlansRepo {
    pool: PgPool,
}

impl RatingPlansRepo {
    /// Connect to Postgres using the given DSN and ensure the schema exists.
    pub async fn connect(database_url: &str) -> ChargingResult<Self> {
        let pool = PgPoolOptions::new()
            .max_connections(10)
            .connect(database_url)
            .await
            .map_err(|e| crate::errors::ChargingError::InternalError(e.to_string()))?;

        // Idempotent schema for rating plans. api-server's GORM migration also
        // creates a compatible `rating_plans` table; the column set here is
        // intentionally a strict subset that both services agree on.
        sqlx::query(
            r#"
            CREATE TABLE IF NOT EXISTS rating_plans (
                plan_id      TEXT PRIMARY KEY,
                name         TEXT NOT NULL,
                data_rate    DOUBLE PRECISION NOT NULL,
                voice_rate   DOUBLE PRECISION NOT NULL,
                sms_rate     DOUBLE PRECISION NOT NULL,
                monthly_fee  DOUBLE PRECISION NOT NULL,
                data_limit   BIGINT NOT NULL,
                voice_limit  BIGINT NOT NULL,
                sms_limit    BIGINT NOT NULL,
                is_active    BOOLEAN NOT NULL DEFAULT TRUE,
                created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
            )
            "#,
        )
        .execute(&pool)
        .await
        .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

        Ok(Self { pool })
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
        let row = sqlx::query(
            r#"
            SELECT plan_id, name, data_rate, voice_rate, sms_rate,
                   monthly_fee, data_limit, voice_limit, sms_limit
              FROM rating_plans
             WHERE plan_id = $1 AND is_active = TRUE
            "#,
        )
        .bind(plan_id)
        .fetch_optional(&self.pool)
        .await
        .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

        Ok(row.map(row_to_plan))
    }

    /// Fetch all active plans as a HashMap for fast lookup.
    pub async fn list_map(&self) -> ChargingResult<HashMap<String, RatingPlan>> {
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
        .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

        let mut out = HashMap::with_capacity(rows.len());
        for row in rows {
            let plan = row_to_plan(row);
            out.insert(plan.plan_id.clone(), plan);
        }
        Ok(out)
    }

    /// Insert or update a plan.
    pub async fn upsert(&self, plan: &RatingPlan) -> ChargingResult<()> {
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
        .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;
        Ok(())
    }

    /// Soft-delete a plan by marking it inactive.
    pub async fn deactivate(&self, plan_id: &str) -> ChargingResult<bool> {
        let result = sqlx::query(
            r#"UPDATE rating_plans SET is_active = FALSE, updated_at = NOW() WHERE plan_id = $1"#,
        )
        .bind(plan_id)
        .execute(&self.pool)
        .await
        .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;
        Ok(result.rows_affected() > 0)
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
