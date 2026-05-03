package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/mvno"
	"gorm.io/gorm"
)

// MVNOFilter defines filtering options for MVNO queries
type MVNOFilter struct {
	Status        mvno.MVNOStatus `json:"status,omitempty"`
	Plan          mvno.MVNOPlan   `json:"plan,omitempty"`
	BusinessID    string          `json:"business_id,omitempty"`
	Limit         int             `json:"limit,omitempty"`
	Offset        int             `json:"offset,omitempty"`
	CreatedAfter  *time.Time      `json:"created_after,omitempty"`
	CreatedBefore *time.Time      `json:"created_before,omitempty"`
}

// CreateMVNO creates a new MVNO record
func (r *GormRepository) CreateMVNO(ctx context.Context, mvno *mvno.MVNO) error {
	if err := r.db.WithContext(ctx).Create(mvno).Error; err != nil {
		return fmt.Errorf("failed to create MVNO: %w", err)
	}

	r.logger.Info("MVNO created", "id", mvno.ID, "business_id", mvno.BusinessID)
	return nil
}

// GetMVNO retrieves an MVNO by ID
func (r *GormRepository) GetMVNO(ctx context.Context, id string) (*mvno.MVNO, error) {
	var mvno mvno.MVNO
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&mvno).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("MVNO not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get MVNO: %w", err)
	}
	return &mvno, nil
}

// GetMVNOByBusinessID retrieves an MVNO by business ID
func (r *GormRepository) GetMVNOByBusinessID(ctx context.Context, businessID string) (*mvno.MVNO, error) {
	var mvno mvno.MVNO
	if err := r.db.WithContext(ctx).Where("business_id = ?", businessID).First(&mvno).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("MVNO not found for business ID: %s", businessID)
		}
		return nil, fmt.Errorf("failed to get MVNO by business ID: %w", err)
	}
	return &mvno, nil
}

// UpdateMVNO updates an existing MVNO
func (r *GormRepository) UpdateMVNO(ctx context.Context, mvno *mvno.MVNO) error {
	if err := r.db.WithContext(ctx).Save(mvno).Error; err != nil {
		return fmt.Errorf("failed to update MVNO: %w", err)
	}

	r.logger.Info("MVNO updated", "id", mvno.ID, "status", mvno.Status)
	return nil
}

// ListMVNOs lists MVNOs with optional filtering
func (r *GormRepository) ListMVNOs(ctx context.Context, filter *mvno.MVNOFilter) ([]*mvno.MVNO, error) {
	query := r.db.WithContext(ctx).Model(&mvno.MVNO{})

	// Apply filters
	if filter != nil {
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.Plan != "" {
			query = query.Where("plan = ?", filter.Plan)
		}
		if filter.BusinessID != "" {
			query = query.Where("business_id = ?", filter.BusinessID)
		}
		if filter.CreatedAfter != nil {
			query = query.Where("created_at >= ?", *filter.CreatedAfter)
		}
		if filter.CreatedBefore != nil {
			query = query.Where("created_at <= ?", *filter.CreatedBefore)
		}
		if filter.Limit > 0 {
			query = query.Limit(filter.Limit)
		}
		if filter.Offset > 0 {
			query = query.Offset(filter.Offset)
		}
	}

	var mvnos []*mvno.MVNO
	if err := query.Order("created_at DESC").Find(&mvnos).Error; err != nil {
		return nil, fmt.Errorf("failed to list MVNOs: %w", err)
	}

	return mvnos, nil
}

// DeleteMVNO soft deletes an MVNO
func (r *GormRepository) DeleteMVNO(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&mvno.MVNO{}).Error; err != nil {
		return fmt.Errorf("failed to delete MVNO: %w", err)
	}

	r.logger.Info("MVNO deleted", "id", id)
	return nil
}

// GetMVNOStats returns statistics about MVNOs
func (r *GormRepository) GetMVNOStats(ctx context.Context) (map[string]any, error) {
	stats := make(map[string]any)

	// Total count
	var totalCount int64
	if err := r.db.WithContext(ctx).Model(&mvno.MVNO{}).Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count total MVNOs: %w", err)
	}
	stats["total"] = totalCount

	// Count by status
	var statusCounts []struct {
		Status mvno.MVNOStatus `gorm:"column:status"`
		Count  int64           `gorm:"column:count"`
	}
	if err := r.db.WithContext(ctx).Model(&mvno.MVNO{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to count by status: %w", err)
	}

	statusMap := make(map[string]int64)
	for _, sc := range statusCounts {
		statusMap[string(sc.Status)] = sc.Count
	}
	stats["by_status"] = statusMap

	// Count by plan
	var planCounts []struct {
		Plan  mvno.MVNOPlan `gorm:"column:plan"`
		Count int64         `gorm:"column:count"`
	}
	if err := r.db.WithContext(ctx).Model(&mvno.MVNO{}).
		Select("plan, COUNT(*) as count").
		Group("plan").
		Scan(&planCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to count by plan: %w", err)
	}

	planMap := make(map[string]int64)
	for _, pc := range planCounts {
		planMap[string(pc.Plan)] = pc.Count
	}
	stats["by_plan"] = planMap

	return stats, nil
}

// UpdateMVNOStatus updates only the status of an MVNO
func (r *GormRepository) UpdateMVNOStatus(ctx context.Context, id string, status mvno.MVNOStatus) error {
	if err := r.db.WithContext(ctx).Model(&mvno.MVNO{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update MVNO status: %w", err)
	}

	r.logger.Info("MVNO status updated", "id", id, "new_status", status)
	return nil
}
