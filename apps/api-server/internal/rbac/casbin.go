package rbac

import (
	"context"
	"fmt"
	"log"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// CasbinService provides sophisticated RBAC using Casbin
type CasbinService struct {
	enforcer *casbin.Enforcer
	db       *gorm.DB
}

// NewCasbinService creates a new Casbin service
func NewCasbinService(db *gorm.DB) (*CasbinService, error) {
	// Initialize GORM adapter for Casbin
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin adapter: %w", err)
	}

	// Casbin model configuration for RBAC
	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
	`

	// Create model from text
	m, err := model.NewModelFromString(modelText)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin model: %w", err)
	}

	// Create enforcer
	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}

	service := &CasbinService{
		enforcer: enforcer,
		db:       db,
	}

	// Initialize default policies
	if err := service.initializeDefaultPolicies(); err != nil {
		log.Printf("Warning: failed to initialize default policies: %v", err)
	}

	return service, nil
}

// initializeDefaultPolicies sets up default RBAC policies
func (s *CasbinService) initializeDefaultPolicies() error {
	// Default role permissions
	defaultPolicies := [][]string{
		// Admin permissions (full access)
		{"admin", "/v1/services", "GET"},
		{"admin", "/v1/services", "POST"},
		{"admin", "/v1/monitoring", "GET"},
		{"admin", "/v1/deploy", "GET"},
		{"admin", "/v1/deploy", "POST"},
		{"admin", "/v1/plugins", "GET"},
		{"admin", "/v1/plugins", "POST"},
		{"admin", "/v1/plugins", "DELETE"},
		{"admin", "/v1/automation", "GET"},
		{"admin", "/v1/automation", "POST"},
		{"admin", "/v1/billing", "GET"},
		{"admin", "/v1/billing", "POST"},
		{"admin", "/v1/config", "GET"},
		{"admin", "/v1/config", "POST"},
		{"admin", "/v1/chaos", "GET"},
		{"admin", "/v1/chaos", "POST"},

		// Operator permissions (can manage services and monitoring)
		{"operator", "/v1/services", "GET"},
		{"operator", "/v1/services", "POST"},
		{"operator", "/v1/monitoring", "GET"},
		{"operator", "/v1/deploy", "GET"},
		{"operator", "/v1/plugins", "GET"},
		{"operator", "/v1/automation", "GET"},
		{"operator", "/v1/billing", "GET"},
		{"operator", "/v1/config", "GET"},
		{"operator", "/v1/chaos", "GET"},

		// Viewer permissions (read-only access)
		{"viewer", "/v1/services", "GET"},
		{"viewer", "/v1/monitoring", "GET"},
		{"viewer", "/v1/deploy", "GET"},
		{"viewer", "/v1/plugins", "GET"},
		{"viewer", "/v1/automation", "GET"},
		{"viewer", "/v1/billing", "GET"},
		{"viewer", "/v1/config", "GET"},
		{"viewer", "/v1/chaos", "GET"},
	}

	// Add policies if they don't exist
	for _, policy := range defaultPolicies {
		ok, err := s.enforcer.AddPolicy(policy)
		if err != nil {
			return fmt.Errorf("failed to add policy %v: %w", policy, err)
		}
		if ok {
			log.Printf("Added default policy: %v", policy)
		}
	}

	return nil
}

// CheckPermission checks if a user has permission for a specific action on a resource
func (s *CasbinService) CheckPermission(ctx context.Context, userID uint, resource, action string) (bool, error) {
	// Get user with role
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	// Check user's role for permission
	allowed, err := s.enforcer.Enforce(user.Role, resource, action)
	if err != nil {
		return false, fmt.Errorf("failed to enforce policy: %w", err)
	}

	return allowed, nil
}

// AddPolicy adds a new policy to Casbin
func (s *CasbinService) AddPolicy(role, resource, action string) error {
	ok, err := s.enforcer.AddPolicy([]string{role, resource, action})
	if err != nil {
		return fmt.Errorf("failed to add policy: %w", err)
	}
	if !ok {
		return fmt.Errorf("policy already exists")
	}
	return nil
}

// RemovePolicy removes a policy from Casbin
func (s *CasbinService) RemovePolicy(role, resource, action string) error {
	ok, err := s.enforcer.RemovePolicy([]string{role, resource, action})
	if err != nil {
		return fmt.Errorf("failed to remove policy: %w", err)
	}
	if !ok {
		return fmt.Errorf("policy does not exist")
	}
	return nil
}

// GetPolicies returns all policies
func (s *CasbinService) GetPolicies() ([][]string, error) {
	policies, err := s.enforcer.GetPolicy()
	return policies, err
}

// GetRolesForUser returns all roles assigned to a user
func (s *CasbinService) GetRolesForUser(userID uint) ([]string, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return []string{user.Role}, nil
}

// AssignRole assigns a role to a user
func (s *CasbinService) AssignRole(userID uint, roleName string) error {
	// This would typically involve updating the user_roles junction table
	// For now, we'll assume this is handled by the user service
	return nil
}

// RemoveRole removes a role from a user
func (s *CasbinService) RemoveRole(userID uint, roleName string) error {
	// This would typically involve updating the user_roles junction table
	// For now, we'll assume this is handled by the user service
	return nil
}

// GetPermissionsForRole returns all permissions for a role
func (s *CasbinService) GetPermissionsForRole(role string) ([][]string, error) {
	permissions, err := s.enforcer.GetPermissionsForUser(role)
	return permissions, err
}

// HasPermissionForRole checks if a role has a specific permission
func (s *CasbinService) HasPermissionForRole(role, resource, action string) (bool, error) {
	return s.enforcer.Enforce(role, resource, action)
}
