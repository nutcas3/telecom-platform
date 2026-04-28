package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/database"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// ConfigStoreService provides persisted runtime-tunable configuration entries
// stored in the database. The static `internal/config` package continues to
// provide startup/bootstrap configuration; this service complements it.
type ConfigStoreService struct {
	db *database.Database
}

func NewConfigStoreService(db *database.Database) *ConfigStoreService {
	return &ConfigStoreService{db: db}
}

func (s *ConfigStoreService) List(ctx context.Context, section string) ([]models.ConfigEntry, error) {
	q := s.db.DB.WithContext(ctx).Model(&models.ConfigEntry{})
	if section != "" {
		q = q.Where("section = ?", section)
	}
	var out []models.ConfigEntry
	if err := q.Order("section, key").Find(&out).Error; err != nil {
		return nil, err
	}
	// Mask sensitive values
	for i := range out {
		if out[i].Sensitive {
			out[i].Value = "********"
		}
	}
	return out, nil
}

type UpsertConfigInput struct {
	Section     string `json:"section" binding:"required"`
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value"`
	Type        string `json:"type"`
	Sensitive   bool   `json:"sensitive"`
	Description string `json:"description"`
	UpdatedBy   string `json:"updated_by"`
}

func (s *ConfigStoreService) Upsert(ctx context.Context, in UpsertConfigInput) (*models.ConfigEntry, error) {
	var e models.ConfigEntry
	err := s.db.DB.WithContext(ctx).
		Where("section = ? AND key = ?", in.Section, in.Key).
		First(&e).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	e.Section = in.Section
	e.Key = in.Key
	e.Value = in.Value
	e.Type = defaultStr(in.Type, "string")
	e.Sensitive = in.Sensitive
	e.Description = in.Description
	e.UpdatedBy = in.UpdatedBy
	if e.ID == 0 {
		e.CreatedAt = time.Now()
	}
	e.UpdatedAt = time.Now()
	if err := s.db.DB.WithContext(ctx).Save(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

// Validate runs type/constraint checks across all entries (or a single section).
func (s *ConfigStoreService) Validate(ctx context.Context, section string) (*ConfigValidationResult, error) {
	entries, err := s.List(ctx, section)
	if err != nil {
		return nil, err
	}
	res := &ConfigValidationResult{Valid: true}
	for _, e := range entries {
		switch e.Type {
		case "integer":
			if _, err := strconv.ParseInt(e.Value, 10, 64); err != nil && e.Value != "" {
				res.Valid = false
				res.Errors = append(res.Errors, ConfigIssue{
					Section: e.Section, Key: e.Key,
					Message: fmt.Sprintf("value %q is not a valid integer", e.Value),
					Code:    "INVALID_INTEGER",
				})
			}
		case "boolean":
			if e.Value != "" {
				if _, err := strconv.ParseBool(e.Value); err != nil {
					res.Valid = false
					res.Errors = append(res.Errors, ConfigIssue{
						Section: e.Section, Key: e.Key,
						Message: fmt.Sprintf("value %q is not a valid boolean", e.Value),
						Code:    "INVALID_BOOLEAN",
					})
				}
			}
		case "duration":
			if e.Value != "" {
				if _, err := time.ParseDuration(e.Value); err != nil {
					res.Valid = false
					res.Errors = append(res.Errors, ConfigIssue{
						Section: e.Section, Key: e.Key,
						Message: fmt.Sprintf("value %q is not a valid duration", e.Value),
						Code:    "INVALID_DURATION",
					})
				}
			}
		}
	}
	return res, nil
}

type ConfigValidationResult struct {
	Valid    bool          `json:"valid"`
	Errors   []ConfigIssue `json:"errors"`
	Warnings []ConfigIssue `json:"warnings"`
}

type ConfigIssue struct {
	Section string `json:"section"`
	Key     string `json:"key"`
	Message string `json:"message"`
	Code    string `json:"code"`
}
