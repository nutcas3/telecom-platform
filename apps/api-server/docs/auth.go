package docs

// LoginRequest represents a login request
// @Description LoginRequest represents a login request
type LoginRequest struct {
	// Username of the user
	// @Required: true
	// @Example: admin
	Username string `json:"username" binding:"required"`
	
	// Password of the user
	// @Required: true
	// @MinLength: 6
	// @Example: password123
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
// @Description LoginResponse represents a login response
type LoginResponse struct {
	// JWT access token (15-minute expiry)
	AccessToken string `json:"access_token"`
	
	// JWT refresh token (7-day expiry)
	RefreshToken string `json:"refresh_token"`
	
	// User information
	User User `json:"user"`
}

// RegisterRequest represents a registration request
// @Description RegisterRequest represents a registration request
type RegisterRequest struct {
	// Username of the user (must be unique)
	// @Required: true
	// @Example: testuser
	Username string `json:"username" binding:"required"`
	
	// Email address of the user (must be unique)
	// @Required: true
	// @Format: email
	// @Example: user@example.com
	Email string `json:"email" binding:"required,email"`
	
	// Password for the user account
	// @Required: true
	// @MinLength: 6
	// @Example: password123
	Password string `json:"password" binding:"required,min=6"`
	
	// First name of the user
	// @Required: true
	// @Example: John
	FirstName string `json:"first_name" binding:"required"`
	
	// Last name of the user
	// @Required: true
	// @Example: Doe
	LastName string `json:"last_name" binding:"required"`
	
	// Role of the user (admin, operator, viewer)
	// @Required: true
	// @Enum: admin,operator,viewer
	// @Example: viewer
	Role string `json:"role" binding:"required,oneof=admin operator viewer"`
}

// ChangePasswordRequest represents a password change request
// @Description ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	// Current password of the user
	// @Required: true
	// @Example: oldpassword123
	OldPassword string `json:"old_password" binding:"required"`
	
	// New password for the user account
	// @Required: true
	// @MinLength: 6
	// @Example: newpassword123
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// User represents a user account
// @Description User represents a user account in the system
type User struct {
	// Unique identifier of the user
	// @ReadOnly: true
	ID uint `json:"id"`
	
	// Username of the user
	// @Example: testuser
	Username string `json:"username"`
	
	// Email address of the user
	// @Format: email
	// @Example: user@example.com
	Email string `json:"email"`
	
	// First name of the user
	// @Example: John
	FirstName string `json:"first_name"`
	
	// Last name of the user
	// @Example: Doe
	LastName string `json:"last_name"`
	
	// Role of the user
	// @Enum: admin,operator,viewer
	// @Example: viewer
	Role string `json:"role"`
	
	// Whether the user account is active
	// @ReadOnly: true
	// @Example: true
	IsActive bool `json:"is_active"`
	
	// Last login time of the user
	// @ReadOnly: true
	// @Format: datetime
	LastLogin *string `json:"last_login"`
	
	// Account creation time
	// @ReadOnly: true
	// @Format: datetime
	CreatedAt string `json:"created_at"`
	
	// Account last update time
	// @ReadOnly: true
	// @Format: datetime
	UpdatedAt string `json:"updated_at"`
}

// ErrorResponse represents an error response
// @Description ErrorResponse represents an error response from the API
type ErrorResponse struct {
	// Error message describing what went wrong
	// @Example: Invalid credentials
	Error string `json:"error"`
}

// SuccessResponse represents a success response
// @Description SuccessResponse represents a success response from the API
type SuccessResponse struct {
	// Success message
	// @Example: Operation completed successfully
	Message string `json:"message"`
}

// PaginationResponse represents pagination information
// @Description PaginationResponse represents pagination information for list endpoints
type PaginationResponse struct {
	// Current page number
	// @Example: 1
	Page int `json:"page"`
	
	// Number of items per page
	// @Example: 20
	Limit int `json:"limit"`
	
	// Total number of items
	// @Example: 100
	Total int64 `json:"total"`
	
	// Total number of pages
	// @Example: 5
	TotalPages int64 `json:"total_pages"`
}

// UsersResponse represents a response with user list and pagination
// @Description UsersResponse represents a response with user list and pagination
type UsersResponse struct {
	// List of users
	Users []User `json:"users"`
	
	// Pagination information
	Pagination PaginationResponse `json:"pagination"`
}
