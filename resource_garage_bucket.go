package main

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"

	garage "git.deuxfleurs.fr/garage-sdk/garage-admin-sdk-golang"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGarageBucket() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGarageBucketCreate,
		ReadContext:   resourceGarageBucketRead,
		UpdateContext: resourceGarageBucketUpdate,
		DeleteContext: resourceGarageBucketDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The bucket ID (computed if not provided)",
			},
			"global_alias": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Global alias for the bucket (this appears as the name in garage bucket list)",
			},
			"bytes": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of bytes used by objects in this bucket",
			},
			"objects": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of objects in this bucket",
			},
			"expiration_days": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Number of days after which objects in this bucket will be automatically deleted. Set to 0 to disable expiration.",
			},
		},
	}
}

func resourceGarageBucketCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*GarageClient)
	globalAlias := d.Get("global_alias").(string)

	bucketInfo := garage.NewCreateBucketRequest()
	if globalAlias != "" {
		bucketInfo.SetGlobalAlias(globalAlias)
	}

	bucket, resp, err := client.Client.BucketAPI.CreateBucket(ctx).CreateBucketRequest(*bucketInfo).Execute()
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create bucket: %w", err))
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	d.SetId(bucket.Id)
	if err := d.Set("id", bucket.Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("bytes", bucket.Bytes); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("objects", bucket.Objects); err != nil {
		return diag.FromErr(err)
	}
	if len(bucket.GlobalAliases) > 0 {
		if err := d.Set("global_alias", bucket.GlobalAliases[0]); err != nil {
			return diag.FromErr(err)
		}
	}

	// Set expiration policy if specified
	if expirationDays, ok := d.GetOk("expiration_days"); ok && expirationDays.(int) > 0 {
		if err := setBucketLifecyclePolicy(ctx, client, bucket.Id, expirationDays.(int)); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set expiration policy: %w", err))
		}
		if err := d.Set("expiration_days", expirationDays.(int)); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceGarageBucketRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*GarageClient)
	bucketID := d.Id()

	bucket, resp, err := client.Client.BucketAPI.GetBucketInfo(ctx).Id(bucketID).Execute()
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("failed to read bucket: %w", err))
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	if err := d.Set("id", bucket.Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("bytes", bucket.Bytes); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("objects", bucket.Objects); err != nil {
		return diag.FromErr(err)
	}
	if len(bucket.GlobalAliases) > 0 {
		if err := d.Set("global_alias", bucket.GlobalAliases[0]); err != nil {
			return diag.FromErr(err)
		}
	}

	// Read expiration policy if it exists
	expirationDays, err := getBucketLifecyclePolicy(ctx, client, bucket.Id)
	if err == nil && expirationDays > 0 {
		if err := d.Set("expiration_days", expirationDays); err != nil {
			return diag.FromErr(err)
		}
	} else if d.HasChange("expiration_days") {
		// If expiration_days was removed, clear it
		if err := d.Set("expiration_days", 0); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceGarageBucketUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*GarageClient)
	bucketID := d.Id()

	// Handle expiration policy changes
	if d.HasChange("expiration_days") {
		expirationDays := d.Get("expiration_days").(int)
		if expirationDays > 0 {
			if err := setBucketLifecyclePolicy(ctx, client, bucketID, expirationDays); err != nil {
				return diag.FromErr(fmt.Errorf("failed to update expiration policy: %w", err))
			}
		} else {
			// Remove lifecycle policy if expiration_days is 0 or removed
			if err := deleteBucketLifecyclePolicy(ctx, client, bucketID); err != nil {
				return diag.FromErr(fmt.Errorf("failed to remove expiration policy: %w", err))
			}
		}
	}

	return resourceGarageBucketRead(ctx, d, m)
}

func resourceGarageBucketDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Note: Garage API v1 doesn't have a delete bucket endpoint
	// We'll just remove from state
	d.SetId("")
	return nil
}

// S3 Lifecycle Configuration structures
type LifecycleConfiguration struct {
	XMLName xml.Name `xml:"LifecycleConfiguration"`
	Rules   []Rule   `xml:"Rule"`
}

type Rule struct {
	ID         string      `xml:"ID"`
	Status     string      `xml:"Status"`
	Filter     *Filter     `xml:"Filter,omitempty"`
	Expiration *Expiration `xml:"Expiration,omitempty"`
}

type Filter struct {
	Prefix string `xml:"Prefix"`
}

type Expiration struct {
	Days int `xml:"Days"`
}

