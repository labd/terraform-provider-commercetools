package business_unit

import (
	"reflect"

	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
)

const (
	// Store modes for business units.
	StoreModeExplicit   = "Explicit"
	StoreModeFromParent = "FromParent"

	// Statuses for business units.
	BusinessUnitActive   = "Active"
	BusinessUnitInactive = "Inactive"

	// Unit types for business units.
	CompanyType  = "Company"
	DivisionType = "Division"

	// Associate modes for business units.
	ExplicitAssociateMode              = "Explicit"
	ExplicitAndFromParentAssociateMode = "ExplicitAndFromParent"

	// Associate role inheritance for business units.
	AssociateRoleInheritanceEnabled  = "Enabled"
	AssociateRoleInheritanceDisabled = "Disabled"

	// Store type id.
	StoreTypeID         = "store"
	AssociateRoleTypeID = "associate-role"
	BusinessUnitTypeID  = "business-unit"
	CustomerTypeID      = "customer"
)

// BusinessUnit is a type to model the fields that all types of Business Units have in common.
type BusinessUnit struct {
	ID                       types.String             `tfsdk:"id"`
	Version                  types.Int64              `tfsdk:"version"`
	Key                      types.String             `tfsdk:"key"`
	Status                   types.String             `tfsdk:"status"`
	Stores                   []StoreKeyReference      `tfsdk:"stores"`
	StoreMode                types.String             `tfsdk:"store_mode"`
	UnitType                 types.String             `tfsdk:"unit_type"`
	Name                     types.String             `tfsdk:"name"`
	ContactEmail             types.String             `tfsdk:"contact_email"`
	Addresses                []Address                `tfsdk:"addresses"`
	ShippingAddressIDs       []types.String           `tfsdk:"shipping_address_ids"`
	DefaultShippingAddressID types.String             `tfsdk:"default_shipping_address_id"`
	BillingAddressIDs        []types.String           `tfsdk:"billing_address_ids"`
	DefaultBillingAddressID  types.String             `tfsdk:"default_billing_address_id"`
	AssociateMode            types.String             `tfsdk:"associate_mode"`
	Associates               []Associate              `tfsdk:"associates"`
	InheritedAssociates      []InheritedAssociate     `tfsdk:"inherited_associates"`
	ParentUnit               BusinessUnitKeyReference `tfsdk:"parent_unit"`
	TopLevelUnit             BusinessUnitKeyReference `tfsdk:"top_level_unit"`
}

func (bu BusinessUnit) draft() platform.BusinessUnitDraft {
	switch bu.UnitType.ValueString() {
	case CompanyType:
		return bu.draftCompany()
	case DivisionType:
		return bu.draftDivision()
	default:
		return new(platform.BusinessUnitDraft)
	}
}

func (bu BusinessUnit) draftCompany() platform.CompanyDraft {
	status := platform.BusinessUnitStatus(bu.Status.ValueString())
	mode := platform.BusinessUnitStoreMode(bu.StoreMode.ValueString())
	assoc := platform.BusinessUnitAssociateMode(bu.AssociateMode.ValueString())
	dsa := pie.Int(bu.DefaultShippingAddressID.ValueString())
	dba := pie.Int(bu.DefaultBillingAddressID.ValueString())

	return platform.CompanyDraft{
		Key:    bu.Key.ValueString(),
		Status: &status,
		Stores: pie.Map(bu.Stores, func(s StoreKeyReference) platform.StoreResourceIdentifier {
			return platform.StoreResourceIdentifier{
				Key: s.Key.ValueStringPointer(),
				ID:  s.Key.ValueStringPointer(),
			}
		}),
		StoreMode:     &mode,
		Name:          bu.Name.ValueString(),
		ContactEmail:  bu.ContactEmail.ValueStringPointer(),
		AssociateMode: &assoc,
		Associates: pie.Map(bu.Associates, func(a Associate) platform.AssociateDraft {
			return a.draft()
		}),
		Addresses: pie.Map(bu.Addresses, func(a Address) platform.BaseAddress {
			return a.draft()
		}),
		ShippingAddresses: pie.Map(bu.ShippingAddressIDs, func(id types.String) int {
			return pie.Int(id.ValueString())
		}),
		DefaultShippingAddress: &dsa,
		BillingAddresses: pie.Map(bu.BillingAddressIDs, func(id types.String) int {
			return pie.Int(id.ValueString())
		}),
		DefaultBillingAddress: &dba,
	}
}

