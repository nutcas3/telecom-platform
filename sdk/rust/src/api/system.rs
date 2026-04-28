use crate::client::HTTPClient;
use crate::error::TelecomError;
use crate::types::{HealthStatus, SystemStats};

pub struct SystemAPI {
    client: HTTPClient,
}

impl SystemAPI {
    pub fn new(client: HTTPClient) -> Self {
        Self { client }
    }

    pub async fn get_stats(&self) -> Result<SystemStats, TelecomError> {
        self.client.get("/v1/system/stats", None).await
    }

    pub async fn get_health(&self) -> Result<HealthStatus, TelecomError> {
        self.client.get("/v1/health", None).await
    }
}
