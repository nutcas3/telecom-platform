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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/middleware"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/services"
)

func setupTestRouter() (*gin.Engine, *services.AuthService) {
	gin.SetMode(gin.TestMode)

	// Create test auth service
	authService := setupAuthServiceForTests()

	// Create auth handler
	authHandler := NewAuthHandler(authService)

	// Use authHandler to avoid unused variable warning
	_ = authHandler

	// Setup router
	router := gin.New()

	// Add auth routes
	v1 := router.Group("/v1")
	{
		// Authentication routes (no auth required)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Protected routes (authentication required)
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			// User management
			authProtected := protected.Group("/auth")
			{
				authProtected.POST("/logout", authHandler.Logout)
				authProtected.GET("/profile", authHandler.GetProfile)
				authProtected.POST("/change-password", authHandler.ChangePassword)
			}

			// Admin-only user management
			users := protected.Group("/users")
			users.Use(middleware.RequireRole("admin"))
			{
				users.GET("", authHandler.GetUsers)
				users.POST("", authHandler.CreateUser)
				users.PUT("/:id", authHandler.UpdateUser)
				users.DELETE("/:id", authHandler.DeleteUser)
			}
		}
	}

	return router, authService
}

func setupAuthServiceForTests() *services.AuthService {
	// Use in-memory SQLite for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("Failed to create test database")
	}

	// Migrate required tables
	err = db.AutoMigrate(
		&models.User{},
		&models.AuthSession{},
		&models.Role{},
		&models.Permission{},
	)
	if err != nil {
		panic("Failed to migrate test database")
	}

	// Create auth service with proper database
	return services.NewAuthService(db, "test-secret")
}

func TestAuthHandler_Register(t *testing.T) {
	router, _ := setupTestRouter()

	// Test successful registration
	reqBody := map[string]any{
		"username":   "testuser",
		"email":      "test@example.com",
		"password":   "password123",
		"first_name": "Test",
		"last_name":  "User",
		"role":       "viewer",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "User registered successfully", response["message"])
	assert.NotNil(t, response["user"])
}

func TestAuthHandler_Login(t *testing.T) {
	router, authService := setupTestRouter()

	// Use authService to avoid unused variable warning
	_ = authService

	// First register a user
	reqBody := map[string]any{
		"username":   "testuser",
		"email":      "test@example.com",
		"password":   "password123",
		"first_name": "Test",
		"last_name":  "User",
		"role":       "admin",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Test successful login
	loginReq := map[string]any{
		"username": "testuser",
		"password": "password123",
	}

	loginBody, _ := json.Marshal(loginReq)
	loginReqHTTP := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(loginBody))
	loginReqHTTP.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, loginReqHTTP)

	assert.Equal(t, http.StatusOK, w.Code)

	var loginResponse map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
	require.NoError(t, err)
	assert.NotEmpty(t, loginResponse["access_token"])
	assert.NotEmpty(t, loginResponse["refresh_token"])
	assert.NotNil(t, loginResponse["user"])
}

func TestAuthHandler_LoginInvalidCredentials(t *testing.T) {
	router, _ := setupTestRouter()

	// Test invalid username
	loginReq := map[string]any{
		"username": "nonexistent",
		"password": "password123",
	}

	loginBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "invalid credentials", response["error"])
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	router, authService := setupTestRouter()

	// Use authService to avoid unused variable warning
	_ = authService

	// Register and login to get tokens
	// First register a user
	reqBody := map[string]any{
		"username":   "testuser",
		"email":      "test@example.com",
		"password":   "password123",
		"first_name": "Test",
		"last_name":  "User",
		"role":       "admin",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Login to get tokens
	loginReq := map[string]any{
		"username": "testuser",
		"password": "password123",
	}

	loginBody, _ := json.Marshal(loginReq)
	loginReqHTTP := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(loginBody))
	loginReqHTTP.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, loginReqHTTP)

	assert.Equal(t, http.StatusOK, w.Code)

	var loginResponse map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
	require.NoError(t, err)

	refreshToken := loginResponse["refresh_token"].(string)
	assert.NotEmpty(t, refreshToken)

	// Test token refresh
	refreshReq := map[string]any{
		"refresh_token": refreshToken,
	}

	refreshBody, _ := json.Marshal(refreshReq)
	refreshReqHTTP := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewBuffer(refreshBody))
	refreshReqHTTP.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, refreshReqHTTP)

	assert.Equal(t, http.StatusOK, w.Code)

	var refreshResponse map[string]any
	err = json.Unmarshal(w.Body.Bytes(), &refreshResponse)
	require.NoError(t, err)
	assert.NotEmpty(t, refreshResponse["access_token"])
	assert.NotEmpty(t, refreshResponse["refresh_token"])
}

