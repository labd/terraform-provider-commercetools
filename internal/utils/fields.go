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

func FromOptionalString(value *string) basetypes.StringValue {
	if value == nil {
		return types.StringUnknown()
	}
	return types.StringValue(*value)
}

func StringRef(value any) *string {
	if value == nil {
		return nil
	}
	result := value.(string)
	return &result
}
