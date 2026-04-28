package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/database"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/middleware"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/services"
)

// TestSetup creates a test environment with database and handlers
type TestSetup struct {
	DB            *database.Database
	Router        *gin.Engine
	AuthService   *services.AuthService
	TestUser      *models.User
	AdminUser     *models.User
	OperatorUser  *models.User
	ViewerUser    *models.User
	JWTToken      string
	AdminToken    string
	OperatorToken string
	ViewerToken   string
}

func setupTest(t *testing.T) *TestSetup {
	// Use in-memory SQLite for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Migrate only the tables we need for testing
	err = db.AutoMigrate(
		&models.User{},
		&models.Plugin{},
		&models.Automation{},
		&models.ConfigEntry{},
		&models.AuthSession{},
		&models.Role{},
		&models.Permission{},
	)
	require.NoError(t, err)

	// Create database wrapper
	database := &database.Database{DB: db}

	// Initialize services
	authService := services.NewAuthService(db, "test-secret")

	// Create test users with different roles
	testUsers := []*models.User{
		{Username: "admin", Email: "admin@test.com", Password: "password123", Role: "admin", FirstName: "Admin", LastName: "User"},
		{Username: "operator", Email: "operator@test.com", Password: "password123", Role: "operator", FirstName: "Operator", LastName: "User"},
		{Username: "viewer", Email: "viewer@test.com", Password: "password123", Role: "viewer", FirstName: "Viewer", LastName: "User"},
		{Username: "user", Email: "user@test.com", Password: "password123", Role: "user", FirstName: "Regular", LastName: "User"},
	}

	for _, user := range testUsers {
		// Hash the password properly for testing
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		require.NoError(t, err)
		user.Password = string(hashedPassword)
		err = db.Create(user).Error
		require.NoError(t, err)
	}

	// Generate JWT tokens for each user
	adminToken, _, err := authService.GetJWTService().GenerateToken(testUsers[0])
	require.NoError(t, err)

	operatorToken, _, err := authService.GetJWTService().GenerateToken(testUsers[1])
	require.NoError(t, err)

	viewerToken, _, err := authService.GetJWTService().GenerateToken(testUsers[2])
	require.NoError(t, err)

	userToken, _, err := authService.GetJWTService().GenerateToken(testUsers[3])
	require.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create handlers
	servicesHandler := NewServicesHandler(nil)     // No K8s in tests
	monitoringHandler := NewMonitoringHandler(nil) // No Prometheus in tests
	pluginsHandler := NewPluginsHandler(services.NewPluginService(database))
	automationHandler := NewAutomationHandler(services.NewAutomationService(database))
	configHandler := NewConfigHandler(services.NewConfigStoreService(database))
	chaosHandler := NewChaosHandler(services.NewChaosService(database))
	authHandler := NewAuthHandler(authService)

	// Setup routes with authentication and RBAC
	v1 := router.Group("/v1")
	{
		// Authentication routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			// User management (admin only)
			users := protected.Group("/users")
			users.Use(middleware.RequireRole("admin"))
			{
				users.GET("", authHandler.GetUsers)
			}

			// Services Management
			svcs := protected.Group("/services")
			{
				svcs.GET("", servicesHandler.List)
				svcsWrite := svcs.Group("/")
				svcsWrite.Use(middleware.RequireRole("admin", "operator"))
				{
					svcsWrite.POST("/:id/restart", servicesHandler.Restart)
				}
			}

			// Monitoring
			mon := protected.Group("/monitoring")
			{
				mon.GET("/metrics", monitoringHandler.Metrics)
			}

			// Plugin Management
			plugs := protected.Group("/plugins")
			{
				plugs.GET("", pluginsHandler.List)
				plugsWrite := plugs.Group("/")
				plugsWrite.Use(middleware.RequireRole("admin"))
				{
					plugsWrite.POST("/install", pluginsHandler.Install)
					plugsWrite.POST("/:id/enable", pluginsHandler.Enable)
					plugsWrite.POST("/:id/disable", pluginsHandler.Disable)
				}
				plugsDelete := plugs.Group("/")
				plugsDelete.Use(middleware.RequireRole("admin"))
				plugsDelete.DELETE("/:id", pluginsHandler.Uninstall)
			}

			// Automation Management
			auto := protected.Group("/automation")
			{
				auto.GET("", automationHandler.List)
				autoWrite := auto.Group("/")
				autoWrite.Use(middleware.RequireRole("admin", "operator"))
				{
					autoWrite.POST("", automationHandler.Create)
				}
			}

			// Configuration Management
			cfg := protected.Group("/config")
			{
				cfg.GET("", configHandler.Get)
				cfgWrite := cfg.Group("/")
				cfgWrite.Use(middleware.RequireRole("admin"))
				{
					cfgWrite.POST("", configHandler.Update)
				}
			}

			// Chaos Engineering
			chaosGroup := protected.Group("/chaos")
			{
				chaosGroup.GET("", chaosHandler.List)
				chaosWrite := chaosGroup.Group("/")
				chaosWrite.Use(middleware.RequireRole("admin"))
				{
					chaosWrite.POST("", chaosHandler.Run)
				}
			}
		}
	}

	return &TestSetup{
		DB:            database,
		Router:        router,
		AuthService:   authService,
		TestUser:      testUsers[3],
		AdminUser:     testUsers[0],
		OperatorUser:  testUsers[1],
		ViewerUser:    testUsers[2],
		JWTToken:      userToken,
		AdminToken:    adminToken,
		OperatorToken: operatorToken,
		ViewerToken:   viewerToken,
	}
}

