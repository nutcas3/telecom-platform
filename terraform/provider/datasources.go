package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type SystemStats struct {
	ActiveSessions   int     `json:"active_sessions"`
	TotalAccounts    int     `json:"total_accounts"`
	BlockedUsers     int     `json:"blocked_users"`
	LowBalanceAlerts int     `json:"low_balance_alerts"`
	Uptime           float64 `json:"uptime"`
}

func dataSourceSubscriber() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSubscriberRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"imsi": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"msisdn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"first_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organization_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"plan_id": {
				Type:     schema.TypeInt,
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

func dataSourceSubscriberRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*APIClient)
	id := d.Get("id").(string)

	var result Subscriber
	if err := client.get(ctx, "/api/subscribers/"+id, &result); err != nil {
		return diag.FromErr(fmt.Errorf("failed to read subscriber: %w", err))
	}

	d.SetId(strconv.Itoa(result.ID))
	d.Set("imsi", result.IMSI)
	d.Set("msisdn", result.MSISDN)
	d.Set("first_name", result.FirstName)
	d.Set("last_name", result.LastName)
	d.Set("email", result.Email)
	d.Set("organization_id", result.OrganizationID)
	d.Set("status", result.Status)
	d.Set("plan_id", result.PlanID)
	d.Set("balance", result.Balance)
	d.Set("created_at", result.CreatedAt)
	d.Set("updated_at", result.UpdatedAt)

	return nil
}

func dataSourceSystemStats() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSystemStatsRead,
		Schema: map[string]*schema.Schema{
			"active_sessions": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"total_accounts": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"blocked_users": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"low_balance_alerts": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"uptime": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
		},
	}
}

func dataSourceSystemStatsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*APIClient)

	var result SystemStats
	if err := client.get(ctx, "/api/system/stats", &result); err != nil {
		return diag.FromErr(fmt.Errorf("failed to read system stats: %w", err))
	}

	d.SetId("system-stats")
	d.Set("active_sessions", result.ActiveSessions)
	d.Set("total_accounts", result.TotalAccounts)
	d.Set("blocked_users", result.BlockedUsers)
	d.Set("low_balance_alerts", result.LowBalanceAlerts)
	d.Set("uptime", result.Uptime)

	return nil
}
