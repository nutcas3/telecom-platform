package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/database"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// DeploymentService persists deployment records. Actual orchestration
// (Helm/ArgoCD/kubectl) is handled by an external controller or operator;
// this service is the system of record for deployment history and status.
type DeploymentService struct {
	db *database.Database
}

func NewDeploymentService(db *database.Database) *DeploymentService {
	return &DeploymentService{db: db}
}

type DeploymentFilter struct {
	Service     string
	Environment string
	Status      string
	Limit       int
	Offset      int
}

func (s *DeploymentService) List(ctx context.Context, f DeploymentFilter) ([]models.DeploymentRecord, int64, error) {
	q := s.db.DB.WithContext(ctx).Model(&models.DeploymentRecord{})
	if f.Service != "" {
		q = q.Where("service = ?", f.Service)
	}
	if f.Environment != "" {
		q = q.Where("environment = ?", f.Environment)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if f.Limit <= 0 {
		f.Limit = 20
	}
	var out []models.DeploymentRecord
	if err := q.Order("created_at DESC").Limit(f.Limit).Offset(f.Offset).Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (s *DeploymentService) Get(ctx context.Context, id uint) (*models.DeploymentRecord, error) {
	var d models.DeploymentRecord
	if err := s.db.DB.WithContext(ctx).First(&d, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("deployment %d not found", id)
		}
		return nil, err
	}
	return &d, nil
}

type StartDeploymentInput struct {
	Service     string `json:"service" binding:"required"`
	Version     string `json:"version" binding:"required"`
	Environment string `json:"environment" binding:"required"`
	Strategy    string `json:"strategy"`
	Replicas    int    `json:"replicas"`
	TriggeredBy string `json:"triggered_by"`
	Reason      string `json:"reason"`
}

func (s *DeploymentService) Start(ctx context.Context, in StartDeploymentInput) (*models.DeploymentRecord, error) {
	now := time.Now()
	d := models.DeploymentRecord{
		Service:     in.Service,
		Version:     in.Version,
		Environment: in.Environment,
		Strategy:    defaultStr(in.Strategy, "rolling"),
		Replicas:    in.Replicas,
		TriggeredBy: defaultStr(in.TriggeredBy, "api"),
		Reason:      in.Reason,
		Status:      "in_progress",
		StartedAt:   &now,
	}
	if err := s.db.DB.WithContext(ctx).Create(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

type RollbackInput struct {
	Service   string `json:"service" binding:"required"`
	ToVersion string `json:"to_version" binding:"required"`
	Reason    string `json:"reason"`
	TriggeredBy string `json:"triggered_by"`
}

func (s *DeploymentService) Rollback(ctx context.Context, in RollbackInput) (*models.DeploymentRecord, error) {
	// Ensure the target version exists in history for that service.
	var prior models.DeploymentRecord
	err := s.db.DB.WithContext(ctx).
		Where("service = ? AND version = ?", in.Service, in.ToVersion).
		Order("created_at DESC").
		First(&prior).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no prior deployment of %s at version %s", in.Service, in.ToVersion)
		}
		return nil, err
	}

	now := time.Now()
	d := models.DeploymentRecord{
		Service:     in.Service,
		Version:     in.ToVersion,
		Environment: prior.Environment,
		Strategy:    "rolling",
		Replicas:    prior.Replicas,
		TriggeredBy: defaultStr(in.TriggeredBy, "api"),
		Reason:      in.Reason,
		Status:      "rolled_back",
		StartedAt:   &now,
		RollbackTo:  in.ToVersion,
	}
	if err := s.db.DB.WithContext(ctx).Create(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

// MarkCompleted updates deployment status; intended to be called by the
// orchestration layer (controller) once the real rollout finishes.
func (s *DeploymentService) MarkCompleted(ctx context.Context, id uint, success bool) (*models.DeploymentRecord, error) {
	d, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	d.CompletedAt = &now
	if success {
		d.Status = "completed"
	} else {
		d.Status = "failed"
	}
	if err := s.db.DB.WithContext(ctx).Save(d).Error; err != nil {
		return nil, err
	}
	return d, nil
}
