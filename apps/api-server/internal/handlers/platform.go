// Package handlers hosts HTTP handler types split by domain:
//
//	platform_services.go     - ServicesHandler (Kubernetes services)
//	platform_monitoring.go   - MonitoringHandler (Prometheus metrics)
//	platform_deployments.go  - DeploymentsHandler (deployment lifecycle)
//	platform_plugins.go      - PluginsHandler (plugin lifecycle)
//	platform_automation.go   - AutomationHandler (automations, schedules)
//	platform_config.go       - ConfigHandler (runtime config)
//	platform_chaos.go        - ChaosHandler (chaos experiments)
//	platform_billing.go      - BillingHandler (invoices, payments)
//
// This file holds only the shared ErrorResponse envelope and small helpers.
package handlers

import (
"net/http"
"strconv"

"github.com/gin-gonic/gin"
)

// ErrorResponse is a shared error envelope used across platform handlers.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

func badRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request", Details: err.Error(), Code: "BAD_REQUEST"})
}

func serverError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Internal error", Details: err.Error(), Code: "INTERNAL_ERROR"})
}

func notFound(c *gin.Context, err error) {
	c.JSON(http.StatusNotFound, ErrorResponse{Error: "Not found", Details: err.Error(), Code: "NOT_FOUND"})
}

// parseUintParam parses a uint route parameter and writes a 400 on failure.
func parseUintParam(c *gin.Context, name string) (uint, bool) {
	v, err := strconv.ParseUint(c.Param(name), 10, 32)
	if err != nil {
		badRequest(c, err)
		return 0, false
	}
	return uint(v), true
}

// parseBoolQuery returns *bool for optional tri-state filters.
func parseBoolQuery(c *gin.Context, key string) *bool {
	raw := c.Query(key)
	if raw == "" {
		return nil
	}
	b, err := strconv.ParseBool(raw)
	if err != nil {
		return nil
	}
	return &b
}
