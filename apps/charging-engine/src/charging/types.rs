use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SubscriberAccount {
    pub imsi: String,
    pub balance: i64,        // in smallest currency unit (e.g., cents)
    pub data_limit: u64,    // bytes
    pub data_used: u64,     // bytes
    pub voice_limit: u64,   // seconds
    pub voice_used: u64,    // seconds
    pub sms_limit: u64,     // count
    pub sms_used: u64,      // count
    pub status: AccountStatus,
    pub last_updated: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum AccountStatus {
    Active,
    Suspended,
    Terminated,
    Blocked,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UsageEvent {
    pub imsi: String,
    pub session_id: String,
    pub usage_type: UsageType,
    pub volume: u64,        // bytes, seconds, or count
    pub timestamp: DateTime<Utc>,
    pub rate: f64,          // cost per unit
    pub cost: f64,          // total cost
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum UsageType {
    Data,
    Voice,
    SMS,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RatingPlan {
    pub plan_id: String,
    pub name: String,
    pub data_rate: f64,     // cost per MB
    pub voice_rate: f64,    // cost per minute
    pub sms_rate: f64,      // cost per SMS
    pub monthly_fee: f64,
    pub data_limit: u64,    // bytes
    pub voice_limit: u64,   // seconds
    pub sms_limit: u64,     // count
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChargingSession {
    pub session_id: String,
    pub imsi: String,
    pub start_time: DateTime<Utc>,
    pub end_time: Option<DateTime<Utc>>,
    pub data_bytes: u64,
    pub voice_seconds: u64,
    pub sms_count: u64,
    pub total_cost: f64,
    pub status: SessionStatus,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum SessionStatus {
    Active,
    Completed,
    Terminated,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChargingRule {
    pub rule_id: String,
    pub name: String,
    pub priority: u32,
    pub conditions: Vec<Condition>,
    pub actions: Vec<Action>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Condition {
    pub field: String,
    pub operator: String,
    pub value: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Action {
    pub action_type: String,
    pub parameters: std::collections::HashMap<String, String>,
}
