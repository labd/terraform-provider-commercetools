package commercetools

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCartDiscountResourceV0() *schema.Resource {
	return &schema.Resource{
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
							ValidateFunc: validateValueType,
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
							Description: "Gift Line Item discount specific field",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"variant": {
							Description: "Gift Line Item discount specific field",
							Type:        schema.TypeInt,
							Optional:    true,
						},
						"supply_channel_id": {
							Description: "Gift Line Item discount specific field",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"distribution_channel_id": {
							Description: "Gift Line Item discount specific field",
							Type:        schema.TypeString,
							Optional:    true,
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
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
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
		},
	}
}

func migrateCartDiscountStateV0toV1(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	transformToList(rawState, "target")
	return rawState, nil
}
