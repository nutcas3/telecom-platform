package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/config"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/database"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/payment/gateways"
)

// SubscriberService handles subscriber management operations
type SubscriberService struct {
	db         *database.Database
	config     *config.Config
	amfClient  *AMFClient
	es2Service *ES2Service
	stripeGW   *gateways.StripeGateway
}

// NewSubscriberService creates a new subscriber service
func NewSubscriberService(db *database.Database, cfg *config.Config) *SubscriberService {
	return &SubscriberService{
		db:         db,
		config:     cfg,
		amfClient:  NewAMFClient("http://localhost:8081"), // Default AMF URL
		es2Service: NewES2Service(&cfg.ES2),
		stripeGW:   gateways.NewStripeGateway(cfg.Payment.StripeAPIKey, cfg.Payment.StripeWebhookSecret),
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

	// Use the subscriber's service plan for limits and get actual usage from database
	var dataUsed, voiceUsed, smsUsed float64

	// Get actual usage records from database
	if err := s.db.DB.WithContext(ctx).Model(&models.UsageEvent{}).
		Where("imsi = ? AND created_at >= ?", subscriber.IMSI, time.Now().AddDate(0, -1, 0)).
		Select("SUM(CASE WHEN usage_type = ? THEN volume ELSE 0 END) as data_used, SUM(CASE WHEN usage_type = ? THEN volume ELSE 0 END) as voice_used, SUM(CASE WHEN usage_type = ? THEN volume ELSE 0 END) as sms_used",
			models.UsageTypeData, models.UsageTypeVoice, models.UsageTypeSMS).
		Row().Scan(&dataUsed, &voiceUsed, &smsUsed); err != nil {
		// If no usage records found, start with zero
		dataUsed, voiceUsed, smsUsed = 0, 0, 0
	}

	// Get actual balance from billing system or database
	var balance float64
	if err := s.db.DB.WithContext(ctx).Model(&models.SubscriberAccount{}).
		Where("imsi = ?", subscriber.IMSI).
		Select("balance").Scan(&balance).Error; err != nil {
		// If no account exists, start with default balance
		balance = 0.0
	}

	account := &models.SubscriberAccount{
		IMSI:        string(subscriber.IMSI),
		Balance:     balance,
		DataLimit:   float64(subscriber.Plan.DataLimit),
		DataUsed:    dataUsed,
		VoiceLimit:  float64(subscriber.Plan.VoiceLimit),
		VoiceUsed:   voiceUsed,
		SMSLimit:    float64(subscriber.Plan.SMSLimit),
		SMSUsed:     smsUsed,
		Status:      string(subscriber.Status),
		LastUpdated: time.Now(),
	}

	return account, nil
}

// ListInvoices lists invoices for a subscriber
func (s *SubscriberService) ListInvoices(ctx context.Context, limit int, offset int, imsi *string, status *models.InvoiceStatus) ([]*models.Invoice, int64, error) {
	query := s.db.DB.WithContext(ctx).Model(&models.Invoice{})

	// Apply filters
	if imsi != nil {
		// Get subscriber ID from IMSI
		var subscriber models.Subscriber
		if err := s.db.DB.WithContext(ctx).Where("imsi = ?", *imsi).First(&subscriber).Error; err != nil {
			return nil, 0, fmt.Errorf("subscriber not found: %w", err)
		}
		query = query.Where("subscriber_id = ?", subscriber.ID)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get invoices with pagination
	var invoices []*models.Invoice
	if err := query.Preload("Subscriber").Limit(limit).Offset(offset).Order("created_at DESC").Find(&invoices).Error; err != nil {
		return nil, 0, err
	}

	return invoices, total, nil
}

// GetInvoice gets a specific invoice
func (s *SubscriberService) GetInvoice(ctx context.Context, id string) (*models.Invoice, error) {
	var invoice models.Invoice

	// Parse ID to integer
	invoiceID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice ID: %w", err)
	}

	// Get invoice from database with line items and subscriber
	if err := s.db.DB.WithContext(ctx).
		Preload("LineItems").
		Preload("Subscriber").
		First(&invoice, invoiceID).Error; err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}

	return &invoice, nil
}

// ListRatingPlans lists available rating plans
func (s *SubscriberService) ListRatingPlans(ctx context.Context) ([]*models.RatingPlan, error) {
	var plans []*models.RatingPlan

	// Get all active rating plans from database
	if err := s.db.DB.WithContext(ctx).
		Where("is_active = ?", true).
		Order("monthly_fee ASC").
		Find(&plans).Error; err != nil {
		return nil, fmt.Errorf("failed to list rating plans: %w", err)
	}

	return plans, nil
}

