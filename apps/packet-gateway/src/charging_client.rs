use anyhow::{Context, Result};
use chrono::Utc;
use reqwest::Client;
use serde::{Deserialize, Serialize};
use tracing::{info, warn, error};

#[derive(Debug, Serialize)]
struct UsageEvent {
    imsi: String,
    session_id: String,
    volume: f64,
    usage_type: String,
    timestamp: i64,
}

#[derive(Debug, Deserialize)]
struct UsageResponse {
    success: bool,
    message: Option<String>,
}

pub struct ChargingEngineClient {
    client: Client,
    base_url: String,
}

impl ChargingEngineClient {
    pub fn new(base_url: String) -> Self {
        Self {
            client: Client::new(),
            base_url,
        }
    }

    pub async fn report_usage(
        &self,
        imsi: String,
        session_id: String,
        volume_bytes: u64,
    ) -> Result<()> {
        let volume_mb = volume_bytes as f64 / (1024.0 * 1024.0);
        let timestamp = Utc::now().timestamp();

        let event = UsageEvent {
            imsi: imsi.clone(),
            session_id,
            volume: volume_mb,
            usage_type: "Data".to_string(),
            timestamp,
        };

        let url = format!("{}/v1/usage/process", self.base_url);
        
        match self
            .client
            .post(&url)
            .json(&event)
            .send()
            .await
        {
            Ok(response) => {
                if response.status().is_success() {
                    info!("Successfully reported usage for IMSI {}: {} MB", imsi, volume_mb);
                    Ok(())
                } else {
                    let status = response.status();
                    let body = response.text().await.unwrap_or_default();
                    warn!(
                        "Failed to report usage for IMSI {}: {} - {}",
                        imsi, status, body
                    );
                    Err(anyhow::anyhow!(
                        "Charging engine returned error {}: {}",
                        status,
                        body
                    ))
                }
            }
            Err(e) => {
                error!("Failed to send usage to charging engine: {}", e);
                Err(anyhow::anyhow!("Failed to send usage to charging engine: {}", e))
            }
        }
    }

    pub async fn report_batch_usage(
        &self,
        usage_data: Vec<(String, String, u64)>,
    ) -> Result<()> {
        for (imsi, session_id, volume_bytes) in usage_data {
            if let Err(e) = self.report_usage(imsi, session_id, volume_bytes).await {
                warn!("Failed to report usage for one session: {}", e);
                // Continue with other sessions instead of failing the entire batch
            }
        }
        Ok(())
    }
}
