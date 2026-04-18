package telecom

import (
	"context"
	"time"
)

// Subscriber represents a telecom subscriber
type Subscriber struct {
	ID             int64     `json:"id"`
	IMSI           string    `json:"imsi"`
	MSISDN         string    `json:"msisdn"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Email          string    `json:"email"`
	OrganizationID *string   `json:"organization_id,omitempty"`
	Status         string    `json:"status"`
	PlanID         int64     `json:"plan_id"`
	Balance        float64   `json:"balance"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// SubscriberList represents a paginated list of subscribers
type SubscriberList struct {
	Subscribers []Subscriber `json:"subscribers"`
	Total       int64        `json:"total"`
	Page        int32        `json:"page"`
	PageSize    int32        `json:"page_size"`
	HasNext     bool         `json:"has_next"`
	HasPrev     bool         `json:"has_prev"`
}

// CreateSubscriberRequest represents a request to create a subscriber
type CreateSubscriberRequest struct {
	IMSI           string  `json:"imsi"`
	MSISDN         string  `json:"msisdn"`
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
	Email          string  `json:"email"`
	PlanID         int64   `json:"plan_id"`
	OrganizationID *string `json:"organization_id,omitempty"`
}

// UpdateSubscriberRequest represents a request to update a subscriber
type UpdateSubscriberRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Email     *string `json:"email,omitempty"`
	PlanID    *int64  `json:"plan_id,omitempty"`
	Status    *string `json:"status,omitempty"`
}

// UsageStats represents usage statistics
type UsageStats struct {
	SubscriberID string    `json:"subscriber_id"`
	DataUp       int64     `json:"data_up"`
	DataDown     int64     `json:"data_down"`
	VoiceSeconds int64     `json:"voice_seconds"`
	SMSCount     int64     `json:"sms_count"`
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
	Cost         float64   `json:"cost"`
}

// PaymentTransaction represents a payment transaction
type PaymentTransaction struct {
	ID            string                 `json:"id"`
	SubscriberID  string                 `json:"subscriber_id"`
	Amount        float64                `json:"amount"`
	Currency      string                 `json:"currency"`
	Status        string                 `json:"status"`
	Gateway       string                 `json:"gateway"`
	TransactionID *string                `json:"transaction_id,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// CreatePaymentRequest represents a request to create a payment
type CreatePaymentRequest struct {
	SubscriberID string                 `json:"subscriber_id"`
	Amount       float64                `json:"amount"`
	Currency     string                 `json:"currency"`
	Gateway      string                 `json:"gateway"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// SystemStats represents system statistics
type SystemStats struct {
	ActiveSessions   int64   `json:"active_sessions"`
	TotalAccounts    int64   `json:"total_accounts"`
	BlockedUsers     int64   `json:"blocked_users"`
	LowBalanceAlerts int64   `json:"low_balance_alerts"`
	Uptime           float64 `json:"uptime"`
	CPUUsage         float64 `json:"cpu_usage"`
	MemoryUsage      float64 `json:"memory_usage"`
}

// HealthStatus represents system health status
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]interface{} `json:"checks"`
	Uptime    float64                `json:"uptime"`
}

// RatingPlan represents a rating plan
type RatingPlan struct {
	PlanID     string  `json:"plan_id"`
	Name       string  `json:"name"`
	DataRate   float64 `json:"data_rate"`
	VoiceRate  float64 `json:"voice_rate"`
	SMSRate    float64 `json:"sms_rate"`
	MonthlyFee float64 `json:"monthly_fee"`
	DataLimit  int64   `json:"data_limit"`
	VoiceLimit int64   `json:"voice_limit"`
	SMSLimit   int64   `json:"sms_limit"`
}

