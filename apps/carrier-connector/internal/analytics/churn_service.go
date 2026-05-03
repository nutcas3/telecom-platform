package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ChurnAnalysisService provides churn analysis and prediction
type ChurnAnalysisService struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewChurnAnalysisService creates a new churn analysis service
func NewChurnAnalysisService(db *gorm.DB, logger *logrus.Logger) *ChurnAnalysisService {
	return &ChurnAnalysisService{db: db, logger: logger}
}

// PredictChurn predicts churn risk for a specific profile
func (s *ChurnAnalysisService) PredictChurn(ctx context.Context, profileID string) (*ChurnPrediction, error) {
	factors := s.analyzeFactors(ctx, profileID)
	score := s.calculateScore(factors)

	prediction := &ChurnPrediction{
		ProfileID:       profileID,
		RiskScore:       score,
		RiskLevel:       RiskLevelFromScore(score),
		Reasons:         s.generateReasons(factors),
		Recommendations: RecommendationsForRisk(RiskLevelFromScore(score)),
		LastUpdated:     time.Now(),
	}

	if prediction.RiskLevel == ChurnRiskHigh || prediction.RiskLevel == ChurnRiskCritical {
		date := time.Now().AddDate(0, 0, int((100-score)*3))
		prediction.PredictedChurnDate = &date
	}

	return prediction, nil
}

// GetChurnMetrics calculates overall churn metrics
func (s *ChurnAnalysisService) GetChurnMetrics(ctx context.Context, period string) (*ChurnMetrics, error) {
	start, end := periodDates(period)
	metrics := &ChurnMetrics{Period: period, RiskDistribution: make(map[ChurnRiskLevel]int64), GeneratedAt: time.Now()}

	s.db.WithContext(ctx).Table("profiles").Where("created_at < ?", start).Count(&metrics.TotalSubscribers)
	s.db.WithContext(ctx).Table("rate_plan_subscriptions").Where("ended_at BETWEEN ? AND ?", start, end).Count(&metrics.ChurnedSubscribers)

	if metrics.TotalSubscribers > 0 {
		metrics.ChurnRate = float64(metrics.ChurnedSubscribers) / float64(metrics.TotalSubscribers) * 100
		metrics.MonthlyChurnRate = metrics.ChurnRate
		metrics.AnnualChurnRate = metrics.ChurnRate * 12
	}

	s.db.WithContext(ctx).Table("rate_plan_subscriptions").Where("status = ?", "cancelled").
		Select("AVG(EXTRACT(EPOCH FROM (ended_at - started_at))/86400)").Scan(&metrics.AverageTenure)

	// Estimated distribution
	metrics.RiskDistribution[ChurnRiskLow] = int64(float64(metrics.TotalSubscribers) * 0.6)
	metrics.RiskDistribution[ChurnRiskMedium] = int64(float64(metrics.TotalSubscribers) * 0.25)
	metrics.RiskDistribution[ChurnRiskHigh] = int64(float64(metrics.TotalSubscribers) * 0.12)
	metrics.RiskDistribution[ChurnRiskCritical] = int64(float64(metrics.TotalSubscribers) * 0.03)

	return metrics, nil
}

// GetChurnFactors returns the top factors contributing to churn
func (s *ChurnAnalysisService) GetChurnFactors(_ context.Context) ([]ChurnFactor, error) {
	return DefaultChurnFactors(), nil
}

// GetAtRiskCustomers returns customers at high risk of churn
func (s *ChurnAnalysisService) GetAtRiskCustomers(ctx context.Context, riskLevel ChurnRiskLevel, limit int) ([]*ChurnPrediction, error) {
	var subs []struct{ ID string }
	s.db.WithContext(ctx).Table("profiles").Where("status = ?", "active").Limit(limit).Find(&subs)

	var predictions []*ChurnPrediction
	for _, sub := range subs {
		pred, err := s.PredictChurn(ctx, sub.ID)
		if err != nil {
			continue
		}
		if pred.RiskLevel == riskLevel || (riskLevel == ChurnRiskHigh && pred.RiskLevel == ChurnRiskCritical) {
			predictions = append(predictions, pred)
		}
	}
	return predictions, nil
}