// Helper function to make authenticated requests
func (ts *TestSetup) makeRequest(method, path string, body any, token string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, _ := http.NewRequest(method, path, reqBody)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)
	return w
}

// Test Authentication Endpoints
func TestAuthenticationEndpoints(t *testing.T) {
	ts := setupTest(t)
	defer ts.DB.Close()

	t.Run("Login with valid credentials", func(t *testing.T) {
		loginData := map[string]string{
			"username": "admin",
			"password": "password123",
		}

		w := ts.makeRequest("POST", "/v1/auth/login", loginData, "")
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "access_token")
		assert.Contains(t, response, "refresh_token")
	})

	t.Run("Login with invalid credentials", func(t *testing.T) {
		loginData := map[string]string{
			"username": "admin",
			"password": "wrongpassword",
		}

		w := ts.makeRequest("POST", "/v1/auth/login", loginData, "")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Access protected endpoint without token", func(t *testing.T) {
		w := ts.makeRequest("GET", "/v1/services", nil, "")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Access protected endpoint with valid token", func(t *testing.T) {
		w := ts.makeRequest("GET", "/v1/services", nil, ts.AdminToken)
		// Services endpoint returns 503 since Kubernetes is not configured in tests
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

// Test RBAC Permissions
func TestRBACPermissions(t *testing.T) {
	ts := setupTest(t)
	defer ts.DB.Close()

	t.Run("Admin can access user management", func(t *testing.T) {
		w := ts.makeRequest("GET", "/v1/users", nil, ts.AdminToken)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Operator cannot access user management", func(t *testing.T) {
		w := ts.makeRequest("GET", "/v1/users", nil, ts.OperatorToken)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Viewer can read monitoring", func(t *testing.T) {
		w := ts.makeRequest("GET", "/v1/monitoring/metrics", nil, ts.ViewerToken)
		assert.Equal(t, http.StatusServiceUnavailable, w.Code) // Expected due to no Prometheus
	})

	t.Run("Viewer cannot write to plugins", func(t *testing.T) {
		pluginData := map[string]any{
			"name":    "test-plugin",
			"version": "1.0.0",
		}
		w := ts.makeRequest("POST", "/v1/plugins/install", pluginData, ts.ViewerToken)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Admin can write to plugins", func(t *testing.T) {
		pluginData := map[string]any{
			"name":    "test-plugin",
			"version": "1.0.0",
		}
		w := ts.makeRequest("POST", "/v1/plugins/install", pluginData, ts.AdminToken)
		// Should not be forbidden (may fail for other reasons like validation)
		assert.NotEqual(t, http.StatusForbidden, w.Code)
	})
}

// Test Service Endpoints
func TestServiceEndpoints(t *testing.T) {
	ts := setupTest(t)
	defer ts.DB.Close()

	t.Run("List services", func(t *testing.T) {
		w := ts.makeRequest("GET", "/v1/services", nil, ts.AdminToken)
		assert.Equal(t, http.StatusServiceUnavailable, w.Code) // Expected due to no K8s
	})

	t.Run("Restart service requires write permissions", func(t *testing.T) {
		w := ts.makeRequest("POST", "/v1/services/123/restart", nil, ts.ViewerToken)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Restart service with admin permissions", func(t *testing.T) {
		w := ts.makeRequest("POST", "/v1/services/123/restart", nil, ts.AdminToken)
		assert.Equal(t, http.StatusServiceUnavailable, w.Code) // Expected due to no K8s
	})
}

// Test Plugin Endpoints
func TestPluginEndpoints(t *testing.T) {
	ts := setupTest(t)
	defer ts.DB.Close()

	t.Run("List plugins", func(t *testing.T) {
		w := ts.makeRequest("GET", "/v1/plugins", nil, ts.AdminToken)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "plugins")
	})

	t.Run("Install plugin without permissions", func(t *testing.T) {
		pluginData := map[string]any{
			"name":    "test-plugin",
			"version": "1.0.0",
		}
		w := ts.makeRequest("POST", "/v1/plugins/install", pluginData, ts.ViewerToken)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Uninstall plugin without permissions", func(t *testing.T) {
		w := ts.makeRequest("DELETE", "/v1/plugins/1", nil, ts.ViewerToken)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// Test Automation Endpoints
func TestAutomationEndpoints(t *testing.T) {
	ts := setupTest(t)
	defer ts.DB.Close()

	t.Run("List automations", func(t *testing.T) {
		w := ts.makeRequest("GET", "/v1/automation", nil, ts.AdminToken)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "automations")
	})

	t.Run("Create automation without permissions", func(t *testing.T) {
		automationData := map[string]any{
			"name":        "test-automation",
			"description": "Test automation",
			"type":        "scheduled",
		}
		w := ts.makeRequest("POST", "/v1/automation/", automationData, ts.ViewerToken)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Create automation with operator permissions", func(t *testing.T) {
		automationData := map[string]any{
			"name":        "test-automation",
			"description": "Test automation",
			"type":        "scheduled",
		}
		w := ts.makeRequest("POST", "/v1/automation/", automationData, ts.OperatorToken)
		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

// Test Configuration Endpoints
func TestConfigurationEndpoints(t *testing.T) {
	ts := setupTest(t)
	defer ts.DB.Close()

	t.Run("List configuration", func(t *testing.T) {
		w := ts.makeRequest("GET", "/v1/config", nil, ts.AdminToken)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Update configuration without permissions", func(t *testing.T) {
		configData := map[string]any{
			"section": "test",
			"key":     "test.config",
			"value":   "test-value",
			"type":    "string",
		}
		w := ts.makeRequest("POST", "/v1/config/", configData, ts.ViewerToken)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Update configuration with admin permissions", func(t *testing.T) {
		configData := map[string]any{
			"section": "test",
			"key":     "test.config",
			"value":   "test-value",
			"type":    "string",
		}
		w := ts.makeRequest("POST", "/v1/config/", configData, ts.AdminToken)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// Test Chaos Engineering Endpoints
func TestChaosEndpoints(t *testing.T) {
	ts := setupTest(t)
	defer ts.DB.Close()

	t.Run("List chaos experiments", func(t *testing.T) {
		w := ts.makeRequest("GET", "/v1/chaos", nil, ts.AdminToken)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "active")
		assert.Contains(t, response, "history")
	})

	t.Run("Run chaos experiment without permissions", func(t *testing.T) {
		chaosData := map[string]any{
			"name":     "test-chaos",
			"type":     "pod-delete",
			"target":   "test-pod",
			"duration": "30s",
		}
		w := ts.makeRequest("POST", "/v1/chaos/", chaosData, ts.ViewerToken)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Run chaos experiment with admin permissions", func(t *testing.T) {
		chaosData := map[string]any{
			"name":     "test-chaos",
			"type":     "pod-delete",
			"target":   "test-pod",
			"duration": "30s",
		}
		w := ts.makeRequest("POST", "/v1/chaos/", chaosData, ts.AdminToken)
		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

// Test Error Handling
func TestErrorHandling(t *testing.T) {
	ts := setupTest(t)
	defer ts.DB.Close()

	t.Run("Invalid JSON payload", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/plugins/install", bytes.NewBufferString("invalid json"))
		req.Header.Set("Authorization", "Bearer "+ts.AdminToken)
		req.Header.Set("Content-Type", "application/json")
		ts.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Missing required fields", func(t *testing.T) {
		pluginData := map[string]any{
			"version": "1.0.0",
			// Missing name field
		}
		w := ts.makeRequest("POST", "/v1/plugins/install", pluginData, ts.AdminToken)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid token format", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/services", nil)
		req.Header.Set("Authorization", "InvalidFormat token")
		ts.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// Test JWT Token Validation
func TestJWTTokenValidation(t *testing.T) {
	ts := setupTest(t)
	defer ts.DB.Close()

	t.Run("Valid token should be accepted", func(t *testing.T) {
		w := ts.makeRequest("GET", "/v1/services", nil, ts.AdminToken)
		assert.NotEqual(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid token should be rejected", func(t *testing.T) {
		w := ts.makeRequest("GET", "/v1/services", nil, "invalid.token.here")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// Benchmark Tests
func BenchmarkAPIRequests(b *testing.B) {
	ts := setupTest(&testing.T{})
	defer ts.DB.Close()

	b.Run("ListServices", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w := ts.makeRequest("GET", "/v1/services", nil, ts.AdminToken)
			_ = w.Code
		}
	})

	b.Run("ListPlugins", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w := ts.makeRequest("GET", "/v1/plugins", nil, ts.AdminToken)
			_ = w.Code
		}
	})
}
