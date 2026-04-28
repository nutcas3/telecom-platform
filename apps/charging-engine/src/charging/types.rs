use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use redis::{FromRedisValue, ToRedisArgs, ToSingleRedisArg};

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub enum ChargingRule {
    Allowed,
    InsufficientCredit,
    DataLimitExceeded,
    VoiceLimitExceeded,
    SmsLimitExceeded,
    UserBlocked,
    Blocked,
}

impl ChargingRule {
    pub fn as_str(&self) -> &str {
        match self {
            ChargingRule::Allowed => "ALLOWED",
            ChargingRule::InsufficientCredit => "INSUFFICIENT_CREDIT",
            ChargingRule::DataLimitExceeded => "DATA_LIMIT_EXCEEDED",
            ChargingRule::VoiceLimitExceeded => "VOICE_LIMIT_EXCEEDED",
            ChargingRule::SmsLimitExceeded => "SMS_LIMIT_EXCEEDED",
            ChargingRule::UserBlocked => "USER_BLOCKED",
            ChargingRule::Blocked => "BLOCKED",
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SubscriberAccount {
    pub imsi: String,
    pub balance: i64,
    pub data_limit: u64,
    pub data_used: u64,
    pub voice_limit: u64,
    pub voice_used: u64,
    pub sms_limit: u64,
    pub sms_used: u64,
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
    pub volume: u64,
    pub timestamp: DateTime<Utc>,
    pub rate: f64,
    pub cost: f64,
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
    pub data_rate: f64,
    pub voice_rate: f64,
    pub sms_rate: f64,
    pub monthly_fee: f64,
    pub data_limit: u64,
    pub voice_limit: u64,
    pub sms_limit: u64,
}

// Redis trait implementations for serialization
impl FromRedisValue for SubscriberAccount {
    fn from_redis_value(v: redis::Value) -> Result<Self, redis::ParsingError> {
        let json: String = redis::from_redis_value(v)?;
        let account: SubscriberAccount = serde_json::from_str(&json)
            .map_err(|e| redis::ParsingError::from(e.to_string()))?;
        Ok(account)
    }
}

impl ToRedisArgs for SubscriberAccount {
    fn write_redis_args<W>(&self, out: &mut W)
    where
        W: redis::RedisWrite + ?Sized,
    {
        let json = serde_json::to_string(self)
            .expect("Failed to serialize SubscriberAccount");
        json.write_redis_args(out)
    }
}

impl ToSingleRedisArg for SubscriberAccount {}

impl FromRedisValue for UsageEvent {
    fn from_redis_value(v: redis::Value) -> Result<Self, redis::ParsingError> {
        let json: String = redis::from_redis_value(v)?;
        let event: UsageEvent = serde_json::from_str(&json)
            .map_err(|e| redis::ParsingError::from(e.to_string()))?;
        Ok(event)
    }
}

impl ToRedisArgs for UsageEvent {
    fn write_redis_args<W>(&self, out: &mut W)
    where
        W: redis::RedisWrite + ?Sized,
    {
        let json = serde_json::to_string(self)
            .expect("Failed to serialize UsageEvent");
        json.write_redis_args(out)
    }
}

impl ToSingleRedisArg for UsageEvent {}
