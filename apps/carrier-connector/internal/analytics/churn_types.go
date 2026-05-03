package analytics

import "time"

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
	RiskScore          float64        `json:"risk_score"`
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
	Impact      float64 `json:"impact"`
	Description string  `json:"description"`
	Weight      float64 `json:"weight"`
}

// DefaultChurnFactors returns standard churn factors
func DefaultChurnFactors() []ChurnFactor {
	return []ChurnFactor{
		{Factor: "Low Usage", Impact: 0.85, Description: "Low data/voice usage increases churn risk", Weight: 0.25},
		{Factor: "Payment Issues", Impact: 0.92, Description: "Failed payments increase churn risk", Weight: 0.20},
		{Factor: "Poor Support", Impact: 0.78, Description: "High support ticket resolution time", Weight: 0.15},
		{Factor: "Network Quality", Impact: 0.70, Description: "Dropped calls and slow data speeds", Weight: 0.20},
		{Factor: "Price Sensitivity", Impact: 0.65, Description: "Higher-priced plans have higher churn", Weight: 0.10},
		{Factor: "Competitor Offers", Impact: 0.60, Description: "Better competitor deals", Weight: 0.10},
	}
}

// RiskLevelFromScore determines risk level from score
func RiskLevelFromScore(score float64) ChurnRiskLevel {
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

// RecommendationsForRisk returns recommendations based on risk level
func RecommendationsForRisk(level ChurnRiskLevel) []string {
	switch level {
	case ChurnRiskCritical:
		return []string{"Immediate intervention required", "Offer retention discount", "Schedule proactive support call"}
	case ChurnRiskHigh:
		return []string{"Send personalized retention offer", "Review and address service issues"}
	case ChurnRiskMedium:
		return []string{"Monitor usage patterns closely", "Send value-add content"}
	default:
		return []string{"Continue standard engagement"}
	}
}
