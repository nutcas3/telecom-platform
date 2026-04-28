package discovery

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hashicorp/consul/api"
)

// Service represents a service instance
type Service struct {
	ID      string
	Name    string
	Address string
	Port    int
	Tags    []string
}

// ServiceDiscovery provides service registration and discovery
type ServiceDiscovery struct {
	client *api.Client
}

// NewServiceDiscovery creates a new service discovery client
func NewServiceDiscovery() (*ServiceDiscovery, error) {
	consulAddr := os.Getenv("CONSUL_ADDR")
	if consulAddr == "" {
		consulAddr = "127.0.0.1:8500"
	}

	config := api.DefaultConfig()
	config.Address = consulAddr

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Consul client: %w", err)
	}

	return &ServiceDiscovery{
		client: client,
	}, nil
}

// Register registers a service with the discovery system
func (sd *ServiceDiscovery) Register(service Service, ttl time.Duration) error {
	registration := &api.AgentServiceRegistration{
		ID:      service.ID,
		Name:    service.Name,
		Address: service.Address,
		Port:    service.Port,
		Tags:    service.Tags,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", service.Address, service.Port),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: ttl.String(),
		},
	}

	err := sd.client.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("failed to register service %s: %w", service.Name, err)
	}

	log.Printf("Registered service: %s (%s:%d)", service.Name, service.Address, service.Port)
	return nil
}

// Deregister deregisters a service from the discovery system
func (sd *ServiceDiscovery) Deregister(serviceID string) error {
	err := sd.client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		return fmt.Errorf("failed to deregister service %s: %w", serviceID, err)
	}

	log.Printf("Deregistered service: %s", serviceID)
	return nil
}

// Discover discovers a service by name
func (sd *ServiceDiscovery) Discover(serviceName string) ([]Service, error) {
	services, _, err := sd.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to discover service %s: %w", serviceName, err)
	}

	result := make([]Service, 0, len(services))
	for _, entry := range services {
		if entry.Service == nil {
			continue
		}
		result = append(result, Service{
			ID:      entry.Service.ID,
			Name:    entry.Service.Service,
			Address: entry.Service.Address,
			Port:    entry.Service.Port,
			Tags:    entry.Service.Tags,
		})
	}

	return result, nil
}

// DiscoverOne discovers a single service instance (round-robin)
func (sd *ServiceDiscovery) DiscoverOne(serviceName string) (*Service, error) {
	services, err := sd.Discover(serviceName)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("no instances found for service %s", serviceName)
	}

	// Simple round-robin: return first instance
	// In production, use a proper load balancing strategy
	return &services[0], nil
}

// HealthCheck checks if a service is healthy
func (sd *ServiceDiscovery) HealthCheck(serviceName string) (bool, error) {
	services, _, err := sd.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return false, fmt.Errorf("failed to check health of service %s: %w", serviceName, err)
	}

	for _, entry := range services {
		if entry.Checks.AggregatedStatus() == "passing" {
			return true, nil
		}
	}

	return false, nil
}

// Watch watches for changes to a service
func (sd *ServiceDiscovery) Watch(ctx context.Context, serviceName string, callback func([]Service)) error {
	// Start a ticker to periodically check for service changes
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var lastServices []Service

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			services, err := sd.Discover(serviceName)
			if err != nil {
				log.Printf("Failed to discover service %s: %v", serviceName, err)
				continue
			}

			// Check if services have changed
			if !servicesEqual(services, lastServices) {
				callback(services)
				lastServices = services
			}
		}
	}
}

// servicesEqual compares two service lists for equality
func servicesEqual(a, b []Service) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].ID != b[i].ID || a[i].Address != b[i].Address || a[i].Port != b[i].Port {
			return false
		}
	}

	return true
}

// Close closes the service discovery client
func (sd *ServiceDiscovery) Close() error {
	// Consul client doesn't need explicit closing
	return nil
}
