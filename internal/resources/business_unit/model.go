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

/*
Model types for the business unit resource.
*/
type Company struct {
	ID                       types.String        `tfsdk:"id"`
	Version                  types.Int64         `tfsdk:"version"`
	Key                      types.String        `tfsdk:"key"`
	Status                   types.String        `tfsdk:"status"`
	Stores                   []StoreKeyReference `tfsdk:"store"`
	Name                     types.String        `tfsdk:"name"`
	ContactEmail             types.String        `tfsdk:"contact_email"`
	Addresses                []Address           `tfsdk:"address"`
	ShippingAddressIDs       []types.String      `tfsdk:"shipping_address_ids"`
	DefaultShippingAddressID types.String        `tfsdk:"default_shipping_address_id"`
	BillingAddressIDs        []types.String      `tfsdk:"billing_address_ids"`
	DefaultBillingAddressID  types.String        `tfsdk:"default_billing_address_id"`
	AssociateMode            types.String        `tfsdk:"associate_mode"`
	Associates               []Associate         `tfsdk:"associate"`
}

func (c Company) draft() platform.CompanyDraft {
	status := platform.BusinessUnitStatus(c.Status.ValueString())
	associateMode := platform.BusinessUnitAssociateMode(c.AssociateMode.ValueString())
	dsa := pie.Int(c.DefaultShippingAddressID.ValueString())
	dba := pie.Int(c.DefaultBillingAddressID.ValueString())

	return platform.CompanyDraft{
		Key:    c.Key.ValueString(),
		Status: &status,
		Stores: pie.Map(c.Stores, func(s StoreKeyReference) platform.StoreResourceIdentifier {
			return platform.StoreResourceIdentifier{
				Key: s.Key.ValueStringPointer(),
				ID:  s.Key.ValueStringPointer(),
			}
		}),
		Name:          c.Name.ValueString(),
		ContactEmail:  c.ContactEmail.ValueStringPointer(),
		AssociateMode: &associateMode,
		Associates: pie.Map(c.Associates, func(a Associate) platform.AssociateDraft {
			return a.draft()
		}),
		Addresses: pie.Map(c.Addresses, func(a Address) platform.BaseAddress {
			return a.draft()
		}),
		ShippingAddresses: pie.Map(c.ShippingAddressIDs, func(id types.String) int {
			return pie.Int(id.ValueString())
		}),
		DefaultShippingAddress: &dsa,
		BillingAddresses: pie.Map(c.BillingAddressIDs, func(id types.String) int {
			return pie.Int(id.ValueString())
		}),
		DefaultBillingAddress: &dba,
	}
}

