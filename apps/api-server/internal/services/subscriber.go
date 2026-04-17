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

// GetAccount gets subscriber account information
func (s *SubscriberService) GetAccount(ctx context.Context, imsi string) (*models.SubscriberAccount, error) {
	subscriber, err := s.db.GetSubscriberByIMSI(ctx, models.IMSI(imsi))
	if err != nil {
		return nil, err
	}

	// Use the subscriber's service plan for limits
	account := &models.SubscriberAccount{
		IMSI:        string(subscriber.IMSI),
		Balance:     100.0, // Placeholder - would get from billing system
		DataLimit:   float64(subscriber.Plan.DataLimit),
		DataUsed:    0.0, // Placeholder - would get from usage records
		VoiceLimit:  float64(subscriber.Plan.VoiceLimit),
		VoiceUsed:   0.0, // Placeholder - would get from usage records
		SMSLimit:    float64(subscriber.Plan.SMSLimit),
		SMSUsed:     0.0, // Placeholder - would get from usage records
		Status:      string(subscriber.Status),
		LastUpdated: time.Now(),
	}

	return account, nil
}

// ListInvoices lists invoices for a subscriber
func (s *SubscriberService) ListInvoices(ctx context.Context, limit int, offset int, imsi *string, status *models.InvoiceStatus) ([]*models.Invoice, int64, error) {
	// This would get actual invoices in a real implementation
	// For now, return placeholder data
	invoices := []*models.Invoice{
		{
			ID:           1,
			SubscriberID: 1,
			Amount:       25.50,
			Currency:     "USD",
			Status:       *status,
			DueDate:      time.Now().Add(30 * 24 * time.Hour),
			CreatedAt:    time.Now(),
			LineItems: []models.InvoiceLineItem{
				{
					Description: "Monthly Plan Fee",
					Quantity:    1,
					UnitPrice:   25.00,
					Amount:      25.00,
				},
				{
					Description: "Data Overage",
					Quantity:    1,
					UnitPrice:   0.50,
					Amount:      0.50,
				},
			},
		},
	}
	return invoices, int64(len(invoices)), nil
}

// GetInvoice gets a specific invoice
func (s *SubscriberService) GetInvoice(ctx context.Context, id string) (*models.Invoice, error) {
	// This would get actual invoice in a real implementation
	// For now, return placeholder data
	invoice := &models.Invoice{
		ID:           1,
		SubscriberID: 1,
		Amount:       25.50,
		Currency:     "USD",
		Status:       models.InvoiceStatusPaid,
		DueDate:      time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:    time.Now(),
		LineItems: []models.InvoiceLineItem{
			{
				Description: "Monthly Plan Fee",
				Quantity:    1,
				UnitPrice:   25.00,
				Amount:      25.00,
			},
		},
	}
	return invoice, nil
}

// ListRatingPlans lists available rating plans
func (s *SubscriberService) ListRatingPlans(ctx context.Context) ([]*models.RatingPlan, error) {
	// This would get actual rating plans in a real implementation
	// For now, return placeholder data
	plans := []*models.RatingPlan{
		{
			PlanID:     "basic",
			Name:       "Basic Plan",
			DataRate:   0.01,
			VoiceRate:  0.05,
			SMSRate:    0.10,
			MonthlyFee: 10.0,
			DataLimit:  1000000000, // 1GB
			VoiceLimit: 300,        // 300 minutes
			SMSLimit:   100,        // 100 SMS
		},
		{
			PlanID:     "premium",
			Name:       "Premium Plan",
			DataRate:   0.005,
			VoiceRate:  0.02,
			SMSRate:    0.05,
			MonthlyFee: 25.0,
			DataLimit:  5000000000, // 5GB
			VoiceLimit: 1000,       // 1000 minutes
			SMSLimit:   500,        // 500 SMS
		},
	}
	return plans, nil
}

// GetRatingPlan gets a specific rating plan
func (s *SubscriberService) GetRatingPlan(ctx context.Context, planId string) (*models.RatingPlan, error) {
	// This would get actual rating plan in a real implementation
	// For now, return placeholder data
	plan := &models.RatingPlan{
		PlanID:     planId,
		Name:       "Basic Plan",
		DataRate:   0.01,
		VoiceRate:  0.05,
		SMSRate:    0.10,
		MonthlyFee: 10.0,
		DataLimit:  1000000000, // 1GB
		VoiceLimit: 300,        // 300 minutes
		SMSLimit:   100,        // 100 SMS
	}
	return plan, nil
}

// ListAlerts lists alerts
func (s *SubscriberService) ListAlerts(ctx context.Context, limit int, offset int, subscriberId *int, severity *models.AlertSeverity, resolved *bool) ([]*models.Alert, int64, error) {
	// This would get actual alerts in a real implementation
	// For now, return placeholder data
	alerts := []*models.Alert{
		{
			ID:           1,
			Type:         models.AlertTypeLowBalance,
			Severity:     *severity,
			Message:      "Low balance warning",
			SubscriberID: subscriberId,
			Timestamp:    time.Now(),
			Resolved:     *resolved,
		},
	}
	return alerts, int64(len(alerts)), nil
}

