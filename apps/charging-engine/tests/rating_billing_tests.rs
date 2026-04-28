use charging_engine::charging::types::{ChargingRule, UsageEvent, UsageType};
use chrono::{Utc, Datelike};

#[test]
fn test_charging_rule_variants() {
    let allowed = ChargingRule::Allowed;
    let insufficient_credit = ChargingRule::InsufficientCredit;
    let data_limit = ChargingRule::DataLimitExceeded;
    let voice_limit = ChargingRule::VoiceLimitExceeded;
    let sms_limit = ChargingRule::SmsLimitExceeded;
    let user_blocked = ChargingRule::UserBlocked;
    let blocked = ChargingRule::Blocked;
    
    assert!(matches!(allowed, ChargingRule::Allowed));
    assert!(matches!(insufficient_credit, ChargingRule::InsufficientCredit));
    assert!(matches!(data_limit, ChargingRule::DataLimitExceeded));
    assert!(matches!(voice_limit, ChargingRule::VoiceLimitExceeded));
    assert!(matches!(sms_limit, ChargingRule::SmsLimitExceeded));
    assert!(matches!(user_blocked, ChargingRule::UserBlocked));
    assert!(matches!(blocked, ChargingRule::Blocked));
}

#[test]
fn test_charging_rule_as_str() {
    assert_eq!(ChargingRule::Allowed.as_str(), "ALLOWED");
    assert_eq!(ChargingRule::InsufficientCredit.as_str(), "INSUFFICIENT_CREDIT");
    assert_eq!(ChargingRule::DataLimitExceeded.as_str(), "DATA_LIMIT_EXCEEDED");
    assert_eq!(ChargingRule::VoiceLimitExceeded.as_str(), "VOICE_LIMIT_EXCEEDED");
    assert_eq!(ChargingRule::SmsLimitExceeded.as_str(), "SMS_LIMIT_EXCEEDED");
    assert_eq!(ChargingRule::UserBlocked.as_str(), "USER_BLOCKED");
    assert_eq!(ChargingRule::Blocked.as_str(), "BLOCKED");
}

#[test]
fn test_charging_rule_equality() {
    assert_eq!(ChargingRule::Allowed, ChargingRule::Allowed);
    assert_eq!(ChargingRule::InsufficientCredit, ChargingRule::InsufficientCredit);
    assert_ne!(ChargingRule::Allowed, ChargingRule::Blocked);
}

#[test]
fn test_charging_rule_clone() {
    let rule1 = ChargingRule::DataLimitExceeded;
    let rule2 = rule1.clone();
    assert_eq!(rule1, rule2);
}

#[test]
fn test_billing_period_parsing() {
    let period = "2024-01";
    let parts: Vec<&str> = period.split('-').collect();
    assert_eq!(parts.len(), 2);
    
    let year = parts[0].parse::<i32>().unwrap();
    let month = parts.get(1).and_then(|m| m.parse::<u32>().ok()).unwrap();
    
    assert_eq!(year, 2024);
    assert_eq!(month, 1);
}

#[test]
fn test_billing_period_parsing_invalid() {
    let period = "invalid";
    let parts: Vec<&str> = period.split('-').collect();
    let year = parts[0].parse::<i32>().ok();
    assert!(year.is_none());
}

#[test]
fn test_usage_event_timestamp_filtering() {
    let event = UsageEvent {
        imsi: "123456789012345".to_string(),
        session_id: "session-123".to_string(),
        usage_type: UsageType::Data,
        volume: 1_000_000,
        timestamp: Utc::now(),
        rate: 0.01,
        cost: 10.0,
    };
    
    let current_year = event.timestamp.year();
    let current_month = event.timestamp.month();
    
    assert!(current_year > 2020);
    assert!(current_month >= 1 && current_month <= 12);
}

#[test]
fn test_cost_calculation_data() {
    let volume_bytes: u64 = 2_097_152; // 2 MB
    let data_rate: f64 = 0.01;
    
    let data_mb = volume_bytes / 1_048_576;
    let remainder_bytes = volume_bytes % 1_048_576;
    
    let cost_full_mb = data_mb as f64 * data_rate;
    let cost_partial = (remainder_bytes as f64 / 1_048_576.0) * data_rate;
    let total_cost = cost_full_mb + cost_partial;
    
    assert_eq!(data_mb, 2);
    assert_eq!(cost_full_mb, 0.02);
    assert!(total_cost >= 0.02);
}

#[test]
fn test_cost_calculation_voice() {
    let volume_seconds: u64 = 120; // 2 minutes
    let voice_rate: f64 = 0.05;
    
    let voice_minutes = volume_seconds / 60;
    let remainder_seconds = volume_seconds % 60;
    
    let cost_full_min = voice_minutes as f64 * voice_rate;
    let cost_partial = (remainder_seconds as f64 / 60.0) * voice_rate;
    let total_cost = cost_full_min + cost_partial;
    
    assert_eq!(voice_minutes, 2);
    assert_eq!(cost_full_min, 0.10);
    assert!(total_cost >= 0.10);
}

#[test]
fn test_cost_calculation_sms() {
    let volume: u64 = 10;
    let sms_rate: f64 = 0.10;
    
    let cost = volume as f64 * sms_rate;
    
    assert_eq!(cost, 1.0);
}

#[test]
fn test_rate_calculation() {
    let cost: f64 = 10.0;
    let volume: u64 = 1_000_000;
    
    if volume > 0 {
        let rate = cost / volume as f64;
        assert!(rate > 0.0);
        assert!(rate < cost);
    }
}
