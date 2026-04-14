package services

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/config"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// ES2Service handles GSMA ES2+ operations for eSIM provisioning
type ES2Service struct {
	httpClient *http.Client
	config     *config.ES2Config
}

// NewES2Service creates a new ES2 service
func NewES2Service(cfg *config.ES2Config) *ES2Service {
	return &ES2Service{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: cfg.InsecureSkipVerify,
				},
			},
		},
		config: cfg,
	}
}

// ProvisionProfile provisions an eSIM profile for a subscriber
func (e *ES2Service) ProvisionProfile(ctx context.Context, subscriber *models.Subscriber) (*ProfileInfo, error) {
	// Check if subscriber has EUICCID
	if subscriber.EUICCID == "" {
		return nil, fmt.Errorf("subscriber must have EUICCID for eSIM provisioning")
	}

	// Create profile download request
	req := DownloadOrderRequest{
		EID:          subscriber.EUICCID,
		ICCID:        "", // Will be assigned by SM-SR
		ProfileType:  "operational",
		Confirmation: true,
		Metadata: map[string]string{
			"subscriber_id": fmt.Sprintf("%d", subscriber.ID),
			"imsi":          string(subscriber.IMSI),
			"organization":  subscriber.OrganizationID,
		},
	}

	// Call SM-SR to download profile
	resp, err := e.downloadProfile(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to download profile: %w", err)
	}

	// Create profile info
	profileInfo := &ProfileInfo{
		ICCID:       resp.ICCID,
		ProfileID:   resp.ProfileID,
		ProfileName: fmt.Sprintf("Profile-%s", subscriber.IMSI),
		State:       "downloaded",
		Operator:    "Telecom Platform",
		Activation: ProfileActivation{
			ActivationCode: resp.ActivationCode,
			ConfAddress:    resp.ConfirmationAddress,
		},
	}

	return profileInfo, nil
}
