package services

import (
	"context"
	"fmt"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/id"
	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/mvno"
	"github.com/sirupsen/logrus"
)

// OnboardingService handles MVNO onboarding process
type OnboardingService struct {
	logger      *logrus.Logger
	validator   *mvno.OnboardingValidator
	provisioner *mvno.ProductionProvisioner
	monitor     *mvno.OnboardingMonitor
}

// NewOnboardingService creates a new onboarding service
func NewOnboardingService(logger *logrus.Logger) *OnboardingService {
	return &OnboardingService{
		logger:    logger,
		validator: mvno.NewOnboardingValidator(logger),
		monitor:   mvno.NewOnboardingMonitor(logger),
		// Note: ProductionProvisioner will be initialized with real services in main.go
	}
}

// StartOnboarding initiates the MVNO onboarding process
func (s *OnboardingService) StartOnboarding(ctx context.Context, req *mvno.OnboardingRequest) (*mvno.MVNO, error) {
	// Validate the onboarding request
	if err := s.validator.ValidateRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create MVNO record
	mvnoRecord := &mvno.MVNO{
		ID:         id.GeneratePrefixed("mvno"),
		BusinessID: req.BusinessID,
		Name:       req.BusinessName,
		Status:     mvno.StatusPending,
		Plan:       req.Plan,
		Config: mvno.MVNOConfig{
			MaxSubscribers:    s.getMaxSubscribersForPlan(req.Plan),
			AllowedCountries:  req.TargetCountries,
			CustomBranding:    req.Plan != mvno.PlanStarter,
			APIAccess:         req.Plan != mvno.PlanStarter,
			AdvancedAnalytics: req.Plan == mvno.PlanScale || req.Plan == mvno.PlanEnterprise,
		},
		CreatedAt: time.Now(),
	}

	// Start onboarding progress tracking
	progress := &mvno.OnboardingProgress{
		MVNOID:    mvnoRecord.ID,
		Steps:     s.getOnboardingSteps(),
		Progress:  0.0,
		StartedAt: time.Now(),
	}

	s.logger.WithFields(logrus.Fields{
		"mvno_id":     mvnoRecord.ID,
		"business_id": req.BusinessID,
		"plan":        req.Plan,
	}).Info("Starting MVNO onboarding")

	// Execute onboarding steps asynchronously
	go s.executeOnboarding(ctx, mvnoRecord, progress)

	return mvnoRecord, nil
}

// executeOnboarding runs all onboarding steps
func (s *OnboardingService) executeOnboarding(ctx context.Context, mvno *mvno.MVNO, progress *mvno.OnboardingProgress) {
	for i, step := range progress.Steps {
		select {
		case <-ctx.Done():
			s.logger.WithField("mvno_id", mvno.ID).Error("Onboarding cancelled")
			return
		default:
		}

		step.Status = "running"
		s.monitor.UpdateProgress(mvno.ID, progress)

		if err := s.executeStep(ctx, mvno, &step); err != nil {
			step.Status = "failed"
			step.Error = err.Error()
			s.logger.WithError(err).WithField("step", step.Name).Error("Step failed")
			break
		}

		step.Status = "completed"
		step.CompletedAt = time.Now()
		progress.Progress = float64(i+1) / float64(len(progress.Steps)) * 100

		s.monitor.UpdateProgress(mvno.ID, progress)
	}

	if progress.Progress == 100.0 {
		mvno.Status = "active"
		progress.CompletedAt = time.Now()
		s.logger.WithField("mvno_id", mvno.ID).Info("Onboarding completed")
	}
}

// executeStep executes a single onboarding step
func (s *OnboardingService) executeStep(ctx context.Context, mvno *mvno.MVNO, step *mvno.OnboardingStep) error {
	switch step.Name {
	case "validation":
		return s.validator.ValidateMVNO(ctx, mvno)
	case "provisioning":
		return s.provisioner.ProvisionResources(ctx, mvno)
	case "carrier_setup":
		return s.provisioner.SetupCarriers(ctx, mvno)
	case "billing_setup":
		return s.provisioner.SetupBilling(ctx, mvno)
	case "api_access":
		return s.provisioner.SetupAPIAccess(ctx, mvno)
	default:
		return fmt.Errorf("unknown step: %s", step.Name)
	}
}

// getOnboardingSteps returns the standard onboarding workflow
func (s *OnboardingService) getOnboardingSteps() []mvno.OnboardingStep {
	return []mvno.OnboardingStep{
		{Name: "validation", Status: "pending"},
		{Name: "provisioning", Status: "pending"},
		{Name: "carrier_setup", Status: "pending"},
		{Name: "billing_setup", Status: "pending"},
		{Name: "api_access", Status: "pending"},
	}
}

// getMaxSubscribersForPlan returns subscriber limits per plan
func (s *OnboardingService) getMaxSubscribersForPlan(plan mvno.MVNOPlan) int {
	switch plan {
	case mvno.PlanStarter:
		return 1000
	case mvno.PlanGrowth:
		return 10000
	case mvno.PlanScale:
		return 100000
	case mvno.PlanEnterprise:
		return -1 // Unlimited
	default:
		return 1000
	}
}
