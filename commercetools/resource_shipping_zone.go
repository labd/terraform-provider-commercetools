package commercetools

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
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
				Type:     schema.TypeString,
				Optional: true,
			},
			"location": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"country": {
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

	var shippingZone *commercetools.Zone

	input := d.Get("location").([]interface{})
	locations := resourceShippingZoneGetLocation(input)

	draft := &commercetools.ZoneDraft{
		Key:         d.Get("key").(string),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Locations:   locations,
	}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error

		shippingZone, err = client.ZoneCreate(draft)
		if err != nil {
			return resource.RetryableError(err)
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

	shippingZone, err := client.ZoneGetWithID(d.Id())

	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
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

	input := &commercetools.ZoneUpdateWithIDInput{
		ID:      d.Id(),
		Version: d.Get("version").(int),
		Actions: []commercetools.ZoneUpdateAction{},
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.ZoneSetKeyAction{Key: newKey})
	}
	if d.HasChange("name") {
		newName := d.Get("name").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.ZoneChangeNameAction{Name: newName})
	}

	if d.HasChange("description") {
		newDescription := d.Get("description").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.ZoneSetDescriptionAction{Description: newDescription})
	}

	if d.HasChange("location") {
		old, new := d.GetChange("location")

		oldLocations := resourceShippingZoneGetLocation(old)
		newLocations := resourceShippingZoneGetLocation(new)

		for i, location := range oldLocations {
			if !_locationInSlice(location, newLocations) {
				input.Actions = append(
					input.Actions,
					&commercetools.ZoneRemoveLocationAction{Location: &oldLocations[i]})
			}
		}
		for i, location := range newLocations {
			if !_locationInSlice(location, oldLocations) {
				input.Actions = append(
					input.Actions,
					&commercetools.ZoneAddLocationAction{Location: &newLocations[i]})
			}
		}
	}

	_, err := client.ZoneUpdateWithID(input)
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
	_, err := client.ZoneDeleteWithID(d.Id(), version)
	if err != nil {
		return err
	}

	return nil
}

func resourceShippingZoneGetLocation(input interface{}) []commercetools.Location {
	inputSlice := input.([]interface{})
	var result []commercetools.Location

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

		result = append(result, commercetools.Location{
			Country: commercetools.CountryCode(country),
			State:   state,
		})
	}

	return result
}

func _locationInSlice(needle commercetools.Location, haystack []commercetools.Location) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
