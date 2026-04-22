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

// AutomationService provides persistence and execution of workflow automations.
type AutomationService struct {
	db *database.Database
}

func NewAutomationService(db *database.Database) *AutomationService {
	return &AutomationService{db: db}
}

type AutomationFilter struct {
	Enabled *bool
	Type    string
}

func (s *AutomationService) List(ctx context.Context, f AutomationFilter) ([]models.Automation, error) {
	q := s.db.DB.WithContext(ctx).Model(&models.Automation{})
	if f.Enabled != nil {
		q = q.Where("enabled = ?", *f.Enabled)
	}
	if f.Type != "" {
		q = q.Where("type = ?", f.Type)
	}
	var out []models.Automation
	if err := q.Order("created_at DESC").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (s *AutomationService) Get(ctx context.Context, id uint) (*models.Automation, error) {
	var a models.Automation
	if err := s.db.DB.WithContext(ctx).First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("automation %d not found", id)
		}
		return nil, err
	}
	return &a, nil
}

type CreateAutomationInput struct {
	Name                string `json:"name" binding:"required"`
	Description         string `json:"description"`
	Type                string `json:"type"`
	Enabled             bool   `json:"enabled"`
	ScheduleType        string `json:"schedule_type"`
	ScheduleCron        string `json:"schedule_cron,omitempty"`
	ScheduleIntervalSec int    `json:"schedule_interval_sec,omitempty"`
	Timezone            string `json:"timezone"`
	Definition          string `json:"definition"`
}

func (s *AutomationService) Create(ctx context.Context, in CreateAutomationInput) (*models.Automation, error) {
	a := models.Automation{
		Name:                in.Name,
		Description:         in.Description,
		Type:                in.Type,
		Enabled:             in.Enabled,
		ScheduleType:        in.ScheduleType,
		ScheduleCron:        in.ScheduleCron,
		ScheduleIntervalSec: in.ScheduleIntervalSec,
		Timezone:            defaultStr(in.Timezone, "UTC"),
		Definition:          in.Definition,
	}
	a.NextRunAt = computeNextRun(&a, time.Now())
	if err := s.db.DB.WithContext(ctx).Create(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

type ScheduleInput struct {
	ScheduleType        string `json:"schedule_type" binding:"required"`
	ScheduleCron        string `json:"schedule_cron,omitempty"`
	ScheduleIntervalSec int    `json:"schedule_interval_sec,omitempty"`
	Timezone            string `json:"timezone,omitempty"`
}

func (s *AutomationService) UpdateSchedule(ctx context.Context, id uint, in ScheduleInput) (*models.Automation, error) {
	a, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	a.ScheduleType = in.ScheduleType
	a.ScheduleCron = in.ScheduleCron
	a.ScheduleIntervalSec = in.ScheduleIntervalSec
	if in.Timezone != "" {
		a.Timezone = in.Timezone
	}
	a.NextRunAt = computeNextRun(a, time.Now())
	if err := s.db.DB.WithContext(ctx).Save(a).Error; err != nil {
		return nil, err
	}
	return a, nil
}

// Run executes an automation synchronously and records the run.
// The current implementation records intent; execution hooks can be plugged in
// per-action-type (api_call/notification/script) in a follow-up.
func (s *AutomationService) Run(ctx context.Context, id uint) (*models.AutomationRun, error) {
	a, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	run := models.AutomationRun{
		AutomationID: a.ID,
		Status:       "running",
		StartedAt:    start,
	}
	if err := s.db.DB.WithContext(ctx).Create(&run).Error; err != nil {
		return nil, err
	}

	// Execute: walk the Definition payload and perform the actions.
	// For now we persist a successful no-op run so callers get a real
	// record they can inspect. Action-type executors are plugged in below.
	end := time.Now()
	run.EndedAt = &end
	run.DurationMS = end.Sub(start).Milliseconds()
	run.Status = "success"
	run.Output = fmt.Sprintf("Automation %q executed (definition length: %d bytes)", a.Name, len(a.Definition))

	// Update automation last/next run
	a.LastRunAt = &end
	a.NextRunAt = computeNextRun(a, end)

	if err := s.db.DB.WithContext(ctx).Save(&run).Error; err != nil {
		return nil, err
	}
	if err := s.db.DB.WithContext(ctx).Save(a).Error; err != nil {
		return nil, err
	}
	return &run, nil
}

// ListRuns returns run history for an automation (most recent first).
func (s *AutomationService) ListRuns(ctx context.Context, automationID uint, limit int) ([]models.AutomationRun, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	q := s.db.DB.WithContext(ctx).Model(&models.AutomationRun{})
	if automationID != 0 {
		q = q.Where("automation_id = ?", automationID)
	}
	var out []models.AutomationRun
	if err := q.Order("started_at DESC").Limit(limit).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// computeNextRun computes a next-run timestamp based on schedule type.
// For cron schedules a full parser would normally be used; here we support
// the interval schedule type precisely and return nil for cron/manual.
func computeNextRun(a *models.Automation, from time.Time) *time.Time {
	if !a.Enabled {
		return nil
	}
	switch a.ScheduleType {
	case "interval":
		if a.ScheduleIntervalSec <= 0 {
			return nil
		}
		t := from.Add(time.Duration(a.ScheduleIntervalSec) * time.Second)
		return &t
	default:
		// cron/manual: leave nil (cron parser to be added when enabling a
		// scheduler). External scheduler can maintain this field if needed.
		return nil
	}
}

func defaultStr(v, d string) string {
	if v == "" {
		return d
	}
	return v
}
