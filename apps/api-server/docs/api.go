package docs

// Subscriber represents a telecom subscriber
// @Description Subscriber represents a telecom subscriber in the system
type Subscriber struct {
	// Unique identifier of the subscriber
	// @ReadOnly: true
	ID uint `json:"id"`

	// IMSI (International Mobile Subscriber Identity)
	// @Example: 208930000000001
	IMSI string `json:"imsi"`

	// MSISDN (Mobile Station International Subscriber Directory Number)
	// @Example: +33612345678
	MSISDN string `json:"msisdn"`

	// First name of the subscriber
	// @Example: John
	FirstName string `json:"first_name"`

	// Last name of the subscriber
	// @Example: Doe
	LastName string `json:"last_name"`

	// Email address of the subscriber
	// @Format: email
	// @Example: john.doe@example.com
	Email string `json:"email"`

	// Current status of the subscriber
	// @Enum: active,inactive,suspended
	// @Example: active
	Status string `json:"status"`

	// Data plan assigned to the subscriber
	// @Example: premium_5g
	PlanID string `json:"plan_id"`

	// eSIM profile status
	// @Enum: active,provisioning,inactive,failed
	// @Example: active
	ProfileStatus string `json:"profile_status"`

	// Account creation time
	// @ReadOnly: true
	// @Format: datetime
	CreatedAt string `json:"created_at"`

	// Account last update time
	// @ReadOnly: true
	// @Format: datetime
	UpdatedAt string `json:"updated_at"`
}

// Service represents a Kubernetes service
// @Description Service represents a Kubernetes service in the platform
type Service struct {
	// Name of the service
	// @Example: api-server
	Name string `json:"name"`

	// Kubernetes namespace where the service is deployed
	// @Example: telecom-platform
	Namespace string `json:"namespace"`

	// Number of desired replicas
	// @Example: 3
	Replicas int32 `json:"replicas"`

	// Number of ready replicas
	// @Example: 3
	ReadyReplicas int32 `json:"ready_replicas"`

	// Number of available replicas
	// @Example: 3
	AvailableReplicas int32 `json:"available_replicas"`

	// Service creation timestamp
	// @Format: datetime
	CreationTimestamp string `json:"creation_timestamp"`
}

// Deployment represents a deployment in the platform
// @Description Deployment represents a deployment in the platform
type Deployment struct {
	// Unique identifier of the deployment
	// @ReadOnly: true
	ID uint `json:"id"`

	// Name of the service being deployed
	// @Example: api-server
	Service string `json:"service"`

	// Version being deployed
	// @Example: v1.2.3
	Version string `json:"version"`

	// Current status of the deployment
	// @Enum: completed,failed,running,pending,rolling_back
	// @Example: completed
	Status string `json:"status"`

	// Deployment start time
	// @Format: datetime
	StartedAt string `json:"started_at"`

	// Deployment completion time
	// @Format: datetime
	CompletedAt *string `json:"completed_at"`

	// Deployment metadata
	Metadata map[string]any `json:"metadata"`
}

// Plugin represents a platform plugin
// @Description Plugin represents a platform plugin
type Plugin struct {
	// Unique identifier of the plugin
	// @ReadOnly: true
	ID uint `json:"id"`

	// Name of the plugin
	// @Example: rate-limiter
	Name string `json:"name"`

	// Version of the plugin
	// @Example: v1.0.0
	Version string `json:"version"`

	// Author of the plugin
	// @Example: Platform Team
	Author string `json:"author"`

	// Description of the plugin
	// @Example: Rate limiting plugin for API endpoints
	Description string `json:"description"`

	// Whether the plugin is enabled
	// @Example: true
	Enabled bool `json:"enabled"`

	// Plugin configuration
	// @Example: {"rate_limit": 100, "window": "1m"}
	Config map[string]any `json:"config"`

	// Plugin installation time
	// @Format: datetime
	CreatedAt string `json:"created_at"`

	// Plugin last update time
	// @Format: datetime
	UpdatedAt string `json:"updated_at"`
}

// Automation represents a workflow automation
// @Description Automation represents a workflow automation
type Automation struct {
	// Unique identifier of the automation
	// @ReadOnly: true
	ID uint `json:"id"`

	// Name of the automation
	// @Example: Daily Backup
	Name string `json:"name"`

	// Description of the automation
	// @Example: Backup database daily at midnight
	Description string `json:"description"`

	// Trigger type for the automation
	// @Enum: manual,schedule,webhook
	// @Example: schedule
	Trigger string `json:"trigger"`

	// Cron schedule for the automation
	// @Example: 0 0 * * *
	Schedule string `json:"schedule"`

	// Whether the automation is enabled
	// @Example: true
	Enabled bool `json:"enabled"`

	// Automation configuration
	// @Example: {"backup_path": "/backups", "retention": "7d"}
	Config map[string]any `json:"config"`

	// Automation creation time
	// @Format: datetime
	CreatedAt string `json:"created_at"`

	// Automation last update time
	// @Format: datetime
	UpdatedAt string `json:"updated_at"`
}

