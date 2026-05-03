package telecom

import "context"

// SecurityAPI provides access to security endpoints
type SecurityAPI struct {
	client *HTTPClient
}

// NewSecurityAPI creates a new SecurityAPI client
func NewSecurityAPI(client *HTTPClient) *SecurityAPI {
	return &SecurityAPI{client: client}
}

// AnalyzeTransaction analyzes a transaction for fraud
func (s *SecurityAPI) AnalyzeTransaction(ctx context.Context, transaction map[string]interface{}) (*FraudAlert, error) {
	var result FraudAlert
	err := s.client.Post(ctx, "/api/v1/security/fraud/analyze", transaction, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFraudAlerts retrieves fraud alerts
func (s *SecurityAPI) GetFraudAlerts(ctx context.Context, filter FraudAlertFilter) ([]*FraudAlert, error) {
	var result []*FraudAlert
	err := s.client.Post(ctx, "/api/v1/security/fraud/alerts", filter, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateAlertStatus updates the status of a fraud alert
func (s *SecurityAPI) UpdateAlertStatus(ctx context.Context, alertID, status string, actions []string) error {
	req := map[string]interface{}{
		"status":  status,
		"actions": actions,
	}
	return s.client.Put(ctx, "/api/v1/security/fraud/alerts/"+alertID, req, nil)
}

// GetFraudMetrics returns fraud detection metrics
func (s *SecurityAPI) GetFraudMetrics(ctx context.Context, period string) (*FraudMetrics, error) {
	var result FraudMetrics
	err := s.client.Get(ctx, "/api/v1/security/fraud/metrics", &result, map[string]string{"period": period})
	if err != nil {
		return nil, err
	}
	return &result, nil
}
