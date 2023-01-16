package custommodifiers

// From https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#creating-attribute-plan-modifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// stringDefaultModifier is a plan modifier that sets a default value for a
// types.StringType attribute when it is not configured. The attribute must be
// marked as Optional and Computed. When setting the state during the resource
// Create, Read, or Update methods, this default value must also be included or
// the Terraform CLI will generate an error.
type listEmptyListModifier struct{}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to understand its impact.
func (m listEmptyListModifier) Description(ctx context.Context) string {
	return "If value is not configured, defaults to empty list"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
func (m listEmptyListModifier) MarkdownDescription(ctx context.Context) string {
	return "If value is not configured, defaults to empty list"
}

func (m listEmptyListModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if resp.PlanValue.IsNull() {
		val, diags := basetypes.NewListValue(types.StringType, []attr.Value{})
		resp.Diagnostics = append(resp.Diagnostics, diags...)
		if diags.HasError() {
			return
		}
		resp.PlanValue = val
	}
}

func EmptyList() planmodifier.List {
	return listEmptyListModifier{}
}
