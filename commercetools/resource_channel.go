package commercetools

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
)

func resourceChannel() *schema.Resource {
	return &schema.Resource{
		Description: "Channels represent a source or destination of different entities. They can be used to model " +
			"warehouses or stores.\n\n" +
			"See also the [Channels API Documentation](https://docs.commercetools.com/api/projects/channels)",
		CreateContext: resourceChannelCreate,
		ReadContext:   resourceChannelRead,
		UpdateContext: resourceChannelUpdate,
		DeleteContext: resourceChannelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "Any arbitrary string key that uniquely identifies this channel within the project",
				Type:        schema.TypeString,
				Required:    true,
			},
			"roles": {
				Description: "The [roles](https://docs.commercetools.com/api/projects/channels#channelroleenum) " +
					"of this channel. Each channel must have at least one role",
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
			"address": AddressFieldSchema(),
			"custom":  CustomFieldSchema(),
			"geolocation": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"coordinates": {
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeFloat},
							MinItems: 2,
							MaxItems: 2,
							Required: true,
						},
					},
				},
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceChannelCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	name := expandLocalizedString(d.Get("name"))
	description := expandLocalizedString(d.Get("description"))

	roles := []platform.ChannelRoleEnum{}
	for _, value := range expandStringArray(d.Get("roles").([]any)) {
		roles = append(roles, platform.ChannelRoleEnum(value))
	}

	client := getClient(m)

	custom, err := CreateCustomFieldDraft(ctx, client, d)
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	draft := platform.ChannelDraft{
		Key:         d.Get("key").(string),
		Roles:       roles,
		Name:        &name,
		Description: &description,
		Address:     CreateAddressFieldDraft(d),
		Custom:      custom,
		GeoLocation: expandGeoLocation(d),
	}

	var channel *platform.Channel
	err = resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		var err error

		channel, err = client.Channels().Post(draft).Execute(ctx)
		return processRemoteError(err)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(channel.ID)
	d.Set("version", channel.Version)
	return resourceChannelRead(ctx, d, m)
}

func resourceChannelRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	channel, err := client.Channels().WithId(d.Id()).Get().Execute(ctx)

	if err != nil {
		if IsResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(channel.ID)
	d.Set("version", channel.Version)
	if channel.Name != nil {
		d.Set("name", *channel.Name)
	}
	if channel.Description != nil {
		d.Set("description", *channel.Description)
	}
	d.Set("roles", channel.Roles)
	d.Set("address", flattenAddress(channel.Address))
	d.Set("geolocation", flattenGeoLocation(channel.GeoLocation))
	d.Set("custom", flattenCustomFields(channel.Custom))
	return nil
}

func resourceChannelUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	input := platform.ChannelUpdate{
		Version: d.Get("version").(int),
		Actions: []platform.ChannelUpdateAction{},
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&platform.ChannelChangeKeyAction{Key: newKey})
	}

	if d.HasChange("name") {
		newName := expandLocalizedString(d.Get("name"))
		input.Actions = append(
			input.Actions,
			&platform.ChannelChangeNameAction{Name: newName})
	}

	if d.HasChange("description") {
		newDescription := expandLocalizedString(d.Get("description"))
		input.Actions = append(
			input.Actions,
			&platform.ChannelChangeDescriptionAction{Description: newDescription})
	}

	if d.HasChange("roles") {
		roles := []platform.ChannelRoleEnum{}
		for _, value := range expandStringArray(d.Get("roles").([]any)) {
			roles = append(roles, platform.ChannelRoleEnum(value))
		}
		input.Actions = append(
			input.Actions,
			&platform.ChannelSetRolesAction{Roles: roles})
	}

	if d.HasChange("address") {
		newAddr := CreateAddressFieldDraft(d)
		input.Actions = append(
			input.Actions,
			&platform.ChannelSetAddressAction{Address: newAddr})
	}

	if d.HasChange("geolocation") {
		newGeoLocation := expandGeoLocation(d)
		input.Actions = append(
			input.Actions,
			&platform.ChannelSetGeoLocationAction{GeoLocation: newGeoLocation})
	}

	if d.HasChange("custom") {
		actions, err := CustomFieldUpdateActions[platform.ChannelSetCustomTypeAction, platform.ChannelSetCustomFieldAction](ctx, client, d)
		if err != nil {
			return diag.FromErr(err)
		}
		for i := range actions {
			input.Actions = append(input.Actions, actions[i].(platform.ChannelUpdateAction))
		}
	}

	err := resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		_, err := client.Channels().WithId(d.Id()).Post(input).Execute(ctx)
		return processRemoteError(err)
	})
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	return resourceChannelRead(ctx, d, m)
}

func resourceChannelDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	version := d.Get("version").(int)
	err := resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		_, err := client.Channels().WithId(d.Id()).Delete().Version(version).Execute(ctx)
		return processRemoteError(err)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func flattenGeoLocation(loc platform.GeoJson) []map[string]any {
	switch l := loc.(type) {
	case platform.GeoJsonPoint:
		return []map[string]any{
			{
				"coordinates": l.Coordinates,
			},
		}
	}

	return []map[string]any{}
}

func expandGeoLocation(d *schema.ResourceData) platform.GeoJson {
	if geolocation, err := elementFromList(d, "geolocation"); err == nil {
		if geolocation == nil {
			return nil
		}

		points := geolocation["coordinates"].([]any)
		return platform.GeoJsonPoint{
			Coordinates: []float64{
				points[0].(float64),
				points[1].(float64),
			},
		}
	}
	return nil
}
