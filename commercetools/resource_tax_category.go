package commercetools

import (
	"fmt"
	"log"
	"reflect"
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
			"rate": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"amount": {
							Type:         schema.TypeFloat,
							Required:     true,
							ValidateFunc: resourceTaxCategoryValidateAmount,
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
										ValidateFunc: resourceTaxCategoryValidateAmount,
									},
								},
							},
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

func resourceTaxCategoryValidateAmount(val interface{}, key string) (warns []string, errs []error) {
	v := val.(float64)
	if v < 0 || v > 1 {
		errs = append(errs, fmt.Errorf("%q must be between 0 and 1 inclusive, got: %f", key, v))
	}
	return
}

func resourceTaxCategoryCreate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	var ctType *commercetools.TaxCategory

	rates, err := resourceTaxCategoryGetRates(d)

	if err != nil {
		return err
	}

	draft := &commercetools.TaxCategoryDraft{
		Key:         d.Get("key").(string),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Rates:       rates,
	}

	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error

		ctType, err = client.TaxCategoryCreate(draft)
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

	if ctType == nil {
		log.Fatal("No tax category created?")
	}

	d.SetId(ctType.ID)
	d.Set("version", ctType.Version)

	return resourceTaxCategoryRead(d, m)
}

func resourceTaxCategoryGetRates(d *schema.ResourceData) ([]commercetools.TaxRateDraft, error) {
	input := d.Get("rate").([]interface{})
	var result []commercetools.TaxRateDraft

	for _, raw := range input {
		fieldDef, err := resourceTaxCategoryGetRate(raw.(map[string]interface{}), true)

		if err != nil {
			return nil, err
		}

		result = append(result, fieldDef.(commercetools.TaxRateDraft))
	}

	return result, nil
}

func resourceTaxCategoryGetRate(input map[string]interface{}, draft bool) (interface{}, error) {
	var subrates []commercetools.SubRate
	var err error
	if subRateRaw, ok := input["sub_rate"]; ok {
		subrates, err = resourceTaxCategoryGetSubRates(subRateRaw.([]interface{}))
		if err != nil {
			return nil, err
		}
	}

	if draft {
		return commercetools.TaxRateDraft{
			Name:            input["name"].(string),
			Amount:          input["amount"].(float64),
			IncludedInPrice: input["included_in_price"].(bool),
			Country:         input["country"].(commercetools.CountryCode),
			State:           input["state"].(string),
			SubRates:        subrates,
		}, nil
	}
	return commercetools.TaxRate{
		ID:              input["id"].(string),
		Name:            input["name"].(string),
		Amount:          input["amount"].(float64),
		IncludedInPrice: input["included_in_price"].(bool),
		Country:         input["country"].(commercetools.CountryCode),
		State:           input["state"].(string),
		SubRates:        subrates,
	}, nil
}

func resourceTaxCategoryGetSubRates(input []interface{}) ([]commercetools.SubRate, error) {
	result := []commercetools.SubRate{}

	for _, raw := range input {
		raw := raw.(map[string]interface{})
		result = append(result, commercetools.SubRate{
			Name:   raw["name"].(string),
			Amount: raw["amount"].(float64),
		})
	}
	return result, nil
}

func resourceTaxCategoryRead(d *schema.ResourceData, m interface{}) error {
	log.Print("[DEBUG] Reading tax category from commercetools")
	client := getClient(m)

	ctType, err := client.TaxCategoryGetByID(d.Id())

	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if ctType == nil {
		log.Print("[DEBUG] No tax category found")
		d.SetId("")
	} else {
		log.Print("[DEBUG] Found following tax category:")
		log.Print(stringFormatObject(ctType))

		taxRates := make([]map[string]interface{}, len(ctType.Rates))
		for i, rate := range ctType.Rates {
			rateData := make(map[string]interface{})

			subRateData := make([]map[string]interface{}, len(rate.SubRates))
			for srIndex, subrate := range rate.SubRates {
				subRateData[srIndex] = map[string]interface{}{
					"name":   subrate.Name,
					"amount": subrate.Amount,
				}
			}

			rateData["id"] = rate.ID
			rateData["name"] = rate.Name
			rateData["amount"] = rate.Amount
			rateData["included_in_price"] = rate.IncludedInPrice
			rateData["country"] = rate.Country
			rateData["state"] = rate.State
			rateData["sub_rate"] = subRateData

			taxRates[i] = rateData
		}

		d.Set("version", ctType.Version)
		d.Set("key", ctType.Key)
		d.Set("name", ctType.Name)
		d.Set("description", ctType.Description)
		d.Set("rate", taxRates)
	}
	return nil
}

func resourceTaxCategoryUpdate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)

	input := &commercetools.TaxCategoryUpdateInput{
		ID:      d.Id(),
		Version: d.Get("version").(int),
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

	if d.HasChange("rate") {
		old, new := d.GetChange("rate")
		rateChangeActions, err := resourceTaxCategoryRateChangeActions(
			old.([]interface{}), new.([]interface{}))
		if err != nil {
			return err
		}
		input.Actions = append(input.Actions, rateChangeActions...)
	}

	log.Printf(
		"[DEBUG] Will perform update operation with the following actions:\n%s",
		stringFormatActions(input.Actions))

	_, err := client.TaxCategoryUpdate(input)
	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			log.Printf("[DEBUG] %v: %v", ctErr, stringFormatErrorExtras(ctErr))
		}
		return err
	}

	return resourceTaxCategoryRead(d, m)
}

func resourceTaxCategoryRateChangeActions(oldValues []interface{}, newValues []interface{}) ([]commercetools.TaxCategoryUpdateAction, error) {
	oldLookup := createLookup(oldValues, "name")
	newLookup := createLookup(newValues, "name")
	actions := []commercetools.TaxCategoryUpdateAction{}

	for name, value := range oldLookup {
		if _, ok := newLookup[name]; !ok {
			oldV := value.(map[string]interface{})
			log.Printf("[DEBUG] Tax Rate deleted: %s", name)
			id := oldV["id"].(string)
			actions = append(actions, commercetools.TaxCategoryRemoveTaxRateAction{TaxRateID: id})
		}
	}

	for name, value := range newLookup {
		oldValue, existingField := oldLookup[name]
		newV := value.(map[string]interface{})

		var taxRateDraft commercetools.TaxRateDraft
		if output, err := resourceTaxCategoryGetRate(newV, true); err == nil {
			taxRateDraft = output.(commercetools.TaxRateDraft)
		} else {
			return nil, err
		}

		if !existingField {
			log.Printf("[DEBUG] Tax rate added: %s", name)
			actions = append(
				actions,
				commercetools.TaxCategoryAddTaxRateAction{TaxRate: &taxRateDraft})
			continue
		}

		if !reflect.DeepEqual(oldValue, newV) {
			actions = append(
				actions,
				commercetools.TaxCategoryReplaceTaxRateAction{
					TaxRateID: newV["id"].(string),
					TaxRate:   &taxRateDraft,
				})
		}
	}

	return actions, nil
}

func resourceTaxCategoryDelete(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	version := d.Get("version").(int)
	_, err := client.TaxCategoryDeleteByID(d.Id(), version)
	if err != nil {
		return err
	}

	return nil
}
