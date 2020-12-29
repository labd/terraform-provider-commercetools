package commercetools

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

func resourceShippingZoneRate() *schema.Resource {
	return &schema.Resource{
		Description: "Defines shipping rates (prices) for a specific zone.\n\n" +
			"See also [ZoneRate API Documentation](https://docs.commercetools.com/api/projects/shippingMethods#zonerate)",
		Create: resourceShippingZoneRateCreate,
		Read:   resourceShippingZoneRateRead,
		Update: resourceShippingZoneRateUpdate,
		Delete: resourceShippingZoneRateDelete,
		Importer: &schema.ResourceImporter{
			State: resourceShippingZoneRateImportState,
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
								string(commercetools.ShippingRateTierTypeCartValue),
								string(commercetools.ShippingRateTierTypeCartScore),
								string(commercetools.ShippingRateTierTypeCartClassification),
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

func resourceShippingZoneRateImportState(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := getClient(meta)
	shippingMethodID, _, _ := getShippingIDs(d.Id())

	shippingMethod, err := client.ShippingMethodGetWithID(context.Background(), shippingMethodID)

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

func resourceShippingZoneRateCreate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	shippingZoneID := d.Get("shipping_zone_id").(string)
	shippingMethodID := d.Get("shipping_method_id").(string)

	// Lock to prevent concurrent updates due to Version number conflicts
	ctMutexKV.Lock(shippingMethodID)
	defer ctMutexKV.Unlock(shippingMethodID)

	shippingMethod, err := client.ShippingMethodGetWithID(context.Background(), shippingMethodID)

	if err != nil {
		return err
	}

	input := commercetools.ShippingMethodUpdateWithIDInput{
		ID:      shippingMethod.ID,
		Version: shippingMethod.Version,
		Actions: []commercetools.ShippingMethodUpdateAction{},
	}
	price := d.Get("price").([]interface{})[0].(map[string]interface{})

	var freeAbove *commercetools.Money
	if freeAboveState, ok := d.GetOk("free_above"); ok {
		log.Printf("[DEBUG] Free above state: %s", stringFormatObject(freeAboveState))
		freeAboveMap := freeAboveState.([]interface{})[0].(map[string]interface{})
		freeAbove = &commercetools.Money{
			CurrencyCode: commercetools.CurrencyCode(freeAboveMap["currency_code"].(string)),
			CentAmount:   freeAboveMap["cent_amount"].(int),
		}
	}
	log.Printf("[DEBUG] Setting freeAbove: %s", stringFormatObject(freeAbove))

	var shippingRatePriceTiers []commercetools.ShippingRatePriceTier
	if shippingRatePriceTierState, ok := d.GetOk("shipping_rate_price_tier"); ok {
		shippingRatePriceTiers, err = createShippingRatePriceTiers(shippingRatePriceTierState.([]interface{}))
		if err != nil {
			return err
		}
	}
	log.Printf("[DEBUG] Setting shippingRatePriceTiers: %s", stringFormatObject(shippingRatePriceTiers))

	priceCurrencyCode := commercetools.CurrencyCode(price["currency_code"].(string))

	zoneNotFound := true
	for _, v := range shippingMethod.ZoneRates {
		if v.Zone.ID == shippingZoneID {
			zoneNotFound = false
			break
		}
	}

	if zoneNotFound {
		input.Actions = append(input.Actions, commercetools.ShippingMethodAddZoneAction{
			Zone: &commercetools.ZoneResourceIdentifier{ID: shippingZoneID},
		})
	}

	input.Actions = append(input.Actions, commercetools.ShippingMethodAddShippingRateAction{
		Zone: &commercetools.ZoneResourceIdentifier{ID: shippingZoneID},
		ShippingRate: &commercetools.ShippingRateDraft{
			Price: &commercetools.Money{
				CurrencyCode: priceCurrencyCode,
				CentAmount:   price["cent_amount"].(int),
			},
			FreeAbove: freeAbove,
			Tiers:     shippingRatePriceTiers,
		},
	})

	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error

		shippingMethod, err = client.ShippingMethodUpdateWithID(context.Background(), &input)
		if err != nil {
			return handleCommercetoolsError(err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	if shippingMethod == nil {
		log.Fatal("No shipping method created?")
	}

	d.SetId(buildShippingZoneRateID(shippingMethod.ID, shippingZoneID, string(priceCurrencyCode)))

	return resourceShippingZoneRateRead(d, m)
}

func createShippingRatePriceTiers(tierStateMap []interface{}) ([]commercetools.ShippingRatePriceTier, error) {
	var tiers []commercetools.ShippingRatePriceTier
	for tierState := range tierStateMap {
		tierMap := tierStateMap[tierState].(map[string]interface{})

		var price *commercetools.Money
		priceMap := tierMap["price"].([]interface{})[0].(map[string]interface{})
		price = &commercetools.Money{
			CurrencyCode: commercetools.CurrencyCode(priceMap["currency_code"].(string)),
			CentAmount:   priceMap["cent_amount"].(int),
		}

		tierType := tierMap["type"].(string)
		switch tierType {
		case string(commercetools.ShippingRateTierTypeCartValue):
			tiers = append(tiers, commercetools.CartValueTier{
				MinimumCentAmount: tierMap["minimum_cent_amount"].(int),
				Price:             price,
			})
		case string(commercetools.ShippingRateTierTypeCartClassification):
			tiers = append(tiers, commercetools.CartClassificationTier{
				Value: tierMap["value"].(string),
				Price: price,
			})
		case string(commercetools.ShippingRateTierTypeCartScore):
			tiers = append(tiers, commercetools.CartScoreTier{
				Score: tierMap["score"].(float64),
				Price: price,
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

func resourceShippingZoneRateRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] Reading shipping zone rate from commercetools, with id: %s", d.Id())

	shippingMethodID, _, _ := getShippingIDs(d.Id())

	client := getClient(m)

	shippingMethod, err := client.ShippingMethodGetWithID(context.Background(), shippingMethodID)

	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if shippingMethod == nil {
		log.Print("[DEBUG] No shipping method found")
		d.SetId("")
	} else {
		log.Print("[DEBUG] Found following shipping method:")
		log.Print(stringFormatObject(shippingMethod))

		err = setShippingZoneRateState(d, shippingMethod)
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceShippingZoneRateUpdate(d *schema.ResourceData, m interface{}) error {
	shippingMethodID, shippingZoneID, currencyCode := getShippingIDs(d.Id())
	ctMutexKV.Lock(shippingMethodID)
	defer ctMutexKV.Unlock(shippingMethodID)

	client := getClient(m)
	shippingMethod, err := client.ShippingMethodGetWithID(context.Background(), shippingMethodID)
	if err != nil {
		return err
	}

	shippingRate, err := findShippingZoneRate(shippingZoneID, currencyCode, shippingMethod)

	if err != nil {
		return err
	}

	input := &commercetools.ShippingMethodUpdateWithIDInput{
		ID:      shippingMethodID,
		Version: shippingMethod.Version,
		Actions: []commercetools.ShippingMethodUpdateAction{},
	}

	if d.HasChange("price") || d.HasChange("free_above") || d.HasChange("shipping_rate_price_tier") {
		zoneResourceIdentifier := commercetools.ZoneResourceIdentifier{
			ID: shippingZoneID,
		}

		oldTypedPrice := shippingRate.Price.(commercetools.CentPrecisionMoney)
		var oldFreeAboveMoney *commercetools.Money
		if shippingRate.FreeAbove != nil {
			oldFreeAbove := shippingRate.FreeAbove.(commercetools.CentPrecisionMoney)
			oldFreeAboveMoney = &commercetools.Money{
				CurrencyCode: commercetools.CurrencyCode(currencyCode),
				CentAmount:   oldFreeAbove.CentAmount,
			}
		}
		var oldShippingRatePriceTiers []commercetools.ShippingRatePriceTier
		if shippingRate.Tiers != nil {
			oldShippingRatePriceTiers = shippingRate.Tiers
		}

		oldShippingRateDraft := commercetools.ShippingRateDraft{
			Price: &commercetools.Money{
				CurrencyCode: commercetools.CurrencyCode(currencyCode),
				CentAmount:   oldTypedPrice.CentAmount,
			},
			FreeAbove: oldFreeAboveMoney,
			Tiers:     oldShippingRatePriceTiers,
		}

		price := d.Get("price").([]interface{})[0].(map[string]interface{})
		var newFreeAboveMoney *commercetools.Money
		if freeAbove, ok := d.GetOk("free_above"); ok {
			freeAboveMap := freeAbove.([]interface{})[0].(map[string]interface{})
			newFreeAboveMoney = &commercetools.Money{
				CurrencyCode: commercetools.CurrencyCode(currencyCode),
				CentAmount:   freeAboveMap["cent_amount"].(int),
			}
		}

		var newShippingRatePriceTiers []commercetools.ShippingRatePriceTier
		if shippingRatePriceTiers, ok := d.GetOk("shipping_rate_price_tier"); ok {
			newShippingRatePriceTiers, err = createShippingRatePriceTiers(shippingRatePriceTiers.([]interface{}))
			if err != nil {
				return err
			}
		}

		newShippingRateDraft := commercetools.ShippingRateDraft{
			Price: &commercetools.Money{
				CurrencyCode: commercetools.CurrencyCode(currencyCode),
				CentAmount:   price["cent_amount"].(int),
			},
			FreeAbove: newFreeAboveMoney,
			Tiers:     newShippingRatePriceTiers,
		}

		input.Actions = append(
			input.Actions,
			&commercetools.ShippingMethodRemoveShippingRateAction{
				Zone: &commercetools.ZoneResourceIdentifier{
					ID: shippingZoneID,
				},
				ShippingRate: &oldShippingRateDraft,
			})
		input.Actions = append(
			input.Actions,
			&commercetools.ShippingMethodAddShippingRateAction{
				Zone:         &zoneResourceIdentifier,
				ShippingRate: &newShippingRateDraft,
			})
	}

	log.Printf(
		"[DEBUG] Will perform update operation with the following actions:\n%s",
		stringFormatActions(input.Actions))

	_, err = client.ShippingMethodUpdateWithID(context.Background(), input)
	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			log.Printf("[DEBUG] %v: %v", ctErr, stringFormatErrorExtras(ctErr))
		}
		return err
	}

	return resourceShippingZoneRateRead(d, m)
}

func resourceShippingZoneRateDelete(d *schema.ResourceData, m interface{}) error {
	shippingMethodID := d.Get("shipping_method_id").(string)
	ctMutexKV.Lock(shippingMethodID)
	defer ctMutexKV.Unlock(shippingMethodID)

	client := getClient(m)
	shippingMethod, err := client.ShippingMethodGetWithID(context.Background(), shippingMethodID)
	if err != nil {
		return err
	}

	input := &commercetools.ShippingMethodUpdateWithIDInput{
		ID:      shippingMethod.ID,
		Version: shippingMethod.Version,
		Actions: []commercetools.ShippingMethodUpdateAction{},
	}

	price := d.Get("price").([]interface{})[0].(map[string]interface{})
	var newFreeAboveMoney *commercetools.Money
	if freeAbove, ok := d.GetOk("free_above"); ok {
		freeAboveMap := freeAbove.([]interface{})[0].(map[string]interface{})
		newFreeAboveMoney = &commercetools.Money{
			CurrencyCode: commercetools.CurrencyCode(freeAboveMap["currency_code"].(string)),
			CentAmount:   freeAboveMap["cent_amount"].(int),
		}
	}

	var newShippingRatePriceTiers []commercetools.ShippingRatePriceTier
	if shippingRatePriceTiers, ok := d.GetOk("shipping_rate_price_tier"); ok {
		newShippingRatePriceTiers, err = createShippingRatePriceTiers(shippingRatePriceTiers.([]interface{}))
		if err != nil {
			return err
		}
	}

	shippingZoneID := d.Get("shipping_zone_id").(string)
	removeAction := commercetools.ShippingMethodRemoveShippingRateAction{
		Zone: &commercetools.ZoneResourceIdentifier{ID: shippingZoneID},
		ShippingRate: &commercetools.ShippingRateDraft{
			Price: &commercetools.Money{
				CurrencyCode: commercetools.CurrencyCode(price["currency_code"].(string)),
				CentAmount:   price["cent_amount"].(int),
			},
			FreeAbove: newFreeAboveMoney,
			Tiers:     newShippingRatePriceTiers,
		},
	}

	input.Actions = append(input.Actions, removeAction)

	for _, v := range shippingMethod.ZoneRates {
		if v.Zone.ID == shippingZoneID && len(v.ShippingRates) == 1 {
			input.Actions = append(input.Actions, commercetools.ShippingMethodRemoveZoneAction{
				Zone: &commercetools.ZoneResourceIdentifier{ID: shippingZoneID},
			})
			break
		}
	}

	_, err = client.ShippingMethodUpdateWithID(context.Background(), input)
	if err != nil {
		return err
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

func findShippingZoneRate(shippingZoneID string, currencyCode string, shippingMethod *commercetools.ShippingMethod) (*commercetools.ShippingRate, error) {
	for _, zoneRate := range shippingMethod.ZoneRates {
		if zoneRate.Zone.ID == shippingZoneID {
			for _, shippingRate := range zoneRate.ShippingRates {
				if shippingRate.Price.(commercetools.CentPrecisionMoney).CurrencyCode == commercetools.CurrencyCode(currencyCode) {
					return &shippingRate, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("Couldn't find shipping zone rate")
}

func setShippingZoneRateState(d *schema.ResourceData, shippingMethod *commercetools.ShippingMethod) error {
	shippingMethodID, shippingZoneID, currencyCode := getShippingIDs(d.Id())

	d.Set("shipping_method_id", shippingMethodID)
	d.Set("shipping_zone_id", shippingZoneID)

	shippingRate, err := findShippingZoneRate(shippingZoneID, currencyCode, shippingMethod)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Found shipping rate: %s", stringFormatObject(shippingRate))

	if typedPrice, ok := shippingRate.Price.(commercetools.CentPrecisionMoney); ok {
		price := map[string]interface{}{
			"currency_code": string(typedPrice.CurrencyCode),
			"cent_amount":   typedPrice.CentAmount,
		}
		err = d.Set("price", []interface{}{price})
		if err != nil {
			return err
		}
	}

	if typedFreeAbove, ok := shippingRate.FreeAbove.(commercetools.CentPrecisionMoney); ok {
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
