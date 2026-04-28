use crate::auth::AuthProvider;
use crate::error::TelecomError;
use reqwest::{Client, StatusCode};
use std::collections::HashMap;
use std::time::Duration;

/// HTTP client for making API requests
pub struct HTTPClient {
    client: Client,
    auth_provider: AuthProvider,
    base_url: String,
    max_retries: u32,
    retry_delay: Duration,
}

impl HTTPClient {
    pub fn new(
        base_url: String,
        auth_provider: AuthProvider,
        timeout: Duration,
        max_retries: u32,
        retry_delay: Duration,
    ) -> Self {
        let client = Client::builder()
            .timeout(timeout)
            .build()
            .expect("Failed to create HTTP client");

        Self {
            client,
            auth_provider,
            base_url,
            max_retries,
            retry_delay,
        }
    }

    pub async fn get<T>(&self, path: &str, params: Option<&HashMap<String, String>>) -> Result<T, TelecomError>
    where
        T: serde::de::DeserializeOwned,
    {
        self.request("GET", path, None, params).await
    }

    pub async fn post<T>(&self, path: &str, body: Option<&impl serde::Serialize>) -> Result<T, TelecomError>
    where
        T: serde::de::DeserializeOwned,
    {
        self.request("POST", path, body, None).await
    }

    pub async fn put<T>(&self, path: &str, body: Option<&impl serde::Serialize>) -> Result<T, TelecomError>
    where
        T: serde::de::DeserializeOwned,
    {
        self.request("PUT", path, body, None).await
    }

    pub async fn delete(&self, path: &str) -> Result<(), TelecomError> {
        self.request::<()>("DELETE", path, None, None).await
    }

    async fn request<T>(
        &self,
        method: &str,
        path: &str,
        body: Option<&impl serde::Serialize>,
        params: Option<&HashMap<String, String>>,
    ) -> Result<T, TelecomError>
    where
        T: serde::de::DeserializeOwned,
    {
        let url = format!("{}{}", self.base_url, path);
        let mut request = self.client.request(method, &url);

        // Set headers
        let headers = self.auth_provider.get_headers();
        for (key, value) in headers {
            request = request.header(&key, &value);
        }

        // Add query parameters
        if let Some(params) = params {
            request = request.query(params);
        }

        // Add body
        if let Some(body) = body {
            request = request.json(body);
        }

        // Retry logic
        let mut last_error = None;
        for attempt in 0..=self.max_retries {
            match request.try_clone().unwrap().send().await {
                Ok(response) => {
                    let status = response.status();
                    
                    if status == StatusCode::UNAUTHORIZED {
                        return Err(TelecomError::AuthenticationError);
                    } else if status == StatusCode::TOO_MANY_REQUESTS {
                        return Err(TelecomError::RateLimitError);
                    } else if status.is_client_error() {
                        let error_text = response.text().await.unwrap_or_else(|_| "Bad request".to_string());
                        return Err(TelecomError::APIError(error_text));
                    } else if status.is_server_error() {
                        return Err(TelecomError::ServerError(status));
                    }

                    return response.json::<T>().await.map_err(TelecomError::from);
                }
                Err(e) => {
                    last_error = Some(TelecomError::NetworkError(e));
                    if attempt < self.max_retries {
                        tokio::time::sleep(self.retry_delay * (1 << attempt)).await;
                    }
                }
            }
        }

        Err(last_error.unwrap_or_else(|| TelecomError::APIError("Request failed".to_string())))
    }

    pub fn close(&self) {
        // HTTP client doesn't need explicit closing
    }
}
