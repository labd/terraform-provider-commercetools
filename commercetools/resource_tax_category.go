package commercetools

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

func resourceTaxCategory() *schema.Resource {
	return &schema.Resource{
		Create: resourceTaxCategoryCreate,
		Read:   resourceTaxCategoryRead,
		Update: resourceTaxCategoryUpdate,
		Delete: resourceTaxCategoryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceTaxCategoryValidateAmount(val interface{}, key string) (warns []string, errs []error) {
	v := val.(float64)
	if v < 0 || v > 1 {
		errs = append(errs, fmt.Errorf("%q must be between 0 and 1 inclusive, got: %f", key, v))
	}
	return
}

func resourceTaxCategoryCreate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	var taxCategory *commercetools.TaxCategory
	emptyTaxRates := []commercetools.TaxRateDraft{}

	draft := &commercetools.TaxCategoryDraft{
		Key:         d.Get("key").(string),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Rates:       emptyTaxRates,
	}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error

		taxCategory, err = client.TaxCategoryCreate(draft)
		if err != nil {
			if ctErr, ok := err.(commercetools.ErrorResponse); ok {
				if _, ok := ctErr.Errors[0].(commercetools.InvalidJSONInputError); ok {
					return resource.NonRetryableError(ctErr)
				}
			} else {
				log.Printf("[DEBUG] Received error: %s", err)
			}
			return resource.RetryableError(err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	if taxCategory == nil {
		log.Fatal("No tax category created?")
	}

	d.SetId(taxCategory.ID)
	d.Set("version", taxCategory.Version)

	return resourceTaxCategoryRead(d, m)
}

func resourceTaxCategoryRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] Reading tax category from commercetools, with taxCategory id: %s", d.Id())
	client := getClient(m)

	taxCategory, err := client.TaxCategoryGetByID(d.Id())

	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if taxCategory == nil {
		log.Print("[DEBUG] No tax category found")
		d.SetId("")
	} else {
		log.Print("[DEBUG] Found following tax category:")
		log.Print(stringFormatObject(taxCategory))

		d.Set("version", taxCategory.Version)
		d.Set("key", taxCategory.Key)
		d.Set("name", taxCategory.Name)
		d.Set("description", taxCategory.Description)
	}
	return nil
}

func resourceTaxCategoryUpdate(d *schema.ResourceData, m interface{}) error {
	// Lock to prevent concurrent updates due to Version number conflicts
	ctMutexKV.Lock(d.Id())
	defer ctMutexKV.Unlock(d.Id())

	client := getClient(m)
	taxCategory, err := client.TaxCategoryGetByID(d.Id())
	if err != nil {
		return err
	}

	input := &commercetools.TaxCategoryUpdateInput{
		ID:      d.Id(),
		Version: taxCategory.Version,
		Actions: []commercetools.TaxCategoryUpdateAction{},
	}

	if d.HasChange("name") {
		newName := d.Get("name").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.TaxCategoryChangeNameAction{Name: newName})
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.TaxCategorySetKeyAction{Key: newKey})
	}

	if d.HasChange("description") {
		newDescription := d.Get("description").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.TaxCategorySetDescriptionAction{Description: newDescription})
	}

	log.Printf(
		"[DEBUG] Will perform update operation with the following actions:\n%s",
		stringFormatActions(input.Actions))

	_, err = client.TaxCategoryUpdate(input)
	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			log.Printf("[DEBUG] %v: %v", ctErr, stringFormatErrorExtras(ctErr))
		}
		return err
	}

	return resourceTaxCategoryRead(d, m)
}

func resourceTaxCategoryDelete(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)

	// Lock to prevent concurrent updates due to Version number conflicts
	ctMutexKV.Lock(d.Id())
	defer ctMutexKV.Unlock(d.Id())

	taxCategory, err := client.TaxCategoryGetByID(d.Id())
	if err != nil {
		return err
	}
	_, err = client.TaxCategoryDeleteByID(d.Id(), taxCategory.Version)
	if err != nil {
		return err
	}

	return nil
}
