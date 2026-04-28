package services

import (
	"context"
	"fmt"
	"time"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// AddPaymentMethod creates a new payment method for the given subscriber.
func (s *SubscriberService) AddPaymentMethod(ctx context.Context, subscriberId int, req *models.AddPaymentMethodRequest) (*models.PaymentMethod, error) {
	paymentMethodID := fmt.Sprintf("pm_%d_%d", subscriberId, time.Now().Unix())

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

	if err := s.db.DB.WithContext(ctx).Create(paymentMethod).Error; err != nil {
		return nil, fmt.Errorf("failed to add payment method: %w", err)
	}

	return paymentMethod, nil
}

// processPaymentToken resolves a Stripe token to card details.
func (s *SubscriberService) processPaymentToken(ctx context.Context, token string) (last4, brand string, expiryMonth, expiryYear int, err error) {
	details, err := s.stripeGW.RetrievePaymentMethodFromToken(ctx, token)
	if err != nil {
		return "", "", 0, 0, fmt.Errorf("failed to retrieve payment method from token: %w", err)
	}
	return details.Last4, details.Brand, details.ExpiryMonth, details.ExpiryYear, nil
}

// RemovePaymentMethod removes a payment method, preventing removal of the only one and
// reassigning the default before deletion.
func (s *SubscriberService) RemovePaymentMethod(ctx context.Context, paymentMethodId string) (bool, error) {
	var paymentMethod models.PaymentMethod
	if err := s.db.DB.WithContext(ctx).First(&paymentMethod, "id = ?", paymentMethodId).Error; err != nil {
		return false, fmt.Errorf("payment method not found: %w", err)
	}

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

	if err := s.db.DB.WithContext(ctx).Delete(&paymentMethod).Error; err != nil {
		return false, fmt.Errorf("failed to delete payment method: %w", err)
	}

	return true, nil
}

// SetDefaultPaymentMethod marks one payment method as default and unsets others in a transaction.
func (s *SubscriberService) SetDefaultPaymentMethod(ctx context.Context, paymentMethodId string) (*models.PaymentMethod, error) {
	var paymentMethod models.PaymentMethod
	tx := s.db.DB.WithContext(ctx).Begin()

	if err := tx.Model(&models.PaymentMethod{}).
		Where("id = ?", paymentMethodId).
		First(&paymentMethod).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("payment method not found: %w", err)
	}

	if err := tx.Model(&models.PaymentMethod{}).
		Where("subscriber_id = ? AND id != ?", paymentMethod.SubscriberID, paymentMethodId).
		Update("is_default", false).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update other payment methods: %w", err)
	}

	if err := tx.Model(&paymentMethod).
		Update("is_default", true).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to set default payment method: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &paymentMethod, nil
}
