package commercetools

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

func resourceShippingZoneRate() *schema.Resource {
	return &schema.Resource{
		Create: resourceShippingZoneRateCreate,
		Read:   resourceShippingZoneRateRead,
		Update: resourceShippingZoneRateUpdate,
		Delete: resourceShippingZoneRateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"shipping_method_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"shipping_zone_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"price": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"currency_code": {
							Type:     schema.TypeString,
							Required: true,
						},
						"cent_amount": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"free_above": {
				Type:     schema.TypeList,
				MinItems: 1,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"currency_code": {
							Type:     schema.TypeString,
							Required: true,
						},
						"cent_amount": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceShippingZoneRateCreate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	var zoneRate *commercetools.ZoneRate
	shippingZoneID := d.Get("shipping_zone_id").(string)
	shippingMethodID := d.Get("shipping_method_id").(string)

	// Lock to prevent concurrent updates due to Version number conflicts
	ctMutexKV.Lock(shippingMethodID)
	defer ctMutexKV.Unlock(shippingMethodID)

	shippingMethod, err := client.ShippingMethodGetByID(shippingMethodID)

	if err != nil {
		return err
	}

	input := commercetools.ShippingMethodUpdateInput{
		ID:      shippingMethod.ID,
		Version: shippingMethod.Version,
		Actions: []commercetools.ShippingMethodUpdateAction{},
	}
	price := d.Get("price").([]interface{})[0].(map[string]interface{})
	freeAbove := d.Get("free_above").([]interface{})[0].(map[string]interface{})
	input.Actions = append(input.Actions, commercetools.ShippingMethodAddShippingRateAction{
		Zone: &commercetools.ZoneReference{ID: shippingZoneID},
		ShippingRate: &commercetools.ShippingRateDraft{
			Price: &commercetools.Money{
				CurrencyCode: price["currency_code"].(commercetools.CurrencyCode),
				CentAmount:   price["cent_amount"].(int),
			},
			FreeAbove: &commercetools.Money{
				CurrencyCode: freeAbove["currency_code"].(commercetools.CurrencyCode),
				CentAmount:   freeAbove["cent_amount"].(int),
			},
		},
	})

	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		shippingMethod, err = client.ShippingMethodUpdateByID(&input)
		if err != nil {
			if ctErr, ok := err.(commercetools.ErrorResponse); ok {
				if _, ok := ctErr.Errors[0].(commercetools.InvalidJSONInputError); ok {
					return resource.NonRetryableError(ctErr)
				}
			} else {
				log.Printf("[DEBUG] Received error: %s", err)
			}
			return resource.RetryableError(err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	if shippingMethod == nil {
		log.Fatal("No shipping method created?")
	}

	newID := shippingMethod.ID + "@" + shippingZoneID
	d.SetId(newID)

	return resourceShippingZoneRateRead(d, m)
}

func resourceShippingZoneRateRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] Reading shipping zone rate from commercetools, with id: %s", d.Id())

	shippingMethodID, shippingZoneID, _ := getShippingIDs(d.Id())

	client := getClient(m)

	shippingMethod, err := client.ShippingMethodGetByID(shippingMethodID)

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
	shippingMethodID := d.Get("shipping_method_id").(string)
	ctMutexKV.Lock(shippingMethodID)
	defer ctMutexKV.Unlock(shippingMethodID)

	client := getClient(m)
	shippingMethod, err := client.ShippingMethodGetByID(shippingMethodID)
	if err != nil {
		return err
	}

	input := &commercetools.ShippingMethodUpdateInput{
		ID:      d.Id(),
		Version: shippingMethod.Version,
		Actions: []commercetools.ShippingMethodUpdateAction{},
	}

	if d.HasChange("price") || d.HasChange("free_above") {
		// TODO: figure out if you can remove/add at the same time (probably not)
		input.Actions = append(
			input.Actions,
			&commercetools.ShippingMethodRemoveShippingRateAction{})
		input.Actions = append(
			input.Actions,
			&commercetools.ShippingMethodAddShippingRateAction{})

	}

	log.Printf(
		"[DEBUG] Will perform update operation with the following actions:\n%s",
		stringFormatActions(input.Actions))

	_, err = client.ShippingMethodUpdateByID(input)
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
	shippingMethod, err := client.ShippingMethodGetByID(shippingMethodID)
	if err != nil {
		return err
	}

	input := &commercetools.ShippingMethodUpdateInput{
		ID:      shippingMethod.ID,
		Version: shippingMethod.Version,
		Actions: []commercetools.ShippingMethodUpdateAction{},
	}

	price := d.Get("price").([]interface{})[0].(map[string]interface{})
	freeAbove := d.Get("free_above").([]interface{})[0].(map[string]interface{})

	removeAction := commercetools.ShippingMethodRemoveShippingRateAction{
		Zone: &commercetools.ZoneReference{
			ID: d.Get("shipping_zone_id").(string),
		},
		ShippingRate: &commercetools.ShippingRateDraft{
			Price: &commercetools.Money{
				CurrencyCode: price["currency_code"].(commercetools.CurrencyCode),
				CentAmount:   price["cent_amount"].(int),
			},
			FreeAbove: &commercetools.Money{
				CurrencyCode: freeAbove["currency_code"].(commercetools.CurrencyCode),
				CentAmount:   freeAbove["cent_amount"].(int),
			},
		},
	}

	_, err = client.ShippingMethodUpdateByID(input)
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
	for idx, zoneRate := range shippingMethod.ZoneRates {
		if zoneRate.Zone.ID == shippingZoneID {
			for idx, shippingRate := range zoneRate.ShippingRates {
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
	typedPrice := shippingRate.Price.(commercetools.CentPrecisionMoney)
	typedFreeAbove := shippingRate.FreeAbove.(commercetools.CentPrecisionMoney)

	price := map[string]interface{}{
		"currency_code": string(typedPrice.CurrencyCode),
		"cent_amount":   typedPrice.CentAmount,
	}

	freeAbove := map[string]interface{}{
		"currency_code": string(typedFreeAbove.CurrencyCode),
		"cent_amount":   typedFreeAbove.CentAmount,
	}

	d.Set("price", []interface{}{price})
	d.Set("free_above", []interface{}{freeAbove})

	return nil
}
