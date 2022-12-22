package custommodifier

// From https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#creating-attribute-plan-modifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type createOnlyModifier struct{}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to understand its impact.
func (m createOnlyModifier) Description(ctx context.Context) string {
	return fmt.Sprintf("TODO")
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
func (m createOnlyModifier) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("TODO")
}

func (m createOnlyModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If the value is unknown or known, do not set default value.
	if req.PlanValue.IsNull() || resp.PlanValue.IsUnknown() {
		return
	}

	if !req.StateValue.IsUnknown() {
		resp.PlanValue = types.StringUnknown()
	}

}

func CreateOnly() planmodifier.String {
	return createOnlyModifier{}
}
