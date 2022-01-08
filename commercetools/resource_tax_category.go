package commercetools

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
)

func resourceTaxCategory() *schema.Resource {
	return &schema.Resource{
		Description: "Tax Categories define how products are to be taxed in different countries.\n\n" +
			"See also the [Tax Category API Documentation](https://docs.commercetools.com/api/projects/taxCategories)",
		Create: resourceTaxCategoryCreate,
		Read:   resourceTaxCategoryRead,
		Update: resourceTaxCategoryUpdate,
		Delete: resourceTaxCategoryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "User-specific unique identifier for the category",
				Type:        schema.TypeString,
				Optional:    true,
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

func resourceTaxCategoryCreate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	var taxCategory *platform.TaxCategory
	emptyTaxRates := []platform.TaxRateDraft{}

	draft := platform.TaxCategoryDraft{
		Key:         stringRef(d.Get("key")),
		Name:        d.Get("name").(string),
		Description: stringRef(d.Get("description")),
		Rates:       emptyTaxRates,
	}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error

		taxCategory, err = client.TaxCategories().Post(draft).Execute(context.Background())
		if err != nil {
			return handleCommercetoolsError(err)
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

	taxCategory, err := client.TaxCategories().WithId(d.Id()).Get().Execute(context.Background())

	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
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
	taxCategory, err := client.TaxCategories().WithId(d.Id()).Get().Execute(context.Background())
	if err != nil {
		return err
	}

	input := platform.TaxCategoryUpdate{
		Version: taxCategory.Version,
		Actions: []platform.TaxCategoryUpdateAction{},
	}

	if d.HasChange("name") {
		newName := d.Get("name").(string)
		input.Actions = append(
			input.Actions,
			&platform.TaxCategoryChangeNameAction{Name: newName})
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&platform.TaxCategorySetKeyAction{Key: &newKey})
	}

	if d.HasChange("description") {
		newDescription := d.Get("description").(string)
		input.Actions = append(
			input.Actions,
			&platform.TaxCategorySetDescriptionAction{Description: &newDescription})
	}

	log.Printf(
		"[DEBUG] Will perform update operation with the following actions:\n%s",
		stringFormatActions(input.Actions))

	_, err = client.TaxCategories().WithId(d.Id()).Post(input).Execute(context.Background())
	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
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

	taxCategory, err := client.TaxCategories().WithId(d.Id()).Get().Execute(context.Background())
	if err != nil {
		return err
	}
	_, err = client.TaxCategories().WithId(d.Id()).Delete().WithQueryParams(platform.ByProjectKeyTaxCategoriesByIDRequestMethodDeleteInput{
		Version: taxCategory.Version,
	}).Execute(context.Background())
	if err != nil {
		return err
	}

	return nil
}
