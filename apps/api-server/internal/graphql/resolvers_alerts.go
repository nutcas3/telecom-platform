package graphql

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// Alerts resolves a paginated connection of alerts.
func (r *Resolver) Alerts(ctx context.Context, first *int, after *string, subscriberId *int, severity *models.AlertSeverity, resolved *bool) (*AlertConnection, error) {
	limit := 50
	if first != nil {
		limit = *first
	}
	offset := 0
	if after != nil {
		cursor, err := base64.StdEncoding.DecodeString(*after)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor")
		}
		offset = parseCursorOffset(string(cursor))
	}

	alerts, total, err := r.subscriber.ListAlerts(ctx, limit, offset, subscriberId, severity, resolved)
	if err != nil {
		return nil, err
	}

	edges := make([]*AlertEdge, len(alerts))
	for i, alert := range alerts {
		edges[i] = &AlertEdge{
			Node:   alert,
			Cursor: base64.StdEncoding.EncodeToString(fmt.Appendf(nil, "%d", offset+i+1)),
		}
	}

	return &AlertConnection{
		Edges:      edges,
		PageInfo:   buildConnectionPageInfo(offset, len(alerts), total),
		TotalCount: int(total),
	}, nil
}

// SystemStats resolves system-wide stats.
func (r *Resolver) SystemStats(ctx context.Context) (*models.SystemStats, error) {
	return r.charging.GetSystemStats(ctx)
}

// HealthStatus resolves the platform health status.
func (r *Resolver) HealthStatus(ctx context.Context) (*models.HealthStatus, error) {
	return r.charging.GetHealthStatus(ctx)
}

// ResolveAlert marks an alert as resolved.
func (r *Resolver) ResolveAlert(ctx context.Context, alertId string) (*models.Alert, error) {
	return r.subscriber.ResolveAlert(ctx, alertId)
}

// CreateAlert creates a new alert from GraphQL input.
func (r *Resolver) CreateAlert(ctx context.Context, input CreateAlertInput) (*models.Alert, error) {
	return r.subscriber.CreateAlert(ctx, &models.CreateAlertRequest{
		Type:         models.AlertType(input.Type),
		Severity:     models.AlertSeverity(input.Severity),
		Message:      input.Message,
		SubscriberID: input.SubscriberID,
	})
}

// TriggerSystemMaintenance triggers a maintenance cycle.
func (r *Resolver) TriggerSystemMaintenance(ctx context.Context) (bool, error) {
	return r.charging.TriggerMaintenance(ctx)
}
