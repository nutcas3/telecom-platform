package mvno

import (
	"context"
	"fmt"
	"time"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/id"
	"github.com/sirupsen/logrus"
)

// OnboardingService handles MVNO onboarding process
type OnboardingService struct {
	logger      *logrus.Logger
	validator   *OnboardingValidator
	provisioner *MVNOProvisioner
	monitor     *OnboardingMonitor
}

// NewOnboardingService creates a new onboarding service
func NewOnboardingService(logger *logrus.Logger) *OnboardingService {
	return &OnboardingService{
		logger:      logger,
		validator:   NewOnboardingValidator(logger),
		provisioner: NewMVNOProvisioner(logger),
		monitor:     NewOnboardingMonitor(logger),
	}
}

// StartOnboarding initiates the MVNO onboarding process
func (s *OnboardingService) StartOnboarding(ctx context.Context, req *OnboardingRequest) (*MVNO, error) {
	// Validate the onboarding request
	if err := s.validator.ValidateRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create MVNO record
	mvno := &MVNO{
		ID:         id.GeneratePrefixed("mvno"),
		BusinessID: req.BusinessID,
		Name:       req.BusinessName,
		Status:     StatusPending,
		Plan:       req.Plan,
		Config: MVNOConfig{
			MaxSubscribers:    s.getMaxSubscribersForPlan(req.Plan),
			AllowedCountries:  req.TargetCountries,
			CustomBranding:    req.Plan != PlanStarter,
			APIAccess:         req.Plan != PlanStarter,
			AdvancedAnalytics: req.Plan == PlanScale || req.Plan == PlanEnterprise,
		},
		CreatedAt: time.Now(),
	}

	// Start onboarding progress tracking
	progress := &OnboardingProgress{
		MVNOID:    mvno.ID,
		Steps:     s.getOnboardingSteps(),
		Progress:  0.0,
		StartedAt: time.Now(),
	}

	s.logger.WithFields(logrus.Fields{
		"mvno_id":     mvno.ID,
		"business_id": req.BusinessID,
		"plan":        req.Plan,
	}).Info("Starting MVNO onboarding")

	// Execute onboarding steps asynchronously
	go s.executeOnboarding(ctx, mvno, progress)

	return mvno, nil
}

// executeOnboarding runs all onboarding steps
func (s *OnboardingService) executeOnboarding(ctx context.Context, mvno *MVNO, progress *OnboardingProgress) {
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
		mvno.Status = StatusActive
		progress.CompletedAt = time.Now()
		s.logger.WithField("mvno_id", mvno.ID).Info("Onboarding completed")
	}
}

// executeStep executes a single onboarding step
func (s *OnboardingService) executeStep(ctx context.Context, mvno *MVNO, step *OnboardingStep) error {
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
func (s *OnboardingService) getOnboardingSteps() []OnboardingStep {
	return []OnboardingStep{
		{Name: "validation", Status: "pending"},
		{Name: "provisioning", Status: "pending"},
		{Name: "carrier_setup", Status: "pending"},
		{Name: "billing_setup", Status: "pending"},
		{Name: "api_access", Status: "pending"},
	}
}

// getMaxSubscribersForPlan returns subscriber limits per plan
func (s *OnboardingService) getMaxSubscribersForPlan(plan MVNOPlan) int {
	switch plan {
	case PlanStarter:
		return 1000
	case PlanGrowth:
		return 10000
	case PlanScale:
		return 100000
	case PlanEnterprise:
		return -1 // Unlimited
	default:
		return 1000
	}
}
