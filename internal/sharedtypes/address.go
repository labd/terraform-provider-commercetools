package sharedtypes

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
	"slices"
)

var (
	AddressBlockSchema = schema.ListNestedBlock{
		MarkdownDescription: "Addresses used by the Business Unit.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					MarkdownDescription: "Unique identifier of the Address",
					Computed:            true,
				},
				"key": schema.StringAttribute{
					MarkdownDescription: "User-defined identifier of the Address that must be unique when multiple " +
						"addresses are referenced in BusinessUnits, Customers, and itemShippingAddresses " +
						"(LineItem-specific addresses) of a Cart, Order, QuoteRequest, or Quote.",
					Required: true,
				},
				"external_id": schema.StringAttribute{
					MarkdownDescription: "ID for the contact used in an external system",
					Optional:            true,
				},
				"country": schema.StringAttribute{
					MarkdownDescription: "Name of the country",
					Required:            true,
				},
				"title": schema.StringAttribute{
					MarkdownDescription: "Title of the contact, for example Dr., Prof.",
					Optional:            true,
				},
				"salutation": schema.StringAttribute{
					MarkdownDescription: "Salutation of the contact, for example Ms., Mr.",
					Optional:            true,
				},
				"first_name": schema.StringAttribute{
					MarkdownDescription: "First name of the contact",
					Optional:            true,
				},
				"last_name": schema.StringAttribute{
					MarkdownDescription: "Last name of the contact",
					Optional:            true,
				},
				"street_name": schema.StringAttribute{
					MarkdownDescription: "Name of the street",
					Optional:            true,
				},
				"street_number": schema.StringAttribute{
					MarkdownDescription: "Street number",
					Optional:            true,
				},
				"additional_street_info": schema.StringAttribute{
					MarkdownDescription: "Further information on the street address",
					Optional:            true,
				},
				"postal_code": schema.StringAttribute{
					MarkdownDescription: "Postal code",
					Optional:            true,
				},
				"city": schema.StringAttribute{
					MarkdownDescription: "Name of the city",
					Optional:            true,
				},
				"region": schema.StringAttribute{
					MarkdownDescription: "Name of the region",
					Optional:            true,
				},
				"state": schema.StringAttribute{
					MarkdownDescription: "Name of the state",
					Optional:            true,
				},
				"company": schema.StringAttribute{
					MarkdownDescription: "Name of the company",
					Optional:            true,
				},
				"department": schema.StringAttribute{
					MarkdownDescription: "Name of the department",
					Optional:            true,
				},
				"building": schema.StringAttribute{
					MarkdownDescription: "Name or number of the building",
					Optional:            true,
				},
				"apartment": schema.StringAttribute{
					MarkdownDescription: "Name or number of the apartment",
					Optional:            true,
				},
				"po_box": schema.StringAttribute{
					MarkdownDescription: "Post office box number",
					Optional:            true,
				},
				"phone": schema.StringAttribute{
					MarkdownDescription: "Phone number",
					Optional:            true,
				},
				"mobile": schema.StringAttribute{
					MarkdownDescription: "Mobile phone number",
					Optional:            true,
				},
				"email": schema.StringAttribute{
					MarkdownDescription: "Email address",
					Optional:            true,
				},
				"fax": schema.StringAttribute{
					MarkdownDescription: "Fax number",
					Optional:            true,
				},
				"additional_address_info": schema.StringAttribute{
					MarkdownDescription: "Further information on the Address",
					Optional:            true,
				},
			},
		},
	}
)

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

func (a Address) Equal(other Address) bool {
	return a.Key.Equal(other.Key) &&
		a.ExternalID.Equal(other.ExternalID) &&
		a.Country.Equal(other.Country) &&
		a.Title.Equal(other.Title) &&
		a.Salutation.Equal(other.Salutation) &&
		a.FirstName.Equal(other.FirstName) &&
		a.LastName.Equal(other.LastName) &&
		a.StreetName.Equal(other.StreetName) &&
		a.StreetNumber.Equal(other.StreetNumber) &&
		a.AdditionalStreetInfo.Equal(other.AdditionalStreetInfo) &&
		a.PostalCode.Equal(other.PostalCode) &&
		a.City.Equal(other.City) &&
		a.Region.Equal(other.Region) &&
		a.State.Equal(other.State) &&
		a.Company.Equal(other.Company) &&
		a.Department.Equal(other.Department) &&
		a.Building.Equal(other.Building) &&
		a.Apartment.Equal(other.Apartment) &&
		a.POBox.Equal(other.POBox) &&
		a.Phone.Equal(other.Phone) &&
		a.Mobile.Equal(other.Mobile) &&
		a.Email.Equal(other.Email) &&
		a.Fax.Equal(other.Fax) &&
		a.AdditionalAddressInfo.Equal(other.AdditionalAddressInfo)
}

func (a Address) Draft() platform.BaseAddress {
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

type DeleteAddressAction interface {
	platform.BusinessUnitRemoveAddressAction
}

type AddAddressAction interface {
	platform.BusinessUnitAddAddressAction
}

type ChangeAddressAction interface {
	platform.BusinessUnitChangeAddressAction
}

func AddressesAddActions[A AddAddressAction](currentAddresses []Address, plannedAddresses []Address) []any {
	var actions []any

	for _, pa := range plannedAddresses {
		if !slices.ContainsFunc(currentAddresses, func(ca Address) bool { return pa.Key.Equal(ca.Key) }) {
			actions = append(actions, A{
				Address: pa.Draft(),
			})
		}
	}

	return actions
}

func AddressesDeleteActions[D DeleteAddressAction](currentAddresses []Address, plannedAddresses []Address) []any {
	var actions []any

	for _, ca := range currentAddresses {
		if !slices.ContainsFunc(plannedAddresses, func(pa Address) bool { return ca.Key.Equal(pa.Key) }) {
			actions = append(actions, D{
				AddressKey: ca.Key.ValueStringPointer(),
			})
		}
	}

	return actions
}

func AddressesChangeActions[C ChangeAddressAction](currentAddresses []Address, plannedAddresses []Address) []any {
	var actions []any

	for _, ca := range currentAddresses {
		pai := slices.IndexFunc(plannedAddresses, func(pa Address) bool { return ca.Key.Equal(pa.Key) })
		if pai == -1 {
			continue
		}

		pa := plannedAddresses[pai]

		if !ca.Equal(pa) {
			actions = append(actions, C{
				AddressKey: pa.Key.ValueStringPointer(),
				Address:    pa.Draft(),
			})
		}
	}

	return actions
}
