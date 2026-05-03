package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/es2"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/handlers"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/mq"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/repository"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/services"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/webhook"
)

// setupRoutes registers all HTTP routes.
func setupRoutes(router *gin.Engine, client *es2.ES2Client, profileRepo repository.ProfileRepository, webhookClient *webhook.WebhookClient, messageQueue *mq.MessageQueue, repo repository.Repository, db *gorm.DB, logger *logrus.Logger) {
	api := router.Group("/api/v1")

	// Health and metrics
	api.GET("/health", healthHandler)
	api.GET("/health/ready", readinessHandler(profileRepo))
	api.GET("/health/live", livenessHandler)
	api.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// eSIM profile management
	esim := api.Group("/esim")
	esim.POST("/profiles", handlers.OrderProfileHandlerWithRepo(client, profileRepo, webhookClient, messageQueue))
	esim.GET("/profiles", handlers.ListProfilesHandler(profileRepo))
	esim.GET("/profiles/:profileId", handlers.GetProfileHandler(profileRepo))
	esim.DELETE("/profiles/:profileId", handlers.DeleteProfileHandler(profileRepo))

	// Carrier info
	carrier := api.Group("/carrier")
	carrier.GET("/info", handlers.GetCarrierInfoHandler(client))
	carrier.GET("/connectivity", handlers.CheckConnectivityHandler(client))

	// MVNO routes
	registerMVNORoutes(api, repo, logger)

	// Domain routes
	registerRatePlanRoutes(api, repo, logger)
	registerPricingRoutes(api, logger)
	registerCurrencyRoutes(api)
	registerTenantRoutes(api, db, logger)
	registerSMDPRoutes(api, profileRepo)
}

// registerMVNORoutes registers MVNO onboarding and management routes.
func registerMVNORoutes(api *gin.RouterGroup, repo repository.Repository, logger *logrus.Logger) {
	mvno := api.Group("/mvno")

	onboardingService := services.NewOnboardingService(logger)
	mvnoHandler := handlers.NewMVNOHandler(onboardingService, repo, logger)
	managementHandler := handlers.NewManagementHandler(repo, logger)

	mvno.POST("/onboarding", mvnoHandler.StartOnboarding)
	mvno.GET("", mvnoHandler.ListMVNOs)
	mvno.GET("/:id", mvnoHandler.GetMVNO)
	mvno.PUT("/:id/status", managementHandler.UpdateMVNOStatus)
	mvno.GET("/stats", managementHandler.GetMVNOStats)
}

// healthHandler returns a simple liveness response.
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "carrier-connector",
		"timestamp": time.Now().UTC(),
	})
}

// livenessHandler returns a simple liveness check.
func livenessHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"service":   "carrier-connector",
		"timestamp": time.Now().UTC(),
	})
}

// readinessHandler checks if the service is ready to accept requests.
func readinessHandler(repo repository.ProfileRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := repo.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":    "not ready",
				"service":   "carrier-connector",
				"timestamp": time.Now().UTC(),
				"error":     "database connection failed",
				"details":   err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"service":   "carrier-connector",
			"timestamp": time.Now().UTC(),
			"checks":    gin.H{"database": "ok"},
		})
	}
}
