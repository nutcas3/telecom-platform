package graphql

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// Invoices resolves a paginated connection of invoices.
func (r *Resolver) Invoices(ctx context.Context, first *int, after *string, imsi *string, status *models.InvoiceStatus) (*InvoiceConnection, error) {
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

	invoices, total, err := r.subscriber.ListInvoices(ctx, limit, offset, imsi, status)
	if err != nil {
		return nil, err
	}

	edges := make([]*InvoiceEdge, len(invoices))
	for i, inv := range invoices {
		edges[i] = &InvoiceEdge{
			Node:   inv,
			Cursor: base64.StdEncoding.EncodeToString(fmt.Appendf(nil, "%d", offset+i+1)),
		}
	}

	return &InvoiceConnection{
		Edges:      edges,
		PageInfo:   buildConnectionPageInfo(offset, len(invoices), total),
		TotalCount: int(total),
	}, nil
}

// Invoice resolves a single invoice by ID.
func (r *Resolver) Invoice(ctx context.Context, id string) (*models.Invoice, error) {
	return r.subscriber.GetInvoice(ctx, id)
}

// RatingPlans resolves all available rating plans.
func (r *Resolver) RatingPlans(ctx context.Context) ([]*models.RatingPlan, error) {
	return r.subscriber.ListRatingPlans(ctx)
}

// RatingPlan resolves a single rating plan by plan_id.
func (r *Resolver) RatingPlan(ctx context.Context, planId string) (*models.RatingPlan, error) {
	return r.subscriber.GetRatingPlan(ctx, planId)
}

// TopUpBalance tops up a subscriber balance (mutation).
func (r *Resolver) TopUpBalance(ctx context.Context, imsi string, input TopUpInput) (*models.SubscriberAccount, error) {
	return r.subscriber.TopUpBalance(ctx, imsi, &models.TopUpRequest{
		Amount:          input.Amount,
		PaymentMethodID: input.PaymentMethodID,
	})
}
