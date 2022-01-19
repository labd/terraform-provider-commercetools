package commercetools

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"log"
// 	"time"

// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
// 	"github.com/labd/commercetools-go-sdk/ctutils"
// 	"github.com/labd/commercetools-go-sdk/platform"
// )

// func resourceCartDiscount() *schema.Resource {
// 	return &schema.Resource{
// 		Description: "Cart discounts are used to change the prices of different elements within a cart.\n\n" +
// 			"See also the [Cart Discount API Documentation](https://docs.commercetools.com/api/projects/cartDiscounts)",
// 		Create: resourceCartDiscountCreate,
// 		Read:   resourceCartDiscountRead,
// 		Update: resourceCartDiscountUpdate,
// 		Delete: resourceCartDiscountDelete,
// 		Importer: &schema.ResourceImporter{
// 			State: schema.ImportStatePassthrough,
// 		},
// 		SchemaVersion: 1,
// 		StateUpgraders: []schema.StateUpgrader{
// 			{
// 				Type:    resourceCartDiscountResourceV0().CoreConfigSchema().ImpliedType(),
// 				Upgrade: migrateCartDiscountStateV0toV1,
// 				Version: 0,
// 			},
// 		},
// 		Schema: map[string]*schema.Schema{
// 			"key": {
// 				Description: "User-specific unique identifier for a cart discount. Must be unique across a project",
// 				Type:        types.StringType,
// 				Optional:    true,
// 			},
// 			"name": {
// 				Description: "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
// 				Type:        TypeLocalizedString,
// 				Required:    true,
// 			},
// 			"description": {
// 				Description: "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
// 				Type:        TypeLocalizedString,
// 				Optional:    true,
// 			},
// 			"value": {
// 				Description: "Defines the effect the discount will have. " +
// 					"[CartDiscountValue](https://docs.commercetools.com/api/projects/cartDiscounts#cartdiscountvalue)",
// 				Type:     types.ListType{ElemType: types.StringType},
// 				MaxItems: 1,
// 				Required: true,
// 				Elem: &schema.Resource{
// 					Schema: map[string]*schema.Schema{
// 						"type": {
// 							Description:  "Currently supports absolute/relative/giftLineItem",
// 							Type:         types.StringType,
// 							Required:     true,
// 							ValidateFunc: validateValueType,
// 						},
// 						"permyriad": {
// 							Description: "Relative discount specific fields",
// 							Type:        types.Int64Type,
// 							Optional:    true,
// 						},
// 						"money": {
// 							Description: "Absolute discount specific fields",
// 							Type:        types.ListType{ElemType: types.StringType},
// 							Optional:    true,
// 							Elem: &schema.Resource{
// 								Schema: map[string]*schema.Schema{
// 									"currency_code": {
// 										Description:  "The currency code compliant to [ISO 4217](https://en.wikipedia.org/wiki/ISO_4217)",
// 										Type:         types.StringType,
// 										Required:     true,
// 										ValidateFunc: ValidateCurrencyCode,
// 									},
// 									"cent_amount": {
// 										Description: "The amount in cents (the smallest indivisible unit of the currency)",
// 										Type:        types.Int64Type,
// 										Required:    true,
// 									},
// 								},
// 							},
// 						},
// 						"product_id": {
// 							Description: "Gift Line Item discount specific field",
// 							Type:        types.StringType,
// 							Optional:    true,
// 						},
// 						"variant": {
// 							Description: "Gift Line Item discount specific field",
// 							Type:        types.Int64Type,
// 							Optional:    true,
// 						},
// 						"supply_channel_id": {
// 							Description: "Gift Line Item discount specific field",
// 							Type:        types.StringType,
// 							Optional:    true,
// 						},
// 						"distribution_channel_id": {
// 							Description: "Gift Line Item discount specific field",
// 							Type:        types.StringType,
// 							Optional:    true,
// 						},
// 					},
// 				},
// 			},
// 			"predicate": {
// 				Description: "A valid [Cart Predicate](https://docs.commercetools.com/api/projects/predicates#cart-predicates)",
// 				Type:        types.StringType,
// 				Required:    true,
// 			},
// 			"target": {
// 				Description: "Empty when the value has type giftLineItem, otherwise a " +
// 					"[CartDiscountTarget](https://docs.commercetools.com/api/projects/cartDiscounts#cartdiscounttarget)",
// 				Type:     types.ListType{ElemType: types.StringType},
// 				MaxItems: 1,
// 				Optional: true,
// 				Elem: &schema.Resource{
// 					Schema: map[string]*schema.Schema{
// 						"type": {
// 							Description:  "Supports lineItems/customLineItems/shipping",
// 							Type:         types.StringType,
// 							Required:     true,
// 							ValidateFunc: validateTargetType,
// 						},
// 						"predicate": {
// 							Description: "LineItems/CustomLineItems target specific fields",
// 							Type:        types.StringType,
// 							Optional:    true,
// 						},
// 					},
// 				},
// 			},
// 			"sort_order": {
// 				Description: "The string must contain a number between 0 and 1. All matching cart discounts are " +
// 					"applied to a cart in the order defined by this field. A discount with greater sort order is " +
// 					"prioritized higher than a discount with lower sort order. The sort order is unambiguous among all cart discounts",
// 				Type:     types.StringType,
// 				Required: true,
// 			},
// 			"is_active": {
// 				Description: "Only active discount can be applied to the cart",
// 				Type:        types.BoolType,
// 				Optional:    true,
// 				Default:     true,
// 			},
// 			"valid_from": {
// 				Type:     types.StringType,
// 				Optional: true,
// 			},
// 			"valid_until": {
// 				Type:     types.StringType,
// 				Optional: true,
// 			},
// 			"requires_discount_code": {
// 				Description: "States whether the discount can only be used in a connection with a " +
// 					"[DiscountCode](https://docs.commercetools.com/api/projects/discountCodes#discountcode)",
// 				Type:     types.BoolType,
// 				Optional: true,
// 				Default:  false,
// 			},
// 			"stacking_mode": {
// 				Description:  "Specifies whether the application of this discount causes the following discounts to be ignored",
// 				Type:         types.StringType,
// 				Optional:     true,
// 				ValidateFunc: validateStackingMode,
// 				Default:      "Stacking",
// 			},
// 			"version": {
// 				Type:     types.Int64Type,
// 				Computed: true,
// 			},
// 		},
// 	}
// }

