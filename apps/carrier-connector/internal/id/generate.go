package id

import (
	"fmt"
	"time"
)

// GenerateRuleID generates a rule ID using Snowflake ID
func GenerateRuleID() string {
	return GeneratePrefixed("rule")
}

// GenerateEventID generates an event ID using Snowflake ID
func GenerateEventID() string {
	return GeneratePrefixed("evt")
}

// GenerateUsageID generates a usage ID using Snowflake ID with resourceType and tenantID
func GenerateUsageID(resourceType, tenantID string) string {
	// Create a more traceable ID that includes resource type and tenant info
	// Format: usage_<resourceType>_<tenantID_short>_<snowflake>
	tenantShort := tenantID
	if len(tenantID) > 8 {
		tenantShort = tenantID[:8]
	}
	return fmt.Sprintf("usage_%s_%s_%s", resourceType, tenantShort, GenerateString())
}

// GenerateAPIID generates an API ID using Snowflake ID
func GenerateAPIID() string {
	return GeneratePrefixed("api")
}

// GenerateTenantID generates a tenant ID using Snowflake ID
func GenerateTenantID() string {
	return GeneratePrefixed("tnt")
}

// GenerateProfileID generates a profile ID using Snowflake ID
func GenerateProfileID() string {
	return GeneratePrefixed("prf")
}

// GenerateSubscriptionID generates a subscription ID using Snowflake ID
func GenerateSubscriptionID() string {
	return GeneratePrefixed("sub")
}

// GenerateCarrierID generates a carrier ID using Snowflake ID
func GenerateCarrierID() string {
	return GeneratePrefixed("car")
}

// GeneratePricingID generates a pricing ID using Snowflake ID
func GeneratePricingID() string {
	return GeneratePrefixed("prc")
}

// GetCurrentTime returns the current time
func GetCurrentTime() time.Time {
	return time.Now()
}

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}
