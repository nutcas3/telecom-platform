use axum::{
    extract::{Request, State},
    http::StatusCode,
    middleware::Next,
    response::Response,
};
use jsonwebtoken::{decode, encode, DecodingKey, EncodingKey, Header, Validation};
use serde::{Deserialize, Serialize};
use std::env;

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Claims {
    pub sub: String,
    pub exp: usize,
}

#[derive(Clone)]
pub struct AuthConfig {
    pub secret: String,
    pub api_keys: Vec<String>,
}

impl AuthConfig {
    pub fn from_env() -> Self {
        let secret = env::var("JWT_SECRET").unwrap_or_else(|_| "default_secret_key_change_in_production".to_string());
        
        // Parse API keys from comma-separated environment variable
        let api_keys_env = env::var("API_KEYS").unwrap_or_else(|_| String::new());
        let api_keys: Vec<String> = api_keys_env
            .split(',')
            .map(|s| s.trim().to_string())
            .filter(|s| !s.is_empty())
            .collect();
        
        Self { secret, api_keys }
    }
    
    pub fn validate_api_key(&self, api_key: &str) -> bool {
        self.api_keys.iter().any(|key| key == api_key)
    }
}

pub fn create_token(user_id: &str, config: &AuthConfig) -> Result<String, StatusCode> {
    let expiration = chrono::Utc::now()
        .checked_add_signed(chrono::Duration::hours(24))
        .expect("valid timestamp")
        .timestamp() as usize;

    let claims = Claims {
        sub: user_id.to_string(),
        exp: expiration,
    };

    encode(
        &Header::default(),
        &claims,
        &EncodingKey::from_secret(config.secret.as_ref()),
    )
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)
}

pub fn validate_token(token: &str, config: &AuthConfig) -> Result<Claims, StatusCode> {
    decode::<Claims>(
        token,
        &DecodingKey::from_secret(config.secret.as_ref()),
        &Validation::default(),
    )
    .map(|data| data.claims)
    .map_err(|_| StatusCode::UNAUTHORIZED)
}

pub async fn auth_middleware(
    State(config): State<AuthConfig>,
    request: Request,
    next: Next,
) -> Result<Response, StatusCode> {
    // Check if authentication is enforced via environment variable
    let enforce_auth = env::var("ENFORCE_AUTH")
        .unwrap_or_else(|_| "true".to_string())
        .to_lowercase() == "true";

    if !enforce_auth {
        // Allow requests without auth for development/testing
        return Ok(next.run(request).await);
    }

    let auth_header = request
        .headers()
        .get("Authorization")
        .and_then(|h| h.to_str().ok());

    // Check for API key authentication
    let api_key_header = request
        .headers()
        .get("X-API-Key")
        .and_then(|h| h.to_str().ok());

    if let Some(api_key) = api_key_header {
        if config.validate_api_key(api_key) {
            return Ok(next.run(request).await);
        }
    }

    // Check for JWT authentication
    if let Some(auth_header) = auth_header {
        if let Some(token) = auth_header.strip_prefix("Bearer ") {
            match validate_token(token, &config) {
                Ok(_claims) => return Ok(next.run(request).await),
                Err(_) => return Err(StatusCode::UNAUTHORIZED),
            }
        }
    }

    // No valid authentication found
    Err(StatusCode::UNAUTHORIZED)
}
