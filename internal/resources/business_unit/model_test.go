package business_unit

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestBusinessUnit_UpdateActions(t *testing.T) {
	cases := []struct {
		name     string
		state    BusinessUnit
		plan     BusinessUnit
		expected platform.BusinessUnitUpdate
	}{
		{
			"business unit update name",
			BusinessUnit{
				Name: types.StringValue("Example business unit"),
			},
			BusinessUnit{
				Name: types.StringValue("Example business unit 2"),
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitChangeNameAction{
						Name: "Example business unit 2",
					},
				},
			},
		},
		{
			"business unit update status",
			BusinessUnit{
				Status: types.StringValue("Active"),
			},
			BusinessUnit{
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
			"business unit add a new store",
			BusinessUnit{
				StoreMode: types.StringValue(StoreModeExplicit),
				Stores: []StoreKeyReference{
					{
						Key: types.StringValue("store-1"),
					},
				},
			},
			BusinessUnit{
				StoreMode: types.StringValue(StoreModeExplicit),
				Stores: []StoreKeyReference{
					{
						Key: types.StringValue("store-1"),
					},
					{
						Key: types.StringValue("store-2"),
					},
				},
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitSetStoresAction{
						Stores: []platform.StoreResourceIdentifier{
							{
								Key: utils.StringRef("store-1"),
							},
							{
								Key: utils.StringRef("store-2"),
							},
						},
					},
				},
			},
		},
		{
			"business unit remove a store",
			BusinessUnit{
				StoreMode: types.StringValue(StoreModeExplicit),
				Stores: []StoreKeyReference{
					{
						Key: types.StringValue("store-1"),
					},
					{
						Key: types.StringValue("store-2"),
					},
				},
			},
			BusinessUnit{
				StoreMode: types.StringValue(StoreModeExplicit),
				Stores: []StoreKeyReference{
					{
						Key: types.StringValue("store-1"),
					},
				},
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitSetStoresAction{
						Stores: []platform.StoreResourceIdentifier{
							{
								Key: utils.StringRef("store-1"),
							},
						},
					},
				},
			},
		},
		{
			"business unit change store mode to explicit",
			BusinessUnit{
				StoreMode: types.StringValue(StoreModeFromParent),
			},
			BusinessUnit{
				StoreMode: types.StringValue(StoreModeExplicit),
				Stores: []StoreKeyReference{
					{
						Key: types.StringValue("store-1"),
					},
					{
						Key: types.StringValue("store-2"),
					},
				},
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitSetStoreModeAction{
						StoreMode: StoreModeExplicit,
						Stores: []platform.StoreResourceIdentifier{
							{
								Key: utils.StringRef("store-1"),
							},
							{
								Key: utils.StringRef("store-2"),
							},
						},
					},
				},
			},
		},
		{
			"business unit change contact email",
			BusinessUnit{
				ContactEmail: types.StringValue("current@example.com"),
			},
			BusinessUnit{
				ContactEmail: types.StringValue("new@example.com"),
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitSetContactEmailAction{
						ContactEmail: utils.StringRef("new@example.com"),
					},
				},
			},
		},
		{
			"business unit change associate mode",
			BusinessUnit{
				AssociateMode: types.StringValue(ExplicitAndFromParentAssociateMode),
			},
			BusinessUnit{
				AssociateMode: types.StringValue(ExplicitAssociateMode),
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitChangeAssociateModeAction{
						AssociateMode: ExplicitAssociateMode,
					},
				},
			},
		},
		{
			"business unit add address",
			BusinessUnit{
				Addresses: []Address{},
			},
			BusinessUnit{
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
			BusinessUnit{
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
			BusinessUnit{
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
			"business unit update default shipping address",
			BusinessUnit{
				DefaultShippingAddressID: types.StringValue("new-york-office"),
			},
			BusinessUnit{
				DefaultShippingAddressID: types.StringValue("new-york-office-2"),
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitSetDefaultShippingAddressAction{
						AddressKey: utils.StringRef("new-york-office-2"),
					},
				},
			},
		},
		{
			"business unit update default billing address",
			BusinessUnit{
				DefaultBillingAddressID: types.StringValue("new-york-office"),
			},
			BusinessUnit{
				DefaultBillingAddressID: types.StringValue("new-york-office-2"),
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitSetDefaultBillingAddressAction{
						AddressKey: utils.StringRef("new-york-office-2"),
					},
				},
			},
		},
		{
			"business unit add address",
			BusinessUnit{
				Addresses: []Address{},
			},
			BusinessUnit{
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
			"business unit add shipping address id",
			BusinessUnit{
				ShippingAddressIDs: []types.String{},
			},
			BusinessUnit{
				ShippingAddressIDs: []types.String{
					types.StringValue("new-york-office"),
				},
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitAddShippingAddressIdAction{
						AddressKey: utils.StringRef("new-york-office"),
					},
				},
			},
		},
		{
			"business unit remove shipping address id",
			BusinessUnit{
				ShippingAddressIDs: []types.String{
					types.StringValue("new-york-office"),
				},
			},
			BusinessUnit{
				ShippingAddressIDs: []types.String{},
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitRemoveShippingAddressIdAction{
						AddressKey: utils.StringRef("new-york-office"),
					},
				},
			},
		},
		{
			"business unit add billing address id",
			BusinessUnit{
				BillingAddressIDs: []types.String{},
			},
			BusinessUnit{
				BillingAddressIDs: []types.String{
					types.StringValue("new-york-office"),
				},
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitAddBillingAddressIdAction{
						AddressKey: utils.StringRef("new-york-office"),
					},
				},
			},
		},
		{
			"business unit remove billing address id",
			BusinessUnit{
				BillingAddressIDs: []types.String{
					types.StringValue("new-york-office"),
				},
			},
			BusinessUnit{
				BillingAddressIDs: []types.String{},
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitRemoveBillingAddressIdAction{
						AddressKey: utils.StringRef("new-york-office"),
					},
				},
			},
		},
		{
			"business unit set associates",
			BusinessUnit{},
			BusinessUnit{
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

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			updateActions := tc.state.updateActions(tc.plan)
			assert.Equal(t, tc.expected, updateActions)
		})
	}
}
