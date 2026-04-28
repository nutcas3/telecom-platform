//! Telecom Platform Rust SDK
//! 
//! Provides async/await support with tokio for the Telecom Platform API.

pub mod api;
pub mod auth;
pub mod client;
pub mod error;
pub mod types;

use auth::AuthProvider;
use client::HTTPClient;
use error::TelecomError;
use std::time::Duration;
use types::*;

pub use api::{GraphQLAPI, PaymentAPI, RatingPlanAPI, SubscriberAPI, SystemAPI, UsageAPI};

/// Telecom Platform SDK client
#[derive(Clone)]
pub struct TelecomClient {
    auth_provider: AuthProvider,
    http_client: HTTPClient,
    
    // API modules
    pub subscribers: SubscriberAPI,
    pub usage: UsageAPI,
    pub payments: PaymentAPI,
    pub rating_plans: RatingPlanAPI,
    pub system: SystemAPI,
    pub graphql: GraphQLAPI,
}

#[derive(Debug, Clone)]
pub struct TelecomConfig {
    pub api_url: String,
    pub api_key: Option<String>,
    pub jwt_secret: Option<String>,
    pub timeout: Duration,
    pub max_retries: u32,
    pub retry_delay: Duration,
}

impl Default for TelecomConfig {
    fn default() -> Self {
        Self {
            api_url: "http://localhost:8000".to_string(),
            api_key: None,
            jwt_secret: None,
            timeout: Duration::from_secs(30),
            max_retries: 3,
            retry_delay: Duration::from_secs(1),
        }
    }
}

impl TelecomClient {
    /// Create a new Telecom client
    pub fn new(config: TelecomConfig) -> Self {
        let auth_provider = AuthProvider::new(config.api_key.clone(), config.jwt_secret.clone());
        
        let http_client = HTTPClient::new(
            config.api_url.clone(),
            auth_provider.clone(),
            config.timeout,
            config.max_retries,
            config.retry_delay,
        );

        // Initialize API modules
        let subscribers = SubscriberAPI::new(http_client.clone());
        let usage = UsageAPI::new(http_client.clone());
        let payments = PaymentAPI::new(http_client.clone());
        let rating_plans = RatingPlanAPI::new(http_client.clone());
        let system = SystemAPI::new(http_client.clone());
        let graphql = GraphQLAPI::new(http_client.clone());

        Self {
            auth_provider,
            http_client,
            subscribers,
            usage,
            payments,
            rating_plans,
            system,
            graphql,
        }
    }

    /// Generate a JWT token for authentication
    pub fn generate_jwt_token(
        &mut self,
        user_id: String,
        expiry_hours: i64,
        additional_claims: std::collections::HashMap<String, serde_json::Value>,
    ) -> Result<String, String> {
        self.auth_provider.generate_jwt_token(user_id, expiry_hours, additional_claims)
    }

    /// Validate a JWT token
    pub fn validate_jwt_token(&self, token: &str) -> Result<JWTClaims, String> {
        self.auth_provider.validate_jwt_token(token)
    }

    /// Close the client and cleanup resources
    pub fn close(&self) {
        self.http_client.close();
    }
}
