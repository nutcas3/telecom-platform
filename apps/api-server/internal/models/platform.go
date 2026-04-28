package models

import (
	"time"

	"gorm.io/gorm"
)

// Plugin represents a real, persisted plugin record.
type Plugin struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"uniqueIndex;not null"`
	Version     string         `json:"version" gorm:"not null"`
	Description string         `json:"description"`
	Author      string         `json:"author"`
	Type        string         `json:"type"`
	Category    string         `json:"category"`
	License     string         `json:"license"`
	Homepage    string         `json:"homepage"`
	Repository  string         `json:"repository"`
	Enabled     bool           `json:"enabled" gorm:"default:false"`
	Status      string         `json:"status" gorm:"default:'installed'"`
	Config      string         `json:"config,omitempty" gorm:"type:jsonb"`
	InstalledAt time.Time      `json:"installed_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// Automation represents a workflow automation definition.
type Automation struct {
	ID                  uint           `json:"id" gorm:"primaryKey"`
	Name                string         `json:"name" gorm:"uniqueIndex;not null"`
	Description         string         `json:"description"`
	Type                string         `json:"type"`
	Enabled             bool           `json:"enabled" gorm:"default:true"`
	ScheduleType        string         `json:"schedule_type"` // cron|interval|manual
	ScheduleCron        string         `json:"schedule_cron,omitempty"`
	ScheduleIntervalSec int            `json:"schedule_interval_sec,omitempty"`
	Timezone            string         `json:"timezone" gorm:"default:'UTC'"`
	Definition          string         `json:"definition" gorm:"type:jsonb"` // actions/conditions payload
	LastRunAt           *time.Time     `json:"last_run_at,omitempty"`
	NextRunAt           *time.Time     `json:"next_run_at,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `json:"-" gorm:"index"`
}

// AutomationRun represents an execution of an automation.
type AutomationRun struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	AutomationID uint       `json:"automation_id" gorm:"index;not null"`
	Status       string     `json:"status"` // pending|running|success|failed
	StartedAt    time.Time  `json:"started_at"`
	EndedAt      *time.Time `json:"ended_at,omitempty"`
	DurationMS   int64      `json:"duration_ms"`
	Output       string     `json:"output"`
	Error        string     `json:"error,omitempty"`
	Details      string     `json:"details,omitempty" gorm:"type:jsonb"`
	CreatedAt    time.Time  `json:"created_at"`
}

// ConfigEntry represents a persisted runtime-tunable config value.
type ConfigEntry struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Section     string    `json:"section" gorm:"index:idx_section_key,unique;not null"`
	Key         string    `json:"key" gorm:"index:idx_section_key,unique;not null"`
	Value       string    `json:"value"`
	Type        string    `json:"type"` // string|integer|boolean|duration|json
	Sensitive   bool      `json:"sensitive" gorm:"default:false"`
	Description string    `json:"description"`
	UpdatedBy   string    `json:"updated_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DeploymentRecord is a persisted deployment history entry.
type DeploymentRecord struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Service     string     `json:"service" gorm:"index;not null"`
	Version     string     `json:"version" gorm:"not null"`
	Environment string     `json:"environment" gorm:"index"`
	Status      string     `json:"status"` // pending|in_progress|completed|failed|rolled_back
	Strategy    string     `json:"strategy"`
	Replicas    int        `json:"replicas"`
	TriggeredBy string     `json:"triggered_by"`
	Reason      string     `json:"reason"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	RollbackTo  string     `json:"rollback_to,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ChaosExperimentRecord persists chaos experiment history.
type ChaosExperimentRecord struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	ExternalID  string     `json:"external_id" gorm:"uniqueIndex"`
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	Target      string     `json:"target"`
	Status      string     `json:"status"`
	DurationMS  int64      `json:"duration_ms"`
	Probability float64    `json:"probability"`
	Amount      int        `json:"amount"`
	StartedAt   time.Time  `json:"started_at"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
	Error       string     `json:"error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
