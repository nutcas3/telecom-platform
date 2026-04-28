package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// ListAlerts lists alerts with optional filtering by subscriber, severity, and resolved status.
func (s *SubscriberService) ListAlerts(ctx context.Context, limit int, offset int, subscriberId *int, severity *models.AlertSeverity, resolved *bool) ([]*models.Alert, int64, error) {
	query := s.db.DB.WithContext(ctx).Model(&models.Alert{})

	if subscriberId != nil {
		query = query.Where("subscriber_id = ?", *subscriberId)
	}
	if severity != nil {
		query = query.Where("severity = ?", *severity)
	}
	if resolved != nil {
		query = query.Where("resolved = ?", *resolved)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var alerts []*models.Alert
	if err := query.Preload("Subscriber").Limit(limit).Offset(offset).
		Order("timestamp DESC").Find(&alerts).Error; err != nil {
		return nil, 0, err
	}

	return alerts, total, nil
}

// SearchSubscribers searches subscribers by name, MSISDN, or IMSI.
func (s *SubscriberService) SearchSubscribers(ctx context.Context, query string, limit int) ([]*models.Subscriber, error) {
	var subscribers []*models.Subscriber
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

// ResolveAlert marks an alert as resolved.
func (s *SubscriberService) ResolveAlert(ctx context.Context, alertId string) (*models.Alert, error) {
	alertID, err := strconv.ParseUint(alertId, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid alert ID: %w", err)
	}

	var alert models.Alert
	if err := s.db.DB.WithContext(ctx).
		Model(&alert).
		Where("id = ?", alertID).
		Update("resolved", true).
		Update("resolved_at", time.Now()).
		Error; err != nil {
		return nil, fmt.Errorf("failed to resolve alert: %w", err)
	}

	if err := s.db.DB.WithContext(ctx).
		Preload("Subscriber").
		First(&alert, alertID).Error; err != nil {
		return nil, fmt.Errorf("alert not found: %w", err)
	}

	return &alert, nil
}

// CreateAlert creates a new alert for a subscriber.
func (s *SubscriberService) CreateAlert(ctx context.Context, req *models.CreateAlertRequest) (*models.Alert, error) {
	alert := &models.Alert{
		Type:         req.Type,
		Severity:     req.Severity,
		Message:      req.Message,
		SubscriberID: req.SubscriberID,
		Timestamp:    time.Now(),
		Resolved:     false,
	}

	if err := s.db.DB.WithContext(ctx).Create(alert).Error; err != nil {
		return nil, fmt.Errorf("failed to create alert: %w", err)
	}

	if err := s.db.DB.WithContext(ctx).
		Preload("Subscriber").
		First(alert, alert.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve created alert: %w", err)
	}

	return alert, nil
}