// func validateValueType(val interface{}, key string) (warns []string, errs []error) {
// 	switch val {
// 	case
// 		"relative",
// 		"absolute",
// 		"giftLineItem":
// 		return
// 	default:
// 		errs = append(errs, fmt.Errorf("%q not a valid value for %q", val, key))
// 	}
// 	return
// }

// func validateTargetType(val interface{}, key string) (warns []string, errs []error) {
// 	switch val {
// 	case
// 		"lineItems",
// 		"customLineItems",
// 		"shipping":
// 		return
// 	default:
// 		errs = append(errs, fmt.Errorf("%q not a valid value for %q", val, key))
// 	}
// 	return
// }

// func validateStackingMode(val interface{}, key string) (warns []string, errs []error) {
// 	switch val {
// 	case
// 		"Stacking",
// 		"StopAfterThisDiscount":
// 		return
// 	default:
// 		errs = append(errs, fmt.Errorf("%q not a valid value for %q", val, key))
// 	}
// 	return
// }

// func resourceCartDiscountCreate(d *schema.ResourceData, m interface{}) error {
// 	client := getClient(m)
// 	var cartDiscount *platform.CartDiscount

// 	name := platform.LocalizedString(
// 		expandStringMap(d.Get("name").(map[string]interface{})))
// 	description := platform.LocalizedString(
// 		expandStringMap(d.Get("description").(map[string]interface{})))

// 	value, err := unmarshallCartDiscountValue(d)
// 	if err != nil {
// 		return err
// 	}

// 	stackingMode, err := unmarshallCartDiscountStackingMode(d)
// 	if err != nil {
// 		return err
// 	}

// 	draft := platform.CartDiscountDraft{
// 		Key:                  stringRef(d.Get("key")),
// 		Name:                 name,
// 		Description:          &description,
// 		Value:                &value,
// 		CartPredicate:        d.Get("predicate").(string),
// 		SortOrder:            d.Get("sort_order").(string),
// 		IsActive:             boolRef(d.Get("is_active")),
// 		RequiresDiscountCode: ctutils.BoolRef(d.Get("requires_discount_code").(bool)),
// 		StackingMode:         &stackingMode,
// 	}

// 	if val, err := unmarshallCartDiscountTarget(d); err == nil {
// 		draft.Target = val
// 	} else {
// 		return err
// 	}

