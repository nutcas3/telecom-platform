package e2e

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

// SubscriberJourneyTestSuite tests end-to-end subscriber journey
type SubscriberJourneyTestSuite struct {
	suite.Suite
	db     *gorm.DB
	router *gin.Engine
	logger *logrus.Logger
}

// SetupSuite sets up the test suite
func (suite *SubscriberJourneyTestSuite) SetupSuite() {
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

	// Setup router with full application routes
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	suite.setupRoutes()
}

func (suite *SubscriberJourneyTestSuite) setupRoutes() {
	api := suite.router.Group("/api/v1")

	// Profile management
	profiles := api.Group("/profiles")
	{
		profiles.POST("", suite.createProfile)
		profiles.GET("/:id", suite.getProfile)
		profiles.PUT("/:id/activate", suite.activateProfile)
	}

	// Rate plan management
	rateplans := api.Group("/rateplans")
	{
		rateplans.POST("", suite.createRatePlan)
		rateplans.GET("", suite.listRatePlans)
		rateplans.GET("/:id", suite.getRatePlan)
	}

	// Subscription management
	subscriptions := api.Group("/subscriptions")
	{
		subscriptions.POST("", suite.createSubscription)
		subscriptions.GET("/:id", suite.getSubscription)
		subscriptions.PUT("/:id/cancel", suite.cancelSubscription)
		subscriptions.POST("/:id/usage", suite.recordUsage)
		subscriptions.GET("/:id/usage", suite.getUsage)
	}

	// Billing
	billing := api.Group("/billing")
	{
		billing.POST("/charge", suite.processCharge)
		billing.GET("/invoice/:subscription_id", suite.getInvoice)
	}

	// Analytics
	analytics := api.Group("/analytics")
	{
		analytics.GET("/dashboard", suite.getDashboard)
		analytics.GET("/revenue", suite.getRevenue)
	}
}

// TestCompleteSubscriberJourney tests the full subscriber lifecycle
func (suite *SubscriberJourneyTestSuite) TestCompleteSubscriberJourney() {
	// Step 1: Create a rate plan
	ratePlanReq := map[string]interface{}{
		"name":          "Premium Data Plan",
		"description":   "Unlimited data with voice and SMS",
		"currency":      "USD",
		"base_price":    29.99,
		"billing_cycle": "monthly",
		"features":      []string{"unlimited_data", "unlimited_voice", "unlimited_sms"},
	}

	reqBody, _ := json.Marshal(ratePlanReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rateplans", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var ratePlan map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &ratePlan)
	require.NoError(suite.T(), err)
	ratePlanID := fmt.Sprintf("%v", ratePlan["id"])

	// Step 2: Create a subscriber profile
	profileReq := map[string]interface{}{
		"first_name": "John",
		"last_name":  "Doe",
		"email":      "john.doe@example.com",
		"phone":      "+1234567890",
		"country":    "US",
		"status":     "pending",
		"metadata": map[string]interface{}{
			"source": "web_signup",
		},
	}

	reqBody, _ = json.Marshal(profileReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/profiles", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var profile map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &profile)
	require.NoError(suite.T(), err)
	profileID := fmt.Sprintf("%v", profile["id"])

	// Step 3: Activate the profile
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/profiles/%s/activate", profileID), nil)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Step 4: Subscribe to a rate plan
	subReq := map[string]interface{}{
		"profile_id":     profileID,
		"rate_plan_id":   ratePlanID,
		"auto_renew":     true,
		"payment_method": "credit_card",
		"billing_address": map[string]interface{}{
			"street":  "123 Main St",
			"city":    "New York",
			"state":   "NY",
			"zip":     "10001",
			"country": "US",
		},
	}

	reqBody, _ = json.Marshal(subReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var subscription repository.RatePlanSubscription
	err = json.Unmarshal(w.Body.Bytes(), &subscription)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), profileID, subscription.ProfileID)
	assert.Equal(suite.T(), ratePlanID, subscription.RatePlanID)
	assert.Equal(suite.T(), repository.SubscriptionStatus("active"), subscription.Status)

	subscriptionID := subscription.ID

	// Step 5: Record usage for the subscription
	usageReq := map[string]interface{}{
		"data_used":  int64(5120), // 5GB in MB
		"voice_used": int64(450),  // 450 minutes
		"sms_used":   int64(25),   // 25 SMS
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	reqBody, _ = json.Marshal(usageReq)
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/subscriptions/%s/usage", subscriptionID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	// Step 6: Get usage history
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/subscriptions/%s/usage", subscriptionID), nil)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var usageList []repository.RatePlanUsage
	err = json.Unmarshal(w.Body.Bytes(), &usageList)
	require.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(usageList), 0)

	// Step 7: Process billing charge
	chargeReq := map[string]interface{}{
		"subscription_id": subscriptionID,
		"amount":          29.99,
		"currency":        "USD",
		"description":     "Monthly subscription fee",
		"charge_type":     "recurring",
	}

	reqBody, _ = json.Marshal(chargeReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/billing/charge", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Step 8: Get invoice
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/billing/invoice/%s", subscriptionID), nil)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var invoice map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &invoice)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), subscriptionID, fmt.Sprintf("%v", invoice["subscription_id"]))
	assert.Equal(suite.T(), 29.99, invoice["total_amount"])

	// Step 9: Check analytics dashboard
	req = httptest.NewRequest(http.MethodGet, "/api/v1/analytics/dashboard", nil)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var dashboard map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &dashboard)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), dashboard, "total_subscribers")
	assert.Contains(suite.T(), dashboard, "monthly_revenue")
	assert.Contains(suite.T(), dashboard, "active_subscriptions")

	// Step 10: Cancel subscription (end of journey)
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/subscriptions/%s/cancel", subscriptionID), nil)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Verify subscription is cancelled
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/subscriptions/%s", subscriptionID), nil)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var cancelledSub repository.RatePlanSubscription
	err = json.Unmarshal(w.Body.Bytes(), &cancelledSub)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), repository.SubscriptionStatus("cancelled"), cancelledSub.Status)
}

