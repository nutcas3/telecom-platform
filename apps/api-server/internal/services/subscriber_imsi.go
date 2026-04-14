package services

import (
	"context"
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// allocateIMSI allocates a new IMSI for a subscriber
func (s *SubscriberService) allocateIMSI(ctx context.Context) (models.IMSI, error) {
	// Get current IMSI allocation
	alloc, err := s.db.GetIMSIAllocation(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get IMSI allocation: %w", err)
	}

	// Check if we've reached the maximum
	if alloc.LastIMSI >= alloc.MaxIMSI {
		return "", fmt.Errorf("IMSI range exhausted: max %d reached", alloc.MaxIMSI)
	}

	// Allocate next IMSI
	nextIMSI := alloc.LastIMSI + 1
	alloc.LastIMSI = nextIMSI

	// Update allocation state
	if err := s.db.UpdateIMSIAllocation(ctx, alloc); err != nil {
		return "", fmt.Errorf("failed to update IMSI allocation: %w", err)
	}

	// Format IMSI: MCC (3) + MNC (2-3) + subscriber number
	imsiStr := fmt.Sprintf("%s%010d", s.config.IMSI.Prefix, nextIMSI)
	return models.IMSI(imsiStr), nil
}
