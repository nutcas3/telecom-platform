package es2

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/config"
)

type ES2Client struct {
	httpClient *http.Client
	config     *config.ES2Config
	baseURL    string
	maxRetries int
	retryDelay time.Duration
}

func NewES2Client(cfg *config.ES2Config) *ES2Client {
	return &ES2Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: cfg.InsecureSkipVerify,
				},
			},
		},
		config:     cfg,
		baseURL:    cfg.BaseURL,
		maxRetries: 3,
		retryDelay: 1 * time.Second,
	}
}

// retryableError checks if an error is retryable
func (c *ES2Client) retryableError(err error, statusCode int) bool {
	if err != nil {
		return true // Network errors are retryable
	}
	// Retry on 5xx server errors and 429 rate limit
	return statusCode >= 500 || statusCode == 429
}

// executeWithRetry executes an HTTP request with retry logic
func (c *ES2Client) executeWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	var lastErr error
	var lastResp *http.Response

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := c.retryDelay * time.Duration(math.Pow(2, float64(attempt-1)))
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if !c.retryableError(err, 0) {
				return nil, err
			}
			continue
		}

		if resp.StatusCode == http.StatusOK {
			return resp, nil
		}

		// Save response body for potential error reporting
		lastResp = resp

		if !c.retryableError(nil, resp.StatusCode) {
			return resp, nil
		}

		resp.Body.Close()
		lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	if lastResp != nil {
		return lastResp, lastErr
	}
	return nil, lastErr
}

func (c *ES2Client) DownloadProfile(ctx context.Context, req *DownloadProfileRequest) (*DownloadProfileResponse, error) {
	url := fmt.Sprintf("%s/es2plus/downloadProfile", c.baseURL)

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.executeWithRetry(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request after retries: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var downloadResp DownloadProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&downloadResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &downloadResp, nil
}

func (c *ES2Client) GetProfileStatus(ctx context.Context, req *GetProfileStatusRequest) (*GetProfileStatusResponse, error) {
	url := fmt.Sprintf("%s/es2plus/getProfileStatus", c.baseURL)

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.executeWithRetry(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request after retries: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var statusResp GetProfileStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &statusResp, nil
}

func (c *ES2Client) DeleteProfile(ctx context.Context, req *DeleteProfileRequest) (*DeleteProfileResponse, error) {
	url := fmt.Sprintf("%s/es2plus/deleteProfile", c.baseURL)

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.executeWithRetry(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request after retries: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var deleteResp DeleteProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&deleteResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &deleteResp, nil
}

func (c *ES2Client) EnableProfile(ctx context.Context, req *EnableProfileRequest) (*EnableProfileResponse, error) {
	url := fmt.Sprintf("%s/es2plus/enableProfile", c.baseURL)

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.executeWithRetry(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request after retries: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var enableResp EnableProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&enableResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &enableResp, nil
}

func (c *ES2Client) DisableProfile(ctx context.Context, req *DisableProfileRequest) (*DisableProfileResponse, error) {
	url := fmt.Sprintf("%s/es2plus/disableProfile", c.baseURL)

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.executeWithRetry(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request after retries: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var disableResp DisableProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&disableResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &disableResp, nil
}

func (c *ES2Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Telecom-Platform/1.0")

	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}

	if c.config.FunctionalityRequesterID != "" {
		req.Header.Set("X-Admin-Protocol", "gsma-rsp")
		req.Header.Set("X-Request-ID", generateRequestID())
	}
}

func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

type DownloadProfileRequest struct {
	EID              string `json:"eid"`
	ICCID            string `json:"iccid"`
	ProfileType      string `json:"profileType"`
	ConfirmationCode string `json:"confirmationCode,omitempty"`
}

type DownloadProfileResponse struct {
	ExecutionStatus string `json:"executionStatus"`
	StatusMessage   string `json:"statusMessage"`
}

type GetProfileStatusRequest struct {
	EID   string `json:"eid"`
	ICCID string `json:"iccid"`
}

type GetProfileStatusResponse struct {
	ExecutionStatus string `json:"executionStatus"`
	StatusMessage   string `json:"statusMessage"`
	ProfileState    string `json:"profileState,omitempty"`
}

type DeleteProfileRequest struct {
	EID   string `json:"eid"`
	ICCID string `json:"iccid"`
}

type DeleteProfileResponse struct {
	ExecutionStatus string `json:"executionStatus"`
	StatusMessage   string `json:"statusMessage"`
}

type EnableProfileRequest struct {
	EID   string `json:"eid"`
	ICCID string `json:"iccid"`
}

type EnableProfileResponse struct {
	ExecutionStatus string `json:"executionStatus"`
	StatusMessage   string `json:"statusMessage"`
}

type DisableProfileRequest struct {
	EID   string `json:"eid"`
	ICCID string `json:"iccid"`
}

type DisableProfileResponse struct {
	ExecutionStatus string `json:"executionStatus"`
	StatusMessage   string `json:"statusMessage"`
}
