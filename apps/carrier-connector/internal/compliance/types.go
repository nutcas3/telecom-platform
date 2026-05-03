package compliance

import "time"

// Regulation represents a compliance regulation
type Regulation string

const (
	RegulationGDPR  Regulation = "GDPR"
	RegulationCCPA  Regulation = "CCPA"
	RegulationLGPD  Regulation = "LGPD"
	RegulationPDPA  Regulation = "PDPA"
	RegulationPIPL  Regulation = "PIPL"
)

// ConsentType represents types of data processing consent
type ConsentType string

const (
	ConsentTypeMarketing    ConsentType = "marketing"
	ConsentTypeAnalytics    ConsentType = "analytics"
	ConsentTypeThirdParty   ConsentType = "third_party"
	ConsentTypeDataSharing  ConsentType = "data_sharing"
	ConsentTypePersonalized ConsentType = "personalized"
)

// DataSubjectRequest represents a GDPR/CCPA data subject request
type DataSubjectRequest struct {
	ID            string            `json:"id" gorm:"primaryKey"`
	TenantID      string            `json:"tenant_id" gorm:"index"`
	SubjectID     string            `json:"subject_id" gorm:"index"`
	SubjectEmail  string            `json:"subject_email"`
	RequestType   DSRType           `json:"request_type"`
	Regulation    Regulation        `json:"regulation"`
	Status        DSRStatus         `json:"status" gorm:"index"`
	RequestedAt   time.Time         `json:"requested_at"`
	VerifiedAt    *time.Time        `json:"verified_at"`
	CompletedAt   *time.Time        `json:"completed_at"`
	DueDate       time.Time         `json:"due_date"`
	Notes         string            `json:"notes"`
	DataExportURL string            `json:"data_export_url,omitempty"`
	Metadata      map[string]any    `json:"metadata" gorm:"serializer:json"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// DSRType represents data subject request types
type DSRType string

const (
	DSRTypeAccess      DSRType = "access"
	DSRTypeRectify     DSRType = "rectification"
	DSRTypeErasure     DSRType = "erasure"
	DSRTypePortability DSRType = "portability"
	DSRTypeRestrict    DSRType = "restriction"
	DSRTypeObject      DSRType = "objection"
)

// DSRStatus represents request status
type DSRStatus string

const (
	DSRStatusPending    DSRStatus = "pending"
	DSRStatusVerifying  DSRStatus = "verifying"
	DSRStatusProcessing DSRStatus = "processing"
	DSRStatusCompleted  DSRStatus = "completed"
	DSRStatusRejected   DSRStatus = "rejected"
)

// ConsentRecord tracks user consent
type ConsentRecord struct {
	ID           string      `json:"id" gorm:"primaryKey"`
	TenantID     string      `json:"tenant_id" gorm:"index"`
	SubjectID    string      `json:"subject_id" gorm:"index"`
	ConsentType  ConsentType `json:"consent_type"`
	Granted      bool        `json:"granted"`
	GrantedAt    *time.Time  `json:"granted_at"`
	RevokedAt    *time.Time  `json:"revoked_at"`
	IPAddress    string      `json:"ip_address"`
	UserAgent    string      `json:"user_agent"`
	ConsentText  string      `json:"consent_text"`
	Version      string      `json:"version"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

// DataResidencyConfig defines data residency requirements
type DataResidencyConfig struct {
	ID               string     `json:"id" gorm:"primaryKey"`
	TenantID         string     `json:"tenant_id" gorm:"uniqueIndex"`
	PrimaryRegion    string     `json:"primary_region"`
	AllowedRegions   []string   `json:"allowed_regions" gorm:"serializer:json"`
	RestrictedData   []string   `json:"restricted_data" gorm:"serializer:json"`
	EncryptionReq    bool       `json:"encryption_required"`
	RetentionDays    int        `json:"retention_days"`
	CrossBorderRules []CrossBorderRule `json:"cross_border_rules" gorm:"serializer:json"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// CrossBorderRule defines rules for cross-border data transfer
type CrossBorderRule struct {
	FromRegion    string `json:"from_region"`
	ToRegion      string `json:"to_region"`
	Allowed       bool   `json:"allowed"`
	RequiresSCC   bool   `json:"requires_scc"`
	RequiresDPIA  bool   `json:"requires_dpia"`
}

// AuditLog represents a compliance audit log entry
type AuditLog struct {
	ID           string         `json:"id" gorm:"primaryKey"`
	TenantID     string         `json:"tenant_id" gorm:"index"`
	Jurisdiction string         `json:"jurisdiction" gorm:"index"`
	ActorID      string         `json:"actor_id" gorm:"index"`
	ActorType    string         `json:"actor_type"`
	Action       string         `json:"action" gorm:"index"`
	ResourceType string         `json:"resource_type"`
	ResourceID   string         `json:"resource_id"`
	OldValue     map[string]any `json:"old_value" gorm:"serializer:json"`
	NewValue     map[string]any `json:"new_value" gorm:"serializer:json"`
	IPAddress    string         `json:"ip_address"`
	UserAgent    string         `json:"user_agent"`
	Timestamp    time.Time      `json:"timestamp" gorm:"index"`
	Metadata     map[string]any `json:"metadata" gorm:"serializer:json"`
}

// RegulatoryReport represents a regulatory compliance report
type RegulatoryReport struct {
	ID           string     `json:"id" gorm:"primaryKey"`
	TenantID     string     `json:"tenant_id" gorm:"index"`
	Regulation   Regulation `json:"regulation"`
	ReportType   string     `json:"report_type"`
	PeriodStart  time.Time  `json:"period_start"`
	PeriodEnd    time.Time  `json:"period_end"`
	Status       string     `json:"status"`
	FileURL      string     `json:"file_url"`
	SubmittedAt  *time.Time `json:"submitted_at"`
	GeneratedAt  time.Time  `json:"generated_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

// PrivacyPolicy tracks privacy policy versions
type PrivacyPolicy struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	TenantID    string    `json:"tenant_id" gorm:"index"`
	Version     string    `json:"version"`
	Content     string    `json:"content"`
	Regulation  Regulation `json:"regulation"`
	EffectiveAt time.Time `json:"effective_at"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}
