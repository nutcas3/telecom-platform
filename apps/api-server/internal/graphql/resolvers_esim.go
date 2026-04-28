package graphql

import (
	"context"
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// ProvisionESIM provisions an eSIM profile for the given IMSI.
func (r *Resolver) ProvisionESIM(ctx context.Context, imsi string) (*ESIMProvisionResult, error) {
	sub, err := r.subscriber.GetSubscriberByIMSI(ctx, models.IMSI(imsi))
	if err != nil {
		return nil, fmt.Errorf("subscriber not found for IMSI %s: %w", imsi, err)
	}
	result, err := r.es2Service.ProvisionProfile(ctx, sub)
	if err != nil {
		return nil, err
	}
	return &ESIMProvisionResult{
		ProfileID:      result.ProfileID,
		ActivationCode: result.Activation.ActivationCode,
	}, nil
}

// ActivateESIM activates the given eSIM profile.
func (r *Resolver) ActivateESIM(ctx context.Context, imsi string, profileId string) (bool, error) {
	err := r.es2Service.ActivateProfile(ctx, imsi, profileId)
	return err == nil, err
}

// DeactivateESIM deactivates the given eSIM profile.
func (r *Resolver) DeactivateESIM(ctx context.Context, imsi string, profileId string) (bool, error) {
	err := r.es2Service.DeactivateProfile(ctx, imsi, profileId)
	return err == nil, err
}

// AddPaymentMethod adds a payment method for the given subscriber.
func (r *Resolver) AddPaymentMethod(ctx context.Context, subscriberId int, input PaymentMethodInput) (*models.PaymentMethod, error) {
	isDefault := false
	if input.IsDefault != nil {
		isDefault = *input.IsDefault
	}
	return r.subscriber.AddPaymentMethod(ctx, subscriberId, &models.AddPaymentMethodRequest{
		Type:      models.PaymentMethodType(input.Type),
		Token:     input.Token,
		IsDefault: isDefault,
	})
}

// RemovePaymentMethod removes a payment method.
func (r *Resolver) RemovePaymentMethod(ctx context.Context, paymentMethodId string) (bool, error) {
	return r.subscriber.RemovePaymentMethod(ctx, paymentMethodId)
}

// SetDefaultPaymentMethod marks a payment method as default.
func (r *Resolver) SetDefaultPaymentMethod(ctx context.Context, paymentMethodId string) (*models.PaymentMethod, error) {
	return r.subscriber.SetDefaultPaymentMethod(ctx, paymentMethodId)
}