func (c Company) updateActions(plan Company) platform.BusinessUnitUpdate {
	result := platform.BusinessUnitUpdate{
		Version: int(c.Version.ValueInt64()),
		Actions: []platform.BusinessUnitUpdateAction{},
	}

	if c.Name.ValueString() != plan.Name.ValueString() {
		result.Actions = append(result.Actions, platform.BusinessUnitChangeNameAction{
			Name: plan.Name.ValueString(),
		})
	}

	if c.ContactEmail.ValueString() != plan.ContactEmail.ValueString() {
		result.Actions = append(result.Actions, platform.BusinessUnitSetContactEmailAction{
			ContactEmail: plan.ContactEmail.ValueStringPointer(),
		})
	}

	if c.Status.ValueString() != plan.Status.ValueString() {
		result.Actions = append(result.Actions, platform.BusinessUnitChangeStatusAction{
			Status: plan.Status.ValueString(),
		})
	}

	if c.DefaultShippingAddressID.ValueString() != plan.DefaultShippingAddressID.ValueString() {
		result.Actions = append(result.Actions, platform.BusinessUnitSetDefaultShippingAddressAction{
			AddressId: plan.DefaultShippingAddressID.ValueStringPointer(),
		})
	}

	if c.DefaultBillingAddressID.ValueString() != plan.DefaultBillingAddressID.ValueString() {
		result.Actions = append(result.Actions, platform.BusinessUnitSetDefaultBillingAddressAction{
			AddressId: plan.DefaultBillingAddressID.ValueStringPointer(),
		})
	}

	if c.AssociateMode.ValueString() != plan.AssociateMode.ValueString() {
		result.Actions = append(result.Actions, platform.BusinessUnitChangeAssociateModeAction{
			AssociateMode: platform.BusinessUnitAssociateMode(plan.AssociateMode.ValueString()),
		})
	}

	if !reflect.DeepEqual(c.Stores, plan.Stores) {
		// find stores to be added
		for _, store := range plan.Stores {
			if !pie.Contains(c.Stores, store) {
				result.Actions = append(result.Actions, platform.BusinessUnitAddStoreAction{
					Store: platform.StoreResourceIdentifier{
						Key: store.Key.ValueStringPointer(),
					},
				})
			}
		}

		// find stores to be removed
		for _, store := range c.Stores {
			if !pie.Contains(plan.Stores, store) {
				result.Actions = append(result.Actions, platform.BusinessUnitRemoveStoreAction{
					Store: platform.StoreResourceIdentifier{
						Key: store.Key.ValueStringPointer(),
					},
				})
			}
		}
	}

	if !reflect.DeepEqual(c.Addresses, plan.Addresses) {
		// find addresses to be added
		for _, address := range plan.Addresses {
			if !pie.Contains(c.Addresses, address) {
				result.Actions = append(result.Actions, platform.BusinessUnitAddAddressAction{
					Address: address.draft(),
				})
			}
		}

		// find addresses to be removed
		for _, address := range c.Addresses {
			if !pie.Contains(plan.Addresses, address) {
				result.Actions = append(result.Actions, platform.BusinessUnitRemoveAddressAction{
					AddressId:  address.ID.ValueStringPointer(),
					AddressKey: address.Key.ValueStringPointer(),
				})
			}
		}
	}

	if !reflect.DeepEqual(c.ShippingAddressIDs, plan.ShippingAddressIDs) {
		// find shipping addresses to be added
		for _, id := range plan.ShippingAddressIDs {
			if !pie.Contains(c.ShippingAddressIDs, id) {
				result.Actions = append(result.Actions, platform.BusinessUnitAddShippingAddressIdAction{
					AddressId: id.ValueStringPointer(),
				})
			}
		}

		// find shipping addresses to be removed
		for _, id := range c.ShippingAddressIDs {
			if !pie.Contains(plan.ShippingAddressIDs, id) {
				result.Actions = append(result.Actions, platform.BusinessUnitRemoveShippingAddressIdAction{
					AddressId: id.ValueStringPointer(),
				})
			}
		}
	}

	if !reflect.DeepEqual(c.BillingAddressIDs, plan.BillingAddressIDs) {
		// find billing addresses to be added
		for _, id := range plan.BillingAddressIDs {
			if !pie.Contains(c.BillingAddressIDs, id) {
				result.Actions = append(result.Actions, platform.BusinessUnitAddBillingAddressIdAction{
					AddressId: id.ValueStringPointer(),
				})
			}
		}

		// find billing addresses to be removed
		for _, id := range c.BillingAddressIDs {
			if !pie.Contains(plan.BillingAddressIDs, id) {
				result.Actions = append(result.Actions, platform.BusinessUnitRemoveBillingAddressIdAction{
					AddressId: id.ValueStringPointer(),
				})
			}
		}
	}

	if !reflect.DeepEqual(c.Associates, plan.Associates) {
		result.Actions = append(result.Actions, platform.BusinessUnitSetAssociatesAction{
			Associates: pie.Map(plan.Associates, func(a Associate) platform.AssociateDraft {
				return a.draft()
			}),
		})
	}

	return result
}

type Division struct {
	ID                       types.String                   `tfsdk:"id"`
	Version                  types.Int64                    `tfsdk:"version"`
	Key                      types.String                   `tfsdk:"key"`
	Status                   types.String                   `tfsdk:"status"`
	ParentUnit               BusinessUnitResourceIdentifier `tfsdk:"parent_unit"`
	Stores                   []StoreKeyReference            `tfsdk:"store"`
	StoreMode                types.String                   `tfsdk:"store_mode"`
	Name                     types.String                   `tfsdk:"name"`
	ContactEmail             types.String                   `tfsdk:"contact_email"`
	Addresses                []Address                      `tfsdk:"address"`
	ShippingAddressIDs       []types.String                 `tfsdk:"shipping_address_ids"`
	DefaultShippingAddressID types.String                   `tfsdk:"default_shipping_address_id"`
	BillingAddressIDs        []types.String                 `tfsdk:"billing_address_ids"`
	DefaultBillingAddressID  types.String                   `tfsdk:"default_billing_address_id"`
	AssociateMode            types.String                   `tfsdk:"associate_mode"`
	Associates               []Associate                    `tfsdk:"associate"`
	InheritedAssociates      []InheritedAssociate           `tfsdk:"inherited_associates"`
}

