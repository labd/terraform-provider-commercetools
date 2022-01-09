package commercetools

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
	"log"
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
				Description: "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
				Type:        TypeLocalizedString,
				Required:    true,
			},
			"key": {
				Description: "User-specific unique identifier for a product discount. Must be unique across a project",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"description": {
				Description: "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
				Type:        TypeLocalizedString,
				Optional:    true,
			},
			"predicate": {
				Description: "A valid [Cart Predicate](https://docs.commercetools.com/api/projects/predicates#cart-predicates)",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "1=1",
			},
			"sort_order": {
				Description: "The string must contain a number between 0 and 1. All matching product discounts are " +
					"applied to a cart in the order defined by this field. A discount with greater sort order is " +
					"prioritized higher than a discount with lower sort order. The sort order is unambiguous among all product discounts",
				Type:     schema.TypeString,
				Optional: true,
			},
			"is_active": {
				Description: "Only active discount can be applied to the product",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"valid_from": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"valid_until": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"value": {
				Description: "Defines the effect the discount will have. " +
					"[ProductDiscountValue](https://docs.commercetools.com/api/projects/productDiscounts#productdiscountvalue)",
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Description:  "Currently supports absolute/relative/external",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateProductDiscountType,
						},
						"money": {
							Description: "Absolute discount specific fields",
							Type:        schema.TypeList,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"currency_code": {
										Description:  "The currency code compliant to [ISO 4217](https://en.wikipedia.org/wiki/ISO_4217)",
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: ValidateCurrencyCode,
									},
									"cent_amount": {
										Description: "The amount in cents (the smallest indivisible unit of the currency)",
										Type:        schema.TypeInt,
										Required:    true,
									},
								},
							},
						},
						"permyriad": {
							Description: "Relative discount specific fields",
							Type:        schema.TypeInt,
							Optional:    true,
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

func validateProductDiscountType(val interface{}, key string) (warns []string, errs []error) {
	var v = val.(string)

	switch v {
	case
		"external",
		"relative",
		"absolute":
		return
	default:
		errs = append(errs, fmt.Errorf("%q not a valid value for %q", val, key))
	}
	return
}

func resourceProductDiscountCreate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)

	name := platform.LocalizedString(
		expandStringMap(d.Get("name").(map[string]interface{})))
	description := platform.LocalizedString(
		expandStringMap(d.Get("description").(map[string]interface{})))

	value, err := resourceProductDiscountGetValue(d)
	if err != nil {
		return err
	}

	draft := platform.ProductDiscountDraft{
		Name:        name,
		Key:         stringRef(d.Get("key")),
		Description: &description,
		Predicate:   d.Get("predicate").(string),
		Value:       value,
		SortOrder:   d.Get("sort_order").(string),
		IsActive:    d.Get("is_active").(bool),
	}

	if val := d.Get("valid_from").(string); len(val) > 0 {
		validFrom, err := expandDate(val)
		if err != nil {
			return err
		}
		draft.ValidFrom = &validFrom
	}
	if val := d.Get("valid_until").(string); len(val) > 0 {
		validUntil, err := expandDate(val)
		if err != nil {
			return err
		}
		draft.ValidUntil = &validUntil
	}

	log.Printf("[DEBUG] Going to create draft: %#v", draft)

	productDiscount, err := client.ProductDiscounts().Post(draft).Execute(context.Background())
	if err != nil {
		return err
	}

	d.SetId(productDiscount.ID)
	d.Set("version", productDiscount.Version)

	return resourceProductDiscountRead(d, m)
}

func resourceProductDiscountGetValue(d *schema.ResourceData) (platform.ProductDiscountValue, error) {
	value := d.Get("value").([]interface{})[0].(map[string]interface{})

	log.Printf("[DEBUG] Product discount value: %#v", value)

	switch value["type"].(string) {
	case "external":
		return platform.ProductDiscountValueExternal{}, nil
	case "absolute":
		money := resourceProductDiscountGetMoney(value)
		return platform.ProductDiscountValueAbsolute{
			Money: money,
		}, nil
	case "relative":
		return platform.ProductDiscountValueRelative{
			Permyriad: value["permyriad"].(int),
		}, nil
	default:
		return nil, fmt.Errorf("Value type %s not implemented", value["type"])
	}
}

