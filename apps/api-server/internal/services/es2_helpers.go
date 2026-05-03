package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Helper methods for ES2 operations
func (e *ES2Service) sendActivationRequest(ctx context.Context, req ActivationRequest, action string) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal activation request: %w", err)
	}

	url := fmt.Sprintf("%s/es2/%s", e.config.BaseURL, action)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Request-ID", generateRequestID())
	httpReq.Header.Set("Authorization", "Bearer "+e.config.APIKey)

	resp, err := e.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send activation request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ES2 server returned status %d", resp.StatusCode)
	}

	return nil
}

func (e *ES2Service) sendDeactivationRequest(ctx context.Context, req any) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal deactivation request: %w", err)
	}

	url := fmt.Sprintf("%s/es2/deactivate", e.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Request-ID", generateRequestID())
	httpReq.Header.Set("Authorization", "Bearer "+e.config.APIKey)

	resp, err := e.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send deactivation request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ES2 server returned status %d", resp.StatusCode)
	}

	return nil
}

func (e *ES2Service) sendDeletionRequest(ctx context.Context, req DeletionRequest) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal deletion request: %w", err)
	}

	url := fmt.Sprintf("%s/es2/delete", e.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Request-ID", generateRequestID())
	httpReq.Header.Set("Authorization", "Bearer "+e.config.APIKey)

	resp, err := e.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send deletion request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ES2 server returned status %d", resp.StatusCode)
	}

	return nil
}

// ValidateEID validates an EID format
func (e *ES2Service) ValidateEID(eid string) error {
	if len(eid) != 32 {
		return fmt.Errorf("invalid EID length: expected 32 characters, got %d", len(eid))
	}

	// Additional validation logic can be added here
	return nil
}

// ValidateICCID validates an ICCID format
func (e *ES2Service) ValidateICCID(iccid string) error {
	if len(iccid) < 18 || len(iccid) > 22 {
		return fmt.Errorf("invalid ICCID length: expected 18-22 characters, got %d", len(iccid))
	}

	// Additional validation logic can be added here
	return nil
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
