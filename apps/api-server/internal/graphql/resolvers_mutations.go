package graphql

import (
	"context"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/services"
)

// CreateSubscriber creates a new subscriber from GraphQL input.
func (r *Resolver) CreateSubscriber(ctx context.Context, input CreateSubscriberInput) (*models.Subscriber, error) {
	req := &services.CreateSubscriberRequest{
		MSISDN:    input.MSISDN,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		PlanID:    uint(input.PlanID),
	}
	if input.OrganizationID != nil {
		req.OrganizationID = *input.OrganizationID
	}
	if input.EUICCID != nil {
		req.EUICCID = *input.EUICCID
	}
	return r.subscriber.CreateSubscriber(ctx, req)
}

// UpdateSubscriber updates subscriber fields.
func (r *Resolver) UpdateSubscriber(ctx context.Context, id string, input UpdateSubscriberInput) (*models.Subscriber, error) {
	sid, err := parseID(id)
	if err != nil {
		return nil, err
	}
	var planID *uint
	if input.PlanID != nil {
		p := uint(*input.PlanID)
		planID = &p
	}
	return r.subscriber.UpdateSubscriber(ctx, sid, &services.UpdateSubscriberRequest{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		PlanID:    planID,
	})
}

// DeleteSubscriber deletes a subscriber by GraphQL ID.
func (r *Resolver) DeleteSubscriber(ctx context.Context, id string) (bool, error) {
	sid, err := parseID(id)
	if err != nil {
		return false, err
	}
	return r.subscriber.DeleteSubscriber(ctx, sid)
}

// SuspendSubscriber suspends and returns the refreshed subscriber.
func (r *Resolver) SuspendSubscriber(ctx context.Context, id string) (*models.Subscriber, error) {
	sid, err := parseID(id)
	if err != nil {
		return nil, err
	}
	if err := r.subscriber.SuspendSubscriber(ctx, sid); err != nil {
		return nil, err
	}
	return r.subscriber.GetSubscriber(ctx, sid)
}

// ActivateSubscriber activates and returns the subscriber.
func (r *Resolver) ActivateSubscriber(ctx context.Context, id string) (*models.Subscriber, error) {
	sid, err := parseID(id)
	if err != nil {
		return nil, err
	}
	return r.subscriber.ActivateSubscriber(ctx, sid)
}

// TerminateSubscriber terminates and returns the refreshed subscriber.
func (r *Resolver) TerminateSubscriber(ctx context.Context, id string) (*models.Subscriber, error) {
	sid, err := parseID(id)
	if err != nil {
		return nil, err
	}
	if err := r.subscriber.TerminateSubscriber(ctx, sid); err != nil {
		return nil, err
	}
	return r.subscriber.GetSubscriber(ctx, sid)
}
