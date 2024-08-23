package business_unit_company_test

import (
	"context"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/internal/resources/business_unit_company"
	"regexp"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func TestBusinessUnitCompanySchemaImplementation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schemaRequest := fwresource.SchemaRequest{}
	schemaResponse := &fwresource.SchemaResponse{}

	business_unit_company.NewCompanyResource().Schema(ctx, schemaRequest, schemaResponse)

	if schemaResponse.Diagnostics.HasError() {
		t.Fatalf("Schema method diagnostics: %+v", schemaResponse.Diagnostics)
	}

	// schema validation
	diagnostics := schemaResponse.Schema.ValidateImplementation(ctx)

	if diagnostics.HasError() {
		t.Fatalf("Schema validation diagnostics: %+v", diagnostics)
	}
}

func TestBusinessUnitResource(t *testing.T) {
	r := "commercetools_business_unit_company.acme_company"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testBusinessUnitDestroy,
		Steps: []resource.TestStep{
			{
				Config: businessUnitTFResourceDef(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "key", "acme-company"),
					resource.TestCheckResourceAttr(r, "name", "Acme Company Business Unit"),
					resource.TestCheckResourceAttr(r, "status", string(platform.BusinessUnitConfigurationStatusActive)),
					resource.TestCheckResourceAttr(r, "contact_email", "acme@example.com"),
					resource.TestCheckResourceAttr(r, "address.#", "1"),
				),
			},
			{
				Config:      businessUnitTFResourceDef(withBusinessUnitCompanyKey("acme-company-updated")),
				ExpectError: regexp.MustCompile(`key is immutable`),
			},
			{
				Config: businessUnitTFResourceDef(withBusinessUnitCompanyName("Acme Business Unit - Updated")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "name", "Acme Business Unit - Updated"),
				),
			},
			{
				Config: businessUnitTFResourceDef(withBusinessUnitCompanyStatus(platform.BusinessUnitConfigurationStatusInactive)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "status", string(platform.BusinessUnitConfigurationStatusInactive)),
				),
			},
			{
				Config: businessUnitTFResourceDef(withBusinessUnitCompanyContactEmail("acme-updated@example.com")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "contact_email", "acme-updated@example.com"),
				),
			},
		},
	})
}

func testBusinessUnitDestroy(_ *terraform.State) error {
	return nil
}

type option func(map[string]interface{})

func withBusinessUnitCompanyKey(key string) option {
	return func(data map[string]interface{}) {
		data["key"] = key
	}
}

func withBusinessUnitCompanyName(name string) option {
	return func(data map[string]interface{}) {
		data["name"] = name
	}
}

func withBusinessUnitCompanyStatus(status platform.BusinessUnitConfigurationStatus) option {
	return func(data map[string]interface{}) {
		data["status"] = status
	}
}

func withBusinessUnitCompanyContactEmail(email string) option {
	return func(data map[string]interface{}) {
		data["contact_email"] = email
	}
}

func businessUnitTFResourceDef(options ...option) string {
	data := map[string]interface{}{
		"key":           "acme-company",
		"status":        platform.BusinessUnitConfigurationStatusActive,
		"contact_email": "acme@example.com",
		"name":          "Acme Company Business Unit",
	}

	for _, option := range options {
		option(data)
	}

	return utils.HCLTemplate(`	
	resource "commercetools_business_unit_company" "acme_company" {
		key              = "{{ .key }}"
		name             = "{{ .name }}"
		status           = "{{ .status }}"
		contact_email    = "{{ .contact_email }}"
	
		address {
			key                     = "acme-business-unit-address"
			title                   = "Acme Business Unit Address"
			salutation              = "Mr."
			first_name              = "John"
			last_name               = "Doe"
			street_name             = "Main Street"
			street_number           = "1"
			additional_street_info  = "Additional Street Info"
			postal_code             = "12345"
			city                    = "Berlin"
			region                  = "Berlin"
			country                 = "DE"
			company                 = "Acme"
			department              = "IT"
			building                = "Building"
			apartment               = "Apartment"
			po_box                  = "P.O. Box"
			phone                   = "123456789"
			mobile                  = "987654321"
		}
		shipping_address_keys = ["acme-business-unit-address"]
		billing_address_keys = ["acme-business-unit-address"]
		default_shipping_address_key     = "acme-business-unit-address"
		default_billing_address_key      = "acme-business-unit-address"
	}
	`, data)
}
