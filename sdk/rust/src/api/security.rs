use crate::client::HTTPClient;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

/// Security API for fraud detection and SIM swap protection
#[derive(Clone)]
pub struct SecurityAPI {
    client: HTTPClient,
}

impl SecurityAPI {
    pub fn new(client: HTTPClient) -> Self {
        Self { client }
    }

    // Fraud Detection

    /// Analyze a transaction for fraud
    pub async fn analyze_transaction(
        &self,
        transaction: &serde_json::Value,
    ) -> Result<Option<FraudAlert>, crate::error::TelecomError> {
        match self.client.post("/api/v1/security/fraud/analyze", transaction).await {
            Ok(alert) => Ok(Some(alert)),
            Err(crate::error::TelecomError::APIError(404, _)) => Ok(None),
            Err(e) => Err(e),
        }
    }

    /// Get fraud alerts with filtering
    pub async fn get_fraud_alerts(
        &self,
        filter: Option<FraudAlertFilter>,
    ) -> Result<Vec<FraudAlert>, crate::error::TelecomError> {
        let body = if let Some(f) = filter {
            serde_json::to_value(f).unwrap_or_default()
        } else {
            serde_json::Value::Null
        };
        self.client.post("/api/v1/security/fraud/alerts", &body).await
    }

    /// Update fraud alert status
    pub async fn update_alert_status(
        &self,
        alert_id: &str,
        status: &str,
        actions: Vec<String>,
    ) -> Result<HashMap<String, serde_json::Value>, crate::error::TelecomError> {
        let body = serde_json::json!({
            "status": status,
            "actions": actions
        });
        self.client.put(&format!("/api/v1/security/fraud/alerts/{}", alert_id), &body).await
    }

    /// Get fraud detection metrics
    pub async fn get_fraud_metrics(&self, period: Option<&str>) -> Result<FraudMetrics, crate::error::TelecomError> {
        let mut params = HashMap::new();
        if let Some(p) = period {
            params.insert("period".to_string(), p.to_string());
        }
        self.client.get("/api/v1/security/fraud/metrics", &params).await
    }

    /// Get detected fraud patterns
    pub async fn get_fraud_patterns(&self) -> Result<HashMap<String, serde_json::Value>, crate::error::TelecomError> {
        self.client.get("/api/v1/security/fraud/patterns", &HashMap::new()).await
    }

    // SIM Swap Protection

    /// Verify SIM swap request
    pub async fn verify_sim_swap(
        &self,
        profile_id: &str,
        msisdn: &str,
    ) -> Result<HashMap<String, serde_json::Value>, crate::error::TelecomError> {
        let body = serde_json::json!({
            "profile_id": profile_id,
            "msisdn": msisdn
        });
        self.client.post("/api/v1/security/simswap/verify", &body).await
    }

    /// Get SIM swap history for a profile
    pub async fn get_sim_swap_history(&self, profile_id: &str) -> Result<HashMap<String, serde_json::Value>, crate::error::TelecomError> {
        self.client.get(&format!("/api/v1/security/simswap/history/{}", profile_id), &HashMap::new()).await
    }
}

// Security Types

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FraudAlert {
    pub id: String,
    #[serde(rename = "type")]
    pub alert_type: String,
    pub severity: String,
    #[serde(rename = "profile_id")]
    pub profile_id: String,
    pub description: String,
    #[serde(rename = "risk_score")]
    pub risk_score: f64,
    pub evidence: Vec<String>,
    #[serde(rename = "ip_address")]
    pub ip_address: String,
    pub timestamp: String,
    pub status: String,
    #[serde(rename = "actions_taken")]
    pub actions_taken: Vec<String>,
    pub metadata: HashMap<String, serde_json::Value>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FraudMetrics {
    pub period: String,
    #[serde(rename = "total_alerts")]
    pub total_alerts: i64,
    #[serde(rename = "resolved_alerts")]
    pub resolved_alerts: i64,
    #[serde(rename = "false_positives")]
    pub false_positives: i64,
    #[serde(rename = "resolution_rate")]
    pub resolution_rate: f64,
    #[serde(rename = "false_positive_rate")]
    pub false_positive_rate: f64,
    #[serde(rename = "by_type")]
    pub by_type: HashMap<String, i64>,
    #[serde(rename = "by_severity")]
    pub by_severity: HashMap<String, i64>,
    #[serde(rename = "generated_at")]
    pub generated_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FraudAlertFilter {
    #[serde(rename = "type", skip_serializing_if = "Option::is_none")]
    pub alert_type: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub severity: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub status: Option<String>,
    #[serde(default = "default_limit")]
    pub limit: u32,
}

fn default_limit() -> u32 {
    50
}

impl Default for FraudAlertFilter {
    fn default() -> Self {
        Self {
            alert_type: None,
            severity: None,
            status: None,
            limit: 50,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum FraudType {
    AccountTakeover,
    SubscriptionFraud,
    PaymentFraud,
    UsageAnomaly,
    SimSwap,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum FraudSeverity {
    Low,
    Medium,
    High,
    Critical,
}
