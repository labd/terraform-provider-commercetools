package business_unit_division

import (
	"fmt"
	"github.com/labd/terraform-provider-commercetools/internal/sharedtypes"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
	"reflect"
	"slices"
	"sort"

	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
)

type Division struct {
	ID                        types.String                    `tfsdk:"id"`
	Version                   types.Int64                     `tfsdk:"version"`
	Key                       types.String                    `tfsdk:"key"`
	Status                    types.String                    `tfsdk:"status"`
	StoreMode                 types.String                    `tfsdk:"store_mode"`
	ApprovalRuleMode          types.String                    `tfsdk:"approval_rule_mode"`
	Name                      types.String                    `tfsdk:"name"`
	ContactEmail              types.String                    `tfsdk:"contact_email"`
	AssociateMode             types.String                    `tfsdk:"associate_mode"`
	ShippingAddressKeys       []types.String                  `tfsdk:"shipping_address_keys"`
	DefaultShippingAddressKey types.String                    `tfsdk:"default_shipping_address_key"`
	BillingAddressKeys        []types.String                  `tfsdk:"billing_address_keys"`
	DefaultBillingAddressKey  types.String                    `tfsdk:"default_billing_address_key"`
	ParentUnit                BusinessUnitResourceIdentifier  `tfsdk:"parent_unit"`
	Stores                    []sharedtypes.StoreKeyReference `tfsdk:"store"`
	Addresses                 []sharedtypes.Address           `tfsdk:"address"`
	Custom                    *sharedtypes.Custom             `tfsdk:"custom"`
}

// BusinessUnitResourceIdentifier is a resource identifier for a business unit.
type BusinessUnitResourceIdentifier struct {
	ID  types.String `tfsdk:"id"`
	Key types.String `tfsdk:"key"`
}

func (d *Division) draft(t *platform.Type) (platform.DivisionDraft, error) {
	mode := platform.BusinessUnitStoreMode(d.StoreMode.ValueString())
	associateMode := platform.BusinessUnitAssociateMode(d.AssociateMode.ValueString())
	status := platform.BusinessUnitStatus(d.Status.ValueString())
	approvalRuleMode := platform.BusinessUnitApprovalRuleMode(d.ApprovalRuleMode.ValueString())

	var addresses = make([]platform.BaseAddress, 0, len(d.Addresses))
	for _, a := range d.Addresses {
		addresses = append(addresses, a.Draft())
	}

	var stores = make([]platform.StoreResourceIdentifier, 0, len(d.Stores))
	for _, s := range d.Stores {
		stores = append(stores, platform.StoreResourceIdentifier{
			Key: s.Key.ValueStringPointer(),
		})
	}

	var shippingAddressIndexes []int
	for _, key := range d.ShippingAddressKeys {
		i := slices.IndexFunc(d.Addresses, func(a sharedtypes.Address) bool {
			return a.Key.ValueString() == key.ValueString()
		})

		if i == -1 {
			return platform.DivisionDraft{}, fmt.Errorf("shipping address key %s is not in addresses", key.ValueString())
		}

		shippingAddressIndexes = append(shippingAddressIndexes, i)
	}

	var billingAddressIndexes []int
	for _, key := range d.BillingAddressKeys {
		i := slices.IndexFunc(d.Addresses, func(a sharedtypes.Address) bool {
			return a.Key.ValueString() == key.ValueString()
		})

		if i == -1 {
			return platform.DivisionDraft{}, fmt.Errorf("billing address key %s is not in addresses", key.ValueString())
		}

		billingAddressIndexes = append(billingAddressIndexes, i)
	}

	var defaultBillingAddressIndex *int
	if !d.DefaultBillingAddressKey.IsNull() {
		i := slices.IndexFunc(d.Addresses, func(a sharedtypes.Address) bool {
			return a.Key.ValueString() == d.DefaultBillingAddressKey.ValueString()
		})

		if i == -1 {
			return platform.DivisionDraft{}, fmt.Errorf("default billing address key %s is not in addresses", d.DefaultBillingAddressKey.ValueString())
		}

		defaultBillingAddressIndex = &i
	}

	var defaultShippingAddressIndex *int
	if !d.DefaultShippingAddressKey.IsNull() {
		i := slices.IndexFunc(d.Addresses, func(a sharedtypes.Address) bool {
			return a.Key.ValueString() == d.DefaultShippingAddressKey.ValueString()
		})

		if i == -1 {
			return platform.DivisionDraft{}, fmt.Errorf("default shipping address key %s is not in addresses", d.DefaultShippingAddressKey.ValueString())
		}

		defaultShippingAddressIndex = &i
	}

	custom, err := d.Custom.Draft(t)
	if err != nil {
		return platform.DivisionDraft{}, err
	}

	return platform.DivisionDraft{
		Key:              d.Key.ValueString(),
		Status:           &status,
		StoreMode:        &mode,
		AssociateMode:    &associateMode,
		ApprovalRuleMode: &approvalRuleMode,
		ParentUnit: platform.BusinessUnitResourceIdentifier{
			ID:  d.ParentUnit.ID.ValueStringPointer(),
			Key: d.ParentUnit.Key.ValueStringPointer(),
		},
		Stores:                 stores,
		Name:                   d.Name.ValueString(),
		ContactEmail:           d.ContactEmail.ValueStringPointer(),
		Addresses:              addresses,
		ShippingAddresses:      shippingAddressIndexes,
		BillingAddresses:       billingAddressIndexes,
		DefaultShippingAddress: defaultShippingAddressIndex,
		DefaultBillingAddress:  defaultBillingAddressIndex,
		Custom:                 custom,
	}, nil
}

