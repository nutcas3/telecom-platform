package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/database"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/middleware"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/services"
)

// TestAPIIntegration tests the complete API integration
func TestAPIIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Use in-memory SQLite for testing
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate all required tables
	err = gormDB.AutoMigrate(
		&models.User{},
		&models.Plugin{},
		&models.Automation{},
		&models.ConfigEntry{},
		&models.AuthSession{},
		&models.Role{},
		&models.Permission{},
		&models.Transaction{},
	)
	assert.NoError(t, err)

	// Create database wrapper
	db := &database.Database{DB: gormDB}
	defer db.Close()

	// Initialize services
	authSvc := services.NewAuthService(db.DB, "test-secret-key")

	// Create router
	router := gin.New()

	// Add health endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "api-server",
		})
	})

	// Add Swagger documentation endpoints
	router.GET("/swagger/*any", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "swagger endpoint"})
	})

	// Initialize other services
	chaosSvc := services.NewChaosService(db)
	invoiceSvc := services.NewInvoiceService(db)
	pluginSvc := services.NewPluginService(db)
	automationSvc := services.NewAutomationService(db)
	configStoreSvc := services.NewConfigStoreService(db)
	deploymentSvc := services.NewDeploymentService(db)

	// Build handlers
	authH := NewAuthHandler(authSvc)
	servicesH := NewServicesHandler(nil) // Kubernetes service will be nil for testing
	deploymentsH := NewDeploymentsHandler(deploymentSvc)
	pluginsH := NewPluginsHandler(pluginSvc)
	automationH := NewAutomationHandler(automationSvc)
	configH := NewConfigHandler(configStoreSvc)
	chaosH := NewChaosHandler(chaosSvc)
	billingH := NewBillingHandler(invoiceSvc, db.DB)

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// Authentication routes (no auth required)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authH.Login)
			auth.POST("/register", authH.Register)
			auth.POST("/refresh", authH.RefreshToken)
		}

		// Protected routes (authentication required)
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(authSvc))
		{
			// User management
			authProtected := protected.Group("/auth")
			{
				authProtected.POST("/logout", authH.Logout)
				authProtected.GET("/profile", authH.GetProfile)
				authProtected.POST("/change-password", authH.ChangePassword)
			}

			// Admin-only user management
			users := protected.Group("/users")
			users.Use(middleware.RequireRole("admin"))
			{
				users.GET("", authH.GetUsers)
				users.POST("", authH.CreateUser)
				users.PUT("/:id", authH.UpdateUser)
				users.DELETE("/:id", authH.DeleteUser)
			}
		}

		// Protected API endpoints (authentication required)
		apiProtected := v1.Group("/")
		apiProtected.Use(middleware.AuthMiddleware(authSvc))
		{
			// Services Management
			svcs := apiProtected.Group("/services")
			{
				svcs.GET("", servicesH.List)
				svcs.GET("/:id", servicesH.Get)
				svcs.POST("/:id/restart", servicesH.Restart)
				svcs.GET("/:id/logs", servicesH.Logs)
				svcs.POST("/:id/start", servicesH.Start)
				svcs.POST("/:id/stop", servicesH.Stop)
				svcs.POST("/:id/scale", servicesH.Scale)
				svcs.GET("/:id/health", servicesH.Health)
				svcs.GET("/:id/pods", servicesH.PodStatus)
				svcs.GET("/:id/events", servicesH.Events)
			}

			// Deployment Management
			deploy := apiProtected.Group("/deploy")
			{
				deploy.GET("/status", deploymentsH.Status)
				deploy.POST("/start", deploymentsH.Start)
				deploy.POST("/rollback", deploymentsH.Rollback)
				deploy.GET("/history", deploymentsH.History)
			}

			// Plugin Management
			plugs := apiProtected.Group("/plugins")
			{
				plugs.GET("", pluginsH.List)
				plugs.POST("/install", pluginsH.Install)
				plugs.DELETE("/:id", pluginsH.Uninstall)
				plugs.POST("/:id/enable", pluginsH.Enable)
				plugs.POST("/:id/disable", pluginsH.Disable)
			}

			// Automation Management
			auto := apiProtected.Group("/automation")
			{
				auto.GET("", automationH.List)
				auto.POST("", automationH.Create)
				auto.POST("/:id/run", automationH.Run)
				auto.POST("/:id/schedule", automationH.Schedule)
				auto.GET("/logs", automationH.Logs)
			}

			// Billing & Invoicing
			bill := apiProtected.Group("/billing")
			{
				bill.GET("/invoices", billingH.ListInvoices)
				bill.POST("/invoices", billingH.GenerateInvoice)
				bill.GET("/payments", billingH.ListPayments)
			}

			// Configuration Management
			cfg := apiProtected.Group("/config")
			{
				cfg.GET("", configH.Get)
				cfg.POST("", configH.Update)
				cfg.GET("/validate", configH.Validate)
			}

			// Chaos Engineering
			chaosGroup := apiProtected.Group("/chaos")
			{
				chaosGroup.GET("/experiments", chaosH.List)
				chaosGroup.POST("/experiments", chaosH.Run)
				chaosGroup.GET("/status", chaosH.Status)
			}
		}
	}

	// Test health endpoint
	t.Run("Health Endpoint", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "healthy", response["status"])
		assert.Equal(t, "api-server", response["service"])
	})

	// Test user registration
	t.Run("User Registration", func(t *testing.T) {
		registerData := map[string]any{
			"username":   "testuser",
			"email":      "test@example.com",
			"password":   "password123",
			"first_name": "Test",
			"last_name":  "User",
			"role":       "viewer",
		}

		requestBody, _ := json.Marshal(registerData)
		req, _ := http.NewRequest("POST", "/v1/auth/register", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Debug: print the actual response
		t.Logf("Registration response: %+v", response)

		assert.Equal(t, "User registered successfully", response["message"])
	})

	// Test user login
	t.Run("User Login", func(t *testing.T) {
		loginData := map[string]any{
			"username": "testuser",
			"password": "password123",
		}

		requestBody, _ := json.Marshal(loginData)
		req, _ := http.NewRequest("POST", "/v1/auth/login", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "access_token")
		assert.Contains(t, response, "refresh_token")

		// Store token for subsequent tests
		token := response["access_token"].(string)
		assert.NotEmpty(t, token)

		// Test protected endpoint with authentication
		t.Run("Protected Endpoint", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/v1/services", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should return 503 since Kubernetes service is not configured
			assert.Equal(t, http.StatusServiceUnavailable, w.Code)

			var response map[string]any
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "Kubernetes not configured", response["error"])
		})

		// Test plugin management endpoint
		t.Run("Plugin Management", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/v1/plugins", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]any
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response, "plugins")
			assert.Contains(t, response, "total")
		})

		// Test automation management endpoint
		t.Run("Automation Management", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/v1/automation", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]any
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response, "automations")
		})

		// Test billing endpoint (payments).
		// NOTE: /v1/billing/invoices preloads the LineItems relation of the Invoice model,
		// which depends on the Subscriber model (embedded PLMN/SNSSAI types can't be migrated
		// into SQLite). We exercise the payments endpoint instead, which uses the clean
		// Transaction model.
		t.Run("Billing Management", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/v1/billing/payments", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]any
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response, "payments")
			assert.Contains(t, response, "total")
			assert.Contains(t, response, "page")
			assert.Contains(t, response, "page_size")
		})

		// Test configuration endpoint
		t.Run("Configuration Management", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/v1/config", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]any
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response, "entries")
		})

		// Test chaos engineering endpoint
		t.Run("Chaos Engineering", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/v1/chaos/experiments", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]any
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response, "active")
			assert.Contains(t, response, "history")
		})
	})
}

