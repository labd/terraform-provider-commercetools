package commercetools

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

func resourceCustomerGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceCustomerGroupCreate,
		Read:   resourceCustomerGroupRead,
		Update: resourceCustomerGroupUpdate,
		Delete: resourceCustomerGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceCustomerGroupCreate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	var customerGroup *commercetools.CustomerGroup

	draft := &commercetools.CustomerGroupDraft{
		GroupName: d.Get("name").(string),
		Key:       d.Get("key").(string),
	}

	errorResponse := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error

		customerGroup, err = client.CustomerGroupCreate(context.Background(), draft)

		if err != nil {
			return handleCommercetoolsError(err)
		}
		return nil
	})

	if errorResponse != nil {
		return errorResponse
	}

	if customerGroup == nil {
		log.Fatal("No customer group")
	}

	d.SetId(customerGroup.ID)
	d.Set("version", customerGroup.Version)

	return resourceCustomerGroupRead(d, m)
}

func resourceCustomerGroupRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] Reading customer group from commercetools, with customer group id: %s", d.Id())

	client := getClient(m)

	customerGroup, err := client.CustomerGroupGetWithID(context.Background(), d.Id())

	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if customerGroup == nil {
		log.Print("[DEBUG] No customer group found")
		d.SetId("")
	} else {
		log.Print("[DEBUG] Found following customer group:")
		log.Print(stringFormatObject(customerGroup))

		d.Set("version", customerGroup.Version)
		d.Set("name", customerGroup.Name)
		d.Set("key", customerGroup.Key)
	}

	return nil
}

func resourceCustomerGroupUpdate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	customerGroup, err := client.CustomerGroupGetWithID(context.Background(), d.Id())
	if err != nil {
		return err
	}

	input := &commercetools.CustomerGroupUpdateWithIDInput{
		ID:      d.Id(),
		Version: customerGroup.Version,
		Actions: []commercetools.CustomerGroupUpdateAction{},
	}

	if d.HasChange("name") {
		newName := d.Get("name").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.CustomerGroupChangeNameAction{Name: newName})
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.CustomerGroupSetKeyAction{Key: newKey})
	}

	log.Printf(
		"[DEBUG] Will perform update operation with the following actions:\n%s",
		stringFormatActions(input.Actions))

	_, err = client.CustomerGroupUpdateWithID(context.Background(), input)
	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			log.Printf("[DEBUG] %v: %v", ctErr, stringFormatErrorExtras(ctErr))
		}
		return err
	}

	return resourceCustomerGroupRead(d, m)
}

func resourceCustomerGroupDelete(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	version := d.Get("version").(int)
	_, err := client.CustomerGroupDeleteWithID(context.Background(), d.Id(), version)
	if err != nil {
		log.Printf("[ERROR] Error during deleting customer group resource %s", err)
		return nil
	}
	return nil
}
