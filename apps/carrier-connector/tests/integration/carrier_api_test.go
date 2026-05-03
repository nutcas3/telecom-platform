package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// CarrierAPITestSuite tests carrier API integration
type CarrierAPITestSuite struct {
	suite.Suite
	db     *gorm.DB
	router *gin.Engine
}

func (suite *CarrierAPITestSuite) SetupSuite() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(suite.T(), err)

	err = db.AutoMigrate(&repository.RatePlan{}, &repository.RatePlanSubscription{}, &repository.RatePlanUsage{})
	require.NoError(suite.T(), err)

	suite.db = db
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	suite.setupRoutes()
}

func (suite *CarrierAPITestSuite) setupRoutes() {
	api := suite.router.Group("/api/v1")
	api.POST("/rateplans", suite.createRatePlan)
	api.GET("/rateplans", suite.listRatePlans)
	api.GET("/rateplans/:id", suite.getRatePlan)
	api.POST("/subscriptions", suite.createSubscription)
	api.GET("/subscriptions/:id", suite.getSubscription)
	api.POST("/usage", suite.recordUsage)
	api.GET("/health", suite.healthCheck)
}

func (suite *CarrierAPITestSuite) TestRatePlanLifecycle() {
	plan := map[string]interface{}{
		"name": "Test Plan", "description": "Test", "carrier_id": "carrier-1",
		"region": "US", "base_price": 10.0, "currency": "USD", "is_active": true,
	}
	body, _ := json.Marshal(plan)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rateplans", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var created repository.RatePlan
	json.Unmarshal(w.Body.Bytes(), &created)
	assert.Equal(suite.T(), "Test Plan", created.Name)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/rateplans/"+created.ID, nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/rateplans", nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *CarrierAPITestSuite) TestSubscriptionLifecycle() {
	plan := &repository.RatePlan{ID: "plan-1", Name: "Plan", CarrierID: "c1", BasePrice: 10, Currency: "USD", IsActive: true}
	suite.db.Create(plan)

	sub := map[string]interface{}{"profile_id": "profile-1", "rate_plan_id": "plan-1", "auto_renew": true}
	body, _ := json.Marshal(sub)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var created repository.RatePlanSubscription
	json.Unmarshal(w.Body.Bytes(), &created)
	assert.Equal(suite.T(), "profile-1", created.ProfileID)
}

func (suite *CarrierAPITestSuite) TestUsageRecording() {
	usage := map[string]interface{}{"profile_id": "p1", "rate_plan_id": "rp1", "data_used": 100, "voice_used": 10, "sms_used": 5}
	body, _ := json.Marshal(usage)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/usage", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
}

func (suite *CarrierAPITestSuite) TestHealthCheck() {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *CarrierAPITestSuite) createRatePlan(c *gin.Context) {
	var plan repository.RatePlan
	if err := c.ShouldBindJSON(&plan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	plan.ID = "rp-" + time.Now().Format("20060102150405")
	plan.CreatedAt = time.Now()
	plan.UpdatedAt = time.Now()
	suite.db.Create(&plan)
	c.JSON(http.StatusCreated, plan)
}

func (suite *CarrierAPITestSuite) listRatePlans(c *gin.Context) {
	var plans []repository.RatePlan
	suite.db.Find(&plans)
	c.JSON(http.StatusOK, plans)
}

func (suite *CarrierAPITestSuite) getRatePlan(c *gin.Context) {
	var plan repository.RatePlan
	if err := suite.db.First(&plan, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, plan)
}

func (suite *CarrierAPITestSuite) createSubscription(c *gin.Context) {
	var sub repository.RatePlanSubscription
	if err := c.ShouldBindJSON(&sub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sub.ID = "sub-" + time.Now().Format("20060102150405")
	sub.Status = repository.SubscriptionStatusActive
	sub.StartedAt = time.Now()
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()
	suite.db.Create(&sub)
	c.JSON(http.StatusCreated, sub)
}

func (suite *CarrierAPITestSuite) getSubscription(c *gin.Context) {
	var sub repository.RatePlanSubscription
	if err := suite.db.First(&sub, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, sub)
}

func (suite *CarrierAPITestSuite) recordUsage(c *gin.Context) {
	var usage repository.RatePlanUsage
	if err := c.ShouldBindJSON(&usage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	usage.ID = "usage-" + time.Now().Format("20060102150405")
	usage.LastUpdated = time.Now()
	suite.db.Create(&usage)
	c.JSON(http.StatusCreated, usage)
}

func (suite *CarrierAPITestSuite) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy", "timestamp": time.Now()})
}

func (suite *CarrierAPITestSuite) TearDownSuite() {
	if db, err := suite.db.DB(); err == nil {
		db.Close()
	}
}

func TestCarrierAPITestSuite(t *testing.T) {
	suite.Run(t, new(CarrierAPITestSuite))
}
