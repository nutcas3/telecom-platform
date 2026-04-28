package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type Subscriber struct {
	ID             int     `json:"id"`
	IMSI           string  `json:"imsi"`
	MSISDN         string  `json:"msisdn"`
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
	Email          string  `json:"email"`
	OrganizationID string  `json:"organization_id"`
	Status         string  `json:"status"`
	PlanID         int     `json:"plan_id"`
	Balance        float64 `json:"balance"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

type RatingPlan struct {
	PlanID     string  `json:"plan_id"`
	Name       string  `json:"name"`
	DataRate   float64 `json:"data_rate"`
	VoiceRate  float64 `json:"voice_rate"`
	SMSRate    float64 `json:"sms_rate"`
	MonthlyFee float64 `json:"monthly_fee"`
	DataLimit  float64 `json:"data_limit"`
	VoiceLimit float64 `json:"voice_limit"`
	SMSLimit   float64 `json:"sms_limit"`
}

func resourceSubscriber() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSubscriberCreate,
		ReadContext:   resourceSubscriberRead,
		UpdateContext: resourceSubscriberUpdate,
		DeleteContext: resourceSubscriberDelete,
		Schema: map[string]*schema.Schema{
			"imsi": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"msisdn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"first_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"last_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"organization_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"plan_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"balance": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSubscriberCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*APIClient)

	sub := map[string]any{
		"imsi":            d.Get("imsi").(string),
		"msisdn":          d.Get("msisdn").(string),
		"first_name":      d.Get("first_name").(string),
		"last_name":       d.Get("last_name").(string),
		"email":           d.Get("email").(string),
		"organization_id": d.Get("organization_id").(string),
		"plan_id":         d.Get("plan_id").(int),
	}

	var result Subscriber
	if err := client.post(ctx, "/api/subscribers", sub, &result); err != nil {
		return diag.FromErr(fmt.Errorf("failed to create subscriber: %w", err))
	}

	d.SetId(strconv.Itoa(result.ID))
	return resourceSubscriberRead(ctx, d, meta)
}

func resourceSubscriberRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*APIClient)
	id := d.Id()

	var result Subscriber
	if err := client.get(ctx, "/api/subscribers/"+id, &result); err != nil {
		return diag.FromErr(fmt.Errorf("failed to read subscriber: %w", err))
	}

	d.Set("imsi", result.IMSI)
	d.Set("msisdn", result.MSISDN)
	d.Set("first_name", result.FirstName)
	d.Set("last_name", result.LastName)
	d.Set("email", result.Email)
	d.Set("organization_id", result.OrganizationID)
	d.Set("plan_id", result.PlanID)
	d.Set("status", result.Status)
	d.Set("balance", result.Balance)
	d.Set("created_at", result.CreatedAt)
	d.Set("updated_at", result.UpdatedAt)

	return nil
}

func resourceSubscriberUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*APIClient)
	id := d.Id()

	sub := map[string]any{
		"first_name":      d.Get("first_name").(string),
		"last_name":       d.Get("last_name").(string),
		"email":           d.Get("email").(string),
		"organization_id": d.Get("organization_id").(string),
		"plan_id":         d.Get("plan_id").(int),
	}

	var result Subscriber
	if err := client.put(ctx, "/api/subscribers/"+id, sub, &result); err != nil {
		return diag.FromErr(fmt.Errorf("failed to update subscriber: %w", err))
	}

	return resourceSubscriberRead(ctx, d, meta)
}

func resourceSubscriberDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*APIClient)
	id := d.Id()

	if err := client.delete(ctx, "/api/subscribers/"+id); err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete subscriber: %w", err))
	}

	return nil
}

func resourceRatingPlan() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRatingPlanCreate,
		ReadContext:   resourceRatingPlanRead,
		UpdateContext: resourceRatingPlanUpdate,
		DeleteContext: resourceRatingPlanDelete,
		Schema: map[string]*schema.Schema{
			"plan_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"data_rate": {
				Type:     schema.TypeFloat,
				Required: true,
			},
			"voice_rate": {
				Type:     schema.TypeFloat,
				Required: true,
			},
			"sms_rate": {
				Type:     schema.TypeFloat,
				Required: true,
			},
			"monthly_fee": {
				Type:     schema.TypeFloat,
				Required: true,
			},
			"data_limit": {
				Type:     schema.TypeFloat,
				Required: true,
			},
			"voice_limit": {
				Type:     schema.TypeFloat,
				Required: true,
			},
			"sms_limit": {
				Type:     schema.TypeFloat,
				Required: true,
			},
		},
	}
}

func resourceRatingPlanCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Placeholder - would implement actual plan creation
	d.SetId(d.Get("plan_id").(string))
	return resourceRatingPlanRead(ctx, d, meta)
}

func resourceRatingPlanRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Placeholder - would fetch actual plan data
	d.Set("plan_id", d.Get("plan_id"))
	d.Set("name", d.Get("name"))
	d.Set("data_rate", d.Get("data_rate"))
	d.Set("voice_rate", d.Get("voice_rate"))
	d.Set("sms_rate", d.Get("sms_rate"))
	d.Set("monthly_fee", d.Get("monthly_fee"))
	d.Set("data_limit", d.Get("data_limit"))
	d.Set("voice_limit", d.Get("voice_limit"))
	d.Set("sms_limit", d.Get("sms_limit"))
	return nil
}

func resourceRatingPlanUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return resourceRatingPlanRead(ctx, d, meta)
}

func resourceRatingPlanDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return nil
}