func (d *Division) updateActions(t *platform.Type, plan Division) (platform.BusinessUnitUpdate, error) {
	result := platform.BusinessUnitUpdate{
		Version: int(d.Version.ValueInt64()),
		Actions: []platform.BusinessUnitUpdateAction{},
	}

	if !d.Key.Equal(plan.Key) {
		return result, fmt.Errorf("key is immutable. Delete this resource instead if a change is intended")
	}

	if !d.Name.Equal(plan.Name) {
		result.Actions = append(result.Actions, platform.BusinessUnitChangeNameAction{
			Name: plan.Name.ValueString(),
		})
	}

	if !d.ContactEmail.Equal(plan.ContactEmail) {
		result.Actions = append(result.Actions, platform.BusinessUnitSetContactEmailAction{
			ContactEmail: plan.ContactEmail.ValueStringPointer(),
		})
	}

	if !d.Status.Equal(plan.Status) {
		result.Actions = append(result.Actions, platform.BusinessUnitChangeStatusAction{
			Status: plan.Status.ValueString(),
		})
	}

	if !d.AssociateMode.Equal(plan.AssociateMode) {
		result.Actions = append(result.Actions, platform.BusinessUnitChangeAssociateModeAction{
			AssociateMode: platform.BusinessUnitAssociateMode(plan.AssociateMode.ValueString()),
		})
	}

	if !d.ApprovalRuleMode.Equal(plan.ApprovalRuleMode) {
		result.Actions = append(result.Actions, platform.BusinessUnitChangeApprovalRuleModeAction{
			ApprovalRuleMode: platform.BusinessUnitApprovalRuleMode(plan.ApprovalRuleMode.ValueString()),
		})
	}

	if !d.StoreMode.Equal(plan.StoreMode) {
		result.Actions = append(result.Actions, platform.BusinessUnitSetStoreModeAction{
			StoreMode: platform.BusinessUnitStoreMode(plan.StoreMode.ValueString()),
			Stores:    []platform.StoreResourceIdentifier{},
		})
	}

	if !reflect.DeepEqual(d.Stores, plan.Stores) {
		result.Actions = append(result.Actions, platform.BusinessUnitSetStoresAction{
			Stores: pie.Map(plan.Stores, func(s sharedtypes.StoreKeyReference) platform.StoreResourceIdentifier {
				return platform.StoreResourceIdentifier{
					Key: s.Key.ValueStringPointer(),
				}
			}),
		})
	}

	if !reflect.DeepEqual(d.Addresses, plan.Addresses) {
		addressAddActions := sharedtypes.AddressesAddActions(d.Addresses, plan.Addresses)
		for _, action := range addressAddActions {
			result.Actions = append(result.Actions, action)
		}

		addressChangeActions := sharedtypes.AddressesChangeActions(d.Addresses, plan.Addresses)
		for _, action := range addressChangeActions {
			result.Actions = append(result.Actions, action)
		}
	}

	if !d.DefaultShippingAddressKey.Equal(plan.DefaultShippingAddressKey) {
		if plan.DefaultShippingAddressKey.IsNull() {
			result.Actions = append(result.Actions, platform.BusinessUnitSetDefaultShippingAddressAction{})
		} else {
			if !pie.Contains(plan.ShippingAddressKeys, plan.DefaultShippingAddressKey) {
				return result, fmt.Errorf("default shipping address key %s is not in shipping address keys", plan.DefaultShippingAddressKey.ValueString())
			}

			result.Actions = append(result.Actions, platform.BusinessUnitSetDefaultShippingAddressAction{
				AddressKey: plan.DefaultShippingAddressKey.ValueStringPointer(),
			})
		}
	}

	if !d.DefaultBillingAddressKey.Equal(plan.DefaultBillingAddressKey) {
		if plan.DefaultBillingAddressKey.IsNull() {
			result.Actions = append(result.Actions, platform.BusinessUnitSetDefaultBillingAddressAction{})
		} else {
			if !pie.Contains(plan.BillingAddressKeys, plan.DefaultBillingAddressKey) {
				return result, fmt.Errorf("default billing address key %s is not in billing address keys", plan.DefaultBillingAddressKey.ValueString())
			}

			result.Actions = append(result.Actions, platform.BusinessUnitSetDefaultBillingAddressAction{
				AddressKey: plan.DefaultBillingAddressKey.ValueStringPointer(),
			})
		}
	}

	if !reflect.DeepEqual(d.ShippingAddressKeys, plan.ShippingAddressKeys) {
		// find shipping addresses to be added
		for _, i := range plan.ShippingAddressKeys {
			if !pie.Contains(d.ShippingAddressKeys, i) {
				result.Actions = append(result.Actions, platform.BusinessUnitAddShippingAddressIdAction{
					AddressKey: i.ValueStringPointer(),
				})
			}
		}

		// find shipping addresses to be removed
		for _, i := range d.ShippingAddressKeys {
			if !pie.Contains(plan.ShippingAddressKeys, i) {
				result.Actions = append(result.Actions, platform.BusinessUnitRemoveShippingAddressIdAction{
					AddressKey: i.ValueStringPointer(),
				})
			}
		}
	}

	if !reflect.DeepEqual(d.BillingAddressKeys, plan.BillingAddressKeys) {
		// find billing addresses to be added
		for _, i := range plan.BillingAddressKeys {
			if !pie.Contains(d.BillingAddressKeys, i) {
				result.Actions = append(result.Actions, platform.BusinessUnitAddBillingAddressIdAction{
					AddressKey: i.ValueStringPointer(),
				})
			}
		}

		// find billing addresses to be removed
		for _, i := range d.BillingAddressKeys {
			if !pie.Contains(plan.BillingAddressKeys, i) {
				result.Actions = append(result.Actions, platform.BusinessUnitRemoveBillingAddressIdAction{
					AddressKey: i.ValueStringPointer(),
				})
			}
		}
	}

	// We need to delete addresses only after we have removed keys
	if !reflect.DeepEqual(d.Addresses, plan.Addresses) {
		addressDeleteActions := sharedtypes.AddressesDeleteActions(d.Addresses, plan.Addresses)
		for _, action := range addressDeleteActions {
			result.Actions = append(result.Actions, action)
		}
	}

	// setCustomFields
	if !reflect.DeepEqual(d.Custom, plan.Custom) {
		actions, err := sharedtypes.CustomFieldUpdateActions[
			platform.BusinessUnitSetCustomTypeAction,
			platform.BusinessUnitSetCustomFieldAction,
		](t, d.Custom, plan.Custom)
		if err != nil {
			return platform.BusinessUnitUpdate{}, err
		}

		for i := range actions {
			result.Actions = append(result.Actions, actions[i].(platform.BusinessUnitUpdateAction))
		}
	}

	return result, nil
}