func resourceProductDiscountGetMoney(d map[string]interface{}) []platform.TypedMoney {
	input := d["money"].([]interface{})
	var result []platform.TypedMoney

	for _, raw := range input {
		i := raw.(map[string]interface{})
		priceCurrencyCode := i["currency_code"].(string)

		result = append(result, platform.Money{
			CurrencyCode: priceCurrencyCode,
			CentAmount:   i["cent_amount"].(int),
		})
	}

	return result
}

func resourceProductDiscountRead(d *schema.ResourceData, m interface{}) error {
	log.Print("[DEBUG] Reading product discount from platform")

	client := getClient(m)

	productDiscount, err := client.ProductDiscounts().WithId(d.Id()).Get().Execute(context.Background())

	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if productDiscount == nil {
		log.Print("[DEBUG] No product discount found")
		d.SetId("")
	} else {
		log.Printf("[DEBUG] Found following product discount: %#v", productDiscount)
		log.Print(stringFormatObject(productDiscount))

		d.Set("version", productDiscount.Version)
		d.Set("name", productDiscount.Name)
		d.Set("key", productDiscount.Key)
		d.Set("description", productDiscount.Description)
		d.Set("value", productDiscount.Value)
		d.Set("predicate", productDiscount.Predicate)
		d.Set("sort_order", productDiscount.SortOrder)
		d.Set("is_active", productDiscount.IsActive)
		d.Set("valid_from", productDiscount.ValidFrom)
		d.Set("valid_until", productDiscount.ValidUntil)
	}

	return nil
}

func resourceProductDiscountUpdate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	productDiscount, err := client.ProductDiscounts().WithId(d.Id()).Get().Execute(context.Background())
	if err != nil {
		return err
	}

	input := platform.ProductDiscountUpdate{
		Version: productDiscount.Version,
		Actions: []platform.ProductDiscountUpdateAction{},
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&platform.ProductDiscountSetKeyAction{Key: &newKey})
	}

	if d.HasChange("name") {
		newName := platform.LocalizedString(
			expandStringMap(d.Get("name").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&platform.ProductDiscountChangeNameAction{Name: newName})
	}

	if d.HasChange("description") {
		newDescription := platform.LocalizedString(
			expandStringMap(d.Get("description").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&platform.ProductDiscountSetDescriptionAction{Description: &newDescription})
	}

	if d.HasChange("value") {
		newValue, err := resourceProductDiscountGetValue(d)
		if err != nil {
			return err
		}
		input.Actions = append(
			input.Actions,
			&platform.ProductDiscountChangeValueAction{Value: newValue})
	}

	if d.HasChange("predicate") {
		newPredicate := d.Get("predicate").(string)
		input.Actions = append(
			input.Actions,
			&platform.ProductDiscountChangePredicateAction{Predicate: newPredicate})
	}

	if d.HasChange("sort_order") {
		newSortOrder := d.Get("sort_order").(string)
		input.Actions = append(
			input.Actions,
			&platform.ProductDiscountChangeSortOrderAction{SortOrder: newSortOrder})
	}

	if d.HasChange("is_active") {
		newIsActive := d.Get("is_active").(bool)
		input.Actions = append(
			input.Actions,
			&platform.ProductDiscountChangeIsActiveAction{IsActive: newIsActive})
	}

	if d.HasChange("valid_from") || d.HasChange("valid_until") {
		var action = &platform.ProductDiscountSetValidFromAndUntilAction{}
		if val := d.Get("valid_from").(string); len(val) > 0 {
			newValidFrom, err := expandDate(d.Get("valid_from").(string))
			if err != nil {
				return err
			}
			action.ValidFrom = &newValidFrom
		}
		if val := d.Get("valid_until").(string); len(val) > 0 {
			newValidUntil, err := expandDate(d.Get("valid_until").(string))
			if err != nil {
				return err
			}
			action.ValidUntil = &newValidUntil
		}

		input.Actions = append(input.Actions, action)
	}

	log.Printf(
		"[DEBUG] Will perform update operation with the following actions:\n%s",
		stringFormatActions(input.Actions))

	_, err = client.ProductDiscounts().WithId(d.Id()).Post(input).Execute(context.Background())
	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
			log.Printf("[DEBUG] %v: %v", ctErr, stringFormatErrorExtras(ctErr))
		}
		return err
	}

	return resourceProductDiscountRead(d, m)
}

func resourceProductDiscountDelete(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	version := d.Get("version").(int)
	_, err := client.ProductDiscounts().WithId(d.Id()).Delete().Version(version).Execute(context.Background())
	if err != nil {
		return err
	}

	return nil
}