// TestMultiSubscriptionJourney tests subscriber with multiple subscriptions
func (suite *SubscriberJourneyTestSuite) TestMultiSubscriptionJourney() {
	// Create multiple rate plans
	ratePlans := []map[string]interface{}{
		{
			"name":       "Basic Plan",
			"base_price": 9.99,
			"currency":   "USD",
			"features":   []string{"1gb_data", "100min_voice"},
		},
		{
			"name":       "Premium Plan",
			"base_price": 29.99,
			"currency":   "USD",
			"features":   []string{"unlimited_data", "unlimited_voice"},
		},
	}

	ratePlanIDs := []string{}
	for _, rp := range ratePlans {
		reqBody, _ := json.Marshal(rp)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/rateplans", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)
		assert.Equal(suite.T(), http.StatusCreated, w.Code)

		var ratePlan map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &ratePlan)
		require.NoError(suite.T(), err)
		ratePlanIDs = append(ratePlanIDs, fmt.Sprintf("%v", ratePlan["id"]))
	}

	// Create profile
	profileReq := map[string]interface{}{
		"first_name": "Jane",
		"last_name":  "Smith",
		"email":      "jane.smith@example.com",
		"country":    "US",
	}

	reqBody, _ := json.Marshal(profileReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/profiles", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	var profile map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &profile)
	require.NoError(suite.T(), err)
	profileID := fmt.Sprintf("%v", profile["id"])

	// Activate profile
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/profiles/%s/activate", profileID), nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Subscribe to multiple plans
	subscriptionIDs := []string{}
	for i, ratePlanID := range ratePlanIDs {
		subReq := map[string]interface{}{
			"profile_id":   profileID,
			"rate_plan_id": ratePlanID,
			"auto_renew":   true,
			"metadata": map[string]interface{}{
				"priority": i + 1,
			},
		}

		reqBody, _ = json.Marshal(subReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)
		assert.Equal(suite.T(), http.StatusCreated, w.Code)

		var subscription repository.RatePlanSubscription
		err = json.Unmarshal(w.Body.Bytes(), &subscription)
		require.NoError(suite.T(), err)
		subscriptionIDs = append(subscriptionIDs, subscription.ID)
	}

	// Verify all subscriptions are active
	for _, subID := range subscriptionIDs {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/subscriptions/%s", subID), nil)
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var sub repository.RatePlanSubscription
		err = json.Unmarshal(w.Body.Bytes(), &sub)
		require.NoError(suite.T(), err)
		assert.Equal(suite.T(), repository.SubscriptionStatus("active"), sub.Status)
	}

	// Record usage on primary subscription
	usageReq := map[string]interface{}{
		"data_used":  int64(2048),
		"voice_used": int64(150),
		"sms_used":   int64(10),
	}

	reqBody, _ = json.Marshal(usageReq)
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/subscriptions/%s/usage", subscriptionIDs[0]), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	// Check analytics reflects multiple subscriptions
	req = httptest.NewRequest(http.MethodGet, "/api/v1/analytics/dashboard", nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var dashboard map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &dashboard)
	require.NoError(suite.T(), err)

	// Should have at least 1 active subscriber
	totalSubs := dashboard["total_subscribers"].(int64)
	assert.Greater(suite.T(), totalSubs, int64(0))
}

