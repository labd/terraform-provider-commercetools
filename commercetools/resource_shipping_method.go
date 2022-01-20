package commercetools

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
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
	var shippingMethod *platform.ShippingMethod
	taxCategory := platform.TaxCategoryResourceIdentifier{}
	if taxCategoryID, ok := d.GetOk("tax_category_id"); ok {
		taxCategory.ID = stringRef(taxCategoryID)
	}

	localizedDescription := platform.LocalizedString(
		expandStringMap(d.Get("localized_description").(map[string]interface{})))

	localizedName := platform.LocalizedString(
		expandStringMap(d.Get("localized_name").(map[string]interface{})))

	draft := platform.ShippingMethodDraft{
		Key:                  stringRef(d.Get("key")),
		Name:                 d.Get("name").(string),
		Description:          stringRef(d.Get("description")),
		LocalizedDescription: &localizedDescription,
		LocalizedName:        &localizedName,
		IsDefault:            d.Get("is_default").(bool),
		TaxCategory:          taxCategory,
		Predicate:            stringRef(d.Get("predicate")),
	}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error

		shippingMethod, err = client.ShippingMethods().Post(draft).Execute(context.Background())
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

	shippingMethod, err := client.ShippingMethods().WithId(d.Id()).Get().Execute(context.Background())

	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
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
	shippingMethod, err := client.ShippingMethods().WithId(d.Id()).Get().Execute(context.Background())
	if err != nil {
		return err
	}

	input := platform.ShippingMethodUpdate{
		Version: shippingMethod.Version,
		Actions: []platform.ShippingMethodUpdateAction{},
	}

	if d.HasChange("name") {
		newName := d.Get("name").(string)
		input.Actions = append(
			input.Actions,
			&platform.ShippingMethodChangeNameAction{Name: newName})
	}

	if d.HasChange("localized_name") {
		newLocalizedName := platform.LocalizedString(
			expandStringMap(d.Get("localized_name").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&platform.ShippingMethodSetLocalizedNameAction{LocalizedName: &newLocalizedName})
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&platform.ShippingMethodSetKeyAction{Key: &newKey})
	}

	if d.HasChange("description") {
		newDescription := d.Get("description").(string)
		input.Actions = append(
			input.Actions,
			&platform.ShippingMethodSetDescriptionAction{Description: &newDescription})
	}

	if d.HasChange("localized_description") {
		newLocalizedDescription := platform.LocalizedString(
			expandStringMap(d.Get("localized_description").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&platform.ShippingMethodSetLocalizedDescriptionAction{LocalizedDescription: &newLocalizedDescription})
	}

	if d.HasChange("is_default") {
		newIsDefault := d.Get("is_default").(bool)
		input.Actions = append(
			input.Actions,
			&platform.ShippingMethodChangeIsDefaultAction{IsDefault: newIsDefault})
	}

	if d.HasChange("tax_category_id") {
		taxCategoryID := d.Get("tax_category_id").(string)
		input.Actions = append(
			input.Actions,
			&platform.ShippingMethodChangeTaxCategoryAction{TaxCategory: platform.TaxCategoryResourceIdentifier{ID: &taxCategoryID}})
	}

	if d.HasChange("predicate") {
		newPredicate := d.Get("predicate").(string)
		input.Actions = append(
			input.Actions,
			&platform.ShippingMethodSetPredicateAction{Predicate: &newPredicate})
	}

	log.Printf(
		"[DEBUG] Will perform update operation with the following actions:\n%s",
		stringFormatActions(input.Actions))

	_, err = client.ShippingMethods().WithId(shippingMethod.ID).Post(input).Execute(context.Background())
	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
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

	shippingMethod, err := client.ShippingMethods().WithId(d.Id()).Get().Execute(context.Background())
	if err != nil {
		return err
	}

	_, err = client.ShippingMethods().WithId(d.Id()).Delete().WithQueryParams(platform.ByProjectKeyShippingMethodsByIDRequestMethodDeleteInput{
		Version: shippingMethod.Version,
	}).Execute(context.Background())
	if err != nil {
		return err
	}

	return nil
}
