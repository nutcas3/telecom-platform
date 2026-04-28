package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/es2"
	handler "github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/handler"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/mq"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/repository"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/webhook"
)

// setupRoutes registers all HTTP routes, wiring the ES2+ client, profile repo, webhook client, and message queue.
func setupRoutes(router *gin.Engine, client *es2.ES2Client, repo repository.ProfileRepository, webhookClient *webhook.WebhookClient, messageQueue *mq.MessageQueue) {
	api := router.Group("/api/v1")
	api.GET("/health", healthHandler)
	api.GET("/health/ready", readinessHandler(repo))
	api.GET("/health/live", livenessHandler)
	api.GET("/metrics", gin.WrapH(promhttp.Handler()))

	esim := api.Group("/esim")
	{
		esim.POST("/profiles", handler.OrderProfileHandlerWithRepo(client, repo, webhookClient, messageQueue))
		esim.GET("/profiles", handler.ListProfilesHandlerWithRepo(repo))
		esim.GET("/profiles/:profileId", handler.GetProfileHandlerWithRepo(client, repo))
		esim.DELETE("/profiles/:profileId", handler.DeleteProfileHandlerWithRepo(client, repo, webhookClient, messageQueue))
	}

	carrier := api.Group("/carrier")
	{
		carrier.GET("/info", handler.GetCarrierInfoHandler(client))
		carrier.GET("/connectivity", handler.CheckConnectivityHandler(client))
	}
}

// healthHandler returns a simple liveness response.
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "carrier-connector",
		"timestamp": time.Now().UTC(),
	})
}

// livenessHandler returns a simple liveness check (always healthy if service is running).
func livenessHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"service":   "carrier-connector",
		"timestamp": time.Now().UTC(),
	})
}

// readinessHandler checks if the service is ready to accept requests (database connectivity).
func readinessHandler(repo repository.ProfileRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check database connectivity
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
			"checks": gin.H{
				"database": "ok",
			},
		})
	}
}