func (bu BusinessUnit) draftDivision() platform.DivisionDraft {
	status := platform.BusinessUnitStatus(bu.Status.ValueString())
	mode := platform.BusinessUnitStoreMode(bu.StoreMode.ValueString())
	assoc := platform.BusinessUnitAssociateMode(bu.AssociateMode.ValueString())
	dsa := pie.Int(bu.DefaultShippingAddressID.ValueString())
	dba := pie.Int(bu.DefaultBillingAddressID.ValueString())

	return platform.DivisionDraft{
		Key:    bu.Key.ValueString(),
		Status: &status,
		Stores: pie.Map(bu.Stores, func(s StoreKeyReference) platform.StoreResourceIdentifier {
			return platform.StoreResourceIdentifier{
				Key: s.Key.ValueStringPointer(),
				ID:  s.Key.ValueStringPointer(),
			}
		}),
		StoreMode:     &mode,
		Name:          bu.Name.ValueString(),
		ContactEmail:  bu.ContactEmail.ValueStringPointer(),
		AssociateMode: &assoc,
		Associates: pie.Map(bu.Associates, func(a Associate) platform.AssociateDraft {
			return a.draft()
		}),
		Addresses: pie.Map(bu.Addresses, func(a Address) platform.BaseAddress {
			return a.draft()
		}),
		ShippingAddresses: pie.Map(bu.ShippingAddressIDs, func(id types.String) int {
			return pie.Int(id.ValueString())
		}),
		DefaultShippingAddress: &dsa,
		BillingAddresses: pie.Map(bu.BillingAddressIDs, func(id types.String) int {
			return pie.Int(id.ValueString())
		}),
		DefaultBillingAddress: &dba,
		ParentUnit: platform.BusinessUnitResourceIdentifier{
			Key: bu.ParentUnit.Key.ValueStringPointer(),
			ID:  bu.ParentUnit.Key.ValueStringPointer(),
		},
	}
}

