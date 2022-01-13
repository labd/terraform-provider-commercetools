package commercetools

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

func resourceShippingMethod() *schema.Resource {
	return &schema.Resource{
		Description: "With Shipping Methods you can specify which shipping services you want to provide to your " +
			"customers for deliveries to different areas of the world at rates you can define.\n\n" +
			"See also the [Shipping Methods API Documentation](https://docs.commercetools.com/api/projects/shippingMethods)",
		Create: resourceShippingMethodCreate,
		Read:   resourceShippingMethodRead,
		Update: resourceShippingMethodUpdate,
		Delete: resourceShippingMethodDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "User-specific unique identifier for the shipping method",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"localized_name": {
				Description: "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
				Type:        TypeLocalizedString,
				Optional:    true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"localized_description": {
				Description: "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
				Type:        TypeLocalizedString,
				Optional:    true,
			},
			"is_default": {
				Description: "One shipping method in a project can be default",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tax_category_id": {
				Description: "ID of a [Tax Category](https://docs.commercetools.com/api/projects/taxCategories#taxcategory)",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"predicate": {
				Description: "A Cart predicate which can be used to more precisely select a shipping method for a cart",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func resourceShippingMethodCreate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	var shippingMethod *commercetools.ShippingMethod
	taxCategory := commercetools.TaxCategoryResourceIdentifier{}
	if taxCategoryID, ok := d.GetOk("tax_category_id"); ok {
		taxCategory.ID = taxCategoryID.(string)
	}

	localizedDescription := commercetools.LocalizedString(
		expandStringMap(d.Get("localized_description").(map[string]interface{})))

	localizedName := commercetools.LocalizedString(
		expandStringMap(d.Get("localized_name").(map[string]interface{})))

	draft := &commercetools.ShippingMethodDraft{
		Key:                  d.Get("key").(string),
		Name:                 d.Get("name").(string),
		Description:          d.Get("description").(string),
		LocalizedDescription: &localizedDescription,
		LocalizedName:        &localizedName,
		IsDefault:            d.Get("is_default").(bool),
		TaxCategory:          &taxCategory,
		Predicate:            d.Get("predicate").(string),
	}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error

		shippingMethod, err = client.ShippingMethodCreate(context.Background(), draft)
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

	d.SetId(shippingMethod.ID)
	d.Set("version", shippingMethod.Version)

	return resourceShippingMethodRead(d, m)
}

func resourceShippingMethodRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] Reading shipping method from commercetools, with shippingMethod id: %s", d.Id())

	client := getClient(m)

	shippingMethod, err := client.ShippingMethodGetWithID(context.Background(), d.Id())

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

		d.Set("version", shippingMethod.Version)
		d.Set("key", shippingMethod.Key)
		d.Set("name", shippingMethod.Name)
		d.Set("localized_name", shippingMethod.LocalizedName)
		d.Set("description", shippingMethod.Description)
		d.Set("localized_description", shippingMethod.LocalizedDescription)
		d.Set("is_default", shippingMethod.IsDefault)
		d.Set("tax_category_id", shippingMethod.TaxCategory.ID)
		d.Set("predicate", shippingMethod.Predicate)
	}

	return nil
}

func resourceShippingMethodUpdate(d *schema.ResourceData, m interface{}) error {
	ctMutexKV.Lock(d.Id())
	defer ctMutexKV.Unlock(d.Id())

	client := getClient(m)
	shippingMethod, err := client.ShippingMethodGetWithID(context.Background(), d.Id())
	if err != nil {
		return err
	}

	input := &commercetools.ShippingMethodUpdateWithIDInput{
		ID:      d.Id(),
		Version: shippingMethod.Version,
		Actions: []commercetools.ShippingMethodUpdateAction{},
	}

	if d.HasChange("name") {
		newName := d.Get("name").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.ShippingMethodChangeNameAction{Name: newName})
	}

	if d.HasChange("localized_name") {
		newLocalizedName := commercetools.LocalizedString(
			expandStringMap(d.Get("localized_name").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&commercetools.ShippingMethodSetLocalizedNameAction{LocalizedName: &newLocalizedName})
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.ShippingMethodSetKeyAction{Key: newKey})
	}

	if d.HasChange("description") {
		newDescription := d.Get("description").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.ShippingMethodSetDescriptionAction{Description: newDescription})
	}

	if d.HasChange("localized_description") {
		newLocalizedDescription := commercetools.LocalizedString(
			expandStringMap(d.Get("localized_description").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&commercetools.ShippingMethodSetLocalizedDescriptionAction{LocalizedDescription: &newLocalizedDescription})
	}

	if d.HasChange("is_default") {
		newIsDefault := d.Get("is_default").(bool)
		input.Actions = append(
			input.Actions,
			&commercetools.ShippingMethodChangeIsDefaultAction{IsDefault: newIsDefault})
	}

	if d.HasChange("tax_category_id") {
		taxCategoryID := d.Get("tax_category_id").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.ShippingMethodChangeTaxCategoryAction{TaxCategory: &commercetools.TaxCategoryResourceIdentifier{ID: taxCategoryID}})
	}

	if d.HasChange("predicate") {
		newPredicate := d.Get("predicate").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.ShippingMethodSetPredicateAction{Predicate: newPredicate})
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

	return resourceShippingMethodRead(d, m)
}

func resourceShippingMethodDelete(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)

	ctMutexKV.Lock(d.Id())
	defer ctMutexKV.Unlock(d.Id())

	shippingMethod, err := client.ShippingMethodGetWithID(context.Background(), d.Id())
	if err != nil {
		return err
	}

	_, err = client.ShippingMethodDeleteWithID(context.Background(), d.Id(), shippingMethod.Version)
	if err != nil {
		return err
	}

	return nil
}
