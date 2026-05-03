package mvno

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/sirupsen/logrus"
)

// OnboardingValidator validates MVNO onboarding requests and configurations
type OnboardingValidator struct {
	logger *logrus.Logger
}

// NewOnboardingValidator creates a new validator instance
func NewOnboardingValidator(logger *logrus.Logger) *OnboardingValidator {
	return &OnboardingValidator{logger: logger}
}

// ValidateRequest validates the initial onboarding request
func (v *OnboardingValidator) ValidateRequest(req *OnboardingRequest) error {
	if err := v.validateBusinessInfo(req); err != nil {
		return fmt.Errorf("business validation failed: %w", err)
	}

	if err := v.validatePlan(req); err != nil {
		return fmt.Errorf("plan validation failed: %w", err)
	}

	if err := v.validateTechnicalRequirements(req); err != nil {
		return fmt.Errorf("technical validation failed: %w", err)
	}

	return nil
}

// ValidateMVNO performs comprehensive MVNO validation
func (v *OnboardingValidator) ValidateMVNO(ctx context.Context, mvno *MVNO) error {
	// Validate business registration
	if err := v.validateBusinessRegistration(mvno.BusinessID); err != nil {
		return fmt.Errorf("business registration validation failed: %w", err)
	}

	// Validate compliance requirements
	if err := v.validateCompliance(ctx, mvno); err != nil {
		return fmt.Errorf("compliance validation failed: %w", err)
	}

	// Validate technical feasibility
	if err := v.validateTechnicalFeasibility(mvno); err != nil {
		return fmt.Errorf("technical feasibility validation failed: %w", err)
	}

	return nil
}

// validateBusinessInfo validates business information
func (v *OnboardingValidator) validateBusinessInfo(req *OnboardingRequest) error {
	if len(req.BusinessName) < 3 {
		return fmt.Errorf("business name must be at least 3 characters")
	}

	if !v.isValidBusinessID(req.BusinessID) {
		return fmt.Errorf("invalid business ID format")
	}

	if !v.isValidEmail(req.ContactEmail) {
		return fmt.Errorf("invalid contact email format")
	}

	if len(req.ContactPhone) < 10 {
		return fmt.Errorf("contact phone must be at least 10 digits")
	}

	return nil
}

// validatePlan validates the selected plan
func (v *OnboardingValidator) validatePlan(req *OnboardingRequest) error {
	validPlans := []MVNOPlan{PlanStarter, PlanGrowth, PlanScale, PlanEnterprise}
	isValid := slices.Contains(validPlans, req.Plan)
	if !isValid {
		return fmt.Errorf("invalid plan selected")
	}

	if req.EstimatedSubs < 1 {
		return fmt.Errorf("estimated subscribers must be at least 1")
	}

	// Validate plan-specific requirements
	switch req.Plan {
	case PlanStarter:
		if req.EstimatedSubs > 1000 {
			return fmt.Errorf("starter plan limited to 1000 subscribers")
		}
	case PlanGrowth:
		if req.EstimatedSubs > 10000 {
			return fmt.Errorf("growth plan limited to 10000 subscribers")
		}
	}

	return nil
}

// validateTechnicalRequirements validates technical requirements
func (v *OnboardingValidator) validateTechnicalRequirements(req *OnboardingRequest) error {
	if len(req.TargetCountries) == 0 {
		return fmt.Errorf("at least one target country must be specified")
	}

	if len(req.UseCase) < 10 {
		return fmt.Errorf("use case description too short")
	}

	// Validate country codes
	for _, country := range req.TargetCountries {
		if len(country) != 2 {
			return fmt.Errorf("invalid country code format: %s", country)
		}
	}

	return nil
}

// validateBusinessRegistration validates business registration
func (v *OnboardingValidator) validateBusinessRegistration(businessID string) error {
	// In production, this would validate against business registry
	// For now, basic format validation
	if len(businessID) < 8 {
		return fmt.Errorf("business ID too short")
	}
	return nil
}

// validateCompliance validates regulatory compliance
func (v *OnboardingValidator) validateCompliance(_ context.Context, mvno *MVNO) error {
	// Check regulatory compliance for target countries
	for _, country := range mvno.Config.AllowedCountries {
		if err := v.checkCountryCompliance(country); err != nil {
			return fmt.Errorf("compliance check failed for %s: %w", country, err)
		}
	}
	return nil
}

// validateTechnicalFeasibility validates technical implementation feasibility
func (v *OnboardingValidator) validateTechnicalFeasibility(mvno *MVNO) error {
	// Validate carrier availability for target countries
	for _, country := range mvno.Config.AllowedCountries {
		if !v.hasCarrierCoverage(country) {
			return fmt.Errorf("no carrier coverage available for %s", country)
		}
	}
	return nil
}

// isValidBusinessID validates business ID format
func (v *OnboardingValidator) isValidBusinessID(id string) bool {
	// Basic validation - alphanumeric with possible hyphens
	matched, _ := regexp.MatchString(`^[A-Za-z0-9\-]{8,}$`, id)
	return matched
}

// isValidEmail validates email format
func (v *OnboardingValidator) isValidEmail(email string) bool {
	matched, _ := regexp.MatchString(`^[^\s@]+@[^\s@]+\.[^\s@]+$`, email)
	return matched
}

// checkCountryCompliance checks if country is compliant
func (v *OnboardingValidator) checkCountryCompliance(country string) error {
	// In production, this would check regulatory requirements
	// For now, basic validation
	restricted := []string{"XX", "YY"} // Example restricted countries
	if slices.Contains(restricted, strings.ToUpper(country)) {
		return fmt.Errorf("country not supported")
	}
	return nil
}

// hasCarrierCoverage checks if carrier coverage exists
func (v *OnboardingValidator) hasCarrierCoverage(country string) bool {
	// In production, this would check carrier availability
	// For now, assume most countries have coverage
	supported := []string{"US", "GB", "DE", "FR", "JP", "AU", "CA"}
	return slices.Contains(supported, strings.ToUpper(country))
}
