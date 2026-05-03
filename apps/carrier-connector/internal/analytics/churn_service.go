package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ChurnRiskLevel represents the risk level of customer churn
type ChurnRiskLevel string

const (
	ChurnRiskLow      ChurnRiskLevel = "low"
	ChurnRiskMedium   ChurnRiskLevel = "medium"
	ChurnRiskHigh     ChurnRiskLevel = "high"
	ChurnRiskCritical ChurnRiskLevel = "critical"
)

// ChurnPrediction represents a churn prediction for a customer
type ChurnPrediction struct {
	ProfileID          string         `json:"profile_id"`
	RiskLevel          ChurnRiskLevel `json:"risk_level"`
	RiskScore          float64        `json:"risk_score"` // 0-100
	PredictedChurnDate *time.Time     `json:"predicted_churn_date,omitempty"`
	Reasons            []string       `json:"reasons"`
	Recommendations    []string       `json:"recommendations"`
	LastUpdated        time.Time      `json:"last_updated"`
}

// ChurnMetrics represents churn analysis metrics
type ChurnMetrics struct {
	Period             string                   `json:"period"`
	TotalSubscribers   int64                    `json:"total_subscribers"`
	ChurnedSubscribers int64                    `json:"churned_subscribers"`
	ChurnRate          float64                  `json:"churn_rate"`
	MonthlyChurnRate   float64                  `json:"monthly_churn_rate"`
	AnnualChurnRate    float64                  `json:"annual_churn_rate"`
	AverageTenure      float64                  `json:"average_tenure_days"`
	RiskDistribution   map[ChurnRiskLevel]int64 `json:"risk_distribution"`
	GeneratedAt        time.Time                `json:"generated_at"`
}

// ChurnFactor represents factors contributing to churn
type ChurnFactor struct {
	Factor      string  `json:"factor"`
	Impact      float64 `json:"impact"` // 0-1
	Description string  `json:"description"`
	Weight      float64 `json:"weight"`
}

// ChurnAnalysisService provides churn analysis and prediction
type ChurnAnalysisService struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewChurnAnalysisService creates a new churn analysis service
func NewChurnAnalysisService(db *gorm.DB, logger *logrus.Logger) *ChurnAnalysisService {
	return &ChurnAnalysisService{
		db:     db,
		logger: logger,
	}
}

// PredictChurn predicts churn risk for a specific profile
func (s *ChurnAnalysisService) PredictChurn(ctx context.Context, profileID string) (*ChurnPrediction, error) {
	// Get profile activity and usage data
	prediction := &ChurnPrediction{
		ProfileID:   profileID,
		LastUpdated: time.Now(),
	}

	// Analyze various churn factors
	factors := s.analyzeChurnFactors(ctx, profileID)

	// Calculate churn risk score
	riskScore := s.calculateChurnScore(factors)
	prediction.RiskScore = riskScore

	// Determine risk level
	prediction.RiskLevel = s.determineRiskLevel(riskScore)

	// Generate reasons and recommendations
	prediction.Reasons = s.generateChurnReasons(factors)
	prediction.Recommendations = s.generateRecommendations(prediction.RiskLevel, factors)

	// Predict churn date if high risk
	if prediction.RiskLevel == ChurnRiskHigh || prediction.RiskLevel == ChurnRiskCritical {
		predictedDate := s.predictChurnDate(riskScore)
		prediction.PredictedChurnDate = &predictedDate
	}

	return prediction, nil
}

