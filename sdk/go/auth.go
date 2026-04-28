package telecom

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// AuthProvider handles authentication for the Telecom SDK
type AuthProvider struct {
	apiKey     string
	jwtSecret  string
	tokenCache string
	tokenExpiry time.Time
}

// NewAuthProvider creates a new authentication provider
func NewAuthProvider(apiKey, jwtSecret string) *AuthProvider {
	return &AuthProvider{
		apiKey:    apiKey,
		jwtSecret: jwtSecret,
	}
}

// GetHeaders returns authentication headers for API requests
func (a *AuthProvider) GetHeaders() map[string]string {
	headers := map[string]string{
		"Content-Type": "application/json",
		"User-Agent":    "Telecom-Go-SDK/1.0.0",
	}

	if a.apiKey != "" {
		headers["X-API-Key"] = a.apiKey
	}

	if a.tokenCache != "" && a.isTokenValid() {
		headers["Authorization"] = "Bearer " + a.tokenCache
	}

	return headers
}

// GenerateJWTToken generates a JWT token for authentication
func (a *AuthProvider) GenerateJWTToken(userID string, expiryHours int, additionalClaims map[string]interface{}) (string, error) {
	if a.jwtSecret == "" {
		return "", fmt.Errorf("JWT secret not configured")
	}

	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub": userID,
		"exp": now + int64(expiryHours*3600),
		"iat": now,
	}

	for k, v := range additionalClaims {
		claims[k] = v
	}

	header := map[string]interface{}{
		"alg": "HS256",
		"typ": "JWT",
	}

	encodedHeader := a.base64URLEncode(header)
	encodedPayload := a.base64URLEncode(claims)
	signature := a.sign(encodedHeader + "." + encodedPayload)

	token := encodedHeader + "." + encodedPayload + "." + signature
	a.tokenCache = token
	a.tokenExpiry = time.Unix(claims["exp"].(int64), 0)

	return token, nil
}

// ValidateJWTToken validates a JWT token
func (a *AuthProvider) ValidateJWTToken(token string) (map[string]interface{}, error) {
	if a.jwtSecret == "" {
		return nil, fmt.Errorf("JWT secret not configured")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	encodedHeader, encodedPayload, signature := parts[0], parts[1], parts[2]
	expectedSignature := a.sign(encodedHeader + "." + encodedPayload)

	if signature != expectedSignature {
		return nil, fmt.Errorf("invalid token signature")
	}

	payloadBytes, err := a.base64URLDecode(encodedPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if exp, ok := payload["exp"].(float64); ok {
		if float64(time.Now().Unix()) > exp {
			return nil, fmt.Errorf("token has expired")
		}
	}

	return payload, nil
}

// ClearTokenCache clears the cached JWT token
func (a *AuthProvider) ClearTokenCache() {
	a.tokenCache = ""
	a.tokenExpiry = time.Time{}
}

// isTokenValid checks if the cached token is still valid
func (a *AuthProvider) isTokenValid() bool {
	if a.tokenCache == "" || a.tokenExpiry.IsZero() {
		return false
	}
	return time.Now().Before(a.tokenExpiry)
}

// base64URLEncode encodes data to base64 URL-safe format
func (a *AuthProvider) base64URLEncode(data interface{}) string {
	jsonBytes, _ := json.Marshal(data)
	return strings.TrimRight(base64.URLEncoding.EncodeToString(jsonBytes), "=")
}

// base64URLDecode decodes base64 URL-safe format
func (a *AuthProvider) base64URLDecode(data string) ([]byte, error) {
	// Add padding if needed
	for len(data)%4 != 0 {
		data += "="
	}
	return base64.URLEncoding.DecodeString(data)
}

// sign creates a HMAC-SHA256 signature
func (a *AuthProvider) sign(data string) string {
	h := hmac.New(sha256.New, []byte(a.jwtSecret))
	h.Write([]byte(data))
	return a.base64URLEncode(h.Sum(nil))
}
