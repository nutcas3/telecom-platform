package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// GetAccount returns subscriber account with balance and usage aggregates.
func (s *SubscriberService) GetAccount(ctx context.Context, imsi string) (*models.SubscriberAccount, error) {
	subscriber, err := s.db.GetSubscriberByIMSI(ctx, models.IMSI(imsi))
	if err != nil {
		return nil, err
	}

	var dataUsed, voiceUsed, smsUsed float64
	if err := s.db.DB.WithContext(ctx).Model(&models.UsageEvent{}).
		Where("imsi = ? AND created_at >= ?", subscriber.IMSI, time.Now().AddDate(0, -1, 0)).
		Select("SUM(CASE WHEN usage_type = ? THEN volume ELSE 0 END) as data_used, SUM(CASE WHEN usage_type = ? THEN volume ELSE 0 END) as voice_used, SUM(CASE WHEN usage_type = ? THEN volume ELSE 0 END) as sms_used",
			models.UsageTypeData, models.UsageTypeVoice, models.UsageTypeSMS).
		Row().Scan(&dataUsed, &voiceUsed, &smsUsed); err != nil {
		dataUsed, voiceUsed, smsUsed = 0, 0, 0
	}

	var balance float64
	if err := s.db.DB.WithContext(ctx).Model(&models.SubscriberAccount{}).
		Where("imsi = ?", subscriber.IMSI).
		Select("balance").Scan(&balance).Error; err != nil {
		balance = 0.0
	}

	return &models.SubscriberAccount{
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
	}, nil
}

// ListInvoices lists invoices with pagination and filters.
func (s *SubscriberService) ListInvoices(ctx context.Context, limit int, offset int, imsi *string, status *models.InvoiceStatus) ([]*models.Invoice, int64, error) {
	query := s.db.DB.WithContext(ctx).Model(&models.Invoice{})

	if imsi != nil {
		var subscriber models.Subscriber
		if err := s.db.DB.WithContext(ctx).Where("imsi = ?", *imsi).First(&subscriber).Error; err != nil {
			return nil, 0, fmt.Errorf("subscriber not found: %w", err)
		}
		query = query.Where("subscriber_id = ?", subscriber.ID)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var invoices []*models.Invoice
	if err := query.Preload("Subscriber").Limit(limit).Offset(offset).
		Order("created_at DESC").Find(&invoices).Error; err != nil {
		return nil, 0, err
	}

	return invoices, total, nil
}

// GetInvoice retrieves a specific invoice by ID.
func (s *SubscriberService) GetInvoice(ctx context.Context, id string) (*models.Invoice, error) {
	invoiceID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice ID: %w", err)
	}

	var invoice models.Invoice
	if err := s.db.DB.WithContext(ctx).
		Preload("LineItems").
		Preload("Subscriber").
		First(&invoice, invoiceID).Error; err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}

	return &invoice, nil
}

// ListRatingPlans lists available active rating plans.
func (s *SubscriberService) ListRatingPlans(ctx context.Context) ([]*models.RatingPlan, error) {
	var plans []*models.RatingPlan
	if err := s.db.DB.WithContext(ctx).
		Where("is_active = ?", true).
		Order("monthly_fee ASC").
		Find(&plans).Error; err != nil {
		return nil, fmt.Errorf("failed to list rating plans: %w", err)
	}
	return plans, nil
}

// GetRatingPlan retrieves a specific rating plan by its plan_id.
func (s *SubscriberService) GetRatingPlan(ctx context.Context, planId string) (*models.RatingPlan, error) {
	var plan models.RatingPlan
	if err := s.db.DB.WithContext(ctx).
		Where("plan_id = ? AND is_active = ?", planId, true).
		First(&plan).Error; err != nil {
		return nil, fmt.Errorf("rating plan not found: %w", err)
	}
	return &plan, nil
}

// TopUpBalance tops up a subscriber's account balance.
func (s *SubscriberService) TopUpBalance(ctx context.Context, imsi string, req *models.TopUpRequest) (*models.SubscriberAccount, error) {
	var account models.SubscriberAccount

	if err := s.db.DB.WithContext(ctx).Where("imsi = ?", imsi).First(&account).Error; err != nil {
		return nil, fmt.Errorf("subscriber account not found: %w", err)
	}

	if err := s.db.DB.WithContext(ctx).
		Model(&account).
		Where("imsi = ?", imsi).
		Update("balance", account.Balance+req.Amount).
		Update("last_updated", time.Now()).Error; err != nil {
		return nil, fmt.Errorf("failed to top up balance: %w", err)
	}

	if err := s.db.DB.WithContext(ctx).Where("imsi = ?", imsi).First(&account).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve updated account: %w", err)
	}

	return &account, nil
}
