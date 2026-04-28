package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a system user for authentication and authorization
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"uniqueIndex;not null"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"` // Hashed password
	FirstName string         `json:"first_name" gorm:"not null"`
	LastName  string         `json:"last_name" gorm:"not null"`
	Role      string         `json:"role" gorm:"not null;default:'viewer'"` // admin, operator, viewer
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	LastLogin *time.Time     `json:"last_login"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Role represents a system role with permissions
type Role struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Name        string       `json:"name" gorm:"uniqueIndex;not null"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// Permission represents a specific permission
type Permission struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	Resource    string    `json:"resource" gorm:"not null"` // e.g., subscribers, services, deployments
	Action      string    `json:"action" gorm:"not null"`   // e.g., read, write, delete, admin
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AuthSession represents an active user authentication session
type AuthSession struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id" gorm:"not null"`
	User         User      `json:"user" gorm:"foreignKey:UserID"`
	Token        string    `json:"token" gorm:"not null"`
	RefreshToken string    `json:"refresh_token" gorm:"not null"`
	ExpiresAt    time.Time `json:"expires_at"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// APIKey represents an API key for service-to-service authentication
type APIKey struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Name        string     `json:"name" gorm:"not null"`
	Key         string     `json:"key" gorm:"uniqueIndex;not null"`
	Secret      string     `json:"secret" gorm:"not null"`
	Permissions []string   `json:"permissions" gorm:"serializer:json"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	ExpiresAt   *time.Time `json:"expires_at"`
	LastUsed    *time.Time `json:"last_used"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TableName returns the table name for the User model
func (User) TableName() string {
	return "users"
}

// TableName returns the table name for the Role model
func (Role) TableName() string {
	return "roles"
}

// TableName returns the table name for the Permission model
func (Permission) TableName() string {
	return "permissions"
}

// TableName returns the table name for the AuthSession model
func (AuthSession) TableName() string {
	return "auth_sessions"
}

// TableName returns the table name for the APIKey model
func (APIKey) TableName() string {
	return "api_keys"
}

// HasPermission checks if a user has a specific permission
func (u *User) HasPermission(db *gorm.DB, resource, action string) bool {
	var permission Permission
	return db.Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN roles ON roles.id = role_permissions.role_id").
		Where("roles.name = ? AND permissions.resource = ? AND permissions.action = ?", u.Role, resource, action).
		First(&permission).Error == nil
}

// GetPermissions returns all permissions for a user's role
func (u *User) GetPermissions(db *gorm.DB) ([]Permission, error) {
	var permissions []Permission
	err := db.Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN roles ON roles.id = role_permissions.role_id").
		Where("roles.name = ?", u.Role).
		Find(&permissions).Error
	return permissions, err
}
