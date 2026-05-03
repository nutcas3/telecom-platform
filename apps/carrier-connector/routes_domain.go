package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/handlers"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/repository"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/services"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/smdp"
)

// registerRatePlanRoutes registers rate plan CRUD, subscription, and management routes.
func registerRatePlanRoutes(api *gin.RouterGroup, repo repository.Repository, logger *logrus.Logger) {
	rateplanGroup := api.Group("/rateplans")

	ratePlanService := services.NewService(repo, logger)
	ratePlanAdapter := services.NewRatePlanAdapter(ratePlanService)
	ratePlanHandler := handlers.NewRatePlanHandler(ratePlanAdapter)

	rateplanGroup.POST("", ratePlanHandler.CreateRatePlan)
	rateplanGroup.GET("", ratePlanHandler.ListRatePlans)
	rateplanGroup.GET("/:id", ratePlanHandler.GetRatePlan)
	rateplanGroup.PUT("/:id", ratePlanHandler.UpdateRatePlan)
	rateplanGroup.DELETE("/:id", ratePlanHandler.DeleteRatePlan)
	rateplanGroup.GET("/search", ratePlanHandler.SearchRatePlans)

	rateplanGroup.POST("/subscribe", ratePlanHandler.SubscribeToPlan)
	rateplanGroup.GET("/subscriptions", ratePlanHandler.ListSubscriptions)
	rateplanGroup.GET("/subscriptions/:id", ratePlanHandler.GetSubscription)
	rateplanGroup.GET("/subscriptions/active", ratePlanHandler.GetActiveSubscription)
	rateplanGroup.DELETE("/subscriptions/:id", ratePlanHandler.CancelSubscription)

	rateplanGroup.GET("/dashboard", ratePlanHandler.GetManagementDashboard)
	rateplanGroup.GET("/overview", ratePlanHandler.GetSystemOverview)
	rateplanGroup.POST("/bulk", ratePlanHandler.BulkCreateRatePlans)
	rateplanGroup.PUT("/:id/activate", ratePlanHandler.ActivateRatePlan)
	rateplanGroup.PUT("/:id/deactivate", ratePlanHandler.DeactivateRatePlan)
	rateplanGroup.POST("/:id/duplicate", ratePlanHandler.DuplicateRatePlan)
}

// registerPricingRoutes registers pricing rule management routes.
func registerPricingRoutes(api *gin.RouterGroup, logger *logrus.Logger) {
	pricingGroup := api.Group("/pricing")

	pricingService := services.NewPricingService(nil, nil, nil, logger)
	pricingHandler := handlers.NewPricingHandler(pricingService)

	pricingGroup.POST("/rules", pricingHandler.CreateRule)
	pricingGroup.GET("/rules", pricingHandler.ListRules)
	pricingGroup.GET("/rules/:id", pricingHandler.GetRule)
	pricingGroup.PUT("/rules/:id", pricingHandler.UpdateRule)
	pricingGroup.DELETE("/rules/:id", pricingHandler.DeleteRule)
}

// registerCurrencyRoutes registers currency conversion and billing routes.
func registerCurrencyRoutes(api *gin.RouterGroup) {
	currencyGroup := api.Group("/currency")

	currencyGroup.POST("/convert", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Currency conversion endpoint ready", "status": "placeholder"})
	})
	currencyGroup.GET("/exchange/:from/:to", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Exchange rate endpoint ready", "status": "placeholder"})
	})
	currencyGroup.GET("/exchange/:from/:to/history", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Exchange rate history endpoint ready", "status": "placeholder"})
	})
	currencyGroup.GET("/currencies", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Supported currencies endpoint ready", "status": "placeholder"})
	})
	currencyGroup.POST("/exchange/refresh", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Exchange rate refresh endpoint ready", "status": "placeholder"})
	})

	currencyGroup.POST("/billing", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Billing processing endpoint ready", "status": "placeholder"})
	})
	currencyGroup.GET("/billing/history/:profile_id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Billing history endpoint ready", "status": "placeholder"})
	})
	currencyGroup.GET("/billing/summary/:profile_id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Billing summary endpoint ready", "status": "placeholder"})
	})
	currencyGroup.POST("/billing/refund/:transaction_id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Billing refund endpoint ready", "status": "placeholder"})
	})
	currencyGroup.GET("/billing/analytics", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Billing analytics endpoint ready", "status": "placeholder"})
	})
}