func (bu BusinessUnit) updateActions(plan BusinessUnit) platform.BusinessUnitUpdate {
	result := platform.BusinessUnitUpdate{
		Version: int(bu.Version.ValueInt64()),
		Actions: []platform.BusinessUnitUpdateAction{},
	}

	// update business unit status.
	if bu.Status != plan.Status {
		var newStatus platform.BusinessUnitStatus
		if !plan.Status.IsNull() && !plan.Status.IsUnknown() {
			newStatus = platform.BusinessUnitStatus(plan.Status.ValueString())
		}

		result.Actions = append(
			result.Actions,
			platform.BusinessUnitChangeStatusAction{Status: string(newStatus)},
		)
	}

	// update stores associated to a business unit.
	if !reflect.DeepEqual(bu.Stores, plan.Stores) {
		if bu.StoreMode == types.StringValue(StoreModeExplicit) {
			result.Actions = append(
				result.Actions,
				platform.BusinessUnitSetStoresAction{
					Stores: pie.Map(plan.Stores, func(s StoreKeyReference) platform.StoreResourceIdentifier {
						return platform.StoreResourceIdentifier{
							Key: s.Key.ValueStringPointer(),
						}
					}),
				},
			)
		}
	}

	// update business unit store mode.
	if bu.StoreMode != plan.StoreMode {
		var newMode platform.BusinessUnitStoreMode
		if !plan.StoreMode.IsNull() && !plan.StoreMode.IsUnknown() {
			newMode = platform.BusinessUnitStoreMode(plan.StoreMode.ValueString())
		}

		actions := platform.BusinessUnitSetStoreModeAction{StoreMode: newMode}

		// if the new store mode is explicit, we need to add the stores.
		if newMode == platform.BusinessUnitStoreModeExplicit {
			actions.Stores = pie.Map(plan.Stores, func(s StoreKeyReference) platform.StoreResourceIdentifier {
				return platform.StoreResourceIdentifier{
					Key: s.Key.ValueStringPointer(),
				}
			})
		}

		result.Actions = append(result.Actions, actions)
	}

	// update business unit name.
	if bu.Name != plan.Name {
		var newName string
		if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
			newName = plan.Name.ValueString()
		}

		result.Actions = append(
			result.Actions,
			platform.BusinessUnitChangeNameAction{Name: newName},
		)
	}

	// update business unit contact email.
	if bu.ContactEmail != plan.ContactEmail {
		var newEmail *string
		if !plan.ContactEmail.IsNull() && !plan.ContactEmail.IsUnknown() {
			newEmail = plan.ContactEmail.ValueStringPointer()
		}

		result.Actions = append(
			result.Actions,
			platform.BusinessUnitSetContactEmailAction{ContactEmail: newEmail},
		)
	}

	// update business unit associate mode.
	if bu.AssociateMode != plan.AssociateMode {
		var newMode platform.BusinessUnitAssociateMode
		if !plan.AssociateMode.IsNull() && !plan.AssociateMode.IsUnknown() {
			newMode = platform.BusinessUnitAssociateMode(plan.AssociateMode.ValueString())
		}

		result.Actions = append(
			result.Actions,
			platform.BusinessUnitChangeAssociateModeAction{AssociateMode: newMode},
		)
	}

	// update business unit associates.
	if !reflect.DeepEqual(bu.Associates, plan.Associates) {
		result.Actions = append(
			result.Actions,
			platform.BusinessUnitSetAssociatesAction{
				Associates: pie.Map(plan.Associates, func(a Associate) platform.AssociateDraft {
					return a.draft()
				}),
			},
		)
	}

	// update business unit addresses.
	if !reflect.DeepEqual(bu.Addresses, plan.Addresses) {
		// find addresses to remove.
		for _, a := range bu.Addresses {
			if !pie.Contains(plan.Addresses, a) {
				result.Actions = append(
					result.Actions,
					platform.BusinessUnitRemoveAddressAction{
						AddressKey: a.Key.ValueStringPointer(),
					},
				)
			}
		}

		// find addresses to add.
		for _, a := range plan.Addresses {
			if !pie.Contains(bu.Addresses, a) {
				result.Actions = append(
					result.Actions,
					platform.BusinessUnitAddAddressAction{
						Address: a.draft(),
					},
				)
			}
		}
	}

	// update business unit shipping addresses.
	if !reflect.DeepEqual(bu.ShippingAddressIDs, plan.ShippingAddressIDs) {
		// find shipping addresses to remove.
		for _, a := range bu.ShippingAddressIDs {
			if !pie.Contains(plan.ShippingAddressIDs, a) {
				result.Actions = append(
					result.Actions,
					platform.BusinessUnitRemoveShippingAddressIdAction{AddressKey: a.ValueStringPointer()},
				)
			}
		}

		// find shipping addresses to add.
		for _, a := range plan.ShippingAddressIDs {
			if !pie.Contains(bu.ShippingAddressIDs, a) {
				result.Actions = append(
					result.Actions,
					platform.BusinessUnitAddShippingAddressIdAction{AddressKey: a.ValueStringPointer()},
				)
			}
		}
	}

	// update business unit default shipping address.
	if bu.DefaultShippingAddressID != plan.DefaultShippingAddressID {
		var newDefault string
		if !plan.DefaultShippingAddressID.IsNull() && !plan.DefaultShippingAddressID.IsUnknown() {
			newDefault = plan.DefaultShippingAddressID.ValueString()
		}

		result.Actions = append(
			result.Actions,
			platform.BusinessUnitSetDefaultShippingAddressAction{AddressKey: &newDefault},
		)
	}

	// update business unit billing addresses.
	if !reflect.DeepEqual(bu.BillingAddressIDs, plan.BillingAddressIDs) {
		// find billing addresses to remove.
		for _, a := range bu.BillingAddressIDs {
			if !pie.Contains(plan.BillingAddressIDs, a) {
				result.Actions = append(
					result.Actions,
					platform.BusinessUnitRemoveBillingAddressIdAction{AddressKey: a.ValueStringPointer()},
				)
			}
		}

		// find billing addresses to add.
		for _, a := range plan.BillingAddressIDs {
			if !pie.Contains(bu.BillingAddressIDs, a) {
				result.Actions = append(
					result.Actions,
					platform.BusinessUnitAddBillingAddressIdAction{AddressKey: a.ValueStringPointer()},
				)
			}
		}
	}

	// update business unit default billing address.
	if bu.DefaultBillingAddressID != plan.DefaultBillingAddressID {
		var newDefault string
		if !plan.DefaultBillingAddressID.IsNull() && !plan.DefaultBillingAddressID.IsUnknown() {
			newDefault = plan.DefaultBillingAddressID.ValueString()
		}

		result.Actions = append(
			result.Actions,
			platform.BusinessUnitSetDefaultBillingAddressAction{AddressKey: &newDefault},
		)
	}

	return result
}

