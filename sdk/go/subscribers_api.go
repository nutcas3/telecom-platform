package telecom

import (
	"context"
	"fmt"
)

// SubscriberAPI handles subscriber-related API calls
type SubscriberAPI struct {
	client *HTTPClient
}

// NewSubscriberAPI creates a new SubscriberAPI
func NewSubscriberAPI(client *HTTPClient) *SubscriberAPI {
	return &SubscriberAPI{client: client}
}

// Get retrieves a subscriber by ID
func (s *SubscriberAPI) Get(ctx context.Context, id int64) (*Subscriber, error) {
	var subscriber Subscriber
	err := s.client.Get(ctx, fmt.Sprintf("/v1/subscribers/%d", id), &subscriber)
	return &subscriber, err
}

// List retrieves a list of subscribers
func (s *SubscriberAPI) List(ctx context.Context, page, pageSize int32, status string) (*SubscriberList, error) {
	params := map[string]string{
		"page":      fmt.Sprintf("%d", page),
		"page_size": fmt.Sprintf("%d", pageSize),
	}
	if status != "" {
		params["status"] = status
	}

	var list SubscriberList
	err := s.client.Get(ctx, "/v1/subscribers", &list, params)
	return &list, err
}

// Create creates a new subscriber
func (s *SubscriberAPI) Create(ctx context.Context, req *CreateSubscriberRequest) (*Subscriber, error) {
	var subscriber Subscriber
	err := s.client.Post(ctx, "/v1/subscribers", req, &subscriber)
	return &subscriber, err
}

// Update updates an existing subscriber
func (s *SubscriberAPI) Update(ctx context.Context, id int64, req *UpdateSubscriberRequest) (*Subscriber, error) {
	var subscriber Subscriber
	err := s.client.Put(ctx, fmt.Sprintf("/v1/subscribers/%d", id), req, &subscriber)
	return &subscriber, err
}

// Delete deletes a subscriber
func (s *SubscriberAPI) Delete(ctx context.Context, id int64) error {
	return s.client.Delete(ctx, fmt.Sprintf("/v1/subscribers/%d", id))
}

// Suspend suspends a subscriber
func (s *SubscriberAPI) Suspend(ctx context.Context, id int64) (*Subscriber, error) {
	var subscriber Subscriber
	err := s.client.Post(ctx, fmt.Sprintf("/v1/subscribers/%d/suspend", id), nil, &subscriber)
	return &subscriber, err
}

// Activate activates a suspended subscriber
func (s *SubscriberAPI) Activate(ctx context.Context, id int64) (*Subscriber, error) {
	var subscriber Subscriber
	err := s.client.Post(ctx, fmt.Sprintf("/v1/subscribers/%d/activate", id), nil, &subscriber)
	return &subscriber, err
}