// UsageEvent represents a usage event
type UsageEvent struct {
	ID           string                 `json:"id"`
	SubscriberID string                 `json:"subscriber_id"`
	UsageType    string                 `json:"usage_type"`
	Amount       int64                  `json:"amount"`
	Cost         float64                `json:"cost"`
	Timestamp    time.Time              `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// CurrentSession represents a current active session
type CurrentSession struct {
	SessionID    string    `json:"session_id"`
	StartTime    time.Time `json:"start_time"`
	DataUp       int64     `json:"data_up"`
	DataDown     int64     `json:"data_down"`
	VoiceSeconds int64     `json:"voice_seconds"`
	SMSCount     int64     `json:"sms_count"`
}

// RealTimeUsage represents real-time usage data
type RealTimeUsage struct {
	CurrentSession *CurrentSession  `json:"current_session,omitempty"`
	TodayUsage     map[string]int64 `json:"today_usage,omitempty"`
}

// gRPC Service Interfaces

// SubscriberServiceClient interface for gRPC subscriber service
type SubscriberServiceClient interface {
	GetSubscriber(ctx context.Context, req *GetSubscriberRequest) (*Subscriber, error)
	ListSubscribers(ctx context.Context, req *ListSubscribersRequest) (*SubscriberList, error)
	CreateSubscriber(ctx context.Context, req *CreateSubscriberRequest) (*Subscriber, error)
	UpdateSubscriber(ctx context.Context, req *UpdateSubscriberRequest) (*Subscriber, error)
	DeleteSubscriber(ctx context.Context, req *DeleteSubscriberRequest) (*DeleteSubscriberResponse, error)
}

// GetSubscriberRequest represents a gRPC request to get a subscriber
type GetSubscriberRequest struct {
	Id int64 `json:"id"`
}

// ListSubscribersRequest represents a gRPC request to list subscribers
type ListSubscribersRequest struct {
	Page     int32  `json:"page"`
	PageSize int32  `json:"page_size"`
	Status   string `json:"status,omitempty"`
}

// DeleteSubscriberRequest represents a gRPC request to delete a subscriber
type DeleteSubscriberRequest struct {
	Id int64 `json:"id"`
}

// DeleteSubscriberResponse represents a gRPC response for delete subscriber
type DeleteSubscriberResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// PaymentServiceClient interface for gRPC payment service
type PaymentServiceClient interface {
	CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*PaymentTransaction, error)
	GetPayment(ctx context.Context, req *GetPaymentRequest) (*PaymentTransaction, error)
	ListPayments(ctx context.Context, req *ListPaymentsRequest) (*PaymentTransactionList, error)
}

// GetPaymentRequest represents a gRPC request to get a payment
type GetPaymentRequest struct {
	Id string `json:"id"`
}

// ListPaymentsRequest represents a gRPC request to list payments
type ListPaymentsRequest struct {
	SubscriberID string `json:"subscriber_id,omitempty"`
	Status       string `json:"status,omitempty"`
	Page         int32  `json:"page"`
	PageSize     int32  `json:"page_size"`
}

// PaymentTransactionList represents a list of payment transactions
type PaymentTransactionList struct {
	Transactions []PaymentTransaction `json:"transactions"`
	Total        int64                `json:"total"`
	Page         int32                `json:"page"`
	PageSize     int32                `json:"page_size"`
	HasNext      bool                 `json:"has_next"`
	HasPrev      bool                 `json:"has_prev"`
}

// UsageServiceClient interface for gRPC usage service
type UsageServiceClient interface {
	GetUsageStats(ctx context.Context, req *GetUsageStatsRequest) (*UsageStats, error)
	GetRealTimeUsage(ctx context.Context, req *GetRealTimeUsageRequest) (*RealTimeUsage, error)
	ListUsageEvents(ctx context.Context, req *ListUsageEventsRequest) (*UsageEventList, error)
}

// GetUsageStatsRequest represents a gRPC request to get usage stats
type GetUsageStatsRequest struct {
	SubscriberID int64     `json:"subscriber_id"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
}

// GetRealTimeUsageRequest represents a gRPC request to get real-time usage
type GetRealTimeUsageRequest struct {
	SubscriberID int64 `json:"subscriber_id"`
}

// ListUsageEventsRequest represents a gRPC request to list usage events
type ListUsageEventsRequest struct {
	SubscriberID *int64     `json:"subscriber_id,omitempty"`
	UsageType    string     `json:"usage_type,omitempty"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	Page         int32      `json:"page"`
	PageSize     int32      `json:"page_size"`
}

// UsageEventList represents a list of usage events
type UsageEventList struct {
	Events   []UsageEvent `json:"events"`
	Total    int64        `json:"total"`
	Page     int32        `json:"page"`
	PageSize int32        `json:"page_size"`
	HasNext  bool         `json:"has_next"`
	HasPrev  bool         `json:"has_prev"`
}

// SystemServiceClient interface for gRPC system service
type SystemServiceClient interface {
	GetSystemStats(ctx context.Context, req *GetSystemStatsRequest) (*SystemStats, error)
	GetHealthStatus(ctx context.Context, req *GetHealthStatusRequest) (*HealthStatus, error)
}

// GetSystemStatsRequest represents a gRPC request to get system stats
type GetSystemStatsRequest struct{}

// GetHealthStatusRequest represents a gRPC request to get health status
type GetHealthStatusRequest struct{}
