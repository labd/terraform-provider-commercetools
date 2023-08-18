package commercetools

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/ctutils"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func resourceCartDiscount() *schema.Resource {
	return &schema.Resource{
		Description: "Cart discounts are used to change the prices of different elements within a cart.\n\n" +
			"See also the [Cart Discount API Documentation](https://docs.commercetools.com/api/projects/cartDiscounts)",
		CreateContext: resourceCartDiscountCreate,
		ReadContext:   resourceCartDiscountRead,
		UpdateContext: resourceCartDiscountUpdate,
		DeleteContext: resourceCartDiscountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceCartDiscountResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: migrateCartDiscountStateV0toV1,
				Version: 0,
			},
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "User-specific unique identifier for a cart discount. Must be unique across a project",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description:      "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
				Type:             TypeLocalizedString,
				ValidateDiagFunc: validateLocalizedStringKey,
				Required:         true,
			},
			"description": {
				Description:      "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
				Type:             TypeLocalizedString,
				ValidateDiagFunc: validateLocalizedStringKey,
				Optional:         true,
			},
			"value": {
				Description: "Defines the effect the discount will have. " +
					"[CartDiscountValue](https://docs.commercetools.com/api/projects/cartDiscounts#cartdiscountvalue)",
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Description:  "Currently supports absolute/relative/giftLineItem",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateCartDiscountValueType,
						},
						"permyriad": {
							Description: "Relative discount specific fields",
							Type:        schema.TypeInt,
							Optional:    true,
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
						"product_id": {
							Description: "ResourceIdentifier of a Product. Required when value type is giftLineItem",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"variant_id": {
							Description: "ProductVariant of the Product. Required when value type is giftLineItem",
							Type:        schema.TypeInt,
							Optional:    true,
						},
						"supply_channel_id": {
							Description: "Channel must have the role InventorySupply. " +
								"Optional when value type is giftLineItem",
							Type:     schema.TypeString,
							Optional: true,
						},
						"distribution_channel_id": {
							Description: "Channel must have the role ProductDistribution. " +
								"Optional when value type is giftLineItem",
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"predicate": {
				Description: "A valid [Cart Predicate](https://docs.commercetools.com/api/projects/predicates#cart-predicates)",
				Type:        schema.TypeString,
				Required:    true,
			},
			"target": {
				Description: "Empty when the value has type giftLineItem, otherwise a " +
					"[CartDiscountTarget](https://docs.commercetools.com/api/projects/cartDiscounts#cartdiscounttarget)",
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Description:  "Supports lineItems/customLineItems/multiBuyLineItems/multiBuyCustomLineItems/shipping",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateTargetType,
						},
						"predicate": {
							Description: "LineItems/CustomLineItems/MultiBuyLineItems/MultiBuyCustomLineItems target specific fields",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"trigger_quantity": {
							Description: "MultiBuyLineItems/MultiBuyCustomLineItems target specific fields",
							Type:        schema.TypeInt,
							Optional:    true,
						},
						"discounted_quantity": {
							Description: "MultiBuyLineItems/MultiBuyCustomLineItems target specific fields",
							Type:        schema.TypeInt,
							Optional:    true,
						},
						"max_occurrence": {
							Description: "MultiBuyLineItems/MultiBuyCustomLineItems target specific fields",
							Type:        schema.TypeInt,
							Optional:    true,
						},
						"selection_mode": {
							Description:  "MultiBuyLineItems/MultiBuyCustomLineItems target specific fields",
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateSelectionMode,
						},
					},
				},
			},
			"sort_order": {
				Description: "The string must contain a number between 0 and 1. All matching cart discounts are " +
					"applied to a cart in the order defined by this field. A discount with greater sort order is " +
					"prioritized higher than a discount with lower sort order. The sort order is unambiguous among all cart discounts",
				Type:     schema.TypeString,
				Required: true,
			},
			"is_active": {
				Description: "Only active discount can be applied to the cart",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"valid_from": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: diffSuppressDateString,
			},
			"valid_until": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: diffSuppressDateString,
			},
			"requires_discount_code": {
				Description: "States whether the discount can only be used in a connection with a " +
					"[DiscountCode](https://docs.commercetools.com/api/projects/discountCodes#discountcode)",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"stacking_mode": {
				Description:  "Specifies whether the application of this discount causes the following discounts to be ignored",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateStackingMode,
				Default:      "Stacking",
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"custom": CustomFieldSchema(),
		},
	}
}

