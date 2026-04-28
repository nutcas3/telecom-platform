use charging_engine::errors::{validate_bytes, validate_session_id, validate_amount, validate_ip, ChargingError};

#[test]
fn test_validate_bytes_zero() {
    let result = validate_bytes(0);
    assert!(result.is_err());
    match result.unwrap_err() {
        ChargingError::InvalidInput(msg) => assert!(msg.contains("cannot be zero")),
        _ => panic!("Expected InvalidInput error"),
    }
}

#[test]
fn test_validate_bytes_exceeds_limit() {
    std::env::set_var("MAX_BYTES_LIMIT", "1000000000000");
    let result = validate_bytes(2_000_000_000_000); // 2TB
    assert!(result.is_err());
    match result.unwrap_err() {
        ChargingError::InvalidInput(_) => (),
        _ => panic!("Expected InvalidInput error"),
    }
}

#[test]
fn test_validate_bytes_valid() {
    let result = validate_bytes(1_000_000_000); // 1GB
    assert!(result.is_ok());
}

#[test]
fn test_validate_session_id_empty() {
    let result = validate_session_id("");
    assert!(result.is_err());
    match result.unwrap_err() {
        ChargingError::InvalidInput(msg) => assert!(msg.contains("cannot be empty")),
        _ => panic!("Expected InvalidInput error"),
    }
}

#[test]
fn test_validate_session_id_too_long() {
    let result = validate_session_id(&"a".repeat(65));
    assert!(result.is_err());
    match result.unwrap_err() {
        ChargingError::InvalidInput(msg) => assert!(msg.contains("too long")),
        _ => panic!("Expected InvalidInput error"),
    }
}

#[test]
fn test_validate_session_id_valid() {
    let result = validate_session_id("session-123");
    assert!(result.is_ok());
}

#[test]
fn test_validate_amount_zero() {
    let result = validate_amount(0.0);
    assert!(result.is_err());
    match result.unwrap_err() {
        ChargingError::InvalidInput(msg) => assert!(msg.contains("must be positive")),
        _ => panic!("Expected InvalidInput error"),
    }
}

#[test]
fn test_validate_amount_negative() {
    let result = validate_amount(-1.0);
    assert!(result.is_err());
    match result.unwrap_err() {
        ChargingError::InvalidInput(msg) => assert!(msg.contains("must be positive")),
        _ => panic!("Expected InvalidInput error"),
    }
}

#[test]
fn test_validate_amount_exceeds_limit() {
    std::env::set_var("MAX_AMOUNT_LIMIT", "10000");
    let result = validate_amount(20_000.0);
    assert!(result.is_err());
    match result.unwrap_err() {
        ChargingError::InvalidInput(_) => (),
        _ => panic!("Expected InvalidInput error"),
    }
}

#[test]
fn test_validate_amount_valid() {
    let result = validate_amount(100.0);
    assert!(result.is_ok());
}

#[test]
fn test_validate_ip_empty() {
    let result = validate_ip("");
    assert!(result.is_err());
    match result.unwrap_err() {
        ChargingError::InvalidInput(msg) => assert!(msg.contains("cannot be empty")),
        _ => panic!("Expected InvalidInput error"),
    }
}

#[test]
fn test_validate_ip_invalid_format() {
    let result = validate_ip("invalid");
    assert!(result.is_err());
    match result.unwrap_err() {
        ChargingError::InvalidInput(msg) => assert!(msg.contains("Invalid IP address format")),
        _ => panic!("Expected InvalidInput error"),
    }
}

#[test]
fn test_validate_ip_zero_octet() {
    let result = validate_ip("0.0.0.0");
    assert!(result.is_err());
    match result.unwrap_err() {
        ChargingError::InvalidInput(msg) => assert!(msg.contains("cannot be 0.0.0.0")),
        _ => panic!("Expected InvalidInput error"),
    }
}

#[test]
fn test_validate_ip_valid() {
    let result = validate_ip("192.168.1.1");
    assert!(result.is_ok());
}
