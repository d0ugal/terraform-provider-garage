package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	garage "git.deuxfleurs.fr/garage-sdk/garage-admin-sdk-golang"
)

func resourceGarageKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGarageKeyCreate,
		ReadContext:   resourceGarageKeyRead,
		UpdateContext: resourceGarageKeyUpdate,
		DeleteContext: resourceGarageKeyDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the access key",
			},
			"access_key_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The access key ID",
			},
			"secret_access_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The secret access key (only available on create)",
			},
		},
	}
}

func resourceGarageKeyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*GarageClient)
	name := d.Get("name").(string)

	keyBody := garage.NewUpdateKeyRequestBody()
	keyBody.SetName(name)

	key, resp, err := client.Client.AccessKeyAPI.CreateKey(ctx).Body(*keyBody).Execute()
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create key: %w", err))
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	d.SetId(key.AccessKeyId)
	if err := d.Set("access_key_id", key.AccessKeyId); err != nil {
		return diag.FromErr(err)
	}
	if key.SecretAccessKey.IsSet() {
		secret := key.SecretAccessKey.Get()
		if secret != nil {
			if err := d.Set("secret_access_key", *secret); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return nil
}

func resourceGarageKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*GarageClient)
	keyID := d.Id()

	key, resp, err := client.Client.AccessKeyAPI.GetKeyInfo(ctx).Id(keyID).Execute()
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("failed to read key: %w", err))
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	if err := d.Set("access_key_id", key.AccessKeyId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", key.Name); err != nil {
		return diag.FromErr(err)
	}
	// Note: secret_access_key is not available on read, only on create

	return nil
}

func resourceGarageKeyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Garage doesn't support updating key names, so we recreate if name changes
	if d.HasChange("name") {
		// Delete and recreate
		diags := resourceGarageKeyDelete(ctx, d, m)
		if diags.HasError() {
			return diags
		}
		return resourceGarageKeyCreate(ctx, d, m)
	}
	return nil
}

func resourceGarageKeyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Note: Garage API doesn't have a delete key endpoint in v1
	// We'll just remove from state
	d.SetId("")
	return nil
}

