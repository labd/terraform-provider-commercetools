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
				Description: "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
				Type:        TypeLocalizedString,
				Optional:    true,
			},
			"description": {
				Description: "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
				Type:        TypeLocalizedString,
				Optional:    true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceChannelCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	name := unmarshallLocalizedString(d.Get("name"))
	description := unmarshallLocalizedString(d.Get("description"))

	roles := []platform.ChannelRoleEnum{}
	for _, value := range expandStringArray(d.Get("roles").([]interface{})) {
		roles = append(roles, platform.ChannelRoleEnum(value))
	}

	draft := platform.ChannelDraft{
		Key:         d.Get("key").(string),
		Roles:       roles,
		Name:        &name,
		Description: &description,
	}

	client := getClient(m)
	var channel *platform.Channel

	err := resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		var err error

		channel, err = client.Channels().Post(draft).Execute(ctx)
		if err != nil {
			return handleCommercetoolsError(err)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(channel.ID)
	d.Set("version", channel.Version)
	return resourceChannelRead(ctx, d, m)
}

func resourceChannelRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)
	channel, err := client.Channels().WithId(d.Id()).Get().Execute(ctx)

	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
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
	return nil
}

func resourceChannelUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		newName := unmarshallLocalizedString(d.Get("name"))
		input.Actions = append(
			input.Actions,
			&platform.ChannelChangeNameAction{Name: newName})
	}

	if d.HasChange("description") {
		newDescription := unmarshallLocalizedString(d.Get("description"))
		input.Actions = append(
			input.Actions,
			&platform.ChannelChangeDescriptionAction{Description: newDescription})
	}

	if d.HasChange("roles") {
		roles := []platform.ChannelRoleEnum{}
		for _, value := range expandStringArray(d.Get("roles").([]interface{})) {
			roles = append(roles, platform.ChannelRoleEnum(value))
		}
		input.Actions = append(
			input.Actions,
			&platform.ChannelSetRolesAction{Roles: roles})
	}

	_, err := client.Channels().WithId(d.Id()).Post(input).Execute(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceChannelRead(ctx, d, m)
}

func resourceChannelDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)
	version := d.Get("version").(int)
	_, err := client.Channels().WithId(d.Id()).Delete().WithQueryParams(platform.ByProjectKeyChannelsByIDRequestMethodDeleteInput{
		Version: version,
	}).Execute(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
