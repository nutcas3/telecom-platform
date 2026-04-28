use axum::{
    http::StatusCode,
    response::{IntoResponse, Response},
};
use tower_governor::{
    governor::GovernorConfig,
    GovernorLayer,
};

pub fn create_rate_limiter() -> GovernorLayer {
    let governor_conf = GovernorConfig::default();

    GovernorLayer::new(governor_conf)
}

pub struct RateLimitResponse;

impl IntoResponse for RateLimitResponse {
    fn into_response(self) -> Response {
        (StatusCode::TOO_MANY_REQUESTS, "Rate limit exceeded").into_response()
    }
}
