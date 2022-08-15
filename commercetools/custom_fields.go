package commercetools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
)

var cacheTypes map[string]*platform.Type

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

func CreateCustomFieldDraft(ctx context.Context, client *platform.ByProjectKeyRequestBuilder, d *schema.ResourceData) (*platform.CustomFieldsDraft, error) {
	customData, err := elementFromList(d, "custom")
	if err != nil {
		return nil, err
	}

	t, err := getTypeResource(ctx, client, d)
	if err != nil {
		return nil, err
	}

	return CreateCustomFieldDraftRaw(customData, t)
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

func CustomFieldCreateFieldContainer(data map[string]any) *platform.FieldContainer {
	if raw, ok := data["fields"].(map[string]any); ok {
		fields := platform.FieldContainer(raw)
		return &fields
	}
	return nil
}

func customFieldEncodeValue(t *platform.Type, name string, value any) (any, error) {
	// Sub-optimal to do this everytime, however performance is not that
	// important here and impact is neglible
	fieldTypes := map[string]platform.FieldType{}
	for _, field := range t.FieldDefinitions {
		fieldTypes[field.Name] = field.Type
	}

	fieldType, ok := fieldTypes[name]
	if !ok {
		return nil, fmt.Errorf("no field '%s' defined in type %s (%s)", name, t.Key, t.ID)
	}

	switch fieldType.(type) {
	case platform.CustomFieldLocalizedStringType:
		result := platform.LocalizedString{}
		if err := json.Unmarshal([]byte(value.(string)), &result); err != nil {
			return nil, fmt.Errorf("value for field '%s' is not a valid LocalizedString value: '%v'", name, value)
		}
		return result, nil
	case platform.CustomFieldBooleanType:
		if value == "true" {
			return true, nil
		}
		if value == "false" {
			return false, nil
		}
		return nil, fmt.Errorf("unrecognized boolean value")
	case platform.CustomFieldNumberType:
		result, err := strconv.ParseInt(value.(string), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("value for field '%s' is not a valid Number value: '%v'", name, value)
		}
		return result, nil
	default:
		return value, nil
	}
}

func CreateCustomFieldDraftRaw(data map[string]any, t *platform.Type) (*platform.CustomFieldsDraft, error) {
	if data["type_id"] == nil {
		return nil, nil
	}

	draft := &platform.CustomFieldsDraft{}
	if val, ok := data["type_id"].(string); ok {
		draft.Type.ID = stringRef(val)
	}

	fieldTypes := map[string]platform.FieldType{}
	for _, field := range t.FieldDefinitions {
		fieldTypes[field.Name] = field.Type
	}

	if raw, ok := data["fields"].(map[string]any); ok {
		container := platform.FieldContainer{}
		for key, value := range raw {
			enc, err := customFieldEncodeValue(t, key, value)
			if err != nil {
				return nil, err
			}
			container[key] = enc
		}
		draft.Fields = &container
	}

	return draft, nil
}

func flattenCustomFields(c *platform.CustomFields) []map[string]any {
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

// getTypeResource returns the platform.Type for the type_id in the custom
// field. The type_id is cached to minimize API calls when multiple resource
// use the same type
func getTypeResource(ctx context.Context, client *platform.ByProjectKeyRequestBuilder, d *schema.ResourceData) (*platform.Type, error) {
	custom := d.Get("custom")
	data := firstElementFromSlice(custom.([]any))
	if data == nil {
		return nil, nil
	}

	if type_id, ok := data["type_id"].(string); ok {
		if cacheTypes == nil {
			cacheTypes = make(map[string]*platform.Type)
		}
		if t, exists := cacheTypes[type_id]; exists {
			if t == nil {
				return nil, fmt.Errorf("type %s not in cache due to previous error", type_id)
			}
			return t, nil
		}

		t, err := client.Types().WithId(type_id).Get().Execute(ctx)
		cacheTypes[type_id] = t
		return t, err
	}
	return nil, fmt.Errorf("missing type_id for custom fields")
}

func CustomFieldUpdateActions[T SetCustomTypeAction, F SetCustomFieldAction](ctx context.Context, client *platform.ByProjectKeyRequestBuilder, d *schema.ResourceData) ([]any, error) {
	t, err := getTypeResource(ctx, client, d)
	if err != nil {
		return nil, err
	}

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
		value, err := CreateCustomFieldDraftRaw(new_data, t)
		if err != nil {
			return nil, err
		}
		action := platform.StoreSetCustomTypeAction{
			Type:   &value.Type,
			Fields: value.Fields,
		}
		return []any{action}, nil
	}

	changes := diffSlices(
		old_data["fields"].(map[string]any),
		new_data["fields"].(map[string]any))

	result := []any{}
	for key := range changes {
		val, err := customFieldEncodeValue(t, key, changes[key])
		if err != nil {
			return nil, err
		}
		result = append(result, F{
			Name:  key,
			Value: val,
		})
	}
	return result, nil
}
