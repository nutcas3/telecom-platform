package security

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// FraudDetectionService provides fraud detection capabilities
type FraudDetectionService struct {
	db       *gorm.DB
	logger   *logrus.Logger
	patterns []FraudPattern
	alerts   []*FraudAlert
	mu       sync.RWMutex
}

// NewFraudDetectionService creates a new fraud detection service
func NewFraudDetectionService(db *gorm.DB, logger *logrus.Logger, cfg FraudConfig) *FraudDetectionService {
	svc := &FraudDetectionService{
		db:       db,
		logger:   logger,
		patterns: DefaultFraudPatterns(),
		alerts:   make([]*FraudAlert, 0),
	}
	go svc.cleanupAlerts(cfg.AlertRetentionDays)
	return svc
}

// AnalyzeTransaction analyzes a transaction for fraud
func (s *FraudDetectionService) AnalyzeTransaction(ctx context.Context, tx map[string]interface{}) (*FraudAlert, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	alert := &FraudAlert{
		ID:        fmt.Sprintf("fraud-%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Status:    "new",
		Metadata:  make(map[string]any),
	}

	if id, ok := tx["profile_id"].(string); ok {
		alert.ProfileID = id
	}
	if ip, ok := tx["ip_address"].(string); ok {
		alert.IPAddress = ip
	}

	score, evidence := s.evaluatePatterns(ctx, tx)
	alert.RiskScore = math.Min(100, score)
	alert.Evidence = evidence
	alert.Severity = SeverityFromScore(alert.RiskScore)
	alert.Type = s.detectType(evidence)
	alert.Description = fmt.Sprintf("%s %s fraud detected", alert.Severity, alert.Type)

	if alert.RiskScore >= 80 {
		alert.Actions = append(alert.Actions, "auto_blocked")
		s.blockProfile(ctx, alert.ProfileID)
	} else if alert.RiskScore >= 60 {
		alert.Actions = append(alert.Actions, "flagged_for_review")
	}

	s.alerts = append(s.alerts, alert)
	s.logger.WithField("alert_id", alert.ID).WithField("risk_score", alert.RiskScore).Warn("Fraud detected")

	return alert, nil
}

// GetFraudAlerts retrieves fraud alerts
func (s *FraudDetectionService) GetFraudAlerts(_ context.Context, filter FraudAlertFilter) ([]*FraudAlert, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*FraudAlert
	for _, a := range s.alerts {
		if s.matchesFilter(a, filter) {
			result = append(result, a)
		}
	}
	return result, nil
}

// UpdateAlertStatus updates the status of a fraud alert
func (s *FraudDetectionService) UpdateAlertStatus(_ context.Context, alertID, status string, actions []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, a := range s.alerts {
		if a.ID == alertID {
			a.Status = status
			a.Actions = append(a.Actions, actions...)
			return nil
		}
	}
	return fmt.Errorf("alert not found: %s", alertID)
}

// GetFraudMetrics returns fraud detection metrics
func (s *FraudDetectionService) GetFraudMetrics(_ context.Context, period string) (*FraudMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	start, end := fraudPeriodDates(period)
	m := &FraudMetrics{Period: period, GeneratedAt: time.Now(), ByType: make(map[FraudType]int64), BySeverity: make(map[FraudSeverity]int64)}

	for _, a := range s.alerts {
		if a.Timestamp.After(start) && a.Timestamp.Before(end) {
			m.TotalAlerts++
			m.ByType[a.Type]++
			m.BySeverity[a.Severity]++
			switch a.Status {
			case "resolved":
				m.ResolvedAlerts++
			case "false_positive":
				m.FalsePositives++
			}
		}
	}

	if m.TotalAlerts > 0 {
		m.ResolutionRate = float64(m.ResolvedAlerts) / float64(m.TotalAlerts) * 100
		m.FalsePositiveRate = float64(m.FalsePositives) / float64(m.TotalAlerts) * 100
	}
	return m, nil
}

func (s *FraudDetectionService) evaluatePatterns(ctx context.Context, tx map[string]interface{}) (float64, []string) {
	var score float64
	var evidence []string

	profileID, _ := tx["profile_id"].(string)
	if profileID == "" {
		return 0, nil
	}

	for _, p := range s.patterns {
		if !p.Enabled {
			continue
		}
		ps, ev := s.checkPattern(ctx, profileID, p)
		if ps > 0 {
			score += ps * p.Weight
			evidence = append(evidence, ev...)
		}
	}
	return score, evidence
}

func (s *FraudDetectionService) checkPattern(ctx context.Context, profileID string, p FraudPattern) (float64, []string) {
	var count int64

	switch p.ID {
	case "multiple_subs":
		s.db.WithContext(ctx).Table("rate_plan_subscriptions").Where("profile_id = ? AND status = ?", profileID, "active").Count(&count)
		if count > int64(p.Threshold) {
			return float64(count) * 20, []string{fmt.Sprintf("%d active subscriptions", count)}
		}
	case "rapid_sub":
		s.db.WithContext(ctx).Table("rate_plan_subscriptions").Where("profile_id = ? AND created_at > ?", profileID, time.Now().Add(-time.Hour)).Count(&count)
		if count > int64(p.Threshold) {
			return float64(count) * 15, []string{fmt.Sprintf("%d subscriptions in last hour", count)}
		}
	case "payment_fail":
		s.db.WithContext(ctx).Table("billing_transactions").Where("profile_id = ? AND status = ? AND created_at > ?", profileID, "failed", time.Now().Add(-24*time.Hour)).Count(&count)
		if count > int64(p.Threshold) {
			return float64(count) * 25, []string{fmt.Sprintf("%d payment failures", count)}
		}
	case "sim_swap":
		s.db.WithContext(ctx).Table("profiles").Where("id = ? AND updated_at > ?", profileID, time.Now().Add(-time.Hour)).Count(&count)
		if count > 2 {
			return 80, []string{"possible SIM swap detected"}
		}
	}
	return 0, nil
}

func (s *FraudDetectionService) detectType(evidence []string) FraudType {
	for _, e := range evidence {
		if contains(e, "subscription") {
			return FraudTypeSubscriptionFraud
		}
		if contains(e, "payment") {
			return FraudTypePaymentFraud
		}
		if contains(e, "SIM") {
			return FraudTypeSIMSwap
		}
	}
	return FraudTypeSubscriptionFraud
}

func (s *FraudDetectionService) blockProfile(ctx context.Context, profileID string) {
	s.db.WithContext(ctx).Table("profiles").Where("id = ?", profileID).Updates(map[string]interface{}{"status": "blocked", "blocked_at": time.Now()})
	s.logger.WithField("profile_id", profileID).Warn("Profile blocked for fraud")
}

func (s *FraudDetectionService) matchesFilter(a *FraudAlert, f FraudAlertFilter) bool {
	if f.Type != "" && a.Type != f.Type {
		return false
	}
	if f.Severity != "" && a.Severity != f.Severity {
		return false
	}
	if f.Status != "" && a.Status != f.Status {
		return false
	}
	if f.FromDate != nil && a.Timestamp.Before(*f.FromDate) {
		return false
	}
	if f.ToDate != nil && a.Timestamp.After(*f.ToDate) {
		return false
	}
	return true
}

func (s *FraudDetectionService) cleanupAlerts(days int) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		cutoff := time.Now().AddDate(0, 0, -days)
		var filtered []*FraudAlert
		for _, a := range s.alerts {
			if a.Timestamp.After(cutoff) {
				filtered = append(filtered, a)
			}
		}
		s.alerts = filtered
		s.mu.Unlock()
	}
}

func fraudPeriodDates(period string) (time.Time, time.Time) {
	now := time.Now()
	switch period {
	case "daily":
		return now.Truncate(24 * time.Hour), now
	case "weekly":
		return now.AddDate(0, 0, -7), now
	case "monthly":
		return now.AddDate(0, -1, 0), now
	default:
		return now.AddDate(0, -1, 0), now
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
