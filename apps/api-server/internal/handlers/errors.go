package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// Error codes for different types of errors
const (
	// Validation errors
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeMissingRequired  = "MISSING_REQUIRED"
	ErrCodeInvalidFormat    = "INVALID_FORMAT"

	// Authentication/Authorization errors
	ErrCodeUnauthorized      = "UNAUTHORIZED"
	ErrCodeForbidden         = "FORBIDDEN"
	ErrCodeInvalidToken      = "INVALID_TOKEN"
	ErrCodeTokenExpired      = "TOKEN_EXPIRED"
	ErrCodeInsufficientPerms = "INSUFFICIENT_PERMISSIONS"

	// Resource errors
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeAlreadyExists  = "ALREADY_EXISTS"
	ErrCodeConflict       = "CONFLICT"
	ErrCodeResourceLocked = "RESOURCE_LOCKED"

	// System errors
	ErrCodeInternalError      = "INTERNAL_ERROR"
	ErrCodeDatabaseError      = "DATABASE_ERROR"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrCodeTimeout            = "TIMEOUT"
	ErrCodeRateLimited        = "RATE_LIMITED"

	// Business logic errors
	ErrCodeInvalidState    = "INVALID_STATE"
	ErrCodeQuotaExceeded   = "QUOTA_EXCEEDED"
	ErrCodeOperationFailed = "OPERATION_FAILED"
)