// GetChurnMetrics calculates overall churn metrics
func (s *ChurnAnalysisService) GetChurnMetrics(ctx context.Context, period string) (*ChurnMetrics, error) {
	metrics := &ChurnMetrics{
		Period:           period,
		RiskDistribution: make(map[ChurnRiskLevel]int64),
		GeneratedAt:      time.Now(),
	}

	// Calculate churn rates based on period
	startDate, endDate := s.getPeriodDates(period)

	// Get total subscribers at start of period
	var totalSubs int64
	s.db.WithContext(ctx).Table("profiles").
		Where("created_at < ?", startDate).
		Count(&totalSubs)
	metrics.TotalSubscribers = totalSubs

	// Get churned subscribers (cancelled subscriptions)
	var churnedSubs int64
	s.db.WithContext(ctx).Table("rate_plan_subscriptions").
		Where("ended_at BETWEEN ? AND ?", startDate, endDate).
		Count(&churnedSubs)
	metrics.ChurnedSubscribers = churnedSubs

	// Calculate churn rates
	if totalSubs > 0 {
		metrics.ChurnRate = float64(churnedSubs) / float64(totalSubs) * 100
		metrics.MonthlyChurnRate = metrics.ChurnRate
		metrics.AnnualChurnRate = metrics.ChurnRate * 12
	}

	// Calculate average tenure
	var avgTenure float64
	s.db.WithContext(ctx).Table("rate_plan_subscriptions").
		Where("status = ?", "cancelled").
		Select("AVG(EXTRACT(EPOCH FROM (ended_at - started_at))/86400)").
		Scan(&avgTenure)
	metrics.AverageTenure = avgTenure

	// Get risk distribution (simplified - would need actual predictions in production)
	metrics.RiskDistribution[ChurnRiskLow] = int64(float64(totalSubs) * 0.6)
	metrics.RiskDistribution[ChurnRiskMedium] = int64(float64(totalSubs) * 0.25)
	metrics.RiskDistribution[ChurnRiskHigh] = int64(float64(totalSubs) * 0.12)
	metrics.RiskDistribution[ChurnRiskCritical] = int64(float64(totalSubs) * 0.03)

	return metrics, nil
}

// GetChurnFactors returns the top factors contributing to churn
func (s *ChurnAnalysisService) GetChurnFactors(ctx context.Context) ([]ChurnFactor, error) {
	factors := []ChurnFactor{
		{
			Factor:      "Low Usage",
			Impact:      0.85,
			Description: "Customers with low data/voice usage are more likely to churn",
			Weight:      0.25,
		},
		{
			Factor:      "Payment Issues",
			Impact:      0.92,
			Description: "Failed payments and billing disputes increase churn risk",
			Weight:      0.20,
		},
		{
			Factor:      "Poor Support Experience",
			Impact:      0.78,
			Description: "High support ticket resolution time correlates with churn",
			Weight:      0.15,
		},
		{
			Factor:      "Network Quality",
			Impact:      0.70,
			Description: "Dropped calls and slow data speeds impact retention",
			Weight:      0.20,
		},
		{
			Factor:      "Price Sensitivity",
			Impact:      0.65,
			Description: "Customers on higher-priced plans have higher churn rates",
			Weight:      0.10,
		},
		{
			Factor:      "Competitor Offers",
			Impact:      0.60,
			Description: "Better deals from competitors increase churn likelihood",
			Weight:      0.10,
		},
	}

	return factors, nil
}

// GetAtRiskCustomers returns customers at high risk of churn
func (s *ChurnAnalysisService) GetAtRiskCustomers(ctx context.Context, riskLevel ChurnRiskLevel, limit int) ([]*ChurnPrediction, error) {
	// This would typically query a pre-computed predictions table
	// For now, we'll simulate by analyzing active subscribers

	var predictions []*ChurnPrediction

	// Get active subscribers
	var subscribers []struct {
		ID string
	}
	s.db.WithContext(ctx).Table("profiles").
		Where("status = ?", "active").
		Limit(limit).
		Find(&subscribers)

	for _, sub := range subscribers {
		prediction, err := s.PredictChurn(ctx, sub.ID)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to predict churn for subscriber", "profile_id", sub.ID)
			continue
		}

		if prediction.RiskLevel == riskLevel ||
			(riskLevel == ChurnRiskHigh && (prediction.RiskLevel == ChurnRiskHigh || prediction.RiskLevel == ChurnRiskCritical)) {
			predictions = append(predictions, prediction)
		}
	}

	return predictions, nil
}

// analyzeChurnFactors analyzes churn factors for a profile
func (s *ChurnAnalysisService) analyzeChurnFactors(ctx context.Context, profileID string) []ChurnFactor {
	factors, _ := s.GetChurnFactors(ctx)
	profileFactors := make([]ChurnFactor, 0)

	for _, factor := range factors {
		impact := s.calculateFactorImpact(ctx, profileID, factor)
		if impact > 0.1 { // Only include significant factors
			profileFactors = append(profileFactors, ChurnFactor{
				Factor:      factor.Factor,
				Impact:      impact,
				Description: factor.Description,
				Weight:      factor.Weight,
			})
		}
	}

	return profileFactors
}

