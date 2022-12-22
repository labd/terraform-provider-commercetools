package customvalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func DependencyValidator(value string, expressions ...path.Expression) validator.String {
	return dependencyValidator{
		value:           value,
		pathExpressions: expressions,
	}
}

var _ validator.String = dependencyValidator{}

// anyValidator implements the validator.
type dependencyValidator struct {
	value           string
	pathExpressions path.Expressions
}

// Description describes the validation in plain text formatting.
func (v dependencyValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("Magic")
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v dependencyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString performs the validation.
func (v dependencyValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	sourceVal := req.ConfigValue.ValueString()
	if sourceVal != v.value {
		return
	}

	expressions := req.PathExpression.MergeExpressions(v.pathExpressions...)

	for _, expression := range expressions {
		matchedPaths, diags := req.Config.PathMatches(ctx, expression)

		resp.Diagnostics.Append(diags...)

		// Collect all errors
		if diags.HasError() {
			continue
		}

		for _, mp := range matchedPaths {
			// If the user specifies the same attribute this validator is applied to,
			// also as part of the input, skip it
			if mp.Equal(req.Path) {
				continue
			}

			var mpVal attr.Value
			diags := req.Config.GetAttribute(ctx, mp, &mpVal)
			resp.Diagnostics.Append(diags...)

			// Collect all errors
			if diags.HasError() {
				continue
			}

			// Delay validation until all involved attribute have a known value
			if mpVal.IsUnknown() {
				return
			}

			if mpVal.IsNull() {
				resp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
					req.Path,
					fmt.Sprintf("Attribute %q must be specified when %q is %q", mp, req.Path, sourceVal),
				))
			}
		}
	}

}
