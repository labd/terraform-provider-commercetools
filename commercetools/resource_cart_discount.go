package commercetools

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

func resourceCartDiscount() *schema.Resource {
	return &schema.Resource{
		Create: resourceCartDiscountCreate,
		Read:   resourceCartDiscountRead,
		Update: resourceCartDiscountUpdate,
		Delete: resourceCartDiscountDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     TypeLocalizedString,
				Required: true,
			},
			"description": {
				Type:     TypeLocalizedString,
				Optional: true,
			},
			"value": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateValueType,
						},
						// Relative discount specific fields
						"permyriad": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						// Absolute discount specific fields
						"money": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"currency_code": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: ValidateCurrencyCode,
									},
									"cent_amount": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
						// Gift Line Item discount specific fields
						"product_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"variant": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"supply_channel_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"distribution_channel_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"predicate": {
				Type:     schema.TypeString,
				Required: true,
			},
			"target": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateTargetType,
						},
						// LineItems/CustomLineItems target specific fields
						"predicate": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"sort_order": {
				Type:     schema.TypeString,
				Required: true,
			},
			"is_active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"valid_from": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"valid_until": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"requires_discount_code": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"stacking_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateStackingMode,
				Default:      "Stacking",
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func validateValueType(val interface{}, key string) (warns []string, errs []error) {
	switch val {
	case
		"relative",
		"absolute",
		"giftLineItem":
		return
	default:
		errs = append(errs, fmt.Errorf("%q not a valid value for %q", val, key))
	}
	return
}

func validateTargetType(val interface{}, key string) (warns []string, errs []error) {
	switch val {
	case
		"lineItems",
		"customLineItems",
		"shipping":
		return
	default:
		errs = append(errs, fmt.Errorf("%q not a valid value for %q", val, key))
	}
	return
}

func validateStackingMode(val interface{}, key string) (warns []string, errs []error) {
	switch val {
	case
		"Stacking",
		"StopAfterThisDiscount":
		return
	default:
		errs = append(errs, fmt.Errorf("%q not a valid value for %q", val, key))
	}
	return
}

func resourceCartDiscountCreate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	var cartDiscount *commercetools.CartDiscount

	name := commercetools.LocalizedString(
		expandStringMap(d.Get("name").(map[string]interface{})))
	description := commercetools.LocalizedString(
		expandStringMap(d.Get("description").(map[string]interface{})))

	value, err := resourceCartDiscountGetValue(d)
	if err != nil {
		return err
	}

	stackingMode, err := resourceCartDiscountGetStackingMode(d)
	if err != nil {
		return err
	}

	draft := &commercetools.CartDiscountDraft{
		Key:                  d.Get("key").(string),
		Name:                 &name,
		Description:          &description,
		Value:                &value,
		CartPredicate:        d.Get("predicate").(string),
		SortOrder:            d.Get("sort_order").(string),
		IsActive:             d.Get("is_active").(bool),
		RequiresDiscountCode: d.Get("requires_discount_code").(bool),
		StackingMode:         stackingMode,
	}

	if val := d.Get("target").(map[string]interface{}); len(val) > 0 {
		target, err := resourceCartDiscountGetTarget(d)
		if err != nil {
			return err
		}
		draft.Target = &target
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

	errorResponse := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error

		cartDiscount, err = client.CartDiscountCreate(context.Background(), draft)

		if err != nil {
			return handleCommercetoolsError(err)
		}
		return nil
	})

	if errorResponse != nil {
		return errorResponse
	}

	if cartDiscount == nil {
		log.Fatal("No cart discount created")
	}

	d.SetId(cartDiscount.ID)
	d.Set("version", cartDiscount.Version)

	return resourceCartDiscountRead(d, m)
}

func resourceCartDiscountRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] Reading cart discount from commercetools, with cartDiscount id: %s", d.Id())

	client := getClient(m)

	cartDiscount, err := client.CartDiscountGetWithID(context.Background(), d.Id())

	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if cartDiscount == nil {
		log.Print("[DEBUG] No cart discount found")
		d.SetId("")
	} else {
		log.Print("[DEBUG] Found following cart discount:")
		log.Print(stringFormatObject(cartDiscount))

		d.Set("version", cartDiscount.Version)
		d.Set("key", cartDiscount.Key)
		d.Set("name", cartDiscount.Name)
		d.Set("description", cartDiscount.Description)
		d.Set("value", cartDiscount.Value)
		d.Set("predicate", cartDiscount.CartPredicate)
		d.Set("target", cartDiscount.Target)
		d.Set("sort_order", cartDiscount.SortOrder)
		d.Set("is_active", cartDiscount.IsActive)
		d.Set("valid_from", cartDiscount.ValidFrom)
		d.Set("valid_until", cartDiscount.ValidUntil)
		d.Set("requires_discount_code", cartDiscount.RequiresDiscountCode)
		d.Set("stacking_mode", cartDiscount.StackingMode)
	}

	return nil
}

func resourceCartDiscountUpdate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	cartDiscount, err := client.CartDiscountGetWithID(context.Background(), d.Id())
	if err != nil {
		return err
	}

	input := &commercetools.CartDiscountUpdateWithIDInput{
		ID:      d.Id(),
		Version: cartDiscount.Version,
		Actions: []commercetools.CartDiscountUpdateAction{},
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.CartDiscountSetKeyAction{Key: newKey})
	}

	if d.HasChange("name") {
		newName := commercetools.LocalizedString(
			expandStringMap(d.Get("name").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&commercetools.CartDiscountChangeNameAction{Name: &newName})
	}

	if d.HasChange("description") {
		newDescription := commercetools.LocalizedString(
			expandStringMap(d.Get("description").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&commercetools.CartDiscountSetDescriptionAction{Description: &newDescription})
	}

	if d.HasChange("value") {
		value, err := resourceCartDiscountGetValue(d)
		if err != nil {
			return err
		}
		input.Actions = append(
			input.Actions,
			&commercetools.CartDiscountChangeValueAction{Value: value})
	}

	if d.HasChange("predicate") {
		newPredicate := d.Get("predicate").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.CartDiscountChangeCartPredicateAction{CartPredicate: newPredicate})
	}

	if d.HasChange("target") {
		if val := d.Get("target").(map[string]interface{}); len(val) > 0 {
			target, err := resourceCartDiscountGetTarget(d)
			if err != nil {
				return err
			}
			input.Actions = append(
				input.Actions,
				&commercetools.CartDiscountChangeTargetAction{Target: target})
		} else {
			return errors.New("Cannot change target to empty")
		}

	}

	if d.HasChange("sort_order") {
		newSortOrder := d.Get("sort_order").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.CartDiscountChangeSortOrderAction{SortOrder: newSortOrder})
	}

	if d.HasChange("is_active") {
		newIsActive := d.Get("is_active").(bool)
		input.Actions = append(
			input.Actions,
			&commercetools.CartDiscountChangeIsActiveAction{IsActive: newIsActive})
	}

	if d.HasChange("valid_from") {
		if val := d.Get("valid_from").(string); len(val) > 0 {
			newValidFrom, err := expandDate(d.Get("valid_from").(string))
			if err != nil {
				return err
			}
			input.Actions = append(
				input.Actions,
				&commercetools.CartDiscountSetValidFromAction{ValidFrom: &newValidFrom})
		} else {
			input.Actions = append(
				input.Actions,
				&commercetools.CartDiscountSetValidFromAction{})
		}
	}

	if d.HasChange("valid_until") {
		if val := d.Get("valid_until").(string); len(val) > 0 {
			newValidUntil, err := expandDate(d.Get("valid_until").(string))
			if err != nil {
				return err
			}
			input.Actions = append(
				input.Actions,
				&commercetools.CartDiscountSetValidUntilAction{ValidUntil: &newValidUntil})
		} else {
			input.Actions = append(
				input.Actions,
				&commercetools.CartDiscountSetValidUntilAction{})
		}
	}

	if d.HasChange("requires_discount_code") {
		newRequiresDiscountCode := d.Get("requires_discount_code").(bool)
		input.Actions = append(
			input.Actions,
			&commercetools.CartDiscountChangeRequiresDiscountCodeAction{RequiresDiscountCode: newRequiresDiscountCode})
	}

	if d.HasChange("stacking_mode") {
		newStackingMode, err := resourceCartDiscountGetStackingMode(d)
		if err != nil {
			return err
		}
		input.Actions = append(
			input.Actions,
			&commercetools.CartDiscountChangeStackingModeAction{StackingMode: newStackingMode})
	}

	log.Printf(
		"[DEBUG] Will perform update operation with the following actions:\n%s",
		stringFormatActions(input.Actions))

	_, err = client.CartDiscountUpdateWithID(context.Background(), input)
	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			log.Printf("[DEBUG] %v: %v", ctErr, stringFormatErrorExtras(ctErr))
		}
		return err
	}

	return resourceCartDiscountRead(d, m)
}