// NewBusinessUnitFromNative creates a new BusinessUnit from a platform.BusinessUnit.
func NewBusinessUnitFromNative(bu platform.BusinessUnit) BusinessUnit {
	if val, ok := bu.(platform.Company); ok {
		return newCompanyFromInterface(val)
	} else if val, ok := bu.(platform.Division); ok {
		return newDivisionFromInterface(val)
	}

	return BusinessUnit{}
}

func newCompanyFromInterface(bu platform.Company) BusinessUnit {
	tf := BusinessUnit{
		ID:                       types.StringValue(bu.ID),
		Version:                  types.Int64Value(int64(bu.Version)),
		Key:                      types.StringValue(bu.Key),
		Status:                   types.StringValue(string(bu.Status)),
		StoreMode:                types.StringValue(string(bu.StoreMode)),
		Stores:                   make([]StoreKeyReference, len(bu.Stores)),
		UnitType:                 types.StringValue(CompanyType),
		Name:                     types.StringValue(bu.Name),
		ContactEmail:             types.StringPointerValue(bu.ContactEmail),
		Addresses:                make([]Address, len(bu.Addresses)),
		ShippingAddressIDs:       make([]types.String, len(bu.ShippingAddressIds)),
		DefaultShippingAddressID: types.StringPointerValue(bu.DefaultShippingAddressId),
		BillingAddressIDs:        make([]types.String, len(bu.BillingAddressIds)),
		DefaultBillingAddressID:  types.StringPointerValue(bu.DefaultBillingAddressId),
		AssociateMode:            types.StringValue(ExplicitAssociateMode),
		Associates:               make([]Associate, len(bu.Associates)),
		InheritedAssociates:      make([]InheritedAssociate, len(bu.InheritedAssociates)),
		ParentUnit:               BusinessUnitKeyReference{},
	}

	for i, a := range bu.Addresses {
		tf.Addresses[i] = NewAddressFromNative(&a)
	}

	for i, a := range bu.ShippingAddressIds {
		tf.ShippingAddressIDs[i] = types.StringValue(a)
	}

	for i, a := range bu.BillingAddressIds {
		tf.BillingAddressIDs[i] = types.StringValue(a)
	}

	for i, a := range bu.Associates {
		tf.Associates[i] = NewAssociateFromNative(&a)
	}

	for i, a := range bu.InheritedAssociates {
		tf.InheritedAssociates[i] = NewInheritedAssociateFromNative(&a)
	}

	for i, a := range bu.Stores {
		tf.Stores[i] = NewStoreKeyReferenceFromNative(&a)
	}

	return tf
}

