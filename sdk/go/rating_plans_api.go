package telecom

import (
	"context"
	"fmt"
)

// RatingPlanAPI handles rating plan-related API calls
type RatingPlanAPI struct {
	client *HTTPClient
}

// NewRatingPlanAPI creates a new RatingPlanAPI
func NewRatingPlanAPI(client *HTTPClient) *RatingPlanAPI {
	return &RatingPlanAPI{client: client}
}

// List retrieves all available rating plans
func (r *RatingPlanAPI) List(ctx context.Context) ([]RatingPlan, error) {
	var plans []RatingPlan
	err := r.client.Get(ctx, "/v1/rating-plans", &plans)
	return plans, err
}

// Get retrieves a rating plan by ID
func (r *RatingPlanAPI) Get(ctx context.Context, planID string) (*RatingPlan, error) {
	var plan RatingPlan
	err := r.client.Get(ctx, fmt.Sprintf("/v1/rating-plans/%s", planID), &plan)
	return &plan, err
}
