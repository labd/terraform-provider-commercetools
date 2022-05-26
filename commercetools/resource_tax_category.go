package commercetools

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
)

func resourceTaxCategory() *schema.Resource {
	return &schema.Resource{
		Description: "Tax Categories define how products are to be taxed in different countries.\n\n" +
			"See also the [Tax Category API Documentation](https://docs.commercetools.com/api/projects/taxCategories)",
		CreateContext: resourceTaxCategoryCreate,
		ReadContext:   resourceTaxCategoryRead,
		UpdateContext: resourceTaxCategoryUpdate,
		DeleteContext: resourceTaxCategoryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceTaxCategoryCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)
	emptyTaxRates := []platform.TaxRateDraft{}

	draft := platform.TaxCategoryDraft{
		Name:        d.Get("name").(string),
		Description: stringRef(d.Get("description")),
		Rates:       emptyTaxRates,
	}

	key := stringRef(d.Get("key"))
	if *key != "" {
		draft.Key = key
	}

	var taxCategory *platform.TaxCategory
	err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		var err error

		taxCategory, err = client.TaxCategories().Post(draft).Execute(ctx)
		return processRemoteError(err)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if taxCategory == nil {
		return diag.Errorf("No tax category created?")
	}

	d.SetId(taxCategory.ID)
	d.Set("version", taxCategory.Version)

	return resourceTaxCategoryRead(ctx, d, m)
}

func resourceTaxCategoryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Reading tax category from commercetools, with taxCategory id: %s", d.Id())
	client := getClient(m)

	taxCategory, err := client.TaxCategories().WithId(d.Id()).Get().Execute(ctx)

	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
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

func resourceTaxCategoryUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Lock to prevent concurrent updates due to Version number conflicts
	ctMutexKV.Lock(d.Id())
	defer ctMutexKV.Unlock(d.Id())

	client := getClient(m)
	taxCategory, err := client.TaxCategories().WithId(d.Id()).Get().Execute(ctx)
	if err != nil {
		return diag.FromErr(err)
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

	_, err = client.TaxCategories().WithId(d.Id()).Post(input).Execute(ctx)
	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
			log.Printf("[DEBUG] %v: %v", ctErr, stringFormatErrorExtras(ctErr))
		}
		return diag.FromErr(err)
	}

	return resourceTaxCategoryRead(ctx, d, m)
}

func resourceTaxCategoryDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)

	// Lock to prevent concurrent updates due to Version number conflicts
	ctMutexKV.Lock(d.Id())
	defer ctMutexKV.Unlock(d.Id())

	taxCategory, err := client.TaxCategories().WithId(d.Id()).Get().Execute(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = client.TaxCategories().WithId(d.Id()).Delete().Version(taxCategory.Version).Execute(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
