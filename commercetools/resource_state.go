package commercetools

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/labd/commercetools-go-sdk/platform"
)

func resourceState() *schema.Resource {
	return &schema.Resource{
		Description: "The commercetools platform allows you to model states of certain objects, such as orders, line " +
			"items, products, reviews, and payments to define finite state machines reflecting the business " +
			"logic you'd like to implement.\n\n" +
			"See also the [State API Documentation](https://docs.commercetools.com/api/projects/states)",
		CreateContext: resourceStateCreate,
		ReadContext:   resourceStateRead,
		UpdateContext: resourceStateUpdate,
		DeleteContext: resourceStateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "A unique identifier for the state",
				Type:        schema.TypeString,
				Required:    true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"type": {
				Description: "[StateType](https://docs.commercetools.com/api/projects/states#statetype)",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					string(platform.StateTypeEnumOrderState),
					string(platform.StateTypeEnumLineItemState),
					string(platform.StateTypeEnumProductState),
					string(platform.StateTypeEnumReviewState),
					string(platform.StateTypeEnumPaymentState),
				}, false),
			},
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
			"initial": {
				Description: "A state can be declared as an initial state for any state machine. When a workflow " +
					"starts, this first state must be an initial state",
				Type:     schema.TypeBool,
				Optional: true,
			},
			"roles": {
				Description: "Array of [State Role](https://docs.commercetools.com/api/projects/states#staterole)",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						string(platform.StateRoleEnumReviewIncludedInStatistics),
						string(platform.StateRoleEnumReturn),
					}, false),
				},
			},
		},
	}
}

func resourceStateCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	name := expandLocalizedString(d.Get("name"))
	description := expandLocalizedString(d.Get("description"))

	roles := []platform.StateRoleEnum{}
	for _, value := range expandStringArray(d.Get("roles").([]any)) {
		roles = append(roles, platform.StateRoleEnum(value))
	}

	draft := platform.StateDraft{
		Key:         d.Get("key").(string),
		Type:        platform.StateTypeEnum(d.Get("type").(string)),
		Name:        &name,
		Description: &description,
		Roles:       roles,
	}

	// Note the use of GetOk since it's an optional bool type
	if _, exists := d.GetOk("initial"); exists {
		draft.Initial = boolRef(d.Get("initial"))
	}

	client := getClient(m)
	var state *platform.State

	err := resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		var err error
		state, err = client.States().Post(draft).Execute(ctx)
		return processRemoteError(err)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(state.ID)
	d.Set("version", state.Version)
	return resourceStateRead(ctx, d, m)
}

func resourceStateRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	state, err := client.States().WithId(d.Id()).Get().Execute(ctx)
	if err != nil {
		if IsResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(state.ID)
	d.Set("version", state.Version)
	d.Set("key", state.Key)
	d.Set("type", state.Type)
	if state.Name != nil {
		d.Set("name", *state.Name)
	}
	if state.Description != nil {
		d.Set("description", *state.Description)
	}
	d.Set("initial", state.Initial)
	if state.Roles != nil {
		d.Set("roles", state.Roles)
	}
	return nil
}

func resourceStateUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	input := platform.StateUpdate{
		Version: d.Get("version").(int),
		Actions: []platform.StateUpdateAction{},
	}

	if d.HasChange("name") {
		newName := expandLocalizedString(d.Get("name"))
		input.Actions = append(
			input.Actions,
			&platform.StateSetNameAction{Name: newName})
	}

	if d.HasChange("description") {
		newDescription := expandLocalizedString(d.Get("description"))
		input.Actions = append(
			input.Actions,
			&platform.StateSetDescriptionAction{Description: newDescription})
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&platform.StateChangeKeyAction{Key: newKey})
	}

	if d.HasChange("type") {
		newType := d.Get("type").(platform.StateTypeEnum)
		input.Actions = append(
			input.Actions,
			&platform.StateChangeTypeAction{Type: newType})
	}

	if d.HasChange("initial") {
		newInitial := d.Get("initial").(bool)
		input.Actions = append(
			input.Actions,
			&platform.StateChangeInitialAction{Initial: newInitial})
	}

	if d.HasChange("roles") {
		roles := []platform.StateRoleEnum{}
		for _, value := range expandStringArray(d.Get("roles").([]any)) {
			roles = append(roles, platform.StateRoleEnum(value))
		}
		input.Actions = append(
			input.Actions,
			&platform.StateSetRolesAction{Roles: roles})
	}

	err := resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		_, err := client.States().WithId(d.Id()).Post(input).Execute(ctx)
		return processRemoteError(err)
	})
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	return resourceStateRead(ctx, d, m)
}

func resourceStateDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	version := d.Get("version").(int)
	err := resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		_, err := client.States().WithId(d.Id()).Delete().Version(version).Execute(ctx)
		return processRemoteError(err)
	})
	return diag.FromErr(err)
}
