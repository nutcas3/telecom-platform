package services

import (
	"context"
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
	"github.com/sirupsen/logrus"
)

// CreateSubscriber creates a new subscriber with allocated IMSI.
func (s *SubscriberService) CreateSubscriber(ctx context.Context, req *CreateSubscriberRequest) (*models.Subscriber, error) {
	imsi, err := s.allocateIMSI(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate IMSI: %w", err)
	}

	authKey, opc, err := s.generateAuthKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth keys: %w", err)
	}

	subscriber := &models.Subscriber{
		IMSI:           imsi,
		MSISDN:         req.MSISDN,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Email:          req.Email,
		OrganizationID: req.OrganizationID,
		Status:         models.SubscriberStatusActive,
		PlanID:         req.PlanID,
		AuthKey:        authKey,
		OPc:            opc,
		ServingPLMN:    models.PLMN{MCC: "208", MNC: "93"},
		ProfileStatus:  models.ProfileStatusInactive,
	}

	if err := s.db.CreateSubscriber(ctx, subscriber); err != nil {
		return nil, fmt.Errorf("failed to create subscriber: %w", err)
	}

	if req.EUICCID != "" {
		subscriber.EUICCID = req.EUICCID
		go func() {
			// Use derived context for goroutine to respect parent cancellation
			bgCtx := context.WithoutCancel(ctx)
			if err := s.provisionESIMProfile(bgCtx, subscriber.ID); err != nil {
				logrus.WithError(err).WithField("subscriber_id", subscriber.ID).Warn("Failed to provision eSIM profile")
			}
		}()
	} else {
		subscriber.Status = models.SubscriberStatusActive
		subscriber.ProfileStatus = models.ProfileStatusActive
		s.db.UpdateSubscriber(ctx, subscriber)
	}

	return subscriber, nil
}

// GetSubscriber retrieves a subscriber by ID.
func (s *SubscriberService) GetSubscriber(ctx context.Context, id uint) (*models.Subscriber, error) {
	subscriber, err := s.db.GetSubscriber(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber: %w", err)
	}
	return subscriber, nil
}

// GetSubscriberByIMSI retrieves a subscriber by IMSI.
func (s *SubscriberService) GetSubscriberByIMSI(ctx context.Context, imsi models.IMSI) (*models.Subscriber, error) {
	subscriber, err := s.db.GetSubscriberByIMSI(ctx, imsi)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber by IMSI: %w", err)
	}
	return subscriber, nil
}

// UpdateSubscriber updates subscriber information.
func (s *SubscriberService) UpdateSubscriber(ctx context.Context, id uint, req *UpdateSubscriberRequest) (*models.Subscriber, error) {
	subscriber, err := s.db.GetSubscriber(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber: %w", err)
	}

	if req.FirstName != nil && *req.FirstName != "" {
		subscriber.FirstName = *req.FirstName
	}
	if req.LastName != nil && *req.LastName != "" {
		subscriber.LastName = *req.LastName
	}
	if req.Email != nil && *req.Email != "" {
		subscriber.Email = *req.Email
	}
	if req.Status != "" {
		subscriber.Status = req.Status
	}
	if req.PlanID != nil {
		subscriber.PlanID = *req.PlanID
	}

	if err := s.db.UpdateSubscriber(ctx, subscriber); err != nil {
		return nil, fmt.Errorf("failed to update subscriber: %w", err)
	}

	return subscriber, nil
}

// ListSubscribers lists subscribers with cursor-based pagination and filtering.
func (s *SubscriberService) ListSubscribers(ctx context.Context, req *ListSubscribersRequest) (*ListSubscribersResponse, error) {
	// Set default limit if not provided
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	subscribers, nextCursor, hasMore, err := s.db.ListSubscribersCursor(ctx, req.Cursor, limit, string(req.Status), req.OrganizationID, req.Search)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscribers: %w", err)
	}

	return &ListSubscribersResponse{
		Subscribers: subscribers,
		NextCursor:  nextCursor,
		HasMore:     hasMore,
	}, nil
}
