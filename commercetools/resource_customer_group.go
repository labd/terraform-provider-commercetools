package commercetools

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
)

func resourceCustomerGroup() *schema.Resource {
	return &schema.Resource{
		Description: "A Customer can be a member of a customer group (for example reseller, gold member). " +
			"Special prices can be assigned to specific products based on a customer group.\n\n" +
			"See also the [Custome Group API Documentation](https://docs.commercetools.com/api/projects/customerGroups)",
		CreateContext: resourceCustomerGroupCreate,
		ReadContext:   resourceCustomerGroupRead,
		UpdateContext: resourceCustomerGroupUpdate,
		DeleteContext: resourceCustomerGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "User-specific unique identifier for the customer group",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Description: "Unique within the project",
				Type:        schema.TypeString,
				Required:    true,
			},
			"custom": CustomFieldSchema(),
		},
	}
}

func resourceCustomerGroupCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	custom, err := CreateCustomFieldDraft(ctx, client, d)
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	draft := platform.CustomerGroupDraft{
		GroupName: d.Get("name").(string),
		Custom:    custom,
	}

	key := stringRef(d.Get("key"))
	if *key != "" {
		draft.Key = key
	}

	var customerGroup *platform.CustomerGroup
	err = resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		var err error
		customerGroup, err = client.CustomerGroups().Post(draft).Execute(ctx)
		return processRemoteError(err)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if customerGroup == nil {
		return diag.Errorf("No customer group")
	}

	d.SetId(customerGroup.ID)
	d.Set("version", customerGroup.Version)

	return resourceCustomerGroupRead(ctx, d, m)
}

func resourceCustomerGroupRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	customerGroup, err := client.CustomerGroups().WithId(d.Id()).Get().Execute(ctx)
	if err != nil {
		if IsResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if customerGroup == nil {
		d.SetId("")
	} else {
		d.Set("version", customerGroup.Version)
		d.Set("name", customerGroup.Name)
		d.Set("key", customerGroup.Key)
		d.Set("custom", flattenCustomFields(customerGroup.Custom))
	}

	return nil
}

func resourceCustomerGroupUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	input := platform.CustomerGroupUpdate{
		Version: d.Get("version").(int),
		Actions: []platform.CustomerGroupUpdateAction{},
	}

	if d.HasChange("name") {
		newName := d.Get("name").(string)
		input.Actions = append(
			input.Actions,
			&platform.CustomerGroupChangeNameAction{Name: newName})
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&platform.CustomerGroupSetKeyAction{Key: nilIfEmpty(&newKey)})
	}

	if d.HasChange("custom") {
		actions, err := CustomFieldUpdateActions[platform.CustomerGroupSetCustomTypeAction, platform.CustomerGroupSetCustomFieldAction](ctx, client, d)
		if err != nil {
			return diag.FromErr(err)
		}
		for i := range actions {
			input.Actions = append(input.Actions, actions[i].(platform.CustomerGroupUpdateAction))
		}
	}

	err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		_, err := client.CustomerGroups().WithId(d.Id()).Post(input).Execute(ctx)
		return processRemoteError(err)
	})
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	return resourceCustomerGroupRead(ctx, d, m)
}

func resourceCustomerGroupDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	version := d.Get("version").(int)
	err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		_, err := client.CustomerGroups().WithId(d.Id()).Delete().Version(version).Execute(ctx)
		return processRemoteError(err)
	})
	return diag.FromErr(err)
}
