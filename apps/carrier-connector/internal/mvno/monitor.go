package mvno

import (
	"maps"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// OnboardingMonitor tracks onboarding progress and status
type OnboardingMonitor struct {
	logger   *logrus.Logger
	progress map[string]*OnboardingProgress
	mu       sync.RWMutex
}

// NewOnboardingMonitor creates a new monitor instance
func NewOnboardingMonitor(logger *logrus.Logger) *OnboardingMonitor {
	return &OnboardingMonitor{
		logger:   logger,
		progress: make(map[string]*OnboardingProgress),
	}
}

// UpdateProgress updates the onboarding progress for an MVNO
func (m *OnboardingMonitor) UpdateProgress(mvnoID string, progress *OnboardingProgress) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.progress[mvnoID] = progress

	m.logger.WithFields(map[string]any{
		"mvno_id":  mvnoID,
		"progress": progress.Progress,
		"status":   m.getCurrentStatus(progress),
	}).Info("Onboarding progress updated")
}

// GetProgress retrieves current onboarding progress
func (m *OnboardingMonitor) GetProgress(mvnoID string) (*OnboardingProgress, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	progress, exists := m.progress[mvnoID]
	return progress, exists
}

// GetAllProgress retrieves progress for all MVNOs
func (m *OnboardingMonitor) GetAllProgress() map[string]*OnboardingProgress {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid concurrent access issues
	result := make(map[string]*OnboardingProgress)
	maps.Copy(result, m.progress)
	return result
}

// GetActiveOnboardingCount returns count of active onboarding processes
func (m *OnboardingMonitor) GetActiveOnboardingCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, progress := range m.progress {
		if progress.Progress < 100.0 && !progress.CompletedAt.IsZero() {
			count++
		}
	}
	return count
}

// GetCompletedOnboardingCount returns count of completed onboardings
func (m *OnboardingMonitor) GetCompletedOnboardingCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, progress := range m.progress {
		if progress.Progress == 100.0 && !progress.CompletedAt.IsZero() {
			count++
		}
	}
	return count
}

// GetAverageOnboardingTime calculates average onboarding duration
func (m *OnboardingMonitor) GetAverageOnboardingTime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var totalDuration time.Duration
	completedCount := 0

	for _, progress := range m.progress {
		if progress.Progress == 100.0 && !progress.CompletedAt.IsZero() && !progress.StartedAt.IsZero() {
			totalDuration += progress.CompletedAt.Sub(progress.StartedAt)
			completedCount++
		}
	}

	if completedCount == 0 {
		return 0
	}

	return totalDuration / time.Duration(completedCount)
}

// GetStepSuccessRate calculates success rate for each step
func (m *OnboardingMonitor) GetStepSuccessRate() map[string]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stepStats := make(map[string]map[string]int)

	// Collect step statistics
	for _, progress := range m.progress {
		for _, step := range progress.Steps {
			if _, exists := stepStats[step.Name]; !exists {
				stepStats[step.Name] = map[string]int{"completed": 0, "failed": 0, "total": 0}
			}
			stepStats[step.Name]["total"]++
			switch step.Status {
case "completed":
				stepStats[step.Name]["completed"]++
			case "failed":
				stepStats[step.Name]["failed"]++
			}
		}
	}

	// Calculate success rates
	successRates := make(map[string]float64)
	for stepName, stats := range stepStats {
		if stats["total"] > 0 {
			successRates[stepName] = float64(stats["completed"]) / float64(stats["total"]) * 100
		}
	}

	return successRates
}

// CleanupOldProgress removes old completed progress records
func (m *OnboardingMonitor) CleanupOldProgress(maxAge time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	for mvnoID, progress := range m.progress {
		if progress.Progress == 100.0 && !progress.CompletedAt.IsZero() && progress.CompletedAt.Before(cutoff) {
			delete(m.progress, mvnoID)
			m.logger.WithFields(map[string]any{
				"mvno_id":      mvnoID,
				"completed_at": progress.CompletedAt,
			}).Info("Cleaned up old onboarding progress")
		}
	}
}

// GetFailedOnboardings returns list of failed onboardings
func (m *OnboardingMonitor) GetFailedOnboardings() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var failed []string
	for mvnoID, progress := range m.progress {
		hasFailed := false
		for _, step := range progress.Steps {
			if step.Status == "failed" {
				hasFailed = true
				break
			}
		}
		if hasFailed {
			failed = append(failed, mvnoID)
		}
	}
	return failed
}

// getCurrentStatus determines current status from progress
func (m *OnboardingMonitor) getCurrentStatus(progress *OnboardingProgress) string {
	if progress.Progress == 100.0 {
		return "completed"
	}

	if progress.Progress == 0.0 {
		return "pending"
	}

	// Check for failed steps
	for _, step := range progress.Steps {
		if step.Status == "failed" {
			return "failed"
		}
	}

	return "in_progress"
}

// GetOnboardingMetrics returns comprehensive onboarding metrics
func (m *OnboardingMonitor) GetOnboardingMetrics() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]any{
		"active_count":       m.GetActiveOnboardingCount(),
		"completed_count":    m.GetCompletedOnboardingCount(),
		"average_duration":   m.GetAverageOnboardingTime().String(),
		"step_success_rates": m.GetStepSuccessRate(),
		"failed_count":       len(m.GetFailedOnboardings()),
		"total_onboardings":  len(m.progress),
	}
}
