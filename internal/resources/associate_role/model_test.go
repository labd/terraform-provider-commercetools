package associate_role

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestAssociateRole_UpdateActions(t *testing.T) {
	cases := []struct {
		name     string
		state    AssociateRole
		plan     AssociateRole
		expected platform.AssociateRoleUpdate
	}{
		{
			"associate role update permissions",
			AssociateRole{
				Permissions: []types.String{
					types.StringValue("CreateMyCarts"),
					types.StringValue("DeleteMyCarts"),
					types.StringValue("UpdateMyCarts"),
					types.StringValue("ViewMyCarts"),
				},
			},
			AssociateRole{
				Permissions: []types.String{
					types.StringValue("AddChildUnits"),
					types.StringValue("UpdateBusinessUnitDetails"),
					types.StringValue("UpdateAssociates"),
					types.StringValue("CreateMyCarts"),
					types.StringValue("DeleteMyCarts"),
					types.StringValue("UpdateMyCarts"),
					types.StringValue("ViewMyCarts"),
				},
			},
			platform.AssociateRoleUpdate{
				Actions: []platform.AssociateRoleUpdateAction{
					platform.AssociateRoleSetPermissionsAction{
						Permissions: []platform.Permission{
							"AddChildUnits",
							"UpdateBusinessUnitDetails",
							"UpdateAssociates",
							"CreateMyCarts",
							"DeleteMyCarts",
							"UpdateMyCarts",
							"ViewMyCarts",
						},
					},
				},
			},
		},
		{
			"associate role update name",
			AssociateRole{
				Name: types.StringValue("Example associate role"),
			},
			AssociateRole{
				Name: types.StringValue("Example manager associate role"),
			},
			platform.AssociateRoleUpdate{
				Actions: []platform.AssociateRoleUpdateAction{
					platform.AssociateRoleSetNameAction{
						Name: utils.StringRef("Example manager associate role"),
					},
				},
			},
		},
		{
			"associate role update buyer assignable",
			AssociateRole{
				BuyerAssignable: types.BoolValue(true),
			},
			AssociateRole{
				BuyerAssignable: types.BoolValue(false),
			},
			platform.AssociateRoleUpdate{
				Actions: []platform.AssociateRoleUpdateAction{
					platform.AssociateRoleChangeBuyerAssignableAction{
						BuyerAssignable: false,
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := c.state.updateActions(c.plan)
			assert.EqualValues(t, c.expected, result)
		})
	}
}

func TestNewAssociateRoleFromNative(t *testing.T) {
	cases := []struct {
		name   string
		res    *platform.AssociateRole
		expect AssociateRole
	}{
		{
			"decode remote associate role representation into local resource",
			&platform.AssociateRole{
				ID:              "rand-uuid-or-other-string",
				Version:         1,
				Key:             "sales_manager_europe_associate_role",
				BuyerAssignable: false,
				Name:            utils.StringRef("Sales Manager - Europe"),
				Permissions: []platform.Permission{
					"AddChildUnits",
					"UpdateBusinessUnitDetails",
					"UpdateAssociates",
					"CreateMyCarts",
					"DeleteMyCarts",
					"UpdateMyCarts",
					"ViewMyCarts",
				},
			},
			AssociateRole{
				ID:              types.StringValue("rand-uuid-or-other-string"),
				Key:             types.StringValue("sales_manager_europe_associate_role"),
				Version:         types.Int64Value(1),
				Name:            types.StringValue("Sales Manager - Europe"),
				BuyerAssignable: types.BoolValue(false),
				Permissions: []types.String{
					types.StringValue("AddChildUnits"),
					types.StringValue("UpdateBusinessUnitDetails"),
					types.StringValue("UpdateAssociates"),
					types.StringValue("CreateMyCarts"),
					types.StringValue("DeleteMyCarts"),
					types.StringValue("UpdateMyCarts"),
					types.StringValue("ViewMyCarts"),
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := NewAssociateRoleFromNative(c.res)
			assert.EqualValues(t, got, c.expect)
		})
	}
}
