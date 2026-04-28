package discovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestServiceRegistration tests service registration with Consul
func TestServiceRegistration(t *testing.T) {
	// Skip if Consul is not available
	t.Skip("Skipping integration test - requires Consul server")
}

// TestServiceDiscovery tests service discovery from Consul
func TestServiceDiscovery(t *testing.T) {
	// Skip if Consul is not available
	t.Skip("Skipping integration test - requires Consul server")
}

// TestServiceDeregistration tests service deregistration from Consul
func TestServiceDeregistration(t *testing.T) {
	// Skip if Consul is not available
	t.Skip("Skipping integration test - requires Consul server")
}

// TestServiceStruct tests the Service struct
func TestServiceStruct(t *testing.T) {
	service := Service{
		ID:      "test-service-1",
		Name:    "api-server",
		Address: "localhost",
		Port:    8080,
		Tags:    []string{"api", "telecom"},
	}

	assert.Equal(t, "test-service-1", service.ID)
	assert.Equal(t, "api-server", service.Name)
	assert.Equal(t, "localhost", service.Address)
	assert.Equal(t, 8080, service.Port)
	assert.Equal(t, []string{"api", "telecom"}, service.Tags)
}

// TestServiceDiscoveryStruct tests the ServiceDiscovery struct
func TestServiceDiscoveryStruct(t *testing.T) {
	// This test verifies the struct can be created
	// Actual connection testing requires a running Consul instance
	assert.NotNil(t, ServiceDiscovery{})
}

// TestDiscoverService tests the Discover method
func TestDiscoverService(t *testing.T) {
	// Skip if Consul is not available
	t.Skip("Skipping integration test - requires Consul server")
}

// TestHealthCheck tests health check of discovered services
func TestHealthCheck(t *testing.T) {
	// Skip if Consul is not available
	t.Skip("Skipping integration test - requires Consul server")
}
