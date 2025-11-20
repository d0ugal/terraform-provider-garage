package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"scheme": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "http",
				Description: "The scheme to use for the Garage admin API",
			},
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The host and port for the Garage admin API (e.g., 127.0.0.1:3903)",
			},
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The admin token for the Garage admin API",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"garage_key":        resourceGarageKey(),
			"garage_bucket":     resourceGarageBucket(),
			"garage_bucket_key": resourceGarageBucketKey(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	scheme := d.Get("scheme").(string)
	host := d.Get("host").(string)
	token := d.Get("token").(string)

	client, err := NewGarageClient(scheme, host, token)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("failed to create Garage client: %w", err))
	}

	return client, nil
}
