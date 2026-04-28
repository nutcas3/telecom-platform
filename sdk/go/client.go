package telecom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HTTPClient handles HTTP requests with authentication
type HTTPClient struct {
	client        *http.Client
	authProvider  *AuthProvider
	baseURL      string
	timeout       time.Duration
	maxRetries    int
	retryDelay    time.Duration
	enableLogging bool
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(baseURL string, authProvider *AuthProvider, timeout time.Duration, maxRetries int, retryDelay time.Duration, enableLogging bool) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		authProvider:  authProvider,
		baseURL:      baseURL,
		timeout:       timeout,
		maxRetries:    maxRetries,
		retryDelay:    retryDelay,
		enableLogging: enableLogging,
	}
}

// Get makes an HTTP GET request
func (h *HTTPClient) Get(ctx context.Context, path string, result interface{}, params ...map[string]string) error {
	return h.request(ctx, "GET", path, nil, result, params...)
}

// Post makes an HTTP POST request
func (h *HTTPClient) Post(ctx context.Context, path string, body, result interface{}) error {
	return h.request(ctx, "POST", path, body, result)
}

// Put makes an HTTP PUT request
func (h *HTTPClient) Put(ctx context.Context, path string, body, result interface{}) error {
	return h.request(ctx, "PUT", path, body, result)
}

// Delete makes an HTTP DELETE request
func (h *HTTPClient) Delete(ctx context.Context, path string) error {
	return h.request(ctx, "DELETE", path, nil, nil)
}

// request makes an HTTP request with retry logic
func (h *HTTPClient) request(ctx context.Context, method, path string, body, result interface{}, params ...map[string]string) error {
	var lastErr error

	for attempt := 0; attempt <= h.maxRetries; attempt++ {
		var bodyBytes []byte
		var err error

		if body != nil {
			bodyBytes, err = json.Marshal(body)
			if err != nil {
				return fmt.Errorf("failed to marshal body: %w", err)
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, h.baseURL+path, bytes.NewReader(bodyBytes))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		headers := h.authProvider.GetHeaders()
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		// Add query parameters
		if len(params) > 0 {
			q := req.URL.Query()
			for k, v := range params[0] {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()
		}

		resp, err := h.client.Do(req)
		if err != nil {
			lastErr = err
			if attempt == h.maxRetries {
				return fmt.Errorf("request failed after %d retries: %w", h.maxRetries, err)
			}
			time.Sleep(h.retryDelay * time.Duration(1<<uint(attempt)))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("authentication failed")
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			return fmt.Errorf("rate limit exceeded")
		}
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			var errorResp map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&errorResp)
			return fmt.Errorf("API error: %v", errorResp)
		}
		if resp.StatusCode >= 500 {
			return fmt.Errorf("server error: %d", resp.StatusCode)
		}

		if result != nil {
			if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
				return fmt.Errorf("failed to decode response: %w", err)
			}
		}

		return nil
	}

	return lastErr
}

// Close closes the HTTP client
func (h *HTTPClient) Close() error {
	return nil
}