// TestAuthenticationFlow tests the complete authentication flow
func TestAuthenticationFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Use in-memory SQLite for testing
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate all required tables
	err = gormDB.AutoMigrate(
		&models.User{},
		&models.Plugin{},
		&models.Automation{},
		&models.ConfigEntry{},
		&models.AuthSession{},
		&models.Role{},
		&models.Permission{},
		&models.Transaction{},
	)
	assert.NoError(t, err)

	// Create database wrapper
	db := &database.Database{DB: gormDB}
	defer db.Close()

	// Initialize services
	authSvc := services.NewAuthService(db.DB, "test-secret-key")
	authH := NewAuthHandler(authSvc)

	// Create router
	router := gin.New()

	// Authentication routes
	auth := router.Group("/v1/auth")
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
		auth.POST("/refresh", authH.RefreshToken)
	}

	// Protected routes
	protected := router.Group("/v1")
	protected.Use(middleware.AuthMiddleware(authSvc))
	{
		protected.GET("/profile", authH.GetProfile)
		protected.POST("/logout", authH.Logout)
	}

	// Test registration
	t.Run("Register User", func(t *testing.T) {
		registerData := map[string]any{
			"username":   "testuser2",
			"email":      "test2@example.com",
			"password":   "password123",
			"first_name": "Test",
			"last_name":  "User",
			"role":       "viewer",
		}

		requestBody, _ := json.Marshal(registerData)
		req, _ := http.NewRequest("POST", "/v1/auth/register", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	// Test login with invalid credentials
	t.Run("Invalid Login", func(t *testing.T) {
		loginData := map[string]any{
			"username": "testuser2",
			"password": "wrongpassword",
		}

		requestBody, _ := json.Marshal(loginData)
		req, _ := http.NewRequest("POST", "/v1/auth/login", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Test successful login
	t.Run("Successful Login", func(t *testing.T) {
		loginData := map[string]any{
			"username": "testuser2",
			"password": "password123",
		}

		requestBody, _ := json.Marshal(loginData)
		req, _ := http.NewRequest("POST", "/v1/auth/login", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "access_token")

		token := response["access_token"].(string)
		assert.NotEmpty(t, token)

		// Test accessing protected endpoint with valid token
		req, _ = http.NewRequest("GET", "/v1/profile", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var profileResponse map[string]any
		err = json.Unmarshal(w.Body.Bytes(), &profileResponse)
		assert.NoError(t, err)
		assert.Equal(t, "testuser2", profileResponse["username"])
		assert.Equal(t, "test2@example.com", profileResponse["email"])
	})

	// Test accessing protected endpoint without token
	t.Run("Unauthorized Access", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/profile", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Test accessing protected endpoint with invalid token
	t.Run("Invalid Token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/profile", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
