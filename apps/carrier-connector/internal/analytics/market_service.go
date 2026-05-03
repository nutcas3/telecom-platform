package analytics

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// MarketMetrics represents market penetration analysis
type MarketMetrics struct {
	Period                   string                          `json:"period"`
	TotalMarketSize          int64                           `json:"total_market_size"`
	OurSubscribers           int64                           `json:"our_subscribers"`
	MarketShare              float64                         `json:"market_share_pct"`
	GrowthRate               float64                         `json:"growth_rate_pct"`
	ByCountry                map[string]CountryMetrics       `json:"by_country"`
	ByCarrier                map[string]MarketCarrierMetrics `json:"by_carrier"`
	ByDemographic            map[string]DemoMetrics          `json:"by_demographic"`
	CompetitorAnalysis       []CompetitorMetrics             `json:"competitor_analysis"`
	PenetrationOpportunities []OpportunityMetrics            `json:"penetration_opportunities"`
	GeneratedAt              time.Time                       `json:"generated_at"`
}

// CountryMetrics represents metrics by country
type CountryMetrics struct {
	Country           string  `json:"country"`
	TotalPopulation   int64   `json:"total_population"`
	ActiveSubscribers int64   `json:"active_subscribers"`
	PenetrationRate   float64 `json:"penetration_rate_pct"`
	GrowthRate        float64 `json:"growth_rate_pct"`
	ARPU              float64 `json:"arpu"` // Average Revenue Per User
	MarketPotential   int64   `json:"market_potential"`
}

// MarketCarrierMetrics represents metrics by carrier for market analysis
type MarketCarrierMetrics struct {
	CarrierName    string  `json:"carrier_name"`
	Subscribers    int64   `json:"subscribers"`
	MarketShare    float64 `json:"market_share_pct"`
	ChurnRate      float64 `json:"churn_rate_pct"`
	ARPU           float64 `json:"arpu"`
	NetworkQuality float64 `json:"network_quality_score"`
	Coverage       float64 `json:"coverage_pct"`
}

// DemoMetrics represents metrics by demographic
type DemoMetrics struct {
	Segment     string  `json:"segment"`
	Subscribers int64   `json:"subscribers"`
	MarketShare float64 `json:"market_share_pct"`
	ARPU        float64 `json:"arpu"`
	GrowthRate  float64 `json:"growth_rate_pct"`
}

// CompetitorMetrics represents competitor analysis
type CompetitorMetrics struct {
	Name          string   `json:"name"`
	EstimatedSubs int64    `json:"estimated_subscribers"`
	MarketShare   float64  `json:"market_share_pct"`
	Strengths     []string `json:"strengths"`
	Weaknesses    []string `json:"weaknesses"`
	ThreatLevel   string   `json:"threat_level"`
}

// OpportunityMetrics represents market opportunities
type OpportunityMetrics struct {
	Country            string  `json:"country"`
	OpportunityType    string  `json:"opportunity_type"`
	PotentialSubs      int64   `json:"potential_subscribers"`
	RequiredInvestment float64 `json:"required_investment"`
	ExpectedROI        float64 `json:"expected_roi_pct"`
	TimeToMarket       int     `json:"time_to_market_months"`
}

// MarketAnalysisService provides market penetration analysis
type MarketAnalysisService struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewMarketAnalysisService creates a new market analysis service
func NewMarketAnalysisService(db *gorm.DB, logger *logrus.Logger) *MarketAnalysisService {
	return &MarketAnalysisService{
		db:     db,
		logger: logger,
	}
}

// GetMarketMetrics calculates market penetration metrics
func (s *MarketAnalysisService) GetMarketMetrics(ctx context.Context, period string) (*MarketMetrics, error) {
	metrics := &MarketMetrics{
		Period:                   period,
		ByCountry:                make(map[string]CountryMetrics),
		ByCarrier:                make(map[string]MarketCarrierMetrics),
		ByDemographic:            make(map[string]DemoMetrics),
		CompetitorAnalysis:       make([]CompetitorMetrics, 0),
		PenetrationOpportunities: make([]OpportunityMetrics, 0),
		GeneratedAt:              time.Now(),
	}

	// Calculate overall market metrics
	s.calculateOverallMetrics(ctx, metrics)

	// Calculate country-specific metrics
	s.calculateCountryMetrics(ctx, metrics)

	// Calculate carrier metrics
	s.calculateCarrierMetrics(ctx, metrics)

	// Calculate demographic metrics
	s.calculateDemographicMetrics(ctx, metrics)

	// Analyze competitors
	s.analyzeCompetitors(ctx, metrics)

	// Identify opportunities
	s.identifyOpportunities(ctx, metrics)

	return metrics, nil
}

