package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// downloadProfile is an internal method to download a profile
func (e *ES2Service) downloadProfile(ctx context.Context, req DownloadOrderRequest) (*DownloadOrderResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal download request: %w", err)
	}

	url := fmt.Sprintf("%s/es2/download", e.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Request-ID", generateRequestID())
	httpReq.Header.Set("Authorization", "Bearer "+e.config.APIKey)

	resp, err := e.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send download request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ES2 server returned status %d", resp.StatusCode)
	}

	var response DownloadOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode download response: %w", err)
	}

	return &response, nil
}

// ActivateProfile activates an eSIM profile
func (e *ES2Service) ActivateProfile(ctx context.Context, euiccID, profileID string) error {
	req := ActivationRequest{
		EID:      euiccID,
		ProfileID: profileID,
	}

	return e.sendActivationRequest(ctx, req, "activate")
}

// DeactivateProfile deactivates an eSIM profile
func (e *ES2Service) DeactivateProfile(ctx context.Context, euiccID, profileID string) error {
	req := DeactivationRequest{
		EID:      euiccID,
		ProfileID: profileID,
	}

	return e.sendDeactivationRequest(ctx, req)
}

// DeleteProfile deletes an eSIM profile
func (e *ES2Service) DeleteProfile(ctx context.Context, euiccID, profileID string) error {
	req := DeletionRequest{
		EID:      euiccID,
		ProfileID: profileID,
	}

	return e.sendDeletionRequest(ctx, req)
}
