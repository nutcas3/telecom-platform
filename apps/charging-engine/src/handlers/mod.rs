// Handler modules
pub mod block;
pub mod credit;
pub mod engine;
pub mod health;
pub mod monitoring;
pub mod rating;
pub mod subscriber;
pub mod usage;

// Re-export all handlers for convenience
pub use block::{block_user, is_user_blocked, unblock_user};
pub use credit::{add_credit, check_credit, deduct_credit, get_balance};
pub use engine::{engine_start, engine_stop, engine_uptime};
pub use health::health_check;
pub use monitoring::{detailed_health_check, get_error_stats, get_performance_metrics, get_system_stats, start_sync};
pub use rating::{add_rating_plan, get_rating_plan, list_rating_plans, remove_rating_plan};
pub use subscriber::{get_subscriber, update_subscriber};
pub use usage::{calculate_usage_cost, generate_invoice, process_usage, rate_usage, record_usage};
