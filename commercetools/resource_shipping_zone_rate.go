package commercetools

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/labd/commercetools-go-sdk/platform"
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
							ForceNew:     true,
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
							Type:        schema.TypeFloat,
							Optional:    true,
						},
						"price": {
							Description: "The price of the score, value or minimum_cent_amount tier",
							Type:        schema.TypeList,
							Required:    true,
							MinItems:    1,
							MaxItems:    1,
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
					},
				},
			},
		},
	}
}

func resourceShippingZoneRateImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

	log.Printf("[DEBUG] Importing results: %#v", results)

	return results, nil
}

func resourceShippingZoneRateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	price := d.Get("price").([]interface{})[0].(map[string]interface{})

	var freeAbove *platform.Money
	if freeAboveState, ok := d.GetOk("free_above"); ok {
		log.Printf("[DEBUG] Free above state: %s", stringFormatObject(freeAboveState))
		freeAboveMap := freeAboveState.([]interface{})[0].(map[string]interface{})
		freeAbove = &platform.Money{
			CurrencyCode: freeAboveMap["currency_code"].(string),
			CentAmount:   freeAboveMap["cent_amount"].(int),
		}
	}
	log.Printf("[DEBUG] Setting freeAbove: %s", stringFormatObject(freeAbove))

	shippingRatePriceTiers, err := unmarshallShippingRatePriceTiers(d)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] Setting shippingRatePriceTiers: %s", stringFormatObject(shippingRatePriceTiers))

	priceCurrencyCode := price["currency_code"].(string)

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
		Zone: platform.ZoneResourceIdentifier{ID: &shippingZoneID},
		ShippingRate: platform.ShippingRateDraft{
			Price: platform.Money{
				CurrencyCode: priceCurrencyCode,
				CentAmount:   price["cent_amount"].(int),
			},
			FreeAbove: freeAbove,
			Tiers:     shippingRatePriceTiers,
		},
	})

	err = resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		var err error

		shippingMethod, err = client.ShippingMethods().WithId(shippingMethod.ID).Post(input).Execute(ctx)
		return processRemoteError(err)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if shippingMethod == nil {
		return diag.Errorf("No shipping method created?")
	}

	d.SetId(buildShippingZoneRateID(shippingMethod.ID, shippingZoneID, string(priceCurrencyCode)))

	return resourceShippingZoneRateRead(ctx, d, m)
}

