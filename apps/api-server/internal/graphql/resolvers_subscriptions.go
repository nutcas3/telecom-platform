package graphql

import (
	"context"
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// SubscriberUpdates streams subscriber updates via a channel.
func (r *Resolver) SubscriberUpdates(ctx context.Context, subscriberId string) (<-chan *models.Subscriber, error) {
	return r.subscriber.SubscribeToSubscriberUpdates(ctx, subscriberId)
}

// UsageUpdates streams usage-event updates for an IMSI.
func (r *Resolver) UsageUpdates(ctx context.Context, imsi string) (<-chan *models.UsageEvent, error) {
	return r.charging.SubscribeToUsageUpdates(ctx, imsi)
}

// AlertUpdates streams alert updates filtered by severity.
func (r *Resolver) AlertUpdates(ctx context.Context, severity models.AlertSeverity) (<-chan *models.Alert, error) {
	severityStr := string(severity)
	return r.subscriber.SubscribeToAlertUpdates(ctx, &severityStr)
}

// SystemStatsUpdates streams system stats updates.
func (r *Resolver) SystemStatsUpdates(ctx context.Context) (<-chan *models.SystemStats, error) {
	return r.charging.SubscribeToSystemStatsUpdates(ctx)
}

// ChargingSessions resolves a paginated connection of charging sessions.
// Placeholder implementation until the charging query surface is finalized.
func (r *Resolver) ChargingSessions(ctx context.Context, first *int, after *string, imsi *string, status *SessionStatus) (*ChargingSessionConnection, error) {
	return &ChargingSessionConnection{
		Edges:      []*ChargingSessionEdge{},
		PageInfo:   &PageInfo{},
		TotalCount: 0,
	}, nil
}

// ChargingSession resolves a single charging session by ID.
// Placeholder implementation until the charging query surface is finalized.
func (r *Resolver) ChargingSession(ctx context.Context, sessionId string) (*ChargingSession, error) {
	return nil, fmt.Errorf("charging session %s not found", sessionId)
}
