#[derive(Debug, Clone)]
pub struct SystemStats {
    pub active_sessions: u64,
    pub total_accounts: u64,
    pub blocked_users: u64,
    pub low_balance_alerts: u64,
    pub uptime: u64,
}

#[derive(Debug, Clone)]
pub struct HealthStatus {
    pub redis_connected: bool,
    pub active_sync: bool,
    pub last_sync: chrono::DateTime<chrono::Utc>,
    pub memory_usage: u64,
}

#[derive(Debug, Clone)]
#[allow(dead_code)]
pub struct PerformanceMetrics {
    pub connected_clients: u64,
    pub used_memory: u64,
    pub total_commands_processed: u64,
    pub requests_per_second: f64,
    pub average_response_time: f64,
}

#[derive(Debug, Clone)]
#[allow(dead_code)]
pub struct ErrorStats {
    pub total_errors: u64,
    pub error_types: std::collections::HashMap<String, u64>,
    pub last_error: Option<String>,
}