// registerTenantRoutes registers tenant management, user, API key, and metrics routes.
func registerTenantRoutes(api *gin.RouterGroup, db *gorm.DB, logger *logrus.Logger) {
	tenant := api.Group("/tenants")

	tenantRepo := repository.NewGormTenantRepository(db)
	tenantService := services.NewTenantService(tenantRepo, nil, logger)
	tenantHandler := handlers.NewTenantHandler(tenantService, logger)

	tenant.POST("", tenantHandler.CreateTenant)
	tenant.GET("", tenantHandler.ListTenants)
	tenant.GET("/:id", tenantHandler.GetTenant)
	tenant.GET("/domain/:domain", tenantHandler.GetTenantByDomain)
	tenant.PUT("/:id", tenantHandler.UpdateTenant)
	tenant.DELETE("/:id", tenantHandler.DeleteTenant)

	tenant.POST("/:id/users", tenantHandler.AddUserToTenant)
	tenant.GET("/:id/users", tenantHandler.ListTenantUsers)
	tenant.GET("/:id/users/:user_id", tenantHandler.GetTenantUser)
	tenant.PUT("/:id/users/:user_id", tenantHandler.UpdateTenantUser)
	tenant.DELETE("/:id/users/:user_id", tenantHandler.RemoveUserFromTenant)

	tenant.POST("/:id/apikeys", tenantHandler.CreateAPIKey)
	tenant.GET("/:id/apikeys", tenantHandler.ListAPIKeys)
	tenant.GET("/:id/apikeys/:key_id", tenantHandler.GetAPIKey)
	tenant.PUT("/:id/apikeys/:key_id", tenantHandler.UpdateAPIKey)
	tenant.DELETE("/:id/apikeys/:key_id", tenantHandler.DeleteAPIKey)

	tenant.GET("/:id/usage", tenantHandler.GetUsageStats)
	tenant.GET("/:id/quota", tenantHandler.GetQuotaStatus)
	tenant.GET("/:id/config", tenantHandler.GetTenantConfig)
	tenant.PUT("/:id/config", tenantHandler.UpdateTenantConfig)
	tenant.GET("/:id/metrics", tenantHandler.GetTenantMetrics)
	tenant.GET("/:id/events", tenantHandler.GetTenantEvents)
}

// registerSMDPRoutes registers SMDP management and carrier selection routes.
func registerSMDPRoutes(api *gin.RouterGroup, profileRepo repository.ProfileRepository) {
	smdpConfig := smdp.DefaultManagerConfig()
	smdpManager := smdp.NewSMDPManager(profileRepo.(*repository.PostgresProfileStore), smdpConfig)

	smdpGroup := api.Group("/smdp")
	smdpHandler := handlers.NewSMDPHandler(smdpManager)
	smdpGroup.DELETE("/carriers/:carrier_id", smdpHandler.RemoveCarrier)
	smdpGroup.GET("/carriers/:carrier_id/history", smdpHandler.GetSelectionHistory)
	smdpGroup.PUT("/carriers/:carrier_id/learning", smdpHandler.UpdateLearning)

	selection := api.Group("/selection")
	selectionHandler := handlers.NewSelectionHandler(smdpManager)
	selection.POST("/optimal", func(c *gin.Context) {
		selectionHandler.SelectOptimalCarrier(c.Writer, c.Request)
	})
	selection.GET("/default", func(c *gin.Context) {
		selectionHandler.SelectCarrier(c.Writer, c.Request)
	})
	selection.GET("/analytics", func(c *gin.Context) {
		selectionHandler.GetSelectionAnalytics(c.Writer, c.Request)
	})
}
