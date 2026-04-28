package graphql

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/services"
)

// Subscriber resolves a single subscriber by GraphQL ID.
func (r *Resolver) Subscriber(ctx context.Context, id string) (*models.Subscriber, error) {
	sid, err := parseID(id)
	if err != nil {
		return nil, err
	}
	return r.subscriber.GetSubscriber(ctx, sid)
}

// SubscriberByImsi resolves a subscriber by IMSI.
func (r *Resolver) SubscriberByImsi(ctx context.Context, imsi string) (*models.Subscriber, error) {
	return r.subscriber.GetSubscriberByIMSI(ctx, models.IMSI(imsi))
}

// Subscribers resolves a paginated connection of subscribers.
func (r *Resolver) Subscribers(ctx context.Context, first *int, after *string, filter *SubscriberFilter, sort *SubscriberSort) (*SubscriberConnection, error) {
	limit := 20
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

	resp, err := r.subscriber.ListSubscribers(ctx, &services.ListSubscribersRequest{
		Page:     offset/limit + 1,
		PageSize: limit,
	})
	if err != nil {
		return nil, err
	}

	edges := make([]*SubscriberEdge, len(resp.Subscribers))
	for i, sub := range resp.Subscribers {
		s := sub
		edges[i] = &SubscriberEdge{
			Node:   &s,
			Cursor: base64.StdEncoding.EncodeToString(fmt.Appendf(nil, "%d", offset+i+1)),
		}
	}

	return &SubscriberConnection{
		Edges:      edges,
		PageInfo:   buildConnectionPageInfo(offset, len(resp.Subscribers), resp.Total),
		TotalCount: int(resp.Total),
	}, nil
}

// SubscriberAccount resolves a subscriber's account by IMSI.
func (r *Resolver) SubscriberAccount(ctx context.Context, imsi string) (*models.SubscriberAccount, error) {
	return r.subscriber.GetAccount(ctx, imsi)
}

// SearchSubscribers performs a subscriber search, capped at the given limit.
func (r *Resolver) SearchSubscribers(ctx context.Context, query string, limit int) ([]*models.Subscriber, error) {
	return r.subscriber.SearchSubscribers(ctx, query, limit)
}
