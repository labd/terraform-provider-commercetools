package commercetools

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceProjectSettingsResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "The unique key of the project",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "The name of the project",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"currencies": {
				Description: "A three-digit currency code as per [ISO 4217](https://en.wikipedia.org/wiki/ISO_4217)",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"countries": {
				Description: "A two-digit country code as per [ISO 3166-1 alpha-2](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2)",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"languages": {
				Description: "[IETF Language Tag](https://en.wikipedia.org/wiki/IETF_language_tag)",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"messages": {
				Description: "[Messages Configuration](https://docs.commercetools.com/api/projects/project#messages-configuration)",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeBool,
				},
			},
			"external_oauth": {
				Description: "[External OAUTH](https://docs.commercetools.com/api/projects/project#externaloauth)",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"carts": {
				Description: "[Carts Configuration](https://docs.commercetools.com/api/projects/project#carts-configuration)",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"shipping_rate_input_type": {
				Description: "Three ways to dynamically select a ShippingRatePriceTier exist. The CartValue type uses " +
					"the sum of all line item prices, whereas CartClassification and CartScore use the " +
					"shippingRateInput field on the cart to select a tier",
				Type:     schema.TypeString,
				Optional: true,
			},
			"shipping_rate_cart_classification_value": {
				Description: "If shipping_rate_input_type is set to CartClassification these values are used to create " +
					"tiers\n. Only a key defined inside the values array can be used to create a tier, or to set a value " +
					"for the shippingRateInput on the cart. The keys are checked for uniqueness and the request is " +
					"rejected if keys are not unique",
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"label": {
							Type:             TypeLocalizedString,
							ValidateDiagFunc: validateLocalizedStringKey,
							Optional:         true,
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

func migrateResourceProjectSettingsStateV0toV1(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	transformToList(rawState, "messages")
	transformToList(rawState, "external_oauth")
	transformToList(rawState, "carts")
	return rawState, nil
}
