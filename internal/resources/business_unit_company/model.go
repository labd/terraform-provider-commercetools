package business_unit_company

import (
	"fmt"
	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/internal/sharedtypes"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
	"reflect"
	"slices"
	"sort"
)

// Company is a type to model the fields that all types of Companies have in common.
type Company struct {
	ID                        types.String                    `tfsdk:"id"`
	Version                   types.Int64                     `tfsdk:"version"`
	Key                       types.String                    `tfsdk:"key"`
	Status                    types.String                    `tfsdk:"status"`
	Name                      types.String                    `tfsdk:"name"`
	ContactEmail              types.String                    `tfsdk:"contact_email"`
	ShippingAddressKeys       []types.String                  `tfsdk:"shipping_address_keys"`
	DefaultShippingAddressKey types.String                    `tfsdk:"default_shipping_address_key"`
	BillingAddressKeys        []types.String                  `tfsdk:"billing_address_keys"`
	DefaultBillingAddressKey  types.String                    `tfsdk:"default_billing_address_key"`
	Stores                    []sharedtypes.StoreKeyReference `tfsdk:"store"`
	Addresses                 []sharedtypes.Address           `tfsdk:"address"`
}

func (c *Company) draft() (platform.CompanyDraft, error) {
	status := platform.BusinessUnitStatus(c.Status.ValueString())
	storeMode := platform.BusinessUnitStoreModeExplicit
	associateMode := platform.BusinessUnitAssociateModeExplicit
	approvalRuleMode := platform.BusinessUnitApprovalRuleModeExplicit

	var addresses []platform.BaseAddress
	for _, a := range c.Addresses {
		addresses = append(addresses, a.Draft())
	}

	var stores []platform.StoreResourceIdentifier
	for _, s := range c.Stores {
		stores = append(stores, platform.StoreResourceIdentifier{
			Key: s.Key.ValueStringPointer(),
		})
	}

	var shippingAddressIndexes []int
	for _, key := range c.ShippingAddressKeys {
		i := slices.IndexFunc(c.Addresses, func(a sharedtypes.Address) bool {
			return a.Key.ValueString() == key.ValueString()
		})

		if i == -1 {
			return platform.CompanyDraft{}, fmt.Errorf("shipping address key %s is not in addresses", key.ValueString())
		}

		shippingAddressIndexes = append(shippingAddressIndexes, i)
	}

	var billingAddressIndexes []int
	for _, key := range c.BillingAddressKeys {
		i := slices.IndexFunc(c.Addresses, func(a sharedtypes.Address) bool {
			return a.Key.ValueString() == key.ValueString()
		})

		if i == -1 {
			return platform.CompanyDraft{}, fmt.Errorf("billing address key %s is not in addresses", key.ValueString())
		}

		billingAddressIndexes = append(billingAddressIndexes, i)
	}

	var defaultBillingAddressIndex *int
	if !c.DefaultBillingAddressKey.IsNull() {
		i := slices.IndexFunc(c.Addresses, func(a sharedtypes.Address) bool {
			return a.Key.ValueString() == c.DefaultBillingAddressKey.ValueString()
		})

		if i == -1 {
			return platform.CompanyDraft{}, fmt.Errorf("default billing address key %s is not in addresses", c.DefaultBillingAddressKey.ValueString())
		}

		defaultBillingAddressIndex = &i
	}

	var defaultShippingAddressIndex *int
	if !c.DefaultShippingAddressKey.IsNull() {
		i := slices.IndexFunc(c.Addresses, func(a sharedtypes.Address) bool {
			return a.Key.ValueString() == c.DefaultShippingAddressKey.ValueString()
		})

		if i == -1 {
			return platform.CompanyDraft{}, fmt.Errorf("default shipping address key %s is not in addresses", c.DefaultShippingAddressKey.ValueString())
		}

		defaultShippingAddressIndex = &i
	}

	return platform.CompanyDraft{
		Key:                    c.Key.ValueString(),
		Status:                 &status,
		StoreMode:              &storeMode,
		AssociateMode:          &associateMode,
		ApprovalRuleMode:       &approvalRuleMode,
		Stores:                 stores,
		Name:                   c.Name.ValueString(),
		ContactEmail:           c.ContactEmail.ValueStringPointer(),
		Addresses:              addresses,
		ShippingAddresses:      shippingAddressIndexes,
		BillingAddresses:       billingAddressIndexes,
		DefaultShippingAddress: defaultShippingAddressIndex,
		DefaultBillingAddress:  defaultBillingAddressIndex,
	}, nil
}

