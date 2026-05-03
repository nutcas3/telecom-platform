package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/repository"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// APIIntegrationTestSuite tests API integration
type APIIntegrationTestSuite struct {
	suite.Suite
	db     *gorm.DB
	router *gin.Engine
	logger *logrus.Logger
}

// SetupSuite sets up the test suite
func (suite *APIIntegrationTestSuite) SetupSuite() {
	// Setup in-memory database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(suite.T(), err)

	// Auto-migrate tables
	err = db.AutoMigrate(
		&repository.RatePlanUsage{},
		&repository.RatePlanSubscription{},
	)
	require.NoError(suite.T(), err)

	suite.db = db
	suite.logger = logrus.New()
	suite.logger.SetLevel(logrus.ErrorLevel)

	// Setup router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	suite.setupRoutes()
}

func (suite *APIIntegrationTestSuite) setupRoutes() {
	api := suite.router.Group("/api/v1")

	// Rate plan usage routes
	usage := api.Group("/usage")
	{
		usage.POST("", suite.createUsage)
		usage.GET("/:id", suite.getUsage)
		usage.GET("", suite.listUsage)
	}

	// Subscription routes
	subscriptions := api.Group("/subscriptions")
	{
		subscriptions.POST("", suite.createSubscription)
		subscriptions.GET("/:id", suite.getSubscription)
		subscriptions.PUT("/:id/cancel", suite.cancelSubscription)
	}

	// Health check
	api.GET("/health", suite.healthCheck)
}

// TestRatePlanUsageLifecycle tests usage tracking
func (suite *APIIntegrationTestSuite) TestRatePlanUsageLifecycle() {
	// 1. Create a rate plan usage record
	usageReq := map[string]interface{}{
		"rate_plan_id": "plan-123",
		"profile_id":   "profile-456",
		"cycle_start":  "2026-01-01T00:00:00Z",
		"cycle_end":    "2026-01-31T23:59:59Z",
		"data_used":    int64(1024),
		"voice_used":   int64(300),
		"sms_used":     int64(50),
	}

	reqBody, _ := json.Marshal(usageReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/usage", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var usage repository.RatePlanUsage
	err := json.Unmarshal(w.Body.Bytes(), &usage)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "plan-123", usage.RatePlanID)
	assert.Equal(suite.T(), "profile-456", usage.ProfileID)
	assert.Equal(suite.T(), int64(1024), usage.DataUsed)

	usageID := usage.ID

	// 2. Get the usage record
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/usage/%s", usageID), nil)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var retrievedUsage repository.RatePlanUsage
	err = json.Unmarshal(w.Body.Bytes(), &retrievedUsage)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), usageID, retrievedUsage.ID)

	// 3. List usage records
	req = httptest.NewRequest(http.MethodGet, "/api/v1/usage", nil)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var usageList []repository.RatePlanUsage
	err = json.Unmarshal(w.Body.Bytes(), &usageList)
	require.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(usageList), 0)
}

// TestSubscriptionLifecycle tests subscription management
func (suite *APIIntegrationTestSuite) TestSubscriptionLifecycle() {
	// 1. Create a subscription
	subReq := map[string]interface{}{
		"profile_id":        "profile-789",
		"rate_plan_id":      "plan-abc",
		"status":            "active",
		"billing_cycle":     "monthly",
		"next_billing_date": "2026-02-01T00:00:00Z",
		"auto_renew":        true,
		"applied_discounts": []string{"new-user-10"},
	}

	reqBody, _ := json.Marshal(subReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var subscription repository.RatePlanSubscription
	err := json.Unmarshal(w.Body.Bytes(), &subscription)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "profile-789", subscription.ProfileID)
	assert.Equal(suite.T(), "plan-abc", subscription.RatePlanID)
	assert.Equal(suite.T(), repository.SubscriptionStatus("active"), subscription.Status)

	subscriptionID := subscription.ID

	// 2. Get the subscription
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/subscriptions/%s", subscriptionID), nil)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var retrievedSubscription repository.RatePlanSubscription
	err = json.Unmarshal(w.Body.Bytes(), &retrievedSubscription)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), subscriptionID, retrievedSubscription.ID)

	// 3. Cancel the subscription
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/subscriptions/%s/cancel", subscriptionID), nil)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// 4. Verify subscription is cancelled
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/subscriptions/%s", subscriptionID), nil)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var cancelledSubscription repository.RatePlanSubscription
	err = json.Unmarshal(w.Body.Bytes(), &cancelledSubscription)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), repository.SubscriptionStatus("cancelled"), cancelledSubscription.Status)
}

