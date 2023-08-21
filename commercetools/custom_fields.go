package commercetools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

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
		platform.DiscountCodeSetCustomTypeAction |
		platform.CartDiscountSetCustomTypeAction
}

type SetCustomFieldAction interface {
	platform.ChannelSetCustomFieldAction |
		platform.StoreSetCustomFieldAction |
		platform.CategorySetCustomFieldAction |
		platform.ShippingMethodSetCustomFieldAction |
		platform.CustomerGroupSetCustomFieldAction |
		platform.DiscountCodeSetCustomFieldAction |
		platform.CartDiscountSetCustomFieldAction
}

func customFieldEncodeType(t *platform.Type, name string, value any) (any, error) {
	// Suboptimal to do this everytime, however performance is not that important here and impact is negligible
	fieldTypes := map[string]platform.FieldType{}
	for _, field := range t.FieldDefinitions {
		fieldTypes[field.Name] = field.Type
	}

	fieldType, ok := fieldTypes[name]
	if !ok {
		return nil, fmt.Errorf("no field '%s' defined in type %s (%s)", name, t.Key, t.ID)
	}
	return customFieldEncodeValue(fieldType, name, value)
}

func customFieldEncodeValue(t platform.FieldType, name string, value any) (any, error) {
	switch v := t.(type) {
	case platform.CustomFieldLocalizedStringType:
		result := platform.LocalizedString{}
		if err := json.Unmarshal([]byte(value.(string)), &result); err != nil {
			return nil, fmt.Errorf("value for field '%s' needs to be a LocalizedString: '%v'", name, value)
		}
		return result, nil

	case platform.CustomFieldBooleanType:
		if value == "true" {
			return true, nil
		}
		if value == "false" {
			return false, nil
		}
		return nil, fmt.Errorf("value for field '%s' needs to be 'true' or 'false': '%v'", name, value)

	case platform.CustomFieldNumberType:
		result, err := strconv.ParseInt(value.(string), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("value for field '%s' needs to be a number: '%v'", name, value)
		}
		return result, nil

	case platform.CustomFieldSetType:
		var values []any
		if err := json.Unmarshal([]byte(value.(string)), &values); err != nil {
			return nil, fmt.Errorf("value for field '%s' needs to be an array: '%v'", name, value)
		}

		result := make([]any, len(values))
		for i := range values {
			var element = values[i]
			_, ok := element.(string)

			//We need to re-marshal the data here, so we can recursively pass it back to the encoding function
			if !ok {
				marshalledValue, err := json.Marshal(values[i])
				if err != nil {
					return nil, err
				}
				element = string(marshalledValue)
			}
			itemValue, err := customFieldEncodeValue(v.ElementType, name, element)
			if err != nil {
				return nil, err
			}
			result[i] = itemValue
		}
		return result, nil

	case platform.CustomFieldReferenceType:
		var result any
		if err := json.Unmarshal([]byte(value.(string)), &result); err != nil {
			return nil, fmt.Errorf("value for field '%s' needs to be an object: '%v'", name, value)
		}
		return result, nil

	case platform.CustomFieldMoneyType:
		var result *platform.CentPrecisionMoney
		if err := json.Unmarshal([]byte(value.(string)), &result); err != nil {
			return nil, fmt.Errorf("value for field '%s' needs to be a CentPrecisionMoney: '%v'", name, value)
		}
		return result, nil

	case platform.CustomFieldDateType:
		result, err := time.Parse("2006-01-02", value.(string))
		if err != nil {
			return nil, fmt.Errorf("value for field '%s' needs to be a valid ISO-8601 date (YYYY-MM-DD): '%v'", name, value)
		}
		return result.Format("2006-01-02"), nil

	case platform.CustomFieldDateTimeType:
		result, err := time.Parse(time.RFC3339Nano, value.(string))
		if err != nil {
			return nil, fmt.Errorf("value for field '%s' needs to be a valid ISO-8601 datetime (YYYY-MM-DDThh:mm:ss.sssZ): '%v'", name, value)
		}
		return result.Format("2006-01-02T15:04:05.000Z"), nil

	case platform.CustomFieldTimeType:
		result, err := time.Parse(time.RFC3339Nano, strings.Join([]string{"0001-01-01T", value.(string), "Z"}, ""))
		if err != nil {
			return nil, fmt.Errorf("value for field '%s' needs to be a valid ISO-8601 time (hh:mm:ss.sss): '%v'", name, value)
		}
		return result.Format("15:04:05.000"), nil

	case platform.CustomFieldEnumType, platform.CustomFieldLocalizedEnumType, platform.CustomFieldStringType:
		return value, nil

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
			enc, err := customFieldEncodeType(t, key, value)
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
		switch value.(type) {
		case string:
			fields[key] = value
		default:
			if v, err := json.Marshal(value); err == nil {
				fields[key] = string(v)
			} else {
				panic(err)
			}
		}
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

	oldState, newState := d.GetChange("custom")
	oldData := firstElementFromSlice(oldState.([]any))
	newData := firstElementFromSlice(newState.([]any))
	oldTypeId := oldData["type_id"]
	newTypeId := newData["type_id"]

	// Remove custom field from resource
	if newTypeId == nil {
		action := T{
			Type: nil,
		}
		return []any{action}, nil
	}

	if oldTypeId == nil || (oldTypeId.(string) != newTypeId.(string)) {
		value, err := CreateCustomFieldDraftRaw(newData, t)
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
		oldData["fields"].(map[string]any),
		newData["fields"].(map[string]any))

	var result []any
	for key := range changes {
		if changes[key] == nil {
			result = append(result, F{
				Name:  key,
				Value: nil,
			})
		} else {
			val, err := customFieldEncodeType(t, key, changes[key])
			if err != nil {
				return nil, err
			}
			result = append(result, F{
				Name:  key,
				Value: val,
			})
		}
	}
	return result, nil
}