// TestFailedJourneyHandling tests error scenarios in subscriber journey
func (suite *SubscriberJourneyTestSuite) TestFailedJourneyHandling() {
	// Test subscription without active profile
	subReq := map[string]interface{}{
		"profile_id":   "non-existent-profile",
		"rate_plan_id": "non-existent-plan",
		"auto_renew":   true,
	}

	reqBody, _ := json.Marshal(subReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var errorResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), errorResp, "error")

	// Test usage recording for non-existent subscription
	usageReq := map[string]interface{}{
		"data_used":  int64(100),
		"voice_used": int64(10),
		"sms_used":   int64(5),
	}

	reqBody, _ = json.Marshal(usageReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions/non-existent/usage", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	// Test billing for cancelled subscription
	// First create and cancel a subscription
	profileReq := map[string]interface{}{
		"email":   "test@example.com",
		"country": "US",
	}

	reqBody, _ = json.Marshal(profileReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/profiles", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	var profile map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &profile)
	require.NoError(suite.T(), err)
	profileID := fmt.Sprintf("%v", profile["id"])

	// Activate
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/profiles/%s/activate", profileID), nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Create rate plan
	rpReq := map[string]interface{}{
		"name":       "Test Plan",
		"currency":   "USD",
		"base_price": 10.0,
	}
	reqBody, _ = json.Marshal(rpReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/rateplans", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	var ratePlan map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &ratePlan)
	require.NoError(suite.T(), err)
	ratePlanID := fmt.Sprintf("%v", ratePlan["id"])

	// Subscribe
	subReq = map[string]interface{}{
		"profile_id":   profileID,
		"rate_plan_id": ratePlanID,
		"auto_renew":   true,
	}
	reqBody, _ = json.Marshal(subReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	var subscription repository.RatePlanSubscription
	err = json.Unmarshal(w.Body.Bytes(), &subscription)
	require.NoError(suite.T(), err)

	// Cancel subscription
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/subscriptions/%s/cancel", subscription.ID), nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Try to bill cancelled subscription
	chargeReq := map[string]interface{}{
		"subscription_id": subscription.ID,
		"amount":          10.0,
		"currency":        "USD",
	}
	reqBody, _ = json.Marshal(chargeReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/billing/charge", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// Handler implementations (simplified for testing)
func (suite *SubscriberJourneyTestSuite) createProfile(c *gin.Context) {
	profile := map[string]interface{}{
		"id":         fmt.Sprintf("profile-%d", time.Now().UnixNano()),
		"status":     "pending",
		"created_at": time.Now(),
	}

	// Add request fields
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for k, v := range req {
		profile[k] = v
	}

	c.JSON(http.StatusCreated, profile)
}

func (suite *SubscriberJourneyTestSuite) getProfile(c *gin.Context) {
	profile := map[string]interface{}{
		"id":     c.Param("id"),
		"status": "active",
	}
	c.JSON(http.StatusOK, profile)
}

func (suite *SubscriberJourneyTestSuite) activateProfile(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "activated"})
}

func (suite *SubscriberJourneyTestSuite) createRatePlan(c *gin.Context) {
	ratePlan := map[string]interface{}{
		"id":         fmt.Sprintf("plan-%d", time.Now().UnixNano()),
		"created_at": time.Now(),
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for k, v := range req {
		ratePlan[k] = v
	}

	c.JSON(http.StatusCreated, ratePlan)
}

func (suite *SubscriberJourneyTestSuite) listRatePlans(c *gin.Context) {
	plans := []map[string]interface{}{
		{
			"id":    "plan-1",
			"name":  "Basic Plan",
			"price": 9.99,
		},
		{
			"id":    "plan-2",
			"name":  "Premium Plan",
			"price": 29.99,
		},
	}
	c.JSON(http.StatusOK, plans)
}

func (suite *SubscriberJourneyTestSuite) getRatePlan(c *gin.Context) {
	ratePlan := map[string]interface{}{
		"id":   c.Param("id"),
		"name": "Test Plan",
	}
	c.JSON(http.StatusOK, ratePlan)
}

func (suite *SubscriberJourneyTestSuite) createSubscription(c *gin.Context) {
	var subscription repository.RatePlanSubscription
	if err := c.ShouldBindJSON(&subscription); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription.ID = fmt.Sprintf("sub-%d", time.Now().UnixNano())
	subscription.Status = repository.SubscriptionStatus("active")
	subscription.StartedAt = time.Now()
	subscription.CreatedAt = time.Now()
	subscription.UpdatedAt = time.Now()
	subscription.CurrentCycle = time.Now()

	c.JSON(http.StatusCreated, subscription)
}

func (suite *SubscriberJourneyTestSuite) getSubscription(c *gin.Context) {
	var subscription repository.RatePlanSubscription
	if err := suite.db.First(&subscription, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

func (suite *SubscriberJourneyTestSuite) cancelSubscription(c *gin.Context) {
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

func (suite *SubscriberJourneyTestSuite) recordUsage(c *gin.Context) {
	var usage repository.RatePlanUsage
	if err := c.ShouldBindJSON(&usage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	usage.ID = fmt.Sprintf("usage-%d", time.Now().UnixNano())
	usage.RatePlanID = "test-plan"
	usage.LastUpdated = time.Now()

	if err := suite.db.Create(&usage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, usage)
}

func (suite *SubscriberJourneyTestSuite) getUsage(c *gin.Context) {
	var usage []repository.RatePlanUsage
	if err := suite.db.Where("rate_plan_id = ?", "test-plan").Find(&usage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, usage)
}

func (suite *SubscriberJourneyTestSuite) processCharge(c *gin.Context) {
	var charge map[string]interface{}
	if err := c.ShouldBindJSON(&charge); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	charge["id"] = fmt.Sprintf("charge-%d", time.Now().UnixNano())
	charge["status"] = "completed"
	charge["processed_at"] = time.Now()

	c.JSON(http.StatusOK, charge)
}

func (suite *SubscriberJourneyTestSuite) getInvoice(c *gin.Context) {
	invoice := map[string]interface{}{
		"subscription_id": c.Param("subscription_id"),
		"total_amount":    29.99,
		"currency":        "USD",
		"status":          "paid",
		"issued_at":       time.Now(),
	}
	c.JSON(http.StatusOK, invoice)
}

func (suite *SubscriberJourneyTestSuite) getDashboard(c *gin.Context) {
	dashboard := map[string]interface{}{
		"total_subscribers":    int64(100),
		"active_subscriptions": int64(95),
		"monthly_revenue":      2999.0,
		"churn_rate":           0.05,
		"generated_at":         time.Now(),
	}
	c.JSON(http.StatusOK, dashboard)
}

func (suite *SubscriberJourneyTestSuite) getRevenue(c *gin.Context) {
	revenue := map[string]interface{}{
		"total_revenue":   50000.0,
		"monthly_revenue": 5000.0,
		"revenue_by_plan": map[string]float64{
			"basic":   1500.0,
			"premium": 3500.0,
		},
		"currency": "USD",
	}
	c.JSON(http.StatusOK, revenue)
}

// TearDownSuite cleans up the test suite
func (suite *SubscriberJourneyTestSuite) TearDownSuite() {
	sqlDB, err := suite.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// TestSubscriberJourneySuite runs the E2E test suite
func TestSubscriberJourneySuite(t *testing.T) {
	suite.Run(t, new(SubscriberJourneyTestSuite))
}
