package commercetools

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func resourceDiscountCode() *schema.Resource {
	return &schema.Resource{
		Description: "With discount codes it is possible to give specific cart discounts to an eligible set of users. " +
			"They are defined by a string value which can be added to a cart so that specific cart discounts " +
			"can be applied to the cart.\n\n" +
			"See also the [Discount Code Api Documentation](https://docs.commercetools.com/api/projects/discountCodes)",
		CreateContext: resourceDiscountCodeCreate,
		ReadContext:   resourceDiscountCodeRead,
		UpdateContext: resourceDiscountCodeUpdate,
		DeleteContext: resourceDiscountCodeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description:      "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
				Type:             TypeLocalizedString,
				ValidateDiagFunc: validateLocalizedStringKey,
				Optional:         true,
			},
			"description": {
				Description:      "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
				Type:             TypeLocalizedString,
				ValidateDiagFunc: validateLocalizedStringKey,
				Optional:         true,
			},
			"code": {
				Description: "Unique identifier of this discount code. This value is added to the cart to enable " +
					"the related cart discounts in the cart",
				Type:     schema.TypeString,
				Required: true,
			},
			"valid_from": {
				Description:      "The time from which the discount can be applied on a cart. Before that time the code is invalid",
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: diffSuppressDateString,
			},
			"valid_until": {
				Description:      "The time until the discount can be applied on a cart. After that time the code is invalid",
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: diffSuppressDateString,
			},
			"is_active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"predicate": {
				Description: "[Cart Predicate](https://docs.commercetools.com/api/projects/predicates#cart-predicates)",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"max_applications_per_customer": {
				Description: "The discount code can only be applied maxApplicationsPerCustomer times per customer",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"max_applications": {
				Description: "The discount code can only be applied maxApplications times",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"groups": {
				Description: "The groups to which this discount code belong",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"cart_discounts": {
				Description: "The referenced matching cart discounts can be applied to the cart once the DiscountCode is added",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"custom": CustomFieldSchema(),
		},
	}
}

func resourceDiscountCodeCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	name := expandLocalizedString(d.Get("name"))
	description := expandLocalizedString(d.Get("description"))

	custom, err := CreateCustomFieldDraft(ctx, client, d)
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}
	cartPredicate := stringRef(d.Get("predicate"))
	if cartPredicate != nil && *cartPredicate == "" {
		cartPredicate = nil
	}

	draft := platform.DiscountCodeDraft{
		Name:                       &name,
		Description:                &description,
		Code:                       d.Get("code").(string),
		CartPredicate:              cartPredicate,
		IsActive:                   boolRef(d.Get("is_active")),
		MaxApplicationsPerCustomer: intRef(d.Get("max_applications_per_customer")),
		MaxApplications:            intRef(d.Get("max_applications")),
		Groups:                     expandDiscountCodeGroups(d),
		CartDiscounts:              expandDiscountCodeCartDiscounts(d),
		Custom:                     custom,
	}

	if val := d.Get("valid_from").(string); len(val) > 0 {
		validFrom, err := expandTime(val)
		if err != nil {
			return diag.FromErr(err)
		}
		draft.ValidFrom = &validFrom
	}
	if val := d.Get("valid_until").(string); len(val) > 0 {
		validUntil, err := expandTime(val)
		if err != nil {
			return diag.FromErr(err)
		}
		draft.ValidUntil = &validUntil
	}

	var discountCode *platform.DiscountCode
	err = resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		var err error
		discountCode, err = client.DiscountCodes().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(discountCode.ID)
	d.Set("version", discountCode.Version)

	return resourceDiscountCodeRead(ctx, d, m)
}

func resourceDiscountCodeRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	discountCode, err := client.DiscountCodes().WithId(d.Id()).Get().Execute(ctx)
	if err != nil {
		if utils.IsResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set("version", discountCode.Version)
	d.Set("code", discountCode.Code)
	d.Set("name", discountCode.Name)
	d.Set("description", discountCode.Description)
	d.Set("predicate", discountCode.CartPredicate)
	d.Set("cart_discounts", flattenDiscountCodeCartDiscounts(discountCode.CartDiscounts))
	d.Set("groups", discountCode.Groups)
	d.Set("is_active", discountCode.IsActive)
	d.Set("valid_from", flattenTime(discountCode.ValidFrom))
	d.Set("valid_until", flattenTime(discountCode.ValidUntil))
	d.Set("max_applications_per_customer", discountCode.MaxApplicationsPerCustomer)
	d.Set("max_applications", discountCode.MaxApplications)
	d.Set("custom", flattenCustomFields(discountCode.Custom))
	return nil
}

func resourceDiscountCodeUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	input := platform.DiscountCodeUpdate{
		Version: d.Get("version").(int),
		Actions: []platform.DiscountCodeUpdateAction{},
	}

	if d.HasChange("name") {
		newName := expandLocalizedString(d.Get("name"))
		input.Actions = append(
			input.Actions,
			&platform.DiscountCodeSetNameAction{Name: &newName})
	}

	if d.HasChange("description") {
		newDescription := expandLocalizedString(d.Get("description"))
		input.Actions = append(
			input.Actions,
			&platform.DiscountCodeSetDescriptionAction{Description: &newDescription})
	}

	if d.HasChange("predicate") {
		newPredicate := d.Get("predicate").(string)
		input.Actions = append(
			input.Actions,
			&platform.DiscountCodeSetCartPredicateAction{CartPredicate: &newPredicate})
	}

	if d.HasChange("max_applications") {
		newMaxApplications := d.Get("max_applications").(int)
		input.Actions = append(
			input.Actions,
			&platform.DiscountCodeSetMaxApplicationsAction{MaxApplications: &newMaxApplications})
	}

	if d.HasChange("max_applications_per_customer") {
		newMaxApplications := d.Get("max_applications_per_customer").(int)
		input.Actions = append(
			input.Actions,
			&platform.DiscountCodeSetMaxApplicationsPerCustomerAction{MaxApplicationsPerCustomer: &newMaxApplications})
	}

	if d.HasChange("cart_discounts") {
		newCartDiscounts := expandDiscountCodeCartDiscounts(d)
		input.Actions = append(
			input.Actions,
			&platform.DiscountCodeChangeCartDiscountsAction{CartDiscounts: newCartDiscounts})
	}

	if d.HasChange("groups") {
		newGroups := expandDiscountCodeGroups(d)
		if len(newGroups) > 0 {
			input.Actions = append(
				input.Actions,
				&platform.DiscountCodeChangeGroupsAction{Groups: newGroups})
		} else {
			input.Actions = append(
				input.Actions,
				&platform.DiscountCodeChangeGroupsAction{Groups: []string{}})
		}
	}

	if d.HasChange("is_active") {
		newIsActive := d.Get("is_active").(bool)
		input.Actions = append(
			input.Actions,
			&platform.DiscountCodeChangeIsActiveAction{IsActive: newIsActive})
	}

	if d.HasChange("valid_from") {
		if val := d.Get("valid_from").(string); len(val) > 0 {
			newValidFrom, err := expandTime(d.Get("valid_from").(string))
			if err != nil {
				return diag.FromErr(err)
			}
			input.Actions = append(
				input.Actions,
				&platform.DiscountCodeSetValidFromAction{ValidFrom: &newValidFrom})
		} else {
			input.Actions = append(
				input.Actions,
				&platform.DiscountCodeSetValidFromAction{})
		}
	}

	if d.HasChange("valid_until") {
		if val := d.Get("valid_until").(string); len(val) > 0 {
			newValidUntil, err := expandTime(d.Get("valid_until").(string))
			if err != nil {
				return diag.FromErr(err)
			}
			input.Actions = append(
				input.Actions,
				&platform.DiscountCodeSetValidUntilAction{ValidUntil: &newValidUntil})
		} else {
			input.Actions = append(
				input.Actions,
				&platform.DiscountCodeSetValidUntilAction{})
		}
	}

	if d.HasChange("custom") {
		actions, err := CustomFieldUpdateActions[platform.DiscountCodeSetCustomTypeAction, platform.DiscountCodeSetCustomFieldAction](ctx, client, d)
		if err != nil {
			return diag.FromErr(err)
		}
		for i := range actions {
			input.Actions = append(input.Actions, actions[i].(platform.DiscountCodeUpdateAction))
		}
	}

	err := resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		_, err := client.DiscountCodes().WithId(d.Id()).Post(input).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	return resourceDiscountCodeRead(ctx, d, m)
}

func resourceDiscountCodeDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	version := d.Get("version").(int)

	err := resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		_, err := client.DiscountCodes().WithId(d.Id()).Delete().Version(version).DataErasure(true).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	return diag.FromErr(err)
}

func expandDiscountCodeGroups(d *schema.ResourceData) []string {
	return expandStringArray(d.Get("groups").([]any))
}

func expandDiscountCodeCartDiscounts(d *schema.ResourceData) []platform.CartDiscountResourceIdentifier {
	discounts := d.Get("cart_discounts").([]any)

	cartDiscounts := make([]platform.CartDiscountResourceIdentifier, len(discounts))
	for i := range discounts {
		id := discounts[i].(string)
		cartDiscounts[i] = platform.CartDiscountResourceIdentifier{ID: &id}
	}
	return cartDiscounts
}

func flattenDiscountCodeCartDiscounts(values []platform.CartDiscountReference) []string {
	result := make([]string, len(values))
	for i := range values {
		result[i] = string(values[i].ID)
	}
	return result
}
