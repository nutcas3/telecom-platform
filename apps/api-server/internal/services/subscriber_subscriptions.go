package services

import (
	"context"
	"strconv"
	"time"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// SubscribeToSubscriberUpdates returns a channel that emits the subscriber whenever
// their row is updated. Implemented via polling; swap for LISTEN/NOTIFY when available.
func (s *SubscriberService) SubscribeToSubscriberUpdates(ctx context.Context, subscriberId string) (<-chan *models.Subscriber, error) {
	ch := make(chan *models.Subscriber, 5)

	go func() {
		defer close(ch)

		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		var lastUpdatedAt time.Time

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				var subscriber models.Subscriber
				query := s.db.DB.WithContext(ctx).Where("imsi = ?", subscriberId)
				if !lastUpdatedAt.IsZero() {
					query = query.Where("updated_at > ?", lastUpdatedAt)
				}

				if err := query.First(&subscriber).Error; err != nil {
					continue
				}

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

// SubscribeToAlertUpdates returns a channel that emits newly created alerts, optionally
// filtered by subscriberId (numeric string).
func (s *SubscriberService) SubscribeToAlertUpdates(ctx context.Context, subscriberId *string) (<-chan *models.Alert, error) {
	ch := make(chan *models.Alert, 10)

	go func() {
		defer close(ch)

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		var lastAlertID uint

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				var alerts []models.Alert
				query := s.db.DB.WithContext(ctx).Where("id > ?", lastAlertID)

				if subscriberId != nil {
					if subID, err := strconv.ParseUint(*subscriberId, 10, 32); err == nil {
						query = query.Where("subscriber_id = ?", subID)
					}
				}

				if err := query.Order("id ASC").Limit(10).Find(&alerts).Error; err != nil {
					continue
				}

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
