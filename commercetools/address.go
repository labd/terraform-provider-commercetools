package commercetools

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
)

func AddressFieldSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"key": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"country": {
					Type:     schema.TypeString,
					Required: true,
				},
				"title": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"salutation": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"first_name": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"last_name": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"street_name": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"street_number": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"additional_street_info": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"postal_code": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"city": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"region": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"state": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"company": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"department": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"building": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"apartment": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"po_box": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"phone": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"mobile": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"email": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"fax": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"additional_address_info": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"external_id": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
}

func CreateAddressFieldDraft(d *schema.ResourceData) *platform.BaseAddress {
	address, err := elementFromList(d, "address")
	if err != nil {
		panic(err)
	}
	if address == nil {
		return nil
	}
	return CreateAddressFieldDraftRaw(address)

}

func CreateAddressFieldDraftRaw(data map[string]any) *platform.BaseAddress {
	if len(data) < 1 {
		return nil
	}

	draft := &platform.BaseAddress{
		Key:                   nilIfEmpty(stringRef(data["key"])),
		Country:               data["country"].(string),
		Title:                 nilIfEmpty(stringRef(data["title"])),
		Salutation:            nilIfEmpty(stringRef(data["salutation"])),
		FirstName:             nilIfEmpty(stringRef(data["first_name"])),
		LastName:              nilIfEmpty(stringRef(data["last_name"])),
		StreetName:            nilIfEmpty(stringRef(data["street_name"])),
		StreetNumber:          nilIfEmpty(stringRef(data["street_number"])),
		AdditionalStreetInfo:  nilIfEmpty(stringRef(data["additional_street_info"])),
		PostalCode:            nilIfEmpty(stringRef(data["postal_code"])),
		City:                  nilIfEmpty(stringRef(data["city"])),
		Region:                nilIfEmpty(stringRef(data["region"])),
		State:                 nilIfEmpty(stringRef(data["state"])),
		Company:               nilIfEmpty(stringRef(data["company"])),
		Department:            nilIfEmpty(stringRef(data["department"])),
		Building:              nilIfEmpty(stringRef(data["building"])),
		Apartment:             nilIfEmpty(stringRef(data["apartment"])),
		POBox:                 nilIfEmpty(stringRef(data["po_box"])),
		Phone:                 nilIfEmpty(stringRef(data["phone"])),
		Mobile:                nilIfEmpty(stringRef(data["mobile"])),
		Email:                 nilIfEmpty(stringRef(data["email"])),
		Fax:                   nilIfEmpty(stringRef(data["fax"])),
		AdditionalAddressInfo: nilIfEmpty(stringRef(data["additional_address_info"])),
		ExternalId:            nilIfEmpty(stringRef(data["external_id"])),
	}

	return draft
}

func flattenAddress(c *platform.Address) []map[string]any {
	if c == nil {
		return nil
	}
	item := map[string]any{
		"key":                     c.Key,
		"country":                 c.Country,
		"title":                   c.Title,
		"salutation":              c.Salutation,
		"first_name":              c.FirstName,
		"last_name":               c.LastName,
		"street_name":             c.StreetName,
		"street_number":           c.StreetNumber,
		"additional_street_info":  c.AdditionalAddressInfo,
		"postal_code":             c.PostalCode,
		"city":                    c.City,
		"region":                  c.Region,
		"state":                   c.State,
		"company":                 c.Company,
		"department":              c.Department,
		"building":                c.Building,
		"apartment":               c.Apartment,
		"po_box":                  c.POBox,
		"phone":                   c.Phone,
		"mobile":                  c.Mobile,
		"email":                   c.Email,
		"fax":                     c.Fax,
		"additional_address_info": c.AdditionalAddressInfo,
		"external_id":             c.ExternalId,
	}
	return []map[string]any{item}
}
