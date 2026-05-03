package rateplan

func (pe *PricingEngine) calculateDataOverage(plan *RatePlan, dataUsed int64) float64 {
	if plan.DataAllowance == nil || plan.DataAllowance.Unlimited || plan.OverageRates == nil {
		return 0
	}

	allowanceMB := plan.DataAllowance.Amount
	if plan.DataAllowance.Unit == "GB" {
		allowanceMB *= 1024
	}

	if dataUsed <= allowanceMB {
		return 0
	}

	overageMB := dataUsed - allowanceMB
	return float64(overageMB) * plan.OverageRates.DataRate
}

func (pe *PricingEngine) calculateVoiceOverage(plan *RatePlan, voiceUsed int64) float64 {
	if plan.VoiceAllowance == nil || plan.VoiceAllowance.Unlimited || plan.OverageRates == nil {
		return 0
	}

	if voiceUsed <= plan.VoiceAllowance.Minutes {
		return 0
	}

	overageMinutes := voiceUsed - plan.VoiceAllowance.Minutes
	return float64(overageMinutes) * plan.OverageRates.VoiceRate
}

func (pe *PricingEngine) calculateSMSOverage(plan *RatePlan, smsUsed int64) float64 {
	if plan.SMSAllowance == nil || plan.SMSAllowance.Unlimited || plan.OverageRates == nil {
		return 0
	}

	if smsUsed <= plan.SMSAllowance.Messages {
		return 0
	}

	overageSMS := smsUsed - plan.SMSAllowance.Messages
	return float64(overageSMS) * plan.OverageRates.SMSRate
}

func (pe *PricingEngine) calculateDiscounts(plan *RatePlan, discountIDs []string, baseCost float64) float64 {
	if plan.Discounts == nil || len(discountIDs) == 0 {
		return 0
	}

	totalDiscount := 0.0
	for _, discountID := range discountIDs {
		for _, discount := range plan.Discounts {
			if discount.ID == discountID && discount.IsActive {
				switch discount.Type {
case DiscountTypePercentage:
					totalDiscount += baseCost * discount.Value / 100
				case DiscountTypeFixed:
					totalDiscount += discount.Value
				}
			}
		}
	}

	return totalDiscount
}

func (pe *PricingEngine) calculateRecommendedPrice(plan *RatePlan, marketAverage float64) float64 {
	// Basic pricing strategy: position slightly below market average for competitive advantage
	if marketAverage > 0 {
		return marketAverage * 0.95 // 5% below market average
	}
	return plan.BasePrice
}

