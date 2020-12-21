package commercetools

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

func resourceDiscountCode() *schema.Resource {
	return &schema.Resource{
		Create: resourceDiscountCodeCreate,
		Read:   resourceDiscountCodeRead,
		Update: resourceDiscountCodeUpdate,
		Delete: resourceDiscountCodeDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     TypeLocalizedString,
				Optional: true,
			},
			"description": {
				Type:     TypeLocalizedString,
				Optional: true,
			},
			"code": {
				Type:     schema.TypeString,
				Required: true,
			},
			"valid_from": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"valid_until": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"is_active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"predicate": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"max_applications_per_customer": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"max_applications": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"groups": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"cart_discounts": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceDiscountCodeCreate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	var discountCode *commercetools.DiscountCode

	name := commercetools.LocalizedString(
		expandStringMap(d.Get("name").(map[string]interface{})))
	description := commercetools.LocalizedString(
		expandStringMap(d.Get("description").(map[string]interface{})))

	draft := &commercetools.DiscountCodeDraft{
		Name:                       &name,
		Description:                &description,
		Code:                       d.Get("code").(string),
		CartPredicate:              d.Get("predicate").(string),
		IsActive:                   d.Get("is_active").(bool),
		MaxApplicationsPerCustomer: d.Get("max_applications_per_customer").(int),
		MaxApplications:            d.Get("max_applications").(int),
		Groups:                     resourceDiscountCodeGetGroups(d),
		CartDiscounts:              resourceDiscountCodeGetCartDiscounts(d),
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

		discountCode, err = client.DiscountCodeCreate(context.Background(), draft)

		if err != nil {
			return handleCommercetoolsError(err)
		}
		return nil
	})

	if errorResponse != nil {
		return errorResponse
	}

	if discountCode == nil {
		log.Fatal("No discount code created")
	}

	d.SetId(discountCode.ID)
	d.Set("version", discountCode.Version)

	return resourceDiscountCodeRead(d, m)
}

func resourceDiscountCodeRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] Reading discount code from commercetools, with discount code id: %s", d.Id())

	client := getClient(m)

	discountCode, err := client.DiscountCodeGetWithID(context.Background(), d.Id())

	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if discountCode == nil {
		log.Print("[DEBUG] No discount code found")
		d.SetId("")
	} else {
		log.Print("[DEBUG] Found following discount code:")
		log.Print(stringFormatObject(discountCode))

		d.Set("version", discountCode.Version)
		d.Set("code", discountCode.Code)
		d.Set("name", discountCode.Name)
		d.Set("description", discountCode.Description)
		d.Set("predicate", discountCode.CartPredicate)
		d.Set("cart_discounts", discountCode.CartDiscounts)
		d.Set("groups", discountCode.Groups)
		d.Set("is_active", discountCode.IsActive)
		d.Set("valid_from", discountCode.ValidFrom)
		d.Set("valid_until", discountCode.ValidUntil)
		d.Set("max_applications_per_customer", discountCode.MaxApplicationsPerCustomer)
		d.Set("max_applications", discountCode.MaxApplications)
	}

	return nil
}

func resourceDiscountCodeUpdate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	discountCode, err := client.DiscountCodeGetWithID(context.Background(), d.Id())
	if err != nil {
		return err
	}

	input := &commercetools.DiscountCodeUpdateWithIDInput{
		ID:      d.Id(),
		Version: discountCode.Version,
		Actions: []commercetools.DiscountCodeUpdateAction{},
	}

	if d.HasChange("name") {
		newName := commercetools.LocalizedString(
			expandStringMap(d.Get("name").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&commercetools.DiscountCodeSetNameAction{Name: &newName})
	}

	if d.HasChange("description") {
		newDescription := commercetools.LocalizedString(
			expandStringMap(d.Get("description").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&commercetools.DiscountCodeSetDescriptionAction{Description: &newDescription})
	}

	if d.HasChange("predicate") {
		newPredicate := d.Get("predicate").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.DiscountCodeSetCartPredicateAction{CartPredicate: newPredicate})
	}

	if d.HasChange("max_applications") {
		newMaxApplications := d.Get("max_applications").(int)
		input.Actions = append(
			input.Actions,
			&commercetools.DiscountCodeSetMaxApplicationsAction{MaxApplications: newMaxApplications})
	}

	if d.HasChange("max_applications_per_customer") {
		newMaxApplications := d.Get("max_applications_per_customer").(int)
		input.Actions = append(
			input.Actions,
			&commercetools.DiscountCodeSetMaxApplicationsPerCustomerAction{MaxApplicationsPerCustomer: newMaxApplications})
	}

	if d.HasChange("cart_discounts") {
		newCartDiscounts := resourceDiscountCodeGetCartDiscounts(d)
		input.Actions = append(
			input.Actions,
			&commercetools.DiscountCodeChangeCartDiscountsAction{CartDiscounts: newCartDiscounts})
	}

	if d.HasChange("groups") {
		newGroups := resourceDiscountCodeGetGroups(d)
		if len(newGroups) > 0 {
			input.Actions = append(
				input.Actions,
				&commercetools.DiscountCodeChangeGroupsAction{Groups: newGroups})
		} else {
			input.Actions = append(
				input.Actions,
				&commercetools.DiscountCodeChangeGroupsAction{Groups: []string{}})
		}
	}

	if d.HasChange("is_active") {
		newIsActive := d.Get("is_active").(bool)
		input.Actions = append(
			input.Actions,
			&commercetools.DiscountCodeChangeIsActiveAction{IsActive: newIsActive})
	}

	if d.HasChange("valid_from") {
		if val := d.Get("valid_from").(string); len(val) > 0 {
			newValidFrom, err := expandDate(d.Get("valid_from").(string))
			if err != nil {
				return err
			}
			input.Actions = append(
				input.Actions,
				&commercetools.DiscountCodeSetValidFromAction{ValidFrom: &newValidFrom})
		} else {
			input.Actions = append(
				input.Actions,
				&commercetools.DiscountCodeSetValidFromAction{})
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
				&commercetools.DiscountCodeSetValidUntilAction{ValidUntil: &newValidUntil})
		} else {
			input.Actions = append(
				input.Actions,
				&commercetools.DiscountCodeSetValidUntilAction{})
		}
	}

	log.Printf(
		"[DEBUG] Will perform update operation with the following actions:\n%s",
		stringFormatActions(input.Actions))

	_, err = client.DiscountCodeUpdateWithID(context.Background(), input)
	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			log.Printf("[DEBUG] %v: %v", ctErr, stringFormatErrorExtras(ctErr))
		}
		return err
	}

	return resourceDiscountCodeRead(d, m)
}

func resourceDiscountCodeDelete(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	version := d.Get("version").(int)
	_, err := client.DiscountCodeDeleteWithID(context.Background(), d.Id(), version, false)
	if err != nil {
		log.Printf("[ERROR] Error during deleting discount code resource %s", err)
		return nil
	}
	return nil
}

func resourceDiscountCodeGetGroups(d *schema.ResourceData) []string {
	var groups []string
	for _, group := range expandStringArray(d.Get("groups").([]interface{})) {
		groups = append(groups, group)
	}
	return groups
}

func resourceDiscountCodeGetCartDiscounts(d *schema.ResourceData) []commercetools.CartDiscountResourceIdentifier {
	var cartDiscounts []commercetools.CartDiscountResourceIdentifier
	for _, cartDiscount := range expandStringArray(d.Get("cart_discounts").([]interface{})) {
		cartDiscounts = append(cartDiscounts, commercetools.CartDiscountResourceIdentifier{ID: cartDiscount})
	}
	return cartDiscounts
}
