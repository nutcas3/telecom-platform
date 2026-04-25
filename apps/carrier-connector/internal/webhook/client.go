package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookClient sends notifications to the api-server
type WebhookClient struct {
	httpClient *http.Client
	webhookURL string
	apiKey     string
}

// NewWebhookClient creates a new webhook client
func NewWebhookClient(webhookURL, apiKey string) *WebhookClient {
	return &WebhookClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		webhookURL: webhookURL,
		apiKey:     apiKey,
	}
}

// WebhookEvent represents a webhook notification event
type WebhookEvent struct {
	EventType    string         `json:"event_type"`
	Timestamp    time.Time      `json:"timestamp"`
	ProfileICCID string         `json:"profile_iccid"`
	Data         map[string]any `json:"data"`
}

// Event types
const (
	EventProfileDownloaded   = "profile.downloaded"
	EventProfileActivated    = "profile.activated"
	EventProfileDeactivated  = "profile.deactivated"
	EventProfileDeleted      = "profile.deleted"
	EventProfileStatusChange = "profile.status_changed"
	EventDownloadFailed      = "download.failed"
)

// SendNotification sends a webhook notification to the api-server
func (w *WebhookClient) SendNotification(ctx context.Context, event *WebhookEvent) error {
	if w.webhookURL == "" {
		return nil // Webhook disabled
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", w.webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if w.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+w.apiKey)
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// SendProfileDownloaded sends a notification when a profile is downloaded
func (w *WebhookClient) SendProfileDownloaded(ctx context.Context, iccid string, data map[string]any) error {
	event := &WebhookEvent{
		EventType:    EventProfileDownloaded,
		Timestamp:    time.Now(),
		ProfileICCID: iccid,
		Data:         data,
	}
	return w.SendNotification(ctx, event)
}

// SendProfileActivated sends a notification when a profile is activated
func (w *WebhookClient) SendProfileActivated(ctx context.Context, iccid string, data map[string]any) error {
	event := &WebhookEvent{
		EventType:    EventProfileActivated,
		Timestamp:    time.Now(),
		ProfileICCID: iccid,
		Data:         data,
	}
	return w.SendNotification(ctx, event)
}

// SendProfileDeleted sends a notification when a profile is deleted
func (w *WebhookClient) SendProfileDeleted(ctx context.Context, iccid string, data map[string]any) error {
	event := &WebhookEvent{
		EventType:    EventProfileDeleted,
		Timestamp:    time.Now(),
		ProfileICCID: iccid,
		Data:         data,
	}
	return w.SendNotification(ctx, event)
}

// SendDownloadFailed sends a notification when a profile download fails
func (w *WebhookClient) SendDownloadFailed(ctx context.Context, iccid string, errorMessage string) error {
	event := &WebhookEvent{
		EventType:    EventDownloadFailed,
		Timestamp:    time.Now(),
		ProfileICCID: iccid,
		Data: map[string]any{
			"error": errorMessage,
		},
	}
	return w.SendNotification(ctx, event)
}
