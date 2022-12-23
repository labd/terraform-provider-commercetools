package customtypes

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var reLocalizedStringKey = regexp.MustCompile("^[a-z]{2}(-[A-Z]{2})?$")

type LocalizedStringOpts struct {
	Optional bool
}

func LocalizedString(opts LocalizedStringOpts) schema.MapAttribute {
	attr := schema.MapAttribute{
		ElementType: types.StringType,
		Optional:    opts.Optional,
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
