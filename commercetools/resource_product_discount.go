package commercetools

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

func resourceProductDiscount() *schema.Resource {
	return &schema.Resource{
		Create: resourceProductDiscountCreate,
		Read:   resourceProductDiscountRead,
		Update: resourceProductDiscountUpdate,
		Delete: resourceProductDiscountDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     TypeLocalizedString,
				Required: true,
			},
			"key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"predicate": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"sortOrder": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"isActive": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  false,
			},
			"validFrom": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"validUntil": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"value": {
				Type:     schema.TypeMap,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateDestinationType,
						},
						// Absolute specific fields
						"money": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"currency_code": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: ValidateCurrencyCode,
									},
									"cent_amount": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
						// Relative specific fields
						"permyriad": {
							Type:     schema.TypeInt,
							Optional: true,
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

func resourceProductDiscountCreate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)

	name := expandLocalizedString(d.Get("name"))
	description := expandLocalizedString(d.Get("description"))
	validFrom, err := expandDate(d.Get("validFrom").(string))
	if err != nil {
		return err
	}
	validUntil, err := expandDate(d.Get("validUntil").(string))
	if err != nil {
		return err
	}

	draft := &commercetools.ProductDiscountDraft{
		Name:        &name,
		Key:         d.Get("key").(string),
		Description: &description,
		Value:       expandProductDiscountValue(d),
		Predicate:   d.Get("predicate").(string),
		SortOrder:   d.Get("sortOrder").(string),
		IsActive:    d.Get("isActive").(bool),
		ValidFrom:   validFrom,
		ValidUntil:  validUntil,
	}

	productDiscount, err := client.ProductDiscountCreate(draft)
	if err != nil {
		return err
	}

	d.SetId(productDiscount.ID)
	d.Set("version", productDiscount.Version)

	return resourceProductDiscountRead(d, m)
}

func expandLocalizedString(value interface{}) commercetools.LocalizedString {
	return commercetools.LocalizedString(
		expandStringMap(value.(map[string]interface{})))
}

func expandProductDiscountValue(d *schema.ResourceData) commercetools.ProductDiscountValue {
	value := d.Get("value").(map[string]interface{})
	switch value["type"].(string) {
	case "external":
		return commercetools.ProductDiscountValueExternal{}
	case "absolute":
		moneyData := value["money"].([]map[string]interface{})
		moneyList := make([]commercetools.Money, 0)
		for _, data := range moneyData {
			currencyCode := data["currency_code"].(string)
			centAmount := data["cent_amount"].(int)
			money := commercetools.Money{
				CurrencyCode: commercetools.CurrencyCode(currencyCode),
				CentAmount:   centAmount,
			}
			moneyList = append(moneyList, money)
		}

		return commercetools.ProductDiscountValueAbsolute{
			Money: moneyList,
		}
	case "relative":
		return commercetools.ProductDiscountValueRelative{
			Permyriad: value["permyriad"].(int),
		}
	default:
		return fmt.Errorf("Unknown product discount type %s", value["type"])
	}
}

func flattenProductDiscountValue(productDiscount commercetools.ProductDiscountValue) (out map[string]interface{}) {
	out = make(map[string]interface{})
	if discount, ok := productDiscount.(commercetools.ProductDiscountValueAbsolute); ok {
		out["type"] = "absolute"
		out["money"] = flattenProductDiscountAbsolute(discount.Money)
		return out
	} else if discount, ok := productDiscount.(commercetools.ProductDiscountValueRelative); ok {
		out["type"] = "relative"
		out["permyriad"] = discount.Permyriad
		return out
	} else if _, ok := productDiscount.(commercetools.ProductDiscountValueExternal); ok {
		out["type"] = "external"
		return out
	}

	return out
}

func flattenProductDiscountAbsolute(money []commercetools.Money) []map[string]interface{} {
	var out = make([]map[string]interface{}, len(money), len(money))
	for _, moneyEntry := range money {
		m := make(map[string]interface{})
		m["currency_code"] = string(moneyEntry.CurrencyCode)
		m["cent_amount"] = moneyEntry.CentAmount
		out = append(out, m)
	}
	return out
}

