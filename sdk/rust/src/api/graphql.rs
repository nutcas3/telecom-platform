use crate::client::HTTPClient;
use crate::error::TelecomError;
use crate::types::GraphQLRequest;
use std::collections::HashMap;

pub struct GraphQLAPI {
    client: HTTPClient,
}

impl GraphQLAPI {
    pub fn new(client: HTTPClient) -> Self {
        Self { client }
    }

    pub async fn execute(
        &self,
        query: String,
        variables: Option<HashMap<String, serde_json::Value>>,
    ) -> Result<HashMap<String, serde_json::Value>, TelecomError> {
        let req = GraphQLRequest { query, variables };
        self.client.post("/graphql", Some(&req)).await
    }
}
