package models

import (
	"time"
)

type SubscriberAccount struct {
	IMSI        string    `json:"imsi"`
	Balance     float64   `json:"balance"`
	DataLimit   float64   `json:"data_limit"`
	DataUsed    float64   `json:"data_used"`
	VoiceLimit  float64   `json:"voice_limit"`
	VoiceUsed   float64   `json:"voice_used"`
	SMSLimit    float64   `json:"sms_limit"`
	SMSUsed     float64   `json:"sms_used"`
	Status      string    `json:"status"`
	LastUpdated time.Time `json:"last_updated"`
}

type UsageEvent struct {
	ID         uint        `json:"id"`
	IMSI       string      `json:"imsi"`
	SessionID  string      `json:"session_id"`
	UsageType  UsageType   `json:"usage_type"`
	Volume     float64     `json:"volume"`
	Timestamp  time.Time   `json:"timestamp"`
	Rate       float64     `json:"rate"`
	Cost       float64     `json:"cost"`
	Subscriber *Subscriber `json:"subscriber,omitempty"`
}

type Invoice struct {
	ID           uint              `json:"id"`
	SubscriberID uint              `json:"subscriber_id"`
	Amount       float64           `json:"amount"`
	Currency     string            `json:"currency"`
	Status       InvoiceStatus     `json:"status"`
	DueDate      time.Time         `json:"due_date"`
	CreatedAt    time.Time         `json:"created_at"`
	LineItems    []InvoiceLineItem `json:"line_items"`
	Subscriber   *Subscriber       `json:"subscriber,omitempty"`
}

type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "DRAFT"
	InvoiceStatusPending   InvoiceStatus = "PENDING"
	InvoiceStatusPaid      InvoiceStatus = "PAID"
	InvoiceStatusOverdue   InvoiceStatus = "OVERDUE"
	InvoiceStatusCancelled InvoiceStatus = "CANCELLED"
)

type InvoiceLineItem struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Amount      float64 `json:"amount"`
}

type RatingPlan struct {
	PlanID      string        `json:"plan_id"`
	Name        string        `json:"name"`
	DataRate    float64       `json:"data_rate"`
	VoiceRate   float64       `json:"voice_rate"`
	SMSRate     float64       `json:"sms_rate"`
	MonthlyFee  float64       `json:"monthly_fee"`
	DataLimit   float64       `json:"data_limit"`
	VoiceLimit  float64       `json:"voice_limit"`
	SMSLimit    float64       `json:"sms_limit"`
	Subscribers []*Subscriber `json:"subscribers,omitempty"`
}

type Alert struct {
	ID           uint          `json:"id"`
	Type         AlertType     `json:"type"`
	Severity     AlertSeverity `json:"severity"`
	Message      string        `json:"message"`
	SubscriberID *int          `json:"subscriber_id"`
	Timestamp    time.Time     `json:"timestamp"`
	Resolved     bool          `json:"resolved"`
	Subscriber   *Subscriber   `json:"subscriber,omitempty"`
}

type AlertType string

const (
	AlertTypeLowBalance    AlertType = "LOW_BALANCE"
	AlertTypeHighUsage     AlertType = "HIGH_USAGE"
	AlertTypePaymentFailed AlertType = "PAYMENT_FAILED"
	AlertTypeSystemError   AlertType = "SYSTEM_ERROR"
)

type AlertSeverity string

const (
	AlertSeverityLow      AlertSeverity = "LOW"
	AlertSeverityMedium   AlertSeverity = "MEDIUM"
	AlertSeverityHigh     AlertSeverity = "HIGH"
	AlertSeverityCritical AlertSeverity = "CRITICAL"
)

type SystemStats struct {
	ActiveSessions   int     `json:"active_sessions"`
	TotalAccounts    int     `json:"total_accounts"`
	BlockedUsers     int     `json:"blocked_users"`
	LowBalanceAlerts int     `json:"low_balance_alerts"`
	Uptime           float64 `json:"uptime"`
}

type HealthStatus struct {
	RedisConnected bool      `json:"redis_connected"`
	ActiveSync     bool      `json:"active_sync"`
	LastSync       time.Time `json:"last_sync"`
	MemoryUsage    float64   `json:"memory_usage"`
}

