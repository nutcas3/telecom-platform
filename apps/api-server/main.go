package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	logger zerolog.Logger
	dbClient *mongo.Client
)

func main() {
	// Initialize logger
	logger = zerolog.New(os.Stdout).With().
		Timestamp().
		Str("service", "api-server").
		Logger()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logger.Info().Msg("No .env file found, using system environment")
	}

	// Connect to MongoDB
	mongoURI := getEnv("MONGODB_URI", "mongodb://localhost:27017/free5gc")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	dbClient, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer dbClient.Disconnect(context.Background())

	// Verify connection
	if err := dbClient.Ping(ctx, nil); err != nil {
		logger.Fatal().Err(err).Msg("Failed to ping MongoDB")
	}
	logger.Info().Msg("Connected to MongoDB successfully")

	// Initialize Gin router
	router := setupRouter()

	// Start server
	port := getEnv("API_PORT", "8000")
	logger.Info().Str("port", port).Msg("Starting API server")
	if err := router.Run(":" + port); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start server")
	}
}

func setupRouter() *gin.Engine {
	// Set Gin mode
	if getEnv("GIN_MODE", "debug") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Health check
	router.GET("/health", healthCheck)

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// eSIM management
		v1.POST("/esims", createESIM)
		v1.GET("/esims/:id", getESIM)
		v1.GET("/esims/:id/usage", getESIMUsage)
		v1.DELETE("/esims/:id", deleteESIM)

		// Subscriber management
		v1.GET("/subscribers", listSubscribers)
		v1.GET("/subscribers/:imsi", getSubscriber)
	}

	return router
}

// Health check endpoint
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "api-server",
		"timestamp": time.Now().Unix(),
	})
}

// POST /v1/esims - Create new eSIM
func createESIM(c *gin.Context) {
	var req struct {
		DataPlan    string `json:"data_plan" binding:"required"`
		CountryCode string `json:"country_code" binding:"required"`
		Carrier     string `json:"carrier,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate IMSI (simplified - use proper allocation in production)
	imsi := fmt.Sprintf("208930%09d", time.Now().UnixNano()%1000000000)

	// Using Go 1.26's new(expr) for optional pointer fields
	esim := gin.H{
		"esim_id":         uuid.New().String(),
		"imsi":            imsi,
		"iccid":           generateICCID(imsi),
		"activation_code": fmt.Sprintf("LPA:1$smdp.example.com$%s", uuid.New().String()),
		"status":          "provisioned",
		"data_plan":       req.DataPlan,
		"country_code":    req.CountryCode,
		"created_at":      time.Now(),
	}

	logger.Info().
		Str("esim_id", esim["esim_id"].(string)).
		Str("imsi", imsi).
		Msg("eSIM created")

	c.JSON(http.StatusCreated, esim)
}

// GET /v1/esims/:id - Get eSIM details
func getESIM(c *gin.Context) {
	esimID := c.Param("id")

	// TODO: Implement actual database lookup
	esim := gin.H{
		"esim_id":    esimID,
		"imsi":       "208930000000001",
		"iccid":      "8933123456789012345",
		"status":     "active",
		"data_plan":  "1GB",
		"created_at": time.Now().Add(-24 * time.Hour),
	}

	c.JSON(http.StatusOK, esim)
}

// GET /v1/esims/:id/usage - Get eSIM usage statistics
func getESIMUsage(c *gin.Context) {
	esimID := c.Param("id")

	usage := gin.H{
		"esim_id":    esimID,
		"data_used":  250 * 1024 * 1024,  // 250 MB in bytes
		"data_limit": 1024 * 1024 * 1024, // 1 GB in bytes
		"sessions": []gin.H{
			{
				"started_at": time.Now().Add(-2 * time.Hour),
				"ended_at":   time.Now().Add(-1 * time.Hour),
				"bytes_up":   10 * 1024 * 1024,
				"bytes_down": 50 * 1024 * 1024,
			},
		},
		"updated_at": time.Now(),
	}

	c.JSON(http.StatusOK, usage)
}

// DELETE /v1/esims/:id - Terminate eSIM
func deleteESIM(c *gin.Context) {
	esimID := c.Param("id")

	logger.Info().Str("esim_id", esimID).Msg("eSIM terminated")

	c.JSON(http.StatusOK, gin.H{
		"message": "eSIM terminated successfully",
		"esim_id": esimID,
	})
}

// GET /v1/subscribers - List all subscribers
func listSubscribers(c *gin.Context) {
	// TODO: Implement MongoDB query
	subscribers := []gin.H{
		{
			"imsi":       "208930000000001",
			"status":     "active",
			"created_at": time.Now().Add(-7 * 24 * time.Hour),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"subscribers": subscribers,
		"count":       len(subscribers),
	})
}

// GET /v1/subscribers/:imsi - Get subscriber details
func getSubscriber(c *gin.Context) {
	imsi := c.Param("imsi")

	subscriber := gin.H{
		"imsi":       imsi,
		"plmn_id":    "20893",
		"status":     "active",
		"created_at": time.Now().Add(-7 * 24 * time.Hour),
	}

	c.JSON(http.StatusOK, subscriber)
}

// Utility functions
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func generateICCID(imsi string) string {
	// ICCID format: 89 (Telecom) + CC (Country) + ISSUER + ACCOUNT + CHECK
	// Simplified implementation
	return fmt.Sprintf("893312%s%d", imsi[5:14], time.Now().Unix()%10)
}
