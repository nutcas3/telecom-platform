//! Telecom Platform Rust SDK
//! 
//! Provides async/await support with tokio for the Telecom Platform API.

use chrono::{DateTime, Utc};
use reqwest::{Client, StatusCode};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::time::Duration;
use thiserror::Error;
use tokio_tungstenite::{connect_async, tungstenite::Message};
use url::Url;
use uuid::Uuid;

#[derive(Error, Debug)]
pub enum TelecomError {
    #[error("Authentication failed")]
    AuthenticationError,
    #[error("API error: {0}")]
    APIError(String),
    #[error("Network error: {0}")]
    NetworkError(#[from] reqwest::Error),
    #[error("Validation error: {0}")]
    ValidationError(String),
    #[error("Rate limit exceeded")]
    RateLimitError,
    #[error("Server error: {0}")]
    ServerError(StatusCode),
    #[error("WebSocket error: {0}")]
    WebSocketError(String),
    #[error("JSON error: {0}")]
    JsonError(#[from] serde_json::Error),
}

#[derive(Debug, Clone, Deserialize, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum SubscriberStatus {
    Active,
    Suspended,
    Terminated,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum UsageType {
    Data,
    Voice,
    Sms,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum PaymentStatus {
    Pending,
    Completed,
    Failed,
    Refunded,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct Subscriber {
    pub id: i64,
    pub imsi: String,
    pub msisdn: String,
    pub first_name: String,
    pub last_name: String,
    pub email: String,
    pub organization_id: Option<String>,
    pub status: SubscriberStatus,
    pub plan_id: i64,
    pub balance: f64,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct SubscriberList {
    pub subscribers: Vec<Subscriber>,
    pub total: i64,
    pub page: i32,
    pub page_size: i32,
    pub has_next: bool,
    pub has_prev: bool,
}

#[derive(Debug, Clone, Serialize)]
pub struct CreateSubscriberRequest {
    pub imsi: String,
    pub msisdn: String,
    pub first_name: String,
    pub last_name: String,
    pub email: String,
    pub plan_id: i64,
    pub organization_id: Option<String>,
}

#[derive(Debug, Clone, Serialize)]
pub struct UpdateSubscriberRequest {
    pub first_name: Option<String>,
    pub last_name: Option<String>,
    pub email: Option<String>,
    pub plan_id: Option<i64>,
    pub status: Option<SubscriberStatus>,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct UsageStats {
    pub subscriber_id: String,
    pub data_up: i64,
    pub data_down: i64,
    pub voice_seconds: i64,
    pub sms_count: i64,
    pub period_start: DateTime<Utc>,
    pub period_end: DateTime<Utc>,
    pub cost: f64,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct UsageEvent {
    pub id: String,
    pub subscriber_id: String,
    pub usage_type: UsageType,
    pub amount: i64,
    pub cost: f64,
    pub timestamp: DateTime<Utc>,
    pub metadata: Option<HashMap<String, serde_json::Value>>,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct UsageEventList {
    pub events: Vec<UsageEvent>,
    pub total: i64,
    pub page: i32,
    pub page_size: i32,
    pub has_next: bool,
    pub has_prev: bool,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct CurrentSession {
    pub session_id: String,
    pub start_time: DateTime<Utc>,
    pub data_up: i64,
    pub data_down: i64,
    pub voice_seconds: i64,
    pub sms_count: i64,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct RealTimeUsage {
    pub current_session: Option<CurrentSession>,
    pub today_usage: Option<HashMap<String, i64>>,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct PaymentTransaction {
    pub id: String,
    pub subscriber_id: String,
    pub amount: f64,
    pub currency: String,
    pub status: PaymentStatus,
    pub gateway: String,
    pub transaction_id: Option<String>,
    pub created_at: DateTime<Utc>,
    pub completed_at: Option<DateTime<Utc>>,
    pub metadata: Option<HashMap<String, serde_json::Value>>,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct PaymentTransactionList {
    pub transactions: Vec<PaymentTransaction>,
    pub total: i64,
    pub page: i32,
    pub page_size: i32,
    pub has_next: bool,
    pub has_prev: bool,
}

#[derive(Debug, Clone, Serialize)]
pub struct CreatePaymentRequest {
    pub subscriber_id: String,
    pub amount: f64,
    pub currency: String,
    pub gateway: String,
    pub metadata: Option<HashMap<String, serde_json::Value>>,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct RatingPlan {
    pub plan_id: String,
    pub name: String,
    pub data_rate: f64,
    pub voice_rate: f64,
    pub sms_rate: f64,
    pub monthly_fee: f64,
    pub data_limit: i64,
    pub voice_limit: i64,
    pub sms_limit: i64,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct SystemStats {
    pub active_sessions: i64,
    pub total_accounts: i64,
    pub blocked_users: i64,
    pub low_balance_alerts: i64,
    pub uptime: f64,
    pub cpu_usage: f64,
    pub memory_usage: f64,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct HealthStatus {
    pub status: String,
    pub timestamp: DateTime<Utc>,
    pub checks: HashMap<String, serde_json::Value>,
    pub uptime: f64,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct WebSocketMessage {
    pub r#type: String,
    pub data: HashMap<String, serde_json::Value>,
    pub timestamp: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize)]
pub struct GraphQLRequest {
    pub query: String,
    pub variables: Option<HashMap<String, serde_json::Value>>,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct GraphQLResponse {
    pub data: Option<HashMap<String, serde_json::Value>>,
    pub errors: Option<Vec<GraphQLError>>,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct GraphQLError {
    pub message: String,
    pub locations: Option<Vec<GraphQLErrorLocation>>,
    pub path: Option<Vec<String>>,
    pub extensions: Option<HashMap<String, serde_json::Value>>,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct GraphQLErrorLocation {
    pub line: i32,
    pub column: i32,
}

/// Telecom Platform SDK client
#[derive(Clone)]
pub struct TelecomClient {
    client: Client,
    config: TelecomConfig,
}

#[derive(Debug, Clone)]
pub struct TelecomConfig {
    pub api_url: String,
    pub api_key: Option<String>,
    pub timeout: Duration,
    pub max_retries: u32,
}

impl Default for TelecomConfig {
    fn default() -> Self {
        Self {
            api_url: "http://localhost:8000".to_string(),
            api_key: None,
            timeout: Duration::from_secs(30),
            max_retries: 3,
        }
    }
}

impl TelecomClient {
    /// Create a new Telecom client
    pub fn new(config: TelecomConfig) -> Self {
        let client = Client::builder()
            .timeout(config.timeout)
            .build()
            .expect("Failed to create HTTP client");

        Self { client, config }
    }

    async fn handle_response(&self, response: reqwest::Response) -> Result<reqwest::Response, TelecomError> {
        match response.status() {
            StatusCode::UNAUTHORIZED => Err(TelecomError::AuthenticationError),
            StatusCode::TOO_MANY_REQUESTS => Err(TelecomError::RateLimitError),
            status if status.is_client_error() => {
                let error_text = response.text().await.unwrap_or_else(|_| "Bad request".to_string());
                Err(TelecomError::APIError(error_text))
            }
            status if status.is_server_error() => Err(TelecomError::ServerError(status)),
            _ => Ok(response),
        }
    }

    /// Get subscriber by ID
    pub async fn get_subscriber(&self, id: i64) -> Result<Subscriber, TelecomError> {
        let response = self
            .client
            .get(&format!("{}/v1/subscribers/{}", self.config.api_url, id))
            .send()
            .await?;

        let response = self.handle_response(response).await?;
        Ok(response.json().await?)
    }

    /// List subscribers with pagination
    pub async fn list_subscribers(
        &self,
        page: i32,
        page_size: i32,
        status: Option<SubscriberStatus>,
    ) -> Result<SubscriberList, TelecomError> {
        let mut request = self
            .client
            .get(&format!("{}/v1/subscribers", self.config.api_url))
            .query(&[("page", page), ("page_size", page_size)]);

        if let Some(status) = status {
            request = request.query(&[("status", serde_json::to_string(&status).unwrap())]);
        }

        let response = request.send().await?;
        let response = self.handle_response(response).await?;
        Ok(response.json().await?)
    }

    /// Create a new subscriber
    pub async fn create_subscriber(&self, request: CreateSubscriberRequest) -> Result<Subscriber, TelecomError> {
        let response = self
            .client
            .post(&format!("{}/v1/subscribers", self.config.api_url))
            .json(&request)
            .send()
            .await?;

        let response = self.handle_response(response).await?;
        Ok(response.json().await?)
    }

    /// Update an existing subscriber
    pub async fn update_subscriber(&self, id: i64, request: UpdateSubscriberRequest) -> Result<Subscriber, TelecomError> {
        let response = self
            .client
            .put(&format!("{}/v1/subscribers/{}", self.config.api_url, id))
            .json(&request)
            .send()
            .await?;

        let response = self.handle_response(response).await?;
        Ok(response.json().await?)
    }

    /// Suspend a subscriber
    pub async fn suspend_subscriber(&self, id: i64) -> Result<Subscriber, TelecomError> {
        let response = self
            .client
            .post(&format!("{}/v1/subscribers/{}/suspend", self.config.api_url, id))
            .send()
            .await?;

        let response = self.handle_response(response).await?;
        Ok(response.json().await?)
    }

    /// Activate a suspended subscriber
    pub async fn activate_subscriber(&self, id: i64) -> Result<Subscriber, TelecomError> {
        let response = self
            .client
            .post(&format!("{}/v1/subscribers/{}/activate", self.config.api_url, id))
            .send()
            .await?;

        let response = self.handle_response(response).await?;
        Ok(response.json().await?)
    }

    /// Terminate a subscriber
    pub async fn terminate_subscriber(&self, id: i64) -> Result<bool, TelecomError> {
        let response = self
            .client
            .delete(&format!("{}/v1/subscribers/{}", self.config.api_url, id))
            .send()
            .await?;

        let response = self.handle_response(response).await?;
        let result: serde_json::Value = response.json().await?;
        Ok(result["success"].as_bool().unwrap_or(false))
    }

    /// Get usage statistics for a subscriber
    pub async fn get_usage_stats(
        &self,
        subscriber_id: i64,
        start_date: DateTime<Utc>,
        end_date: DateTime<Utc>,
    ) -> Result<UsageStats, TelecomError> {
        let response = self
            .client
            .get(&format!("{}/v1/subscribers/{}/usage", self.config.api_url, subscriber_id))
            .query(&[
                ("start_date", start_date.to_rfc3339()),
                ("end_date", end_date.to_rfc3339()),
            ])
            .send()
            .await?;

        let response = self.handle_response(response).await?;
        Ok(response.json().await?)
    }

    /// Create a payment transaction
    pub async fn create_payment_transaction(&self, request: CreatePaymentRequest) -> Result<PaymentTransaction, TelecomError> {
        let response = self
            .client
            .post(&format!("{}/v1/payments/transactions", self.config.api_url))
            .json(&request)
            .send()
            .await?;

        let response = self.handle_response(response).await?;
        Ok(response.json().await?)
    }

    /// Get system statistics
    pub async fn get_system_stats(&self) -> Result<SystemStats, TelecomError> {
        let response = self
            .client
            .get(&format!("{}/v1/system/stats", self.config.api_url))
            .send()
            .await?;

        let response = self.handle_response(response).await?;
        Ok(response.json().await?)
    }

    /// Get system health status
    pub async fn get_health_status(&self) -> Result<HealthStatus, TelecomError> {
        let response = self
            .client
            .get(&format!("{}/v1/health", self.config.api_url))
            .send()
            .await?;

        let response = self.handle_response(response).await?;
        Ok(response.json().await?)
    }

    /// Execute a GraphQL query
    pub async fn execute_graphql_query(
        &self,
        query: String,
        variables: Option<HashMap<String, serde_json::Value>>,
    ) -> Result<GraphQLResponse, TelecomError> {
        let request = GraphQLRequest { query, variables };
        let response = self
            .client
            .post(&format!("{}/graphql", self.config.api_url))
            .json(&request)
            .send()
            .await?;

        let response = self.handle_response(response).await?;
        Ok(response.json().await?)
    }

    /// Connect to WebSocket for real-time updates
    pub async fn connect_websocket<F>(&self, message_handler: F) -> Result<(), TelecomError>
    where
        F: Fn(WebSocketMessage) + Send + Sync + 'static,
    {
        let ws_url = self.config.api_url.replace("http://", "ws://") + "/ws";
        let (ws_stream, _) = connect_async(&ws_url).await.map_err(|e| TelecomError::WebSocketError(e.to_string()))?;

        println!("WebSocket connected to {}", ws_url);

        let (mut write, mut read) = ws_stream.split();
        let handler = Arc::new(message_handler);
        
        // Spawn task to handle incoming messages
        let handler_clone = handler.clone();
        let read_task = tokio::spawn(async move {
            while let Some(msg) = read.next().await {
                match msg {
                    Ok(tungstenite::Message::Text(text)) => {
                        if let Ok(ws_msg) = serde_json::from_str::<WebSocketMessage>(&text) {
                            handler_clone(ws_msg);
                        }
                    }
                    Ok(tungstenite::Message::Binary(data)) => {
                        if let Ok(text) = String::from_utf8(data) {
                            if let Ok(ws_msg) = serde_json::from_str::<WebSocketMessage>(&text) {
                                handler_clone(ws_msg);
                            }
                        }
                    }
                    Ok(tungstenite::Message::Close(_)) => {
                        println!("WebSocket connection closed");
                        break;
                    }
                    Err(e) => {
                        eprintln!("WebSocket error: {}", e);
                        break;
                    }
                    _ => {}
                }
            }
        });

        // Return a handle that can be used to manage the connection
        tokio::spawn(async move {
            let _ = read_task.await;
        });
        
        Ok(())
    }
}
