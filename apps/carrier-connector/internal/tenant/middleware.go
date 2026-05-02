package tenant

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/id"
	"github.com/sirupsen/logrus"
)

// TenantMiddleware provides tenant isolation middleware
type TenantMiddleware struct {
	tenantService Service
	logger        *logrus.Logger
}

// NewTenantMiddleware creates a new tenant middleware
func NewTenantMiddleware(tenantService Service, logger *logrus.Logger) *TenantMiddleware {
	return &TenantMiddleware{
		tenantService: tenantService,
		logger:        logger,
	}
}

// TenantContextKey is the key used to store tenant context in the request context
type TenantContextKey string

const (
	TenantCtxKey TenantContextKey = "tenant_context"
	UserIDKey    TenantContextKey = "user_id"
)

// ExtractTenantFromHeader extracts tenant information from HTTP headers
func (tm *TenantMiddleware) ExtractTenantFromHeader(c *gin.Context) (*TenantContext, error) {
	// Try to get tenant from X-Tenant-ID header
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		// Try to get tenant from subdomain
		host := c.Request.Host
		parts := strings.Split(host, ".")
		if len(parts) > 2 {
			tenantID = parts[0]
		}
	}

	if tenantID == "" {
		return nil, errors.New("tenant ID not found in request")
	}

	// Get tenant context
	tenantCtx, err := tm.tenantService.GetTenantContext(c.Request.Context(), tenantID)
	if err != nil {
		return nil, err
	}

	return tenantCtx, nil
}

// ExtractTenantFromAPIKey extracts tenant information from API key
func (tm *TenantMiddleware) ExtractTenantFromAPIKey(c *gin.Context) (*TenantContext, error) {
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		// Try Authorization header with Bearer token
		authHeader := c.GetHeader("Authorization")
		if after, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
			apiKey = after
		}
	}

	if apiKey == "" {
		return nil, errors.New("API key not found in request")
	}

	// Validate API key
	apiKeyData, err := tm.tenantService.ValidateAPIKey(c.Request.Context(), apiKey)
	if err != nil {
		return nil, err
	}

	// Get tenant context
	tenantCtx, err := tm.tenantService.GetTenantContext(c.Request.Context(), apiKeyData.TenantID)
	if err != nil {
		return nil, err
	}

	return tenantCtx, nil
}

// RequireTenant middleware ensures a valid tenant is present
func (tm *TenantMiddleware) RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantCtx, err := tm.ExtractTenantFromHeader(c)
		if err != nil {
			// Try API key extraction
			tenantCtx, err = tm.ExtractTenantFromAPIKey(c)
			if err != nil {
				tm.logger.WithError(err).Error("Failed to extract tenant context")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid tenant or API key"})
				c.Abort()
				return
			}
		}

		// Validate tenant access
		_, err = tm.tenantService.ValidateTenantAccess(c.Request.Context(), tenantCtx.TenantID, "")
		if err != nil {
			tm.logger.WithError(err).Error("Tenant access validation failed")
			c.JSON(http.StatusForbidden, gin.H{"error": "Tenant access denied"})
			c.Abort()
			return
		}

		// Inject tenant context
		ctx := context.WithValue(c.Request.Context(), TenantCtxKey, tenantCtx)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequireTenantRole middleware ensures user has required role
