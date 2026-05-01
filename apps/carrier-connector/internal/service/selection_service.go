package service

import (
	"context"
	"fmt"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/handler"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/smdp"
	"github.com/sirupsen/logrus"
)

// SelectionService provides high-level carrier selection operations
type SelectionService struct {
	manager *smdp.SMDPManager
	handler *handler.SelectionHandler
	logger  *logrus.Logger
}

// NewSelectionService creates a new selection service
func NewSelectionService(manager *smdp.SMDPManager) *SelectionService {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &SelectionService{
		manager: manager,
		handler: handler.NewSelectionHandler(manager),
		logger:  logger,
	}
}

// GetHandler returns the selection handler for API registration
func (s *SelectionService) GetHandler() *handler.SelectionHandler {
	return s.handler
}

// IntelligentCarrierSelection performs intelligent carrier selection with comprehensive criteria
func (s *SelectionService) IntelligentCarrierSelection(ctx context.Context, request *IntelligentSelectionRequest) (*IntelligentSelectionResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"region":           request.Region,
		"profile_type":     request.ProfileType,
		"urgency":          request.Urgency,
		"cost_sensitivity": request.CostSensitivity,
	}).Info("Performing intelligent carrier selection")

	criteria := &smdp.SelectionCriteria{
		Region:            request.Region,
		ProfileType:       request.ProfileType,
		Urgency:           request.Urgency,
		CostSensitivity:   request.CostSensitivity,
		PerformanceWeight: request.PerformanceWeight,
		ReliabilityWeight: request.ReliabilityWeight,
	}

	// Select optimal carrier
	score, err := s.manager.SelectOptimalCarrier(ctx, criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to select optimal carrier: %w", err)
	}

	// Record the selection for analytics
	s.recordSelection(score)

	// Generate recommendations
	recommendations := s.generateRecommendations(score, request)

	response := &IntelligentSelectionResponse{
		Success:         true,
		SelectedCarrier: score.Carrier,
		SelectionScore:  score,
		Recommendations: recommendations,
		SelectionTime:   time.Now(),
	}

	s.logger.WithFields(logrus.Fields{
		"selected_carrier": score.CarrierID,
		"total_score":      score.TotalScore,
		"reason":           score.Reason,
	}).Info("Intelligent carrier selection completed")

	return response, nil
}

// IntelligentSelectionRequest represents a comprehensive carrier selection request
type IntelligentSelectionRequest struct {
	Region            string           `json:"region"`
	ProfileType       string           `json:"profile_type"`
	Urgency           string           `json:"urgency"`
	CostSensitivity   float64          `json:"cost_sensitivity"`
	PerformanceWeight float64          `json:"performance_weight"`
	ReliabilityWeight float64          `json:"reliability_weight"`
	UserPreferences   *UserPreferences `json:"user_preferences,omitempty"`
	BusinessContext   *BusinessContext `json:"business_context,omitempty"`
}

// UserPreferences represents user-specific preferences
type UserPreferences struct {
	PreferredCarriers []string `json:"preferred_carriers"`
	ExcludedCarriers  []string `json:"excluded_carriers"`
	MaxResponseTime   int      `json:"max_response_time_ms"`
	MinSuccessRate    float64  `json:"min_success_rate"`
}

// BusinessContext represents business-specific context
type BusinessContext struct {
	CustomerTier    string `json:"customer_tier"`
	ServiceLevel    string `json:"service_level"`
	BillingModel    string `json:"billing_model"`
	GeographicScope string `json:"geographic_scope"`
}

// IntelligentSelectionResponse represents the response from intelligent carrier selection
type IntelligentSelectionResponse struct {
	Success         bool                 `json:"success"`
	SelectedCarrier *smdp.Carrier        `json:"selected_carrier"`
	SelectionScore  *smdp.CarrierScore   `json:"selection_score"`
	Recommendations []string             `json:"recommendations"`
	SelectionTime   time.Time            `json:"selection_time"`
	Alternatives    []*smdp.CarrierScore `json:"alternatives,omitempty"`
}

// recordSelection records a carrier selection for analytics
func (s *SelectionService) recordSelection(score *smdp.CarrierScore) {
	// This would typically be stored in a database for analytics
	s.logger.WithFields(logrus.Fields{
		"carrier_id": score.CarrierID,
		"score":      score.TotalScore,
		"reason":     score.Reason,
		"timestamp":  score.SelectedAt,
	}).Info("Carrier selection recorded")
}

// generateRecommendations generates recommendations based on selection results
func (s *SelectionService) generateRecommendations(score *smdp.CarrierScore, request *IntelligentSelectionRequest) []string {
	recommendations := []string{}

	// Performance-based recommendations
	if score.PerformanceScore < 70 {
		recommendations = append(recommendations, "Consider monitoring carrier performance closely")
	} else if score.PerformanceScore > 90 {
		recommendations = append(recommendations, "Excellent performance - consider prioritizing this carrier")
	}

	// Cost-based recommendations
	if score.CostScore < 50 && request.CostSensitivity > 0.7 {
		recommendations = append(recommendations, "Higher cost detected - consider cost optimization strategies")
	}

	// Reliability-based recommendations
	if score.ReliabilityScore < 60 {
		recommendations = append(recommendations, "Reliability concerns - ensure failover mechanisms are in place")
	}

	// Regional recommendations
	if score.RegionScore < 50 {
		recommendations = append(recommendations, "Regional compatibility issues - consider regional carrier partnerships")
	}

	// Urgency-based recommendations
	if request.Urgency == "high" && score.TotalScore < 80 {
		recommendations = append(recommendations, "High urgency request - consider manual carrier override if needed")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Carrier selection appears optimal")
	}

	return recommendations
}
