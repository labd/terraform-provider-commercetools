package commercetools

// import (
// 	"context"
// 	"time"

// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
// 	"github.com/labd/commercetools-go-sdk/platform"
// )

// func resourceChannel() *schema.Resource {
// 	return &schema.Resource{
// 		Description: "Channels represent a source or destination of different entities. They can be used to model " +
// 			"warehouses or stores.\n\n" +
// 			"See also the [Channels API Documentation](https://docs.commercetools.com/api/projects/channels)",
// 		Create: resourceChannelCreate,
// 		Read:   resourceChannelRead,
// 		Update: resourceChannelUpdate,
// 		Delete: resourceChannelDelete,
// 		Importer: &schema.ResourceImporter{
// 			State: schema.ImportStatePassthrough,
// 		},
// 		Schema: map[string]*schema.Schema{
// 			"key": {
// 				Description: "Any arbitrary string key that uniquely identifies this channel within the project",
// 				Type:        types.StringType,
// 				Required:    true,
// 			},
// 			"roles": {
// 				Description: "The [roles](https://docs.commercetools.com/api/projects/channels#channelroleenum) " +
// 					"of this channel. Each channel must have at least one role",
// 				Type:     types.ListType{ElemType: types.StringType},
// 				Required: true,
// 				Elem:     &schema.Schema{Type: types.StringType},
// 			},
// 			"name": {
// 				Description: "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
// 				Type:        TypeLocalizedString,
// 				Optional:    true,
// 			},
// 			"description": {
// 				Description: "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
// 				Type:        TypeLocalizedString,
// 				Optional:    true,
// 			},
// 			"version": {
// 				Type:     types.Int64Type,
// 				Computed: true,
// 			},
// 		},
// 	}
// }

// func resourceChannelCreate(d *schema.ResourceData, m interface{}) error {
// 	name := platform.LocalizedString(
// 		expandStringMap(d.Get("name").(map[string]interface{})))
// 	description := platform.LocalizedString(
// 		expandStringMap(d.Get("description").(map[string]interface{})))

// 	roles := []platform.ChannelRoleEnum{}
// 	for _, value := range expandStringArray(d.Get("roles").([]interface{})) {
// 		roles = append(roles, platform.ChannelRoleEnum(value))
// 	}

// 	draft := platform.ChannelDraft{
// 		Key:         d.Get("key").(string),
// 		Roles:       roles,
// 		Name:        &name,
// 		Description: &description,
// 	}

// 	client := getClient(m)
// 	var channel *platform.Channel

// 	err := resource.Retry(20*time.Second, func() *resource.RetryError {
// 		var err error

// 		channel, err = client.Channels().Post(draft).Execute(context.Background())
// 		if err != nil {
// 			return handleCommercetoolsError(err)
// 		}
// 		return nil
// 	})

// 	if err != nil {
// 		return err
// 	}

// 	d.SetId(channel.ID)
// 	d.Set("version", channel.Version)
// 	return resourceChannelRead(d, m)
// }

// func resourceChannelRead(d *schema.ResourceData, m interface{}) error {
// 	client := getClient(m)
// 	channel, err := client.Channels().WithId(d.Id()).Get().Execute(context.Background())

// 	if err != nil {
// 		if ctErr, ok := err.(platform.ErrorResponse); ok {
// 			if ctErr.StatusCode == 404 {
// 				d.SetId("")
// 				return nil
// 			}
// 		}
// 		return err
// 	}

// 	d.SetId(channel.ID)
// 	d.Set("version", channel.Version)

// 	if channel.Name != nil {
// 		d.Set("name", *channel.Name)
// 	}
// 	if channel.Description != nil {
// 		d.Set("description", *channel.Description)
// 	}
// 	d.Set("roles", channel.Roles)
// 	return nil
// }

// func resourceChannelUpdate(d *schema.ResourceData, m interface{}) error {
// 	client := getClient(m)

// 	input := platform.ChannelUpdate{
// 		Version: d.Get("version").(int),
// 		Actions: []platform.ChannelUpdateAction{},
// 	}

// 	if d.HasChange("key") {
// 		newKey := d.Get("key").(string)
// 		input.Actions = append(
// 			input.Actions,
// 			&platform.ChannelChangeKeyAction{Key: newKey})
// 	}

// 	if d.HasChange("name") {
// 		newName := platform.LocalizedString(
// 			expandStringMap(d.Get("name").(map[string]interface{})))
// 		input.Actions = append(
// 			input.Actions,
// 			&platform.ChannelChangeNameAction{Name: newName})
// 	}

// 	if d.HasChange("description") {
// 		newDescription := platform.LocalizedString(
// 			expandStringMap(d.Get("description").(map[string]interface{})))
// 		input.Actions = append(
// 			input.Actions,
// 			&platform.ChannelChangeDescriptionAction{Description: newDescription})
// 	}

// 	if d.HasChange("roles") {
// 		roles := []platform.ChannelRoleEnum{}
// 		for _, value := range expandStringArray(d.Get("roles").([]interface{})) {
// 			roles = append(roles, platform.ChannelRoleEnum(value))
// 		}
// 		input.Actions = append(
// 			input.Actions,
// 			&platform.ChannelSetRolesAction{Roles: roles})
// 	}

// 	_, err := client.Channels().WithId(d.Id()).Post(input).Execute(context.Background())
// 	if err != nil {
// 		return err
// 	}

// 	return resourceChannelRead(d, m)
// }

// func resourceChannelDelete(d *schema.ResourceData, m interface{}) error {
// 	client := getClient(m)
// 	version := d.Get("version").(int)
// 	_, err := client.Channels().WithId(d.Id()).Delete().WithQueryParams(platform.ByProjectKeyChannelsByIDRequestMethodDeleteInput{
// 		Version: version,
// 	}).Execute(context.Background())
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