func resourceCartDiscountDelete(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	version := d.Get("version").(int)
	_, err := client.CartDiscountDeleteWithID(context.Background(), d.Id(), version)
	if err != nil {
		return err
	}
	return nil
}

func resourceCartDiscountGetValue(d *schema.ResourceData) (commercetools.CartDiscountValueDraft, error) {
	value := d.Get("value").([]interface{})[0].(map[string]interface{})
	switch value["type"].(string) {
	case "relative":
		return commercetools.CartDiscountValueRelativeDraft{
			Permyriad: value["permyriad"].(int),
		}, nil
	case "absolute":
		money := resourceCartDiscountGetMoney(value)
		return commercetools.CartDiscountValueAbsoluteDraft{
			Money: money,
		}, nil
	case "giftLineItem":

		draft := &commercetools.CartDiscountValueGiftLineItemDraft{}

		if val := value["supply_channel_id"].(string); len(val) > 0 {
			draft.SupplyChannel = &commercetools.ChannelReference{ID: val}
		}
		if val := value["product_id"].(string); len(val) > 0 {
			draft.Product = &commercetools.ProductReference{ID: val}
		}
		if val := value["distribution_channel_id"].(string); len(val) > 0 {
			draft.DistributionChannel = &commercetools.ChannelReference{ID: val}
		}

		draft.VariantID = value["variant"].(int)

		return draft, nil

	default:
		return nil, fmt.Errorf("Value type %s not implemented", value["type"])
	}
}

func resourceCartDiscountGetMoney(d map[string]interface{}) []commercetools.Money {
	input := d["money"].([]interface{})
	var result []commercetools.Money

	for _, raw := range input {
		i := raw.(map[string]interface{})
		priceCurrencyCode := commercetools.CurrencyCode(i["currency_code"].(string))

		result = append(result, commercetools.Money{
			CurrencyCode: priceCurrencyCode,
			CentAmount:   i["cent_amount"].(int),
		})
	}

	return result
}

func resourceCartDiscountGetTarget(d *schema.ResourceData) (commercetools.CartDiscountTarget, error) {
	input := d.Get("target").(map[string]interface{})

	switch input["type"].(string) {
	case "lineItems":
		return commercetools.CartDiscountLineItemsTarget{
			Predicate: input["predicate"].(string),
		}, nil
	case "customLineItems":
		return commercetools.CartDiscountCustomLineItemsTarget{
			Predicate: input["predicate"].(string),
		}, nil
	case "shipping":
		return commercetools.CartDiscountShippingCostTarget{}, nil
	default:
		return nil, fmt.Errorf("Target type %s not implemented", input["type"])
	}

}

func resourceCartDiscountGetStackingMode(d *schema.ResourceData) (commercetools.StackingMode, error) {
	switch d.Get("stacking_mode").(string) {
	case "Stacking":
		return commercetools.StackingModeStacking, nil
	case "StopAfterThisDiscount":
		return commercetools.StackingModeStopAfterThisDiscount, nil
	default:
		return "", fmt.Errorf("Stacking mode %s not implemented", d.Get("stacking_mode").(string))
	}
}