// calculateChurnScore calculates overall churn risk score
func (s *ChurnAnalysisService) calculateChurnScore(factors []ChurnFactor) float64 {
	totalScore := 0.0
	totalWeight := 0.0

	for _, factor := range factors {
		totalScore += factor.Impact * factor.Weight
		totalWeight += factor.Weight
	}

	if totalWeight == 0 {
		return 0
	}

	return (totalScore / totalWeight) * 100
}

// determineRiskLevel determines risk level from score
func (s *ChurnAnalysisService) determineRiskLevel(score float64) ChurnRiskLevel {
	switch {
	case score >= 80:
		return ChurnRiskCritical
	case score >= 60:
		return ChurnRiskHigh
	case score >= 40:
		return ChurnRiskMedium
	default:
		return ChurnRiskLow
	}
}

// generateChurnReasons generates reasons for churn prediction
func (s *ChurnAnalysisService) generateChurnReasons(factors []ChurnFactor) []string {
	reasons := make([]string, 0)

	for _, factor := range factors {
		if factor.Impact > 0.7 {
			reasons = append(reasons, fmt.Sprintf("High %s detected", factor.Factor))
		} else if factor.Impact > 0.5 {
			reasons = append(reasons, fmt.Sprintf("Moderate %s detected", factor.Factor))
		}
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "Multiple minor risk factors detected")
	}

	return reasons
}

// generateRecommendations generates retention recommendations
func (s *ChurnAnalysisService) generateRecommendations(riskLevel ChurnRiskLevel, factors []ChurnFactor) []string {
	recommendations := make([]string, 0)

	switch riskLevel {
	case ChurnRiskCritical:
		recommendations = append(recommendations, "Immediate intervention required")
		recommendations = append(recommendations, "Offer retention discount or upgrade")
		recommendations = append(recommendations, "Schedule proactive support call")

	case ChurnRiskHigh:
		recommendations = append(recommendations, "Send personalized retention offer")
		recommendations = append(recommendations, "Review and address service issues")

	case ChurnRiskMedium:
		recommendations = append(recommendations, "Monitor usage patterns closely")
		recommendations = append(recommendations, "Send value-add content")

	case ChurnRiskLow:
		recommendations = append(recommendations, "Continue standard engagement")
	}

	// Add specific recommendations based on factors
	for _, factor := range factors {
		switch factor.Factor {
		case "Low Usage":
			recommendations = append(recommendations, "Offer data bonus or plan optimization")
		case "Payment Issues":
			recommendations = append(recommendations, "Review billing and offer payment flexibility")
		case "Poor Support Experience":
			recommendations = append(recommendations, "Assign dedicated support representative")
		case "Network Quality":
			recommendations = append(recommendations, "Investigate network issues in customer area")
		case "Price Sensitivity":
			recommendations = append(recommendations, "Evaluate plan pricing and discounts")
		}
	}

	return recommendations
}

// predictChurnDate predicts when customer might churn
func (s *ChurnAnalysisService) predictChurnDate(riskScore float64) time.Time {
	// Simple prediction: higher risk = sooner churn
	daysUntilChurn := int((100 - riskScore) * 3) // Scale: 0-300 days
	return time.Now().AddDate(0, 0, daysUntilChurn)
}

// calculateFactorImpact calculates the impact of a specific factor for a profile
func (s *ChurnAnalysisService) calculateFactorImpact(ctx context.Context, profileID string, factor ChurnFactor) float64 {
	// This would analyze actual profile data
	// For now, return simulated values based on factor type

	switch factor.Factor {
	case "Low Usage":
		return s.analyzeUsagePattern(ctx, profileID)
	case "Payment Issues":
		return s.analyzePaymentHistory(ctx, profileID)
	case "Poor Support Experience":
		return s.analyzeSupportInteractions(ctx, profileID)
	case "Network Quality":
		return s.analyzeNetworkQuality(ctx, profileID)
	case "Price Sensitivity":
		return s.analyzePriceSensitivity(ctx, profileID)
	case "Competitor Offers":
		return s.analyzeCompetitorThreat(ctx, profileID)
	default:
		return 0.5 // Default medium impact
	}
}

