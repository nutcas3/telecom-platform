package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nutcas3/telecom-platform/apps/cli/internal/types"
)

// Client represents the API client for the telecom platform
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new API client from CLI config
func NewClient(cfg *types.CLIConfig) *Client {
	if cfg == nil {
		cfg = &types.CLIConfig{APIEndpoint: "http://localhost:8000"}
	}
	return &Client{
		baseURL: cfg.APIEndpoint,
		apiKey:  cfg.APIToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// BaseURL returns the configured API endpoint (useful for UI display).
func (c *Client) BaseURL() string {
	return c.baseURL
}

// doGetJSON performs a GET request and decodes JSON into out
func (c *Client) doGetJSON(path string, out any) error {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return err
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API %s %d: %s", path, resp.StatusCode, string(body))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// doPostJSON performs a POST request with JSON body and decodes response
func (c *Client) doPostJSON(path string, body any, out any) error {
	var buf *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}
	req, err := http.NewRequest("POST", c.baseURL+path, buf)
	if err != nil {
		return err
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		rb, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API %s %d: %s", path, resp.StatusCode, string(rb))
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// doDelete performs a DELETE request
func (c *Client) doDelete(path string) error {
	req, err := http.NewRequest("DELETE", c.baseURL+path, nil)
	if err != nil {
		return err
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API %s %d: %s", path, resp.StatusCode, string(body))
	}
	return nil
}

// SubscriberAccount holds account info
type SubscriberAccount struct {
	IMSI    string  `json:"imsi"`
	Name    string  `json:"name"`
	Status  string  `json:"status"`
	Balance float64 `json:"balance"`
}

// ListSubscribers retrieves subscribers
func (c *Client) ListSubscribers() ([]SubscriberAccount, error) {
	var resp struct {
		Data []SubscriberAccount `json:"data"`
	}
	if err := c.doGetJSON("/api/v1/subscribers", &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetSubscriber retrieves a single subscriber
func (c *Client) GetSubscriber(imsi string) (*SubscriberAccount, error) {
	var sub SubscriberAccount
	if err := c.doGetJSON("/api/v1/subscribers/"+imsi, &sub); err != nil {
		return nil, err
	}
	return &sub, nil
}

// CreateSubscriber creates a subscriber
func (c *Client) CreateSubscriber(imsi, name string) (*SubscriberAccount, error) {
	var sub SubscriberAccount
	body := map[string]string{"imsi": imsi, "name": name}
	if err := c.doPostJSON("/api/v1/subscribers", body, &sub); err != nil {
		return nil, err
	}
	return &sub, nil
}

// DeleteSubscriber deletes a subscriber
func (c *Client) DeleteSubscriber(imsi string) error {
	return c.doDelete("/api/v1/subscribers/" + imsi)
}

// Service describes a platform service
type Service struct {
	Name    string  `json:"name"`
	Status  string  `json:"status"`
	Version string  `json:"version"`
	Uptime  string  `json:"uptime"`
	CPU     float64 `json:"cpu"`
	Memory  string  `json:"memory"`
}

// PostRestart requests the API to restart the named service.
func (c *Client) PostRestart(name string) error {
	return c.doPostJSON("/api/v1/services/"+name+"/restart", nil, nil)
}

// ListServices returns platform services
func (c *Client) ListServices() ([]Service, error) {
	var resp struct {
		Data []Service `json:"data"`
	}
	if err := c.doGetJSON("/api/v1/services", &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// Alert represents a monitoring alert
type Alert struct {
	ID       string    `json:"id"`
	Severity string    `json:"severity"`
	Service  string    `json:"service"`
	Message  string    `json:"message"`
	Time     time.Time `json:"time"`
}

// ListAlerts returns alerts
func (c *Client) ListAlerts() ([]Alert, error) {
	var resp struct {
		Data []Alert `json:"data"`
	}
	if err := c.doGetJSON("/api/v1/alerts", &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// SystemStats represents system health/metrics
type SystemStats struct {
	ActiveSessions   int     `json:"active_sessions"`
	TotalAccounts    int     `json:"total_accounts"`
	BlockedUsers     int     `json:"blocked_users"`
	LowBalanceAlerts int     `json:"low_balance_alerts"`
	CPUUsage         float64 `json:"cpu_usage"`
	MemoryUsage      float64 `json:"memory_usage"`
	Uptime           string  `json:"uptime"`
}

// GetSystemStats retrieves system stats
func (c *Client) GetSystemStats() (*SystemStats, error) {
	var stats SystemStats
	if err := c.doGetJSON("/api/v1/system/stats", &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// HealthStatus represents a health check result
type HealthStatus struct {
	Service   string    `json:"service"`
	Status    string    `json:"status"`
	Uptime    string    `json:"uptime"`
	LastCheck time.Time `json:"last_check"`
}

// GetHealth returns per-service health
func (c *Client) GetHealth() ([]HealthStatus, error) {
	var resp struct {
		Data []HealthStatus `json:"data"`
	}
	if err := c.doGetJSON("/api/v1/health", &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// Invoice represents an invoice from the API
type Invoice struct {
	ID           string     `json:"id"`
	SubscriberID string     `json:"subscriber_id"`
	Amount       float64    `json:"amount"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	DueDate      time.Time  `json:"due_date"`
	PaidAt       *time.Time `json:"paid_at,omitempty"`
	Subscriber   Subscriber `json:"subscriber"`
}

// Payment represents a payment from the API
type Payment struct {
	ID          string     `json:"id"`
	InvoiceID   string     `json:"invoice_id"`
	Amount      float64    `json:"amount"`
	Method      string     `json:"method"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}

// Subscriber represents a subscriber from the API
type Subscriber struct {
	ID        uint   `json:"id"`
	IMSI      string `json:"imsi"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Status    string `json:"status"`
}

// GetInvoices retrieves invoices from the API
func (c *Client) GetInvoices() ([]Invoice, error) {
	url := fmt.Sprintf("%s/api/v1/invoices", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Data []Invoice `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Data, nil
}

// GetPayments retrieves payments from the API
func (c *Client) GetPayments() ([]Payment, error) {
	url := fmt.Sprintf("%s/api/v1/payments", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Data []Payment `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Data, nil
}

// GenerateInvoiceRequest represents a request to generate an invoice
type GenerateInvoiceRequest struct {
	SubscriberID string `json:"subscriber_id"`
}

// GenerateInvoiceResponse represents a response from generating an invoice
type GenerateInvoiceResponse struct {
	ID        string    `json:"id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	DueDate   time.Time `json:"due_date"`
}

// GenerateInvoice generates a new invoice for a subscriber
func (c *Client) GenerateInvoice(subscriberID string) (*GenerateInvoiceResponse, error) {
	url := fmt.Sprintf("%s/api/v1/invoices/generate", c.baseURL)

	request := GenerateInvoiceRequest{
		SubscriberID: subscriberID,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response GenerateInvoiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// IsConnected checks if the API server is reachable
func (c *Client) IsConnected() bool {
	url := fmt.Sprintf("%s/health", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