// 	if val := d.Get("valid_from").(string); len(val) > 0 {
// 		validFrom, err := expandDate(val)
// 		if err != nil {
// 			return err
// 		}
// 		draft.ValidFrom = &validFrom
// 	}
// 	if val := d.Get("valid_until").(string); len(val) > 0 {
// 		validUntil, err := expandDate(val)
// 		if err != nil {
// 			return err
// 		}
// 		draft.ValidUntil = &validUntil
// 	}

// 	errorResponse := resource.Retry(1*time.Minute, func() *resource.RetryError {
// 		var err error

// 		cartDiscount, err = client.CartDiscounts().Post(draft).Execute(context.Background())

// 		if err != nil {
// 			return handleCommercetoolsError(err)
// 		}
// 		return nil
// 	})

// 	if errorResponse != nil {
// 		return errorResponse
// 	}

// 	if cartDiscount == nil {
// 		log.Fatal("No cart discount created")
// 	}

// 	d.SetId(cartDiscount.ID)
// 	d.Set("version", cartDiscount.Version)

// 	return resourceCartDiscountRead(d, m)
// }

// func resourceCartDiscountRead(d *schema.ResourceData, m interface{}) error {
// 	log.Printf("[DEBUG] Reading cart discount from commercetools, with cartDiscount id: %s", d.Id())

// 	client := getClient(m)

// 	cartDiscount, err := client.CartDiscounts().WithId(d.Id()).Get().Execute(context.Background())

// 	if err != nil {
// 		if ctErr, ok := err.(platform.ErrorResponse); ok {
// 			if ctErr.StatusCode == 404 {
// 				d.SetId("")
// 				return nil
// 			}
// 		}
// 		return err
// 	}

// 	if cartDiscount == nil {
// 		log.Print("[DEBUG] No cart discount found")
// 		d.SetId("")
// 	} else {
// 		log.Print("[DEBUG] Found following cart discount:")
// 		log.Print(stringFormatObject(cartDiscount))

// 		d.Set("version", cartDiscount.Version)
// 		d.Set("key", cartDiscount.Key)
// 		d.Set("name", cartDiscount.Name)
// 		d.Set("description", cartDiscount.Description)
// 		d.Set("value", marshallCartDiscountValue(cartDiscount.Value))
// 		d.Set("predicate", cartDiscount.CartPredicate)
// 		d.Set("target", marshallCartDiscountTarget(cartDiscount.Target))
// 		d.Set("sort_order", cartDiscount.SortOrder)
// 		d.Set("is_active", cartDiscount.IsActive)
// 		d.Set("valid_from", marshallTime(cartDiscount.ValidFrom))
// 		d.Set("valid_until", marshallTime(cartDiscount.ValidUntil))
// 		d.Set("requires_discount_code", cartDiscount.RequiresDiscountCode)
// 		d.Set("stacking_mode", cartDiscount.StackingMode)
// 	}

// 	return nil
// }

// func resourceCartDiscountUpdate(d *schema.ResourceData, m interface{}) error {
// 	client := getClient(m)
// 	cartDiscount, err := client.CartDiscounts().WithId(d.Id()).Get().Execute(context.Background())
// 	if err != nil {
// 		return err
// 	}

// 	input := platform.CartDiscountUpdate{
// 		Version: cartDiscount.Version,
// 		Actions: []platform.CartDiscountUpdateAction{},
// 	}

// 	if d.HasChange("key") {
// 		newKey := d.Get("key").(string)
// 		input.Actions = append(
// 			input.Actions,
// 			&platform.CartDiscountSetKeyAction{Key: &newKey})
// 	}

// 	if d.HasChange("name") {
// 		newName := platform.LocalizedString(
// 			expandStringMap(d.Get("name").(map[string]interface{})))
// 		input.Actions = append(
// 			input.Actions,
// 			&platform.CartDiscountChangeNameAction{Name: newName})
// 	}

// 	if d.HasChange("description") {
// 		newDescription := platform.LocalizedString(
// 			expandStringMap(d.Get("description").(map[string]interface{})))
// 		input.Actions = append(
// 			input.Actions,
// 			&platform.CartDiscountSetDescriptionAction{Description: &newDescription})
// 	}

// 	if d.HasChange("value") {
// 		value, err := unmarshallCartDiscountValue(d)
// 		if err != nil {
// 			return err
// 		}
// 		input.Actions = append(
// 			input.Actions,
// 			&platform.CartDiscountChangeValueAction{Value: value})
// 	}

// 	if d.HasChange("predicate") {
// 		newPredicate := d.Get("predicate").(string)
// 		input.Actions = append(
// 			input.Actions,
// 			&platform.CartDiscountChangeCartPredicateAction{CartPredicate: newPredicate})
// 	}

