package graphql

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// UsageEvents resolves a paginated connection of usage events.
func (r *Resolver) UsageEvents(ctx context.Context, first *int, after *string, imsi *string, filter *UsageEventFilter) (*UsageEventConnection, error) {
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

	where := buildUsageEventFilter(imsi, filter)
	events, total, err := r.charging.ListUsageEvents(ctx, limit, offset, where)
	if err != nil {
		return nil, err
	}

	edges := make([]*UsageEventEdge, len(events))
	for i, event := range events {
		edges[i] = &UsageEventEdge{
			Node:   event,
			Cursor: base64.StdEncoding.EncodeToString(fmt.Appendf(nil, "%d", offset+i+1)),
		}
	}

	return &UsageEventConnection{
		Edges:      edges,
		PageInfo:   buildConnectionPageInfo(offset, len(events), total),
		TotalCount: int(total),
	}, nil
}

// UsageStats resolves aggregate usage stats for a subscriber and period.
func (r *Resolver) UsageStats(ctx context.Context, input UsageStatsInput) (*UsageStats, error) {
	stats, err := r.charging.GetUsageStats(ctx, input.IMSI, string(input.Period))
	if err != nil {
		return nil, err
	}
	return &UsageStats{
		DataUsage:  stats.DataUsage,
		VoiceUsage: stats.VoiceUsage,
		SmsUsage:   stats.SmsUsage,
		Cost:       stats.Cost,
		Period:     stats.Period,
		Trend: &UsageTrend{
			Direction:        TrendDirection(stats.Trend.Direction),
			Percentage:       stats.Trend.Percentage,
			PeriodOverPeriod: stats.Trend.PeriodOverPeriod,
		},
	}, nil
}

// RealTimeUsage resolves real-time usage for a subscriber by IMSI.
func (r *Resolver) RealTimeUsage(ctx context.Context, imsi string) (*RealTimeUsage, error) {
	usage, err := r.charging.GetRealTimeUsage(ctx, imsi)
	if err != nil {
		return nil, err
	}

	var cs *CurrentSession
	if usage.CurrentSession != nil {
		cs = &CurrentSession{
			SessionID: usage.CurrentSession.SessionID,
			StartTime: usage.CurrentSession.StartTime,
			DataUsed:  usage.CurrentSession.DataUsed,
			VoiceUsed: usage.CurrentSession.VoiceUsed,
			SmsUsed:   usage.CurrentSession.SmsUsed,
			Cost:      usage.CurrentSession.Cost,
		}
	}

	return &RealTimeUsage{
		CurrentSession: cs,
		TodayUsage: &TodayUsage{
			DataUsed:  usage.TodayUsage.DataUsed,
			VoiceUsed: usage.TodayUsage.VoiceUsed,
			SmsUsed:   usage.TodayUsage.SmsUsed,
			Cost:      usage.TodayUsage.Cost,
		},
	}, nil
}

// SearchUsageEvents searches usage events, capped at limit.
func (r *Resolver) SearchUsageEvents(ctx context.Context, query string, limit int) ([]*models.UsageEvent, error) {
	return r.charging.SearchUsageEvents(ctx, query, limit)
}
