use crate::client::HTTPClient;
use crate::error::TelecomError;
use crate::types::RatingPlan;

pub struct RatingPlanAPI {
    client: HTTPClient,
}

impl RatingPlanAPI {
    pub fn new(client: HTTPClient) -> Self {
        Self { client }
    }

    pub async fn list(&self) -> Result<Vec<RatingPlan>, TelecomError> {
        self.client.get("/v1/rating-plans", None).await
    }

    pub async fn get(&self, plan_id: &str) -> Result<RatingPlan, TelecomError> {
        self.client
            .get(&format!("/v1/rating-plans/{}", plan_id), None)
            .await
    }
}