func (c *Company) updateActions(plan Company) (platform.BusinessUnitUpdate, error) {
	result := platform.BusinessUnitUpdate{
		Version: int(c.Version.ValueInt64()),
		Actions: []platform.BusinessUnitUpdateAction{},
	}

	if !c.Key.Equal(plan.Key) {
		return result, fmt.Errorf("key is immutable. Delete this resource instead if a change is intended")
	}

	if !c.Name.Equal(plan.Name) {
		result.Actions = append(result.Actions, platform.BusinessUnitChangeNameAction{
			Name: plan.Name.ValueString(),
		})
	}

	if !c.ContactEmail.Equal(plan.ContactEmail) {
		result.Actions = append(result.Actions, platform.BusinessUnitSetContactEmailAction{
			ContactEmail: plan.ContactEmail.ValueStringPointer(),
		})
	}

	if !c.Status.Equal(plan.Status) {
		result.Actions = append(result.Actions, platform.BusinessUnitChangeStatusAction{
			Status: plan.Status.ValueString(),
		})
	}

	if !reflect.DeepEqual(c.Stores, plan.Stores) {
		result.Actions = append(result.Actions, platform.BusinessUnitSetStoresAction{
			Stores: pie.Map(plan.Stores, func(s sharedtypes.StoreKeyReference) platform.StoreResourceIdentifier {
				return platform.StoreResourceIdentifier{
					Key: s.Key.ValueStringPointer(),
				}
			}),
		})
	}

	if !reflect.DeepEqual(c.Addresses, plan.Addresses) {
		addressAddActions := sharedtypes.AddressesAddActions(c.Addresses, plan.Addresses)
		for _, action := range addressAddActions {
			result.Actions = append(result.Actions, action)
		}

		addressChangeActions := sharedtypes.AddressesChangeActions(c.Addresses, plan.Addresses)
		for _, action := range addressChangeActions {
			result.Actions = append(result.Actions, action)
		}
	}

	if !c.DefaultShippingAddressKey.Equal(plan.DefaultShippingAddressKey) {
		if !pie.Contains(plan.ShippingAddressKeys, plan.DefaultShippingAddressKey) {
			return result, fmt.Errorf("default shipping address key %s is not in shipping address keys", plan.DefaultShippingAddressKey.ValueString())
		}

		result.Actions = append(result.Actions, platform.BusinessUnitSetDefaultShippingAddressAction{
			AddressKey: plan.DefaultShippingAddressKey.ValueStringPointer(),
		})
	}

	if !c.DefaultBillingAddressKey.Equal(plan.DefaultBillingAddressKey) {
		if !pie.Contains(plan.BillingAddressKeys, plan.DefaultBillingAddressKey) {
			return result, fmt.Errorf("default shipping address key %s is not in shipping address keys", plan.DefaultBillingAddressKey.ValueString())
		}

		result.Actions = append(result.Actions, platform.BusinessUnitSetDefaultBillingAddressAction{
			AddressKey: plan.DefaultBillingAddressKey.ValueStringPointer(),
		})
	}

	if !reflect.DeepEqual(c.ShippingAddressKeys, plan.ShippingAddressKeys) {
		// find shipping addresses to be added
		for _, i := range plan.ShippingAddressKeys {
			if !pie.Contains(c.ShippingAddressKeys, i) {
				result.Actions = append(result.Actions, platform.BusinessUnitAddShippingAddressIdAction{
					AddressKey: i.ValueStringPointer(),
				})
			}
		}

		// find shipping addresses to be removed
		for _, i := range c.ShippingAddressKeys {
			if !pie.Contains(plan.ShippingAddressKeys, i) {
				result.Actions = append(result.Actions, platform.BusinessUnitRemoveShippingAddressIdAction{
					AddressKey: i.ValueStringPointer(),
				})
			}
		}
	}

	if !reflect.DeepEqual(c.BillingAddressKeys, plan.BillingAddressKeys) {
		// find billing addresses to be added
		for _, i := range plan.BillingAddressKeys {
			if !pie.Contains(c.BillingAddressKeys, i) {
				result.Actions = append(result.Actions, platform.BusinessUnitAddBillingAddressIdAction{
					AddressKey: i.ValueStringPointer(),
				})
			}
		}

		// find billing addresses to be removed
		for _, i := range c.BillingAddressKeys {
			if !pie.Contains(plan.BillingAddressKeys, i) {
				result.Actions = append(result.Actions, platform.BusinessUnitRemoveBillingAddressIdAction{
					AddressKey: i.ValueStringPointer(),
				})
			}
		}
	}

	// We need to delete addresses only after we have removed keys
	if !reflect.DeepEqual(c.Addresses, plan.Addresses) {
		addressDeleteActions := sharedtypes.AddressesDeleteActions(c.Addresses, plan.Addresses)
		for _, action := range addressDeleteActions {
			result.Actions = append(result.Actions, action)
		}
	}

	return result, nil
}