// analyzeUsagePattern analyzes usage patterns for churn risk
func (s *ChurnAnalysisService) analyzeUsagePattern(ctx context.Context, profileID string) float64 {
	// Get recent usage
	var usage struct {
		DataUsed  int64
		VoiceUsed int64
		SMSUsed   int64
	}

	s.db.WithContext(ctx).Table("rate_plan_usage").
		Where("profile_id = ? AND created_at > ?", profileID, time.Now().AddDate(0, -1, 0)).
		Select("COALESCE(SUM(data_used), 0), COALESCE(SUM(voice_used), 0), COALESCE(SUM(sms_used), 0)").
		Scan(&usage)

	// Low usage indicates higher churn risk
	totalUsage := usage.DataUsed + usage.VoiceUsed + usage.SMSUsed
	if totalUsage < 100 { // Very low usage
		return 0.8
	} else if totalUsage < 500 { // Low usage
		return 0.6
	} else if totalUsage < 1000 { // Moderate usage
		return 0.3
	} else { // High usage
		return 0.1
	}
}

// analyzePaymentHistory analyzes payment patterns
func (s *ChurnAnalysisService) analyzePaymentHistory(ctx context.Context, profileID string) float64 {
	// This would check for failed payments, late payments, etc.
	// For simulation, return a value based on profile age
	var createdAt time.Time
	s.db.WithContext(ctx).Table("profiles").
		Where("id = ?", profileID).
		Select("created_at").
		Scan(&createdAt)

	tenureDays := time.Since(createdAt).Hours() / 24
	if tenureDays < 30 {
		return 0.3 // New customers have payment setup risk
	} else if tenureDays < 90 {
		return 0.2
	} else {
		return 0.1
	}
}

// analyzeSupportInteractions analyzes support ticket patterns
func (s *ChurnAnalysisService) analyzeSupportInteractions(ctx context.Context, profileID string) float64 {
	// This would check support ticket volume and resolution times
	// For simulation, return a moderate risk
	return 0.4
}

// analyzeNetworkQuality analyzes network quality metrics
func (s *ChurnAnalysisService) analyzeNetworkQuality(ctx context.Context, profileID string) float64 {
	// This would check dropped calls, data speeds, etc.
	// For simulation, return a value based on location
	var country string
	s.db.WithContext(ctx).Table("profiles").
		Where("id = ?", profileID).
		Select("country").
		Scan(&country)

	// Simulate different network quality by country
	switch country {
	case "US":
		return 0.2
	case "UK":
		return 0.25
	case "DE":
		return 0.15
	default:
		return 0.4 // Emerging markets might have lower quality
	}
}

// analyzePriceSensitivity analyzes price sensitivity
func (s *ChurnAnalysisService) analyzePriceSensitivity(ctx context.Context, profileID string) float64 {
	// Get current plan price
	var basePrice float64
	s.db.WithContext(ctx).Table("rate_plans rp").
		Joins("JOIN rate_plan_subscriptions rps ON rps.rate_plan_id = rp.id").
		Where("rps.profile_id = ? AND rps.status = ?", profileID, "active").
		Select("rp.base_price").
		Scan(&basePrice)

	// Higher price plans have higher churn risk
	if basePrice > 50 {
		return 0.6
	} else if basePrice > 30 {
		return 0.4
	} else if basePrice > 15 {
		return 0.2
	} else {
		return 0.1
	}
}

// analyzeCompetitorThreat analyzes competitor threat level
func (s *ChurnAnalysisService) analyzeCompetitorThreat(ctx context.Context, profileID string) float64 {
	// This would analyze market conditions and competitor offers
	// For simulation, return a moderate threat level
	return 0.5
}

// getPeriodDates returns start and end dates for a period
func (s *ChurnAnalysisService) getPeriodDates(period string) (time.Time, time.Time) {
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
	case "yearly":
		return now.AddDate(-1, 0, 0), now
	default:
		return now.AddDate(0, -1, 0), now // Default to monthly
	}
}
