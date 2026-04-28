use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

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
