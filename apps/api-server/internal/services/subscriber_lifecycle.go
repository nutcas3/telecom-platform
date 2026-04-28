package services

import (
	"context"
	"fmt"
	"time"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// SuspendSubscriber suspends a subscriber and terminates their sessions.
func (s *SubscriberService) SuspendSubscriber(ctx context.Context, id uint) error {
	subscriber, err := s.db.GetSubscriber(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get subscriber: %w", err)
	}

	subscriber.Status = models.SubscriberStatusSuspended

	if err := s.terminateSubscriberSessions(ctx, subscriber.IMSI); err != nil {
		return fmt.Errorf("failed to terminate sessions: %w", err)
	}

	if err := s.db.UpdateSubscriber(ctx, subscriber); err != nil {
		return fmt.Errorf("failed to suspend subscriber: %w", err)
	}

	return nil
}

// TerminateSubscriber terminates a subscriber, sessions, and eSIM profile.
func (s *SubscriberService) TerminateSubscriber(ctx context.Context, id uint) error {
	subscriber, err := s.db.GetSubscriber(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get subscriber: %w", err)
	}

	subscriber.Status = models.SubscriberStatusTerminated

	if err := s.terminateSubscriberSessions(ctx, subscriber.IMSI); err != nil {
		return fmt.Errorf("failed to terminate sessions: %w", err)
	}

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

// ActivateSubscriber activates a subscriber.
func (s *SubscriberService) ActivateSubscriber(ctx context.Context, subscriberId uint) (*models.Subscriber, error) {
	var subscriber models.Subscriber

	if err := s.db.DB.WithContext(ctx).
		Model(&subscriber).
		Where("id = ?", subscriberId).
		Update("status", models.SubscriberStatusActive).
		Update("activated_at", time.Now()).Error; err != nil {
		return nil, fmt.Errorf("failed to activate subscriber: %w", err)
	}

	if err := s.db.DB.WithContext(ctx).
		Preload("Plan").
		First(&subscriber, subscriberId).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve activated subscriber: %w", err)
	}

	return &subscriber, nil
}

// DeleteSubscriber deletes a subscriber and all related data in a transaction.
func (s *SubscriberService) DeleteSubscriber(ctx context.Context, subscriberId uint) (bool, error) {
	tx := s.db.DB.WithContext(ctx).Begin()

	if err := tx.Where("subscriber_id = ?", subscriberId).Delete(&models.PaymentMethod{}).Error; err != nil {
		tx.Rollback()
		return false, fmt.Errorf("failed to delete payment methods: %w", err)
	}

	if err := tx.Where("subscriber_id = ?", subscriberId).Delete(&models.Alert{}).Error; err != nil {
		tx.Rollback()
		return false, fmt.Errorf("failed to delete alerts: %w", err)
	}

	if err := tx.Where("imsi IN (SELECT imsi FROM subscribers WHERE id = ?)", subscriberId).
		Delete(&models.SubscriberAccount{}).Error; err != nil {
		tx.Rollback()
		return false, fmt.Errorf("failed to delete account: %w", err)
	}

	if err := tx.Delete(&models.Subscriber{}, subscriberId).Error; err != nil {
		tx.Rollback()
		return false, fmt.Errorf("failed to delete subscriber: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return false, fmt.Errorf("failed to commit deletion: %w", err)
	}

	return true, nil
}

// terminateSubscriberSessions terminates all active sessions for a subscriber.
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

		if err := s.amfClient.TerminateSession(ctx, imsi, "Subscriber terminated"); err != nil {
			fmt.Printf("Failed to notify AMF for session termination: %v\n", err)
		}
	}

	return nil
}
