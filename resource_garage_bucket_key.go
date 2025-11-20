package main

import (
	"context"
	"fmt"
	"net/http"

	garage "git.deuxfleurs.fr/garage-sdk/garage-admin-sdk-golang"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGarageBucketKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGarageBucketKeyCreate,
		ReadContext:   resourceGarageBucketKeyRead,
		UpdateContext: resourceGarageBucketKeyUpdate,
		DeleteContext: resourceGarageBucketKeyDelete,
		Schema: map[string]*schema.Schema{
			"bucket_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The bucket ID",
			},
			"access_key_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The access key ID",
			},
			"read": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Grant read permission",
			},
			"write": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Grant write permission",
			},
			"owner": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Grant owner permission",
			},
		},
	}
}

func resourceGarageBucketKeyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*GarageClient)
	bucketID := d.Get("bucket_id").(string)
	keyID := d.Get("access_key_id").(string)
	read := d.Get("read").(bool)
	write := d.Get("write").(bool)
	owner := d.Get("owner").(bool)

	perms := garage.NewApiBucketKeyPerm()
	perms.SetRead(read)
	perms.SetWrite(write)
	perms.SetOwner(owner)

	updateReq := garage.NewBucketKeyPermChangeRequest(keyID, bucketID, *perms)

	_, resp, err := client.Client.PermissionAPI.AllowBucketKey(ctx).Body(*updateReq).Execute()
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to update bucket key permissions: %w", err))
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	d.SetId(fmt.Sprintf("%s/%s", bucketID, keyID))

	return resourceGarageBucketKeyRead(ctx, d, m)
}

func resourceGarageBucketKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*GarageClient)
	bucketID := d.Get("bucket_id").(string)
	keyID := d.Get("access_key_id").(string)

	// Get key to check bucket associations
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

	// Find this bucket in the key's bucket list
	for _, bucket := range key.Buckets {
		if bucket.Id == bucketID {
			if bucket.Permissions.Read != nil {
				if err := d.Set("read", *bucket.Permissions.Read); err != nil {
					return diag.FromErr(err)
				}
			}
			if bucket.Permissions.Write != nil {
				if err := d.Set("write", *bucket.Permissions.Write); err != nil {
					return diag.FromErr(err)
				}
			}
			if bucket.Permissions.Owner != nil {
				if err := d.Set("owner", *bucket.Permissions.Owner); err != nil {
					return diag.FromErr(err)
				}
			}
			return nil
		}
	}

	// Key doesn't have permissions on this bucket
	d.SetId("")
	return nil
}

func resourceGarageBucketKeyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*GarageClient)
	bucketID := d.Get("bucket_id").(string)
	keyID := d.Get("access_key_id").(string)
	read := d.Get("read").(bool)
	write := d.Get("write").(bool)
	owner := d.Get("owner").(bool)

	perms := garage.NewApiBucketKeyPerm()
	perms.SetRead(read)
	perms.SetWrite(write)
	perms.SetOwner(owner)

	updateReq := garage.NewBucketKeyPermChangeRequest(keyID, bucketID, *perms)

	_, resp, err := client.Client.PermissionAPI.AllowBucketKey(ctx).Body(*updateReq).Execute()
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to update bucket key permissions: %w", err))
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	return resourceGarageBucketKeyRead(ctx, d, m)
}

func resourceGarageBucketKeyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*GarageClient)
	bucketID := d.Get("bucket_id").(string)
	keyID := d.Get("access_key_id").(string)

	// Remove all permissions by denying them
	perms := garage.NewApiBucketKeyPerm()
	perms.SetRead(false)
	perms.SetWrite(false)
	perms.SetOwner(false)

	updateReq := garage.NewBucketKeyPermChangeRequest(keyID, bucketID, *perms)

	_, resp, err := client.Client.PermissionAPI.DenyBucketKey(ctx).Body(*updateReq).Execute()
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to remove bucket key permissions: %w", err))
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	d.SetId("")
	return nil
}
