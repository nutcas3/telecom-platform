// Package handlers hosts HTTP handlers for the carrier-connector service.
//
// Profile handlers (order, get, list, delete) live in profile_handlers.go and
// take a repository.ProfileRepository so profile state can be persisted and
// queried. Carrier-level handlers (info, connectivity) that only touch the
// ES2+ client live here.
package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/es2"
)

// GetCarrierInfoHandler returns static carrier information from environment config.
func GetCarrierInfoHandler(_ *es2.ES2Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"carrier": gin.H{
				"name":        GetEnv("CARRIER_NAME", "Example Carrier"),
				"mcc":         GetEnv("CARRIER_MCC", "208"),
				"mnc":         GetEnv("CARRIER_MNC", "93"),
				"smdpUrl":     GetEnv("SMDP_URL", "https://smdp.example.com"),
				"supported":   true,
				"lastChecked": time.Now().UTC(),
			},
		})
	}
}

// CheckConnectivityHandler probes the SM-DP+ via an ES2+ GetProfileStatus call
// with synthetic identifiers. A non-error response indicates the endpoint is
// reachable and accepting authenticated requests.
func CheckConnectivityHandler(client *es2.ES2Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := &es2.GetProfileStatusRequest{EID: "connectivity-probe", ICCID: "connectivity-probe"}
		_, err := client.GetProfileStatus(context.Background(), req)

		connected := err == nil
		var errMsg any
		if err != nil {
			errMsg = err.Error()
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"connectivity": gin.H{
				"smdpConnected": connected,
				"checkedAt":     time.Now().UTC(),
				"error":         errMsg,
			},
		})
	}
}

// parsePositiveInt parses a positive integer from a string, returning fallback on failure.
func parsePositiveInt(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}