// SearchSubscribers searches subscribers
func (s *SubscriberService) SearchSubscribers(ctx context.Context, query string, limit int) ([]*models.Subscriber, error) {
	// This would search actual subscribers in a real implementation
	// For now, return placeholder data
	subscribers := []*models.Subscriber{
		{
			ID:        1,
			IMSI:      "123456789012345",
			MSISDN:    "1234567890",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
			Status:    models.SubscriberStatusActive,
			PlanID:    1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	return subscribers, nil
}

// ResolveAlert resolves an alert
func (s *SubscriberService) ResolveAlert(ctx context.Context, alertId string) (*models.Alert, error) {
	// This would resolve actual alert in a real implementation
	// For now, return placeholder data
	alert := &models.Alert{
		ID:        1,
		Type:      models.AlertTypeLowBalance,
		Severity:  models.AlertSeverityLow,
		Message:   "Low balance warning",
		Timestamp: time.Now(),
		Resolved:  true,
	}
	return alert, nil
}

// CreateAlert creates an alert
func (s *SubscriberService) CreateAlert(ctx context.Context, req *models.CreateAlertRequest) (*models.Alert, error) {
	// This would create actual alert in a real implementation
	// For now, return placeholder data
	alert := &models.Alert{
		ID:           1,
		Type:         req.Type,
		Severity:     req.Severity,
		Message:      req.Message,
		SubscriberID: req.SubscriberID,
		Timestamp:    time.Now(),
		Resolved:     false,
	}
	return alert, nil
}

// SubscribeToSubscriberUpdates subscribes to subscriber updates
func (s *SubscriberService) SubscribeToSubscriberUpdates(ctx context.Context, subscriberId string) (<-chan *models.Subscriber, error) {
	// This would set up actual subscription in a real implementation
	// For now, return a closed channel
	ch := make(chan *models.Subscriber)
	close(ch)
	return ch, nil
}

// SubscribeToAlertUpdates subscribes to alert updates
func (s *SubscriberService) SubscribeToAlertUpdates(ctx context.Context, severity models.AlertSeverity) (<-chan *models.Alert, error) {
	// This would set up actual subscription in a real implementation
	// For now, return a closed channel
	ch := make(chan *models.Alert)
	close(ch)
	return ch, nil
}

// AddPaymentMethod adds a payment method
func (s *SubscriberService) AddPaymentMethod(ctx context.Context, subscriberId int, req *models.AddPaymentMethodRequest) (*models.PaymentMethod, error) {
	// This would add actual payment method in a real implementation
	// For now, return placeholder data
	paymentMethod := &models.PaymentMethod{
		ID:          "pm_123",
		Type:        req.Type,
		CustomerID:  "customer_123",
		Last4:       "1234",
		Brand:       "visa",
		ExpiryMonth: 12,
		ExpiryYear:  2025,
		IsDefault:   req.IsDefault,
		CreatedAt:   time.Now(),
	}
	return paymentMethod, nil
}

// RemovePaymentMethod removes a payment method
func (s *SubscriberService) RemovePaymentMethod(ctx context.Context, paymentMethodId string) (bool, error) {
	// This would remove actual payment method in a real implementation
	// For now, return success
	return true, nil
}

// SetDefaultPaymentMethod sets default payment method
func (s *SubscriberService) SetDefaultPaymentMethod(ctx context.Context, paymentMethodId string) (*models.PaymentMethod, error) {
	// This would set actual default payment method in a real implementation
	// For now, return placeholder data
	paymentMethod := &models.PaymentMethod{
		ID:          paymentMethodId,
		Type:        models.PaymentMethodTypeCreditCard,
		CustomerID:  "customer_123",
		Last4:       "1234",
		Brand:       "visa",
		ExpiryMonth: 12,
		ExpiryYear:  2025,
		IsDefault:   true,
		CreatedAt:   time.Now(),
	}
	return paymentMethod, nil
}

// TopUpBalance tops up subscriber balance
func (s *SubscriberService) TopUpBalance(ctx context.Context, imsi string, req *models.TopUpRequest) (*models.SubscriberAccount, error) {
	// This would top up actual balance in a real implementation
	// For now, return placeholder data
	account := &models.SubscriberAccount{
		IMSI:        imsi,
		Balance:     150.0,      // Previous balance + top-up
		DataLimit:   1000000000, // 1GB
		DataUsed:    500000000,  // 500MB
		VoiceLimit:  300,
		VoiceUsed:   45,
		SMSLimit:    100,
		SMSUsed:     8,
		Status:      "active",
		LastUpdated: time.Now(),
	}
	return account, nil
}

// DeleteSubscriber deletes a subscriber
func (s *SubscriberService) DeleteSubscriber(ctx context.Context, subscriberId uint) (bool, error) {
	// This would delete actual subscriber in a real implementation
	// For now, return success
	return true, nil
}

// ActivateSubscriber activates a subscriber
func (s *SubscriberService) ActivateSubscriber(ctx context.Context, subscriberId uint) (*models.Subscriber, error) {
	// This would activate actual subscriber in a real implementation
	// For now, return placeholder data
	subscriber := &models.Subscriber{
		ID:        subscriberId,
		IMSI:      "123456789012345",
		MSISDN:    "1234567890",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Status:    models.SubscriberStatusActive,
		PlanID:    1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return subscriber, nil
}
