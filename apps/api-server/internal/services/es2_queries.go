package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetProfileStatus retrieves the status of an eSIM profile
func (e *ES2Service) GetProfileStatus(ctx context.Context, euiccID, profileID string) (*ProfileInfo, error) {
	url := fmt.Sprintf("%s/es2/eid/%s/profile/%s/status", e.config.BaseURL, euiccID, profileID)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("X-Request-ID", generateRequestID())
	httpReq.Header.Set("Authorization", "Bearer "+e.config.APIKey)

	resp, err := e.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send status request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ES2 server returned status %d", resp.StatusCode)
	}

	var response ProfileStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode status response: %w", err)
	}

	profileInfo := &ProfileInfo{
		ICCID:       response.ICCID,
		ProfileID:   response.ProfileID,
		ProfileName: response.ProfileName,
		State:       response.State,
		Operator:    response.Operator,
	}

	return profileInfo, nil
}

// ListProfiles lists all profiles on an eUICC
func (e *ES2Service) ListProfiles(ctx context.Context, euiccID string) ([]*ProfileInfo, error) {
	url := fmt.Sprintf("%s/es2/eid/%s/profiles", e.config.BaseURL, euiccID)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("X-Request-ID", generateRequestID())
	httpReq.Header.Set("Authorization", "Bearer "+e.config.APIKey)

	resp, err := e.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send list request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ES2 server returned status %d", resp.StatusCode)
	}

	var response ListProfilesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode list response: %w", err)
	}

	var profiles []*ProfileInfo
	for _, profile := range response.Profiles {
		profileInfo := &ProfileInfo{
			ICCID:       profile.ICCID,
			ProfileID:   profile.ProfileID,
			ProfileName: profile.ProfileName,
			State:       profile.State,
			Operator:    profile.Operator,
		}
		profiles = append(profiles, profileInfo)
	}

	return profiles, nil
}
