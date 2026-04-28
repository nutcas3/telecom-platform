use charging_engine::errors::{ChargingError, log_error};

#[test]
fn test_charging_error_display() {
    let err = ChargingError::RedisConnection("connection failed".to_string());
    assert_eq!(err.to_string(), "Redis connection error: connection failed");
    
    let err = ChargingError::RedisOperation("operation failed".to_string());
    assert_eq!(err.to_string(), "Redis operation error: operation failed");
    
    let err = ChargingError::DatabaseError("query failed".to_string());
    assert_eq!(err.to_string(), "Database error: query failed");
    
    let err = ChargingError::SubscriberNotFound("12345".to_string());
    assert_eq!(err.to_string(), "Subscriber not found: 12345");
    
    let err = ChargingError::RatingPlanNotFound("basic".to_string());
    assert_eq!(err.to_string(), "Rating plan not found: basic");
    
    let err = ChargingError::InsufficientCredit { available: 100, requested: 200 };
    assert_eq!(err.to_string(), "Insufficient credit: available=100, requested=200");
    
    let err = ChargingError::UsageBlocked("user blocked".to_string());
    assert_eq!(err.to_string(), "Usage blocked: user blocked");
    
    let err = ChargingError::InvalidInput("bad input".to_string());
    assert_eq!(err.to_string(), "Invalid input: bad input");
    
    let err = ChargingError::SerializationError("json error".to_string());
    assert_eq!(err.to_string(), "Serialization error: json error");
    
    let err = ChargingError::InternalError("internal error".to_string());
    assert_eq!(err.to_string(), "Internal error: internal error");
}

#[test]
fn test_charging_error_clone() {
    let err1 = ChargingError::RedisConnection("test".to_string());
    let err2 = err1.clone();
    assert_eq!(err1.to_string(), err2.to_string());
}

#[test]
fn test_log_error() {
    // This test just ensures log_error doesn't panic
    log_error(&ChargingError::RedisConnection("test".to_string()));
    log_error(&ChargingError::DatabaseError("test".to_string()));
    log_error(&ChargingError::SubscriberNotFound("123".to_string()));
    log_error(&ChargingError::InsufficientCredit { available: 100, requested: 200 });
    log_error(&ChargingError::InvalidInput("test".to_string()));
}