func TestAuthHandler_GetProfile(t *testing.T) {
	router, authService := setupTestRouter()

	// Use authService to avoid unused variable warning
	_ = authService

	// Register and login to get token
	reqBody := map[string]any{
		"username":   "testuser",
		"email":      "test@example.com",
		"password":   "password123",
		"first_name": "Test",
		"last_name":  "User",
		"role":       "admin",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Login
	loginReq := map[string]any{
		"username": "testuser",
		"password": "password123",
	}

	loginBody, _ := json.Marshal(loginReq)
	loginReqHTTP := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(loginBody))
	loginReqHTTP.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, loginReqHTTP)

	assert.Equal(t, http.StatusOK, w.Code)

	var loginResponse map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
	require.NoError(t, err)

	accessToken := loginResponse["access_token"].(string)

	// Test getting profile
	profileReq := httptest.NewRequest(http.MethodGet, "/v1/auth/profile", nil)
	profileReq.Header.Set("Authorization", "Bearer "+accessToken)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, profileReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var profileResponse map[string]any
	err = json.Unmarshal(w.Body.Bytes(), &profileResponse)
	require.NoError(t, err)
	assert.Equal(t, "testuser", profileResponse["username"])
	assert.Equal(t, "test@example.com", profileResponse["email"])
}

func TestAuthHandler_GetProfileUnauthorized(t *testing.T) {
	router, _ := setupTestRouter()

	// Test getting profile without token
	profileReq := httptest.NewRequest(http.MethodGet, "/v1/auth/profile", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, profileReq)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Authorization header required", response["error"])
}

func TestAuthHandler_ChangePassword(t *testing.T) {
	router, authService := setupTestRouter()

	// Use authService to avoid unused variable warning
	_ = authService

	// Register and login to get token
	reqBody := map[string]any{
		"username":   "testuser",
		"email":      "test@example.com",
		"password":   "password123",
		"first_name": "Test",
		"last_name":  "User",
		"role":       "admin",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Login
	loginReq := map[string]any{
		"username": "testuser",
		"password": "password123",
	}

	loginBody, _ := json.Marshal(loginReq)
	loginReqHTTP := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(loginBody))
	loginReqHTTP.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, loginReqHTTP)

	assert.Equal(t, http.StatusOK, w.Code)

	var loginResponse map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
	require.NoError(t, err)

	accessToken := loginResponse["access_token"].(string)

	// Test changing password
	changePasswordReq := map[string]any{
		"old_password": "password123",
		"new_password": "newpassword123",
	}

	changePasswordBody, _ := json.Marshal(changePasswordReq)
	changePasswordReqHTTP := httptest.NewRequest(http.MethodPost, "/v1/auth/change-password", bytes.NewBuffer(changePasswordBody))
	changePasswordReqHTTP.Header.Set("Authorization", "Bearer "+accessToken)
	changePasswordReqHTTP.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, changePasswordReqHTTP)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Password changed successfully", response["message"])
}

func TestAuthHandler_Logout(t *testing.T) {
	router, authService := setupTestRouter()

	// Use authService to avoid unused variable warning
	_ = authService

	// Register and login to get token
	reqBody := map[string]any{
		"username":   "testuser",
		"email":      "test@example.com",
		"password":   "password123",
		"first_name": "Test",
		"last_name":  "User",
		"role":       "admin",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Login
	loginReq := map[string]any{
		"username": "testuser",
		"password": "password123",
	}

	loginBody, _ := json.Marshal(loginReq)
	loginReqHTTP := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(loginBody))
	loginReqHTTP.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, loginReqHTTP)

	assert.Equal(t, http.StatusOK, w.Code)

	var loginResponse map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
	require.NoError(t, err)

	accessToken := loginResponse["access_token"].(string)

	// Test logout
	logoutReq := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", nil)
	logoutReq.Header.Set("Authorization", "Bearer "+accessToken)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, logoutReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Logged out successfully", response["message"])
}

func TestAuthHandler_RegisterValidation(t *testing.T) {
	router, _ := setupTestRouter()

	testCases := []struct {
		name     string
		reqBody  map[string]any
		expected int
	}{
		{
			name: "Missing username",
			reqBody: map[string]any{
				"email":      "test@example.com",
				"password":   "password123",
				"first_name": "Test",
				"last_name":  "User",
				"role":       "viewer",
			},
			expected: http.StatusBadRequest,
		},
		{
			name: "Invalid email",
			reqBody: map[string]any{
				"username":   "testuser",
				"email":      "invalid-email",
				"password":   "password123",
				"first_name": "Test",
				"last_name":  "User",
				"role":       "viewer",
			},
			expected: http.StatusBadRequest,
		},
		{
			name: "Short password",
			reqBody: map[string]any{
				"username":   "testuser",
				"email":      "test@example.com",
				"password":   "123",
				"first_name": "Test",
				"last_name":  "User",
				"role":       "viewer",
			},
			expected: http.StatusBadRequest,
		},
		{
			name: "Invalid role",
			reqBody: map[string]any{
				"username":   "testuser",
				"email":      "test@example.com",
				"password":   "password123",
				"first_name": "Test",
				"last_name":  "User",
				"role":       "invalid",
			},
			expected: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expected, w.Code)
		})
	}
}

func TestAuthHandler_LoginValidation(t *testing.T) {
	router, _ := setupTestRouter()

	testCases := []struct {
		name     string
		reqBody  map[string]any
		expected int
	}{
		{
			name: "Missing username",
			reqBody: map[string]any{
				"password": "password123",
			},
			expected: http.StatusBadRequest,
		},
		{
			name: "Missing password",
			reqBody: map[string]any{
				"username": "testuser",
			},
			expected: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expected, w.Code)
		})
	}
}
