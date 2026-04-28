package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// JWTService handles JWT token generation and validation
type JWTService struct {
	db        *gorm.DB
	secretKey string
}

// NewJWTService creates a new JWT service
func NewJWTService(db *gorm.DB, secretKey string) *JWTService {
	return &JWTService{
		db:        db,
		secretKey: secretKey,
	}
}

// Claims represents JWT claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken generates a new JWT token for a user
func (s *JWTService) GenerateToken(user *models.User) (string, string, error) {
	// Generate access token (short-lived)
	accessTokenExpiry := time.Now().Add(15 * time.Minute)
	accessClaims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessTokenExpiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "telecom-platform",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token (long-lived)
	refreshTokenExpiry := time.Now().Add(7 * 24 * time.Hour) // 7 days
	refreshClaims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshTokenExpiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "telecom-platform",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	// Store session in database
	sessionID := generateSessionID()
	session := &models.AuthSession{
		ID:           sessionID,
		UserID:       user.ID,
		Token:        accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessTokenExpiry,
		IsActive:     true,
	}

	// Delete any existing inactive sessions for this user to clean up
	s.db.Where("user_id = ? AND is_active = ?", user.ID, false).Delete(&models.AuthSession{})

	if err := s.db.Create(session).Error; err != nil {
		return "", "", fmt.Errorf("failed to create session: %w", err)
	}

	return accessTokenString, refreshTokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Check if session is still active
		var session models.AuthSession
		if err := s.db.Where("token = ? AND is_active = ?", tokenString, true).First(&session).Error; err != nil {
			return nil, fmt.Errorf("session not found or inactive")
		}

		if time.Now().After(session.ExpiresAt) {
			return nil, fmt.Errorf("session expired")
		}

		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// validateTokenWithoutSessionCheck validates a JWT token without checking the database session
func (s *JWTService) validateTokenWithoutSessionCheck(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshToken generates new tokens using a refresh token
func (s *JWTService) RefreshToken(refreshTokenString string) (string, string, error) {
	// First validate the refresh token by checking it in the database
	var session models.AuthSession
	if err := s.db.Where("refresh_token = ? AND is_active = ?", refreshTokenString, true).First(&session).Error; err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		return "", "", fmt.Errorf("refresh token expired")
	}

	// Parse the refresh token to get claims
	claims, err := s.validateTokenWithoutSessionCheck(refreshTokenString)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Get user from database
	var user models.User
	if err := s.db.First(&user, claims.UserID).Error; err != nil {
		return "", "", fmt.Errorf("user not found: %w", err)
	}

	// Deactivate old session
	s.db.Model(&models.AuthSession{}).Where("refresh_token = ?", refreshTokenString).Update("is_active", false)

	// Generate new tokens
	return s.GenerateToken(&user)
}

// InvalidateSession invalidates a user session
func (s *JWTService) InvalidateSession(tokenString string) error {
	return s.db.Model(&models.AuthSession{}).Where("token = ?", tokenString).Update("is_active", false).Error
}

// InvalidateAllUserSessions invalidates all sessions for a user
func (s *JWTService) InvalidateAllUserSessions(userID uint) error {
	return s.db.Model(&models.AuthSession{}).Where("user_id = ?", userID).Update("is_active", false).Error
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword verifies a password against its hash
func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// generateSessionID generates a random session ID
func generateSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// AuthService handles authentication operations
type AuthService struct {
	db         *gorm.DB
	jwtService *JWTService
}

// NewAuthService creates a new authentication service
func NewAuthService(db *gorm.DB, secretKey string) *AuthService {
	return &AuthService{
		db:         db,
		jwtService: NewJWTService(db, secretKey),
	}
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(username, password string) (string, string, *models.User, error) {
	var user models.User
	if err := s.db.Where("username = ? AND is_active = ?", username, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", nil, fmt.Errorf("invalid credentials")
		}
		return "", "", nil, fmt.Errorf("database error: %w", err)
	}

	if err := CheckPassword(password, user.Password); err != nil {
		return "", "", nil, fmt.Errorf("invalid credentials")
	}

	// Update last login
	now := time.Now()
	s.db.Model(&user).Update("last_login", &now)

	// Generate tokens
	accessToken, refreshToken, err := s.jwtService.GenerateToken(&user)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return accessToken, refreshToken, &user, nil
}

// Register creates a new user
func (s *AuthService) Register(username, email, password, firstName, lastName, role string) (*models.User, error) {
	// Hash password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Username:  username,
		Email:     email,
		Password:  hashedPassword,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		IsActive:  true,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if err := CheckPassword(oldPassword, user.Password); err != nil {
		return fmt.Errorf("invalid old password")
	}

	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := s.db.Model(&user).Update("password", hashedPassword).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate all sessions for this user
	s.jwtService.InvalidateAllUserSessions(userID)

	return nil
}

// GetJWTService returns the JWT service
func (s *AuthService) GetJWTService() *JWTService {
	return s.jwtService
}

// GetDB returns the database instance
func (s *AuthService) GetDB() *gorm.DB {
	return s.db
}