func validateCartDiscountValueType(val any, key string) (warns []string, errs []error) {
	switch val {
	case
		"relative",
		"absolute",
		"fixed",
		"giftLineItem":
		return
	default:
		errs = append(errs, fmt.Errorf("%q not a valid value for %q", val, key))
	}
	return
}

func validateTargetType(val any, key string) (warns []string, errs []error) {
	switch val {
	case
		"lineItems",
		"customLineItems",
		"multiBuyLineItems",
		"multiBuyCustomLineItems",
		"shipping":
		return
	default:
		errs = append(errs, fmt.Errorf("%q not a valid value for %q", val, key))
	}
	return
}

func validateStackingMode(val any, key string) (warns []string, errs []error) {
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

func validateSelectionMode(val any, key string) (warns []string, errs []error) {
	switch val {
	case
		"Cheapest",
		"MostExpensive":
		return
	default:
		errs = append(errs, fmt.Errorf("%q not a valid value for %q", val, key))
	}
	return
}

func resourceCartDiscountCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	name := expandLocalizedString(d.Get("name"))
	description := expandLocalizedString(d.Get("description"))

	value, err := expandCartDiscountValue(d)
	if err != nil {
		return diag.FromErr(err)
	}

	stackingMode, err := expandCartDiscountStackingMode(d)
	if err != nil {
		return diag.FromErr(err)
	}

	custom, err := CreateCustomFieldDraft(ctx, client, d)
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	draft := platform.CartDiscountDraft{
		Name:                 name,
		Description:          &description,
		Value:                &value,
		CartPredicate:        d.Get("predicate").(string),
		SortOrder:            d.Get("sort_order").(string),
		IsActive:             boolRef(d.Get("is_active")),
		RequiresDiscountCode: ctutils.BoolRef(d.Get("requires_discount_code").(bool)),
		Custom:               custom,
		StackingMode:         &stackingMode,
	}

	key := stringRef(d.Get("key"))
	if *key != "" {
		draft.Key = key
	}

	if val, err := expandCartDiscountTarget(d); err == nil {
		draft.Target = val
	} else {
		return diag.FromErr(err)
	}

	if val := d.Get("valid_from").(string); len(val) > 0 {
		validFrom, err := expandTime(val)
		if err != nil {
			return diag.FromErr(err)
		}
		draft.ValidFrom = &validFrom
	}
	if val := d.Get("valid_until").(string); len(val) > 0 {
		validUntil, err := expandTime(val)
		if err != nil {
			return diag.FromErr(err)
		}
		draft.ValidUntil = &validUntil
	}

	var cartDiscount *platform.CartDiscount
	err = resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		var err error
		cartDiscount, err = client.CartDiscounts().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if cartDiscount == nil {
		return diag.Errorf("No cart discount created")
	}

	d.SetId(cartDiscount.ID)
	_ = d.Set("version", cartDiscount.Version)

	return resourceCartDiscountRead(ctx, d, m)
}

func resourceCartDiscountRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	cartDiscount, err := client.CartDiscounts().WithId(d.Id()).Get().Execute(ctx)
	if err != nil {
		if utils.IsResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("version", cartDiscount.Version)
	_ = d.Set("key", cartDiscount.Key)
	_ = d.Set("name", cartDiscount.Name)
	_ = d.Set("description", cartDiscount.Description)
	_ = d.Set("value", flattenCartDiscountValue(cartDiscount.Value))
	_ = d.Set("predicate", cartDiscount.CartPredicate)
	_ = d.Set("target", flattenCartDiscountTarget(cartDiscount.Target))
	_ = d.Set("sort_order", cartDiscount.SortOrder)
	_ = d.Set("is_active", cartDiscount.IsActive)
	_ = d.Set("valid_from", flattenTime(cartDiscount.ValidFrom))
	_ = d.Set("valid_until", flattenTime(cartDiscount.ValidUntil))
	_ = d.Set("requires_discount_code", cartDiscount.RequiresDiscountCode)
	_ = d.Set("stacking_mode", cartDiscount.StackingMode)
	_ = d.Set("custom", flattenCustomFields(cartDiscount.Custom))
	return nil
}

func resourceCartDiscountUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	input := platform.CartDiscountUpdate{
		Version: d.Get("version").(int),
		Actions: []platform.CartDiscountUpdateAction{},
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&platform.CartDiscountSetKeyAction{Key: &newKey})
	}

	if d.HasChange("name") {
		newName := expandLocalizedString(d.Get("name"))
		input.Actions = append(
			input.Actions,
			&platform.CartDiscountChangeNameAction{Name: newName})
	}

	if d.HasChange("description") {
		newDescription := expandLocalizedString(d.Get("description"))
		input.Actions = append(
			input.Actions,
			&platform.CartDiscountSetDescriptionAction{Description: &newDescription})
	}

	if d.HasChange("value") {
		value, err := expandCartDiscountValue(d)
		if err != nil {
			return diag.FromErr(err)
		}
		input.Actions = append(
			input.Actions,
			&platform.CartDiscountChangeValueAction{Value: value})
	}

	if d.HasChange("predicate") {
		newPredicate := d.Get("predicate").(string)
		input.Actions = append(
			input.Actions,
			&platform.CartDiscountChangeCartPredicateAction{CartPredicate: newPredicate})
	}

	if d.HasChange("target") {
		if val, err := expandCartDiscountTarget(d); err == nil {
			if val != nil {
				input.Actions = append(
					input.Actions,
					&platform.CartDiscountChangeTargetAction{Target: val})
			} else {
				return diag.Errorf("Cannot change target to empty")
			}
		} else {
			return diag.FromErr(err)
		}

	}

	if d.HasChange("sort_order") {
		newSortOrder := d.Get("sort_order").(string)
		input.Actions = append(
			input.Actions,
			&platform.CartDiscountChangeSortOrderAction{SortOrder: newSortOrder})
	}

	if d.HasChange("is_active") {
		newIsActive := d.Get("is_active").(bool)
		input.Actions = append(
			input.Actions,
			&platform.CartDiscountChangeIsActiveAction{IsActive: newIsActive})
	}

	if d.HasChange("valid_from") {
		if val := d.Get("valid_from").(string); len(val) > 0 {
			newValidFrom, err := expandTime(d.Get("valid_from").(string))
			if err != nil {
				return diag.FromErr(err)
			}
			input.Actions = append(
				input.Actions,
				&platform.CartDiscountSetValidFromAction{ValidFrom: &newValidFrom})
		} else {
			input.Actions = append(
				input.Actions,
				&platform.CartDiscountSetValidFromAction{})
		}
	}

	if d.HasChange("valid_until") {
		if val := d.Get("valid_until").(string); len(val) > 0 {
			newValidUntil, err := expandTime(d.Get("valid_until").(string))
			if err != nil {
				return diag.FromErr(err)
			}
			input.Actions = append(
				input.Actions,
				&platform.CartDiscountSetValidUntilAction{ValidUntil: &newValidUntil})
		} else {
			input.Actions = append(
				input.Actions,
				&platform.CartDiscountSetValidUntilAction{})
		}
	}

	if d.HasChange("requires_discount_code") {
		newRequiresDiscountCode := d.Get("requires_discount_code").(bool)
		input.Actions = append(
			input.Actions,
			&platform.CartDiscountChangeRequiresDiscountCodeAction{RequiresDiscountCode: newRequiresDiscountCode})
	}

	if d.HasChange("stacking_mode") {
		newStackingMode, err := expandCartDiscountStackingMode(d)
		if err != nil {
			return diag.FromErr(err)
		}
		input.Actions = append(
			input.Actions,
			&platform.CartDiscountChangeStackingModeAction{StackingMode: newStackingMode})
	}

	if d.HasChange("custom") {
		actions, err := CustomFieldUpdateActions[platform.CartDiscountSetCustomTypeAction, platform.CartDiscountSetCustomFieldAction](ctx, client, d)
		if err != nil {
			return diag.FromErr(err)
		}
		for i := range actions {
			input.Actions = append(input.Actions, actions[i].(platform.CartDiscountUpdateAction))
		}
	}

	err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		_, err := client.CartDiscounts().WithId(d.Id()).Post(input).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})

	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	return resourceCartDiscountRead(ctx, d, m)
}

func resourceCartDiscountDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	version := d.Get("version").(int)

	err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		_, err := client.CartDiscounts().WithId(d.Id()).Delete().Version(version).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func flattenCartDiscountValue(val platform.CartDiscountValue) []map[string]any {
	if val == nil {
		return []map[string]any{}
	}

	switch v := val.(type) {
	case platform.CartDiscountValueAbsolute:
		manyMoney := make([]map[string]any, len(v.Money))
		for i, money := range v.Money {
			manyMoney[i] = flattenTypedMoney(money)
		}
		return []map[string]any{{
			"type":  "absolute",
			"money": manyMoney,
		}}
	case platform.CartDiscountValueFixed:
		manyMoney := make([]map[string]any, len(v.Money))
		for i, money := range v.Money {
			manyMoney[i] = flattenTypedMoney(money)
		}
		return []map[string]any{{
			"type":  "fixed",
			"money": manyMoney,
		}}
	case platform.CartDiscountValueGiftLineItem:
		var supplyChannelID string
		if v.SupplyChannel != nil {
			supplyChannelID = v.SupplyChannel.ID
		}

		var distributionChannelID string
		if v.DistributionChannel != nil {
			distributionChannelID = v.DistributionChannel.ID
		}

		return []map[string]any{{
			"type":                    "giftLineItem",
			"supply_channel_id":       supplyChannelID,
			"distribution_channel_id": distributionChannelID,
			"product_id":              v.Product.ID,
			"variant_id":              v.VariantId,
		}}
	case platform.CartDiscountValueRelative:
		return []map[string]any{{
			"type":      "relative",
			"permyriad": v.Permyriad,
		}}
	}
	panic("Unable to flatten cart discount value")
}

func expandCartDiscountValue(d *schema.ResourceData) (platform.CartDiscountValueDraft, error) {
	value := d.Get("value").([]any)[0].(map[string]any)
	switch value["type"].(string) {
	case "relative":
		return platform.CartDiscountValueRelativeDraft{
			Permyriad: value["permyriad"].(int),
		}, nil
	case "absolute":
		money := expandTypedMoney(value)
		return platform.CartDiscountValueAbsoluteDraft{
			Money: money,
		}, nil
	case "fixed":
		money := expandTypedMoney(value)
		return platform.CartDiscountValueFixedDraft{
			Money: money,
		}, nil
	case "giftLineItem":
		draft := &platform.CartDiscountValueGiftLineItemDraft{}

		if val := value["supply_channel_id"].(string); len(val) > 0 {
			draft.SupplyChannel = &platform.ChannelResourceIdentifier{ID: &val}
		}
		if val := value["product_id"].(string); len(val) > 0 {
			draft.Product = platform.ProductResourceIdentifier{ID: &val}
		}
		if val := value["distribution_channel_id"].(string); len(val) > 0 {
			draft.DistributionChannel = &platform.ChannelResourceIdentifier{ID: &val}
		}

		draft.VariantId = value["variant_id"].(int)

		return draft, nil

	default:
		return nil, fmt.Errorf("value type %s not implemented", value["type"])
	}
}

