package es2

import (
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
