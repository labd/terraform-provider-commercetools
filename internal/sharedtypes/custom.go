package sharedtypes

import (
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/commercetools"
)

var (
	CustomSchema schema.Block = schema.SingleNestedBlock{
		MarkdownDescription: "Custom fields for this resource.",
		Attributes: map[string]schema.Attribute{
			"type_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the custom type to use for this resource.",
				Optional:            true,
			},
			"fields": schema.MapAttribute{
				ElementType: types.StringType,
				MarkdownDescription: "CustomValue fields for this resource. Note that " +
					"the values need to be provided as JSON encoded " +
					"strings: `my-value = jsonencode({\"key\": \"value\"})`",
				Optional: true,
			},
		},
		Validators: []validator.Object{
			// Ensure that a type_id is set if the custom block is set
			objectvalidator.AlsoRequires(path.MatchRelative().AtName("type_id")),
		},
	}
)

type CustomFieldTypeEncoder func(t *platform.Type, name string, value any) (any, error)

type Custom struct {
	TypeID *string           `tfsdk:"type_id"`
	Fields map[string]string `tfsdk:"fields"`
}

func (c *Custom) IsSet() bool {
	return c != nil && c.TypeID != nil
}

func (c *Custom) fieldsInterface() map[string]any {
	var result = make(map[string]any, len((*c).Fields))
	for key, value := range (*c).Fields {
		result[key] = value
	}

	return result
}

// Draft generates a custom fields draft. It uses the default custom field encoder.
func (c *Custom) Draft(t *platform.Type) (*platform.CustomFieldsDraft, error) {
	return c.draftWithEncoder(t, commercetools.CustomFieldEncodeType)
}

// draftWithEncoder generates a custom fields draft with a custom field encoder.
func (c *Custom) draftWithEncoder(t *platform.Type, encoder CustomFieldTypeEncoder) (*platform.CustomFieldsDraft, error) {
	if t == nil || c == nil || c.TypeID == nil {
		return nil, nil
	}

	draft := &platform.CustomFieldsDraft{
		Type: platform.TypeResourceIdentifier{
			ID: c.TypeID,
		},
	}

	fieldTypes := map[string]platform.FieldType{}
	for _, field := range t.FieldDefinitions {
		fieldTypes[field.Name] = field.Type
	}

	container := platform.FieldContainer{}
	for key, value := range c.Fields {
		enc, err := encoder(t, key, value)
		if err != nil {
			return nil, err
		}
		container[key] = enc
	}
	draft.Fields = &container

	return draft, nil
}

// CustomFieldUpdateActions generates the update actions for custom fields. It uses the default custom field encoder.
func CustomFieldUpdateActions[T commercetools.SetCustomTypeAction, F commercetools.SetCustomFieldAction](t *platform.Type, current, plan *Custom) ([]any, error) {
	return customFieldUpdateActionsWithEncoder[T, F](t, commercetools.CustomFieldEncodeType, current, plan)
}

// customFieldUpdateActionsWithEncoder generates the update actions for custom fields with a custom field encoder.
func customFieldUpdateActionsWithEncoder[T commercetools.SetCustomTypeAction, F commercetools.SetCustomFieldAction](t *platform.Type, encoder CustomFieldTypeEncoder, current, plan *Custom) ([]any, error) {
	// Remove custom field from resource
	if plan == nil {
		action := T{
			Type: nil,
		}
		return []any{action}, nil
	}

	if current == nil || (current.TypeID != plan.TypeID) {
		value, err := plan.draftWithEncoder(t, encoder)
		if err != nil {
			return nil, err
		}
		action := T{
			Type:   &value.Type,
			Fields: value.Fields,
		}
		return []any{action}, nil
	}

	changes := commercetools.DiffSlices(current.fieldsInterface(), plan.fieldsInterface())

	var result []any
	for key := range changes {
		if changes[key] == nil {
			result = append(result, F{
				Name:  key,
				Value: nil,
			})
		} else {
			val, err := encoder(t, key, changes[key])
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

func NewCustomFromNative(c *platform.CustomFields) (*Custom, error) {
	if c == nil {
		return nil, nil
	}

	var fields = map[string]string{}
	for key, value := range c.Fields {
		switch value.(type) {
		case string:
			fields[key] = value.(string)
		default:
			if v, err := json.Marshal(value); err == nil {
				fields[key] = string(v)
			} else {
				return nil, err
			}
		}
	}

	if len(fields) == 0 {
		fields = nil
	}

	return &Custom{
		TypeID: &c.Type.ID,
		Fields: fields,
	}, nil
}
