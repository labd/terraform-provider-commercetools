package customvalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func RequireValueValidator(value string, expressions ...path.Expression) validator.List {
	return requireValueValidator{
		value:           value,
		pathExpressions: expressions,
	}
}

var _ validator.List = requireValueValidator{}

// anyValidator implements the validator.
type requireValueValidator struct {
	value           string
	pathExpressions path.Expressions
}

// Description describes the validation in plain text formatting.
func (v requireValueValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("Magic")
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v requireValueValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString performs the validation.
func (v requireValueValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {

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

			val, ok := mpVal.(basetypes.StringValue)
			if !ok {
				resp.Diagnostics.Append(validatordiag.BugInProviderDiagnostic(
					fmt.Sprintf("Expression %q must resolve to a string", mp),
				))
			}

			if len(req.ConfigValue.Elements()) == 0 {
				continue
			}

			if val.ValueString() != v.value {
				resp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
					req.Path,
					fmt.Sprintf("Block %q can only be specified when %q is %q", req.Path, mp, v.value),
				))
			}
		}
	}

}
