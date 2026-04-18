package telecom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client represents the Telecom Platform SDK client
type Client struct {
	config     *Config
	httpClient *HTTPClient
}

// Config holds the SDK configuration
type Config struct {
	APIURL     string
	APIKey     string
	Timeout    time.Duration
	Retries    int
	EnableHTTP bool
}

// HTTPClient handles HTTP requests
type HTTPClient struct {
	client  *http.Client
	apiKey  string
	baseURL string
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		APIURL:     "http://localhost:8000",
		APIKey:     "",
		Timeout:    30 * time.Second,
		Retries:    3,
		EnableHTTP: true,
	}
}

// NewClient creates a new Telecom SDK client
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	client := &Client{
		config: config,
	}

	// Initialize HTTP client
	if config.EnableHTTP {
		client.httpClient = &HTTPClient{
			client: &http.Client{
				Timeout: config.Timeout,
			},
			apiKey:  config.APIKey,
			baseURL: config.APIURL,
		}
	}

	return client, nil
}

// Close closes the client connections
func (c *Client) Close() error {
	return nil
}

// HTTP Methods

// GetSubscriber retrieves a subscriber by ID
func (c *Client) GetSubscriber(ctx context.Context, id int64) (*Subscriber, error) {
	if !c.config.EnableHTTP {
		return nil, fmt.Errorf("HTTP client not enabled")
	}

	var subscriber Subscriber
	err := c.httpClient.Get(ctx, fmt.Sprintf("/v1/subscribers/%d", id), &subscriber)
	return &subscriber, err
}

// ListSubscribers retrieves a list of subscribers
func (c *Client) ListSubscribers(ctx context.Context, page, pageSize int32) (*SubscriberList, error) {
	if !c.config.EnableHTTP {
		return nil, fmt.Errorf("HTTP client not enabled")
	}

	params := map[string]string{
		"page":      fmt.Sprintf("%d", page),
		"page_size": fmt.Sprintf("%d", pageSize),
	}

	var list SubscriberList
	err := c.httpClient.Get(ctx, "/v1/subscribers", &list, params)
	return &list, err
}

// CreateSubscriber creates a new subscriber
func (c *Client) CreateSubscriber(ctx context.Context, req *CreateSubscriberRequest) (*Subscriber, error) {
	if !c.config.EnableHTTP {
		return nil, fmt.Errorf("HTTP client not enabled")
	}

	var subscriber Subscriber
	err := c.httpClient.Post(ctx, "/v1/subscribers", req, &subscriber)
	return &subscriber, err
}

// UpdateSubscriber updates an existing subscriber
func (c *Client) UpdateSubscriber(ctx context.Context, id int64, req *UpdateSubscriberRequest) (*Subscriber, error) {
	if !c.config.EnableHTTP {
		return nil, fmt.Errorf("HTTP client not enabled")
	}

	var subscriber Subscriber
	err := c.httpClient.Put(ctx, fmt.Sprintf("/v1/subscribers/%d", id), req, &subscriber)
	return &subscriber, err
}

// DeleteSubscriber deletes a subscriber
func (c *Client) DeleteSubscriber(ctx context.Context, id int64) error {
	if !c.config.EnableHTTP {
		return fmt.Errorf("HTTP client not enabled")
	}

	return c.httpClient.Delete(ctx, fmt.Sprintf("/v1/subscribers/%d", id))
}

// GetUsageStats retrieves usage statistics for a subscriber
func (c *Client) GetUsageStats(ctx context.Context, subscriberID int64, startDate, endDate time.Time) (*UsageStats, error) {
	if !c.config.EnableHTTP {
		return nil, fmt.Errorf("HTTP client not enabled")
	}

	params := map[string]string{
		"start_date": startDate.Format(time.RFC3339),
		"end_date":   endDate.Format(time.RFC3339),
	}

	var stats UsageStats
	err := c.httpClient.Get(ctx, fmt.Sprintf("/v1/subscribers/%d/usage", subscriberID), &stats, params)
	return &stats, err
}

// CreatePaymentTransaction creates a new payment transaction
func (c *Client) CreatePaymentTransaction(ctx context.Context, req *CreatePaymentRequest) (*PaymentTransaction, error) {
	if !c.config.EnableHTTP {
		return nil, fmt.Errorf("HTTP client not enabled")
	}

	var transaction PaymentTransaction
	err := c.httpClient.Post(ctx, "/v1/payments/transactions", req, &transaction)
	return &transaction, err
}

// GetSystemStats retrieves system statistics
func (c *Client) GetSystemStats(ctx context.Context) (*SystemStats, error) {
	if !c.config.EnableHTTP {
		return nil, fmt.Errorf("HTTP client not enabled")
	}

	var stats SystemStats
	err := c.httpClient.Get(ctx, "/v1/system/stats", &stats)
	return &stats, err
}

// HTTPClient implementation

// Get makes an HTTP GET request
func (h *HTTPClient) Get(ctx context.Context, path string, result interface{}, params ...map[string]string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", h.baseURL+path, nil)
	if err != nil {
		return err
	}

	if h.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.apiKey)
	}

	if len(params) > 0 {
		q := req.URL.Query()
		for k, v := range params[0] {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

// Post makes an HTTP POST request
func (h *HTTPClient) Post(ctx context.Context, path string, body, result interface{}) error {
	return h.doRequest(ctx, "POST", path, body, result)
}

// Put makes an HTTP PUT request
func (h *HTTPClient) Put(ctx context.Context, path string, body, result interface{}) error {
	return h.doRequest(ctx, "PUT", path, body, result)
}

// Delete makes an HTTP DELETE request
func (h *HTTPClient) Delete(ctx context.Context, path string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", h.baseURL+path, nil)
	if err != nil {
		return err
	}

	if h.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.apiKey)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	return nil
}

// doRequest makes an HTTP request with a body
func (h *HTTPClient) doRequest(ctx context.Context, method, path string, body, result interface{}) error {
	var bodyBytes []byte
	var err error

	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, h.baseURL+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}

	if h.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}
