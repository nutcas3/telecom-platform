use axum::{
    http::StatusCode,
    response::{IntoResponse, Response},
};
use governor::{
    clock::DefaultClock,
    state::InMemoryState,
    Quota,
};
use std::num::NonZeroU32;
use std::sync::Arc;
use tower_governor::{
    governor::GovernorConfig,
    key_extractor::SmartIpKeyExtractor,
    GovernorLayer,
};

pub fn create_rate_limiter() -> GovernorLayer<SmartIpKeyExtractor, InMemoryState, DefaultClock> {
    let requests_per_second: u32 = std::env::var("RATE_LIMIT_RPS")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(10);
    let burst_size: u32 = std::env::var("RATE_LIMIT_BURST")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(20);

    let quota = Quota::per_second(NonZeroU32::new(requests_per_second).unwrap())
        .allow_burst(NonZeroU32::new(burst_size).unwrap());

    let config = GovernorConfig::builder()
        .per_second(quota)
        .burst_size(burst_size)
        .finish()
        .unwrap();

    GovernorLayer::new(&config)
}

pub struct RateLimitResponse;

impl IntoResponse for RateLimitResponse {
    fn into_response(self) -> Response {
        (StatusCode::TOO_MANY_REQUESTS, "Rate limit exceeded").into_response()
    }
}
