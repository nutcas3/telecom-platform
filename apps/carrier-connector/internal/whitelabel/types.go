package whitelabel

import "time"

// BrandingConfig defines partner branding configuration
type BrandingConfig struct {
	ID              string            `json:"id" gorm:"primaryKey"`
	TenantID        string            `json:"tenant_id" gorm:"index"`
	CompanyName     string            `json:"company_name"`
	LogoURL         string            `json:"logo_url"`
	FaviconURL      string            `json:"favicon_url"`
	PrimaryColor    string            `json:"primary_color"`
	SecondaryColor  string            `json:"secondary_color"`
	AccentColor     string            `json:"accent_color"`
	FontFamily      string            `json:"font_family"`
	CustomCSS       string            `json:"custom_css"`
	CustomDomain    string            `json:"custom_domain" gorm:"uniqueIndex"`
	EmailFromName   string            `json:"email_from_name"`
	EmailFromAddr   string            `json:"email_from_address"`
	SupportEmail    string            `json:"support_email"`
	SupportPhone    string            `json:"support_phone"`
	TermsURL        string            `json:"terms_url"`
	PrivacyURL      string            `json:"privacy_url"`
	FooterText      string            `json:"footer_text"`
	SocialLinks     map[string]string `json:"social_links" gorm:"serializer:json"`
	Features        FeatureFlags      `json:"features" gorm:"embedded;embeddedPrefix:feature_"`
	IsActive        bool              `json:"is_active"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// FeatureFlags controls which features are enabled for a whitelabel partner
type FeatureFlags struct {
	ShowPoweredBy      bool `json:"show_powered_by"`
	CustomEmailDomain  bool `json:"custom_email_domain"`
	CustomAPIDomain    bool `json:"custom_api_domain"`
	AdvancedAnalytics  bool `json:"advanced_analytics"`
	WhitelabelMobile   bool `json:"whitelabel_mobile"`
	CustomWebhooks     bool `json:"custom_webhooks"`
	APIDocsBranding    bool `json:"api_docs_branding"`
	MultiLanguage      bool `json:"multi_language"`
}

// EmailTemplate defines customizable email templates
type EmailTemplate struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	TenantID    string    `json:"tenant_id" gorm:"index"`
	TemplateKey string    `json:"template_key" gorm:"index"`
	Subject     string    `json:"subject"`
	HTMLBody    string    `json:"html_body"`
	TextBody    string    `json:"text_body"`
	Variables   []string  `json:"variables" gorm:"serializer:json"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PartnerTier defines partnership levels with different capabilities
type PartnerTier string

const (
	PartnerTierBasic      PartnerTier = "basic"
	PartnerTierProfession PartnerTier = "professional"
	PartnerTierEnterprise PartnerTier = "enterprise"
	PartnerTierPlatinum   PartnerTier = "platinum"
)

// PartnerConfig defines partner-specific configuration
type PartnerConfig struct {
	ID                string            `json:"id" gorm:"primaryKey"`
	TenantID          string            `json:"tenant_id" gorm:"uniqueIndex"`
	Tier              PartnerTier       `json:"tier"`
	RevenueSharePct   float64           `json:"revenue_share_pct"`
	MinMonthlyCommit  float64           `json:"min_monthly_commit"`
	MaxAPIRequests    int64             `json:"max_api_requests"`
	MaxProfiles       int64             `json:"max_profiles"`
	AllowedCountries  []string          `json:"allowed_countries" gorm:"serializer:json"`
	AllowedCarriers   []string          `json:"allowed_carriers" gorm:"serializer:json"`
	CustomPricing     map[string]any    `json:"custom_pricing" gorm:"serializer:json"`
	ContractStartDate time.Time         `json:"contract_start_date"`
	ContractEndDate   time.Time         `json:"contract_end_date"`
	IsActive          bool              `json:"is_active"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}
