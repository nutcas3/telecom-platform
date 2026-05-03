package rateplan

import (
	"fmt"
	"time"
)

func (pe *PricingEngine) validateBasicFields(plan *RatePlan) error {
	if plan.Name == "" {
		return fmt.Errorf("rate plan name is required")
	}
	if plan.CarrierID == "" {
		return fmt.Errorf("carrier ID is required")
	}
	if plan.Region == "" {
		return fmt.Errorf("region is required")
	}
	if plan.BasePrice < 0 {
		return fmt.Errorf("base price cannot be negative")
	}
	if plan.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	if plan.BillingCycle == "" {
		return fmt.Errorf("billing cycle is required")
	}
	return nil
}

func (pe *PricingEngine) validateDates(plan *RatePlan) error {
	if plan.ValidFrom.IsZero() {
		return fmt.Errorf("valid from date is required")
	}

	now := time.Now()
	if plan.ValidFrom.After(now) {
		pe.logger.Warning("Rate plan valid from date is in the future")
	}

	if plan.ValidTo != nil && plan.ValidTo.Before(plan.ValidFrom) {
		return fmt.Errorf("valid to date cannot be before valid from date")
	}

	return nil
}

func (pe *PricingEngine) validateAllowances(plan *RatePlan) error {
	// Validate data allowance
	if plan.DataAllowance != nil {
		if plan.DataAllowance.Amount <= 0 && !plan.DataAllowance.Unlimited {
			return fmt.Errorf("data allowance amount must be positive or unlimited")
		}
		if plan.DataAllowance.Unit == "" {
			return fmt.Errorf("data allowance unit is required")
		}
	}

	// Validate voice allowance
	if plan.VoiceAllowance != nil {
		if plan.VoiceAllowance.Minutes <= 0 && !plan.VoiceAllowance.Unlimited {
			return fmt.Errorf("voice allowance minutes must be positive or unlimited")
		}
	}

	// Validate SMS allowance
	if plan.SMSAllowance != nil {
		if plan.SMSAllowance.Messages <= 0 && !plan.SMSAllowance.Unlimited {
			return fmt.Errorf("SMS allowance messages must be positive or unlimited")
		}
	}

	return nil
}

func (pe *PricingEngine) validateOverageRates(plan *RatePlan) error {
	if plan.OverageRates != nil {
		if plan.OverageRates.DataRate < 0 {
			return fmt.Errorf("data overage rate cannot be negative")
		}
		if plan.OverageRates.VoiceRate < 0 {
			return fmt.Errorf("voice overage rate cannot be negative")
		}
		if plan.OverageRates.SMSRate < 0 {
			return fmt.Errorf("SMS overage rate cannot be negative")
		}
		if plan.OverageRates.Currency == "" {
			return fmt.Errorf("overage rates currency is required")
		}
	}
	return nil
}

func (pe *PricingEngine) validateDiscounts(plan *RatePlan) error {
	if plan.Discounts != nil {
		for _, discount := range plan.Discounts {
			if discount.Name == "" {
				return fmt.Errorf("discount name is required")
			}
			if discount.Value <= 0 {
				return fmt.Errorf("discount value must be positive")
			}
			if discount.ValidFrom.IsZero() {
				return fmt.Errorf("discount valid from date is required")
			}
			if discount.ValidTo != nil && discount.ValidTo.Before(discount.ValidFrom) {
				return fmt.Errorf("discount valid to date cannot be before valid from date")
			}
		}
	}
	return nil
}

func (pe *PricingEngine) validateEarlyTermination(plan *RatePlan) error {
	if plan.EarlyTermination != nil && plan.EarlyTermination.Enabled {
		if plan.EarlyTermination.FeeType == "" {
			return fmt.Errorf("early termination fee type is required")
		}
		if plan.EarlyTermination.FeeType == "fixed" && plan.EarlyTermination.FeeAmount <= 0 {
			return fmt.Errorf("early termination fee amount must be positive for fixed fee type")
		}
		if plan.EarlyTermination.FeeType == "percentage" && (plan.EarlyTermination.FeePercentage <= 0 || plan.EarlyTermination.FeePercentage > 100) {
			return fmt.Errorf("early termination fee percentage must be between 0 and 100")
		}
	}
	return nil
}

func (pe *PricingEngine) validateSubscriptionDiscounts(plan *RatePlan, discountIDs []string) error {
	if plan.Discounts == nil {
		return fmt.Errorf("no discounts available for this rate plan")
	}

	for _, discountID := range discountIDs {
		found := false
		for _, discount := range plan.Discounts {
			if discount.ID == discountID {
				if !discount.IsActive {
					return fmt.Errorf("discount %s is not active", discountID)
				}
				now := time.Now()
				if now.Before(discount.ValidFrom) || (discount.ValidTo != nil && now.After(*discount.ValidTo)) {
					return fmt.Errorf("discount %s is not currently valid", discountID)
				}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("discount %s not found", discountID)
		}
	}

	return nil
}
