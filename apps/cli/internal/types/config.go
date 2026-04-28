package types

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
)

type CLIConfig struct {
	APIEndpoint string
	APIToken    string
	Profile     string
	Verbose     bool
	NoColor     bool
	Theme       string
}

// Validate checks the configuration for errors and returns a list of validation errors
func (c *CLIConfig) Validate() []error {
	var errors []error

	// Validate API endpoint
	if c.APIEndpoint == "" {
		errors = append(errors, fmt.Errorf("API endpoint is required"))
	} else {
		if _, err := url.Parse(c.APIEndpoint); err != nil {
			errors = append(errors, fmt.Errorf("invalid API endpoint URL: %w", err))
		}
		if !strings.HasPrefix(c.APIEndpoint, "http://") && !strings.HasPrefix(c.APIEndpoint, "https://") {
			errors = append(errors, fmt.Errorf("API endpoint must start with http:// or https://"))
		}
	}

	// Validate API token (optional but warn if empty)
	if c.APIToken == "" {
		errors = append(errors, fmt.Errorf("API token is empty (some operations may fail)"))
	} else if len(c.APIToken) < 10 {
		errors = append(errors, fmt.Errorf("API token appears too short (expected at least 10 characters)"))
	}

	// Validate profile
	validProfiles := []string{"default", "dev", "staging", "prod"}
	if c.Profile != "" {
		valid := slices.Contains(validProfiles, c.Profile)
		if !valid {
			errors = append(errors, fmt.Errorf("invalid profile '%s', must be one of: %v", c.Profile, validProfiles))
		}
	}

	// Validate theme
	validThemes := []string{"default", "light", "dark", "high-contrast"}
	if c.Theme != "" {
		valid := slices.Contains(validThemes, c.Theme)
		if !valid {
			errors = append(errors, fmt.Errorf("invalid theme '%s', must be one of: %v", c.Theme, validThemes))
		}
	}

	return errors
}