func newDivisionFromInterface(bu platform.Division) BusinessUnit {
	tf := BusinessUnit{
		ID:                       types.StringValue(bu.ID),
		Version:                  types.Int64Value(int64(bu.Version)),
		Key:                      types.StringValue(bu.Key),
		Status:                   types.StringValue(string(bu.Status)),
		StoreMode:                types.StringValue(string(bu.StoreMode)),
		Stores:                   make([]StoreKeyReference, len(bu.Stores)),
		UnitType:                 types.StringValue(DivisionType),
		Name:                     types.StringValue(bu.Name),
		ContactEmail:             types.StringPointerValue(bu.ContactEmail),
		Addresses:                make([]Address, len(bu.Addresses)),
		ShippingAddressIDs:       make([]types.String, len(bu.ShippingAddressIds)),
		DefaultShippingAddressID: types.StringPointerValue(bu.DefaultShippingAddressId),
		BillingAddressIDs:        make([]types.String, len(bu.BillingAddressIds)),
		DefaultBillingAddressID:  types.StringPointerValue(bu.DefaultBillingAddressId),
		AssociateMode:            types.StringValue(string(bu.AssociateMode)),
		Associates:               make([]Associate, len(bu.Associates)),
		InheritedAssociates:      make([]InheritedAssociate, len(bu.InheritedAssociates)),
		ParentUnit:               NewBusinessUnitKeyReferenceFromNative(&bu.ParentUnit),
	}

	for i, a := range bu.Addresses {
		tf.Addresses[i] = NewAddressFromNative(&a)
	}

	for i, a := range bu.ShippingAddressIds {
		tf.ShippingAddressIDs[i] = types.StringValue(a)
	}

	for i, a := range bu.BillingAddressIds {
		tf.BillingAddressIDs[i] = types.StringValue(a)
	}

	for i, a := range bu.Associates {
		tf.Associates[i] = NewAssociateFromNative(&a)
	}

	for i, a := range bu.InheritedAssociates {
		tf.InheritedAssociates[i] = NewInheritedAssociateFromNative(&a)
	}

	for i, a := range bu.Stores {
		tf.Stores[i] = NewStoreKeyReferenceFromNative(&a)
	}

	return tf
}

// StoreResourceIdentifier is a type to model the fields that all types of
// Store Resource Identifiers have in common.
type StoreResourceIdentifier struct {
	ID     types.String `tfsdk:"id"`
	Key    types.String `tfsdk:"key"`
	TypeID types.String `tfsdk:"type_id"`
}

func (s StoreResourceIdentifier) draft() platform.StoreResourceIdentifier {
	if !s.ID.IsNull() || !s.ID.IsUnknown() {
		return platform.StoreResourceIdentifier{
			ID: s.ID.ValueStringPointer(),
		}
	}

	if !s.Key.IsNull() || !s.Key.IsUnknown() {
		return platform.StoreResourceIdentifier{
			Key: s.Key.ValueStringPointer(),
		}
	}

	return platform.StoreResourceIdentifier{}
}

// StoreKeyReference is a type to model the fields that all types of
// Store Key References have in common.
type StoreKeyReference struct {
	Key    types.String `tfsdk:"key"`
	TypeID types.String `tfsdk:"type_id"`
}

func (s StoreKeyReference) draft() platform.StoreKeyReference {
	return platform.StoreKeyReference{
		Key: s.Key.ValueString(),
	}
}

// NewStoreKeyReferenceFromNative creates a new StoreKeyReference from a
// platform.StoreKeyReference.
func NewStoreKeyReferenceFromNative(kr *platform.StoreKeyReference) StoreKeyReference {
	return StoreKeyReference{
		Key:    types.StringValue(kr.Key),
		TypeID: types.StringValue(StoreTypeID),
	}
}

