package commercetools

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
)

func CustomFieldSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"type_id": {
					Type:     schema.TypeString,
					Required: true,
				},
				"fields": {
					Type:     schema.TypeMap,
					Optional: true,
				},
			},
		},
	}
}

func CreateCustomFieldDraft(d *schema.ResourceData) *platform.CustomFieldsDraft {
	customData, err := elementFromList(d, "custom")
	if err != nil {
		panic(err)
	}
	return CreateCustomFieldDraftRaw(customData)
}

type SetCustomTypeAction interface {
	platform.ChannelSetCustomTypeAction
}

type SetCustomFieldAction interface {
	platform.ChannelSetCustomFieldAction
}

func CustomFieldCreateFieldContainer(data map[string]interface{}) *platform.FieldContainer {

	if raw, ok := data["fields"].(map[string]interface{}); ok {
		fields := platform.FieldContainer(raw)
		return &fields
	}
	return nil
}

func CreateCustomFieldDraftRaw(data map[string]interface{}) *platform.CustomFieldsDraft {
	draft := &platform.CustomFieldsDraft{}
	if val, ok := data["type_id"].(string); ok {
		draft.Type.ID = stringRef(val)
	}

	if raw, ok := data["fields"].(map[string]interface{}); ok {
		container := platform.FieldContainer(raw)
		draft.Fields = &container
	}

	return draft
}
