use crate::client::HTTPClient;
use crate::error::TelecomError;
use crate::types::{RealTimeUsage, UsageEventList, UsageStats};
use chrono::{DateTime, Utc};
use std::collections::HashMap;

pub struct UsageAPI {
    client: HTTPClient,
}

impl UsageAPI {
    pub fn new(client: HTTPClient) -> Self {
        Self { client }
    }

    pub async fn get_stats(
        &self,
        subscriber_id: i64,
        start_date: DateTime<Utc>,
        end_date: DateTime<Utc>,
    ) -> Result<UsageStats, TelecomError> {
        let mut params = HashMap::new();
        params.insert("start_date".to_string(), start_date.to_rfc3339());
        params.insert("end_date".to_string(), end_date.to_rfc3339());

        self.client
            .get(&format!("/v1/subscribers/{}/usage", subscriber_id), Some(&params))
            .await
    }

    pub async fn list_events(
        &self,
        subscriber_id: Option<i64>,
        usage_type: Option<String>,
        start_date: Option<DateTime<Utc>>,
        end_date: Option<DateTime<Utc>>,
        page: i32,
        page_size: i32,
    ) -> Result<UsageEventList, TelecomError> {
        let mut params = HashMap::new();
        params.insert("page".to_string(), page.to_string());
        params.insert("page_size".to_string(), page_size.to_string());

        if let Some(subscriber_id) = subscriber_id {
            params.insert("subscriber_id".to_string(), subscriber_id.to_string());
        }
        if let Some(usage_type) = usage_type {
            params.insert("usage_type".to_string(), usage_type);
        }
        if let Some(start_date) = start_date {
            params.insert("start_date".to_string(), start_date.to_rfc3339());
        }
        if let Some(end_date) = end_date {
            params.insert("end_date".to_string(), end_date.to_rfc3339());
        }

        self.client.get("/v1/usage/events", Some(&params)).await
    }

    pub async fn get_realtime(&self, subscriber_id: i64) -> Result<RealTimeUsage, TelecomError> {
        self.client
            .get(&format!("/v1/subscribers/{}/realtime", subscriber_id), None)
            .await
    }
}