func resourceProductDiscountRead(d *schema.ResourceData, m interface{}) error {
	log.Print("[DEBUG] Reading product discount from commercetools")
	client := getClient(m)

	productDiscount, err := client.ProductDiscountGetWithID(d.Id())

	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if productDiscount == nil {
		log.Print("[DEBUG] No product type found")
		d.SetId("")
	} else {
		log.Printf("[DEBUG] Found following product discount: %#v", productDiscount)
		log.Print(stringFormatObject(productDiscount))

		d.Set("version", productDiscount.Version)
		d.Set("name", productDiscount.Name)
		d.Set("key", productDiscount.Key)
		d.Set("description", productDiscount.Description)
		if err := d.Set("value", flattenProductDiscountValue(productDiscount)); err != nil {
			return err
		}
		d.Set("predicate", productDiscount.Predicate)
		d.Set("sortOrder", productDiscount.SortOrder)
		d.Set("isActive", productDiscount.IsActive)
		d.Set("validFrom", flattenDateToString(productDiscount.ValidFrom))
		d.Set("validUntil", flattenDateToString(productDiscount.ValidUntil))
	}

	return nil
}

func expandDate(input string) (time.Time, error) {
	return time.Parse("2006-01-02", input)
}

func flattenDateToString(input time.Time) string {
	return input.Format("2006-01-02")
}

func resourceProductDiscountUpdate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)

	input := &commercetools.ProductDiscountUpdateWithIDInput{
		ID:      d.Id(),
		Version: d.Get("version").(int),
		Actions: []commercetools.ProductDiscountUpdateAction{},
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.ProductDiscountSetKeyAction{Key: newKey})
	}

	if d.HasChange("isActive") {
		isActive := d.Get("isActive").(bool)
		input.Actions = append(
			input.Actions,
			&commercetools.ProductDiscountChangeIsActiveAction{IsActive: isActive})
	}

	if d.HasChange("predicate") {
		newPredicate := d.Get("predicate").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.ProductDiscountChangePredicateAction{Predicate: newPredicate})
	}

	if d.HasChange("sortOrder") {
		newSortOrder := d.Get("sortOrder").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.ProductDiscountChangeSortOrderAction{SortOrder: newSortOrder})
	}

	if d.HasChange("validFrom") {
		validFrom, err := expandDate(d.Get("validFrom").(string))
		if err != nil {
			return err
		}
		input.Actions = append(
			input.Actions,
			&commercetools.ProductDiscountSetValidFromAction{ValidFrom: validFrom})
	}

	if d.HasChange("validUntil") {
		validUntil, err := expandDate(d.Get("validUntil").(string))
		if err != nil {
			return err
		}
		input.Actions = append(
			input.Actions,
			&commercetools.ProductDiscountSetValidUntilAction{ValidUntil: validUntil})
	}

	if d.HasChange("name") {
		newName := expandLocalizedString(d.Get("name"))
		input.Actions = append(
			input.Actions,
			&commercetools.ProductDiscountChangeNameAction{Name: &newName})
	}

	if d.HasChange("description") {
		newDescr := expandLocalizedString(d.Get("description"))
		input.Actions = append(
			input.Actions,
			&commercetools.ProductDiscountSetDescriptionAction{Description: &newDescr})
	}

	log.Printf(
		"[DEBUG] Will perform update operation with the following actions:\n%s",
		stringFormatActions(input.Actions))

	_, err := client.ProductDiscountUpdateWithID(input)
	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			log.Printf("[DEBUG] %v: %v", ctErr, stringFormatErrorExtras(ctErr))
		}
		return err
	}

	return resourceProductDiscountRead(d, m)
}

func resourceProductDiscountDelete(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	version := d.Get("version").(int)
	_, err := client.ProductDiscountDeleteWithID(d.Id(), version)
	if err != nil {
		return err
	}

	return nil
}
