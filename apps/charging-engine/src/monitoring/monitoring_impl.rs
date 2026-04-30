use crate::errors::ChargingResult;

impl crate::charging::ChargingEngine {
    pub async fn get_performance_metrics(&self) -> ChargingResult<PerformanceMetrics> {
        let mut conn = self.redis_client.get_multiplexed_async_connection().await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

        let info: String = redis::cmd("INFO").query_async(&mut conn).await.unwrap_or_default();
        
        let connected_clients = self.extract_metric(&info, "connected_clients");
        let used_memory = self.extract_metric(&info, "used_memory");
        let total_commands_processed = self.extract_metric(&info, "total_commands_processed");

        let metrics = PerformanceMetrics {
            connected_clients,
            used_memory,
            total_commands_processed,
            requests_per_second: self.calculate_rps().await?,
            average_response_time: self.calculate_avg_response_time().await?,
        };

        Ok(metrics)
    }

    fn extract_metric(&self, info: &str, metric: &str) -> u64 {
        info.lines()
            .find(|line| line.starts_with(metric))
            .and_then(|line| line.split(':').nth(1))
            .and_then(|value| value.parse::<u64>().ok())
            .unwrap_or(0)
    }

    async fn calculate_rps(&self) -> ChargingResult<f64> {
        let mut conn = self.redis_client.get_multiplexed_async_connection().await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

        let info: String = redis::cmd("INFO").query_async(&mut conn).await.unwrap_or_default();
        
        for line in info.lines() {
            if line.starts_with("instantaneous_ops_per_sec:") {
                if let Some(rps_str) = line.split(':').nth(1) {
                    if let Ok(rps) = rps_str.parse::<f64>() {
                        return Ok(rps);
                    }
                }
            }
        }
        
        let total_commands = self.extract_metric(&info, "total_commands_processed");
        let uptime_seconds = self.get_uptime().await?;
        
        if uptime_seconds > 0 {
            Ok(total_commands as f64 / uptime_seconds as f64)
        } else {
            Ok(0.0)
        }
    }

    async fn calculate_avg_response_time(&self) -> ChargingResult<f64> {
        let mut conn = self.redis_client.get_multiplexed_async_connection().await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

        let response_times_key = "metrics:response_times";
        
        let response_times_raw: Vec<String> = redis::AsyncCommands::lrange(&mut conn, response_times_key, 0, 99)
            .await
            .unwrap_or_default();
        let response_times: Vec<f64> = response_times_raw
            .into_iter()
            .filter_map(|s| s.parse::<f64>().ok())
            .collect();

        if response_times.is_empty() {
            let latency_info = self.get_redis_latency(&mut conn).await?;
            Ok(latency_info)
        } else {
            let sum: f64 = response_times.iter().sum();
            Ok(sum / response_times.len() as f64)
        }
    }

    async fn get_redis_latency(&self, conn: &mut redis::aio::MultiplexedConnection) -> ChargingResult<f64> {
        use std::time::Instant;
        
        let start = Instant::now();
        let _: String = redis::cmd("PING").query_async(conn).await.unwrap_or_else(|_| "PONG".to_string());
        let latency = start.elapsed();
        
        Ok(latency.as_millis() as f64)
    }

    pub async fn get_error_statistics(&self) -> ChargingResult<ErrorStats> {
        let mut conn = self.redis_client.get_multiplexed_async_connection().await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

        let pattern = "error:*".to_string();
        let keys: Vec<String> = redis::AsyncCommands::keys(&mut conn, &pattern).await
            .map_err(|e| crate::errors::ChargingError::RedisOperation(e.to_string()))?;

        let mut total_errors = 0u64;
        let mut error_types = std::collections::HashMap::new();

        for key in keys {
            if let Ok(count) = redis::AsyncCommands::get::<_, u64>(&mut conn, &key).await {
                total_errors += count;
                
                let error_type = key.split(':').nth(1).unwrap_or("unknown");
                *error_types.entry(error_type.to_string()).or_insert(0) += count;
            }
        }

        let stats = ErrorStats {
            total_errors,
            error_types,
            last_error: self.get_last_error().await?,
        };

        Ok(stats)
    }

    async fn get_last_error(&self) -> ChargingResult<Option<String>> {
        let mut conn = self.redis_client.get_multiplexed_async_connection().await
            .map_err(|e| crate::errors::ChargingError::RedisConnection(e.to_string()))?;

        let last_error: Option<String> = redis::AsyncCommands::get(&mut conn, "last_error").await.unwrap_or(None);
        Ok(last_error)
    }
}

#[allow(dead_code)]
#[derive(Debug, Clone)]
pub struct PerformanceMetrics {
    pub connected_clients: u64,
    pub used_memory: u64,
    pub total_commands_processed: u64,
    pub requests_per_second: f64,
    pub average_response_time: f64,
}

#[allow(dead_code)]
#[derive(Debug, Clone)]
pub struct ErrorStats {
    pub total_errors: u64,
    pub error_types: std::collections::HashMap<String, u64>,
    pub last_error: Option<String>,
}
