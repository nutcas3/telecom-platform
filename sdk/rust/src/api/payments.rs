use crate::client::HTTPClient;
use crate::error::TelecomError;
use crate::types::{CreatePaymentRequest, PaymentTransaction, SubscriberList};
use std::collections::HashMap;

pub struct PaymentAPI {
    client: HTTPClient,
}

impl PaymentAPI {
    pub fn new(client: HTTPClient) -> Self {
        Self { client }
    }

    pub async fn create_transaction(&self, req: &CreatePaymentRequest) -> Result<PaymentTransaction, TelecomError> {
        self.client.post("/v1/payments/transactions", Some(req)).await
    }

    pub async fn get_transaction(&self, transaction_id: &str) -> Result<PaymentTransaction, TelecomError> {
        self.client
            .get(&format!("/v1/payments/transactions/{}", transaction_id), None)
            .await
    }

    pub async fn list_transactions(
        &self,
        subscriber_id: Option<i64>,
        status: Option<String>,
        page: i32,
        page_size: i32,
    ) -> Result<SubscriberList, TelecomError> {
        let mut params = HashMap::new();
        params.insert("page".to_string(), page.to_string());
        params.insert("page_size".to_string(), page_size.to_string());

        if let Some(subscriber_id) = subscriber_id {
            params.insert("subscriber_id".to_string(), subscriber_id.to_string());
        }
        if let Some(status) = status {
            params.insert("status".to_string(), status);
        }

        self.client.get("/v1/payments/transactions", Some(&params)).await
    }
}
