package attribute_group

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"
)

func TestToDraft(t *testing.T) {
	tests := []struct {
		name string
		res  *platform.AttributeGroupDraft
		want *AttributeGroup
	}{
		{
			name: "Default",
			res: &platform.AttributeGroupDraft{
				Key:         utils.StringRef("attribute-group-key"),
				Name:        map[string]string{"nl-NL": "attribute-group-name"},
				Description: &platform.LocalizedString{"nl-NL": "attribute-group-description"},
				Attributes:  []platform.AttributeReference{{Key: "attribute-key"}},
			},
			want: &AttributeGroup{
				Version: types.Int64Value(1),
				Key:     types.StringValue("attribute-group-key"),
				Name: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"nl-NL": types.StringValue("attribute-group-name"),
				}),
				Description: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"nl-NL": types.StringValue("attribute-group-description"),
				}),
				Attributes: []AttributeReference{
					{Key: types.StringValue("attribute-key")},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toDraft(tt.want)
			assert.Equal(t, tt.res, &result)
		})
	}
}

func TestFromNative(t *testing.T) {
	tests := []struct {
		name string
		res  *AttributeGroup
		want *platform.AttributeGroup
	}{
		{
			name: "Default",
			want: &platform.AttributeGroup{
				ID:          "attribute-group-id",
				Version:     1,
				Key:         utils.StringRef("attribute-group-key"),
				Name:        map[string]string{"nl-NL": "attribute-group-name"},
				Description: &platform.LocalizedString{"nl-NL": "attribute-group-description"},
				Attributes:  []platform.AttributeReference{{Key: "attribute-key"}},
			},
			res: &AttributeGroup{
				ID:      types.StringValue("attribute-group-id"),
				Version: types.Int64Value(1),
				Key:     types.StringValue("attribute-group-key"),
				Name: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"nl-NL": types.StringValue("attribute-group-name"),
				}),
				Description: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"nl-NL": types.StringValue("attribute-group-description"),
				}),
				Attributes: []AttributeReference{
					{Key: types.StringValue("attribute-key")},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fromNative(tt.want)
			assert.Equal(t, tt.res, &result)
		})
	}
}

func TestToUpdateActions(t *testing.T) {
	tests := []struct {
		name   string
		state  AttributeGroup
		plan   AttributeGroup
		action platform.AttributeGroupUpdate
	}{
		{
			name: "Update Key",
			state: AttributeGroup{
				Version: types.Int64Value(1),
				Key:     types.StringValue("attribute-group-key"),
			},
			plan: AttributeGroup{
				Key: types.StringValue("new-attribute-group-key"),
			},
			action: platform.AttributeGroupUpdate{
				Version: 1,
				Actions: []platform.AttributeGroupUpdateAction{
					platform.AttributeGroupSetKeyAction{
						Key: utils.StringRef("new-attribute-group-key"),
					},
				},
			},
		},
		{
			name: "Update Description",
			state: AttributeGroup{
				Version: types.Int64Value(1),
				Description: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"nl-NL": types.StringValue("attribute-group-description"),
				}),
			},
			plan: AttributeGroup{
				Description: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"nl-NL": types.StringValue("new-attribute-group-description"),
				}),
			},
			action: platform.AttributeGroupUpdate{
				Version: 1,
				Actions: []platform.AttributeGroupUpdateAction{
					platform.AttributeGroupSetDescriptionAction{
						Description: &platform.LocalizedString{"nl-NL": "new-attribute-group-description"},
					},
				},
			},
		},
		{
			name: "Update Name",
			state: AttributeGroup{
				Version: types.Int64Value(1),
				Name: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"nl-NL": types.StringValue("attribute-group-name"),
				}),
			},
			plan: AttributeGroup{
				Name: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"nl-NL": types.StringValue("new-attribute-group-name"),
				}),
			},
			action: platform.AttributeGroupUpdate{
				Version: 1,
				Actions: []platform.AttributeGroupUpdateAction{
					platform.AttributeGroupChangeNameAction{
						Name: map[string]string{"nl-NL": "new-attribute-group-name"},
					},
				},
			},
		},
		{
			name: "Update Attributes",
			state: AttributeGroup{
				Version: types.Int64Value(1),
				Attributes: []AttributeReference{
					{Key: types.StringValue("attribute-group-key-1")},
				},
			},
			plan: AttributeGroup{
				Attributes: []AttributeReference{
					{Key: types.StringValue("attribute-group-key-1")},
					{Key: types.StringValue("attribute-group-key-2")},
				},
			},
			action: platform.AttributeGroupUpdate{
				Version: 1,
				Actions: []platform.AttributeGroupUpdateAction{
					platform.AttributeGroupSetAttributesAction{
						Attributes: []platform.AttributeReference{
							{Key: "attribute-group-key-1"},
							{Key: "attribute-group-key-2"},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toUpdateActions(&tt.state, &tt.plan)
			assert.Equal(t, tt.action, result)
		})
	}
}
