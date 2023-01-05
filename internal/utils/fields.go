package utils

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
