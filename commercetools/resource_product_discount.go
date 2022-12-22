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

func resourceProductDiscount() *schema.Resource {
	return &schema.Resource{
		Description: "Product discounts are used to reduce certain product prices.\n\n" +
			"Also see the [Product Discount API Documentation](https://docs.commercetools.com/api/projects/productDiscounts).",
		CreateContext: resourceProductDiscountCreate,
		ReadContext:   resourceProductDiscountRead,
		UpdateContext: resourceProductDiscountUpdate,
		DeleteContext: resourceProductDiscountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"key": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "User-defined unique identifier for the ProductDiscount. Must be unique across a project",
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
							ValidateFunc: validateProductDiscountValueType,
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
									"fraction_digits": {
										Description: "The number of default fraction digits for the given currency, like 2 for EUR or 0 for JPY",
										Type:        schema.TypeInt,
										Optional:    true,
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
			"predicate": {
				Description: "A valid [Product Predicate](https://docs.commercetools.com/api/projects/predicates#product-predicates)",
				Type:        schema.TypeString,
				Required:    true,
			},
			"sort_order": {
				Description: "The string must contain a number between 0 and 1. All matching product discounts are " +
					"applied to a product in the order defined by this field. A discount with greater sort order is " +
					"prioritized higher than a discount with lower sort order. The sort order is unambiguous among all product discounts",
				Type:     schema.TypeString,
				Required: true,
			},
			"is_active": {
				Description: "When set the product discount is applied to products matching the predicate",
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
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceProductDiscountCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	name := expandLocalizedString(d.Get("name"))
	description := expandLocalizedString(d.Get("description"))

	value, err := expandProductDiscountValue(d)
	if err != nil {
		return diag.FromErr(err)
	}

	draft := platform.ProductDiscountDraft{
		Key:         nilIfEmpty(stringRef(d.Get("key"))),
		Name:        name,
		Description: &description,
		Value:       &value,
		Predicate:   d.Get("predicate").(string),
		SortOrder:   d.Get("sort_order").(string),
		IsActive:    d.Get("is_active").(bool),
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

	var productDiscount *platform.ProductDiscount
	err = resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		var err error
		productDiscount, err = client.ProductDiscounts().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(productDiscount.ID)
	d.Set("version", productDiscount.Version)

	return resourceProductDiscountRead(ctx, d, m)
}

func resourceProductDiscountRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	productDiscount, err := client.ProductDiscounts().WithId(d.Id()).Get().Execute(ctx)
	if err != nil {
		if utils.IsResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if productDiscount == nil {
		d.SetId("")
	} else {
		d.Set("version", productDiscount.Version)
		d.Set("key", productDiscount.Key)
		d.Set("name", productDiscount.Name)
		d.Set("description", productDiscount.Description)
		d.Set("value", flattenProductDiscountValue(productDiscount.Value))
		d.Set("predicate", productDiscount.Predicate)
		d.Set("sort_order", productDiscount.SortOrder)
		d.Set("is_active", productDiscount.IsActive)
		d.Set("valid_from", flattenTime(productDiscount.ValidFrom))
		d.Set("valid_until", flattenTime(productDiscount.ValidUntil))
	}

	return nil
}

func resourceProductDiscountUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	input := platform.ProductDiscountUpdate{
		Version: d.Get("version").(int),
		Actions: []platform.ProductDiscountUpdateAction{},
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&platform.ProductDiscountSetKeyAction{Key: &newKey})
	}

	if d.HasChange("name") {
		newName := expandLocalizedString(d.Get("name"))
		input.Actions = append(
			input.Actions,
			&platform.ProductDiscountChangeNameAction{Name: newName})
	}

	if d.HasChange("description") {
		newDescription := expandLocalizedString(d.Get("description"))
		input.Actions = append(
			input.Actions,
			&platform.ProductDiscountSetDescriptionAction{Description: &newDescription})
	}

	if d.HasChange("value") {
		value, err := expandProductDiscountValue(d)
		if err != nil {
			return diag.FromErr(err)
		}
		input.Actions = append(
			input.Actions,
			&platform.ProductDiscountChangeValueAction{Value: value})
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

	if d.HasChange("valid_from") {
		if val := d.Get("valid_from").(string); len(val) > 0 {
			newValidFrom, err := expandTime(d.Get("valid_from").(string))
			if err != nil {
				return diag.FromErr(err)
			}
			input.Actions = append(
				input.Actions,
				&platform.ProductDiscountSetValidFromAction{ValidFrom: &newValidFrom})
		} else {
			input.Actions = append(
				input.Actions,
				&platform.ProductDiscountSetValidFromAction{})
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
				&platform.ProductDiscountSetValidUntilAction{ValidUntil: &newValidUntil})
		} else {
			input.Actions = append(
				input.Actions,
				&platform.ProductDiscountSetValidUntilAction{})
		}
	}

	err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		_, err := client.ProductDiscounts().WithId(d.Id()).Post(input).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})

	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	return resourceProductDiscountRead(ctx, d, m)
}

func resourceProductDiscountDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	version := d.Get("version").(int)

	err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		_, err := client.ProductDiscounts().WithId(d.Id()).Delete().Version(version).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func expandProductDiscountValue(d *schema.ResourceData) (platform.ProductDiscountValueDraft, error) {
	value := d.Get("value").([]any)[0].(map[string]any)
	switch value["type"].(string) {
	case "relative":
		return platform.ProductDiscountValueRelativeDraft{
			Permyriad: value["permyriad"].(int),
		}, nil
	case "absolute":
		money := expandMoneyDraft(value)
		return platform.ProductDiscountValueAbsoluteDraft{
			Money: money,
		}, nil
	case "external":
		return platform.ProductDiscountValueExternalDraft{}, nil
	default:
		return nil, fmt.Errorf("value type %s not implemented", value["type"])
	}
}

func flattenProductDiscountValue(val platform.ProductDiscountValue) []map[string]any {
	if val == nil {
		return []map[string]any{}
	}

	switch v := val.(type) {
	case platform.ProductDiscountValueAbsolute:
		manyMoney := make([]map[string]any, len(v.Money))
		for i, money := range v.Money {
			manyMoney[i] = flattenTypedMoney(money)
		}
		return []map[string]any{{
			"type":      "absolute",
			"money":     manyMoney,
			"permyriad": 0,
		}}
	case platform.ProductDiscountValueExternal:
		return []map[string]any{{
			"type":      "external",
			"permyriad": 0,
			"money":     []any{},
		}}
	case platform.ProductDiscountValueRelative:
		return []map[string]any{{
			"type":      "relative",
			"permyriad": v.Permyriad,
			"money":     []any{},
		}}
	}
	panic("Unable to flatten product discount value")
}

func validateProductDiscountValueType(val any, key string) (warns []string, errs []error) {
	switch val {
	case
		"relative",
		"absolute",
		"external":
		return
	default:
		errs = append(errs, fmt.Errorf("%q not a valid value for %q", val, key))
	}
	return
}
