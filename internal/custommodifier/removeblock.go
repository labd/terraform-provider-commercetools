package custommodifier

// From https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#creating-attribute-plan-modifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

type removeBlockModifier struct{}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to understand its impact.
func (m removeBlockModifier) Description(ctx context.Context) string {
	return fmt.Sprintf("TODO")
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
func (m removeBlockModifier) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("TODO")
}

func (m removeBlockModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	attrs := req.ConfigValue.Attributes()
	if len(attrs) == 0 {
		resp.PlanValue = req.ConfigValue
	}
}

func RemoveBlockModifier() planmodifier.Object {
	return removeBlockModifier{}
}
