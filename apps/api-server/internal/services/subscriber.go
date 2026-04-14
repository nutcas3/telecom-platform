package services

import (
	"context"
	"fmt"
	"time"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/config"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/database"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// SubscriberService handles subscriber management operations
type SubscriberService struct {
	db         *database.Database
	config     *config.Config
	amfClient  *AMFClient
	es2Service *ES2Service
}

// NewSubscriberService creates a new subscriber service
func NewSubscriberService(db *database.Database, cfg *config.Config) *SubscriberService {
	return &SubscriberService{
		db:         db,
		config:     cfg,
		amfClient:  NewAMFClient("http://localhost:8081"), // Default AMF URL
		es2Service: NewES2Service(&cfg.ES2),
	}
}

// CreateSubscriber creates a new subscriber with allocated IMSI
func (s *SubscriberService) CreateSubscriber(ctx context.Context, req *CreateSubscriberRequest) (*models.Subscriber, error) {
	// Allocate IMSI
	imsi, err := s.allocateIMSI(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate IMSI: %w", err)
	}

	// Generate authentication keys
	authKey, opc, err := s.generateAuthKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth keys: %w", err)
	}

	// Create subscriber
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
		ServingPLMN:    models.PLMN{MCC: "208", MNC: "93"}, // France
		ProfileStatus:  models.ProfileStatusInactive,
	}

	if err := s.db.CreateSubscriber(ctx, subscriber); err != nil {
		return nil, fmt.Errorf("failed to create subscriber: %w", err)
	}

	// Initiate eSIM profile provisioning if EUICCID provided
	if req.EUICCID != "" {
		subscriber.EUICCID = req.EUICCID
		go func() {
			ctx := context.Background()
			if err := s.provisionESIMProfile(ctx, subscriber.ID); err != nil {
				// Log error but don't fail subscriber creation
				fmt.Printf("Failed to provision eSIM profile for subscriber %d: %v\n", subscriber.ID, err)
			}
		}()
	} else {
		// Activate immediately for physical SIM
		subscriber.Status = models.SubscriberStatusActive
		subscriber.ProfileStatus = models.ProfileStatusActive
		s.db.UpdateSubscriber(ctx, subscriber)
	}

	return subscriber, nil
}

// GetSubscriber retrieves a subscriber by ID
func (s *SubscriberService) GetSubscriber(ctx context.Context, id uint) (*models.Subscriber, error) {
	subscriber, err := s.db.GetSubscriber(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber: %w", err)
	}
	return subscriber, nil
}

// GetSubscriberByIMSI retrieves a subscriber by IMSI
func (s *SubscriberService) GetSubscriberByIMSI(ctx context.Context, imsi models.IMSI) (*models.Subscriber, error) {
	subscriber, err := s.db.GetSubscriberByIMSI(ctx, imsi)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber by IMSI: %w", err)
	}
	return subscriber, nil
}

// UpdateSubscriber updates subscriber information
func (s *SubscriberService) UpdateSubscriber(ctx context.Context, id uint, req *UpdateSubscriberRequest) (*models.Subscriber, error) {
	subscriber, err := s.db.GetSubscriber(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber: %w", err)
	}

	// Update fields
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

// SuspendSubscriber suspends a subscriber
func (s *SubscriberService) SuspendSubscriber(ctx context.Context, id uint) error {
	subscriber, err := s.db.GetSubscriber(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get subscriber: %w", err)
	}

	subscriber.Status = models.SubscriberStatusSuspended

	// Terminate active sessions
	if err := s.terminateSubscriberSessions(ctx, subscriber.IMSI); err != nil {
		return fmt.Errorf("failed to terminate sessions: %w", err)
	}

	if err := s.db.UpdateSubscriber(ctx, subscriber); err != nil {
		return fmt.Errorf("failed to suspend subscriber: %w", err)
	}

	return nil
}

// TerminateSubscriber terminates a subscriber
func (s *SubscriberService) TerminateSubscriber(ctx context.Context, id uint) error {
	subscriber, err := s.db.GetSubscriber(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get subscriber: %w", err)
	}

	subscriber.Status = models.SubscriberStatusTerminated

	// Terminate active sessions
	if err := s.terminateSubscriberSessions(ctx, subscriber.IMSI); err != nil {
		return fmt.Errorf("failed to terminate sessions: %w", err)
	}

	// Deactivate eSIM profile if active
	if subscriber.EUICCID != "" && subscriber.ProfileStatus == models.ProfileStatusActive {
		if err := s.deactivateESIMProfile(ctx, subscriber.ID); err != nil {
			return fmt.Errorf("failed to deactivate eSIM profile: %w", err)
		}
	}

	if err := s.db.UpdateSubscriber(ctx, subscriber); err != nil {
		return fmt.Errorf("failed to terminate subscriber: %w", err)
	}

	return nil
}

// ListSubscribers lists subscribers with pagination and filtering
func (s *SubscriberService) ListSubscribers(ctx context.Context, req *ListSubscribersRequest) (*ListSubscribersResponse, error) {
	dbReq := &database.ListSubscribersRequest{
		Page:           req.Page,
		PageSize:       req.PageSize,
		Status:         req.Status,
		OrganizationID: req.OrganizationID,
		Search:         req.Search,
	}
	subscribers, total, err := s.db.ListSubscribers(ctx, dbReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscribers: %w", err)
	}

	return &ListSubscribersResponse{
		Subscribers: subscribers,
		Total:       total,
		Page:        req.Page,
		PageSize:    req.PageSize,
	}, nil
}

// terminateSubscriberSessions terminates all active sessions for a subscriber
func (s *SubscriberService) terminateSubscriberSessions(ctx context.Context, imsi models.IMSI) error {
	sessions, err := s.db.GetActiveSessionsByIMSI(ctx, imsi)
	if err != nil {
		return err
	}

	for _, session := range sessions {
		now := time.Now()
		session.Status = models.SessionStatusInactive
		session.EndTime = &now

		if err := s.db.UpdateSession(ctx, &session); err != nil {
			return err
		}

		// Notify AMF to terminate session
		if err := s.amfClient.TerminateSession(ctx, imsi, "Subscriber terminated"); err != nil {
			// Log error but continue with other sessions
			fmt.Printf("Failed to notify AMF for session termination: %v\n", err)
		}
	}

	return nil
}

// Request/Response types
type CreateSubscriberRequest struct {
	MSISDN         string `json:"msisdn" validate:"required"`
	FirstName      string `json:"first_name" validate:"required"`
	LastName       string `json:"last_name" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
	OrganizationID string `json:"organization_id"`
	PlanID         uint   `json:"plan_id" validate:"required"`
	EUICCID        string `json:"euicc_id"`
}

type UpdateSubscriberRequest struct {
	FirstName *string                 `json:"first_name"`
	LastName  *string                 `json:"last_name"`
	Email     *string                 `json:"email"`
	Status    models.SubscriberStatus `json:"status"`
	PlanID    *uint                   `json:"plan_id"`
}

type ListSubscribersRequest struct {
	Page           int                     `json:"page" query:"page"`
	PageSize       int                     `json:"page_size" query:"page_size"`
	Status         models.SubscriberStatus `json:"status" query:"status"`
	OrganizationID string                  `json:"organization_id" query:"organization_id"`
	Search         string                  `json:"search" query:"search"`
}

type ListSubscribersResponse struct {
	Subscribers []models.Subscriber `json:"subscribers"`
	Total       int64               `json:"total"`
	Page        int                 `json:"page"`
	PageSize    int                 `json:"page_size"`
}
