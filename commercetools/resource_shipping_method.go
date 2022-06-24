package commercetools

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
)

func resourceShippingMethod() *schema.Resource {
	return &schema.Resource{
		Description: "With Shipping Methods you can specify which shipping services you want to provide to your " +
			"customers for deliveries to different areas of the world at rates you can define.\n\n" +
			"See also the [Shipping Methods API Documentation](https://docs.commercetoolstools.com/api/projects/shippingMethods)",
		CreateContext: resourceShippingMethodCreate,
		ReadContext:   resourceShippingMethodRead,
		UpdateContext: resourceShippingMethodUpdate,
		DeleteContext: resourceShippingMethodDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"localized_description": {
				Description:      "[LocalizedString](https://docs.commercetoolstools.com/api/types#localizedstring)",
				Type:             TypeLocalizedString,
				ValidateDiagFunc: validateLocalizedStringKey,
				Optional:         true,
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
				Description: "ID of a [Tax Category](https://docs.commercetoolstools.com/api/projects/taxCategories#taxcategory)",
				Type:        schema.TypeString,
				Required:    true,
			},
			"predicate": {
				Description: "A Cart predicate which can be used to more precisely select a shipping method for a cart",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"custom": CustomFieldSchema(),
		},
	}
}

func resourceShippingMethodCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)
	taxCategory := platform.TaxCategoryResourceIdentifier{}
	if taxCategoryID, ok := d.GetOk("tax_category_id"); ok {
		taxCategory.ID = stringRef(taxCategoryID)
	}

	localizedDescription := expandLocalizedString(d.Get("localized_description"))

	draft := platform.ShippingMethodDraft{
		Name:                 d.Get("name").(string),
		Description:          stringRef(d.Get("description")),
		LocalizedDescription: &localizedDescription,
		IsDefault:            d.Get("is_default").(bool),
		TaxCategory:          taxCategory,
		Predicate:            nilIfEmpty(stringRef(d.Get("predicate"))),
		Custom:               CreateCustomFieldDraft(d),
	}

	key := stringRef(d.Get("key"))
	if *key != "" {
		draft.Key = key
	}

	var shippingMethod *platform.ShippingMethod
	err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		var err error
		shippingMethod, err = client.ShippingMethods().Post(draft).Execute(ctx)
		return processRemoteError(err)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(shippingMethod.ID)
	d.Set("version", shippingMethod.Version)

	return resourceShippingMethodRead(ctx, d, m)
}

func resourceShippingMethodRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)
	shippingMethod, err := client.ShippingMethods().WithId(d.Id()).Get().Execute(ctx)
	if err != nil {
		if IsResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if shippingMethod == nil {
		d.SetId("")
	} else {
		d.Set("version", shippingMethod.Version)
		d.Set("key", shippingMethod.Key)
		d.Set("name", shippingMethod.Name)
		d.Set("description", shippingMethod.Description)
		d.Set("localized_description", shippingMethod.LocalizedDescription)
		d.Set("is_default", shippingMethod.IsDefault)
		d.Set("tax_category_id", shippingMethod.TaxCategory.ID)
		d.Set("predicate", shippingMethod.Predicate)
		d.Set("custom", flattenCustomFields(shippingMethod.Custom))
	}

	return nil
}

func resourceShippingMethodUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ctMutexKV.Lock(d.Id())
	defer ctMutexKV.Unlock(d.Id())

	client := getClient(m)
	shippingMethod, err := client.ShippingMethods().WithId(d.Id()).Get().Execute(ctx)
	if err != nil {
		return diag.FromErr(err)
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
		newLocalizedDescription := expandLocalizedString(d.Get("localized_description"))
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
		newPredicate := nilIfEmpty(stringRef(d.Get("predicate").(string)))
		input.Actions = append(
			input.Actions,
			&platform.ShippingMethodSetPredicateAction{Predicate: newPredicate})
	}

	if d.HasChange("custom") {
		actions, err := CustomFieldUpdateActions[platform.ShippingMethodSetCustomTypeAction, platform.ShippingMethodSetCustomFieldAction](d)
		if err != nil {
			return diag.FromErr(err)
		}
		for i := range actions {
			input.Actions = append(input.Actions, actions[i].(platform.ShippingMethodUpdateAction))
		}
	}

	err = resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		_, err := client.ShippingMethods().WithId(shippingMethod.ID).Post(input).Execute(ctx)
		return processRemoteError(err)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceShippingMethodRead(ctx, d, m)
}

func resourceShippingMethodDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)

	ctMutexKV.Lock(d.Id())
	defer ctMutexKV.Unlock(d.Id())

	shippingMethod, err := client.ShippingMethods().WithId(d.Id()).Get().Execute(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	err = resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		_, err := client.ShippingMethods().WithId(d.Id()).Delete().Version(shippingMethod.Version).Execute(ctx)
		return processRemoteError(err)
	})
	return diag.FromErr(err)
}
