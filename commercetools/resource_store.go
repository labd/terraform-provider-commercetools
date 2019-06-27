package commercetools

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

func resourceStore() *schema.Resource {
	return &schema.Resource{
		Create: resourceStoreCreate,
		Read:   resourceStoreRead,
		Update: resourceStoreUpdate,
		Delete: resourceStoreDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     TypeLocalizedString,
				Optional: true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceStoreCreate(d *schema.ResourceData, m interface{}) error {
	name := commercetools.LocalizedString(
		expandStringMap(d.Get("name").(map[string]interface{})))

	draft := &commercetools.StoreDraft{
		Key:  d.Get("key").(string),
		Name: &name,
	}

	client := getClient(m)
	store, err := client.StoreCreate(draft)
	if err != nil {
		return err
	}

	d.SetId(store.ID)
	d.Set("version", store.Version)
	return resourceStoreRead(d, m)
}

func resourceStoreRead(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	store, err := client.StoreGetByID(d.Id())

	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.SetId(store.ID)
	d.Set("key", store.Key)
	d.Set("name", *store.Name)
	d.Set("version", store.Version)
	return nil
}

func resourceStoreUpdate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)

	input := &commercetools.StoreUpdateInput{
		ID:      d.Id(),
		Version: d.Get("version").(int),
		Actions: []commercetools.StoreUpdateAction{},
	}

	if d.HasChange("name") {
		newName := commercetools.LocalizedString(
			expandStringMap(d.Get("name").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&commercetools.StoreSetNameAction{Name: &newName})
	}

	_, err := client.StoreUpdate(input)
	if err != nil {
		return err
	}

	return resourceStoreRead(d, m)
}

func resourceStoreDelete(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	version := d.Get("version").(int)
	_, err := client.StoreDelete(d.Id(), version)
	if err != nil {
		return err
	}

	return nil
}
