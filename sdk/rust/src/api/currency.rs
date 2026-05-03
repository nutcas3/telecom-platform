use crate::client::HTTPClient;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

/// Currency and Billing API
#[derive(Clone)]
pub struct CurrencyAPI {
    client: HTTPClient,
}

impl CurrencyAPI {
    pub fn new(client: HTTPClient) -> Self {
        Self { client }
    }

    // Currency Conversion

    /// Convert currency
    pub async fn convert(
        &self,
        from: &str,
        to: &str,
        amount: f64,
    ) -> Result<ConvertResponse, crate::error::TelecomError> {
        let body = serde_json::json!({
            "from": from,
            "to": to,
            "amount": amount
        });
        self.client.post("/api/v1/currency/convert", &body).await
    }

    /// Get exchange rate between currencies
    pub async fn get_exchange_rate(&self, from: &str, to: &str) -> Result<ExchangeRate, crate::error::TelecomError> {
        self.client.get(&format!("/api/v1/currency/exchange/{}/{}", from, to), &HashMap::new()).await
    }

    /// Get exchange rate history
    pub async fn get_exchange_rate_history(
        &self,
        from: &str,
        to: &str,
        days: Option<u32>,
    ) -> Result<Vec<ExchangeRate>, crate::error::TelecomError> {
        let mut params = HashMap::new();
        if let Some(d) = days {
            params.insert("days".to_string(), d.to_string());
        }
        self.client.get(&format!("/api/v1/currency/exchange/{}/{}", from, to), &params).await
    }

    /// Get supported currencies
    pub async fn get_supported_currencies(&self) -> Result<Vec<Currency>, crate::error::TelecomError> {
        self.client.get("/api/v1/currency/currencies", &HashMap::new()).await
    }

    /// Refresh exchange rates
    pub async fn refresh_exchange_rates(&self) -> Result<HashMap<String, serde_json::Value>, crate::error::TelecomError> {
        self.client.post("/api/v1/currency/exchange/refresh", &serde_json::Value::Null).await
    }

    // Billing

    /// Process billing transaction
    pub async fn process_billing(
        &self,
        billing_data: &serde_json::Value,
    ) -> Result<BillingTransaction, crate::error::TelecomError> {
        self.client.post("/api/v1/currency/billing", billing_data).await
    }

    /// Get billing history for a profile
    pub async fn get_billing_history(
        &self,
        profile_id: &str,
        limit: Option<u32>,
    ) -> Result<Vec<BillingTransaction>, crate::error::TelecomError> {
        let mut params = HashMap::new();
        if let Some(l) = limit {
            params.insert("limit".to_string(), l.to_string());
        }
        self.client.get(&format!("/api/v1/currency/billing/history/{}", profile_id), &params).await
    }

    /// Get billing summary for a profile
    pub async fn get_billing_summary(
        &self,
        profile_id: &str,
        period: Option<&str>,
    ) -> Result<BillingSummary, crate::error::TelecomError> {
        let mut params = HashMap::new();
        if let Some(p) = period {
            params.insert("period".to_string(), p.to_string());
        }
        self.client.get(&format!("/api/v1/currency/billing/summary/{}", profile_id), &params).await
    }

    /// Process refund
    pub async fn process_refund(
        &self,
        transaction_id: &str,
        reason: &str,
    ) -> Result<BillingTransaction, crate::error::TelecomError> {
        let body = serde_json::json!({ "reason": reason });
        self.client.post(&format!("/api/v1/currency/billing/refund/{}", transaction_id), &body).await
    }

    /// Get billing analytics
    pub async fn get_billing_analytics(&self, period: Option<&str>) -> Result<HashMap<String, serde_json::Value>, crate::error::TelecomError> {
        let mut params = HashMap::new();
        if let Some(p) = period {
            params.insert("period".to_string(), p.to_string());
        }
        self.client.get("/api/v1/currency/billing/analytics", &params).await
    }
}

// Currency Types

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConvertRequest {
    pub from: String,
    pub to: String,
    pub amount: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConvertResponse {
    pub from: String,
    pub to: String,
    pub amount: f64,
    pub converted: f64,
    pub rate: f64,
    pub timestamp: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExchangeRate {
    pub from: String,
    pub to: String,
    pub rate: f64,
    pub timestamp: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Currency {
    pub code: String,
    pub name: String,
    pub symbol: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BillingTransaction {
    pub id: String,
    #[serde(rename = "profile_id")]
    pub profile_id: String,
    pub amount: f64,
    pub currency: String,
    #[serde(rename = "type")]
    pub transaction_type: String,
    pub status: String,
    pub description: String,
    #[serde(rename = "created_at")]
    pub created_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BillingSummary {
    #[serde(rename = "profile_id")]
    pub profile_id: String,
    pub period: String,
    #[serde(rename = "total_amount")]
    pub total_amount: f64,
    pub currency: String,
    #[serde(rename = "transaction_count")]
    pub transaction_count: u32,
    pub breakdown: HashMap<String, f64>,
}
