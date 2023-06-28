package commercetools

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func resourceTaxCategoryRate() *schema.Resource {
	return &schema.Resource{
		Description: "Tax rate for Tax Category. \n\n" +
			"See also [Tax Rate API Documentation](https://docs.commercetools.com/api/projects/taxCategories#taxrate)",
		CreateContext: resourceTaxCategoryRateCreate,
		ReadContext:   resourceTaxCategoryRateRead,
		UpdateContext: resourceTaxCategoryRateUpdate,
		DeleteContext: resourceTaxCategoryRateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTaxCategoryRateImportState,
		},
		Schema: map[string]*schema.Schema{
			"tax_category_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"amount": {
				Description: "Number Percentage in the range of [0..1]. The sum of the amounts of all subRates, " +
					"if there are any",
				Type:         schema.TypeFloat,
				Optional:     true,
				ValidateFunc: validateTaxRateAmount,
			},
			"included_in_price": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"country": {
				Description: "A two-digit country code as per [ISO 3166-1 alpha-2](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2)",
				Type:        schema.TypeString,
				Required:    true,
			},
			"state": {
				Description: "The state in the country",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"sub_rate": {
				Description: "For countries (for example the US) where the total tax is a combination of multiple " +
					"taxes (for example state and local taxes)",
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"amount": {
							Description:  "Number Percentage in the range of [0..1]",
							Type:         schema.TypeFloat,
							Required:     true,
							ValidateFunc: validateTaxRateAmount,
						},
					},
				},
			},
		},
	}
}

func resourceTaxCategoryRateImportState(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	client := getClient(meta)
	taxRateID := d.Id()
	// Arbitrary number, safe to assume there won't be more than 500 tax categories...
	taxCategoriesQuery, err := client.TaxCategories().Get().Limit(500).Execute(ctx)
	if err != nil {
		return nil, err
	}

	taxCategory, taxRate := findTaxRate(taxRateID, taxCategoriesQuery.Results)

	if taxRate == nil {
		return nil, fmt.Errorf("tax rate %s does not seem to exist", taxRateID)
	}

	results := make([]*schema.ResourceData, 0)
	taxRateState := resourceTaxCategoryRate().Data(nil)

	taxRateState.SetId(*taxRate.ID)
	taxRateState.SetType("commercetools_tax_category_rate")
	taxRateState.Set("tax_category_id", taxCategory.ID)

	setTaxRateState(taxRateState, taxRate)

	results = append(results, taxRateState)
	return results, nil
}

func resourceTaxCategoryRateGetSubRates(input []any) ([]platform.SubRate, error) {
	result := []platform.SubRate{}

	for _, raw := range input {
		raw := raw.(map[string]any)
		amount := raw["amount"].(float64)
		result = append(result, platform.SubRate{
			Name:   raw["name"].(string),
			Amount: amount,
		})
	}
	return result, nil
}

func resourceTaxCategoryRateCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	taxCategoryID := d.Get("tax_category_id").(string)

	// Lock to prevent concurrent updates due to Version number conflicts
	ctMutexKV.Lock(taxCategoryID)
	defer ctMutexKV.Unlock(taxCategoryID)

	taxCategory, err := client.TaxCategories().WithId(taxCategoryID).Get().Execute(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	oldTaxRateIds := getTaxRateIds(taxCategory)

	input := platform.TaxCategoryUpdate{
		Version: taxCategory.Version,
		Actions: []platform.TaxCategoryUpdateAction{},
	}

	taxRateDraft, err := createTaxRateDraft(d)
	if err != nil {
		return diag.FromErr(err)
	}

	input.Actions = append(input.Actions, platform.TaxCategoryAddTaxRateAction{TaxRate: *taxRateDraft})

	err = resource.RetryContext(ctx, 30*time.Second, func() *resource.RetryError {
		taxCategory, err = client.TaxCategories().WithId(taxCategoryID).Post(input).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	// Refresh the taxCategory. When a tax rate is added the ID is different
	// then the ID returned in the response
	updatedTaxCategory, err := client.TaxCategories().WithId(taxCategoryID).Get().Execute(ctx)
	newTaxRate := findNewTaxRate(updatedTaxCategory, oldTaxRateIds)

	if newTaxRate == nil {
		return diag.Errorf("No tax category rate created?")
	}

	d.SetId(*newTaxRate.ID)
	d.Set("tax_category_id", taxCategory.ID)

	return resourceTaxCategoryRateRead(ctx, d, m)
}

func resourceTaxCategoryRateRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	_, taxRate, err := readResourcesFromStateIDs(ctx, d, m)

	if err != nil {
		d.SetId("")
		return nil
	}

	setTaxRateState(d, taxRate)

	return nil
}

func setTaxRateState(d *schema.ResourceData, taxRate *platform.TaxRate) {
	d.Set("name", taxRate.Name)
	d.Set("amount", taxRate.Amount)
	d.Set("included_in_price", taxRate.IncludedInPrice)
	d.Set("country", taxRate.Country)
	d.Set("state", taxRate.State)

	subRateData := make([]map[string]any, len(taxRate.SubRates))
	for srIndex, subrate := range taxRate.SubRates {
		subRateData[srIndex] = map[string]any{
			"name":   subrate.Name,
			"amount": subrate.Amount,
		}
	}
	d.Set("sub_rate", subRateData)
}

func resourceTaxCategoryRateUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	taxCategoryID := d.Get("tax_category_id").(string)

	// Lock to prevent concurrent updates due to Version number conflicts
	ctMutexKV.Lock(taxCategoryID)
	defer ctMutexKV.Unlock(taxCategoryID)

	taxCategory, _, err := readResourcesFromStateIDs(ctx, d, m)
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	oldTaxRateIds := getTaxRateIds(taxCategory)

	input := platform.TaxCategoryUpdate{
		Version: taxCategory.Version,
		Actions: []platform.TaxCategoryUpdateAction{},
	}

	if d.HasChange("name") || d.HasChange("amount") || d.HasChange("included_in_price") || d.HasChange("country") || d.HasChange("state") || d.HasChange("sub_rate") {
		taxRateDraft, err := createTaxRateDraft(d)
		if err != nil {
			// Workaround invalid state to be written, see
			// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
			d.Partial(true)
			return diag.FromErr(err)
		}
		input.Actions = append(input.Actions, platform.TaxCategoryReplaceTaxRateAction{
			TaxRateId: stringRef(d.Id()),
			TaxRate:   *taxRateDraft,
		})
	}

	client := getClient(m)
	err = resource.RetryContext(ctx, 30*time.Second, func() *resource.RetryError {
		_, err := client.TaxCategories().WithId(taxCategory.ID).Post(input).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	// Refresh the taxCategory. When a tax rate is added the ID is different
	// then the ID returned in the response
	updatedTaxCategory, err := client.TaxCategories().WithId(taxCategoryID).Get().Execute(ctx)
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	newTaxRate := findNewTaxRate(updatedTaxCategory, oldTaxRateIds)
	if newTaxRate == nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.Errorf("No tax category rate created?")
	}

	d.SetId(*newTaxRate.ID)

	return resourceTaxCategoryRateRead(ctx, d, m)
}

func createTaxRateDraft(d *schema.ResourceData) (*platform.TaxRateDraft, error) {
	var subrates []platform.SubRate
	var err error
	if subRateRaw, ok := d.GetOk("sub_rate"); ok {
		subrates, err = resourceTaxCategoryRateGetSubRates(subRateRaw.([]any))
		if err != nil {
			return nil, err
		}
	}

	var countryCode string
	if value, ok := d.Get("country").(string); ok {
		countryCode = value
	}

	amountRaw := d.Get("amount").(float64)

	taxRateDraft := platform.TaxRateDraft{
		Name:            d.Get("name").(string),
		Amount:          &amountRaw,
		IncludedInPrice: d.Get("included_in_price").(bool),
		Country:         countryCode,
		State:           stringRef(d.Get("state")),
		SubRates:        subrates,
	}

	return &taxRateDraft, nil
}

func resourceTaxCategoryRateDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	taxCategoryID := d.Get("tax_category_id").(string)

	// Lock to prevent concurrent updates due to Version number conflicts
	ctMutexKV.Lock(taxCategoryID)
	defer ctMutexKV.Unlock(taxCategoryID)

	taxCategory, taxRate, err := readResourcesFromStateIDs(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	input := platform.TaxCategoryUpdate{
		Version: taxCategory.Version,
		Actions: []platform.TaxCategoryUpdateAction{},
	}

	removeAction := platform.TaxCategoryRemoveTaxRateAction{
		TaxRateId: taxRate.ID,
	}
	input.Actions = append(input.Actions, removeAction)

	client := getClient(m)
	err = resource.RetryContext(ctx, 30*time.Second, func() *resource.RetryError {
		_, err := client.TaxCategories().WithId(taxCategory.ID).Post(input).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	return diag.FromErr(err)
}

func readResourcesFromStateIDs(ctx context.Context, d *schema.ResourceData, m any) (*platform.TaxCategory, *platform.TaxRate, error) {
	client := getClient(m)
	taxCategoryID := d.Get("tax_category_id").(string)
	taxRateID := d.Id()

	taxCategory, err := client.TaxCategories().WithId(taxCategoryID).Get().Execute(ctx)

	if err != nil {
		return nil, nil, err
	}

	taxRate := getTaxRateWithID(taxCategory, taxRateID)
	if taxRate == nil {
		return nil, nil, fmt.Errorf("could not find tax rate %s in tax category %s", taxRateID, taxCategory.ID)
	}
	return taxCategory, taxRate, nil
}

func validateTaxRateAmount(val any, key string) (warns []string, errs []error) {
	v := val.(float64)
	if v < 0 || v > 1 {
		errs = append(errs, fmt.Errorf("%q must be between 0 and 1 inclusive, got: %f", key, v))
	}
	return
}

func getTaxRateIds(taxCategory *platform.TaxCategory) []string {
	taxRateIds := []string{}
	for _, rate := range taxCategory.Rates {
		taxRateIds = append(taxRateIds, *rate.ID)
	}

	return taxRateIds
}

// Find new tax rate by comparing with tax rate ids created just before adding new tax rate
func findNewTaxRate(taxCategory *platform.TaxCategory, oldTaxRateIds []string) *platform.TaxRate {
	for _, taxRate := range taxCategory.Rates {
		if !stringInSlice(*taxRate.ID, oldTaxRateIds) {
			return &taxRate
		}
	}
	return nil
}

func getTaxRateWithID(taxCategory *platform.TaxCategory, taxRateID string) *platform.TaxRate {
	for _, rate := range taxCategory.Rates {
		if *rate.ID == taxRateID {
			return &rate
		}
	}

	return nil
}

func findTaxRate(taxRateID string, taxCategories []platform.TaxCategory) (*platform.TaxCategory, *platform.TaxRate) {
	for _, taxCategory := range taxCategories {
		for _, taxRate := range taxCategory.Rates {
			if *taxRate.ID == taxRateID {
				return &taxCategory, &taxRate
			}
		}
	}
	return nil, nil
}