// calculateOverallMetrics calculates overall market metrics
func (s *MarketAnalysisService) calculateOverallMetrics(ctx context.Context, metrics *MarketMetrics) {
	// Get our subscriber count
	var ourSubs int64
	s.db.WithContext(ctx).Table("profiles").
		Where("status = ?", "active").
		Count(&ourSubs)
	metrics.OurSubscribers = ourSubs

	// Estimate total market size (simplified - would use external data)
	metrics.TotalMarketSize = s.estimateTotalMarketSize()

	// Calculate market share
	if metrics.TotalMarketSize > 0 {
		metrics.MarketShare = float64(ourSubs) / float64(metrics.TotalMarketSize) * 100
	}

	// Calculate growth rate (compare to previous period)
	metrics.GrowthRate = s.calculateGrowthRate(ctx, ourSubs)
}

// calculateCountryMetrics calculates metrics by country
func (s *MarketAnalysisService) calculateCountryMetrics(ctx context.Context, metrics *MarketMetrics) {
	countries := []string{"US", "UK", "DE", "FR", "JP", "AU", "CA", "SG"}

	for _, country := range countries {
		var subs int64
		s.db.WithContext(ctx).Table("profiles").
			Where("country = ? AND status = ?", country, "active").
			Count(&subs)

		countryMetrics := CountryMetrics{
			Country:           country,
			ActiveSubscribers: subs,
			TotalPopulation:   s.getCountryPopulation(country),
			PenetrationRate:   s.calculatePenetrationRate(subs, country),
			GrowthRate:        s.calculateCountryGrowthRate(ctx, country),
			ARPU:              s.calculateCountryARPU(ctx, country),
			MarketPotential:   s.calculateMarketPotential(country),
		}

		metrics.ByCountry[country] = countryMetrics
	}
}

// calculateCarrierMetrics calculates metrics by carrier
func (s *MarketAnalysisService) calculateCarrierMetrics(ctx context.Context, metrics *MarketMetrics) {
	// This would analyze carrier partnerships and performance
	carriers := []string{"AT&T", "Verizon", "T-Mobile", "Vodafone", "Orange", "Deutsche Telekom"}

	for _, carrier := range carriers {
		var subs int64
		s.db.WithContext(ctx).Table("rate_plan_subscriptions rps").
			Joins("JOIN rate_plans rp ON rps.rate_plan_id = rp.id").
			Joins("JOIN profiles p ON rps.profile_id = p.id").
			Where("rp.carrier_name = ? AND rps.status = ?", carrier, "active").
			Count(&subs)

		carrierMetrics := MarketCarrierMetrics{
			CarrierName:    carrier,
			Subscribers:    subs,
			MarketShare:    s.calculateCarrierMarketShare(ctx, carrier),
			ChurnRate:      s.calculateCarrierChurnRate(ctx, carrier),
			ARPU:           s.calculateCarrierARPU(ctx, carrier),
			NetworkQuality: s.getCarrierNetworkQuality(carrier),
			Coverage:       s.getCarrierCoverage(carrier),
		}

		metrics.ByCarrier[carrier] = carrierMetrics
	}
}

// calculateDemographicMetrics calculates metrics by demographic segments
func (s *MarketAnalysisService) calculateDemographicMetrics(ctx context.Context, metrics *MarketMetrics) {
	segments := []string{"18-24", "25-34", "35-44", "45-54", "55+", "Business", "Student"}

	for _, segment := range segments {
		var subs int64
		s.db.WithContext(ctx).Table("profiles").
			Where("demographic_segment = ? AND status = ?", segment, "active").
			Count(&subs)

		demoMetrics := DemoMetrics{
			Segment:     segment,
			Subscribers: subs,
			MarketShare: s.calculateDemoMarketShare(ctx, segment),
			ARPU:        s.calculateDemoARPU(ctx, segment),
			GrowthRate:  s.calculateDemoGrowthRate(ctx, segment),
		}

		metrics.ByDemographic[segment] = demoMetrics
	}
}

