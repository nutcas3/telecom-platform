package telecom

import (
	"context"
	"fmt"
)

// AnalyticsAPI provides access to analytics endpoints
type AnalyticsAPI struct {
	client *HTTPClient
}

// NewAnalyticsAPI creates a new AnalyticsAPI client
func NewAnalyticsAPI(client *HTTPClient) *AnalyticsAPI {
	return &AnalyticsAPI{client: client}
}

// PredictChurn predicts churn risk for a specific profile
func (a *AnalyticsAPI) PredictChurn(ctx context.Context, profileID string) (*ChurnPrediction, error) {
	var result ChurnPrediction
	err := a.client.Post(ctx, "/api/v1/analytics/churn/predict", map[string]string{"profile_id": profileID}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetChurnMetrics retrieves overall churn metrics
func (a *AnalyticsAPI) GetChurnMetrics(ctx context.Context, period string) (*ChurnMetrics, error) {
	var result ChurnMetrics
	err := a.client.Get(ctx, "/api/v1/analytics/churn/metrics", &result, map[string]string{"period": period})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetAtRiskCustomers retrieves customers at high risk of churn
func (a *AnalyticsAPI) GetAtRiskCustomers(ctx context.Context, riskLevel ChurnRiskLevel, limit int) ([]*ChurnPrediction, error) {
	var result []*ChurnPrediction
	err := a.client.Get(ctx, "/api/v1/analytics/churn/at-risk", &result, map[string]string{
		"risk_level": string(riskLevel),
		"limit":      fmt.Sprintf("%d", limit),
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetMarketMetrics retrieves market penetration metrics
func (a *AnalyticsAPI) GetMarketMetrics(ctx context.Context, period string) (*MarketMetrics, error) {
	var result MarketMetrics
	err := a.client.Get(ctx, "/api/v1/analytics/market/metrics", &result, map[string]string{"period": period})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetPredictiveMaintenanceMetrics retrieves infrastructure health metrics
func (a *AnalyticsAPI) GetPredictiveMaintenanceMetrics(ctx context.Context, period string) (*PredictiveMaintenanceMetrics, error) {
	var result PredictiveMaintenanceMetrics
	err := a.client.Get(ctx, "/api/v1/analytics/maintenance/metrics", &result, map[string]string{"period": period})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetPricingMetrics retrieves pricing optimization metrics
func (a *AnalyticsAPI) GetPricingMetrics(ctx context.Context, period string) (*PricingMetrics, error) {
	var result PricingMetrics
	err := a.client.Get(ctx, "/api/v1/analytics/pricing/metrics", &result, map[string]string{"period": period})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// OptimizePrice performs pricing optimization for a rate plan
func (a *AnalyticsAPI) OptimizePrice(ctx context.Context, ratePlanID, strategy string) (*PricingOptimizationResult, error) {
	var result PricingOptimizationResult
	err := a.client.Post(ctx, "/api/v1/analytics/pricing/optimize", map[string]interface{}{
		"rate_plan_id": ratePlanID,
		"strategy":     strategy,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
