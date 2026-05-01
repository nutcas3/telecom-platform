package integration

import (
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/handlers"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/repository"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/service"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/smdp"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SMDPIntegration integrates the multi-carrier SM-DP+ system with the existing carrier connector
type SMDPIntegration struct {
	service    *service.SMDPService
	handler    *handlers.SMDPHandler
	repository *repository.PostgresProfileStore
	logger     *logrus.Logger
}

// NewSMDPIntegration creates a new SM-DP+ integration
func NewSMDPIntegration(repo *repository.PostgresProfileStore) *SMDPIntegration {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	svc := service.NewSMDPService(repo)
	hnd := handlers.NewSMDPHandler(repo)

	integration := &SMDPIntegration{
		service:    svc,
		handler:    hnd,
		repository: repo,
		logger:     logger,
	}

	// Initialize default carriers
	if err := integration.InitializeSystem(); err != nil {
		logger.WithError(err).Error("Failed to initialize SM-DP+ integration")
	}

	return integration
}

// InitializeSystem sets up the SM-DP+ system with default carriers and configuration
func (i *SMDPIntegration) InitializeSystem() error {
	i.logger.Info("Initializing SM-DP+ Multi-Carrier Integration")

	// Initialize default carriers
	if err := i.service.InitializeDefaultCarriers(); err != nil {
		return fmt.Errorf("failed to initialize default carriers: %w", err)
	}

	// Log system status
	carriers := i.service.GetCarrierHealth()
	i.logger.WithField("carrier_count", len(carriers)).Info("SM-DP+ system initialized")

	for id, carrier := range carriers {
		i.logger.WithFields(logrus.Fields{
			"carrier_id":     id,
			"carrier_name":   carrier.Name,
			"country":        carrier.CountryCode,
			"health_status":  carrier.HealthStatus,
			"is_active":      carrier.IsActive,
		}).Info("Carrier loaded")
	}

	return nil
}

// RegisterRoutes registers SM-DP+ routes with the Gin router
func (i *SMDPIntegration) RegisterRoutes(router *gin.RouterGroup) {
	smdp := router.Group("/smdp")
	{
		// Profile management
		smdp.POST("/download", i.handler.DownloadProfile)
		smdp.POST("/status", i.handler.GetProfileStatus)

		// Carrier management
		smdp.POST("/carriers", i.handler.AddCarrier)
		smdp.DELETE("/carriers/:id", i.handler.RemoveCarrier)
		smdp.GET("/carriers/status", i.handler.GetCarrierStatus)

		// Health and metrics
		smdp.GET("/health", i.healthHandler)
		smdp.GET("/metrics", i.metricsHandler)
	}

	i.logger.Info("SM-DP+ routes registered")
}

// healthHandler returns system health status
func (i *SMDPIntegration) healthHandler(c *gin.Context) {
	carriers := i.service.GetCarrierHealth()
	
	healthyCount := 0
	totalCount := len(carriers)
	
	for _, carrier := range carriers {
		if carrier.HealthStatus == smdp.CarrierStatusHealthy {
			healthyCount++
		}
	}

	status := "healthy"
	if healthyCount == 0 {
		status = "unhealthy"
	} else if healthyCount < totalCount {
		status = "degraded"
	}

	c.JSON(200, gin.H{
		"status":           status,
		"healthy_carriers": healthyCount,
		"total_carriers":   totalCount,
		"carriers":         carriers,
	})
}

// metricsHandler returns system metrics
func (i *SMDPIntegration) metricsHandler(c *gin.Context) {
	metrics := i.service.GetCarrierMetrics()
	carriers := i.service.GetCarrierHealth()

	response := gin.H{
		"carrier_metrics": metrics,
		"carrier_details": carriers,
		"timestamp":       gin.H{},
	}

	// Calculate total metrics
	var totalRequests, totalSuccess, totalFailed uint64
	for _, metric := range metrics {
		totalRequests += metric.TotalRequests
		totalSuccess += metric.SuccessfulRequests
		totalFailed += metric.FailedRequests
	}

	response["total_metrics"] = gin.H{
		"total_requests":       totalRequests,
		"successful_requests":  totalSuccess,
		"failed_requests":      totalFailed,
		"success_rate":         float64(totalSuccess) / float64(totalRequests),
	}

	c.JSON(200, response)
}

// GetService returns the SM-DP+ service for direct access
func (i *SMDPIntegration) GetService() *service.SMDPService {
	return i.service
}

// GetHandler returns the SM-DP+ handler for direct access
func (i *SMDPIntegration) GetHandler() *handlers.SMDPHandler {
	return i.handler
}

// ExampleUsage demonstrates how to use the SM-DP+ integration
func ExampleUsage() {
	// This would typically be in your main.go or router setup
	
	/*
	// Initialize repository
	repo, err := repository.NewPostgresProfileStore("postgres://user:pass@localhost/db")
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Close()

	// Create SM-DP+ integration
	smdpIntegration := integration.NewSMDPIntegration(repo)

	// Register routes
	router := gin.Default()
	api := router.Group("/api/v1")
	smdpIntegration.RegisterRoutes(api)

	// Get service for direct use
	service := smdpIntegration.GetService()

	// Example: Download profile
	ctx := context.Background()
	req := &smdp.ProfileRequest{
		EID:         "eid-example",
		ICCID:       "iccid-example",
		ProfileType: "operational",
	}

	response, err := service.DownloadProfile(ctx, req)
	if err != nil {
		log.Printf("Profile download failed: %v", err)
		return
	}

	log.Printf("Profile downloaded successfully from carrier %s", response.CarrierID)
	*/
}
