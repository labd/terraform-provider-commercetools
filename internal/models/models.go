package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

type CustomFieldLocalizedEnumValue struct {
	Key   types.String                     `tfsdk:"key"`
	Label customtypes.LocalizedStringValue `tfsdk:"label"`
}

func (c CustomFieldLocalizedEnumValue) ToNative() platform.CustomFieldLocalizedEnumValue {
	return platform.CustomFieldLocalizedEnumValue{
		Key:   c.Key.ValueString(),
		Label: c.Label.ValueLocalizedString(),
	}
}

func NewCustomFieldLocalizedEnumValue(s platform.CustomFieldLocalizedEnumValue) CustomFieldLocalizedEnumValue {
	label := utils.FromOptionalLocalizedString(&s.Label)
	result := CustomFieldLocalizedEnumValue{
		Key:   types.StringValue(s.Key),
		Label: label,
	}

	return result
}
