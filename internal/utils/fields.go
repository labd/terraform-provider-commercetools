package utils

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
)

func OptionalString(value types.String) *string {
	if value.IsUnknown() || value.IsNull() {
		return nil
	}

	val := value.ValueString()
	return &val
}

func OptionalInt(value types.Int64) *int {
	if value.IsUnknown() || value.IsNull() {
		return nil
	}

	val := int(value.ValueInt64())
	return &val
}

func FromOptionalString(value *string) basetypes.StringValue {
	if value == nil {
		return types.StringNull()
	}
	return types.StringValue(*value)

}
func FromOptionalLocalizedString(value *platform.LocalizedString) customtypes.LocalizedStringValue {
	if value == nil {
		return customtypes.NewLocalizedStringNull()
	}

	return FromLocalizedString(*value)
}

func FromOptionalInt(value *int) basetypes.Int64Value {
	if value == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*value))
}

func FromOptionalBool(value *bool) basetypes.BoolValue {
	if value == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*value)
}

func FromLocalizedString(value platform.LocalizedString) customtypes.LocalizedStringValue {
	result := make(map[string]attr.Value, len(value))
	for k, v := range value {
		result[k] = types.StringValue(v)
	}
	return customtypes.NewLocalizedStringValue(result)
}

func StringRef(value any) *string {
	if value == nil {
		return nil
	}
	result := value.(string)
	return &result
}

func IntRef(value any) *int {
	result := value.(int)
	return &result
}

func BoolRef(value any) *bool {
	result := value.(bool)
	return &result
}