// NewDivisionFromNative creates a new Division from a platform.Division.
func NewDivisionFromNative(bu *platform.BusinessUnit) (Division, error) {
	data := (*bu).(map[string]interface{})
	var d platform.Division
	err := utils.DecodeStruct(data, &d)
	if err != nil {
		return Division{}, err
	}

	parent := BusinessUnitResourceIdentifier{
		Key: types.StringValue(d.ParentUnit.Key),
	}

	var defaultShippingAddressKey *string
	if d.DefaultShippingAddressId != nil {
		i := slices.IndexFunc(d.Addresses, func(a platform.Address) bool {
			return *a.ID == *d.DefaultShippingAddressId
		})
		defaultShippingAddressKey = d.Addresses[i].Key
	}
	var defaultBillingAddressKey *string
	if d.DefaultBillingAddressId != nil {
		i := slices.IndexFunc(d.Addresses, func(a platform.Address) bool {
			return *a.ID == *d.DefaultBillingAddressId
		})
		defaultBillingAddressKey = d.Addresses[i].Key
	}

	var shippingAddressKeys []types.String
	for _, id := range d.ShippingAddressIds {
		i := slices.IndexFunc(d.Addresses, func(a platform.Address) bool {
			return *a.ID == id
		})
		shippingAddressKeys = append(shippingAddressKeys, types.StringPointerValue(d.Addresses[i].Key))
	}

	var billingAddressKeys []types.String
	for _, id := range d.BillingAddressIds {
		i := slices.IndexFunc(d.Addresses, func(a platform.Address) bool {
			return *a.ID == id
		})
		billingAddressKeys = append(billingAddressKeys, types.StringPointerValue(d.Addresses[i].Key))
	}

	var stores []sharedtypes.StoreKeyReference
	for _, s := range d.Stores {
		stores = append(stores, sharedtypes.NewStoreKeyReferenceFromNative(&s))
	}

	var addresses []sharedtypes.Address
	for _, a := range d.Addresses {
		addresses = append(addresses, sharedtypes.NewAddressFromNative(&a))
	}

	custom, err := sharedtypes.NewCustomFromNative(d.Custom)
	if err != nil {
		return Division{}, err
	}

	division := Division{
		ID:                        types.StringValue(d.ID),
		Version:                   types.Int64Value(int64(d.Version)),
		Key:                       types.StringValue(d.Key),
		Status:                    types.StringValue(string(d.Status)),
		ParentUnit:                parent,
		StoreMode:                 types.StringValue(string(d.StoreMode)),
		ApprovalRuleMode:          types.StringValue(string(d.ApprovalRuleMode)),
		Name:                      types.StringValue(d.Name),
		ContactEmail:              types.StringPointerValue(d.ContactEmail),
		DefaultShippingAddressKey: types.StringPointerValue(defaultShippingAddressKey),
		DefaultBillingAddressKey:  types.StringPointerValue(defaultBillingAddressKey),
		AssociateMode:             types.StringValue(string(d.AssociateMode)),
		Stores:                    stores,
		Addresses:                 addresses,
		ShippingAddressKeys:       shippingAddressKeys,
		BillingAddressKeys:        billingAddressKeys,
		Custom:                    custom,
	}

	sort.Slice(division.Addresses, func(i, j int) bool {
		return division.Addresses[i].Key.ValueString() < division.Addresses[j].Key.ValueString()
	})

	sort.Slice(division.ShippingAddressKeys, func(i, j int) bool {
		return division.ShippingAddressKeys[i].ValueString() < division.ShippingAddressKeys[j].ValueString()
	})

	sort.Slice(division.BillingAddressKeys, func(i, j int) bool {
		return division.BillingAddressKeys[i].ValueString() < division.BillingAddressKeys[j].ValueString()
	})

	return division, nil
}