func (s *ChurnAnalysisService) analyzeFactors(ctx context.Context, profileID string) []ChurnFactor {
	factors := DefaultChurnFactors()
	result := make([]ChurnFactor, 0, len(factors))

	for _, f := range factors {
		impact := s.factorImpact(ctx, profileID, f.Factor)
		if impact > 0.1 {
			result = append(result, ChurnFactor{Factor: f.Factor, Impact: impact, Description: f.Description, Weight: f.Weight})
		}
	}
	return result
}

func (s *ChurnAnalysisService) factorImpact(ctx context.Context, profileID, factor string) float64 {
	switch factor {
	case "Low Usage":
		return s.usageImpact(ctx, profileID)
	case "Payment Issues":
		return s.paymentImpact(ctx, profileID)
	case "Price Sensitivity":
		return s.priceImpact(ctx, profileID)
	default:
		return 0.4
	}
}

func (s *ChurnAnalysisService) usageImpact(ctx context.Context, profileID string) float64 {
	var total int64
	s.db.WithContext(ctx).Table("rate_plan_usage").Where("profile_id = ? AND created_at > ?", profileID, time.Now().AddDate(0, -1, 0)).
		Select("COALESCE(SUM(data_used + voice_used + sms_used), 0)").Scan(&total)

	switch {
	case total < 100:
		return 0.8
	case total < 500:
		return 0.6
	case total < 1000:
		return 0.3
	default:
		return 0.1
	}
}

func (s *ChurnAnalysisService) paymentImpact(ctx context.Context, profileID string) float64 {
	var createdAt time.Time
	s.db.WithContext(ctx).Table("profiles").Where("id = ?", profileID).Select("created_at").Scan(&createdAt)
	days := time.Since(createdAt).Hours() / 24
	if days < 30 {
		return 0.3
	} else if days < 90 {
		return 0.2
	}
	return 0.1
}

func (s *ChurnAnalysisService) priceImpact(ctx context.Context, profileID string) float64 {
	var price float64
	s.db.WithContext(ctx).Table("rate_plans rp").Joins("JOIN rate_plan_subscriptions rps ON rps.rate_plan_id = rp.id").
		Where("rps.profile_id = ? AND rps.status = ?", profileID, "active").Select("rp.base_price").Scan(&price)

	switch {
	case price > 50:
		return 0.6
	case price > 30:
		return 0.4
	case price > 15:
		return 0.2
	default:
		return 0.1
	}
}

func (s *ChurnAnalysisService) calculateScore(factors []ChurnFactor) float64 {
	var total, weight float64
	for _, f := range factors {
		total += f.Impact * f.Weight
		weight += f.Weight
	}
	if weight == 0 {
		return 0
	}
	return (total / weight) * 100
}

func (s *ChurnAnalysisService) generateReasons(factors []ChurnFactor) []string {
	var reasons []string
	for _, f := range factors {
		if f.Impact > 0.7 {
			reasons = append(reasons, fmt.Sprintf("High %s detected", f.Factor))
		} else if f.Impact > 0.5 {
			reasons = append(reasons, fmt.Sprintf("Moderate %s detected", f.Factor))
		}
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "Multiple minor risk factors detected")
	}
	return reasons
}

func periodDates(period string) (time.Time, time.Time) {
	now := time.Now()
	switch period {
	case "daily":
		return now.Truncate(24 * time.Hour), now
	case "weekly":
		return now.AddDate(0, 0, -7), now
	case "monthly":
		return now.AddDate(0, -1, 0), now
	case "quarterly":
		return now.AddDate(0, -3, 0), now
	default:
		return now.AddDate(0, -1, 0), now
	}
}