func flattenCartDiscountTarget(val platform.CartDiscountTarget) []map[string]any {
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case platform.CartDiscountLineItemsTarget:
		return []map[string]any{{
			"type":      "lineItems",
			"predicate": v.Predicate,
		}}
	case platform.CartDiscountCustomLineItemsTarget:
		return []map[string]any{{
			"type":      "customLineItems",
			"predicate": v.Predicate,
		}}
	case platform.MultiBuyLineItemsTarget:
		return []map[string]any{{
			"type":                "multiBuyLineItems",
			"predicate":           v.Predicate,
			"trigger_quantity":    v.TriggerQuantity,
			"discounted_quantity": v.DiscountedQuantity,
			"max_occurrence":      v.MaxOccurrence,
			"selection_mode":      v.SelectionMode,
		}}
	case platform.MultiBuyCustomLineItemsTarget:
		return []map[string]any{{
			"type":                "multiBuyCustomLineItems",
			"predicate":           v.Predicate,
			"trigger_quantity":    v.TriggerQuantity,
			"discounted_quantity": v.DiscountedQuantity,
			"max_occurrence":      v.MaxOccurrence,
			"selection_mode":      v.SelectionMode,
		}}
	case platform.CartDiscountShippingCostTarget:
		return []map[string]any{{
			"type": "shipping",
		}}
	}

	panic("Unable to flatten cart discount target")
}

func expandCartDiscountTarget(d *schema.ResourceData) (platform.CartDiscountTarget, error) {
	input, err := elementFromList(d, "target")
	if err != nil {
		return nil, err
	}

	if input == nil {
		return nil, nil
	}

	switch input["type"].(string) {
	case "lineItems":
		return platform.CartDiscountLineItemsTarget{
			Predicate: input["predicate"].(string),
		}, nil
	case "customLineItems":
		return platform.CartDiscountCustomLineItemsTarget{
			Predicate: input["predicate"].(string),
		}, nil
	case "multiBuyLineItems":
		selectionMode, err := expandSelectionMode(input["selection_mode"].(string))
		if err != nil {
			return nil, err
		}
		target := platform.MultiBuyLineItemsTarget{
			Predicate:          input["predicate"].(string),
			TriggerQuantity:    input["trigger_quantity"].(int),
			DiscountedQuantity: input["discounted_quantity"].(int),
			SelectionMode:      selectionMode,
		}
		maxOccurrence := input["max_occurrence"].(int)
		if maxOccurrence > 0 {
			target.MaxOccurrence = &maxOccurrence
		}
		return target, nil
	case "multiBuyCustomLineItems":
		selectionMode, err := expandSelectionMode(input["selection_mode"].(string))
		if err != nil {
			return nil, err
		}
		target := platform.MultiBuyCustomLineItemsTarget{
			Predicate:          input["predicate"].(string),
			TriggerQuantity:    input["trigger_quantity"].(int),
			DiscountedQuantity: input["discounted_quantity"].(int),
			SelectionMode:      selectionMode,
		}
		maxOccurrence := input["max_occurrence"].(int)
		if maxOccurrence > 0 {
			target.MaxOccurrence = &maxOccurrence
		}
		return target, nil

	case "shipping":
		return platform.CartDiscountShippingCostTarget{}, nil
	default:
		return nil, fmt.Errorf("target type %s not implemented", input["type"])
	}

}

func expandCartDiscountStackingMode(d *schema.ResourceData) (platform.StackingMode, error) {
	switch d.Get("stacking_mode").(string) {
	case "Stacking":
		return platform.StackingModeStacking, nil
	case "StopAfterThisDiscount":
		return platform.StackingModeStopAfterThisDiscount, nil
	default:
		return "", fmt.Errorf("stacking mode %s not implemented", d.Get("stacking_mode").(string))
	}
}

func expandSelectionMode(selectionMode string) (platform.SelectionMode, error) {
	switch selectionMode {
	case "Cheapest":
		return platform.SelectionModeCheapest, nil
	case "MostExpensive":
		return platform.SelectionModeMostExpensive, nil
	default:
		return "", fmt.Errorf("selection mode %s not implemented", selectionMode)
	}
}