// TestHealthCheck tests the health endpoint
func (suite *APIIntegrationTestSuite) TestHealthCheck() {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var health map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &health)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "healthy", health["status"])
	assert.Contains(suite.T(), health, "timestamp")
}

// TestErrorHandling tests API error handling
func (suite *APIIntegrationTestSuite) TestErrorHandling() {
	// Test getting non-existent usage
	req := httptest.NewRequest(http.MethodGet, "/api/v1/usage/non-existent", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var errorResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), errorResponse, "error")

	// Test invalid request body
	reqBody := []byte("{invalid json}")
	req = httptest.NewRequest(http.MethodPost, "/api/v1/usage", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestConcurrentRequests tests concurrent API requests
func (suite *APIIntegrationTestSuite) TestConcurrentRequests() {
	// Create some test data first
	usage := repository.RatePlanUsage{
		RatePlanID: "test-plan",
		ProfileID:  "test-profile",
		CycleStart: time.Now(),
		CycleEnd:   time.Now().Add(30 * 24 * time.Hour),
		DataUsed:   100,
		VoiceUsed:  50,
		SMSUsed:    10,
	}
	err := suite.db.Create(&usage).Error
	require.NoError(suite.T(), err)

	// Test concurrent requests
	const numGoroutines = 10
	const numRequests = 100

	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numRequests/numGoroutines; j++ {
				req := httptest.NewRequest(http.MethodGet, "/api/v1/usage", nil)
				w := httptest.NewRecorder()
				suite.router.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					errChan <- fmt.Errorf("unexpected status code: %d", w.Code)
					return
				}
			}
			errChan <- nil
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		err := <-errChan
		assert.NoError(suite.T(), err)
	}
}

// TestRatePlanUsageValidation tests input validation
func (suite *APIIntegrationTestSuite) TestRatePlanUsageValidation() {
	// Test missing required fields
	invalidReq := map[string]interface{}{
		"profile_id": "profile-123",
		// Missing rate_plan_id
	}

	reqBody, _ := json.Marshal(invalidReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/usage", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var errorResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), errorResponse, "error")

	// Test invalid data types
	invalidReq2 := map[string]interface{}{
		"rate_plan_id": "plan-123",
		"profile_id":   "profile-123",
		"data_used":    "not-a-number", // Should be int64
	}

	reqBody, _ = json.Marshal(invalidReq2)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/usage", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// Handler implementations
func (suite *APIIntegrationTestSuite) createUsage(c *gin.Context) {
	var usage repository.RatePlanUsage
	if err := c.ShouldBindJSON(&usage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate ID if not provided
	if usage.ID == "" {
		usage.ID = fmt.Sprintf("usage-%d", time.Now().UnixNano())
	}

	// Set timestamps
	now := time.Now()
	usage.LastUpdated = now

	if err := suite.db.Create(&usage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, usage)
}

func (suite *APIIntegrationTestSuite) getUsage(c *gin.Context) {
	var usage repository.RatePlanUsage
	if err := suite.db.First(&usage, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "usage record not found"})
		return
	}

	c.JSON(http.StatusOK, usage)
}

func (suite *APIIntegrationTestSuite) listUsage(c *gin.Context) {
	var usage []repository.RatePlanUsage
	if err := suite.db.Find(&usage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, usage)
}

func (suite *APIIntegrationTestSuite) createSubscription(c *gin.Context) {
	var subscription repository.RatePlanSubscription
	if err := c.ShouldBindJSON(&subscription); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate ID if not provided
	if subscription.ID == "" {
		subscription.ID = fmt.Sprintf("sub-%d", time.Now().UnixNano())
	}

	// Set timestamps
	now := time.Now()
	subscription.StartedAt = now
	subscription.CreatedAt = now
	subscription.UpdatedAt = now
	subscription.CurrentCycle = now

	if err := suite.db.Create(&subscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, subscription)
}

func (suite *APIIntegrationTestSuite) getSubscription(c *gin.Context) {
	var subscription repository.RatePlanSubscription
	if err := suite.db.First(&subscription, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

func (suite *APIIntegrationTestSuite) cancelSubscription(c *gin.Context) {
	if err := suite.db.Model(&repository.RatePlanSubscription{}).Where("id = ?", c.Param("id")).Updates(map[string]interface{}{
		"status":     "cancelled",
		"ended_at":   time.Now(),
		"updated_at": time.Now(),
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
}

func (suite *APIIntegrationTestSuite) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	})
}

// TearDownSuite cleans up the test suite
func (suite *APIIntegrationTestSuite) TearDownSuite() {
	sqlDB, err := suite.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// TestAPIIntegrationSuite runs the integration test suite
func TestAPIIntegrationSuite(t *testing.T) {
	suite.Run(t, new(APIIntegrationTestSuite))
}