func (tm *TenantMiddleware) RequireTenantRole(requiredRoles ...TenantRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantCtx, err := tm.GetTenantContext(c)
		if err != nil {
			tm.logger.WithError(err).Error("Failed to get tenant context")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid tenant context"})
			c.Abort()
			return
		}

		// Check if user has required role
		hasRole := slices.Contains(requiredRoles, tenantCtx.UserRole)

		if !hasRole {
			tm.logger.WithFields(logrus.Fields{
				"tenant_id":      tenantCtx.TenantID,
				"user_role":      tenantCtx.UserRole,
				"required_roles": requiredRoles,
			}).Error("Insufficient tenant role")
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission middleware ensures user has specific permission
func (tm *TenantMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantCtx, err := tm.GetTenantContext(c)
		if err != nil {
			tm.logger.WithError(err).Error("Failed to get tenant context")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid tenant context"})
			c.Abort()
			return
		}

		userID := tm.GetUserID(c)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
			c.Abort()
			return
		}

		// Check permission
		hasPermission, err := tm.tenantService.HasPermission(c.Request.Context(), tenantCtx.TenantID, userID, permission)
		if err != nil {
			tm.logger.WithError(err).Error("Permission check failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Permission check failed"})
			c.Abort()
			return
		}

		if !hasPermission {
			tm.logger.WithFields(logrus.Fields{
				"tenant_id":  tenantCtx.TenantID,
				"user_id":    userID,
				"permission": permission,
			}).Error("Permission denied")
			c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimit middleware applies rate limiting per tenant
func (tm *TenantMiddleware) RateLimit(endpoint string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement rate limiting per tenant
		c.Next()
	}
}

// GetTenantContext retrieves tenant context from gin context
func (tm *TenantMiddleware) GetTenantContext(c *gin.Context) (*TenantContext, error) {
	tenantCtx, exists := c.Get(string(TenantCtxKey))
	if !exists {
		return nil, errors.New("tenant context not found")
	}

	ctx, ok := tenantCtx.(*TenantContext)
	if !ok {
		return nil, errors.New("invalid tenant context type")
	}

	return ctx, nil
}

// GetUserID retrieves user ID from gin context
func (tm *TenantMiddleware) GetUserID(c *gin.Context) string {
	userID, exists := c.Get(string(UserIDKey))
	if !exists {
		return ""
	}

	id, ok := userID.(string)
	if !ok {
		return ""
	}

	return id
}

// InjectTenantContext injects tenant context into gin context
func (tm *TenantMiddleware) InjectTenantContext(c *gin.Context, tenantCtx *TenantContext) {
	c.Set(string(TenantCtxKey), tenantCtx)
}

// TenantIsolation middleware ensures data isolation between tenants
func (tm *TenantMiddleware) TenantIsolation() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantCtx, err := tm.GetTenantContext(c)
		if err != nil {
			tm.logger.WithError(err).Error("Failed to get tenant context for isolation")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid tenant context"})
			c.Abort()
			return
		}

		// Inject tenant ID into request context for repository layer
		ctx := context.WithValue(c.Request.Context(), "tenant_id", tenantCtx.TenantID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// ValidateResourceAccess validates that tenant can access specific resource
func (tm *TenantMiddleware) ValidateResourceAccess(resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get resource ID from URL parameters
		resourceID := c.Param("id")
		if resourceID == "" {
			resourceID = c.Param("resource_id")
		}

		// Validate resource access - TODO: Implement resource access validation
		// For now, allow all resource access within tenant
		c.Next()
	}
}

// LogTenantActivity logs tenant activity for audit purposes
func (tm *TenantMiddleware) LogTenantActivity(activity string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantCtx, err := tm.GetTenantContext(c)
		if err != nil {
			// If we can't get tenant context, skip logging
			c.Next()
			return
		}

		userID := tm.GetUserID(c)

		// Log tenant event
		event := &TenantEvent{
			ID:        id.GenerateEventID(),
			TenantID:  tenantCtx.TenantID,
			UserID:    userID,
			EventType: TenantEventType(activity),
			EventData: map[string]any{
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"user_agent": c.Request.UserAgent(),
				"ip_address": c.ClientIP(),
			},
			Timestamp: id.GetCurrentTime(),
		}

		if err := tm.tenantService.LogTenantEvent(c.Request.Context(), event); err != nil {
			tm.logger.WithError(err).Error("Failed to log tenant event")
		}

		c.Next()
	}
}

// Helper functions
