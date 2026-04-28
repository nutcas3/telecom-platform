package handlers

import (
	"os"

	"github.com/rs/zerolog"
)

// Logger is the package-level logger, initialized by main.
var Logger zerolog.Logger

// ProfileOrder is the GSMA ES2+ profile order request accepted by the API.
type ProfileOrder struct {
	EID              string `json:"eid"`
	ICCID            string `json:"iccid"`
	IMSI             string `json:"imsi"`
	K                string `json:"k"`
	OPc              string `json:"opc"`
	MCC              string `json:"mcc"`
	MNC              string `json:"mnc"`
	ProfileType      string `json:"profileType"`
	ConfirmationCode string `json:"confirmationCode,omitempty"`
}

// ProfileResponse is the GSMA ES2+ profile order response returned by the API.
type ProfileResponse struct {
	ExecutionStatus string `json:"executionStatus"`
	StatusMessage   string `json:"statusMessage"`
	ProfileID       string `json:"profileId"`
	ActivationCode  string `json:"activationCode,omitempty"`
}

// GetEnv returns the value of an environment variable or fallback if unset.
func GetEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
