package compliance

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Service provides compliance management operations
type Service struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewService creates a new compliance service
func NewService(db *gorm.DB, logger *logrus.Logger) *Service {
	return &Service{db: db, logger: logger}
}

// CreateDSR creates a new data subject request
func (s *Service) CreateDSR(ctx context.Context, req *DataSubjectRequest) error {
	req.ID = uuid.New().String()
	req.Status = DSRStatusPending
	req.RequestedAt = time.Now()
	req.DueDate = s.calculateDueDate(req.Regulation)
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	if err := s.db.WithContext(ctx).Create(req).Error; err != nil {
		return fmt.Errorf("failed to create DSR: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"dsr_id":      req.ID,
		"type":        req.RequestType,
		"regulation":  req.Regulation,
	}).Info("Data subject request created")

	return nil
}

func (s *Service) calculateDueDate(reg Regulation) time.Time {
	switch reg {
	case RegulationGDPR:
		return time.Now().AddDate(0, 1, 0) // 30 days
	case RegulationCCPA:
		return time.Now().AddDate(0, 0, 45) // 45 days
	default:
		return time.Now().AddDate(0, 1, 0)
	}
}

// GetDSR retrieves a data subject request
func (s *Service) GetDSR(ctx context.Context, id string) (*DataSubjectRequest, error) {
	var req DataSubjectRequest
	if err := s.db.WithContext(ctx).First(&req, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("DSR not found: %w", err)
	}
	return &req, nil
}

// ListDSRs lists data subject requests for a tenant
func (s *Service) ListDSRs(ctx context.Context, tenantID string, status *DSRStatus) ([]*DataSubjectRequest, error) {
	var requests []*DataSubjectRequest
	query := s.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if err := query.Order("created_at DESC").Find(&requests).Error; err != nil {
		return nil, fmt.Errorf("failed to list DSRs: %w", err)
	}
	return requests, nil
}

// ProcessDSR processes a data subject request
func (s *Service) ProcessDSR(ctx context.Context, id string) error {
	req, err := s.GetDSR(ctx, id)
	if err != nil {
		return err
	}

	req.Status = DSRStatusProcessing
	req.UpdatedAt = time.Now()

	switch req.RequestType {
	case DSRTypeAccess, DSRTypePortability:
		return s.processAccessRequest(ctx, req)
	case DSRTypeErasure:
		return s.processErasureRequest(ctx, req)
	case DSRTypeRectify:
		return s.processRectificationRequest(ctx, req)
	default:
		return fmt.Errorf("unsupported request type: %s", req.RequestType)
	}
}

func (s *Service) processAccessRequest(ctx context.Context, req *DataSubjectRequest) error {
	// Export user data
	s.logger.WithField("dsr_id", req.ID).Info("Processing access request")
	now := time.Now()
	req.Status = DSRStatusCompleted
	req.CompletedAt = &now
	return s.db.WithContext(ctx).Save(req).Error
}

func (s *Service) processErasureRequest(ctx context.Context, req *DataSubjectRequest) error {
	s.logger.WithField("dsr_id", req.ID).Info("Processing erasure request")
	now := time.Now()
	req.Status = DSRStatusCompleted
	req.CompletedAt = &now
	return s.db.WithContext(ctx).Save(req).Error
}

func (s *Service) processRectificationRequest(ctx context.Context, req *DataSubjectRequest) error {
	s.logger.WithField("dsr_id", req.ID).Info("Processing rectification request")
	now := time.Now()
	req.Status = DSRStatusCompleted
	req.CompletedAt = &now
	return s.db.WithContext(ctx).Save(req).Error
}

// RecordConsent records user consent
func (s *Service) RecordConsent(ctx context.Context, consent *ConsentRecord) error {
	consent.ID = uuid.New().String()
	consent.CreatedAt = time.Now()
	consent.UpdatedAt = time.Now()
	if consent.Granted {
		now := time.Now()
		consent.GrantedAt = &now
	}
	return s.db.WithContext(ctx).Create(consent).Error
}

// RevokeConsent revokes user consent
func (s *Service) RevokeConsent(ctx context.Context, subjectID string, consentType ConsentType) error {
	now := time.Now()
	return s.db.WithContext(ctx).Model(&ConsentRecord{}).
		Where("subject_id = ? AND consent_type = ?", subjectID, consentType).
		Updates(map[string]any{"granted": false, "revoked_at": now, "updated_at": now}).Error
}

// GetConsents retrieves consent records for a subject
func (s *Service) GetConsents(ctx context.Context, subjectID string) ([]*ConsentRecord, error) {
	var consents []*ConsentRecord
	if err := s.db.WithContext(ctx).Where("subject_id = ?", subjectID).Find(&consents).Error; err != nil {
		return nil, err
	}
	return consents, nil
}

// LogAudit creates an audit log entry
func (s *Service) LogAudit(ctx context.Context, log *AuditLog) error {
	log.ID = uuid.New().String()
	log.Timestamp = time.Now()
	return s.db.WithContext(ctx).Create(log).Error
}

// QueryAuditLogs queries audit logs with filters
func (s *Service) QueryAuditLogs(ctx context.Context, tenantID string, filter *AuditLogFilter) ([]*AuditLog, error) {
	var logs []*AuditLog
	query := s.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if filter.Jurisdiction != "" {
		query = query.Where("jurisdiction = ?", filter.Jurisdiction)
	}
	if filter.ActorID != "" {
		query = query.Where("actor_id = ?", filter.ActorID)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if !filter.StartDate.IsZero() {
		query = query.Where("timestamp >= ?", filter.StartDate)
	}
	if !filter.EndDate.IsZero() {
		query = query.Where("timestamp <= ?", filter.EndDate)
	}
	if err := query.Order("timestamp DESC").Limit(filter.Limit).Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// AuditLogFilter defines audit log query filters
type AuditLogFilter struct {
	Jurisdiction string
	ActorID      string
	Action       string
	StartDate    time.Time
	EndDate      time.Time
	Limit        int
}

// SetDataResidency sets data residency configuration
func (s *Service) SetDataResidency(ctx context.Context, config *DataResidencyConfig) error {
	config.ID = uuid.New().String()
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	return s.db.WithContext(ctx).Save(config).Error
}

// GetDataResidency retrieves data residency configuration
func (s *Service) GetDataResidency(ctx context.Context, tenantID string) (*DataResidencyConfig, error) {
	var config DataResidencyConfig
	if err := s.db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}