func (d Division) draft() platform.DivisionDraft {
	mode := platform.BusinessUnitStoreMode(d.StoreMode.ValueString())
	associateMode := platform.BusinessUnitAssociateMode(d.AssociateMode.ValueString())
	status := platform.BusinessUnitStatus(d.Status.ValueString())
	dsa := pie.Int(d.DefaultShippingAddressID.ValueString())
	dba := pie.Int(d.DefaultBillingAddressID.ValueString())

	return platform.DivisionDraft{
		Key:    d.Key.ValueString(),
		Status: &status,
		ParentUnit: platform.BusinessUnitResourceIdentifier{
			ID:  d.ParentUnit.ID.ValueStringPointer(),
			Key: d.ParentUnit.Key.ValueStringPointer(),
		},
		Stores: pie.Map(d.Stores, func(s StoreKeyReference) platform.StoreResourceIdentifier {
			return platform.StoreResourceIdentifier{
				Key: s.Key.ValueStringPointer(),
				ID:  s.Key.ValueStringPointer(),
			}
		}),
		StoreMode:     &mode,
		Name:          d.Name.ValueString(),
		ContactEmail:  d.ContactEmail.ValueStringPointer(),
		AssociateMode: &associateMode,
		Associates: pie.Map(d.Associates, func(a Associate) platform.AssociateDraft {
			return a.draft()
		}),
		Addresses: pie.Map(d.Addresses, func(a Address) platform.BaseAddress {
			return a.draft()
		}),
		ShippingAddresses: pie.Map(d.ShippingAddressIDs, func(id types.String) int {
			return pie.Int(id.ValueString())
		}),
		DefaultShippingAddress: &dsa,
		BillingAddresses: pie.Map(d.BillingAddressIDs, func(id types.String) int {
			return pie.Int(id.ValueString())
		}),
		DefaultBillingAddress: &dba,
	}
}

