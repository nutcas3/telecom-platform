package telecom

import (
	"time"
)

// Client represents the Telecom Platform SDK client
type Client struct {
	config       *Config
	authProvider *AuthProvider
	httpClient   *HTTPClient

	// API modules
	Subscribers *SubscriberAPI
	Usage       *UsageAPI
	Payments    *PaymentAPI
	RatingPlans *RatingPlanAPI
	System      *SystemAPI
	Analytics   *AnalyticsAPI
	Security    *SecurityAPI
	Currency    *CurrencyAPI
}

// Config holds the SDK configuration
type Config struct {
	APIURL        string
	APIKey        string
	JWTSecret     string
	Timeout       time.Duration
	MaxRetries    int
	RetryDelay    time.Duration
	EnableHTTP    bool
	EnableLogging bool
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		APIURL:        "http://localhost:8000",
		Timeout:       30 * time.Second,
		MaxRetries:    3,
		RetryDelay:    1 * time.Second,
		EnableHTTP:    true,
		EnableLogging: false,
	}
}

// NewClient creates a new Telecom SDK client
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	authProvider := NewAuthProvider(config.APIKey, config.JWTSecret)

	var httpClient *HTTPClient
	if config.EnableHTTP {
		httpClient = NewHTTPClient(
			config.APIURL,
			authProvider,
			config.Timeout,
			config.MaxRetries,
			config.RetryDelay,
			config.EnableLogging,
		)
	}

	client := &Client{
		config:       config,
		authProvider: authProvider,
		httpClient:   httpClient,
	}

	// Initialize API modules
	client.Subscribers = NewSubscriberAPI(httpClient)
	client.Usage = NewUsageAPI(httpClient)
	client.Payments = NewPaymentAPI(httpClient)
	client.RatingPlans = NewRatingPlanAPI(httpClient)
	client.System = NewSystemAPI(httpClient)
	client.Analytics = NewAnalyticsAPI(httpClient)
	client.Security = NewSecurityAPI(httpClient)
	client.Currency = NewCurrencyAPI(httpClient)

	return client, nil
}

// Close closes the client connections
func (c *Client) Close() error {
	if c.httpClient != nil {
		return c.httpClient.Close()
	}
	return nil
}

// Authentication methods

// GenerateJWTToken generates a JWT token for authentication
func (c *Client) GenerateJWTToken(userID string, expiryHours int, additionalClaims map[string]interface{}) (string, error) {
	return c.authProvider.GenerateJWTToken(userID, expiryHours, additionalClaims)
}

// ValidateJWTToken validates a JWT token
func (c *Client) ValidateJWTToken(token string) (map[string]interface{}, error) {
	return c.authProvider.ValidateJWTToken(token)
}
