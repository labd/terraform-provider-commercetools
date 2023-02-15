package customtypes

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/labd/commercetools-go-sdk/platform"
)

var reLocalizedStringKey = regexp.MustCompile("^[a-z]{2}(-[A-Z]{2})?$")

type LocalizedStringOpts struct {
	Optional bool
}

type LocalizedStringType struct {
	types.MapType
}

func NewLocalizedStringType() LocalizedStringType {
	return LocalizedStringType{
		MapType: types.MapType{
			ElemType: types.StringType,
		},
	}
}

type LocalizedStringValue struct {
	types.Map
}

func LocalizedString(opts LocalizedStringOpts) schema.MapAttribute {
	attr := schema.MapAttribute{
		Optional:   opts.Optional,
		CustomType: NewLocalizedStringType(),
		Validators: []validator.Map{
			mapvalidator.KeysAre(
				stringvalidator.RegexMatches(
					reLocalizedStringKey,
					"invalid locale specified, ",
				),
			),
		},
	}

	return attr
}

func NewLocalizedStringNull() LocalizedStringValue {
	return LocalizedStringValue{
		basetypes.NewMapNull(types.StringType),
	}
}

func NewLocalizedStringValue(v map[string]attr.Value) LocalizedStringValue {
	val, diags := types.MapValue(types.StringType, v)
	retval := LocalizedStringValue{
		val,
	}

	// The only reason to get an error is when the element types are mixed,
	// e.g. some values are a types.String and some are types.Int.
	// Since we already validate this in the schema this should never happen
	// (unless there is a bug in the provider)
	if diags.HasError() {
		diagsStrings := make([]string, 0, len(diags))
		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}

		panic("NewLocalizedStringValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}
	return retval
}

func (l LocalizedStringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := l.MapType.ValueFromTerraform(ctx, in)

	return LocalizedStringValue{
		// unchecked type assertion
		val.(types.Map),
	}, err
}

// Equal returns true if `o` is also a MapType and has the same ElemType.
func (l LocalizedStringType) Equal(o attr.Type) bool {
	if l.ElemType == nil {
		return false
	}

	if other, ok := o.(LocalizedStringType); ok {
		return l.ElemType.Equal(other.ElemType)
	}

	// in some cases (not sure yet) we receive a MapType here. Since a
	// LocalizedString is just a MapType we accept that too.
	if other, ok := o.(basetypes.MapType); ok {
		return l.ElemType.Equal(other.ElemType)
	}
	return false
}

func (l LocalizedStringValue) ValueLocalizedString() platform.LocalizedString {
	result := platform.LocalizedString{}
	if l.IsUnknown() {
		return nil
	}
	diags := l.ElementsAs(context.Background(), &result, false)
	if diags.HasError() {
		panic("failed to transform localized string")
	}
	return result
}

func (l LocalizedStringValue) ValueLocalizedStringRef() *platform.LocalizedString {
	v := l.ValueLocalizedString()
	return &v
}