// analyzeCompetitors analyzes competitive landscape
func (s *MarketAnalysisService) analyzeCompetitors(ctx context.Context, metrics *MarketMetrics) {
	competitors := []struct {
		name       string
		subs       int64
		share      float64
		strengths  []string
		weaknesses []string
		threat     string
	}{
		{
			name:       "Airalo",
			subs:       5000000,
			share:      25.0,
			strengths:  []string{"Large market share", "Global presence", "Brand recognition"},
			weaknesses: []string{"Higher prices", "Limited customization"},
			threat:     "high",
		},
		{
			name:       "Truphone",
			subs:       2000000,
			share:      10.0,
			strengths:  []string{"Enterprise focus", "Quality network"},
			weaknesses: []string{"Limited consumer market", "Higher pricing"},
			threat:     "medium",
		},
		{
			name:       "Ubigi",
			subs:       1500000,
			share:      7.5,
			strengths:  []string{"Competitive pricing", "Good coverage"},
			weaknesses: []string{"Smaller market share", "Limited features"},
			threat:     "medium",
		},
		{
			name:       "Holafly",
			subs:       1000000,
			share:      5.0,
			strengths:  []string{"Simple pricing", "Good customer service"},
			weaknesses: []string{"Limited carrier partnerships", "Basic features"},
			threat:     "low",
		},
	}

	for _, comp := range competitors {
		metrics.CompetitorAnalysis = append(metrics.CompetitorAnalysis, CompetitorMetrics{
			Name:          comp.name,
			EstimatedSubs: comp.subs,
			MarketShare:   comp.share,
			Strengths:     comp.strengths,
			Weaknesses:    comp.weaknesses,
			ThreatLevel:   comp.threat,
		})
	}
}

// identifyOpportunities identifies market penetration opportunities
func (s *MarketAnalysisService) identifyOpportunities(ctx context.Context, metrics *MarketMetrics) {
	opportunities := []OpportunityMetrics{
		{
			Country:            "India",
			OpportunityType:    "Emerging Market",
			PotentialSubs:      10000000,
			RequiredInvestment: 5000000,
			ExpectedROI:        150.0,
			TimeToMarket:       6,
		},
		{
			Country:            "Brazil",
			OpportunityType:    "Latin America Expansion",
			PotentialSubs:      5000000,
			RequiredInvestment: 3000000,
			ExpectedROI:        120.0,
			TimeToMarket:       4,
		},
		{
			Country:            "South Korea",
			OpportunityType:    "Tech-Savvy Market",
			PotentialSubs:      2000000,
			RequiredInvestment: 2000000,
			ExpectedROI:        80.0,
			TimeToMarket:       3,
		},
		{
			Country:            "Nigeria",
			OpportunityType:    "African Market Entry",
			PotentialSubs:      8000000,
			RequiredInvestment: 4000000,
			ExpectedROI:        180.0,
			TimeToMarket:       8,
		},
	}

	metrics.PenetrationOpportunities = opportunities
}

// estimateTotalMarketSize estimates total addressable market
func (s *MarketAnalysisService) estimateTotalMarketSize() int64 {
	// Simplified estimation - would use market research data
	// Global mobile subscribers ~5.5B, eSIM adoption ~10%
	return 550000000 // 550M eSIM users globally
}

// calculateGrowthRate calculates growth rate compared to previous period
func (s *MarketAnalysisService) calculateGrowthRate(ctx context.Context, currentSubs int64) float64 {
	// Get subscribers from previous month
	var prevSubs int64
	s.db.WithContext(ctx).Table("profiles").
		Where("status = ? AND created_at < ?", "active", time.Now().AddDate(0, -1, 0)).
		Count(&prevSubs)

	if prevSubs == 0 {
		return 0
	}

	return float64(currentSubs-prevSubs) / float64(prevSubs) * 100
}

// getCountryPopulation returns population data for a country
func (s *MarketAnalysisService) getCountryPopulation(country string) int64 {
	populations := map[string]int64{
		"US": 331000000,
		"UK": 67000000,
		"DE": 83000000,
		"FR": 65000000,
		"JP": 126000000,
		"AU": 25000000,
		"CA": 38000000,
		"SG": 5800000,
	}
	return populations[country]
}