// AssociateRoleKeyReference is a type to model the fields that all types of
// Associate Role Key References have in common.
type AssociateRoleKeyReference struct {
	Key    types.String `tfsdk:"key"`
	TypeID types.String `tfsdk:"type_id"`
}

func (kr AssociateRoleKeyReference) draft() platform.AssociateRoleResourceIdentifier {
	if !kr.Key.IsNull() || !kr.Key.IsUnknown() {
		return platform.AssociateRoleResourceIdentifier{
			Key: kr.Key.ValueStringPointer(),
		}
	}

	return platform.AssociateRoleResourceIdentifier{}
}

// NewAssociateRoleKeyReferenceFromNative creates a new AssociateRoleKeyReference
// from a platform.AssociateRoleKeyReference.
func NewAssociateRoleKeyReferenceFromNative(kr *platform.AssociateRoleKeyReference) AssociateRoleKeyReference {
	return AssociateRoleKeyReference{
		Key:    types.StringValue(kr.Key),
		TypeID: types.StringValue(AssociateRoleTypeID),
	}
}

// CustomerReference is a type to model the fields that all types of
// Customer References have in common.
type CustomerReference struct {
	ID     types.String `tfsdk:"id"`
	TypeID types.String `tfsdk:"type_id"`
}

func (cr CustomerReference) draft() platform.CustomerResourceIdentifier {
	if !cr.ID.IsNull() || !cr.ID.IsUnknown() {
		return platform.CustomerResourceIdentifier{
			ID: cr.ID.ValueStringPointer(),
		}
	}

	return platform.CustomerResourceIdentifier{}
}

// NewCustomerReferenceFromNative creates a new CustomerReference from a
// platform.CustomerReference.
func NewCustomerReferenceFromNative(kr *platform.CustomerReference) CustomerReference {
	return CustomerReference{
		ID:     types.StringValue(kr.ID),
		TypeID: types.StringValue(CustomerTypeID),
	}
}

// BusinessUnitKeyReference is a type to model the fields that all types of
// Business Unit Key References have in common.
type BusinessUnitKeyReference struct {
	Key    types.String `tfsdk:"key"`
	TypeID types.String `tfsdk:"type_id"`
}

// NewBusinessUnitKeyReferenceFromNative creates a new BusinessUnitKeyReference
// from a platform.BusinessUnitKeyReference.
func NewBusinessUnitKeyReferenceFromNative(kr *platform.BusinessUnitKeyReference) BusinessUnitKeyReference {
	return BusinessUnitKeyReference{
		Key:    types.StringValue(kr.Key),
		TypeID: types.StringValue(BusinessUnitTypeID),
	}
}

// AssociateRoleAssignment is a type to model the fields that all types of
// Associate Role Assignments have in common.
type AssociateRoleAssignment struct {
	AssociateRole AssociateRoleKeyReference `tfsdk:"associate_role"`
	Inheritance   types.String              `tfsdk:"inheritance"`
}

func (ara AssociateRoleAssignment) draft() platform.AssociateRoleAssignmentDraft {
	if ara.Inheritance.IsNull() || ara.Inheritance.IsUnknown() {
		return platform.AssociateRoleAssignmentDraft{
			AssociateRole: ara.AssociateRole.draft(),
		}
	}

	inheritance := platform.AssociateRoleInheritanceMode(ara.Inheritance.ValueString())

	return platform.AssociateRoleAssignmentDraft{
		AssociateRole: ara.AssociateRole.draft(),
		Inheritance:   &inheritance,
	}
}

// NewAssociateRoleAssignment creates a new AssociateRoleAssignment from a
// platform.AssociateRoleAssignment.
func NewAssociateRoleAssignment(ara *platform.AssociateRoleAssignment) AssociateRoleAssignment {
	ar := AssociateRoleAssignment{}

	ar.AssociateRole = NewAssociateRoleKeyReferenceFromNative(&ara.AssociateRole)
	ar.Inheritance = types.StringValue(string(ara.Inheritance))

	return ar
}

