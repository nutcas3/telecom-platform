use chrono::{DateTime, Utc};
use hmac::{Hmac, Mac};
use serde::{Deserialize, Serialize};
use sha2::Sha256;
use std::collections::HashMap;

type HmacSha256 = Hmac<Sha256>;

/// Authentication provider for the Telecom SDK
#[derive(Clone)]
pub struct AuthProvider {
    api_key: Option<String>,
    jwt_secret: Option<String>,
    token_cache: Option<String>,
    token_expiry: Option<DateTime<Utc>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct JWTClaims {
    pub sub: String,
    pub exp: i64,
    pub iat: i64,
    #[serde(flatten)]
    pub additional: HashMap<String, serde_json::Value>,
}

impl AuthProvider {
    pub fn new(api_key: Option<String>, jwt_secret: Option<String>) -> Self {
        Self {
            api_key,
            jwt_secret,
            token_cache: None,
            token_expiry: None,
        }
    }

    pub fn get_headers(&self) -> HashMap<String, String> {
        let mut headers = HashMap::new();
        headers.insert("Content-Type".to_string(), "application/json".to_string());
        headers.insert("User-Agent".to_string(), "Telecom-Rust-SDK/1.0.0".to_string());

        if let Some(ref api_key) = self.api_key {
            headers.insert("X-API-Key".to_string(), api_key.clone());
        }

        if let Some(ref token) = self.token_cache {
            if self.is_token_valid() {
                headers.insert("Authorization".to_string(), format!("Bearer {}", token));
            }
        }

        headers
    }

    pub fn generate_jwt_token(
        &mut self,
        user_id: String,
        expiry_hours: i64,
        additional_claims: HashMap<String, serde_json::Value>,
    ) -> Result<String, String> {
        let jwt_secret = self.jwt_secret.as_ref().ok_or("JWT secret not configured")?;

        let now = Utc::now().timestamp();
        let exp = now + (expiry_hours * 3600);

        let mut claims = HashMap::new();
        claims.insert("sub".to_string(), serde_json::to_value(&user_id).unwrap());
        claims.insert("exp".to_string(), serde_json::to_value(&exp).unwrap());
        claims.insert("iat".to_string(), serde_json::to_value(&now).unwrap());

        for (k, v) in additional_claims {
            claims.insert(k, v);
        }

        let header = json!({"alg": "HS256", "typ": "JWT"});
        let encoded_header = self.base64_url_encode(&header.to_string());
        let encoded_payload = self.base64_url_encode(&json!(claims).to_string());
        let signature = self.sign(&format!("{}.{}", encoded_header, encoded_payload), jwt_secret);

        let token = format!("{}.{}.{}", encoded_header, encoded_payload, signature);
        self.token_cache = Some(token.clone());
        self.token_expiry = Some(DateTime::from_timestamp(exp, 0).unwrap());

        Ok(token)
    }

    pub fn validate_jwt_token(&self, token: &str) -> Result<JWTClaims, String> {
        let jwt_secret = self.jwt_secret.as_ref().ok_or("JWT secret not configured")?;

        let parts: Vec<&str> = token.split('.').collect();
        if parts.len() != 3 {
            return Err("Invalid token format".to_string());
        }

        let (encoded_header, encoded_payload, signature) = (parts[0], parts[1], parts[2]);
        let expected_signature = self.sign(&format!("{}.{}", encoded_header, encoded_payload), jwt_secret);

        if signature != expected_signature {
            return Err("Invalid token signature".to_string());
        }

        let payload_bytes = self.base64_url_decode(encoded_payload)?;
        let claims: JWTClaims = serde_json::from_slice(&payload_bytes)
            .map_err(|e| format!("Failed to decode payload: {}", e))?;

        if claims.exp < Utc::now().timestamp() {
            return Err("Token has expired".to_string());
        }

        Ok(claims)
    }

    pub fn clear_token_cache(&mut self) {
        self.token_cache = None;
        self.token_expiry = None;
    }

    fn is_token_valid(&self) -> bool {
        if self.token_cache.is_none() || self.token_expiry.is_none() {
            return false;
        }
        Utc::now() < self.token_expiry.unwrap()
    }

    fn base64_url_encode(&self, data: &str) -> String {
        use base64::{engine::general_purpose::URL_SAFE_NO_PAD, Engine};
        URL_SAFE_NO_PAD.encode(data.as_bytes())
    }

    fn base64_url_decode(&self, data: &str) -> Result<Vec<u8>, String> {
        use base64::{engine::general_purpose::URL_SAFE_NO_PAD, Engine};
        URL_SAFE_NO_PAD
            .decode(data)
            .map_err(|e| format!("Failed to decode: {}", e))
    }

    fn sign(&self, data: &str, secret: &str) -> String {
        let mut mac = HmacSha256::new_from_slice(secret.as_bytes()).expect("HMAC can take key of any size");
        mac.update(data.as_bytes());
        let result = mac.finalize();
        self.base64_url_encode(&format!("{:x}", result.into_bytes()))
    }
}