type PaymentMethod struct {
	ID           string            `json:"id" gorm:"primaryKey"`
	SubscriberID uint              `json:"subscriber_id"`
	GatewayID    string            `json:"gateway_id"`
	Type         PaymentMethodType `json:"type"`
	CustomerID   string            `json:"customer_id"`
	Last4        string            `json:"last4"`
	Brand        string            `json:"brand"`
	ExpiryMonth  int               `json:"expiry_month"`
	ExpiryYear   int               `json:"expiry_year"`
	IsDefault    bool              `json:"is_default"`
	CreatedAt    time.Time         `json:"created_at"`
	Metadata     map[string]string `json:"metadata" gorm:"-"`
}

type PaymentMethodType string

const (
	PaymentMethodTypeCreditCard  PaymentMethodType = "CREDIT_CARD"
	PaymentMethodTypeBankAccount PaymentMethodType = "BANK_ACCOUNT"
)

type Transaction struct {
	ID            uint      `json:"id"`
	SubscriberID  uint      `json:"subscriber_id"`
	TransactionID string    `json:"transaction_id"`
	Type          string    `json:"type"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	Description   string    `json:"description"`
	ParentID      *uint     `json:"parent_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Notification struct {
	ID           uint      `json:"id"`
	SubscriberID uint      `json:"subscriber_id"`
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	Message      string    `json:"message"`
	Read         bool      `json:"read"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateSubscriberRequest struct {
	MSISDN         string  `json:"msisdn"`
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
	Email          string  `json:"email"`
	OrganizationID *string `json:"organization_id"`
	PlanID         int     `json:"plan_id"`
	EUICCID        *string `json:"euicc_id"`
}

type UpdateSubscriberRequest struct {
	FirstName      *string `json:"first_name"`
	LastName       *string `json:"last_name"`
	Email          *string `json:"email"`
	OrganizationID *string `json:"organization_id"`
	PlanID         *int    `json:"plan_id"`
}

type TopUpRequest struct {
	Amount          float64 `json:"amount"`
	PaymentMethodID *string `json:"payment_method_id"`
}

type AddPaymentMethodRequest struct {
	Type      PaymentMethodType `json:"type"`
	Token     string            `json:"token"`
	IsDefault bool              `json:"is_default"`
}

type CreateAlertRequest struct {
	Type         AlertType     `json:"type"`
	Severity     AlertSeverity `json:"severity"`
	Message      string        `json:"message"`
	SubscriberID *int          `json:"subscriber_id"`
}

type ListSubscribersRequest struct {
	First  int               `json:"first"`
	After  *string           `json:"after"`
	Filter *SubscriberFilter `json:"filter"`
	Sort   *SubscriberSort   `json:"sort"`
}

type SubscriberFilter struct {
	Status         *SubscriberStatus `json:"status"`
	OrganizationID *string           `json:"organization_id"`
	PlanID         *int              `json:"plan_id"`
	Search         *string           `json:"search"`
	BalanceMin     *float64          `json:"balance_min"`
	BalanceMax     *float64          `json:"balance_max"`
	DataUsageMin   *float64          `json:"data_usage_min"`
	DataUsageMax   *float64          `json:"data_usage_max"`
}

type SubscriberSort string

const (
	SubscriberSortCreatedAtAsc  SubscriberSort = "CREATED_AT_ASC"
	SubscriberSortCreatedAtDesc SubscriberSort = "CREATED_AT_DESC"
	SubscriberSortUpdatedAtAsc  SubscriberSort = "UPDATED_AT_ASC"
	SubscriberSortUpdatedAtDesc SubscriberSort = "UPDATED_AT_DESC"
	SubscriberSortFirstNameAsc  SubscriberSort = "FIRST_NAME_ASC"
	SubscriberSortFirstNameDesc SubscriberSort = "FIRST_NAME_DESC"
	SubscriberSortLastNameAsc   SubscriberSort = "LAST_NAME_ASC"
	SubscriberSortLastNameDesc  SubscriberSort = "LAST_NAME_DESC"
	SubscriberSortEmailAsc      SubscriberSort = "EMAIL_ASC"
	SubscriberSortEmailDesc     SubscriberSort = "EMAIL_DESC"
	SubscriberSortBalanceAsc    SubscriberSort = "BALANCE_ASC"
	SubscriberSortBalanceDesc   SubscriberSort = "BALANCE_DESC"
	SubscriberSortDataUsedAsc   SubscriberSort = "DATA_USED_ASC"
	SubscriberSortDataUsedDesc  SubscriberSort = "DATA_USED_DESC"
)
