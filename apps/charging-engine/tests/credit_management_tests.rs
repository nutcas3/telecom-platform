use charging_engine::charging::types::{SubscriberAccount, AccountStatus, UsageEvent, UsageType};
use chrono::Utc;

#[test]
fn test_subscriber_account_creation() {
    let account = SubscriberAccount {
        imsi: "123456789012345".to_string(),
        balance: 1000,
        data_limit: 1_000_000_000,
        data_used: 0,
        voice_limit: 1000,
        voice_used: 0,
        sms_limit: 100,
        sms_used: 0,
        status: AccountStatus::Active,
        last_updated: Utc::now(),
    };
    
    assert_eq!(account.imsi, "123456789012345");
    assert_eq!(account.balance, 1000);
    assert_eq!(account.data_limit, 1_000_000_000);
    assert!(matches!(account.status, AccountStatus::Active));
}

#[test]
fn test_usage_event_creation() {
    let event = UsageEvent {
        imsi: "123456789012345".to_string(),
        session_id: "session-123".to_string(),
        usage_type: UsageType::Data,
        volume: 1_000_000,
        timestamp: Utc::now(),
        rate: 0.01,
        cost: 10.0,
    };
    
    assert_eq!(event.imsi, "123456789012345");
    assert_eq!(event.volume, 1_000_000);
    assert!(matches!(event.usage_type, UsageType::Data));
    assert_eq!(event.cost, 10.0);
}

#[test]
fn test_usage_type_data() {
    let data_type = UsageType::Data;
    let voice_type = UsageType::Voice;
    let sms_type = UsageType::SMS;
    
    assert!(matches!(data_type, UsageType::Data));
    assert!(matches!(voice_type, UsageType::Voice));
    assert!(matches!(sms_type, UsageType::SMS));
}

#[test]
fn test_account_status_variants() {
    let active = AccountStatus::Active;
    let suspended = AccountStatus::Suspended;
    let terminated = AccountStatus::Terminated;
    let blocked = AccountStatus::Blocked;
    
    assert!(matches!(active, AccountStatus::Active));
    assert!(matches!(suspended, AccountStatus::Suspended));
    assert!(matches!(terminated, AccountStatus::Terminated));
    assert!(matches!(blocked, AccountStatus::Blocked));
}

#[test]
fn test_checked_add_no_overflow() {
    let value: u64 = u64::MAX - 100;
    let result = value.checked_add(100);
    assert!(result.is_some());
    assert_eq!(result.unwrap(), u64::MAX);
}

#[test]
fn test_checked_add_overflow() {
    let value: u64 = u64::MAX;
    let result = value.checked_add(1);
    assert!(result.is_none());
}

#[test]
fn test_checked_sub_no_underflow() {
    let value: u64 = 100;
    let result = value.checked_sub(50);
    assert!(result.is_some());
    assert_eq!(result.unwrap(), 50);
}

#[test]
fn test_checked_sub_underflow() {
    let value: u64 = 10;
    let result = value.checked_sub(20);
    assert!(result.is_none());
}

#[test]
fn test_integer_arithmetic_bytes_to_mb() {
    let bytes: u64 = 2_097_152; // 2 MB
    let mb = bytes / 1_048_576;
    let remainder = bytes % 1_048_576;
    
    assert_eq!(mb, 2);
    assert_eq!(remainder, 0);
}

#[test]
fn test_integer_arithmetic_bytes_to_mb_with_remainder() {
    let bytes: u64 = 1_500_000; // 1.43 MB
    let mb = bytes / 1_048_576;
    let remainder = bytes % 1_048_576;
    
    assert_eq!(mb, 1);
    assert_eq!(remainder, 451_424);
}

#[test]
fn test_integer_arithmetic_seconds_to_minutes() {
    let seconds: u64 = 120; // 2 minutes
    let minutes = seconds / 60;
    let remainder = seconds % 60;
    
    assert_eq!(minutes, 2);
    assert_eq!(remainder, 0);
}

#[test]
fn test_integer_arithmetic_seconds_to_minutes_with_remainder() {
    let seconds: u64 = 90; // 1.5 minutes
    let minutes = seconds / 60;
    let remainder = seconds % 60;
    
    assert_eq!(minutes, 1);
    assert_eq!(remainder, 30);
}
