package pricing

import (
	"math"
)

func (s *PricingOptimizationService) calculatePriceElasticity(data []HistoricalDataPoint) float64 {
	if len(data) < 2 {
		return -1.2
	}

	var sumX, sumY, sumXY, sumX2 float64
	n := float64(len(data))

	for i := 1; i < len(data); i++ {
		priceChange := (data[i].Price - data[i-1].Price) / data[i-1].Price
		demandChange := float64(data[i].Demand-data[i-1].Demand) / float64(data[i-1].Demand)

		if priceChange != 0 {
			logPrice := math.Abs(priceChange)
			logDemand := math.Abs(demandChange)
			sumX += logPrice
			sumY += logDemand
			sumXY += logPrice * logDemand
			sumX2 += logPrice * logPrice
		}
	}

	elasticity := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	if elasticity > 0 {
		elasticity = -elasticity
	}

	if elasticity < -2.0 {
		elasticity = -2.0
	} else if elasticity > -0.3 {
		elasticity = -0.3
	}

	return elasticity
}

func (s *PricingOptimizationService) optimizeForRevenue(ratePlan *RatePlan, data []HistoricalDataPoint) float64 {
	if len(data) < 3 {
		return ratePlan.BasePrice * 1.05
	}

	elasticity := s.calculatePriceElasticity(data)
	optimalPrice := ratePlan.BasePrice * (1.0 - elasticity/(-elasticity+1))

	minPrice := ratePlan.BasePrice * 0.7
	maxPrice := ratePlan.BasePrice * 1.8

	if optimalPrice < minPrice {
		optimalPrice = minPrice
	} else if optimalPrice > maxPrice {
		optimalPrice = maxPrice
	}

	return math.Round(optimalPrice*100) / 100
}

func (s *PricingOptimizationService) optimizeForMarketShare(ratePlan *RatePlan, data []HistoricalDataPoint) float64 {
	if len(data) < 3 {
		return ratePlan.BasePrice * 0.90
	}

	elasticity := s.calculatePriceElasticity(data)
	var priceReduction float64

	if elasticity < -1.0 {
		priceReduction = 0.15
	} else if elasticity < -0.5 {
		priceReduction = 0.10
	} else {
		priceReduction = 0.20
	}

	optimalPrice := ratePlan.BasePrice * (1.0 - priceReduction)
	minPrice := ratePlan.BasePrice * 0.6

	if optimalPrice < minPrice {
		optimalPrice = minPrice
	}

	return math.Round(optimalPrice*100) / 100
}

func (s *PricingOptimizationService) optimizeForProfitMargin(ratePlan *RatePlan, data []HistoricalDataPoint) float64 {
	variableCost := ratePlan.BasePrice * 0.45
	fixedCost := ratePlan.BasePrice * 0.25
	totalCost := variableCost + fixedCost
	targetMargin := 0.40

	optimalPrice := totalCost / (1.0 - targetMargin)

	if len(data) >= 3 {
		elasticity := s.calculatePriceElasticity(data)
		if elasticity < -1.2 {
			targetMargin = 0.25
			optimalPrice = totalCost / (1.0 - targetMargin)
		}
	}

	maxPrice := ratePlan.BasePrice * 2.0
	if optimalPrice > maxPrice {
		optimalPrice = maxPrice
	}

	return math.Round(optimalPrice*100) / 100
}

func (s *PricingOptimizationService) optimizeForCompetitive(ratePlan *RatePlan, data []HistoricalDataPoint) float64 {
	competitorPrices := []float64{9.99, 12.99, 14.99, 16.99}
	medianPrice := competitorPrices[len(competitorPrices)/2]
	return medianPrice * 0.95
}

func (s *PricingOptimizationService) optimizeForChurnReduction(ratePlan *RatePlan, data []HistoricalDataPoint) float64 {
	return ratePlan.BasePrice * 0.9
}

func (s *PricingOptimizationService) predictOutcomes(ratePlan *RatePlan, price float64, data []HistoricalDataPoint) (float64, int64) {
	demand := s.predictDemand(price, data)
	return price * float64(demand), demand
}

func (s *PricingOptimizationService) predictDemand(price float64, data []HistoricalDataPoint) int64 {
	if len(data) < 2 {
		if price < 20 {
			return 5000
		} else if price < 50 {
			return 2000
		} else {
			return 800
		}
	}

	var totalElasticity, totalWeight float64
	for i := 1; i < len(data); i++ {
		priceChange := (data[i].Price - data[i-1].Price) / data[i-1].Price
		if priceChange != 0 {
			demandChange := float64(data[i].Demand-data[i-1].Demand) / float64(data[i-1].Demand)
			elasticity := demandChange / priceChange
			weight := float64(len(data)-i) / float64(len(data))
			totalElasticity += elasticity * weight
			totalWeight += weight
		}
	}

	avgElasticity := totalElasticity / totalWeight
	baseDemand := float64(data[0].Demand)

	// Apply elasticity with non-linear adjustments
	priceChangeNew := (price - data[0].Price) / data[0].Price

	var demandMultiplier float64
	if math.Abs(priceChangeNew) < 0.1 {
		demandMultiplier = 1 + avgElasticity*priceChangeNew
	} else {
		sign := 1.0
		if priceChangeNew < 0 {
			sign = -1.0
		}
		magnitude := math.Abs(priceChangeNew)
		demandMultiplier = 1 + sign*math.Pow(magnitude, 0.8)*avgElasticity
	}

	predictedDemand := baseDemand * demandMultiplier
	maxDemand := baseDemand * 3.0
	minDemand := baseDemand * 0.1

	if predictedDemand > maxDemand {
		predictedDemand = maxDemand
	} else if predictedDemand < minDemand {
		predictedDemand = minDemand
	}

	return int64(math.Max(100, math.Round(predictedDemand)))
}
