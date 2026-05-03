package infra

import (
	"context"
	"net"
	"sync"
)

// GeoRouter handles geographic routing for API requests
type GeoRouter struct {
	regions    map[string]*Region
	defaultReg string
	mu         sync.RWMutex
}

// Region represents a geographic region with endpoints
type Region struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Endpoints []string `json:"endpoints"`
	Priority  int      `json:"priority"`
	IsActive  bool     `json:"is_active"`
	Latency   float64  `json:"latency_ms"`
}

// GeoRoutingConfig configures geographic routing
type GeoRoutingConfig struct {
	DefaultRegion string
	Regions       []*Region
}

// NewGeoRouter creates a new geographic router
func NewGeoRouter(config GeoRoutingConfig) *GeoRouter {
	gr := &GeoRouter{
		regions:    make(map[string]*Region),
		defaultReg: config.DefaultRegion,
	}
	for _, r := range config.Regions {
		gr.regions[r.ID] = r
	}
	return gr
}

// GetRegionForIP determines the best region for an IP address
func (gr *GeoRouter) GetRegionForIP(_ context.Context, ipStr string) (*Region, error) {
	gr.mu.RLock()
	defer gr.mu.RUnlock()

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return gr.regions[gr.defaultReg], nil
	}

	// Simplified geo lookup - in production use MaxMind GeoIP
	regionID := gr.lookupRegion(ip)
	if region, ok := gr.regions[regionID]; ok && region.IsActive {
		return region, nil
	}

	return gr.regions[gr.defaultReg], nil
}

func (gr *GeoRouter) lookupRegion(ip net.IP) string {
	// Simplified region detection based on IP ranges
	if ip.To4() != nil {
		first := ip.To4()[0]
		switch {
		case first >= 1 && first <= 126:
			return "us-east"
		case first >= 128 && first <= 191:
			return "eu-west"
		default:
			return "ap-southeast"
		}
	}
	return gr.defaultReg
}

// GetBestEndpoint returns the best endpoint for a region
func (gr *GeoRouter) GetBestEndpoint(_ context.Context, regionID string) (string, error) {
	gr.mu.RLock()
	defer gr.mu.RUnlock()

	region, ok := gr.regions[regionID]
	if !ok || len(region.Endpoints) == 0 {
		region = gr.regions[gr.defaultReg]
	}

	if len(region.Endpoints) > 0 {
		return region.Endpoints[0], nil
	}
	return "", nil
}

// UpdateRegion updates a region configuration
func (gr *GeoRouter) UpdateRegion(region *Region) {
	gr.mu.Lock()
	defer gr.mu.Unlock()
	gr.regions[region.ID] = region
}

// GetRegions returns all configured regions
func (gr *GeoRouter) GetRegions() []*Region {
	gr.mu.RLock()
	defer gr.mu.RUnlock()

	regions := make([]*Region, 0, len(gr.regions))
	for _, r := range gr.regions {
		regions = append(regions, r)
	}
	return regions
}
