use crate::client::HTTPClient;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

/// Analytics API for churn prediction, market analysis, and pricing optimization
#[derive(Clone)]
pub struct AnalyticsAPI {
    client: HTTPClient,
}

impl AnalyticsAPI {
    pub fn new(client: HTTPClient) -> Self {
        Self { client }
    }

    // Churn Analysis

    /// Predict churn risk for a profile
    pub async fn predict_churn(&self, profile_id: &str) -> Result<ChurnPrediction, crate::error::TelecomError> {
        let body = serde_json::json!({ "profile_id": profile_id });
        self.client.post("/api/v1/analytics/churn/predict", &body).await
    }

    /// Get churn metrics
    pub async fn get_churn_metrics(&self, period: Option<&str>) -> Result<ChurnMetrics, crate::error::TelecomError> {
        let mut params = HashMap::new();
        if let Some(p) = period {
            params.insert("period".to_string(), p.to_string());
        }
        self.client.get("/api/v1/analytics/churn/metrics", &params).await
    }

    /// Get at-risk customers
    pub async fn get_at_risk_customers(
        &self,
        risk_level: ChurnRiskLevel,
        limit: Option<u32>,
    ) -> Result<Vec<ChurnPrediction>, crate::error::TelecomError> {
        let mut body = serde_json::json!({ "risk_level": risk_level });
        if let Some(l) = limit {
            body["limit"] = serde_json::Value::Number(l.into());
        }
        self.client.post("/api/v1/analytics/churn/at-risk", &body).await
    }

    // Market Analytics

    /// Get market metrics
    pub async fn get_market_metrics(&self, period: Option<&str>) -> Result<MarketMetrics, crate::error::TelecomError> {
        let mut params = HashMap::new();
        if let Some(p) = period {
            params.insert("period".to_string(), p.to_string());
        }
        self.client.get("/api/v1/analytics/market/metrics", &params).await
    }

    /// Get competitor analysis
    pub async fn get_competitors(&self) -> Result<HashMap<String, serde_json::Value>, crate::error::TelecomError> {
        self.client.get("/api/v1/analytics/market/competitors", &HashMap::new()).await
    }

    /// Get market opportunities
    pub async fn get_market_opportunities(&self) -> Result<HashMap<String, serde_json::Value>, crate::error::TelecomError> {
        self.client.get("/api/v1/analytics/market/opportunities", &HashMap::new()).await
    }

    // Predictive Maintenance

    /// Get maintenance metrics
    pub async fn get_maintenance_metrics(&self, period: Option<&str>) -> Result<MaintenanceMetrics, crate::error::TelecomError> {
        let mut params = HashMap::new();
        if let Some(p) = period {
            params.insert("period".to_string(), p.to_string());
        }
        self.client.get("/api/v1/analytics/maintenance/metrics", &params).await
    }

    /// Get assets health
    pub async fn get_assets_health(&self) -> Result<HashMap<String, serde_json::Value>, crate::error::TelecomError> {
        self.client.get("/api/v1/analytics/maintenance/assets", &HashMap::new()).await
    }

    /// Get maintenance alerts
    pub async fn get_maintenance_alerts(&self) -> Result<HashMap<String, serde_json::Value>, crate::error::TelecomError> {
        self.client.get("/api/v1/analytics/maintenance/alerts", &HashMap::new()).await
    }

    /// Predict failure for an asset
    pub async fn predict_failure(&self, asset_id: &str) -> Result<HashMap<String, serde_json::Value>, crate::error::TelecomError> {
        self.client.post(&format!("/api/v1/analytics/maintenance/predict/{}", asset_id), &serde_json::Value::Null).await
    }

    // Pricing Optimization

    /// Get pricing metrics
    pub async fn get_pricing_metrics(&self, period: Option<&str>) -> Result<PricingMetrics, crate::error::TelecomError> {
        let mut params = HashMap::new();
        if let Some(p) = period {
            params.insert("period".to_string(), p.to_string());
        }
        self.client.get("/api/v1/analytics/pricing/metrics", &params).await
    }

    /// Optimize pricing for rate plans
    pub async fn optimize_pricing(
        &self,
        rate_plan_ids: Vec<String>,
        strategy: Option<&str>,
    ) -> Result<Vec<PricingOptimizationResult>, crate::error::TelecomError> {
        let mut body = serde_json::json!({ "rate_plan_ids": rate_plan_ids });
        if let Some(s) = strategy {
            body["strategy"] = serde_json::Value::String(s.to_string());
        }
        self.client.post("/api/v1/analytics/pricing/optimize", &body).await
    }

    /// Get price elasticity data
    pub async fn get_price_elasticity(&self) -> Result<HashMap<String, serde_json::Value>, crate::error::TelecomError> {
        self.client.get("/api/v1/analytics/pricing/elasticity", &HashMap::new()).await
    }
}

// Analytics Types

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChurnPrediction {
    pub profile_id: String,
    pub risk_level: String,
    pub risk_score: f64,
    pub predicted_churn_date: Option<String>,
    pub reasons: Vec<String>,
    pub recommendations: Vec<String>,
    pub last_updated: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChurnMetrics {
    pub period: String,
    pub total_subscribers: i64,
    pub churned_subscribers: i64,
    pub churn_rate: f64,
    pub monthly_churn_rate: f64,
    pub annual_churn_rate: f64,
    pub average_tenure_days: f64,
    pub risk_distribution: HashMap<String, i64>,
    pub generated_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MarketMetrics {
    pub period: String,
    pub total_market_size: i64,
    pub our_subscribers: i64,
    pub market_share: f64,
    pub growth_rate: f64,
    pub by_country: HashMap<String, serde_json::Value>,
    pub generated_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MaintenanceMetrics {
    pub period: String,
    pub total_assets: i64,
    pub healthy_assets: i64,
    pub assets_needing_attention: i64,
    pub uptime: f64,
    pub mean_time_to_failure: f64,
    pub mean_time_to_repair: f64,
    pub generated_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PricingMetrics {
    pub period: String,
    pub total_revenue: f64,
    pub arpu: f64,
    pub price_elasticity: f64,
    pub competitive_index: f64,
    pub optimization_roi: f64,
    pub generated_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PricingOptimizationResult {
    pub rate_plan_id: String,
    pub strategy: String,
    pub current_price: f64,
    pub optimal_price: f64,
    pub price_change_pct: f64,
    pub expected_revenue: f64,
    pub expected_demand: f64,
    pub confidence: f64,
    pub reasoning: Vec<String>,
    pub risks: Vec<String>,
    pub recommendations: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ChurnRiskLevel {
    Low,
    Medium,
    High,
    Critical,
}

impl AsRef<str> for ChurnRiskLevel {
    fn as_ref(&self) -> &str {
        match self {
            ChurnRiskLevel::Low => "low",
            ChurnRiskLevel::Medium => "medium",
            ChurnRiskLevel::High => "high",
            ChurnRiskLevel::Critical => "critical",
        }
    }
}