// GetRatingPlan gets a specific rating plan
func (s *SubscriberService) GetRatingPlan(ctx context.Context, planId string) (*models.RatingPlan, error) {
	var plan models.RatingPlan

	// Get rating plan from database
	if err := s.db.DB.WithContext(ctx).
		Where("plan_id = ? AND is_active = ?", planId, true).
		First(&plan).Error; err != nil {
		return nil, fmt.Errorf("rating plan not found: %w", err)
	}

	return &plan, nil
}

// ListAlerts lists alerts
func (s *SubscriberService) ListAlerts(ctx context.Context, limit int, offset int, subscriberId *int, severity *models.AlertSeverity, resolved *bool) ([]*models.Alert, int64, error) {
	query := s.db.DB.WithContext(ctx).Model(&models.Alert{})

	// Apply filters
	if subscriberId != nil {
		query = query.Where("subscriber_id = ?", *subscriberId)
	}

	if severity != nil {
		query = query.Where("severity = ?", *severity)
	}

	if resolved != nil {
		query = query.Where("resolved = ?", *resolved)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get alerts with pagination
	var alerts []*models.Alert
	if err := query.Preload("Subscriber").Limit(limit).Offset(offset).Order("timestamp DESC").Find(&alerts).Error; err != nil {
		return nil, 0, err
	}

	return alerts, total, nil
}

// SearchSubscribers searches subscribers
func (s *SubscriberService) SearchSubscribers(ctx context.Context, query string, limit int) ([]*models.Subscriber, error) {
	var subscribers []*models.Subscriber

	// Search subscribers by name, MSISDN, or IMSI
	searchPattern := "%" + query + "%"
	if err := s.db.DB.WithContext(ctx).
		Where("first_name ILIKE ? OR last_name ILIKE ? OR msisdn ILIKE ? OR imsi ILIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern).
		Limit(limit).
		Find(&subscribers).Error; err != nil {
		return nil, fmt.Errorf("failed to search subscribers: %w", err)
	}

	return subscribers, nil
}

// ResolveAlert resolves an alert
func (s *SubscriberService) ResolveAlert(ctx context.Context, alertId string) (*models.Alert, error) {
	var alert models.Alert

	// Parse ID to integer
	alertID, err := strconv.ParseUint(alertId, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid alert ID: %w", err)
	}

	// Update alert as resolved
	if err := s.db.DB.WithContext(ctx).
		Model(&alert).
		Where("id = ?", alertID).
		Update("resolved", true).
		Update("resolved_at", time.Now()).
		Error; err != nil {
		return nil, fmt.Errorf("failed to resolve alert: %w", err)
	}

	// Get updated alert
	if err := s.db.DB.WithContext(ctx).
		Preload("Subscriber").
		First(&alert, alertID).Error; err != nil {
		return nil, fmt.Errorf("alert not found: %w", err)
	}

	return &alert, nil
}

// CreateAlert creates an alert
func (s *SubscriberService) CreateAlert(ctx context.Context, req *models.CreateAlertRequest) (*models.Alert, error) {
	alert := &models.Alert{
		Type:         req.Type,
		Severity:     req.Severity,
		Message:      req.Message,
		SubscriberID: req.SubscriberID,
		Timestamp:    time.Now(),
		Resolved:     false,
	}

	// Create alert in database
	if err := s.db.DB.WithContext(ctx).Create(alert).Error; err != nil {
		return nil, fmt.Errorf("failed to create alert: %w", err)
	}

	// Get created alert with subscriber
	if err := s.db.DB.WithContext(ctx).
		Preload("Subscriber").
		First(alert, alert.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve created alert: %w", err)
	}

	return alert, nil
}

// SubscribeToSubscriberUpdates subscribes to subscriber updates
func (s *SubscriberService) SubscribeToSubscriberUpdates(ctx context.Context, subscriberId string) (<-chan *models.Subscriber, error) {
	ch := make(chan *models.Subscriber, 5) // Buffered channel to prevent blocking

	go func() {
		defer close(ch)

		// Poll for subscriber updates every 3 seconds
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		var lastUpdatedAt time.Time

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Query for subscriber updates since last check
				var subscriber models.Subscriber
				query := s.db.DB.WithContext(ctx).Where("imsi = ?", subscriberId)
				if !lastUpdatedAt.IsZero() {
					query = query.Where("updated_at > ?", lastUpdatedAt)
				}

				if err := query.First(&subscriber).Error; err != nil {
					// Subscriber not found or error, continue polling
					continue
				}

				// Send updated subscriber to channel
				select {
				case ch <- &subscriber:
					lastUpdatedAt = subscriber.UpdatedAt
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch, nil
}

// SubscribeToAlertUpdates subscribes to alert updates
func (s *SubscriberService) SubscribeToAlertUpdates(ctx context.Context, subscriberId *string) (<-chan *models.Alert, error) {
	ch := make(chan *models.Alert, 10) // Buffered channel to prevent blocking

	go func() {
		defer close(ch)

		// Poll for new alerts every 2 seconds
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		var lastAlertID uint = 0

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Query for new alerts since last check
				var alerts []models.Alert
				query := s.db.DB.WithContext(ctx).Where("id > ?", lastAlertID)

				// Filter by subscriber if specified
				if subscriberId != nil {
					// Convert subscriberId string to uint for database query
					if subID, err := strconv.ParseUint(*subscriberId, 10, 32); err == nil {
						query = query.Where("subscriber_id = ?", subID)
					}
				}

				if err := query.Order("id ASC").Limit(10).Find(&alerts).Error; err != nil {
					// Log error but continue polling
					continue
				}

				// Send new alerts to channel
				for _, alert := range alerts {
					select {
					case ch <- &alert:
						lastAlertID = alert.ID
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return ch, nil
}

// AddPaymentMethod adds a payment method
func (s *SubscriberService) AddPaymentMethod(ctx context.Context, subscriberId int, req *models.AddPaymentMethodRequest) (*models.PaymentMethod, error) {
	// Generate unique payment method ID
	paymentMethodID := fmt.Sprintf("pm_%d_%d", subscriberId, time.Now().Unix())

	// Process payment token using real Stripe API to extract card details
	last4, brand, expiryMonth, expiryYear, err := s.processPaymentToken(ctx, req.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to process payment token: %w", err)
	}

	paymentMethod := &models.PaymentMethod{
		ID:           paymentMethodID,
		SubscriberID: uint(subscriberId),
		GatewayID:    "default_gateway",
		Type:         req.Type,
		CustomerID:   fmt.Sprintf("cus_%d", subscriberId),
		Last4:        last4,
		Brand:        brand,
		ExpiryMonth:  expiryMonth,
		ExpiryYear:   expiryYear,
		IsDefault:    req.IsDefault,
		CreatedAt:    time.Now(),
	}

	// Create payment method in database
	if err := s.db.DB.WithContext(ctx).Create(paymentMethod).Error; err != nil {
		return nil, fmt.Errorf("failed to add payment method: %w", err)
	}

	return paymentMethod, nil
}

// processPaymentToken processes payment gateway token using real Stripe API
func (s *SubscriberService) processPaymentToken(ctx context.Context, token string) (last4, brand string, expiryMonth, expiryYear int, err error) {
	// Call Stripe API to retrieve payment method details from token
	details, err := s.stripeGW.RetrievePaymentMethodFromToken(ctx, token)
	if err != nil {
		return "", "", 0, 0, fmt.Errorf("failed to retrieve payment method from token: %w", err)
	}

	// Return the actual card details from Stripe
	return details.Last4, details.Brand, details.ExpiryMonth, details.ExpiryYear, nil
}

// RemovePaymentMethod removes a payment method
func (s *SubscriberService) RemovePaymentMethod(ctx context.Context, paymentMethodId string) (bool, error) {
	// Check if payment method exists
	var paymentMethod models.PaymentMethod
	if err := s.db.DB.WithContext(ctx).First(&paymentMethod, "id = ?", paymentMethodId).Error; err != nil {
		return false, fmt.Errorf("payment method not found: %w", err)
	}

	// Don't allow removal if it's the only payment method and has default status
	var count int64
	if err := s.db.DB.WithContext(ctx).
		Model(&models.PaymentMethod{}).
		Where("subscriber_id = ?", paymentMethod.SubscriberID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check payment methods: %w", err)
	}

	if count == 1 {
		return false, fmt.Errorf("cannot remove the only payment method")
	}

	// If this is the default payment method, set another one as default
	if paymentMethod.IsDefault {
		var newDefault models.PaymentMethod
		if err := s.db.DB.WithContext(ctx).
			Where("subscriber_id = ? AND id != ?", paymentMethod.SubscriberID, paymentMethodId).
			First(&newDefault).Error; err != nil {
			return false, fmt.Errorf("failed to find alternative payment method: %w", err)
		}

		if err := s.db.DB.WithContext(ctx).
			Model(&newDefault).
			Update("is_default", true).Error; err != nil {
			return false, fmt.Errorf("failed to set new default payment method: %w", err)
		}
	}

	// Delete the payment method
	if err := s.db.DB.WithContext(ctx).Delete(&paymentMethod).Error; err != nil {
		return false, fmt.Errorf("failed to delete payment method: %w", err)
	}

	return true, nil
}

// SetDefaultPaymentMethod sets default payment method
func (s *SubscriberService) SetDefaultPaymentMethod(ctx context.Context, paymentMethodId string) (*models.PaymentMethod, error) {
	var paymentMethod models.PaymentMethod

	// Start transaction
	tx := s.db.DB.WithContext(ctx).Begin()

	// Unset all existing default payment methods for this subscriber
	if err := tx.Model(&models.PaymentMethod{}).
		Where("id = ?", paymentMethodId).
		First(&paymentMethod).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("payment method not found: %w", err)
	}

	// Set all other payment methods as non-default
	if err := tx.Model(&models.PaymentMethod{}).
		Where("subscriber_id = ? AND id != ?", paymentMethod.SubscriberID, paymentMethodId).
		Update("is_default", false).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update other payment methods: %w", err)
	}

	// Set this payment method as default
	if err := tx.Model(&paymentMethod).
		Update("is_default", true).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to set default payment method: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &paymentMethod, nil
}

// TopUpBalance tops up subscriber balance
func (s *SubscriberService) TopUpBalance(ctx context.Context, imsi string, req *models.TopUpRequest) (*models.SubscriberAccount, error) {
	var account models.SubscriberAccount

	// Get subscriber account from database
	if err := s.db.DB.WithContext(ctx).
		Where("imsi = ?", imsi).
		First(&account).Error; err != nil {
		return nil, fmt.Errorf("subscriber account not found: %w", err)
	}

	// Update balance
	if err := s.db.DB.WithContext(ctx).
		Model(&account).
		Where("imsi = ?", imsi).
		Update("balance", account.Balance+req.Amount).
		Update("last_updated", time.Now()).Error; err != nil {
		return nil, fmt.Errorf("failed to top up balance: %w", err)
	}

	// Get updated account
	if err := s.db.DB.WithContext(ctx).
		Where("imsi = ?", imsi).
		First(&account).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve updated account: %w", err)
	}

	return &account, nil
}

// DeleteSubscriber deletes a subscriber
func (s *SubscriberService) DeleteSubscriber(ctx context.Context, subscriberId uint) (bool, error) {
	// Start transaction to delete all related data
	tx := s.db.DB.WithContext(ctx).Begin()

	// Delete subscriber's payment methods
	if err := tx.Where("subscriber_id = ?", subscriberId).Delete(&models.PaymentMethod{}).Error; err != nil {
		tx.Rollback()
		return false, fmt.Errorf("failed to delete payment methods: %w", err)
	}

	// Delete subscriber's alerts
	if err := tx.Where("subscriber_id = ?", subscriberId).Delete(&models.Alert{}).Error; err != nil {
		tx.Rollback()
		return false, fmt.Errorf("failed to delete alerts: %w", err)
	}

	// Delete subscriber's account
	if err := tx.Where("imsi IN (SELECT imsi FROM subscribers WHERE id = ?))", subscriberId).Delete(&models.SubscriberAccount{}).Error; err != nil {
		tx.Rollback()
		return false, fmt.Errorf("failed to delete account: %w", err)
	}

	// Delete subscriber
	if err := tx.Delete(&models.Subscriber{}, subscriberId).Error; err != nil {
		tx.Rollback()
		return false, fmt.Errorf("failed to delete subscriber: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return false, fmt.Errorf("failed to commit deletion: %w", err)
	}

	return true, nil
}

// ActivateSubscriber activates a subscriber
func (s *SubscriberService) ActivateSubscriber(ctx context.Context, subscriberId uint) (*models.Subscriber, error) {
	var subscriber models.Subscriber

	// Update subscriber status to active
	if err := s.db.DB.WithContext(ctx).
		Model(&subscriber).
		Where("id = ?", subscriberId).
		Update("status", models.SubscriberStatusActive).
		Update("activated_at", time.Now()).Error; err != nil {
		return nil, fmt.Errorf("failed to activate subscriber: %w", err)
	}

	// Get updated subscriber
	if err := s.db.DB.WithContext(ctx).
		Preload("Plan").
		First(&subscriber, subscriberId).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve activated subscriber: %w", err)
	}

	return &subscriber, nil
}