func unmarshallShippingRatePriceTiers(d *schema.ResourceData) ([]platform.ShippingRatePriceTier, error) {
	values, ok := d.GetOk("shipping_rate_price_tier")
	if !ok {
		return []platform.ShippingRatePriceTier{}, nil
	}

	var tiers []platform.ShippingRatePriceTier
	for _, priceTier := range values.([]interface{}) {
		tierMap := priceTier.(map[string]interface{})

		priceMap := tierMap["price"].([]interface{})[0].(map[string]interface{})
		price := platform.Money{
			CurrencyCode: priceMap["currency_code"].(string),
			CentAmount:   priceMap["cent_amount"].(int),
		}

		tierType := tierMap["type"].(string)
		switch tierType {
		case string(platform.ShippingRateTierTypeCartValue):
			tiers = append(tiers, platform.CartValueTier{
				MinimumCentAmount: tierMap["minimum_cent_amount"].(int),
				Price:             price,
			})
		case string(platform.ShippingRateTierTypeCartClassification):
			tiers = append(tiers, platform.CartClassificationTier{
				Value: tierMap["value"].(string),
				Price: price,
			})
		case string(platform.ShippingRateTierTypeCartScore):
			tiers = append(tiers, platform.CartScoreTier{
				Score: tierMap["score"].(float64),
				Price: &price,
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

func resourceShippingZoneRateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Reading shipping zone rate from commercetools, with id: %s", d.Id())

	shippingMethodID, _, _ := getShippingIDs(d.Id())

	client := getClient(m)

	shippingMethod, err := client.ShippingMethods().WithId(shippingMethodID).Get().Execute(ctx)

	if err != nil {
		if IsResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if shippingMethod == nil {
		log.Print("[DEBUG] No shipping method found")
		d.SetId("")
	} else {
		log.Print("[DEBUG] Found following shipping method:")
		log.Print(stringFormatObject(shippingMethod))

		err = setShippingZoneRateState(d, shippingMethod)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceShippingZoneRateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	shippingMethodID, shippingZoneID, currencyCode := getShippingIDs(d.Id())
	ctMutexKV.Lock(shippingMethodID)
	defer ctMutexKV.Unlock(shippingMethodID)

	client := getClient(m)
	shippingMethod, err := client.ShippingMethods().WithId(shippingMethodID).Get().Execute(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	shippingRate, err := findShippingZoneRate(shippingZoneID, currencyCode, shippingMethod)

	if err != nil {
		return diag.FromErr(err)
	}

	input := platform.ShippingMethodUpdate{
		Version: shippingMethod.Version,
		Actions: []platform.ShippingMethodUpdateAction{},
	}

	if d.HasChange("price") || d.HasChange("free_above") || d.HasChange("shipping_rate_price_tier") {
		zoneResourceIdentifier := platform.ZoneResourceIdentifier{
			ID: &shippingZoneID,
		}

		oldTypedPrice := shippingRate.Price.(platform.CentPrecisionMoney)
		var oldFreeAboveMoney *platform.Money
		if shippingRate.FreeAbove != nil {
			oldFreeAbove := (shippingRate.FreeAbove).(platform.CentPrecisionMoney)
			oldFreeAboveMoney = &platform.Money{
				CurrencyCode: currencyCode,
				CentAmount:   oldFreeAbove.CentAmount,
			}
		}
		var oldShippingRatePriceTiers []platform.ShippingRatePriceTier
		if shippingRate.Tiers != nil {
			oldShippingRatePriceTiers = shippingRate.Tiers
		}

		oldShippingRateDraft := platform.ShippingRateDraft{
			Price: platform.Money{
				CurrencyCode: currencyCode,
				CentAmount:   oldTypedPrice.CentAmount,
			},
			FreeAbove: oldFreeAboveMoney,
			Tiers:     oldShippingRatePriceTiers,
		}

		price := d.Get("price").([]interface{})[0].(map[string]interface{})
		var newFreeAboveMoney *platform.Money
		if freeAbove, ok := d.GetOk("free_above"); ok {
			freeAboveMap := freeAbove.([]interface{})[0].(map[string]interface{})
			newFreeAboveMoney = &platform.Money{
				CurrencyCode: currencyCode,
				CentAmount:   freeAboveMap["cent_amount"].(int),
			}
		}

		newShippingRatePriceTiers, err := unmarshallShippingRatePriceTiers(d)
		if err != nil {
			return diag.FromErr(err)
		}

		newShippingRateDraft := platform.ShippingRateDraft{
			Price: platform.Money{
				CurrencyCode: currencyCode,
				CentAmount:   price["cent_amount"].(int),
			},
			FreeAbove: newFreeAboveMoney,
			Tiers:     newShippingRatePriceTiers,
		}

		input.Actions = append(
			input.Actions,
			&platform.ShippingMethodRemoveShippingRateAction{
				Zone: platform.ZoneResourceIdentifier{
					ID: &shippingZoneID,
				},
				ShippingRate: oldShippingRateDraft,
			})
		input.Actions = append(
			input.Actions,
			&platform.ShippingMethodAddShippingRateAction{
				Zone:         zoneResourceIdentifier,
				ShippingRate: newShippingRateDraft,
			})
	}

	log.Printf(
		"[DEBUG] Will perform update operation with the following actions:\n%s",
		stringFormatActions(input.Actions))

	_, err = client.ShippingMethods().WithId(shippingMethodID).Post(input).Execute(ctx)
	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
			log.Printf("[DEBUG] %v: %v", ctErr, stringFormatErrorExtras(ctErr))
		}
		return diag.FromErr(err)
	}

	return resourceShippingZoneRateRead(ctx, d, m)
}

func resourceShippingZoneRateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	shippingMethodID := d.Get("shipping_method_id").(string)
	ctMutexKV.Lock(shippingMethodID)
	defer ctMutexKV.Unlock(shippingMethodID)

	client := getClient(m)
	shippingMethod, err := client.ShippingMethods().WithId(shippingMethodID).Get().Execute(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	input := platform.ShippingMethodUpdate{
		Version: shippingMethod.Version,
		Actions: []platform.ShippingMethodUpdateAction{},
	}

	price := d.Get("price").([]interface{})[0].(map[string]interface{})
	var newFreeAboveMoney *platform.Money
	if freeAbove, ok := d.GetOk("free_above"); ok {
		freeAboveMap := freeAbove.([]interface{})[0].(map[string]interface{})
		newFreeAboveMoney = &platform.Money{
			CurrencyCode: freeAboveMap["currency_code"].(string),
			CentAmount:   freeAboveMap["cent_amount"].(int),
		}
	}

	newShippingRatePriceTiers, err := unmarshallShippingRatePriceTiers(d)
	if err != nil {
		return diag.FromErr(err)
	}

	shippingZoneID := d.Get("shipping_zone_id").(string)
	removeAction := platform.ShippingMethodRemoveShippingRateAction{
		Zone: platform.ZoneResourceIdentifier{ID: &shippingZoneID},
		ShippingRate: platform.ShippingRateDraft{
			Price: platform.Money{
				CurrencyCode: price["currency_code"].(string),
				CentAmount:   price["cent_amount"].(int),
			},
			FreeAbove: newFreeAboveMoney,
			Tiers:     newShippingRatePriceTiers,
		},
	}

	input.Actions = append(input.Actions, removeAction)

	for _, v := range shippingMethod.ZoneRates {
		if v.Zone.ID == shippingZoneID && len(v.ShippingRates) == 1 {
			input.Actions = append(input.Actions, platform.ShippingMethodRemoveZoneAction{
				Zone: platform.ZoneResourceIdentifier{ID: &shippingZoneID},
			})
			break
		}
	}

	_, err = client.ShippingMethods().WithId(shippingMethodID).Post(input).Execute(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getShippingIDs(shippingZoneRateID string) (string, string, string) {
	idSplit := strings.Split(shippingZoneRateID, "@")

	shippingMethodID := idSplit[0]
	shippingZoneID := idSplit[1]
	currencyCode := idSplit[2]

	return shippingMethodID, shippingZoneID, currencyCode
}

func findShippingZoneRate(shippingZoneID string, currencyCode string, shippingMethod *platform.ShippingMethod) (*platform.ShippingRate, error) {
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

	shippingRate, err := findShippingZoneRate(shippingZoneID, currencyCode, shippingMethod)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Found shipping rate: %s", stringFormatObject(shippingRate))

	if typedPrice, ok := shippingRate.Price.(platform.CentPrecisionMoney); ok {
		price := map[string]interface{}{
			"currency_code": string(typedPrice.CurrencyCode),
			"cent_amount":   typedPrice.CentAmount,
		}
		err = d.Set("price", []interface{}{price})
		if err != nil {
			return err
		}
	}

	if typedFreeAbove, ok := (shippingRate.FreeAbove).(platform.CentPrecisionMoney); ok {
		freeAbove := map[string]interface{}{
			"currency_code": string(typedFreeAbove.CurrencyCode),
			"cent_amount":   typedFreeAbove.CentAmount,
		}
		err = d.Set("free_above", []interface{}{freeAbove})
		if err != nil {
			return err
		}
	}
	log.Printf("[DEBUG] New state: %#v", d)

	return nil
}