// 	if d.HasChange("target") {
// 		if val, err := unmarshallCartDiscountTarget(d); err == nil {
// 			if val != nil {
// 				input.Actions = append(
// 					input.Actions,
// 					&platform.CartDiscountChangeTargetAction{Target: val})
// 			} else {
// 				return errors.New("Cannot change target to empty")
// 			}
// 		} else {
// 			return err
// 		}

// 	}

// 	if d.HasChange("sort_order") {
// 		newSortOrder := d.Get("sort_order").(string)
// 		input.Actions = append(
// 			input.Actions,
// 			&platform.CartDiscountChangeSortOrderAction{SortOrder: newSortOrder})
// 	}

// 	if d.HasChange("is_active") {
// 		newIsActive := d.Get("is_active").(bool)
// 		input.Actions = append(
// 			input.Actions,
// 			&platform.CartDiscountChangeIsActiveAction{IsActive: newIsActive})
// 	}

// 	if d.HasChange("valid_from") {
// 		if val := d.Get("valid_from").(string); len(val) > 0 {
// 			newValidFrom, err := expandDate(d.Get("valid_from").(string))
// 			if err != nil {
// 				return err
// 			}
// 			input.Actions = append(
// 				input.Actions,
// 				&platform.CartDiscountSetValidFromAction{ValidFrom: &newValidFrom})
// 		} else {
// 			input.Actions = append(
// 				input.Actions,
// 				&platform.CartDiscountSetValidFromAction{})
// 		}
// 	}

// 	if d.HasChange("valid_until") {
// 		if val := d.Get("valid_until").(string); len(val) > 0 {
// 			newValidUntil, err := expandDate(d.Get("valid_until").(string))
// 			if err != nil {
// 				return err
// 			}
// 			input.Actions = append(
// 				input.Actions,
// 				&platform.CartDiscountSetValidUntilAction{ValidUntil: &newValidUntil})
// 		} else {
// 			input.Actions = append(
// 				input.Actions,
// 				&platform.CartDiscountSetValidUntilAction{})
// 		}
// 	}

// 	if d.HasChange("requires_discount_code") {
// 		newRequiresDiscountCode := d.Get("requires_discount_code").(bool)
// 		input.Actions = append(
// 			input.Actions,
// 			&platform.CartDiscountChangeRequiresDiscountCodeAction{RequiresDiscountCode: newRequiresDiscountCode})
// 	}

// 	if d.HasChange("stacking_mode") {
// 		newStackingMode, err := unmarshallCartDiscountStackingMode(d)
// 		if err != nil {
// 			return err
// 		}
// 		input.Actions = append(
// 			input.Actions,
// 			&platform.CartDiscountChangeStackingModeAction{StackingMode: newStackingMode})
// 	}

// 	log.Printf(
// 		"[DEBUG] Will perform update operation with the following actions:\n%s",
// 		stringFormatActions(input.Actions))

// 	_, err = client.CartDiscounts().WithId(d.Id()).Post(input).Execute(context.Background())
// 	if err != nil {
// 		if ctErr, ok := err.(platform.ErrorResponse); ok {
// 			log.Printf("[DEBUG] %v: %v", ctErr, stringFormatErrorExtras(ctErr))
// 		}
// 		return err
// 	}

// 	return resourceCartDiscountRead(d, m)
// }