// calculatePenetrationRate calculates penetration rate for a country
func (s *MarketAnalysisService) calculatePenetrationRate(subs int64, country string) float64 {
	population := s.getCountryPopulation(country)
	if population == 0 {
		return 0
	}
	return float64(subs) / float64(population) * 100
}

// calculateCountryGrowthRate calculates growth rate for a country
func (s *MarketAnalysisService) calculateCountryGrowthRate(ctx context.Context, country string) float64 {
	var currentSubs, prevSubs int64

	now := time.Now()
	prevMonth := now.AddDate(0, -1, 0)

	s.db.WithContext(ctx).Table("profiles").
		Where("country = ? AND status = ?", country, "active").
		Count(&currentSubs)

	s.db.WithContext(ctx).Table("profiles").
		Where("country = ? AND status = ? AND created_at < ?", country, "active", prevMonth).
		Count(&prevSubs)

	if prevSubs == 0 {
		return 0
	}

	return float64(currentSubs-prevSubs) / float64(prevSubs) * 100
}

// calculateCountryARPU calculates ARPU for a country
func (s *MarketAnalysisService) calculateCountryARPU(ctx context.Context, country string) float64 {
	var totalRevenue float64
	var subscriberCount int64

	s.db.WithContext(ctx).Table("billing_transactions bt").
		Joins("JOIN profiles p ON bt.profile_id = p.id").
		Where("p.country = ? AND bt.status = ?", country, "completed").
		Select("COALESCE(SUM(bt.amount), 0)").
		Scan(&totalRevenue)

	s.db.WithContext(ctx).Table("profiles").
		Where("country = ? AND status = ?", country, "active").
		Count(&subscriberCount)

	if subscriberCount == 0 {
		return 0
	}

	return totalRevenue / float64(subscriberCount)
}

// calculateMarketPotential calculates market potential for a country
func (s *MarketAnalysisService) calculateMarketPotential(country string) int64 {
	population := s.getCountryPopulation(country)
	// Assume 20% of population could adopt eSIM in next 2 years
	return int64(float64(population) * 0.2)
}

// calculateCarrierMarketShare calculates market share for a carrier
func (s *MarketAnalysisService) calculateCarrierMarketShare(ctx context.Context, carrier string) float64 {
	var carrierSubs, totalSubs int64

	s.db.WithContext(ctx).Table("rate_plan_subscriptions rps").
		Joins("JOIN rate_plans rp ON rps.rate_plan_id = rp.id").
		Where("rp.carrier_name = ? AND rps.status = ?", carrier, "active").
		Count(&carrierSubs)

	s.db.WithContext(ctx).Table("rate_plan_subscriptions").
		Where("status = ?", "active").
		Count(&totalSubs)

	if totalSubs == 0 {
		return 0
	}

	return float64(carrierSubs) / float64(totalSubs) * 100
}

// calculateCarrierChurnRate calculates churn rate for a carrier
func (s *MarketAnalysisService) calculateCarrierChurnRate(ctx context.Context, carrier string) float64 {
	var totalSubs, churnedSubs int64

	s.db.WithContext(ctx).Table("rate_plan_subscriptions rps").
		Joins("JOIN rate_plans rp ON rps.rate_plan_id = rp.id").
		Where("rp.carrier_name = ?", carrier).
		Count(&totalSubs)

	s.db.WithContext(ctx).Table("rate_plan_subscriptions rps").
		Joins("JOIN rate_plans rp ON rps.rate_plan_id = rp.id").
		Where("rp.carrier_name = ? AND rps.status = ? AND rps.ended_at > ?",
			carrier, "cancelled", time.Now().AddDate(0, -1, 0)).
		Count(&churnedSubs)

	if totalSubs == 0 {
		return 0
	}

	return float64(churnedSubs) / float64(totalSubs) * 100
}

// calculateCarrierARPU calculates ARPU for a carrier
func (s *MarketAnalysisService) calculateCarrierARPU(ctx context.Context, carrier string) float64 {
	var totalRevenue float64
	var subscriberCount int64

	s.db.WithContext(ctx).Table("billing_transactions bt").
		Joins("JOIN rate_plan_subscriptions rps ON bt.subscription_id = rps.id").
		Joins("JOIN rate_plans rp ON rps.rate_plan_id = rp.id").
		Where("rp.carrier_name = ? AND bt.status = ?", carrier, "completed").
		Select("COALESCE(SUM(bt.amount), 0)").
		Scan(&totalRevenue)

	s.db.WithContext(ctx).Table("rate_plan_subscriptions rps").
		Joins("JOIN rate_plans rp ON rps.rate_plan_id = rp.id").
		Where("rp.carrier_name = ? AND rps.status = ?", carrier, "active").
		Count(&subscriberCount)

	if subscriberCount == 0 {
		return 0
	}

	return totalRevenue / float64(subscriberCount)
}