// Associate is a type to model the fields that all types of Associates have in common.
type Associate struct {
	AssociateRoleAssignments []AssociateRoleAssignment `tfsdk:"associate_role_assignments"`
	Customer                 CustomerReference         `tfsdk:"customer"`
}

func (a Associate) draft() platform.AssociateDraft {
	return platform.AssociateDraft{
		AssociateRoleAssignments: pie.Map(a.AssociateRoleAssignments, func(ara AssociateRoleAssignment) platform.AssociateRoleAssignmentDraft {
			return ara.draft()
		}),
		Customer: a.Customer.draft(),
	}
}

// NewAssociateFromNative creates a new Associate from a platform.Associate.
func NewAssociateFromNative(a *platform.Associate) Associate {
	assoc := Associate{
		AssociateRoleAssignments: make([]AssociateRoleAssignment, len(a.AssociateRoleAssignments)),
		Customer:                 NewCustomerReferenceFromNative(&a.Customer),
	}

	for i, ara := range a.AssociateRoleAssignments {
		assoc.AssociateRoleAssignments[i] = NewAssociateRoleAssignment(&ara)
	}

	return assoc
}

// InheritedAssociate is a type to model the fields that all types of Inherited Associates have in common.
type InheritedAssociate struct {
	AssociateRoleAssignments []InheritedAssociateRoleAssignment `tfsdk:"associate_role_assignment"`
	Customer                 CustomerReference                  `tfsdk:"customer"`
}

// NewInheritedAssociateFromNative creates a new InheritedAssociate from a
// platform.InheritedAssociate.
func NewInheritedAssociateFromNative(ia *platform.InheritedAssociate) InheritedAssociate {
	localIA := InheritedAssociate{
		AssociateRoleAssignments: make([]InheritedAssociateRoleAssignment, len(ia.AssociateRoleAssignments)),
		Customer:                 NewCustomerReferenceFromNative(&ia.Customer),
	}

	for i, ara := range ia.AssociateRoleAssignments {
		localIA.AssociateRoleAssignments[i] = NewInheritedAssociateRoleAssignmentFromNative(&ara)
	}

	return localIA
}

// InheritedAssociateRoleAssignment is a type to model the fields that all types of
// Inherited Associate Role Assignments have in common.
type InheritedAssociateRoleAssignment struct {
	AssociateRole AssociateRoleKeyReference `tfsdk:"associate_role"`
	Source        BusinessUnitKeyReference  `tfsdk:"source"`
}

// NewInheritedAssociateRoleAssignmentFromNative creates a new
// InheritedAssociateRoleAssignment from a platform.InheritedAssociateRoleAssignment.
func NewInheritedAssociateRoleAssignmentFromNative(ara *platform.InheritedAssociateRoleAssignment) InheritedAssociateRoleAssignment {
	localARA := InheritedAssociateRoleAssignment{}

	localARA.AssociateRole = NewAssociateRoleKeyReferenceFromNative(&ara.AssociateRole)
	localARA.Source = NewBusinessUnitKeyReferenceFromNative(&ara.Source)

	return localARA
}

// Address is a type to model the fields that all types of Addresses have in common.
type Address struct {
	ID                    types.String `tfsdk:"id"`
	Key                   types.String `tfsdk:"key"`
	ExternalID            types.String `tfsdk:"external_id"`
	Country               types.String `tfsdk:"country"`
	Title                 types.String `tfsdk:"title"`
	Salutation            types.String `tfsdk:"salutation"`
	FirstName             types.String `tfsdk:"first_name"`
	LastName              types.String `tfsdk:"last_name"`
	StreetName            types.String `tfsdk:"street_name"`
	StreetNumber          types.String `tfsdk:"street_number"`
	AdditionalStreetInfo  types.String `tfsdk:"additional_street_info"`
	PostalCode            types.String `tfsdk:"postal_code"`
	City                  types.String `tfsdk:"city"`
	Region                types.String `tfsdk:"region"`
	State                 types.String `tfsdk:"state"`
	Company               types.String `tfsdk:"company"`
	Department            types.String `tfsdk:"department"`
	Building              types.String `tfsdk:"building"`
	Apartment             types.String `tfsdk:"apartment"`
	POBox                 types.String `tfsdk:"po_box"`
	Phone                 types.String `tfsdk:"phone"`
	Mobile                types.String `tfsdk:"mobile"`
	Email                 types.String `tfsdk:"email"`
	Fax                   types.String `tfsdk:"fax"`
	AdditionalAddressInfo types.String `tfsdk:"additional_address_info"`
}