// EnhancedErrorResponse represents an enhanced error response with additional context
type EnhancedErrorResponse struct {
	Error     string      `json:"error"`
	Code      string      `json:"code"`
	Details   string      `json:"details,omitempty"`
	Context   any `json:"context,omitempty"`
	Timestamp string      `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// ValidationError represents field-specific validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrorResponse represents a response with multiple validation errors
type ValidationErrorResponse struct {
	Error     string            `json:"error"`
	Code      string            `json:"code"`
	Errors    []ValidationError `json:"errors"`
	Timestamp string            `json:"timestamp"`
	RequestID string            `json:"request_id,omitempty"`
}

// handleError is the central error handling function
func handleError(c *gin.Context, err error, defaultMessage string, defaultCode string) {
	// Analyze the error type and provide appropriate response
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Resource not found",
			Code:    ErrCodeNotFound,
			Details: fmt.Sprintf("The requested resource could not be found: %s", err.Error()),
		})

	case errors.Is(err, gorm.ErrDuplicatedKey):
		c.JSON(http.StatusConflict, ErrorResponse{
			Error:   "Resource already exists",
			Code:    ErrCodeAlreadyExists,
			Details: "A resource with this identifier already exists",
		})

	case errors.Is(err, gorm.ErrInvalidTransaction):
		fallthrough
	case errors.Is(err, gorm.ErrInvalidData):
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid data",
			Code:    ErrCodeInvalidInput,
			Details: fmt.Sprintf("The provided data is invalid: %s", err.Error()),
		})

	case isValidationError(err):
		handleValidationError(c, err)

	case isAuthError(err):
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Authentication failed",
			Code:    ErrCodeUnauthorized,
			Details: err.Error(),
		})

	case isPermissionError(err):
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Access denied",
			Code:    ErrCodeForbidden,
			Details: "You don't have permission to perform this action",
		})

	case isTimeoutError(err):
		c.JSON(http.StatusRequestTimeout, ErrorResponse{
			Error:   "Request timeout",
			Code:    ErrCodeTimeout,
			Details: "The request took too long to process",
		})

	case isRateLimitError(err):
		c.JSON(http.StatusTooManyRequests, ErrorResponse{
			Error:   "Rate limit exceeded",
			Code:    ErrCodeRateLimited,
			Details: "Too many requests, please try again later",
		})

	default:
		// Log the error for debugging
		c.Error(err)

		statusCode := http.StatusInternalServerError
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(strings.ToLower(err.Error()), "unauthorized") {
			statusCode = http.StatusUnauthorized
		} else if strings.Contains(strings.ToLower(err.Error()), "forbidden") {
			statusCode = http.StatusForbidden
		} else if strings.Contains(strings.ToLower(err.Error()), "invalid") {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   defaultMessage,
			Code:    defaultCode,
			Details: err.Error(),
		})
	}
}

// handleValidationError handles validation errors with field-specific details
func handleValidationError(c *gin.Context, err error) {
	var validationErrors []ValidationError

	// Handle validator.ValidationErrors
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErrs {
			validationErrors = append(validationErrors, ValidationError{
				Field:   fieldErr.Field(),
				Message: getValidationErrorMessage(fieldErr),
				Value:   fmt.Sprintf("%v", fieldErr.Value()),
			})
		}
	} else {
		// Handle other validation errors
		validationErrors = append(validationErrors, ValidationError{
			Field:   "general",
			Message: err.Error(),
		})
	}

	c.JSON(http.StatusBadRequest, ValidationErrorResponse{
		Error:     "Validation failed",
		Code:      ErrCodeValidationFailed,
		Errors:    validationErrors,
		Timestamp: getCurrentTimestamp(),
	})
}

// Helper functions for error type detection
func isValidationError(err error) bool {
	_, ok := err.(validator.ValidationErrors)
	return ok || strings.Contains(strings.ToLower(err.Error()), "validation")
}

func isAuthError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "authentication") ||
		strings.Contains(errStr, "invalid credentials") ||
		strings.Contains(errStr, "invalid token")
}

func isPermissionError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "forbidden") ||
		strings.Contains(errStr, "permission") ||
		strings.Contains(errStr, "access denied") ||
		strings.Contains(errStr, "insufficient")
}

func isTimeoutError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline exceeded")
}

func isRateLimitError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "too many requests")
}

// getValidationErrorMessage converts validator.FieldError to user-friendly message
func getValidationErrorMessage(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fieldErr.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fieldErr.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", fieldErr.Field(), fieldErr.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", fieldErr.Field(), fieldErr.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", fieldErr.Field(), fieldErr.Param())
	case "numeric":
		return fmt.Sprintf("%s must be a number", fieldErr.Field())
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", fieldErr.Field())
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", fieldErr.Field())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", fieldErr.Field(), fieldErr.Param())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", fieldErr.Field())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", fieldErr.Field())
	case "datetime":
		return fmt.Sprintf("%s must be a valid datetime", fieldErr.Field())
	default:
		return fmt.Sprintf("%s is invalid", fieldErr.Field())
	}
}

// getCurrentTimestamp returns the current timestamp in ISO 8601 format
func getCurrentTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// Convenience functions for common error scenarios
func BadRequest(c *gin.Context, err error, details ...string) {
	message := "Invalid request"
	if len(details) > 0 {
		message = details[0]
	}
	handleError(c, err, message, ErrCodeInvalidInput)
}

func Unauthorized(c *gin.Context, err error, details ...string) {
	message := "Unauthorized"
	if len(details) > 0 {
		message = details[0]
	}
	handleError(c, err, message, ErrCodeUnauthorized)
}

func Forbidden(c *gin.Context, err error, details ...string) {
	message := "Access denied"
	if len(details) > 0 {
		message = details[0]
	}
	handleError(c, err, message, ErrCodeForbidden)
}

func NotFound(c *gin.Context, err error, details ...string) {
	message := "Resource not found"
	if len(details) > 0 {
		message = details[0]
	}
	handleError(c, err, message, ErrCodeNotFound)
}

func Conflict(c *gin.Context, err error, details ...string) {
	message := "Resource conflict"
	if len(details) > 0 {
		message = details[0]
	}
	handleError(c, err, message, ErrCodeConflict)
}

func InternalServerError(c *gin.Context, err error, details ...string) {
	message := "Internal server error"
	if len(details) > 0 {
		message = details[0]
	}
	handleError(c, err, message, ErrCodeInternalError)
}

func ServiceUnavailable(c *gin.Context, err error, details ...string) {
	message := "Service unavailable"
	if len(details) > 0 {
		message = details[0]
	}
	handleError(c, err, message, ErrCodeServiceUnavailable)
}

// Enhanced parsing functions with better error handling
func ParseUintParam(c *gin.Context, name string) (uint, bool) {
	value := c.Param(name)
	if value == "" {
		BadRequest(c, fmt.Errorf("parameter %s is required", name), fmt.Sprintf("Missing required parameter: %s", name))
		return 0, false
	}

	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		BadRequest(c, err, fmt.Sprintf("Invalid %s parameter: must be a positive integer", name))
		return 0, false
	}

	return uint(parsed), true
}

func ParseIntParam(c *gin.Context, name string) (int, bool) {
	value := c.Param(name)
	if value == "" {
		BadRequest(c, fmt.Errorf("parameter %s is required", name), fmt.Sprintf("Missing required parameter: %s", name))
		return 0, false
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		BadRequest(c, err, fmt.Sprintf("Invalid %s parameter: must be an integer", name))
		return 0, false
	}

	return parsed, true
}

func ParseBoolQuery(c *gin.Context, key string) *bool {
	raw := c.Query(key)
	if raw == "" {
		return nil
	}

	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		BadRequest(c, err, fmt.Sprintf("Invalid %s query parameter: must be true or false", key))
		return nil
	}

	return &parsed
}

func ParseIntQuery(c *gin.Context, key string, defaultValue int) (int, bool) {
	raw := c.Query(key)
	if raw == "" {
		return defaultValue, true
	}

	parsed, err := strconv.Atoi(raw)
	if err != nil {
		BadRequest(c, err, fmt.Sprintf("Invalid %s query parameter: must be an integer", key))
		return defaultValue, false
	}

	return parsed, true
}
