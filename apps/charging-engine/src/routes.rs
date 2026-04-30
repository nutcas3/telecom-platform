use axum::{
    routing::{delete, get, post, put},
    Router,
};
use tower_http::cors::{Any, CorsLayer};
use tower_governor::{
    governor::GovernorConfigBuilder,
    GovernorLayer,
};

use crate::auth::auth_middleware;
use crate::handlers::{
    add_credit, add_rating_plan, block_user, check_credit, deduct_credit, detailed_health_check,
    engine_start, engine_stop, engine_uptime, get_balance, get_error_stats, get_performance_metrics,
    get_rating_plan, get_subscriber, get_system_stats, is_user_blocked, list_rating_plans,
    record_usage, remove_rating_plan, start_sync, unblock_user, update_subscriber, health_check,
    calculate_usage_cost, rate_usage, process_usage, generate_invoice,
};
use crate::models::AppState;

pub fn create_router(state: AppState) -> Router {
    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let requests_per_second: u32 = std::env::var("RATE_LIMIT_RPS")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(10);
    let burst_size: u32 = std::env::var("RATE_LIMIT_BURST")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(20);

    // Build governor config, fallback to default if None
    let governor_conf = GovernorConfigBuilder::default()
        .per_second(requests_per_second as u64)
        .burst_size(burst_size)
        .finish()
        .unwrap_or_else(|| {
            GovernorConfigBuilder::default()
                .per_second(10)
                .burst_size(20)
                .finish()
                .unwrap()
        });

    Router::new()
        // Public routes (no auth required - for packet-gateway integration)
        .route("/v1/health", get(health_check))
        .route("/v1/credit/:ip/check", post(check_credit))
        .route("/v1/credit/:ip/deduct", post(deduct_credit))
        .route("/v1/credit/:ip/balance", get(get_balance))
        .route("/v1/usage", post(record_usage))
        // Protected routes (require auth)
        .route("/v1/credit/:ip/add", post(add_credit))
        .route("/v1/subscriber/:imsi", get(get_subscriber))
        .route("/v1/subscriber/:imsi", put(update_subscriber))
        .route("/v1/rating-plans", get(list_rating_plans))
        .route("/v1/rating-plans/:id", get(get_rating_plan))
        .route("/v1/rating-plans", post(add_rating_plan))
        .route("/v1/rating-plans/:id", delete(remove_rating_plan))
        .route("/v1/block/:ip", post(block_user))
        .route("/v1/unblock/:ip", post(unblock_user))
        .route("/v1/blocked/:ip", get(is_user_blocked))
        .route("/v1/stats", get(get_system_stats))
        .route("/v1/metrics", get(get_performance_metrics))
        .route("/v1/errors", get(get_error_stats))
        .route("/v1/sync/start", post(start_sync))
        .route("/v1/health/detailed", get(detailed_health_check))
        .route("/v1/engine/start", post(engine_start))
        .route("/v1/engine/stop", post(engine_stop))
        .route("/v1/engine/uptime", get(engine_uptime))
        .route("/v1/usage/calculate-cost", post(calculate_usage_cost))
        .route("/v1/usage/rate", post(rate_usage))
        .route("/v1/usage/process", post(process_usage))
        .route("/v1/invoice/:imsi/:period", get(generate_invoice))
        .layer(cors)
        .layer(GovernorLayer::new(governor_conf))
        .route_layer(axum::middleware::from_fn_with_state(state.auth_config.clone(), auth_middleware))
        .with_state(state)
}
