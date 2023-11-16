package commercetools

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func resourceCustomObject() *schema.Resource {
	return &schema.Resource{
		Description: "Custom objects are a way to store arbitrary JSON-formatted data on the commercetools platform. " +
			"It allows you to persist data that does not fit the standard data model. This frees your application " +
			"completely from any third-party persistence solution and means that all your data stays on the " +
			"commercetools platform.\n\n" +
			"See also the [Custom Object API Documentation](https://docs.commercetools.com/api/projects/custom-objects)",
		CreateContext: resourceCustomObjectCreate,
		ReadContext:   resourceCustomObjectRead,
		UpdateContext: resourceCustomObjectUpdate,
		DeleteContext: resourceCustomObjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"container": {
				Description: "A namespace to group custom objects matching the pattern '[-_~.a-zA-Z0-9]+'",
				Type:        schema.TypeString,
				Required:    true,
			},
			"key": {
				Description: "String matching the pattern '[-_~.a-zA-Z0-9]+'",
				Type:        schema.TypeString,
				Required:    true,
			},
			"value": {
				Description: "JSON types Number, String, Boolean, Array, Object",
				Type:        schema.TypeString,
				Required:    true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceCustomObjectCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	value := _decodeCustomObjectValue(d.Get("value").(string))

	draft := platform.CustomObjectDraft{
		Container: d.Get("container").(string),
		Key:       d.Get("key").(string),
		Value:     value,
	}
	var customObject *platform.CustomObject
	err := retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
		var err error
		customObject, err = client.CustomObjects().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(customObject.ID)
	_ = d.Set("version", customObject.Version)

	return nil
}

func resourceCustomObjectRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	container := d.Get("container").(string)
	key := d.Get("key").(string)
	client := getClient(m)
	customObject, err := client.CustomObjects().WithContainerAndKey(container, key).Get().Execute(ctx)
	if err != nil {
		if utils.IsResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("container", customObject.Container)
	_ = d.Set("key", customObject.Key)
	_ = d.Set("value", flattenCustomObjectValue(customObject))
	_ = d.Set("version", customObject.Version)
	return nil
}

func resourceCustomObjectUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	value := _decodeCustomObjectValue(d.Get("value").(string))
	originalKey, newKey := d.GetChange("key")
	originalContainer, newContainer := d.GetChange("container")
	originalVersion, _ := d.GetChange("version")

	if d.HasChange("container") || d.HasChange("key") {
		// If the container or key has changed we need to delete the old object
		// and create the new object. We first want to create the new value and
		// then the old one
		draft := platform.CustomObjectDraft{
			Container: newContainer.(string),
			Key:       newKey.(string),
			Value:     value,
		}
		var customObject *platform.CustomObject
		err := retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
			var err error
			customObject, err = client.CustomObjects().Post(draft).Execute(ctx)
			return utils.ProcessRemoteError(err)
		})
		if err != nil {
			// Workaround invalid state to be written, see
			// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
			d.Partial(true)
			return diag.FromErr(err)
		}
		d.SetId(customObject.ID)
		_ = d.Set("version", customObject.Version)

		err = retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
			_, err := client.
				CustomObjects().
				WithContainerAndKey(originalContainer.(string), originalKey.(string)).
				Delete().
				Version(originalVersion.(int)).
				DataErasure(true).
				Execute(ctx)
			return utils.ProcessRemoteError(err)
		})
		if err != nil {
			// Workaround invalid state to be written, see
			// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
			d.Partial(true)
			return diag.FromErr(err)
		}
	} else {
		// Update the value by creating an object with the same key/value.
		// Commercetools will then update the value of the object if it already
		// exists
		draft := platform.CustomObjectDraft{
			Container: d.Get("container").(string),
			Key:       d.Get("key").(string),
			Value:     value,
			Version:   intRef(d.Get("version")),
		}
		var customObject *platform.CustomObject
		err := retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
			var err error
			customObject, err = client.CustomObjects().Post(draft).Execute(ctx)
			return utils.ProcessRemoteError(err)
		})
		if err != nil {
			// Workaround invalid state to be written, see
			// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
			d.Partial(true)
			return diag.FromErr(err)
		}

		d.SetId(customObject.ID)
		_ = d.Set("version", customObject.Version)

	}
	return nil
}

func resourceCustomObjectDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	container := d.Get("container").(string)
	key := d.Get("key").(string)

	client := getClient(m)

	// Lock to prevent concurrent updates due to Version number conflicts
	ctMutexKV.Lock(d.Id())
	defer ctMutexKV.Unlock(d.Id())

	customObject, err := client.
		CustomObjects().
		WithContainerAndKey(container, key).
		Get().
		Execute(ctx)
	if err != nil {
		var diags diag.Diagnostics
		diags = append(diags, diag.FromErr(err)...)
		diags = append(diags, diag.Errorf("could not get custom object with container %s and key %s", container, key)...)
		return diags
	}

	err = retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
		_, err := client.
			CustomObjects().
			WithContainerAndKey(container, key).
			Delete().
			Version(customObject.Version).
			DataErasure(false).
			Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		var diags diag.Diagnostics
		diags = append(diags, diag.FromErr(err)...)
		diags = append(diags, diag.Errorf("could not delete custom object with container %s and key %s", container, key)...)
		return diags
	}
	return nil
}

func _decodeCustomObjectValue(value string) any {
	var data any
	_ = json.Unmarshal([]byte(value), &data)
	return data
}

func flattenCustomObjectValue(o *platform.CustomObject) string {
	val, err := json.Marshal(o.Value)
	if err != nil {
		panic(err)
	}
	return string(val)
}
