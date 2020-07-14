package commercetools

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

func resourceTaxCategoryRate() *schema.Resource {
	return &schema.Resource{
		Create: resourceTaxCategoryRateCreate,
		Read:   resourceTaxCategoryRateRead,
		Update: resourceTaxCategoryRateUpdate,
		Delete: resourceTaxCategoryRateDelete,
		Importer: &schema.ResourceImporter{
			State: resourceTaxCategoryRateImportState,
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
				Type:         schema.TypeFloat,
				Optional:     true,
				ValidateFunc: validateTaxRateAmount,
			},
			"included_in_price": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"country": {
				Type:     schema.TypeString,
				Required: true,
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"sub_rate": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"amount": {
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

func resourceTaxCategoryRateImportState(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := getClient(meta)
	taxRateID := d.Id()
	// Arbitrary number, safe to assume there won't be more than 500 tax categories...
	queryInput := commercetools.QueryInput{Limit: 500}
	taxCategoriesQuery, err := client.TaxCategoryQuery(context.Background(), &queryInput)
	if err != nil {
		return nil, err
	}

	taxCategory, taxRate := findTaxRate(taxRateID, taxCategoriesQuery.Results)

	if taxRate == nil {
		return nil, fmt.Errorf("Tax rate %s does not seem to exist", taxRateID)
	}

	results := make([]*schema.ResourceData, 0)
	taxRateState := resourceTaxCategoryRate().Data(nil)

	taxRateState.SetId(taxRate.ID)
	taxRateState.SetType("commercetools_tax_category_rate")
	taxRateState.Set("tax_category_id", taxCategory.ID)

	setTaxRateState(taxRateState, taxRate)

	results = append(results, taxRateState)

	log.Printf("[DEBUG] Importing results: %#v", results)

	return results, nil
}

func resourceTaxCategoryRateGetSubRates(input []interface{}) ([]commercetools.SubRate, error) {
	result := []commercetools.SubRate{}

	for _, raw := range input {
		raw := raw.(map[string]interface{})
		amount := raw["amount"].(float64)
		result = append(result, commercetools.SubRate{
			Name:   raw["name"].(string),
			Amount: &amount,
		})
	}
	return result, nil
}

func resourceTaxCategoryRateCreate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	taxCategoryID := d.Get("tax_category_id").(string)

	// Lock to prevent concurrent updates due to Version number conflicts
	ctMutexKV.Lock(taxCategoryID)
	defer ctMutexKV.Unlock(taxCategoryID)

	taxCategory, err := client.TaxCategoryGetWithID(context.Background(), taxCategoryID)

	if err != nil {
		return err
	}

	oldTaxRateIds := getTaxRateIds(taxCategory)

	input := &commercetools.TaxCategoryUpdateWithIDInput{
		ID:      taxCategoryID,
		Version: taxCategory.Version,
		Actions: []commercetools.TaxCategoryUpdateAction{},
	}

	taxRateDraft, err := createTaxRateDraft(d)
	if err != nil {
		return err
	}

	input.Actions = append(input.Actions, commercetools.TaxCategoryAddTaxRateAction{TaxRate: taxRateDraft})

	err = resource.Retry(30*time.Second, func() *resource.RetryError {
		taxCategory, err = client.TaxCategoryUpdateWithID(context.Background(), input)
		if err != nil {
			return handleCommercetoolsError(err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	newTaxRate := findNewTaxRate(taxCategory, oldTaxRateIds)

	if newTaxRate == nil {
		log.Fatal("No tax category rate created?")
	}

	d.SetId(newTaxRate.ID)
	d.Set("tax_category_id", taxCategory.ID)

	return resourceTaxCategoryRateRead(d, m)
}

func resourceTaxCategoryRateRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] Current tax rate state: %s and m: %s", stringFormatObject(d), stringFormatObject(m))
	_, taxRate, err := readResourcesFromStateIDs(d, m)

	if err != nil {
		d.SetId("")
		return err
	}

	setTaxRateState(d, taxRate)

	return nil
}

func setTaxRateState(d *schema.ResourceData, taxRate *commercetools.TaxRate) {
	log.Printf("[DEBUG] Setting state: %s to taxRate: %s", stringFormatObject(d), stringFormatObject(taxRate))
	d.Set("name", taxRate.Name)
	d.Set("amount", taxRate.Amount)
	d.Set("included_in_price", taxRate.IncludedInPrice)
	d.Set("country", taxRate.Country)
	d.Set("state", taxRate.State)

	subRateData := make([]map[string]interface{}, len(taxRate.SubRates))
	for srIndex, subrate := range taxRate.SubRates {
		subRateData[srIndex] = map[string]interface{}{
			"name":   subrate.Name,
			"amount": subrate.Amount,
		}
	}
	d.Set("sub_rate", subRateData)

	log.Printf("[DEBUG] Updated state to: %s", stringFormatObject(d))
}

func resourceTaxCategoryRateUpdate(d *schema.ResourceData, m interface{}) error {
	taxCategoryID := d.Get("tax_category_id").(string)

	// Lock to prevent concurrent updates due to Version number conflicts
	ctMutexKV.Lock(taxCategoryID)
	defer ctMutexKV.Unlock(taxCategoryID)

	taxCategory, _, err := readResourcesFromStateIDs(d, m)
	if err != nil {
		return err
	}

	oldTaxRateIds := getTaxRateIds(taxCategory)

	input := &commercetools.TaxCategoryUpdateWithIDInput{
		ID:      taxCategory.ID,
		Version: taxCategory.Version,
		Actions: []commercetools.TaxCategoryUpdateAction{},
	}

	if d.HasChange("name") || d.HasChange("amount") || d.HasChange("included_in_price") || d.HasChange("country") || d.HasChange("state") || d.HasChange("sub_rate") {
		taxRateDraft, err := createTaxRateDraft(d)
		if err != nil {
			return err
		}
		input.Actions = append(input.Actions, commercetools.TaxCategoryReplaceTaxRateAction{
			TaxRateID: d.Id(),
			TaxRate:   taxRateDraft,
		})
	}

	log.Printf(
		"[DEBUG] Will perform update operation with the following actions:\n%s",
		stringFormatActions(input.Actions))

	client := getClient(m)
	taxCategory, err = client.TaxCategoryUpdateWithID(context.Background(), input)
	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			log.Printf("[DEBUG] %v: %v", ctErr, stringFormatErrorExtras(ctErr))
		}
		return err
	}

	newTaxRate := findNewTaxRate(taxCategory, oldTaxRateIds)

	if newTaxRate == nil {
		log.Fatal("No tax category rate created?")
	}

	d.SetId(newTaxRate.ID)

	return resourceTaxCategoryRateRead(d, m)
}

func createTaxRateDraft(d *schema.ResourceData) (*commercetools.TaxRateDraft, error) {
	var subrates []commercetools.SubRate
	var err error
	if subRateRaw, ok := d.GetOk("sub_rate"); ok {
		subrates, err = resourceTaxCategoryRateGetSubRates(subRateRaw.([]interface{}))
		if err != nil {
			return nil, err
		}
	}

	var countryCode commercetools.CountryCode
	if value, ok := d.Get("country").(string); ok {
		countryCode = commercetools.CountryCode(value)
	}

	amountRaw := d.Get("amount").(float64)

	log.Printf("[DEBUG] Got amount: %f", amountRaw)

	taxRateDraft := commercetools.TaxRateDraft{
		Name:            d.Get("name").(string),
		Amount:          &amountRaw,
		IncludedInPrice: d.Get("included_in_price").(bool),
		Country:         countryCode,
		State:           d.Get("state").(string),
		SubRates:        subrates,
	}

	log.Printf("[DEBUG] Created tax rate draft: %#v from input %#v", taxRateDraft, d)

	return &taxRateDraft, nil
}

func resourceTaxCategoryRateDelete(d *schema.ResourceData, m interface{}) error {
	taxCategoryID := d.Get("tax_category_id").(string)

	// Lock to prevent concurrent updates due to Version number conflicts
	ctMutexKV.Lock(taxCategoryID)
	defer ctMutexKV.Unlock(taxCategoryID)

	taxCategory, taxRate, err := readResourcesFromStateIDs(d, m)
	if err != nil {
		return err
	}

	input := &commercetools.TaxCategoryUpdateWithIDInput{
		ID:      taxCategory.ID,
		Version: taxCategory.Version,
		Actions: []commercetools.TaxCategoryUpdateAction{},
	}

	removeAction := commercetools.TaxCategoryRemoveTaxRateAction{
		TaxRateID: taxRate.ID,
	}
	input.Actions = append(input.Actions, removeAction)

	client := getClient(m)
	_, err = client.TaxCategoryUpdateWithID(context.Background(), input)
	if err != nil {
		return err
	}

	return nil
}

func readResourcesFromStateIDs(d *schema.ResourceData, m interface{}) (*commercetools.TaxCategory, *commercetools.TaxRate, error) {
	client := getClient(m)
	taxCategoryID := d.Get("tax_category_id").(string)
	taxRateID := d.Id()

	log.Printf("[DEBUG] Reading tax category from commercetools, taxCategory ID: %s, taxRate ID: %s", taxCategoryID, taxRateID)

	taxCategory, err := client.TaxCategoryGetWithID(context.Background(), taxCategoryID)

	if err != nil {
		return nil, nil, err
	}

	log.Print("[DEBUG] Found following tax category:")
	log.Print(stringFormatObject(taxCategory))
	taxRate := getTaxRateWithID(taxCategory, taxRateID)
	if taxRate == nil {
		return nil, nil, fmt.Errorf("Could not find tax rate %s in tax category %s", taxRateID, taxCategory.ID)
	}
	log.Print("[DEBUG] Found following tax rate:")
	log.Print(stringFormatObject(taxRate))

	return taxCategory, taxRate, nil
}

func validateTaxRateAmount(val interface{}, key string) (warns []string, errs []error) {
	v := val.(float64)
	if v < 0 || v > 1 {
		errs = append(errs, fmt.Errorf("%q must be between 0 and 1 inclusive, got: %f", key, v))
	}
	return
}

func getTaxRateIds(taxCategory *commercetools.TaxCategory) []string {
	taxRateIds := []string{}
	for _, rate := range taxCategory.Rates {
		taxRateIds = append(taxRateIds, rate.ID)
	}

	return taxRateIds
}

// Find new tax rate by comparing with tax rate ids created just before adding new tax rate
func findNewTaxRate(taxCategory *commercetools.TaxCategory, oldTaxRateIds []string) *commercetools.TaxRate {
	for _, taxRate := range taxCategory.Rates {
		if !stringInSlice(taxRate.ID, oldTaxRateIds) {
			return &taxRate
		}
	}
	return nil
}

func getTaxRateWithID(taxCategory *commercetools.TaxCategory, taxRateID string) *commercetools.TaxRate {
	for _, rate := range taxCategory.Rates {
		if rate.ID == taxRateID {
			return &rate
		}
	}

	return nil
}

func findTaxRate(taxRateID string, taxCategories []commercetools.TaxCategory) (*commercetools.TaxCategory, *commercetools.TaxRate) {
	for _, taxCategory := range taxCategories {
		for _, taxRate := range taxCategory.Rates {
			if taxRate.ID == taxRateID {
				return &taxCategory, &taxRate
			}
		}
	}
	return nil, nil
}
