package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// AMFClient represents the Access and Mobility Management Function client
type AMFClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAMFClient creates a new AMF client
func NewAMFClient(baseURL string) *AMFClient {
	return &AMFClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SessionTerminationRequest represents a session termination request to AMF
type SessionTerminationRequest struct {
	IMSI      models.IMSI `json:"imsi"`
	Reason    string      `json:"reason"`
	Timestamp time.Time   `json:"timestamp"`
}

// SessionTerminationResponse represents the response from AMF
type SessionTerminationResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
}

// TerminateSession terminates a subscriber session in the AMF
func (a *AMFClient) TerminateSession(ctx context.Context, imsi models.IMSI, reason string) error {
	req := SessionTerminationRequest{
		IMSI:      imsi,
		Reason:    reason,
		Timestamp: time.Now(),
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal session termination request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/sessions/terminate", a.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Request-ID", generateRequestID())

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send session termination request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("AMF returned status %d", resp.StatusCode)
	}

	var response SessionTerminationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode AMF response: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("AMF session termination failed: %s", response.Message)
	}

	return nil
}

// NotifySubscriberUpdate notifies AMF of subscriber status changes
func (a *AMFClient) NotifySubscriberUpdate(ctx context.Context, imsi models.IMSI, status models.SubscriberStatus) error {
	req := map[string]interface{}{
		"imsi":      imsi,
		"status":    status,
		"timestamp": time.Now(),
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal subscriber update request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/subscribers/update", a.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Request-ID", generateRequestID())

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send subscriber update request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("AMF returned status %d", resp.StatusCode)
	}

	return nil
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
