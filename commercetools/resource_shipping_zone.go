package commercetools

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func resourceShippingZone() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceShippingZoneCreate,
		ReadContext:   resourceShippingZoneRead,
		UpdateContext: resourceShippingZoneUpdate,
		DeleteContext: resourceShippingZoneDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"key": {
				Description: "User-specific unique identifier for a zone. Must be unique across a project",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"location": {
				Description: "[Location](https://docs.commercetoolstools.pi/projects/zones#location)",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"country": {
							Description: "A two-digit country code as per " +
								"[ISO 3166-1 alpha-2](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2)",
							Type:     schema.TypeString,
							Required: true,
						},
						"state": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceShippingZoneCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	input := d.Get("location").(*schema.Set)
	locations := expandShippingZoneLocations(input)

	draft := platform.ZoneDraft{
		Name:        d.Get("name").(string),
		Description: stringRef(d.Get("description")),
		Locations:   locations,
	}

	key := stringRef(d.Get("key"))
	if *key != "" {
		draft.Key = key
	}

	var shippingZone *platform.Zone
	err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		var err error
		shippingZone, err = client.Zones().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(shippingZone.ID)
	d.Set("version", shippingZone.Version)

	return resourceShippingZoneRead(ctx, d, m)
}

func resourceShippingZoneRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	shippingZone, err := client.Zones().WithId(d.Id()).Get().Execute(ctx)

	if err != nil {
		if utils.IsResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set("version", shippingZone.Version)
	d.Set("key", shippingZone.Key)
	d.Set("name", shippingZone.Name)
	d.Set("description", shippingZone.Description)
	d.Set("location", flattenShippingZoneLocations(shippingZone.Locations))
	return nil
}

func resourceShippingZoneUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	ctMutexKV.Lock(d.Id())
	defer ctMutexKV.Unlock(d.Id())

	input := platform.ZoneUpdate{
		Version: d.Get("version").(int),
		Actions: []platform.ZoneUpdateAction{},
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&platform.ZoneSetKeyAction{Key: &newKey})
	}
	if d.HasChange("name") {
		newName := d.Get("name").(string)
		input.Actions = append(
			input.Actions,
			&platform.ZoneChangeNameAction{Name: newName})
	}

	if d.HasChange("description") {
		newDescription := d.Get("description").(string)
		input.Actions = append(
			input.Actions,
			&platform.ZoneSetDescriptionAction{Description: &newDescription})
	}

	if d.HasChange("location") {
		old, new := d.GetChange("location")

		oldLocations := expandShippingZoneLocations(old.(*schema.Set))
		newLocations := expandShippingZoneLocations(new.(*schema.Set))

		for i, location := range oldLocations {
			if !_locationInSlice(location, newLocations) {
				input.Actions = append(
					input.Actions,
					&platform.ZoneRemoveLocationAction{Location: oldLocations[i]})
			}
		}
		for i, location := range newLocations {
			if !_locationInSlice(location, oldLocations) {
				input.Actions = append(
					input.Actions,
					&platform.ZoneAddLocationAction{Location: newLocations[i]})
			}
		}
	}

	_, err := client.Zones().WithId(d.Id()).Post(input).Execute(ctx)
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	return resourceShippingZoneRead(ctx, d, m)
}

func resourceShippingZoneDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	// Lock to prevent concurrent updates due to Version number conflicts
	ctMutexKV.Lock(d.Id())
	defer ctMutexKV.Unlock(d.Id())

	version := d.Get("version").(int)
	err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		_, err := client.Zones().WithId(d.Id()).Delete().Version(version).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	return diag.FromErr(err)
}

func expandShippingZoneLocations(input *schema.Set) []platform.Location {
	inputSlice := input.List()
	result := make([]platform.Location, len(inputSlice))

	for i := range inputSlice {
		raw := inputSlice[i].(map[string]any)

		country, ok := raw["country"].(string)
		if !ok {
			country = ""
		}

		var stateRef *string
		if state, ok := raw["state"].(string); ok && state != "" {
			stateRef = &state
		}

		result[i] = platform.Location{
			Country: country,
			State:   stateRef,
		}
	}

	return result
}

func flattenShippingZoneLocations(locations []platform.Location) []map[string]any {
	result := make([]map[string]any, len(locations))

	for i := range locations {
		result[i] = map[string]any{
			"country": locations[i].Country,
			"state":   locations[i].State,
		}
	}

	return result
}

func _locationInSlice(needle platform.Location, haystack []platform.Location) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
