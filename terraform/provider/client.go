package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// APIClient is the HTTP client for the Telecom API
type APIClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	apiURL := d.Get("api_url").(string)
	apiKey := d.Get("api_key").(string)

	client := &APIClient{
		BaseURL: apiURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	return client, nil
}

func (c *APIClient) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	return c.HTTPClient.Do(req)
}

func (c *APIClient) get(ctx context.Context, path string, result interface{}) error {
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d for GET %s", resp.StatusCode, path)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *APIClient) post(ctx context.Context, path string, body, result interface{}) error {
	resp, err := c.doRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("API returned status %d for POST %s", resp.StatusCode, path)
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (c *APIClient) put(ctx context.Context, path string, body, result interface{}) error {
	resp, err := c.doRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d for PUT %s", resp.StatusCode, path)
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (c *APIClient) delete(ctx context.Context, path string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("API returned status %d for DELETE %s", resp.StatusCode, path)
	}

	return nil
}