func (d Division) updateActions(plan Division) platform.BusinessUnitUpdate {
	result := platform.BusinessUnitUpdate{
		Version: int(d.Version.ValueInt64()),
		Actions: []platform.BusinessUnitUpdateAction{},
	}

	if d.Name.ValueString() != plan.Name.ValueString() {
		result.Actions = append(result.Actions, platform.BusinessUnitChangeNameAction{
			Name: plan.Name.ValueString(),
		})
	}

	if d.ContactEmail.ValueString() != plan.ContactEmail.ValueString() {
		result.Actions = append(result.Actions, platform.BusinessUnitSetContactEmailAction{
			ContactEmail: plan.ContactEmail.ValueStringPointer(),
		})
	}

	if d.Status.ValueString() != plan.Status.ValueString() {
		result.Actions = append(result.Actions, platform.BusinessUnitChangeStatusAction{
			Status: plan.Status.ValueString(),
		})
	}

	if d.DefaultShippingAddressID.ValueString() != plan.DefaultShippingAddressID.ValueString() {
		result.Actions = append(result.Actions, platform.BusinessUnitSetDefaultShippingAddressAction{
			AddressId: plan.DefaultShippingAddressID.ValueStringPointer(),
		})
	}

	if d.DefaultBillingAddressID.ValueString() != plan.DefaultBillingAddressID.ValueString() {
		result.Actions = append(result.Actions, platform.BusinessUnitSetDefaultBillingAddressAction{
			AddressId: plan.DefaultBillingAddressID.ValueStringPointer(),
		})
	}

	if d.AssociateMode.ValueString() != plan.AssociateMode.ValueString() {
		result.Actions = append(result.Actions, platform.BusinessUnitChangeAssociateModeAction{
			AssociateMode: platform.BusinessUnitAssociateMode(plan.AssociateMode.ValueString()),
		})
	}

	if d.StoreMode.ValueString() != plan.StoreMode.ValueString() {
		result.Actions = append(result.Actions, platform.BusinessUnitSetStoreModeAction{
			StoreMode: platform.BusinessUnitStoreMode(plan.StoreMode.ValueString()),
		})
	}

	if !reflect.DeepEqual(d.Stores, plan.Stores) {
		// find stores to be added
		for _, store := range plan.Stores {
			if !pie.Contains(d.Stores, store) {
				result.Actions = append(result.Actions, platform.BusinessUnitAddStoreAction{
					Store: platform.StoreResourceIdentifier{
						Key: store.Key.ValueStringPointer(),
					},
				})
			}
		}

		// find stores to be removed
		for _, store := range d.Stores {
			if !pie.Contains(plan.Stores, store) {
				result.Actions = append(result.Actions, platform.BusinessUnitRemoveStoreAction{
					Store: platform.StoreResourceIdentifier{
						Key: store.Key.ValueStringPointer(),
					},
				})
			}
		}

		if !reflect.DeepEqual(d.BillingAddressIDs, plan.BillingAddressIDs) {
			// find billing addresses to be added
			for _, id := range plan.BillingAddressIDs {
				if !pie.Contains(d.BillingAddressIDs, id) {
					result.Actions = append(result.Actions, platform.BusinessUnitAddBillingAddressIdAction{
						AddressId: id.ValueStringPointer(),
					})
				}
			}

			// find billing addresses to be removed
			for _, id := range d.BillingAddressIDs {
				if !pie.Contains(plan.BillingAddressIDs, id) {
					result.Actions = append(result.Actions, platform.BusinessUnitRemoveBillingAddressIdAction{
						AddressId: id.ValueStringPointer(),
					})
				}
			}
		}
	}

	if !reflect.DeepEqual(d.Associates, plan.Associates) {
		result.Actions = append(result.Actions, platform.BusinessUnitSetAssociatesAction{
			Associates: pie.Map(plan.Associates, func(a Associate) platform.AssociateDraft {
				return a.draft()
			}),
		})
	}

	if !reflect.DeepEqual(d.Addresses, plan.Addresses) {
		// find addresses to be added
		for _, address := range plan.Addresses {
			if !pie.Contains(d.Addresses, address) {
				result.Actions = append(result.Actions, platform.BusinessUnitAddAddressAction{
					Address: address.draft(),
				})
			}
		}

		// find addresses to be removed
		for _, address := range d.Addresses {
			if !pie.Contains(plan.Addresses, address) {
				result.Actions = append(result.Actions, platform.BusinessUnitRemoveAddressAction{
					AddressId:  address.ID.ValueStringPointer(),
					AddressKey: address.Key.ValueStringPointer(),
				})
			}
		}
	}

	if !reflect.DeepEqual(d.ShippingAddressIDs, plan.ShippingAddressIDs) {
		// find shipping addresses to be added
		for _, id := range plan.ShippingAddressIDs {
			if !pie.Contains(d.ShippingAddressIDs, id) {
				result.Actions = append(result.Actions, platform.BusinessUnitAddShippingAddressIdAction{
					AddressId: id.ValueStringPointer(),
				})
			}
		}

		// find shipping addresses to be removed
		for _, id := range d.ShippingAddressIDs {
			if !pie.Contains(plan.ShippingAddressIDs, id) {
				result.Actions = append(result.Actions, platform.BusinessUnitRemoveShippingAddressIdAction{
					AddressId: id.ValueStringPointer(),
				})
			}
		}
	}

	return result
}

/*
	Support types for the business unit resource.
*/

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

// BusinessUnitResourceIdentifier is a resource identifier for a business unit.
type BusinessUnitResourceIdentifier struct {
	ID  types.String `tfsdk:"id"`
	Key types.String `tfsdk:"key"`
}

// NewBusinessUnitResourceIdentifierFromNative creates a new BusinessUnitResourceIdentifier
// from a platform.BusinessUnitResourceIdentifier.
func NewBusinessUnitResourceIdentifierFromNative(kr *platform.BusinessUnitResourceIdentifier) BusinessUnitResourceIdentifier {
	return BusinessUnitResourceIdentifier{
		ID:  types.StringValue(*kr.ID),
		Key: types.StringValue(*kr.Key),
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

// StoreKeyReference is a type to model the fields that all types of
// Store Key References have in common.
type StoreKeyReference struct {
	Key    types.String `tfsdk:"key"`
	TypeID types.String `tfsdk:"type_id"`
}
