package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

var logger zerolog.Logger

type ES2Client struct {
	baseURL    string
	apiKey     string
	httpClient *resty.Client
}

type ProfileOrder struct {
	ICCID       string `json:"iccid"`
	IMSI        string `json:"imsi"`
	K           string `json:"k"`
	OPc         string `json:"opc"`
	MCC         string `json:"mcc"`
	MNC         string `json:"mnc"`
	ProfileType string `json:"profileType"`
}

type ProfileResponse struct {
	ActivationCode string `json:"activationCode"`
	ProfileID      string `json:"profileId"`
	Status         string `json:"status"`
}

func main() {
	// Initialize logger
	logger = zerolog.New(os.Stdout).With().
		Timestamp().
		Str("service", "carrier-connector").
		Logger()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logger.Info().Msg("No .env file found, using system environment")
	}

	logger.Info().Msg("Carrier Connector started")

	// Initialize ES2+ client
	smdpURL := getEnv("SMDP_URL", "https://smdp.example.com")
	apiKey := getEnv("SMDP_API_KEY", "test-api-key")

	client := NewES2Client(smdpURL, apiKey)

	// Example: Order a profile
	order := &ProfileOrder{
		ICCID:       "8933123456789012345",
		IMSI:        "208930000000001",
		K:           "465B5CE8B199B49FAA5F0A2EE238A6BC",
		OPc:         "E8ED289DEBA952E4283B54E88E6183CA",
		MCC:         "208",
		MNC:         "93",
		ProfileType: "operational",
	}

	logger.Info().
		Str("imsi", order.IMSI).
		Msg("Ordering eSIM profile from SM-DP+")

	// In production, this would be triggered by API requests
	response, err := client.OrderProfile(order)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to order profile")
	} else {
		logger.Info().
			Str("activation_code", response.ActivationCode).
			Str("profile_id", response.ProfileID).
			Msg("Profile ordered successfully")
	}

	// Keep service running
	logger.Info().Msg("Carrier Connector running. Press Ctrl+C to exit.")
	select {}
}

func NewES2Client(baseURL, apiKey string) *ES2Client {
	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", apiKey)).
		SetHeader("Content-Type", "application/json").
		SetTimeout(10 * time.Second).
		SetTLSClientConfig(&tls.Config{
			MinVersion: tls.VersionTLS12,
		})

	return &ES2Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: client,
	}
}

func (c *ES2Client) OrderProfile(order *ProfileOrder) (*ProfileResponse, error) {
	var response ProfileResponse

	// NOTE: This is a simplified example
	// In production, implement full GSMA ES2+ protocol
	resp, err := c.httpClient.R().
		SetBody(order).
		SetResult(&response).
		Post("/gsma/rsp2/es2plus/createProfile")

	if err != nil {
		return nil, fmt.Errorf("ES2+ API call failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("SM-DP+ returned error: %s", resp.String())
	}

	return &response, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
