package smdp

import (
	"fmt"
	"math/rand"
	"time"
)

// LoadBalancer implements different load balancing strategies for carrier selection
type LoadBalancer struct {
	strategy LoadBalancingStrategy
	rand     *rand.Rand
}

// LoadBalancingStrategy defines the load balancing algorithm
type LoadBalancingStrategy int

const (
	StrategyRoundRobin LoadBalancingStrategy = iota
	StrategyWeightedRoundRobin
	StrategyLeastConnections
	StrategyRandom
	StrategyPriority
)

// NewLoadBalancer creates a new load balancer with default strategy
func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		strategy: StrategyWeightedRoundRobin,
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// SelectCarrier selects the best carrier based on the configured strategy
func (lb *LoadBalancer) SelectCarrier(carriers []*Carrier, req *ProfileRequest) (*Carrier, error) {
	if len(carriers) == 0 {
		return nil, fmt.Errorf("no carriers available")
	}

	switch lb.strategy {
	case StrategyRoundRobin:
		return lb.roundRobinSelect(carriers), nil
	case StrategyWeightedRoundRobin:
		return lb.weightedRoundRobinSelect(carriers), nil
	case StrategyLeastConnections:
		return lb.leastConnectionsSelect(carriers), nil
	case StrategyRandom:
		return lb.randomSelect(carriers), nil
	case StrategyPriority:
		return lb.prioritySelect(carriers), nil
	default:
		return lb.weightedRoundRobinSelect(carriers), nil
	}
}

// roundRobinSelect implements round-robin carrier selection
func (lb *LoadBalancer) roundRobinSelect(carriers []*Carrier) *Carrier {
	// Simple round-robin based on request count
	// In a real implementation, you'd maintain state across requests
	return carriers[0] // Simplified for demo
}

// weightedRoundRobinSelect selects carrier based on priority and health
func (lb *LoadBalancer) weightedRoundRobinSelect(carriers []*Carrier) *Carrier {
	// Calculate weights based on priority and success rate
	var totalWeight int
	weights := make([]int, len(carriers))

	for i, carrier := range carriers {
		weight := carrier.Priority

		// Adjust weight based on success rate
		if carrier.Metrics.TotalRequests > 0 {
			successRate := float64(carrier.Metrics.SuccessfulRequests) / float64(carrier.Metrics.TotalRequests)
			weight = int(float64(weight) * successRate)
		}

		weights[i] = weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return carriers[0]
	}

	// Select carrier based on weighted random
	random := lb.rand.Intn(totalWeight)
	currentWeight := 0

	for i, weight := range weights {
		currentWeight += weight
		if random < currentWeight {
			return carriers[i]
		}
	}

	return carriers[0]
}

// leastConnectionsSelect selects carrier with least current load
func (lb *LoadBalancer) leastConnectionsSelect(carriers []*Carrier) *Carrier {
	bestCarrier := carriers[0]
	minLoad := bestCarrier.Metrics.RequestRate

	for _, carrier := range carriers[1:] {
		if carrier.Metrics.RequestRate < minLoad {
			bestCarrier = carrier
			minLoad = carrier.Metrics.RequestRate
		}
	}

	return bestCarrier
}

// randomSelect selects a random carrier
func (lb *LoadBalancer) randomSelect(carriers []*Carrier) *Carrier {
	return carriers[lb.rand.Intn(len(carriers))]
}

// prioritySelect selects carrier with highest priority
func (lb *LoadBalancer) prioritySelect(carriers []*Carrier) *Carrier {
	bestCarrier := carriers[0]
	highestPriority := bestCarrier.Priority

	for _, carrier := range carriers[1:] {
		if carrier.Priority > highestPriority {
			bestCarrier = carrier
			highestPriority = carrier.Priority
		}
	}

	return bestCarrier
}
