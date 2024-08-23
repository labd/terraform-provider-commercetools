package business_unit_company

import (
	"github.com/labd/terraform-provider-commercetools/internal/sharedtypes"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestBusinessUnit_Company_Draft(t *testing.T) {
	cases := []struct {
		name     string
		company  Company
		expected platform.CompanyDraft
	}{
		{
			name: "Basic company draft",
			company: Company{
				Key:          types.StringValue("company-key"),
				Status:       types.StringValue("Active"),
				Name:         types.StringValue("Company Name"),
				ContactEmail: types.StringValue("contact@example.com"),
				Addresses: []sharedtypes.Address{
					{
						Key:     types.StringValue("address-1"),
						Country: types.StringValue("US"),
						City:    types.StringValue("New York"),
					},
					{
						Key:     types.StringValue("address-2"),
						Country: types.StringValue("US"),
						City:    types.StringValue("Detroit"),
					},
				},
				Stores: []sharedtypes.StoreKeyReference{
					{Key: types.StringValue("store-1")},
				},
				ShippingAddressKeys:       []types.String{types.StringValue("address-1"), types.StringValue("address-2")},
				BillingAddressKeys:        []types.String{types.StringValue("address-1"), types.StringValue("address-2")},
				DefaultBillingAddressKey:  types.StringValue("address-2"),
				DefaultShippingAddressKey: types.StringValue("address-2"),
			},
			expected: platform.CompanyDraft{
				Key:              "company-key",
				Status:           utils.Ref(platform.BusinessUnitStatusActive),
				Name:             "Company Name",
				StoreMode:        utils.Ref(platform.BusinessUnitStoreModeExplicit),
				AssociateMode:    utils.Ref(platform.BusinessUnitAssociateModeExplicit),
				ApprovalRuleMode: utils.Ref(platform.BusinessUnitApprovalRuleModeExplicit),
				ContactEmail:     utils.Ref("contact@example.com"),
				Addresses: []platform.BaseAddress{
					{
						Key:     utils.Ref("address-1"),
						Country: "US",
						City:    utils.Ref("New York"),
					},
					{
						Key:     utils.Ref("address-2"),
						Country: "US",
						City:    utils.Ref("Detroit"),
					},
				},
				Stores: []platform.StoreResourceIdentifier{
					{
						Key: utils.Ref("store-1"),
					},
				},
				DefaultShippingAddress: utils.Ref(1),
				DefaultBillingAddress:  utils.Ref(1),
				ShippingAddresses:      []int{0, 1},
				BillingAddresses:       []int{0, 1},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result, err := c.company.draft()
			assert.NoError(t, err)
			assert.Equal(t, c.expected, result)
		})
	}
}

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
				ShippingAddressKeys:       []types.String{types.StringValue("some-random-id"), types.StringValue("another-random-id")},
				DefaultShippingAddressKey: types.StringValue("some-random-id"),
			},
			Company{
				ShippingAddressKeys:       []types.String{types.StringValue("some-random-id"), types.StringValue("another-random-id")},
				DefaultShippingAddressKey: types.StringValue("another-random-id"),
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitSetDefaultShippingAddressAction{
						AddressKey: types.StringValue("another-random-id").ValueStringPointer(),
					},
				},
			},
		},
		{
			"business unit update default billing address",
			Company{
				BillingAddressKeys:       []types.String{types.StringValue("some-random-id"), types.StringValue("another-random-id")},
				DefaultBillingAddressKey: types.StringValue("some-random-id"),
			},
			Company{
				BillingAddressKeys:       []types.String{types.StringValue("some-random-id"), types.StringValue("another-random-id")},
				DefaultBillingAddressKey: types.StringValue("another-random-id"),
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitSetDefaultBillingAddressAction{
						AddressKey: types.StringValue("another-random-id").ValueStringPointer(),
					},
				},
			},
		},
		{
			"business unit update stores",
			Company{
				Stores: []sharedtypes.StoreKeyReference{
					{
						Key: types.StringValue("store-1"),
					},
					{
						Key: types.StringValue("store-2"),
					},
				},
			},
			Company{
				Stores: []sharedtypes.StoreKeyReference{
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
					platform.BusinessUnitSetStoresAction{
						Stores: []platform.StoreResourceIdentifier{
							{
								Key: types.StringValue("store-1").ValueStringPointer(),
								ID:  nil,
							},
							{
								Key: types.StringValue("store-3").ValueStringPointer(),
								ID:  nil,
							},
						},
					},
				},
			},
		},
		{
			"business unit add address",
			Company{
				Addresses: []sharedtypes.Address{},
			},
			Company{
				Addresses: []sharedtypes.Address{
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
				Addresses: []sharedtypes.Address{
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
				Addresses: []sharedtypes.Address{},
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
			"business unit add billing address id",
			Company{
				BillingAddressKeys: []types.String{},
			},
			Company{
				BillingAddressKeys: []types.String{
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
			Company{
				BillingAddressKeys: []types.String{
					types.StringValue("new-york-office"),
				},
			},
			Company{
				BillingAddressKeys: []types.String{},
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
			"business unit add shipping address id",
			Company{
				ShippingAddressKeys: []types.String{},
			},
			Company{
				ShippingAddressKeys: []types.String{
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
			Company{
				ShippingAddressKeys: []types.String{
					types.StringValue("new-york-office"),
				},
			},
			Company{
				ShippingAddressKeys: []types.String{},
			},
			platform.BusinessUnitUpdate{
				Actions: []platform.BusinessUnitUpdateAction{
					platform.BusinessUnitRemoveShippingAddressIdAction{
						AddressKey: utils.StringRef("new-york-office"),
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result, err := c.state.updateActions(c.plan)
			assert.NoError(t, err)
			assert.EqualValues(t, c.expected, result)
		})
	}
}

func TestBusinessUnit_Company_NewCompanyFromNative(t *testing.T) {
	cases := []struct {
		name     string
		company  map[string]interface{}
		expected Company
	}{
		{
			name: "Basic company draft",
			company: map[string]interface{}{
				"id":           "company-id",
				"key":          "company-key",
				"version":      1,
				"status":       platform.BusinessUnitStatusActive,
				"name":         "Company Name",
				"contactEmail": utils.Ref("contact@example.com"),
				"addresses": []map[string]interface{}{
					{
						"id":      utils.Ref("address-id-1"),
						"key":     utils.Ref("address-1"),
						"country": "US",
						"city":    utils.Ref("New York"),
					},
					{
						"id":      utils.Ref("address-id-2"),
						"key":     utils.Ref("address-2"),
						"country": "US",
						"city":    utils.Ref("Detroit"),
					},
				},
				"stores": []map[string]interface{}{
					{"key": "store-1"},
				},
				"shippingAddressIds":       []string{"address-id-1", "address-id-2"},
				"billingAddressIds":        []string{"address-id-1", "address-id-2"},
				"defaultBillingAddressId":  utils.Ref("address-id-2"),
				"defaultShippingAddressId": utils.Ref("address-id-2"),
			},
			expected: Company{
				ID:           types.StringValue("company-id"),
				Key:          types.StringValue("company-key"),
				Version:      types.Int64Value(1),
				Status:       types.StringValue("Active"),
				Name:         types.StringValue("Company Name"),
				ContactEmail: types.StringValue("contact@example.com"),
				Addresses: []sharedtypes.Address{
					{
						ID:      types.StringValue("address-id-1"),
						Key:     types.StringValue("address-1"),
						Country: types.StringValue("US"),
						City:    types.StringValue("New York"),
					},
					{
						ID:      types.StringValue("address-id-2"),
						Key:     types.StringValue("address-2"),
						Country: types.StringValue("US"),
						City:    types.StringValue("Detroit"),
					},
				},
				Stores: []sharedtypes.StoreKeyReference{
					{Key: types.StringValue("store-1")},
				},
				ShippingAddressKeys:       []types.String{types.StringValue("address-1"), types.StringValue("address-2")},
				BillingAddressKeys:        []types.String{types.StringValue("address-1"), types.StringValue("address-2")},
				DefaultBillingAddressKey:  types.StringValue("address-2"),
				DefaultShippingAddressKey: types.StringValue("address-2"),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var data platform.BusinessUnit
			err := utils.DecodeStruct(c.company, &data)
			assert.NoError(t, err)

			result, err := NewCompanyFromNative(&data)
			assert.NoError(t, err)
			assert.Equal(t, c.expected, result)
		})
	}
}