// setBucketLifecyclePolicy sets the lifecycle expiration policy for a bucket using S3-compatible API
func setBucketLifecyclePolicy(ctx context.Context, client *GarageClient, bucketID string, expirationDays int) error {
	// Get bucket info to find the bucket alias/name
	bucket, resp, err := client.Client.BucketAPI.GetBucketInfo(ctx).Id(bucketID).Execute()
	if err != nil {
		return fmt.Errorf("failed to get bucket info: %w", err)
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	// Get bucket name (alias)
	bucketName := bucketID
	if len(bucket.GlobalAliases) > 0 {
		bucketName = bucket.GlobalAliases[0]
	}

	// Construct S3 endpoint URL (Garage S3 API typically uses port 3900)
	// Replace admin port (3903) with S3 port (3900)
	s3URL := fmt.Sprintf("%s://%s/%s?lifecycle",
		client.Client.GetConfig().Scheme,
		replacePort(client.Client.GetConfig().Host, 3900),
		bucketName)

	// Create lifecycle configuration XML
	lifecycleConfig := LifecycleConfiguration{
		Rules: []Rule{
			{
				ID:     "expire-after-days",
				Status: "Enabled",
				Expiration: &Expiration{
					Days: expirationDays,
				},
			},
		},
	}

	xmlData, err := xml.MarshalIndent(lifecycleConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lifecycle config: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "PUT", s3URL, bytes.NewReader(xmlData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/xml")
	// Use admin token for authorization (Garage may accept this for S3 API too)
	if authHeader, ok := client.Client.GetConfig().DefaultHeader["Authorization"]; ok {
		req.Header.Set("Authorization", authHeader)
	}

	// Execute request
	httpClient := &http.Client{}
	resp, err = httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// getBucketLifecyclePolicy retrieves the lifecycle expiration policy for a bucket
func getBucketLifecyclePolicy(ctx context.Context, client *GarageClient, bucketID string) (int, error) {
	// Get bucket info to find the bucket alias/name
	bucket, resp, err := client.Client.BucketAPI.GetBucketInfo(ctx).Id(bucketID).Execute()
	if err != nil {
		return 0, fmt.Errorf("failed to get bucket info: %w", err)
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	// Get bucket name (alias)
	bucketName := bucketID
	if len(bucket.GlobalAliases) > 0 {
		bucketName = bucket.GlobalAliases[0]
	}

	// Construct S3 endpoint URL
	s3URL := fmt.Sprintf("%s://%s/%s?lifecycle",
		client.Client.GetConfig().Scheme,
		replacePort(client.Client.GetConfig().Host, 3900),
		bucketName)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", s3URL, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Use admin token for authorization
	if authHeader, ok := client.Client.GetConfig().DefaultHeader["Authorization"]; ok {
		req.Header.Set("Authorization", authHeader)
	}

	// Execute request
	httpClient := &http.Client{}
	resp, err = httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return 0, nil // No lifecycle policy set
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse XML response
	var lifecycleConfig LifecycleConfiguration
	if err := xml.NewDecoder(resp.Body).Decode(&lifecycleConfig); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract expiration days from first rule
	if len(lifecycleConfig.Rules) > 0 && lifecycleConfig.Rules[0].Expiration != nil {
		return lifecycleConfig.Rules[0].Expiration.Days, nil
	}

	return 0, nil
}

// deleteBucketLifecyclePolicy removes the lifecycle policy from a bucket
func deleteBucketLifecyclePolicy(ctx context.Context, client *GarageClient, bucketID string) error {
	// Get bucket info to find the bucket alias/name
	bucket, resp, err := client.Client.BucketAPI.GetBucketInfo(ctx).Id(bucketID).Execute()
	if err != nil {
		return fmt.Errorf("failed to get bucket info: %w", err)
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	// Get bucket name (alias)
	bucketName := bucketID
	if len(bucket.GlobalAliases) > 0 {
		bucketName = bucket.GlobalAliases[0]
	}

	// Construct S3 endpoint URL
	s3URL := fmt.Sprintf("%s://%s/%s?lifecycle",
		client.Client.GetConfig().Scheme,
		replacePort(client.Client.GetConfig().Host, 3900),
		bucketName)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "DELETE", s3URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Use admin token for authorization
	if authHeader, ok := client.Client.GetConfig().DefaultHeader["Authorization"]; ok {
		req.Header.Set("Authorization", authHeader)
	}

	// Execute request
	httpClient := &http.Client{}
	resp, err = httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// replacePort replaces the port in a host string (e.g., "127.0.0.1:3903" -> "127.0.0.1:3900")
func replacePort(host string, newPort int) string {
	// Simple implementation - if host contains a port, replace it
	// Otherwise, append the new port
	for i := len(host) - 1; i >= 0; i-- {
		if host[i] == ':' {
			return host[:i+1] + strconv.Itoa(newPort)
		}
	}
	return host + ":" + strconv.Itoa(newPort)
}
