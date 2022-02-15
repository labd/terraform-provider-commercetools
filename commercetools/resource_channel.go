package commercetools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
)

func resourceChannel() *schema.Resource {
	return &schema.Resource{
		Description: "Channels represent a source or destination of different entities. They can be used to model " +
			"warehouses or stores.\n\n" +
			"See also the [Channels API Documentation](https://docs.commercetools.com/api/projects/channels)",
		Create: resourceChannelCreate,
		Read:   resourceChannelRead,
		Update: resourceChannelUpdate,
		Delete: resourceChannelDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
			"custom": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type_key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"field": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The field name of a custom field (https://docs.commercetools.com/api/projects/channels#set-customfield)",
									},
									"value": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The value of a custom field (https://docs.commercetools.com/api/projects/channels#set-customfield) expected as json encoded field to handle all different cases",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceChannelCreate(d *schema.ResourceData, m interface{}) error {
	name := platform.LocalizedString(
		expandStringMap(d.Get("name").(map[string]interface{})))
	description := platform.LocalizedString(
		expandStringMap(d.Get("description").(map[string]interface{})))

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

	//custom fields are set to be filled
	if d.HasChange("custom") {
		typeId, fields := getCustomFieldsData(d)

		draft.Custom = &platform.CustomFieldsDraft{
			Type:   *typeId,
			Fields: fields,
		}
	}

	client := getClient(m)
	var channel *platform.Channel

	err := resource.Retry(20*time.Second, func() *resource.RetryError {
		var err error

		channel, err = client.Channels().Post(draft).Execute(context.Background())
		if err != nil {
			return handleCommercetoolsError(err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	d.SetId(channel.ID)
	if err := d.Set("version", channel.Version); err != nil {
		return fmt.Errorf("error reading channel: %s", err)
	}
	return resourceChannelRead(d, m)
}

func getCustomFieldsData(d *schema.ResourceData) (*platform.TypeResourceIdentifier, *platform.FieldContainer) {
	custom := d.Get("custom").([]interface{})[0].(map[string]interface{})

	typeId := &platform.TypeResourceIdentifier{
		Key: custom["type_key"].(*string),
	}

	fields := &platform.FieldContainer{}

	for _, fieldDef := range custom["field"].([]interface{}) {
		key := fieldDef.(map[string]interface{})["name"].(string)
		value := fieldDef.(map[string]interface{})["value"].(string)
		decodedValue := _decodeCustomFieldValue(value)

		(*fields)[key] = decodedValue

	}
	return typeId, fields
}

func _decodeCustomFieldValue(value string) interface{} {
	var data interface{}
	_ = json.Unmarshal([]byte(value), &data)
	return data
}

func _encodeCustomFieldValue(value interface{}) string {
	data, _ := json.Marshal(value)

	return string(data)
}

func resourceChannelRead(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	channel, err := client.Channels().WithId(d.Id()).Get().Execute(context.Background())

	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.SetId(channel.ID)
	if err := d.Set("version", channel.Version); err != nil {
		return fmt.Errorf("error reading channel: %s", err)
	}

	if channel.Name != nil {
		if err := d.Set("name", *channel.Name); err != nil {
			return fmt.Errorf("error reading channel: %s", err)
		}
	}
	if channel.Description != nil {
		if err := d.Set("description", *channel.Description); err != nil {
			return fmt.Errorf("error reading channel: %s", err)
		}
	}
	if err := d.Set("roles", channel.Roles); err != nil {
		return fmt.Errorf("error reading channel: %s", err)
	}

	if channel.Custom != nil {
		data := _decodeCustomFieldValue(_encodeCustomFieldValue(channel.Custom.Fields))

		customStateFields := make([]interface{}, 0)

		//if the length would be 0 we are reading from a remote channel which has already custom fields set
		//but the terraform state does not match it yet
		//for the case that we read from the remote channel and the state has custom fields we will use the order of the
		//existing terraform state all additional fields will be added to the state at then end of the list
		if len(d.Get("custom").([]interface{})) != 0 {
			customState := d.Get("custom").([]interface{})[0].(map[string]interface{})

			customStateFields = customState["field"].([]interface{})
		}

		for fieldKey, fieldValue := range data.(map[string]interface{}) {

			idx := -1

			for i := range customStateFields {
				if customStateFields[i].(map[string]interface{})["name"] == fieldKey {
					idx = i
					break
				}
			}

			//add to list of fields as the state does not know about this field but remote it exists
			if idx == -1 {

				customStateFields = append(customStateFields, map[string]interface{}{
					"name":  fieldKey,
					"value": _encodeCustomFieldValue(fieldValue),
				})
				continue
			}

			//update field value
			customStateFields[idx].(map[string]interface{})["value"] = _encodeCustomFieldValue(fieldValue)
		}

		customBase := []interface{}{map[string]interface{}{
			"type_key": channel.Custom.Type.Obj.Key,
			"field":    customStateFields,
		}}

		if err := d.Set("custom", customBase); err != nil {
			return fmt.Errorf("error reading channel: %s", err)
		}
	}
	return nil
}

func resourceChannelUpdate(d *schema.ResourceData, m interface{}) error {
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
		newName := platform.LocalizedString(
			expandStringMap(d.Get("name").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&platform.ChannelChangeNameAction{Name: newName})
	}

	if d.HasChange("description") {
		newDescription := platform.LocalizedString(
			expandStringMap(d.Get("description").(map[string]interface{})))
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

	if d.HasChange("custom") {
		typeId, fields := getCustomFieldsData(d)

		input.Actions = append(
			input.Actions,
			&platform.ChannelSetCustomTypeAction{Type: typeId, Fields: fields})
	}

	_, err := client.Channels().WithId(d.Id()).Post(input).Execute(context.Background())
	if err != nil {
		return err
	}

	return resourceChannelRead(d, m)
}

func resourceChannelDelete(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	version := d.Get("version").(int)
	_, err := client.Channels().WithId(d.Id()).Delete().WithQueryParams(platform.ByProjectKeyChannelsByIDRequestMethodDeleteInput{
		Version: version,
	}).Execute(context.Background())
	if err != nil {
		return err
	}

	return nil
}