// getCarrierNetworkQuality returns network quality score for carrier
func (s *MarketAnalysisService) getCarrierNetworkQuality(carrier string) float64 {
	// Simplified network quality scores
	quality := map[string]float64{
		"AT&T":             85.0,
		"Verizon":          87.0,
		"T-Mobile":         83.0,
		"Vodafone":         80.0,
		"Orange":           78.0,
		"Deutsche Telekom": 82.0,
	}
	return quality[carrier]
}

// getCarrierCoverage returns coverage percentage for carrier
func (s *MarketAnalysisService) getCarrierCoverage(carrier string) float64 {
	// Simplified coverage percentages
	coverage := map[string]float64{
		"AT&T":             95.0,
		"Verizon":          96.0,
		"T-Mobile":         94.0,
		"Vodafone":         85.0,
		"Orange":           88.0,
		"Deutsche Telekom": 92.0,
	}
	return coverage[carrier]
}

// calculateDemoMarketShare calculates market share for demographic segment
func (s *MarketAnalysisService) calculateDemoMarketShare(ctx context.Context, segment string) float64 {
	var segmentSubs, totalSubs int64

	s.db.WithContext(ctx).Table("profiles").
		Where("demographic_segment = ? AND status = ?", segment, "active").
		Count(&segmentSubs)

	s.db.WithContext(ctx).Table("profiles").
		Where("status = ?", "active").
		Count(&totalSubs)

	if totalSubs == 0 {
		return 0
	}

	return float64(segmentSubs) / float64(totalSubs) * 100
}

// calculateDemoARPU calculates ARPU for demographic segment
func (s *MarketAnalysisService) calculateDemoARPU(ctx context.Context, segment string) float64 {
	var totalRevenue float64
	var subscriberCount int64

	s.db.WithContext(ctx).Table("billing_transactions bt").
		Joins("JOIN profiles p ON bt.profile_id = p.id").
		Where("p.demographic_segment = ? AND bt.status = ?", segment, "completed").
		Select("COALESCE(SUM(bt.amount), 0)").
		Scan(&totalRevenue)

	s.db.WithContext(ctx).Table("profiles").
		Where("demographic_segment = ? AND status = ?", segment, "active").
		Count(&subscriberCount)

	if subscriberCount == 0 {
		return 0
	}

	return totalRevenue / float64(subscriberCount)
}

// calculateDemoGrowthRate calculates growth rate for demographic segment
func (s *MarketAnalysisService) calculateDemoGrowthRate(ctx context.Context, segment string) float64 {
	var currentSubs, prevSubs int64

	now := time.Now()
	prevMonth := now.AddDate(0, -1, 0)

	s.db.WithContext(ctx).Table("profiles").
		Where("demographic_segment = ? AND status = ?", segment, "active").
		Count(&currentSubs)

	s.db.WithContext(ctx).Table("profiles").
		Where("demographic_segment = ? AND status = ? AND created_at < ?", segment, "active", prevMonth).
		Count(&prevSubs)

	if prevSubs == 0 {
		return 0
	}

	return float64(currentSubs-prevSubs) / float64(prevSubs) * 100
}

// GetMarketTrends returns market penetration trends
func (s *MarketAnalysisService) GetMarketTrends(ctx context.Context, period string) ([]map[string]interface{}, error) {
	trends := make([]map[string]interface{}, 0)

	// Generate trend data for the last 12 months
	for i := 11; i >= 0; i-- {
		date := time.Now().AddDate(0, -i, 0)

		var subs int64
		s.db.WithContext(ctx).Table("profiles").
			Where("status = ? AND created_at <= ?", "active", date).
			Count(&subs)

		trend := map[string]interface{}{
			"month":        date.Format("2006-01"),
			"subscribers":  subs,
			"market_share": float64(subs) / 550000000 * 100, // vs total market
		}

		trends = append(trends, trend)
	}

	return trends, nil
}
