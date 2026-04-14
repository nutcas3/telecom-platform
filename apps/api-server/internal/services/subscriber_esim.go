package services

import (
	"context"
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// provisionESIMProfile provisions an eSIM profile for the subscriber
func (s *SubscriberService) provisionESIMProfile(ctx context.Context, subscriberID uint) error {
	// Get subscriber details
	subscriber, err := s.db.GetSubscriber(ctx, subscriberID)
	if err != nil {
		return fmt.Errorf("failed to get subscriber: %w", err)
	}

	// Validate EID
	if err := s.es2Service.ValidateEID(subscriber.EUICCID); err != nil {
		return fmt.Errorf("invalid EID: %w", err)
	}

	// Provision profile via ES2+ API
	profileInfo, err := s.es2Service.ProvisionProfile(ctx, subscriber)
	if err != nil {
		return fmt.Errorf("failed to provision profile: %w", err)
	}

	// Update subscriber with profile information
	subscriber.ProfileID = profileInfo.ProfileID
	subscriber.ProfileStatus = models.ProfileStatusDownloading

	if err := s.db.UpdateSubscriber(ctx, subscriber); err != nil {
		return fmt.Errorf("failed to update subscriber: %w", err)
	}

	// Activate profile
	if err := s.es2Service.ActivateProfile(ctx, subscriber.EUICCID, profileInfo.ProfileID); err != nil {
		return fmt.Errorf("failed to activate profile: %w", err)
	}

	// Update subscriber status
	subscriber.ProfileStatus = models.ProfileStatusActive
	if err := s.db.UpdateSubscriber(ctx, subscriber); err != nil {
		return fmt.Errorf("failed to update subscriber status: %w", err)
	}

	// Notify AMF of new subscriber
	if err := s.amfClient.NotifySubscriberUpdate(ctx, subscriber.IMSI, models.SubscriberStatusActive); err != nil {
		fmt.Printf("Failed to notify AMF of subscriber activation: %v\n", err)
	}

	return nil
}

// deactivateESIMProfile deactivates an eSIM profile
func (s *SubscriberService) deactivateESIMProfile(ctx context.Context, subscriberID uint) error {
	// Get subscriber details
	subscriber, err := s.db.GetSubscriber(ctx, subscriberID)
	if err != nil {
		return fmt.Errorf("failed to get subscriber: %w", err)
	}

	// Check if subscriber has active profile
	if subscriber.ProfileID == "" || subscriber.ProfileStatus != models.ProfileStatusActive {
		return fmt.Errorf("no active profile to deactivate")
	}

	// Deactivate profile via ES2+ API
	if err := s.es2Service.DeactivateProfile(ctx, subscriber.EUICCID, subscriber.ProfileID); err != nil {
		return fmt.Errorf("failed to deactivate profile: %w", err)
	}

	// Update subscriber status
	subscriber.ProfileStatus = models.ProfileStatusInactive
	subscriber.Status = models.SubscriberStatusInactive

	if err := s.db.UpdateSubscriber(ctx, subscriber); err != nil {
		return fmt.Errorf("failed to update subscriber: %w", err)
	}

	// Terminate all sessions
	if err := s.terminateSubscriberSessions(ctx, subscriber.IMSI); err != nil {
		fmt.Printf("Failed to terminate sessions: %v\n", err)
	}

	// Notify AMF of subscriber deactivation
	if err := s.amfClient.NotifySubscriberUpdate(ctx, subscriber.IMSI, models.SubscriberStatusInactive); err != nil {
		fmt.Printf("Failed to notify AMF of subscriber deactivation: %v\n", err)
	}

	return nil
}
