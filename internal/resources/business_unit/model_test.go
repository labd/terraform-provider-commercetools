package business_unit

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestBusinessUnit_Company_UpdateActions(t *testing.T) {
	cases := []struct {
		name     string
		state    Company
		plan     Company
		expected platform.BusinessUnitUpdate
	}{
		{
			"business unit update name",
			Company{
				Name: types.StringValue("Example business unit"),
			},
			Company{
				Name: types.StringValue("Updated business unit"),
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitChangeNameAction{
						Name: "Updated business unit",
					},
				},
			},
		},
		{
			"business unit update contact email",
			Company{
				ContactEmail: types.StringValue("info@example.com"),
			},
			Company{
				ContactEmail: types.StringValue("new@example.com"),
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitSetContactEmailAction{
						ContactEmail: types.StringValue("new@example.com").ValueStringPointer(),
					},
				},
			},
		},
		{
			"business unit update status",
			Company{
				Status: types.StringValue("Active"),
			},
			Company{
				Status: types.StringValue("Inactive"),
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitChangeStatusAction{
						Status: "Inactive",
					},
				},
			},
		},
		{
			"business unit update default shipping address",
			Company{
				DefaultShippingAddressID: types.StringValue("some-random-id"),
			},
			Company{
				DefaultShippingAddressID: types.StringValue("another-random-id"),
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitSetDefaultShippingAddressAction{
						AddressId: types.StringValue("another-random-id").ValueStringPointer(),
					},
				},
			},
		},
		{
			"business unit update default billing address",
			Company{
				DefaultBillingAddressID: types.StringValue("some-random-id"),
			},
			Company{
				DefaultBillingAddressID: types.StringValue("another-random-id"),
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitSetDefaultBillingAddressAction{
						AddressId: types.StringValue("another-random-id").ValueStringPointer(),
					},
				},
			},
		},
		{
			"business unit update stores",
			Company{
				Stores: []StoreKeyReference{
					{
						Key: types.StringValue("store-1"),
					},
					{
						Key: types.StringValue("store-2"),
					},
				},
			},
			Company{
				Stores: []StoreKeyReference{
					{
						Key: types.StringValue("store-1"),
					},
					{
						Key: types.StringValue("store-3"),
					},
				},
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitAddStoreAction{
						Store: platform.StoreResourceIdentifier{
							Key: types.StringValue("store-3").ValueStringPointer(),
							ID:  nil,
						},
					},
					platform.BusinessUnitRemoveStoreAction{
						Store: platform.StoreResourceIdentifier{
							Key: types.StringValue("store-2").ValueStringPointer(),
							ID:  nil,
						},
					},
				},
			},
		},
		{
			"business unit add address",
			Company{
				Addresses: []Address{},
			},
			Company{
				Addresses: []Address{
					{
						Key:                  types.StringValue("new-york-office"),
						Country:              types.StringValue("US"),
						Salutation:           types.StringValue("Mr."),
						FirstName:            types.StringValue("John"),
						LastName:             types.StringValue("Doe"),
						StreetName:           types.StringValue("Main St."),
						StreetNumber:         types.StringValue("123"),
						AdditionalStreetInfo: types.StringValue("Apt. 1"),
						PostalCode:           types.StringValue("12345"),
						City:                 types.StringValue("New York"),
						Region:               types.StringValue("New York"),
						State:                types.StringValue("New York"),
						Company:              types.StringValue("Example Inc."),
						Department:           types.StringValue("Sales"),
						Building:             types.StringValue("1"),
						Apartment:            types.StringValue("1"),
						POBox:                types.StringValue("123"),
						Phone:                types.StringValue("1234567890"),
						Mobile:               types.StringValue("1234567890"),
						Fax:                  types.StringValue("1234567890"),
					},
				},
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitAddAddressAction{
						Address: platform.BaseAddress{
							Key:                  utils.StringRef("new-york-office"),
							Country:              "US",
							Salutation:           utils.StringRef("Mr."),
							FirstName:            utils.StringRef("John"),
							LastName:             utils.StringRef("Doe"),
							StreetName:           utils.StringRef("Main St."),
							StreetNumber:         utils.StringRef("123"),
							AdditionalStreetInfo: utils.StringRef("Apt. 1"),
							PostalCode:           utils.StringRef("12345"),
							City:                 utils.StringRef("New York"),
							Region:               utils.StringRef("New York"),
							State:                utils.StringRef("New York"),
							Company:              utils.StringRef("Example Inc."),
							Department:           utils.StringRef("Sales"),
							Building:             utils.StringRef("1"),
							Apartment:            utils.StringRef("1"),
							POBox:                utils.StringRef("123"),
							Phone:                utils.StringRef("1234567890"),
							Mobile:               utils.StringRef("1234567890"),
							Fax:                  utils.StringRef("1234567890"),
						},
					},
				},
			},
		},
		{
			"business unit remove address",
			Company{
				Addresses: []Address{
					{
						Key:                  types.StringValue("new-york-office"),
						Country:              types.StringValue("US"),
						Salutation:           types.StringValue("Mr."),
						FirstName:            types.StringValue("John"),
						LastName:             types.StringValue("Doe"),
						StreetName:           types.StringValue("Main St."),
						StreetNumber:         types.StringValue("123"),
						AdditionalStreetInfo: types.StringValue("Apt. 1"),
						PostalCode:           types.StringValue("12345"),
						City:                 types.StringValue("New York"),
						Region:               types.StringValue("New York"),
						State:                types.StringValue("New York"),
						Company:              types.StringValue("Example Inc."),
						Department:           types.StringValue("Sales"),
						Building:             types.StringValue("1"),
						Apartment:            types.StringValue("1"),
						POBox:                types.StringValue("123"),
						Phone:                types.StringValue("1234567890"),
						Mobile:               types.StringValue("1234567890"),
						Fax:                  types.StringValue("1234567890"),
					},
				},
			},
			Company{
				Addresses: []Address{},
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitRemoveAddressAction{
						AddressKey: utils.StringRef("new-york-office"),
					},
				},
			},
		},
		{
			"business unit set associates",
			Company{},
			Company{
				Associates: []Associate{
					{
						AssociateRoleAssignments: []AssociateRoleAssignment{
							{
								AssociateRole: AssociateRoleKeyReference{
									Key: types.StringValue("role-1"),
								},
							},
						},
						Customer: CustomerReference{
							ID: types.StringValue("customer-1"),
						},
					},
				},
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitSetAssociatesAction{
						Associates: []platform.AssociateDraft{
							{
								AssociateRoleAssignments: []platform.AssociateRoleAssignmentDraft{
									{
										AssociateRole: platform.AssociateRoleResourceIdentifier{
											Key: utils.StringRef("role-1"),
										},
									},
								},
								Customer: platform.CustomerResourceIdentifier{
									ID: utils.StringRef("customer-1"),
								},
							},
						},
					},
				},
			},
		},
		{
			"business unit add billing address id",
			Company{
				BillingAddressIDs: []types.String{},
			},
			Company{
				BillingAddressIDs: []types.String{
					types.StringValue("new-york-office"),
				},
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitAddBillingAddressIdAction{
						AddressId: utils.StringRef("new-york-office"),
					},
				},
			},
		},
		{
			"business unit remove billing address id",
			Company{
				BillingAddressIDs: []types.String{
					types.StringValue("new-york-office"),
				},
			},
			Company{
				BillingAddressIDs: []types.String{},
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitRemoveBillingAddressIdAction{
						AddressId: utils.StringRef("new-york-office"),
					},
				},
			},
		},
		{
			"business unit add shipping address id",
			Company{
				ShippingAddressIDs: []types.String{},
			},
			Company{
				ShippingAddressIDs: []types.String{
					types.StringValue("new-york-office"),
				},
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitAddShippingAddressIdAction{
						AddressId: utils.StringRef("new-york-office"),
					},
				},
			},
		},
		{
			"business unit remove shipping address id",
			Company{
				ShippingAddressIDs: []types.String{
					types.StringValue("new-york-office"),
				},
			},
			Company{
				ShippingAddressIDs: []types.String{},
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitRemoveShippingAddressIdAction{
						AddressId: utils.StringRef("new-york-office"),
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

func TestBusinessUnit_Division_UpdateActions(t *testing.T) {
	cases := []struct {
		name     string
		state    Division
		plan     Division
		expected platform.BusinessUnitUpdate
	}{
		{
			"business unit update name",
			Division{
				Name: types.StringValue("Example business unit"),
			},
			Division{
				Name: types.StringValue("Updated business unit"),
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitChangeNameAction{
						Name: "Updated business unit",
					},
				},
			},
		},
		{
			"business unit update contact email",
			Division{
				ContactEmail: types.StringValue("info@example.com"),
			},
			Division{
				ContactEmail: types.StringValue("new@example.com"),
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitSetContactEmailAction{
						ContactEmail: types.StringValue("new@example.com").ValueStringPointer(),
					},
				},
			},
		},
		{
			"business unit update status",
			Division{
				Status: types.StringValue("Active"),
			},
			Division{
				Status: types.StringValue("Inactive"),
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitChangeStatusAction{
						Status: "Inactive",
					},
				},
			},
		},
		{
			"business unit update default shipping address",
			Division{
				DefaultShippingAddressID: types.StringValue("some-random-id"),
			},
			Division{
				DefaultShippingAddressID: types.StringValue("another-random-id"),
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitSetDefaultShippingAddressAction{
						AddressId: types.StringValue("another-random-id").ValueStringPointer(),
					},
				},
			},
		},
		{
			"business unit update default billing address",
			Division{
				DefaultBillingAddressID: types.StringValue("some-random-id"),
			},
			Division{
				DefaultBillingAddressID: types.StringValue("another-random-id"),
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitSetDefaultBillingAddressAction{
						AddressId: types.StringValue("another-random-id").ValueStringPointer(),
					},
				},
			},
		},
		{
			"business unit update associate mode",
			Division{
				AssociateMode: types.StringValue("Explicit"),
			},
			Division{
				AssociateMode: types.StringValue("ExplicitAndFromParent"),
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitChangeAssociateModeAction{
						AssociateMode: "ExplicitAndFromParent",
					},
				},
			},
		},
		{
			"business unit update stores",
			Division{
				Stores: []StoreKeyReference{
					{
						Key: types.StringValue("store-1"),
					},
					{
						Key: types.StringValue("store-2"),
					},
				},
			},
			Division{
				Stores: []StoreKeyReference{
					{
						Key: types.StringValue("store-1"),
					},
					{
						Key: types.StringValue("store-3"),
					},
				},
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitAddStoreAction{
						Store: platform.StoreResourceIdentifier{
							Key: types.StringValue("store-3").ValueStringPointer(),
							ID:  nil,
						},
					},
					platform.BusinessUnitRemoveStoreAction{
						Store: platform.StoreResourceIdentifier{
							Key: types.StringValue("store-2").ValueStringPointer(),
							ID:  nil,
						},
					},
				},
			},
		},
		{
			"business unit add address",
			Division{
				Addresses: []Address{},
			},
			Division{
				Addresses: []Address{
					{
						Key:                  types.StringValue("new-york-office"),
						Country:              types.StringValue("US"),
						Salutation:           types.StringValue("Mr."),
						FirstName:            types.StringValue("John"),
						LastName:             types.StringValue("Doe"),
						StreetName:           types.StringValue("Main St."),
						StreetNumber:         types.StringValue("123"),
						AdditionalStreetInfo: types.StringValue("Apt. 1"),
						PostalCode:           types.StringValue("12345"),
						City:                 types.StringValue("New York"),
						Region:               types.StringValue("New York"),
						State:                types.StringValue("New York"),
						Company:              types.StringValue("Example Inc."),
						Department:           types.StringValue("Sales"),
						Building:             types.StringValue("1"),
						Apartment:            types.StringValue("1"),
						POBox:                types.StringValue("123"),
						Phone:                types.StringValue("1234567890"),
						Mobile:               types.StringValue("1234567890"),
						Fax:                  types.StringValue("1234567890"),
					},
				},
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitAddAddressAction{
						Address: platform.BaseAddress{
							Key:                  utils.StringRef("new-york-office"),
							Country:              "US",
							Salutation:           utils.StringRef("Mr."),
							FirstName:            utils.StringRef("John"),
							LastName:             utils.StringRef("Doe"),
							StreetName:           utils.StringRef("Main St."),
							StreetNumber:         utils.StringRef("123"),
							AdditionalStreetInfo: utils.StringRef("Apt. 1"),
							PostalCode:           utils.StringRef("12345"),
							City:                 utils.StringRef("New York"),
							Region:               utils.StringRef("New York"),
							State:                utils.StringRef("New York"),
							Company:              utils.StringRef("Example Inc."),
							Department:           utils.StringRef("Sales"),
							Building:             utils.StringRef("1"),
							Apartment:            utils.StringRef("1"),
							POBox:                utils.StringRef("123"),
							Phone:                utils.StringRef("1234567890"),
							Mobile:               utils.StringRef("1234567890"),
							Fax:                  utils.StringRef("1234567890"),
						},
					},
				},
			},
		},
		{
			"business unit remove address",
			Division{
				Addresses: []Address{
					{
						Key:                  types.StringValue("new-york-office"),
						Country:              types.StringValue("US"),
						Salutation:           types.StringValue("Mr."),
						FirstName:            types.StringValue("John"),
						LastName:             types.StringValue("Doe"),
						StreetName:           types.StringValue("Main St."),
						StreetNumber:         types.StringValue("123"),
						AdditionalStreetInfo: types.StringValue("Apt. 1"),
						PostalCode:           types.StringValue("12345"),
						City:                 types.StringValue("New York"),
						Region:               types.StringValue("New York"),
						State:                types.StringValue("New York"),
						Company:              types.StringValue("Example Inc."),
						Department:           types.StringValue("Sales"),
						Building:             types.StringValue("1"),
						Apartment:            types.StringValue("1"),
						POBox:                types.StringValue("123"),
						Phone:                types.StringValue("1234567890"),
						Mobile:               types.StringValue("1234567890"),
						Fax:                  types.StringValue("1234567890"),
					},
				},
			},
			Division{
				Addresses: []Address{},
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitRemoveAddressAction{
						AddressKey: utils.StringRef("new-york-office"),
					},
				},
			},
		},
		{
			"business unit set associates",
			Division{},
			Division{
				Associates: []Associate{
					{
						AssociateRoleAssignments: []AssociateRoleAssignment{
							{
								AssociateRole: AssociateRoleKeyReference{
									Key: types.StringValue("role-1"),
								},
							},
						},
						Customer: CustomerReference{
							ID: types.StringValue("customer-1"),
						},
					},
				},
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitSetAssociatesAction{
						Associates: []platform.AssociateDraft{
							{
								AssociateRoleAssignments: []platform.AssociateRoleAssignmentDraft{
									{
										AssociateRole: platform.AssociateRoleResourceIdentifier{
											Key: utils.StringRef("role-1"),
										},
									},
								},
								Customer: platform.CustomerResourceIdentifier{
									ID: utils.StringRef("customer-1"),
								},
							},
						},
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