// Payment represents a payment transaction
// @Description Payment represents a payment transaction
type Payment struct {
	// Unique identifier of the payment
	// @ReadOnly: true
	ID uint `json:"id"`

	// Transaction ID
	// @Example: txn_123456789
	TransactionID string `json:"transaction_id"`

	// Invoice ID associated with the payment
	// @Example: inv_123456789
	InvoiceID string `json:"invoice_id"`

	// Payment method used
	// @Enum: credit_card,debit_card,bank_transfer,paypal
	// @Example: credit_card
	Method string `json:"method"`

	// Payment amount
	// @Example: 29.99
	Amount float64 `json:"amount"`

	// Payment status
	// @Enum: pending,completed,failed,refunded
	// @Example: completed
	Status string `json:"status"`

	// Payment creation time
	// @Format: datetime
	CreatedAt string `json:"created_at"`
}

// ConfigEntry represents a configuration entry
// @Description ConfigEntry represents a configuration entry
type ConfigEntry struct {
	// Configuration key
	// @Example: api_base_url
	Key string `json:"key"`

	// Configuration value
	// @Example: http://localhost:8000
	Value string `json:"value"`

	// Configuration description
	// @Example: Base URL for API endpoints
	Description string `json:"description"`

	// Whether the configuration is encrypted
	// @Example: false
	Encrypted bool `json:"encrypted"`
}

// HealthStatus represents system health status
// @Description HealthStatus represents system health status
type HealthStatus struct {
	// Overall health status
	// @Enum: healthy,degraded,unhealthy
	// @Example: healthy
	Status string `json:"status"`

	// List of service health statuses
	Services []ServiceHealth `json:"services"`

	// Health check timestamp
	// @Format: datetime
	Timestamp string `json:"timestamp"`
}

// ServiceHealth represents individual service health
// @Description ServiceHealth represents individual service health
type ServiceHealth struct {
	// Service name
	// @Example: api-server
	Name string `json:"name"`

	// Service health status
	// @Enum: healthy,degraded,unhealthy
	// @Example: healthy
	Status string `json:"status"`

	// Service response time in milliseconds
	// @Example: 45
	ResponseTime int `json:"response_time"`

	// Last health check time
	// @Format: datetime
	LastCheck string `json:"last_check"`
}

// UsageEvent represents a usage event
// @Description UsageEvent represents a usage event
type UsageEvent struct {
	// Event ID
	// @Example: evt_123456789
	ID string `json:"id"`

	// Subscriber IMSI
	// @Example: 208930000000001
	SubscriberIMSI string `json:"subscriber_imsi"`

	// Type of usage
	// @Enum: data,voice,sms
	// @Example: data
	Type string `json:"type"`

	// Amount of usage (bytes, seconds, or count)
	// @Example: 1048576
	Amount int64 `json:"amount"`

	// Event timestamp
	// @Format: datetime
	Timestamp string `json:"timestamp"`
}

// MetricSample represents a metric data point
// @Description MetricSample represents a metric data point
type MetricSample struct {
	// Metric timestamp
	// @Format: datetime
	Timestamp string `json:"timestamp"`

	// Metric value
	// @Example: 42.5
	Value float64 `json:"value"`
}

// Alert represents a monitoring alert
// @Description Alert represents a monitoring alert
type Alert struct {
	// Alert name
	// @Example: High CPU Usage
	Name string `json:"name"`

	// Alert severity
	// @Enum: critical,warning,info
	// @Example: warning
	Severity string `json:"severity"`

	// Alert state
	// @Enum: firing,resolved
	// @Example: firing
	State string `json:"state"`

	// Alert summary
	// @Example: CPU usage is above 80%
	Summary string `json:"summary"`

	// Alert description
	// @Example: CPU usage has been above 80% for more than 5 minutes
	Description string `json:"description"`

	// Alert start time
	// @Format: datetime
	StartsAt string `json:"starts_at"`

	// Alert end time
	// @Format: datetime
	EndsAt *string `json:"ends_at"`
}

// AutomationRun represents an automation execution run
// @Description AutomationRun represents an automation execution run
type AutomationRun struct {
	// Run ID
	// @Example: run_123456789
	ID uint `json:"id"`

	// Automation ID
	// @Example: 1
	AutomationID uint `json:"automation_id"`

	// Run status
	// @Enum: completed,failed,running,pending
	// @Example: completed
	Status string `json:"status"`

	// Run start time
	// @Format: datetime
	StartedAt string `json:"started_at"`

	// Run completion time
	// @Format: datetime
	CompletedAt *string `json:"completed_at"`

	// Run output
	// @Example: Backup completed successfully
	Output string `json:"output"`

	// Run error (if any)
	// @Example: Backup failed: insufficient disk space
	Error string `json:"error"`
}

// APIResponse represents a standard API response
// @Description APIResponse represents a standard API response
type APIResponse struct {
	// Response data
	Data any `json:"data"`

	// Response message
	Message string `json:"message"`

	// Success status
	Success bool `json:"success"`
}