// func resourceCartDiscountDelete(d *schema.ResourceData, m interface{}) error {
// 	client := getClient(m)
// 	version := d.Get("version").(int)
// 	_, err := client.CartDiscounts().WithId(d.Id()).Delete().WithQueryParams(platform.ByProjectKeyCartDiscountsByIDRequestMethodDeleteInput{
// 		Version: version,
// 	}).Execute(context.Background())
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func marshallCartDiscountValue(val platform.CartDiscountValue) []map[string]interface{} {
// 	if val == nil {
// 		return []map[string]interface{}{}
// 	}

// 	switch v := val.(type) {
// 	case platform.CartDiscountValueAbsolute:
// 		return []map[string]interface{}{{
// 			"type":  "absolute",
// 			"money": marshallTypedMoney(v.Money),
// 		}}
// 	case platform.CartDiscountValueFixed:
// 		return []map[string]interface{}{{
// 			"type":  "fixed",
// 			"money": marshallTypedMoney(v.Money),
// 		}}
// 	case platform.CartDiscountValueGiftLineItem:
// 		return []map[string]interface{}{{
// 			"type":                    "giftLineItem",
// 			"supply_channel_id":       v.SupplyChannel.ID,
// 			"distribution_channel_id": v.DistributionChannel.ID,
// 			"product_id":              v.Product.ID,
// 		}}
// 	case platform.CartDiscountValueRelative:
// 		return []map[string]interface{}{{
// 			"type":      "relative",
// 			"permyriad": v.Permyriad,
// 		}}
// 	}
// 	panic("Unable to marshall cart discount value")
// }

// func unmarshallCartDiscountValue(d *schema.ResourceData) (platform.CartDiscountValueDraft, error) {
// 	value := d.Get("value").([]interface{})[0].(map[string]interface{})
// 	switch value["type"].(string) {
// 	case "relative":
// 		return platform.CartDiscountValueRelativeDraft{
// 			Permyriad: value["permyriad"].(int),
// 		}, nil
// 	case "absolute":
// 		money := unmarshallTypedMoney(value)
// 		return platform.CartDiscountValueAbsoluteDraft{
// 			Money: money,
// 		}, nil
// 	case "giftLineItem":
// 		draft := &platform.CartDiscountValueGiftLineItemDraft{}

// 		if val := value["supply_channel_id"].(string); len(val) > 0 {
// 			draft.SupplyChannel = &platform.ChannelResourceIdentifier{ID: &val}
// 		}
// 		if val := value["product_id"].(string); len(val) > 0 {
// 			draft.Product = platform.ProductResourceIdentifier{ID: &val}
// 		}
// 		if val := value["distribution_channel_id"].(string); len(val) > 0 {
// 			draft.DistributionChannel = &platform.ChannelResourceIdentifier{ID: &val}
// 		}

// 		draft.VariantId = value["variant"].(int)

// 		return draft, nil

// 	default:
// 		return nil, fmt.Errorf("Value type %s not implemented", value["type"])
// 	}
// }

// func marshallCartDiscountTarget(val platform.CartDiscountTarget) []map[string]interface{} {
// 	switch v := val.(type) {
// 	case platform.CartDiscountLineItemsTarget:
// 		return []map[string]interface{}{{
// 			"type":      "lineItems",
// 			"predicate": v.Predicate,
// 		}}
// 	case platform.CartDiscountCustomLineItemsTarget:
// 		return []map[string]interface{}{{
// 			"type":      "customLineItems",
// 			"predicate": v.Predicate,
// 		}}
// 	case platform.CartDiscountShippingCostTarget:
// 		return []map[string]interface{}{{
// 			"type": "shipping",
// 		}}
// 	}

// 	panic("Unable to marshall cart discount target")
// }

// func unmarshallCartDiscountTarget(d *schema.ResourceData) (platform.CartDiscountTarget, error) {
// 	input, err := elementFromList(d, "target")
// 	if err != nil {
// 		return nil, err
// 	}

// 	if input == nil {
// 		return nil, nil
// 	}

// 	switch input["type"].(string) {
// 	case "lineItems":
// 		return platform.CartDiscountLineItemsTarget{
// 			Predicate: input["predicate"].(string),
// 		}, nil
// 	case "customLineItems":
// 		return platform.CartDiscountCustomLineItemsTarget{
// 			Predicate: input["predicate"].(string),
// 		}, nil
// 	case "shipping":
// 		return platform.CartDiscountShippingCostTarget{}, nil
// 	default:
// 		return nil, fmt.Errorf("Target type %s not implemented", input["type"])
// 	}

// }

// func unmarshallCartDiscountStackingMode(d *schema.ResourceData) (platform.StackingMode, error) {
// 	switch d.Get("stacking_mode").(string) {
// 	case "Stacking":
// 		return platform.StackingModeStacking, nil
// 	case "StopAfterThisDiscount":
// 		return platform.StackingModeStopAfterThisDiscount, nil
// 	default:
// 		return "", fmt.Errorf("Stacking mode %s not implemented", d.Get("stacking_mode").(string))
// 	}
// }

// func resourceCartDiscountResourceV0() *schema.Resource {
// 	return &schema.Resource{
// 		Schema: map[string]*schema.Schema{
// 			"key": {
// 				Description: "User-specific unique identifier for a cart discount. Must be unique across a project",
// 				Type:        types.StringType,
// 				Optional:    true,
// 			},
// 			"name": {
// 				Description: "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
// 				Type:        TypeLocalizedString,
// 				Required:    true,
// 			},
// 			"description": {
// 				Description: "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
// 				Type:        TypeLocalizedString,
// 				Optional:    true,
// 			},
// 			"value": {
// 				Description: "Defines the effect the discount will have. " +
// 					"[CartDiscountValue](https://docs.commercetools.com/api/projects/cartDiscounts#cartdiscountvalue)",
// 				Type:     types.ListType{ElemType: types.StringType},
// 				MaxItems: 1,
// 				Required: true,
// 				Elem: &schema.Resource{
// 					Schema: map[string]*schema.Schema{
// 						"type": {
// 							Description:  "Currently supports absolute/relative/giftLineItem",
// 							Type:         types.StringType,
// 							Required:     true,
// 							ValidateFunc: validateValueType,
// 						},
// 						"permyriad": {
// 							Description: "Relative discount specific fields",
// 							Type:        types.Int64Type,
// 							Optional:    true,
// 						},
// 						"money": {
// 							Description: "Absolute discount specific fields",
// 							Type:        types.ListType{ElemType: types.StringType},
// 							Optional:    true,
// 							Elem: &schema.Resource{
// 								Schema: map[string]*schema.Schema{
// 									"currency_code": {
// 										Description:  "The currency code compliant to [ISO 4217](https://en.wikipedia.org/wiki/ISO_4217)",
// 										Type:         types.StringType,
// 										Required:     true,
// 										ValidateFunc: ValidateCurrencyCode,
// 									},
// 									"cent_amount": {
// 										Description: "The amount in cents (the smallest indivisible unit of the currency)",
// 										Type:        types.Int64Type,
// 										Required:    true,
// 									},
// 								},
// 							},
// 						},
// 						"product_id": {
// 							Description: "Gift Line Item discount specific field",
// 							Type:        types.StringType,
// 							Optional:    true,
// 						},
// 						"variant": {
// 							Description: "Gift Line Item discount specific field",
// 							Type:        types.Int64Type,
// 							Optional:    true,
// 						},
// 						"supply_channel_id": {
// 							Description: "Gift Line Item discount specific field",
// 							Type:        types.StringType,
// 							Optional:    true,
// 						},
// 						"distribution_channel_id": {
// 							Description: "Gift Line Item discount specific field",
// 							Type:        types.StringType,
// 							Optional:    true,
// 						},
// 					},
// 				},
// 			},
// 			"predicate": {
// 				Description: "A valid [Cart Predicate](https://docs.commercetools.com/api/projects/predicates#cart-predicates)",
// 				Type:        types.StringType,
// 				Required:    true,
// 			},
// 			"target": {
// 				Description: "Empty when the value has type giftLineItem, otherwise a " +
// 					"[CartDiscountTarget](https://docs.commercetools.com/api/projects/cartDiscounts#cartdiscounttarget)",
// 				Type:     schema.TypeMap,
// 				Optional: true,
// 				Elem: &schema.Schema{
// 					Type: types.StringType,
// 				},
// 			},
// 			"sort_order": {
// 				Description: "The string must contain a number between 0 and 1. All matching cart discounts are " +
// 					"applied to a cart in the order defined by this field. A discount with greater sort order is " +
// 					"prioritized higher than a discount with lower sort order. The sort order is unambiguous among all cart discounts",
// 				Type:     types.StringType,
// 				Required: true,
// 			},
// 			"is_active": {
// 				Description: "Only active discount can be applied to the cart",
// 				Type:        types.BoolType,
// 				Optional:    true,
// 				Default:     true,
// 			},
// 			"valid_from": {
// 				Type:     types.StringType,
// 				Optional: true,
// 			},
// 			"valid_until": {
// 				Type:     types.StringType,
// 				Optional: true,
// 			},
// 			"requires_discount_code": {
// 				Description: "States whether the discount can only be used in a connection with a " +
// 					"[DiscountCode](https://docs.commercetools.com/api/projects/discountCodes#discountcode)",
// 				Type:     types.BoolType,
// 				Optional: true,
// 				Default:  false,
// 			},
// 			"stacking_mode": {
// 				Description:  "Specifies whether the application of this discount causes the following discounts to be ignored",
// 				Type:         types.StringType,
// 				Optional:     true,
// 				ValidateFunc: validateStackingMode,
// 				Default:      "Stacking",
// 			},
// 			"version": {
// 				Type:     types.Int64Type,
// 				Computed: true,
// 			},
// 		},
// 	}
// }

// func migrateCartDiscountStateV0toV1(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
// 	transformToList(rawState, "target")
// 	return rawState, nil
// }
