package commercetools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/commercetools/utils"
)

func resourceShippingZoneRate() *schema.Resource {
	return &schema.Resource{
		Description: "Defines shipping rates (prices) for a specific zone.\n\n" +
			"See also [ZoneRate API Documentation](https://docs.commercetools.com/api/projects/shippingMethods#zonerate)",
		CreateContext: resourceShippingZoneRateCreate,
		ReadContext:   resourceShippingZoneRateRead,
		UpdateContext: resourceShippingZoneRateUpdate,
		DeleteContext: resourceShippingZoneRateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceShippingZoneRateImportState,
		},
		Schema: map[string]*schema.Schema{
			"shipping_method_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"shipping_zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"price": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
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
			"free_above": {
				Description: "The shipping is free if the sum of the (custom) line item prices reaches the freeAbove value",
				Type:        schema.TypeList,
				MinItems:    1,
				MaxItems:    1,
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
			"shipping_rate_price_tier": {
				Description: "A price tier is selected instead of the default price when a certain threshold or " +
					"specific cart value is reached. If no tiered price is suitable for the cart, the base price of the " +
					"shipping rate is used\n. " +
					"See also [Shipping Rate Price Tier API Docs](https://docs.commercetools.com/api/projects/shippingMethods#shippingratepricetier)",
				Type:     schema.TypeList,
				MinItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Description: "CartValue, CartScore or CartClassification",
							Type:        schema.TypeString,
							Required:    true,
							ValidateFunc: validation.StringInSlice([]string{
								string(platform.ShippingRateTierTypeCartValue),
								string(platform.ShippingRateTierTypeCartScore),
								string(platform.ShippingRateTierTypeCartClassification),
							}, false),
						},
						"minimum_cent_amount": {
							Description: "If type is CartValue this represents the cent amount of the tier",
							Type:        schema.TypeInt,
							Optional:    true,
						},
						"value": {
							Description: "If type is CartClassification, must be a valid key of the CartClassification",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"score": {
							Description: "If type is CartScore. Sets a fixed price for this score value",
							Type:        schema.TypeInt,
							Optional:    true,
						},
						"price": {
							Description: "The price of the score, value or minimum_cent_amount tier",
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
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
						"price_function": {
							Description: "If type is CartScore. Allows to calculate a price dynamically for the score.",
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"currency_code": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: ValidateCurrencyCode,
									},
									"function": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceShippingZoneRateImportState(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	client := getClient(meta)
	shippingMethodID, _, _ := getShippingIDs(d.Id())

	shippingMethod, err := client.ShippingMethods().WithId(shippingMethodID).Get().Execute(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]*schema.ResourceData, 0)
	shippingZoneRateState := resourceShippingZoneRate().Data(nil)
	shippingZoneRateState.SetId(d.Id())
	shippingZoneRateState.SetType("commercetools_shipping_zone_rate")

	setShippingZoneRateState(d, shippingMethod)

	results = append(results, shippingZoneRateState)

	return results, nil
}

func resourceShippingZoneRateCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	shippingZoneID := d.Get("shipping_zone_id").(string)
	shippingMethodID := d.Get("shipping_method_id").(string)

	// Lock to prevent concurrent updates due to Version number conflicts
	ctMutexKV.Lock(shippingMethodID)
	defer ctMutexKV.Unlock(shippingMethodID)

	shippingMethod, err := client.ShippingMethods().WithId(shippingMethodID).Get().Execute(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	input := platform.ShippingMethodUpdate{
		Version: shippingMethod.Version,
		Actions: []platform.ShippingMethodUpdateAction{},
	}
	draft, err := expandShippingRateDraft(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Add the zone to the shipping method if it isn't set yet.
	zoneNotFound := true
	for _, v := range shippingMethod.ZoneRates {
		if v.Zone.ID == shippingZoneID {
			zoneNotFound = false
			break
		}
	}

	if zoneNotFound {
		input.Actions = append(input.Actions, platform.ShippingMethodAddZoneAction{
			Zone: platform.ZoneResourceIdentifier{ID: &shippingZoneID},
		})
	}

	input.Actions = append(input.Actions, platform.ShippingMethodAddShippingRateAction{
		Zone:         platform.ZoneResourceIdentifier{ID: &shippingZoneID},
		ShippingRate: *draft,
	})

	err = resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		var err error
		shippingMethod, err = client.ShippingMethods().WithId(shippingMethod.ID).Post(input).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildShippingZoneRateID(shippingMethod.ID, shippingZoneID, string(draft.Price.CurrencyCode)))
	return resourceShippingZoneRateRead(ctx, d, m)
}

func resourceShippingZoneRateRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	shippingMethodID, _, _ := getShippingIDs(d.Id())

	client := getClient(m)
	shippingMethod, err := client.ShippingMethods().WithId(shippingMethodID).Get().Execute(ctx)
	if err != nil {
		if utils.IsResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	err = setShippingZoneRateState(d, shippingMethod)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceShippingZoneRateUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	shippingMethodID, shippingZoneID, currencyCode := getShippingIDs(d.Id())
	ctMutexKV.Lock(shippingMethodID)
	defer ctMutexKV.Unlock(shippingMethodID)

	client := getClient(m)
	shippingMethod, err := client.ShippingMethods().WithId(shippingMethodID).Get().Execute(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	curShippingRate, err := findShippingZoneRate(shippingMethod, shippingZoneID, currencyCode)
	if err != nil {
		return diag.FromErr(err)
	}
	oldShippingRateDraft := createShippingRateDraft(curShippingRate)

	input := platform.ShippingMethodUpdate{
		Version: shippingMethod.Version,
		Actions: []platform.ShippingMethodUpdateAction{},
	}

	if d.HasChange("price") || d.HasChange("free_above") || d.HasChange("shipping_rate_price_tier") {
		zoneResourceIdentifier := platform.ZoneResourceIdentifier{
			ID: &shippingZoneID,
		}

		input.Actions = append(
			input.Actions,
			&platform.ShippingMethodRemoveShippingRateAction{
				Zone:         zoneResourceIdentifier,
				ShippingRate: *oldShippingRateDraft,
			})

		newShippingRateDraft, err := expandShippingRateDraft(d)
		if err != nil {
			return diag.FromErr(err)
		}

		input.Actions = append(
			input.Actions,
			&platform.ShippingMethodAddShippingRateAction{
				Zone:         zoneResourceIdentifier,
				ShippingRate: *newShippingRateDraft,
			})

		d.SetId(buildShippingZoneRateID(shippingMethod.ID, shippingZoneID, string(newShippingRateDraft.Price.CurrencyCode)))
	}

	err = resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		_, err := client.ShippingMethods().WithId(shippingMethodID).Post(input).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	return resourceShippingZoneRateRead(ctx, d, m)
}

func resourceShippingZoneRateDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	shippingMethodID := d.Get("shipping_method_id").(string)
	ctMutexKV.Lock(shippingMethodID)
	defer ctMutexKV.Unlock(shippingMethodID)

	client := getClient(m)
	shippingMethod, err := client.ShippingMethods().WithId(shippingMethodID).Get().Execute(ctx)
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	input := platform.ShippingMethodUpdate{
		Version: shippingMethod.Version,
		Actions: []platform.ShippingMethodUpdateAction{},
	}

	shippingRateDraft, err := expandShippingRateDraft(d)
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	shippingZoneID := d.Get("shipping_zone_id").(string)
	removeAction := platform.ShippingMethodRemoveShippingRateAction{
		Zone:         platform.ZoneResourceIdentifier{ID: &shippingZoneID},
		ShippingRate: *shippingRateDraft,
	}

	input.Actions = append(input.Actions, removeAction)

	// Remove the zone from the shipping methode if there are no rates for the
	// combination anymore.
	for _, v := range shippingMethod.ZoneRates {
		if v.Zone.ID == shippingZoneID && len(v.ShippingRates) == 1 {
			input.Actions = append(input.Actions, platform.ShippingMethodRemoveZoneAction{
				Zone: platform.ZoneResourceIdentifier{ID: &shippingZoneID},
			})
			break
		}
	}

	err = resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		_, err = client.ShippingMethods().WithId(shippingMethodID).Post(input).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})

	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
	}
	return diag.FromErr(err)
}

func createShippingRateDraft(rate *platform.ShippingRate) *platform.ShippingRateDraft {
	var freeAbove *platform.Money
	if rate.FreeAbove != nil {
		m := coerceTypedMoney(rate.FreeAbove)
		freeAbove = &m
	}

	return &platform.ShippingRateDraft{
		Price:     coerceTypedMoney(rate.Price),
		FreeAbove: freeAbove,
		Tiers:     rate.Tiers,
	}

}

func getShippingIDs(shippingZoneRateID string) (string, string, string) {
	idSplit := strings.Split(shippingZoneRateID, "@")

	shippingMethodID := idSplit[0]
	shippingZoneID := idSplit[1]
	currencyCode := idSplit[2]

	return shippingMethodID, shippingZoneID, currencyCode
}

// find the shippingRate in a shippingMethod. This is done by a combination of
// the zone id and the ccurrency of the rate. The currency must be unique within
// commercetools so this should be safe.
func findShippingZoneRate(shippingMethod *platform.ShippingMethod, shippingZoneID string, currencyCode string) (*platform.ShippingRate, error) {
	for _, zoneRate := range shippingMethod.ZoneRates {
		if zoneRate.Zone.ID == shippingZoneID {
			for _, shippingRate := range zoneRate.ShippingRates {
				if shippingRate.Price.(platform.CentPrecisionMoney).CurrencyCode == currencyCode {
					return &shippingRate, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("couldn't find shipping zone rate")
}

func setShippingZoneRateState(d *schema.ResourceData, shippingMethod *platform.ShippingMethod) error {
	shippingMethodID, shippingZoneID, currencyCode := getShippingIDs(d.Id())

	d.Set("shipping_method_id", shippingMethodID)
	d.Set("shipping_zone_id", shippingZoneID)

	shippingRate, err := findShippingZoneRate(shippingMethod, shippingZoneID, currencyCode)

	if err != nil {
		return err
	}

	tiers := flattenShippingZoneRateTiers(shippingRate)
	d.Set("shipping_rate_price_tier", tiers)

	if typedPrice, ok := shippingRate.Price.(platform.CentPrecisionMoney); ok {
		price := map[string]any{
			"currency_code": string(typedPrice.CurrencyCode),
			"cent_amount":   typedPrice.CentAmount,
		}
		err = d.Set("price", []any{price})
		if err != nil {
			return err
		}
	} else {
		d.Set("price", nil)
		if err != nil {
			return err
		}
	}

	if typedFreeAbove, ok := (shippingRate.FreeAbove).(platform.CentPrecisionMoney); ok {
		freeAbove := map[string]any{
			"currency_code": string(typedFreeAbove.CurrencyCode),
			"cent_amount":   typedFreeAbove.CentAmount,
		}
		err = d.Set("free_above", []any{freeAbove})
		if err != nil {
			return err
		}
	} else {
		d.Set("free_above", nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func flattenShippingZoneRateTiers(shippingRate *platform.ShippingRate) []any {
	tiers := []any{}

	for _, v := range shippingRate.Tiers {
		switch shippingRateTier := v.(type) {
		case platform.CartClassificationTier:
			tiers = append(tiers, map[string]any{
				"type":  string(platform.ShippingRateTierTypeCartClassification),
				"value": shippingRateTier.Value,
				"price": []any{
					map[string]any{
						"currency_code": shippingRateTier.Price.CurrencyCode,
						"cent_amount":   shippingRateTier.Price.CentAmount,
					},
				},
			})
		case platform.CartScoreTier:
			tierData := map[string]any{
				"type":  string(platform.ShippingRateTierTypeCartScore),
				"score": shippingRateTier.Score,
			}

			if shippingRateTier.PriceFunction != nil {
				tierData["price_function"] = []any{
					map[string]any{
						"currency_code": shippingRateTier.PriceFunction.CurrencyCode,
						"function":      shippingRateTier.PriceFunction.Function,
					},
				}
			}
			if shippingRateTier.Price != nil {
				tierData["price"] = []any{
					map[string]any{
						"currency_code": shippingRateTier.Price.CurrencyCode,
						"cent_amount":   shippingRateTier.Price.CentAmount,
					},
				}
			}
			tiers = append(tiers, tierData)
		case platform.CartValueTier:
			tiers = append(tiers, map[string]any{
				"type":                string(platform.ShippingRateTierTypeCartValue),
				"minimum_cent_amount": shippingRateTier.MinimumCentAmount,
				"price": []any{
					map[string]any{
						"currency_code": shippingRateTier.Price.CurrencyCode,
						"cent_amount":   shippingRateTier.Price.CentAmount,
					},
				},
			})
		}
	}

	return tiers
}

func expandShippingRateDraft(d *schema.ResourceData) (*platform.ShippingRateDraft, error) {
	shippingRatePriceTiers, err := expandShippingRatePriceTiers(d)
	if err != nil {
		return nil, err
	}

	draft := &platform.ShippingRateDraft{
		Tiers: shippingRatePriceTiers,
	}

	if price, _ := elementFromList(d, "price"); price != nil {
		draft.Price = platform.Money{
			CurrencyCode: price["currency_code"].(string),
			CentAmount:   price["cent_amount"].(int),
		}
	}

	if price, _ := elementFromList(d, "free_above"); price != nil {
		draft.FreeAbove = &platform.Money{
			CurrencyCode: price["currency_code"].(string),
			CentAmount:   price["cent_amount"].(int),
		}
	}

	return draft, nil

}

func expandShippingRatePriceTiers(d *schema.ResourceData) ([]platform.ShippingRatePriceTier, error) {
	values, ok := d.GetOk("shipping_rate_price_tier")
	if !ok {
		return []platform.ShippingRatePriceTier{}, nil
	}

	var tiers []platform.ShippingRatePriceTier
	for _, priceTier := range values.([]any) {
		tierMap := priceTier.(map[string]any)

		var price *platform.Money
		if rawPrice := elementFromSlice(tierMap, "price"); rawPrice != nil {
			price = &platform.Money{
				CurrencyCode: rawPrice["currency_code"].(string),
				CentAmount:   rawPrice["cent_amount"].(int),
			}
		}

		tierType := tierMap["type"].(string)
		switch tierType {
		case string(platform.ShippingRateTierTypeCartValue):
			tiers = append(tiers, platform.CartValueTier{
				MinimumCentAmount: tierMap["minimum_cent_amount"].(int),
				Price:             *price,
			})

		case string(platform.ShippingRateTierTypeCartClassification):
			tiers = append(tiers, platform.CartClassificationTier{
				Value: tierMap["value"].(string),
				Price: *price,
			})

		// CartScore has either a `price` or `price_function` field.
		case string(platform.ShippingRateTierTypeCartScore):
			var function *platform.PriceFunction
			if rawFunc := elementFromSlice(tierMap, "price_function"); rawFunc != nil {
				function = &platform.PriceFunction{
					CurrencyCode: rawFunc["currency_code"].(string),
					Function:     rawFunc["function"].(string),
				}
			}

			tiers = append(tiers, platform.CartScoreTier{
				Score:         tierMap["score"].(int),
				Price:         price,
				PriceFunction: function,
			})
			// Do we want to fail on 1 wrong tier?
		default:
			return nil, fmt.Errorf("invalid shippingRatePriceTier type: %s", tierType)
		}
	}
	return tiers, nil
}

func buildShippingZoneRateID(shippingMethodID string, shippingZoneID string, currencyCode string) string {
	return shippingMethodID + "@" + shippingZoneID + "@" + currencyCode
}
