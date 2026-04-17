package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: Provider,
	})
}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("TELECOM_API_URL", "http://localhost:8000"),
				Description: "The base URL of the Telecom API server.",
			},
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("TELECOM_API_KEY", nil),
				Description: "API key for authenticating with the Telecom API.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"telecom_subscriber": resourceSubscriber(),
			"telecom_rating_plan": resourceRatingPlan(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"telecom_subscriber":  dataSourceSubscriber(),
			"telecom_system_stats": dataSourceSystemStats(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}
