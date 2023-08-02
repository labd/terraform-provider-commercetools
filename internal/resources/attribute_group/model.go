package attribute_group

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
	"reflect"
)

// AttributeGroup is the main resource schema data
type AttributeGroup struct {
	ID          types.String                     `tfsdk:"id"`
	Version     types.Int64                      `tfsdk:"version"`
	Key         types.String                     `tfsdk:"key"`
	Name        customtypes.LocalizedStringValue `tfsdk:"name"`
	Description customtypes.LocalizedStringValue `tfsdk:"description"`
	Attributes  []AttributeReference             `tfsdk:"attribute"`
}

type AttributeReference struct {
	Key types.String `tfsdk:"key"`
}

func toDraft(g *AttributeGroup) platform.AttributeGroupDraft {
	var attributes []platform.AttributeReference
	for _, v := range g.Attributes {
		attributes = append(attributes, platform.AttributeReference{
			Key: v.Key.ValueString(),
		})
	}

	return platform.AttributeGroupDraft{
		Name:        g.Name.ValueLocalizedString(),
		Description: g.Description.ValueLocalizedStringRef(),
		Key:         g.Key.ValueStringPointer(),
		Attributes:  attributes,
	}
}

func fromNative(n *platform.AttributeGroup) AttributeGroup {
	var attributes []AttributeReference
	for _, v := range n.Attributes {
		attributes = append(attributes, AttributeReference{
			Key: types.StringValue(v.Key),
		})
	}

	return AttributeGroup{
		ID:          types.StringValue(n.ID),
		Version:     types.Int64Value(int64(n.Version)),
		Key:         types.StringPointerValue(n.Key),
		Name:        utils.FromLocalizedString(n.Name),
		Description: utils.FromOptionalLocalizedString(n.Description),
		Attributes:  attributes,
	}
}

func toUpdateActions(state *AttributeGroup, plan *AttributeGroup) platform.AttributeGroupUpdate {
	update := platform.AttributeGroupUpdate{
		Version: int(state.Version.ValueInt64()),
		Actions: []platform.AttributeGroupUpdateAction{},
	}

	if !reflect.DeepEqual(state.Name, plan.Name) {
		update.Actions = append(
			update.Actions,
			platform.AttributeGroupChangeNameAction{
				Name: plan.Name.ValueLocalizedString(),
			},
		)
	}

	if !reflect.DeepEqual(state.Description, plan.Description) {
		update.Actions = append(
			update.Actions,
			platform.AttributeGroupSetDescriptionAction{
				Description: plan.Description.ValueLocalizedStringRef(),
			},
		)
	}

	if !reflect.DeepEqual(state.Key, plan.Key) {
		update.Actions = append(
			update.Actions,
			platform.AttributeGroupSetKeyAction{
				Key: utils.StringRef(plan.Key.ValueString()),
			},
		)
	}

	if !reflect.DeepEqual(state.Attributes, plan.Attributes) {
		var attributes []platform.AttributeReference
		for _, v := range plan.Attributes {
			attributes = append(attributes, platform.AttributeReference{
				Key: v.Key.ValueString(),
			})
		}

		update.Actions = append(
			update.Actions,
			platform.AttributeGroupSetAttributesAction{
				Attributes: attributes,
			},
		)
	}

	return update
}
