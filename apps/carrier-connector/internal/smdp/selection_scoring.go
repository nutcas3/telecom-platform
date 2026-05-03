package smdp

import (
	"fmt"
	"math"
	"slices"
)

// calculatePerformanceScore evaluates carrier performance metrics
func (sa *SelectionAlgorithm) calculatePerformanceScore(carrier *Carrier) float64 {
	metrics := carrier.Metrics
	if metrics.TotalRequests == 0 {
		return 50.0 // Neutral score for new carriers
	}

	// Success rate (40% weight)
	successRate := float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests)
	successScore := successRate * 40.0

	// Response time score (30% weight) - lower is better
	responseScore := 30.0
	if metrics.AverageResponseTime > 0 {
		// Normalize: 1s = 0 points, 100ms = 30 points
		responseMs := metrics.AverageResponseTime.Milliseconds()
		if responseMs <= 100 {
			responseScore = 30.0
		} else if responseMs >= 1000 {
			responseScore = 0.0
		} else {
			responseScore = 30.0 * (1.0 - float64(responseMs-100)/900.0)
		}
	}

	// Request rate score (30% weight) - moderate rate is optimal
	requestScore := 15.0 // Base score
	if metrics.RequestRate > 0 && metrics.RequestRate <= 100 {
		requestScore = 30.0
	} else if metrics.RequestRate > 100 {
		// Penalize very high rates (potential overload)
		requestScore = 30.0 * math.Max(0, 1.0-metrics.RequestRate/1000.0)
	}

	return successScore + responseScore + requestScore
}

// calculateReliabilityScore evaluates carrier reliability
func (sa *SelectionAlgorithm) calculateReliabilityScore(carrier *Carrier) float64 {
	// Health status (40% weight)
	healthScore := 0.0
	switch carrier.HealthStatus {
	case CarrierStatusHealthy:
		healthScore = 40.0
	case CarrierStatusDegraded:
		healthScore = 20.0
	case CarrierStatusUnhealthy:
		healthScore = 0.0
	default:
		healthScore = 10.0
	}

	// Uptime based on recent performance (30% weight)
	uptimeScore := 30.0
	if carrier.Metrics.TotalRequests > 0 {
		successRate := float64(carrier.Metrics.SuccessfulRequests) / float64(carrier.Metrics.TotalRequests)
		uptimeScore = successRate * 30.0
	}

	// Priority score (30% weight) - higher priority = more reliable
	priorityScore := float64(carrier.Priority) / 100.0 * 30.0
	if priorityScore > 30.0 {
		priorityScore = 30.0
	}

	return healthScore + uptimeScore + priorityScore
}

// calculateCostScore evaluates cost effectiveness
func (sa *SelectionAlgorithm) calculateCostScore(carrier *Carrier, criteria *SelectionCriteria) float64 {
	// This would integrate with actual pricing data
	// For now, use priority as a proxy (higher priority = potentially more expensive)
	costScore := 100.0 - float64(carrier.Priority)
	if costScore < 0 {
		costScore = 0
	}

	// Apply cost sensitivity
	if criteria.CostSensitivity > 0.5 {
		costScore *= 1.5 // Boost cost score for cost-sensitive requests
	}

	return costScore
}

// calculateRegionScore evaluates regional compatibility
func (sa *SelectionAlgorithm) calculateRegionScore(carrier *Carrier, region string) float64 {
	if region == "" {
		return 50.0 // Neutral score if no region specified
	}

	// Check if carrier supports the region
	if carrier.CountryCode == region {
		return 100.0
	}

	// Check regional compatibility through MCC-to-region mapping
	mccRegion := mccToRegion(carrier.MCC)
	if mccRegion == region {
		return 90.0 // Strong match via MCC
	}

	// Partial match: same continent/area based on MCC range
	if sameRegionGroup(mccRegion, region) {
		return 70.0
	}

	return 30.0 // Low score for distant regions
}

// mccToRegion maps MCC codes to region identifiers
func mccToRegion(mcc string) string {
	if len(mcc) < 1 {
		return ""
	}
	// MCC ranges: 2xx=Europe, 3xx=North America/Caribbean, 4xx=Asia,
	// 5xx=Oceania/Australia, 6xx=Africa, 7xx=South America
	switch mcc[0] {
	case '2':
		return "EU"
	case '3':
		return "NA"
	case '4':
		return "AS"
	case '5':
		return "OC"
	case '6':
		return "AF"
	case '7':
		return "SA"
	default:
		return ""
	}
}

// sameRegionGroup checks if two regions are in the same broader geographic group
func sameRegionGroup(r1, r2 string) bool {
	groups := map[string]string{
		"EU": "EMEA", "AF": "EMEA",
		"NA": "AMER", "SA": "AMER",
		"AS": "APAC", "OC": "APAC",
	}
	return groups[r1] != "" && groups[r1] == groups[r2]
}

// calculateCapabilityScore evaluates carrier capabilities
func (sa *SelectionAlgorithm) calculateCapabilityScore(carrier *Carrier, profileType string) float64 {
	capabilities := carrier.Capabilities
	score := 50.0 // Base score

	// Check if profile type is supported
	if slices.Contains(capabilities.SupportedProfileTypes, profileType) {
		score += 30.0
	}

	// Check for advanced features
	hasBulkDownload := false
	hasRemoteProvisioning := false
	for _, feature := range capabilities.Features {
		if feature == "bulk_download" {
			hasBulkDownload = true
		}
		if feature == "remote_provisioning" {
			hasRemoteProvisioning = true
		}
	}

	if hasBulkDownload {
		score += 10.0
	}
	if hasRemoteProvisioning {
		score += 10.0
	}

	return math.Min(100.0, score)
}

// generateReason creates a human-readable selection reason
func (sa *SelectionAlgorithm) generateReason(score *CarrierScore, _ *SelectionCriteria) string {
	reasons := []string{}

	if score.PerformanceScore > 80 {
		reasons = append(reasons, "excellent performance")
	}
	if score.ReliabilityScore > 80 {
		reasons = append(reasons, "high reliability")
	}
	if score.CostScore > 80 {
		reasons = append(reasons, "cost-effective")
	}
	if score.RegionScore > 80 {
		reasons = append(reasons, "region-compatible")
	}
	if score.CapabilityScore > 80 {
		reasons = append(reasons, "full capability support")
	}

	if len(reasons) == 0 {
		return "best available option"
	}

	return fmt.Sprintf("selected for %s", reasons[0])
}
