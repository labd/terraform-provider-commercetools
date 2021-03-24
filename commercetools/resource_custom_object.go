package commercetools

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

func resourceCustomObject() *schema.Resource {
	return &schema.Resource{
		Description: "Custom objects are a way to store arbitrary JSON-formatted data on the commercetools platform. " +
			"It allows you to persist data that does not fit the standard data model. This frees your application " +
			"completely from any third-party persistence solution and means that all your data stays on the " +
			"commercetools platform.\n\n" +
			"See also the [Custom Object API Documentation](https://docs.commercetools.com/api/projects/custom-objects)",
		Create: resourceCustomObjectCreate,
		Read:   resourceCustomObjectRead,
		Update: resourceCustomObjectUpdate,
		Delete: resourceCustomObjectDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

func resourceCustomObjectCreate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	value := _decodeCustomObjectValue(d.Get("value").(string))

	draft := commercetools.CustomObjectDraft{
		Container: d.Get("container").(string),
		Key:       d.Get("key").(string),
		Value:     value,
	}
	customObject, err := client.CustomObjectCreate(context.Background(), &draft)
	if err != nil {
		return err
	}

	d.SetId(customObject.ID)
	d.Set("version", customObject.Version)

	return nil
}

func resourceCustomObjectRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceCustomObjectUpdate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	value := _decodeCustomObjectValue(d.Get("value").(string))
	ctx := context.Background()

	if d.HasChange("container") || d.HasChange("key") {
		// If the container or key has changed we need to delete the old object
		// and create the new object. We first want to create the new vlaue and
		// then the old one
		draft := commercetools.CustomObjectDraft{
			Container: d.Get("container").(string),
			Key:       d.Get("key").(string),
			Value:     value,
		}
		customObject, err := client.CustomObjectCreate(ctx, &draft)
		if err != nil {
			return err
		}
		d.SetId(customObject.ID)
		d.Set("version", customObject.Version)

		_, err = client.CustomObjectDeleteWithContainerAndKey(
			ctx,
			d.Get("container").(string),
			d.Get("key").(string),
			d.Get("version").(int),
			true,
		)

		// Do we care? Just log an error for now
		if err != nil {
			log.Printf("Failed to remove old custom object")
		}
	} else {

		// Update the value by creating an object with the same key/value.
		// Commercetools will then update the value of the object if it already
		// exists
		draft := commercetools.CustomObjectDraft{
			Container: d.Get("container").(string),
			Key:       d.Get("key").(string),
			Value:     value,
			Version:   d.Get("version").(int),
		}
		customObject, err := client.CustomObjectCreate(ctx, &draft)
		if err != nil {
			return err
		}

		d.SetId(customObject.ID)
		d.Set("version", customObject.Version)

	}
	return nil
}

func resourceCustomObjectDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}

func _decodeCustomObjectValue(value string) interface{} {
	data := make(map[string]interface{})
	json.Unmarshal([]byte(value), &data)
	return data
}
