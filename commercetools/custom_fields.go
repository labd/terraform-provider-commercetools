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
	platform.ChannelSetCustomTypeAction |
		platform.StoreSetCustomTypeAction |
		platform.CategorySetCustomTypeAction |
		platform.ShippingMethodSetCustomTypeAction |
		platform.CustomerGroupSetCustomTypeAction |
		platform.DiscountCodeSetCustomTypeAction
}

type SetCustomFieldAction interface {
	platform.ChannelSetCustomFieldAction |
		platform.StoreSetCustomFieldAction |
		platform.CategorySetCustomFieldAction |
		platform.ShippingMethodSetCustomFieldAction |
		platform.CustomerGroupSetCustomFieldAction |
		platform.DiscountCodeSetCustomFieldAction
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
	if data["type_id"] == nil {
		return nil
	}

	if val, ok := data["type_id"].(string); ok {
		draft.Type.ID = stringRef(val)
	}

	if raw, ok := data["fields"].(map[string]interface{}); ok {
		container := platform.FieldContainer(raw)
		draft.Fields = &container
	}

	return draft
}

func flattenCustomFields(c *platform.CustomFields) any {
	if c == nil {
		return nil
	}
	result := map[string]any{}
	result["type_id"] = c.Type.ID

	fields := map[string]any{}
	for key, value := range c.Fields {
		fields[key] = value
	}
	result["fields"] = fields
	return []map[string]any{result}
}

func CustomFieldUpdateActions[T SetCustomTypeAction, F SetCustomFieldAction](d *schema.ResourceData) ([]any, error) {
	old, new := d.GetChange("custom")
	old_data := firstElementFromSlice(old.([]any))
	new_data := firstElementFromSlice(new.([]any))
	old_type_id := old_data["type_id"]
	new_type_id := new_data["type_id"]

	// Remove custom field from resource
	if new_type_id == nil {
		action := T{
			Type: nil,
		}
		return []any{action}, nil
	}

	if old_type_id == nil || (old_type_id.(string) != new_type_id.(string)) {
		value := CreateCustomFieldDraftRaw(new_data)
		action := platform.StoreSetCustomTypeAction{
			Type:   &value.Type,
			Fields: value.Fields,
		}
		return []any{action}, nil
	}

	changes := diffSlices(
		old_data["fields"].(map[string]interface{}),
		new_data["fields"].(map[string]interface{}))

	result := []any{}
	for key := range changes {
		result = append(result, F{
			Name:  key,
			Value: changes[key],
		})
	}
	return result, nil
}
