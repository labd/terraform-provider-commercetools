package commercetools

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
)

func resourceShippingZone() *schema.Resource {
	return &schema.Resource{
		Create: resourceShippingZoneCreate,
		Read:   resourceShippingZoneRead,
		Update: resourceShippingZoneUpdate,
		Delete: resourceShippingZoneDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
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
				Type:        schema.TypeList,
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

func resourceShippingZoneCreate(d *schema.ResourceData, m interface{}) error {
	log.Print("[DEBUG] Creating shippingzones in commercetools")
	client := getClient(m)

	var shippingZone *platform.Zone

	input := d.Get("location").([]interface{})
	locations := resourceShippingZoneGetLocation(input)

	draft := platform.ZoneDraft{
		Key:         stringRef(d.Get("key")),
		Name:        d.Get("name").(string),
		Description: stringRef(d.Get("description")),
		Locations:   locations,
	}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error

		shippingZone, err = client.Zones().Post(draft).Execute(context.Background())
		if err != nil {
			return handleCommercetoolsError(err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	if shippingZone == nil {
		return fmt.Errorf("Error creating shipping zone")
	}

	d.SetId(shippingZone.ID)
	d.Set("version", shippingZone.Version)

	return resourceShippingZoneRead(d, m)
}

func resourceShippingZoneRead(d *schema.ResourceData, m interface{}) error {
	log.Print("[DEBUG] Reading shippingzones from commercetools")
	client := getClient(m)

	shippingZone, err := client.Zones().WithId(d.Id()).Get().Execute(context.Background())

	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if shippingZone == nil {
		log.Print("[DEBUG] No shippingzones found")
		d.SetId("")
	} else {
		log.Print("[DEBUG] Found following shippingzones:")
		log.Print(stringFormatObject(shippingZone))

		d.Set("version", shippingZone.Version)
		d.Set("key", shippingZone.Key)
		d.Set("name", shippingZone.Name)
		d.Set("description", shippingZone.Description)
		d.Set("locations", shippingZone.Locations)
	}
	return nil
}

func resourceShippingZoneUpdate(d *schema.ResourceData, m interface{}) error {
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

		oldLocations := resourceShippingZoneGetLocation(old)
		newLocations := resourceShippingZoneGetLocation(new)

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

	_, err := client.Zones().WithId(d.Id()).Post(input).Execute(context.Background())
	if err != nil {
		return err
	}

	return resourceShippingZoneRead(d, m)
}

func resourceShippingZoneDelete(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)

	// Lock to prevent concurrent updates due to Version number conflicts
	ctMutexKV.Lock(d.Id())
	defer ctMutexKV.Unlock(d.Id())

	version := d.Get("version").(int)
	_, err := client.Zones().WithId(d.Id()).Delete().WithQueryParams(platform.ByProjectKeyZonesByIDRequestMethodDeleteInput{
		Version: version,
	}).Execute(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func resourceShippingZoneGetLocation(input interface{}) []platform.Location {
	inputSlice := input.([]interface{})
	var result []platform.Location

	for _, raw := range inputSlice {
		i := raw.(map[string]interface{})

		country, ok := i["country"].(string)
		if !ok {
			country = ""
		}

		state, ok := i["state"].(string)
		if !ok {
			state = ""
		}

		result = append(result, platform.Location{
			Country: country,
			State:   &state,
		})
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