func (a Address) draft() platform.BaseAddress {
	return platform.BaseAddress{
		Key:                   a.Key.ValueStringPointer(),
		ExternalId:            a.ExternalID.ValueStringPointer(),
		Country:               a.Country.ValueString(),
		Title:                 a.Title.ValueStringPointer(),
		Salutation:            a.Salutation.ValueStringPointer(),
		FirstName:             a.FirstName.ValueStringPointer(),
		LastName:              a.LastName.ValueStringPointer(),
		StreetName:            a.StreetName.ValueStringPointer(),
		StreetNumber:          a.StreetNumber.ValueStringPointer(),
		AdditionalStreetInfo:  a.AdditionalStreetInfo.ValueStringPointer(),
		PostalCode:            a.PostalCode.ValueStringPointer(),
		City:                  a.City.ValueStringPointer(),
		Region:                a.Region.ValueStringPointer(),
		State:                 a.State.ValueStringPointer(),
		Company:               a.Company.ValueStringPointer(),
		Department:            a.Department.ValueStringPointer(),
		Building:              a.Building.ValueStringPointer(),
		Apartment:             a.Apartment.ValueStringPointer(),
		POBox:                 a.POBox.ValueStringPointer(),
		Phone:                 a.Phone.ValueStringPointer(),
		Mobile:                a.Mobile.ValueStringPointer(),
		Email:                 a.Email.ValueStringPointer(),
		Fax:                   a.Fax.ValueStringPointer(),
		AdditionalAddressInfo: a.AdditionalAddressInfo.ValueStringPointer(),
	}
}

// NewAddressFromNative creates a new Address from a platform.Address.
func NewAddressFromNative(a *platform.Address) Address {
	return Address{
		ID:                    types.StringPointerValue(a.ID),
		Key:                   types.StringPointerValue(a.Key),
		ExternalID:            types.StringPointerValue(a.ExternalId),
		Country:               types.StringValue(a.Country),
		Title:                 types.StringPointerValue(a.Title),
		Salutation:            types.StringPointerValue(a.Salutation),
		FirstName:             types.StringPointerValue(a.FirstName),
		LastName:              types.StringPointerValue(a.LastName),
		StreetName:            types.StringPointerValue(a.StreetName),
		StreetNumber:          types.StringPointerValue(a.StreetNumber),
		AdditionalStreetInfo:  types.StringPointerValue(a.AdditionalStreetInfo),
		PostalCode:            types.StringPointerValue(a.PostalCode),
		City:                  types.StringPointerValue(a.City),
		Region:                types.StringPointerValue(a.Region),
		State:                 types.StringPointerValue(a.State),
		Company:               types.StringPointerValue(a.Company),
		Department:            types.StringPointerValue(a.Department),
		Building:              types.StringPointerValue(a.Building),
		Apartment:             types.StringPointerValue(a.Apartment),
		POBox:                 types.StringPointerValue(a.POBox),
		Phone:                 types.StringPointerValue(a.Phone),
		Mobile:                types.StringPointerValue(a.Mobile),
		Email:                 types.StringPointerValue(a.Email),
		Fax:                   types.StringPointerValue(a.Fax),
		AdditionalAddressInfo: types.StringPointerValue(a.AdditionalAddressInfo),
	}
}
