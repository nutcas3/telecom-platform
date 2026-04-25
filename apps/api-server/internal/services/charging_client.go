package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/config"
)

// ChargingEngineClient is an HTTP client for the Rust charging engine
type ChargingEngineClient struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

// NewChargingEngineClient creates a new client for the Rust charging engine
func NewChargingEngineClient(cfg *config.ChargingEngineConfig) *ChargingEngineClient {
	return &ChargingEngineClient{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// setAPIKeyHeader adds the X-API-Key header to a request if an API key is configured
func (c *ChargingEngineClient) setAPIKeyHeader(req *http.Request) {
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}
}

type CreditCheckRequest struct {
	BytesRequested uint64 `json:"bytes_requested"`
}

type CreditCheckResponse struct {
	Allowed        bool  `json:"allowed"`
	RemainingBytes int64 `json:"remaining_bytes"`
}

type DeductRequest struct {
	BytesUsed uint64 `json:"bytes_used"`
}

type AddCreditRequest struct {
	BytesToAdd uint64 `json:"bytes_to_add"`
}

type BalanceResponse struct {
	IP           string `json:"ip"`
	BalanceBytes int64  `json:"balance_bytes"`
}

type EngineHealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// CheckCredit checks if a subscriber (by IP) has enough credit
func (c *ChargingEngineClient) CheckCredit(ctx context.Context, ip string, bytesRequested uint64) (*CreditCheckResponse, error) {
	body, err := json.Marshal(CreditCheckRequest{BytesRequested: bytesRequested})
	if err != nil {
		return nil, fmt.Errorf("marshal credit check request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/credit/%s/check", c.baseURL, ip)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAPIKeyHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("charging engine request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("charging engine returned status %d", resp.StatusCode)
	}

	var result CreditCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}

// DeductCredit deducts bytes from a subscriber's credit balance
func (c *ChargingEngineClient) DeductCredit(ctx context.Context, ip string, bytesUsed uint64) error {
	body, err := json.Marshal(DeductRequest{BytesUsed: bytesUsed})
	if err != nil {
		return fmt.Errorf("marshal deduct request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/credit/%s/deduct", c.baseURL, ip)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAPIKeyHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("charging engine request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("charging engine returned status %d", resp.StatusCode)
	}
	return nil
}

// AddCredit adds bytes to a subscriber's credit balance
func (c *ChargingEngineClient) AddCredit(ctx context.Context, ip string, bytesToAdd uint64) error {
	body, err := json.Marshal(AddCreditRequest{BytesToAdd: bytesToAdd})
	if err != nil {
		return fmt.Errorf("marshal add credit request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/credit/%s/add", c.baseURL, ip)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAPIKeyHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("charging engine request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("charging engine returned status %d", resp.StatusCode)
	}
	return nil
}

// GetBalance gets the current credit balance for a subscriber
func (c *ChargingEngineClient) GetBalance(ctx context.Context, ip string) (*BalanceResponse, error) {
	url := fmt.Sprintf("%s/v1/credit/%s/balance", c.baseURL, ip)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setAPIKeyHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("charging engine request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("charging engine returned status %d", resp.StatusCode)
	}

	var result BalanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}

// HealthCheck checks if the Rust charging engine is healthy
func (c *ChargingEngineClient) HealthCheck(ctx context.Context) (*EngineHealthResponse, error) {
	url := fmt.Sprintf("%s/health", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setAPIKeyHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("charging engine health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("charging engine returned status %d", resp.StatusCode)
	}

	var result EngineHealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}
