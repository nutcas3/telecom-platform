use crate::client::HTTPClient;
use crate::error::TelecomError;
use crate::types::{
    CreateSubscriberRequest, Subscriber, SubscriberList, UpdateSubscriberRequest,
};
use std::collections::HashMap;

pub struct SubscriberAPI {
    client: HTTPClient,
}

impl SubscriberAPI {
    pub fn new(client: HTTPClient) -> Self {
        Self { client }
    }

    pub async fn get(&self, id: i64) -> Result<Subscriber, TelecomError> {
        self.client.get(&format!("/v1/subscribers/{}", id), None).await
    }

    pub async fn list(
        &self,
        page: i32,
        page_size: i32,
        status: Option<String>,
    ) -> Result<SubscriberList, TelecomError> {
        let mut params = HashMap::new();
        params.insert("page".to_string(), page.to_string());
        params.insert("page_size".to_string(), page_size.to_string());
        
        if let Some(status) = status {
            params.insert("status".to_string(), status);
        }

        self.client.get("/v1/subscribers", Some(&params)).await
    }

    pub async fn create(&self, req: &CreateSubscriberRequest) -> Result<Subscriber, TelecomError> {
        self.client.post("/v1/subscribers", Some(req)).await
    }

    pub async fn update(
        &self,
        id: i64,
        req: &UpdateSubscriberRequest,
    ) -> Result<Subscriber, TelecomError> {
        self.client
            .put(&format!("/v1/subscribers/{}", id), Some(req))
            .await
    }

    pub async fn delete(&self, id: i64) -> Result<(), TelecomError> {
        self.client.delete(&format!("/v1/subscribers/{}", id)).await
    }

    pub async fn suspend(&self, id: i64) -> Result<Subscriber, TelecomError> {
        self.client
            .post(&format!("/v1/subscribers/{}/suspend", id), None::<()>)
            .await
    }

    pub async fn activate(&self, id: i64) -> Result<Subscriber, TelecomError> {
        self.client
            .post(&format!("/v1/subscribers/{}/activate", id), None::<()>)
            .await
    }
}
