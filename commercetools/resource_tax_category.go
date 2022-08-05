package commercetools

import (
	"context"
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

	d.SetId(taxCategory.ID)
	d.Set("version", taxCategory.Version)

	return resourceTaxCategoryRead(ctx, d, m)
}

func resourceTaxCategoryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)
	taxCategory, err := client.TaxCategories().WithId(d.Id()).Get().Execute(ctx)
	if err != nil {
		if IsResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if taxCategory == nil {
		d.SetId("")
	} else {
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
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
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

	err = resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		_, err := client.TaxCategories().WithId(d.Id()).Post(input).Execute(ctx)
		return processRemoteError(err)
	})
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
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
	err = resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		_, err = client.TaxCategories().WithId(d.Id()).Delete().Version(taxCategory.Version).Execute(ctx)
		return processRemoteError(err)
	})
	return diag.FromErr(err)
}
