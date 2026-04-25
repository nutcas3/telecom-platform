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
}

impl AuthConfig {
    pub fn from_env() -> Self {
        let secret = env::var("JWT_SECRET").unwrap_or_else(|_| "default_secret_key_change_in_production".to_string());
        Self { secret }
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
    let auth_header = request
        .headers()
        .get("Authorization")
        .and_then(|h| h.to_str().ok());

    if let Some(auth_header) = auth_header {
        if let Some(token) = auth_header.strip_prefix("Bearer ") {
            match validate_token(token, &config) {
                Ok(_claims) => return Ok(next.run(request).await),
                Err(_) => return Err(StatusCode::UNAUTHORIZED),
            }
        }
    }

    // For now, allow requests without auth for packet-gateway integration
    // In production, you would want to enable strict authentication
    // or use API keys for service-to-service communication
    Ok(next.run(request).await)
}