// NewCompanyFromNative creates a new Company from a platform.Company.
func NewCompanyFromNative(bu *platform.BusinessUnit) (Company, error) {
	data, ok := (*bu).(map[string]interface{})
	if !ok {
		return Company{}, fmt.Errorf("failed to convert business unit to map")
	}
	var c platform.Company
	err := utils.DecodeStruct(data, &c)
	if err != nil {
		return Company{}, err
	}

	var defaultShippingAddressKey *string
	if c.DefaultShippingAddressId != nil {
		i := slices.IndexFunc(c.Addresses, func(a platform.Address) bool {
			return *a.ID == *c.DefaultShippingAddressId
		})
		defaultShippingAddressKey = c.Addresses[i].Key
	}
	var defaultBillingAddressKey *string
	if c.DefaultBillingAddressId != nil {
		i := slices.IndexFunc(c.Addresses, func(a platform.Address) bool {
			return *a.ID == *c.DefaultBillingAddressId
		})
		defaultBillingAddressKey = c.Addresses[i].Key
	}

	var shippingAddressKeys []types.String
	for _, id := range c.ShippingAddressIds {
		i := slices.IndexFunc(c.Addresses, func(a platform.Address) bool {
			return *a.ID == id
		})
		shippingAddressKeys = append(shippingAddressKeys, types.StringPointerValue(c.Addresses[i].Key))
	}

	var billingAddressKeys []types.String
	for _, id := range c.BillingAddressIds {
		i := slices.IndexFunc(c.Addresses, func(a platform.Address) bool {
			return *a.ID == id
		})
		billingAddressKeys = append(billingAddressKeys, types.StringPointerValue(c.Addresses[i].Key))
	}

	var stores []sharedtypes.StoreKeyReference
	for _, s := range c.Stores {
		stores = append(stores, sharedtypes.NewStoreKeyReferenceFromNative(&s))
	}

	var addresses []sharedtypes.Address
	for _, a := range c.Addresses {
		addresses = append(addresses, sharedtypes.NewAddressFromNative(&a))
	}

	company := Company{
		ID:                        types.StringValue(c.ID),
		Version:                   types.Int64Value(int64(c.Version)),
		Key:                       types.StringValue(c.Key),
		Name:                      types.StringValue(c.Name),
		Status:                    types.StringValue(string(c.Status)),
		ContactEmail:              types.StringPointerValue(c.ContactEmail),
		DefaultShippingAddressKey: types.StringPointerValue(defaultShippingAddressKey),
		DefaultBillingAddressKey:  types.StringPointerValue(defaultBillingAddressKey),
		Stores:                    stores,
		Addresses:                 addresses,
		ShippingAddressKeys:       shippingAddressKeys,
		BillingAddressKeys:        billingAddressKeys,
	}

	sort.Slice(company.Addresses, func(i, j int) bool {
		return company.Addresses[i].Key.ValueString() < company.Addresses[j].Key.ValueString()
	})

	sort.Slice(company.ShippingAddressKeys, func(i, j int) bool {
		return company.ShippingAddressKeys[i].ValueString() < company.ShippingAddressKeys[j].ValueString()
	})

	sort.Slice(company.BillingAddressKeys, func(i, j int) bool {
		return company.BillingAddressKeys[i].ValueString() < company.BillingAddressKeys[j].ValueString()
	})

	return company, nil
}
