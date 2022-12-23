package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
)

type CustomFieldLocalizedEnumValue struct {
	Key   types.String            `tfsdk:"key"`
	Label map[string]types.String `tfsdk:"label"`
}

func (c CustomFieldLocalizedEnumValue) ToNative() platform.CustomFieldLocalizedEnumValue {
	label := make(map[string]string, len(c.Label))
	for key, value := range c.Label {
		label[key] = value.ValueString()
	}
	return platform.CustomFieldLocalizedEnumValue{
		Key:   c.Key.ValueString(),
		Label: label,
	}
}

func NewCustomFieldLocalizedEnumValue(s platform.CustomFieldLocalizedEnumValue) CustomFieldLocalizedEnumValue {
	label := make(map[string]types.String, len(s.Label))
	for k, v := range s.Label {
		label[k] = types.StringValue(v)
	}
	return CustomFieldLocalizedEnumValue{
		Key:   types.StringValue(s.Key),
		Label: label,
	}
}
