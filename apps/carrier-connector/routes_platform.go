package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/analytics"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/compliance"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/currency"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/handlers"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/infra"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/whitelabel"
)

// registerAnalyticsRoutes registers analytics dashboard routes (platform-011, 012)
func registerAnalyticsRoutes(api *gin.RouterGroup, db *gorm.DB, logger *logrus.Logger) {
	analyticsService := analytics.NewService(db, logger)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService, logger)

	group := api.Group("/analytics")
	group.GET("/dashboard", analyticsHandler.GetDashboard)
	group.GET("/revenue", analyticsHandler.GetRevenueAnalytics)
	group.POST("/reports", analyticsHandler.CreateScheduledReport)
	group.GET("/reports", analyticsHandler.ListScheduledReports)
}

// registerWhitelabelRoutes registers whitelabel partner routes (platform-009)
func registerWhitelabelRoutes(api *gin.RouterGroup, db *gorm.DB, logger *logrus.Logger) {
	whitelabelService := whitelabel.NewService(db, logger)
	whitelabelHandler := handlers.NewWhitelabelHandler(whitelabelService, logger)

	group := api.Group("/whitelabel")
	group.POST("/branding", whitelabelHandler.CreateBranding)
	group.GET("/branding", whitelabelHandler.GetBranding)
	group.GET("/branding/domain/:domain", whitelabelHandler.GetBrandingByDomain)
	group.PUT("/branding", whitelabelHandler.UpdateBranding)

	group.POST("/partner", whitelabelHandler.CreatePartnerConfig)
	group.GET("/partner", whitelabelHandler.GetPartnerConfig)

	group.POST("/templates", whitelabelHandler.CreateEmailTemplate)
	group.GET("/templates", whitelabelHandler.ListEmailTemplates)
	group.GET("/templates/:key", whitelabelHandler.GetEmailTemplate)
}

// registerComplianceRoutes registers compliance routes (platform-013, 014, 015, 016)
func registerComplianceRoutes(api *gin.RouterGroup, db *gorm.DB, logger *logrus.Logger) {
	complianceService := compliance.NewService(db, logger)
	complianceHandler := handlers.NewComplianceHandler(complianceService, logger)

	group := api.Group("/compliance")

	// Data Subject Requests (GDPR/CCPA)
	group.POST("/dsr", complianceHandler.CreateDSR)
	group.GET("/dsr", complianceHandler.ListDSRs)
	group.GET("/dsr/:id", complianceHandler.GetDSR)
	group.POST("/dsr/:id/process", complianceHandler.ProcessDSR)

	// Consent Management
	group.POST("/consent", complianceHandler.RecordConsent)
	group.GET("/consent/:subject_id", complianceHandler.GetConsents)
	group.DELETE("/consent/:subject_id", complianceHandler.RevokeConsent)

	// Audit Logging
	group.GET("/audit", complianceHandler.QueryAuditLogs)

	// Data Residency
	group.POST("/residency", complianceHandler.SetDataResidency)
	group.GET("/residency", complianceHandler.GetDataResidency)
}

// registerExchangeRateRoutes registers exchange rate routes (platform-017)
func registerExchangeRateRoutes(api *gin.RouterGroup, logger *logrus.Logger) {
	config := currency.ExchangeRateConfig{
		Provider:     currency.ProviderInternal,
		BaseCurrency: "USD",
	}
	exchangeService := currency.NewRealTimeExchangeService(config, logger)

	group := api.Group("/exchange")
	group.GET("/rate/:from/:to", func(c *gin.Context) {
		from := c.Param("from")
		to := c.Param("to")
		rate, err := exchangeService.GetRate(c.Request.Context(), from, to)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"from": from, "to": to, "rate": rate})
	})
	group.GET("/rates", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"rates":       exchangeService.GetAllRates(),
			"last_update": exchangeService.LastUpdateTime(),
		})
	})
	group.GET("/currencies", func(c *gin.Context) {
		c.JSON(200, gin.H{"currencies": exchangeService.GetSupportedCurrencies()})
	})
	group.POST("/refresh", func(c *gin.Context) {
		if err := exchangeService.RefreshRates(c.Request.Context()); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "refreshed"})
	})
}

// registerInfraRoutes registers infrastructure monitoring routes (platform-020, 026, 027)
func registerInfraRoutes(api *gin.RouterGroup, cache *infra.Cache, geoRouter *infra.GeoRouter) {
	group := api.Group("/infra")

	// Cache stats (platform-020)
	group.GET("/cache/stats", func(c *gin.Context) {
		if cache != nil {
			c.JSON(200, cache.Stats())
		} else {
			c.JSON(200, gin.H{"status": "cache not configured"})
		}
	})

	// Geographic routing (platform-026)
	group.GET("/regions", func(c *gin.Context) {
		if geoRouter != nil {
			c.JSON(200, geoRouter.GetRegions())
		} else {
			c.JSON(200, gin.H{"status": "geo routing not configured"})
		}
	})
	group.GET("/region/detect", func(c *gin.Context) {
		if geoRouter != nil {
			region, _ := geoRouter.GetRegionForIP(c.Request.Context(), c.ClientIP())
			c.JSON(200, region)
		} else {
			c.JSON(200, gin.H{"region": "default"})
		}
	})
}
