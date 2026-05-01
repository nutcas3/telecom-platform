package services

import (
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/config"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/database"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/payment/gateways"
)

// SubscriberService handles subscriber management operations. Methods are split across:
//
//	subscriber_crud.go, subscriber_lifecycle.go, subscriber_billing.go,
//	subscriber_alerts.go, subscriber_payment.go, subscriber_subscriptions.go,
//	subscriber_esim.go, subscriber_imsi.go, subscriber_auth.go.
type SubscriberService struct {
	db         *database.Database
	config     *config.Config
	amfClient  *AMFClient
	es2Service *ES2Service
	stripeGW   *gateways.StripeGateway
}

// NewSubscriberService creates a new subscriber service.
func NewSubscriberService(db *database.Database, cfg *config.Config) *SubscriberService {
	return &SubscriberService{
		db:         db,
		config:     cfg,
		amfClient:  NewAMFClient("http://localhost:8081"),
		es2Service: NewES2Service(&cfg.ES2),
		stripeGW:   gateways.NewStripeGateway(cfg.Payment.StripeAPIKey, cfg.Payment.StripeWebhookSecret),
	}
}

// CreateSubscriberRequest is the payload for creating a new subscriber.
type CreateSubscriberRequest struct {
	MSISDN         string `json:"msisdn" validate:"required"`
	FirstName      string `json:"first_name" validate:"required"`
	LastName       string `json:"last_name" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
	OrganizationID string `json:"organization_id"`
	PlanID         uint   `json:"plan_id" validate:"required"`
	EUICCID        string `json:"euicc_id"`
}

// UpdateSubscriberRequest is the payload for updating an existing subscriber.
type UpdateSubscriberRequest struct {
	FirstName *string                 `json:"first_name"`
	LastName  *string                 `json:"last_name"`
	Email     *string                 `json:"email"`
	Status    models.SubscriberStatus `json:"status"`
	PlanID    *uint                   `json:"plan_id"`
}

// ListSubscribersRequest is the query-filter payload for listing subscribers.
type ListSubscribersRequest struct {
	Cursor         string                  `json:"cursor" query:"cursor"`
	Limit          int                     `json:"limit" query:"limit"`
	Status         models.SubscriberStatus `json:"status" query:"status"`
	OrganizationID string                  `json:"organization_id" query:"organization_id"`
	Search         string                  `json:"search" query:"search"`
}

// ListSubscribersResponse is the cursor-paginated subscriber list result.
type ListSubscribersResponse struct {
	Subscribers []models.Subscriber `json:"subscribers"`
	NextCursor  string              `json:"next_cursor,omitempty"`
	HasMore     bool                `json:"has_more"`
}
