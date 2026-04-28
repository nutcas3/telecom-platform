package telecom

import (
	"context"
	"fmt"
	"time"
)

// UsageAPI handles usage-related API calls
type UsageAPI struct {
	client *HTTPClient
}

// NewUsageAPI creates a new UsageAPI
func NewUsageAPI(client *HTTPClient) *UsageAPI {
	return &UsageAPI{client: client}
}

// GetStats retrieves usage statistics for a subscriber
func (u *UsageAPI) GetStats(ctx context.Context, subscriberID int64, startDate, endDate time.Time) (*UsageStats, error) {
	params := map[string]string{
		"start_date": startDate.Format(time.RFC3339),
		"end_date":   endDate.Format(time.RFC3339),
	}

	var stats UsageStats
	err := u.client.Get(ctx, fmt.Sprintf("/v1/subscribers/%d/usage", subscriberID), &stats, params)
	return &stats, err
}

// ListEvents retrieves a list of usage events
func (u *UsageAPI) ListEvents(ctx context.Context, subscriberID int64, usageType string, startDate, endDate time.Time, page, pageSize int32) (*UsageEventList, error) {
	params := map[string]string{
		"page":      fmt.Sprintf("%d", page),
		"page_size": fmt.Sprintf("%d", pageSize),
	}

	if subscriberID > 0 {
		params["subscriber_id"] = fmt.Sprintf("%d", subscriberID)
	}
	if usageType != "" {
		params["usage_type"] = usageType
	}
	if !startDate.IsZero() {
		params["start_date"] = startDate.Format(time.RFC3339)
	}
	if !endDate.IsZero() {
		params["end_date"] = endDate.Format(time.RFC3339)
	}

	var list UsageEventList
	err := u.client.Get(ctx, "/v1/usage/events", &list, params)
	return &list, err
}

// GetRealtime retrieves real-time usage for a subscriber
func (u *UsageAPI) GetRealtime(ctx context.Context, subscriberID int64) (*RealTimeUsage, error) {
	var usage RealTimeUsage
	err := u.client.Get(ctx, fmt.Sprintf("/v1/subscribers/%d/realtime", subscriberID), &usage)
	return &usage, err
}
