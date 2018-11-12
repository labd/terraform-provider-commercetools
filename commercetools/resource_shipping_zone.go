package commercetools

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
	"github.com/labd/commercetools-go-sdk/service/shippingzones"
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
	svc := getShippingZoneService(m)
	var shippingZone *shippingzones.ShippingZone

	// input := d.Get("location").([]interface{})
	// locations := resourceShippingZoneGetLocation(input)
	draft := &shippingzones.ShippingZoneDraft{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error

		shippingZone, err = svc.Create(draft)
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
	svc := getShippingZoneService(m)

	shippingZone, err := svc.GetByID(d.Id())

	if err != nil {
		if ctErr, ok := err.(commercetools.Error); ok {
			if ctErr.Code() == commercetools.ErrResourceNotFound {
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
		d.Set("name", shippingZone.Name)
		d.Set("description", shippingZone.Description)
	}
	return nil
}

func resourceShippingZoneUpdate(d *schema.ResourceData, m interface{}) error {
	svc := getShippingZoneService(m)

	input := &shippingzones.UpdateInput{
		ID:      d.Id(),
		Version: d.Get("version").(int),
		Actions: commercetools.UpdateActions{},
	}

	if d.HasChange("name") {
		newName := d.Get("name").(string)
		input.Actions = append(
			input.Actions,
			&shippingzones.ChangeName{Name: newName})
	}

	if d.HasChange("description") {
		newDescription := d.Get("description").(string)
		input.Actions = append(
			input.Actions,
			&shippingzones.SetDescription{Description: newDescription})
	}

	fmt.Println("TEST")
	if d.HasChange("location") {
		old, new := d.GetChange("location")
		log.Println(old)
		log.Println(new)
		fmt.Println(old)
		fmt.Println(new)

		oldLocations := resourceShippingZoneGetLocation(old)
		newLocations := resourceShippingZoneGetLocation(new)

		for _, location := range oldLocations {
			if !_locationInSlice(location, newLocations) {
				input.Actions = append(
					input.Actions,
					&shippingzones.RemoveLocation{Location: location})
			}
		}
		for _, location := range newLocations {
			if !_locationInSlice(location, oldLocations) {
				input.Actions = append(
					input.Actions,
					&shippingzones.AddLocation{Location: location})
			}
		}
		log.Println(oldLocations)
		log.Println(newLocations)
		fmt.Println(oldLocations)
		fmt.Println(newLocations)
	}

	_, err := svc.Update(input)
	if err != nil {
		return err
	}

	return resourceShippingZoneRead(d, m)
}

func resourceShippingZoneDelete(d *schema.ResourceData, m interface{}) error {
	svc := getShippingZoneService(m)
	version := d.Get("version").(int)
	_, err := svc.DeleteByID(d.Id(), version)
	if err != nil {
		return err
	}

	return nil
}

func getShippingZoneService(m interface{}) *shippingzones.Service {
	client := m.(*commercetools.Client)
	svc := shippingzones.New(client)
	return svc
}

func resourceShippingZoneGetLocation(input interface{}) []shippingzones.Location {
	inputSlice := input.([]interface{})
	var result []shippingzones.Location

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

		result = append(result, shippingzones.Location{
			Country: country,
			State:   state,
		})
	}

	return result
}

func _locationInSlice(needle shippingzones.Location, haystack []shippingzones.Location) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
