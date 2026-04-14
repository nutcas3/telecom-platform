package models

import (
	"time"

	"gorm.io/gorm"
)

type IMSI string

type Subscriber struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	IMSI   IMSI   `json:"imsi" gorm:"uniqueIndex;not null"`
	MSISDN string `json:"msisdn" gorm:"uniqueIndex"` // Mobile phone number
	IMEI   string `json:"imei"`                      // Device identifier

	FirstName      string `json:"first_name" gorm:"not null"`
	LastName       string `json:"last_name" gorm:"not null"`
	Email          string `json:"email" gorm:"uniqueIndex"`
	OrganizationID string `json:"organization_id"`

	Status SubscriberStatus `json:"status" gorm:"default:'active'"`
	PlanID uint             `json:"plan_id"`

	AuthKey     string `json:"auth_key" gorm:"not null"`
	OPc         string `json:"opc" gorm:"not null"`
	ServingPLMN PLMN   `json:"serving_plmn" gorm:"not null"`

	EUICCID       string        `json:"euicc_id"`
	ProfileID     string        `json:"profile_id"`
	ProfileStatus ProfileStatus `json:"profile_status" gorm:"default:'inactive'"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Plan         ServicePlan   `json:"plan" gorm:"foreignKey:PlanID"`
	Sessions     []Session     `json:"sessions,omitempty" gorm:"foreignKey:SubscriberID"`
	UsageRecords []UsageRecord `json:"usage_records,omitempty" gorm:"foreignKey:SubscriberID"`
}

type ServicePlan struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`

	DataLimit  int64 `json:"data_limit"`
	VoiceLimit int   `json:"voice_limit"`
	SMSLimit   int   `json:"sms_limit"`

	MonthlyFee float64 `json:"monthly_fee"`
	DataRate   float64 `json:"data_rate"`
	VoiceRate  float64 `json:"voice_rate"`
	SMSRate    float64 `json:"sms_rate"`

	ARPA int  `json:"arpa"`
	AMBR AMBR `json:"ambr"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type Session struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	SubscriberID uint   `json:"subscriber_id" gorm:"not null"`
	SessionID    string `json:"session_id" gorm:"uniqueIndex;not null"`

	PDUAddress string     `json:"pdu_address"`
	DNN        string     `json:"dnn"`
	SNSSAI     SNSSAI     `json:"snssai"`
	QoS        QoSProfile `json:"qos" gorm:"embedded"`

	Status    SessionStatus `json:"status" gorm:"default:'active'"`
	StartTime time.Time     `json:"start_time"`
	EndTime   *time.Time    `json:"end_time,omitempty"`

	DataUsed  int64 `json:"data_used"`
	VoiceUsed int   `json:"voice_used"`
	SMSUsed   int   `json:"sms_used"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type UsageRecord struct {
	ID           uint `json:"id" gorm:"primaryKey"`
	SubscriberID uint `json:"subscriber_id" gorm:"not null"`
	SessionID    uint `json:"session_id"`

	UsageType UsageType `json:"usage_type"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Volume    int64     `json:"volume"`
	Rate      float64   `json:"rate"`
	Cost      float64   `json:"cost"`

	BillingCycle string `json:"billing_cycle"`
	InvoiceID    *uint  `json:"invoice_id,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SubscriberStatus string

const (
	SubscriberStatusActive       SubscriberStatus = "active"
	SubscriberStatusInactive     SubscriberStatus = "inactive"
	SubscriberStatusSuspended    SubscriberStatus = "suspended"
	SubscriberStatusTerminated   SubscriberStatus = "terminated"
	SubscriberStatusProvisioning SubscriberStatus = "provisioning"
)

type ProfileStatus string

const (
	ProfileStatusInactive    ProfileStatus = "inactive"
	ProfileStatusActive      ProfileStatus = "active"
	ProfileStatusDownloading ProfileStatus = "downloading"
	ProfileStatusError       ProfileStatus = "error"
)

type SessionStatus string

const (
	SessionStatusActive      SessionStatus = "active"
	SessionStatusInactive    SessionStatus = "inactive"
	SessionStatusTerminating SessionStatus = "terminating"
)

type UsageType string

const (
	UsageTypeData  UsageType = "data"
	UsageTypeVoice UsageType = "voice"
	UsageTypeSMS   UsageType = "sms"
)

type PLMN struct {
	MCC string `json:"mcc"`
	MNC string `json:"mnc"`
}

type SNSSAI struct {
	SST uint8  `json:"sst"`
	SD  uint32 `json:"sd"`
}

type AMBR struct {
	Uplink   uint32 `json:"uplink"`
	Downlink uint32 `json:"downlink"`
}

type QoSProfile struct {
	QCI      uint8  `json:"qci"`
	ARPA     uint32 `json:"arpa"`
	MBR      AMBR   `json:"mbr"`
	GBR      AMBR   `json:"gbr"`
	Priority uint8  `json:"priority"`
}

func (Subscriber) TableName() string {
	return "subscribers"
}

func (ServicePlan) TableName() string {
	return "service_plans"
}

func (Session) TableName() string {
	return "sessions"
}

func (UsageRecord) TableName() string {
	return "usage_records"
}
