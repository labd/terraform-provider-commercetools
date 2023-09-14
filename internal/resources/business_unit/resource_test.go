package business_unit_test

import (
	"context"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/terraform-provider-commercetools/internal/acctest"
	"github.com/labd/terraform-provider-commercetools/internal/resources/business_unit"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func TestAssociateRoleSchemaImplementation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schemaRequest := fwresource.SchemaRequest{}
	schemaResponse := &fwresource.SchemaResponse{}

	business_unit.NewResource().Schema(ctx, schemaRequest, schemaResponse)

	if schemaResponse.Diagnostics.HasError() {
		t.Fatalf("Schema method diagnostics: %+v", schemaResponse.Diagnostics)
	}

	// Validate the schema
	diagnostics := schemaResponse.Schema.ValidateImplementation(ctx)

	if diagnostics.HasError() {
		t.Fatalf("Schema validation diagnostics: %+v", diagnostics)
	}
}

func TestBusinessUnitResource_Create(t *testing.T) {
	input := basicBusinessUnitResource()

	r := "commercetools_business_unit.acme_business_unit"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testBusinessUnitDestroy,
		Steps: []resource.TestStep{
			{
				Config: input,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "key", "acme-business-unit"),
					resource.TestCheckResourceAttr(r, "status", "Active"),
					resource.TestCheckResourceAttr(r, "store_mode", "Explicit"),
					resource.TestCheckResourceAttr(r, "unit_type", "Company"),
					resource.TestCheckResourceAttr(r, "stores.#", "2"),
					resource.TestCheckResourceAttr(r, "stores.0.key", "acme-store-dusseldorf"),
					resource.TestCheckResourceAttr(r, "stores.1.key", "acme-store-berlin"),
					resource.TestCheckResourceAttr(r, "addresses.#", "1"),
				),
			},
			{
				Config: updateBusinessUnitResource("Acme Business Unit - Updated", "Inactive"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "key", "acme-business-unit"),
					resource.TestCheckResourceAttr(r, "status", "Inactive"),
					resource.TestCheckResourceAttr(r, "store_mode", "Explicit"),
					resource.TestCheckResourceAttr(r, "unit_type", "Company"),
					resource.TestCheckResourceAttr(r, "stores.#", "2"),
					resource.TestCheckResourceAttr(r, "stores.0.key", "acme-store-dusseldorf"),
					resource.TestCheckResourceAttr(r, "stores.1.key", "acme-store-berlin"),
					resource.TestCheckResourceAttr(r, "addresses.#", "1"),
				),
			},
		},
	})
}

func testBusinessUnitDestroy(_ *terraform.State) error {
	return nil
}

func basicBusinessUnitResource() string {
	return utils.HCLTemplate(`resource "commercetools_business_unit" "acme_business_unit" {
key = "acme-business-unit"
status = "Active"
store_mode = "Explicit"
unit_type = "Company"
name = "Acme Business Unit"

stores {
	key = "acme-store-dusseldorf"
	type_id = "store"
}

stores {
	key = "acme-store-berlin"
	type_id = "store"
}

addresses {
	key = "acme-business-unit-address"
	title = "Acme Business Unit Address"
	salutation = "Mr."
	first_name = "John"
	last_name = "Doe"
	street_name = "Main Street"
	street_number = "1"
	additional_street_info = "Additional Street Info"
	postal_code = "12345"
	city = "Berlin"
	region = "Berlin"
	country = "DE"
	company = "Acme"
	department = "IT"
	building = "Building"
	apartment = "Apartment"
	p_o_box = "P.O. Box"
	phone = "123456789"
	mobile = "987654321"
}

default_shipping_address_id = "acme-business-unit-address"
default_billing_address_id = "acme-business-unit-address"
}`, map[string]any{})
}

func updateBusinessUnitResource(name string, status string) string {
	return utils.HCLTemplate(`resource "commercetools_business_unit" "acme_business_unit" {
key = "acme-business-unit"
status = "{{ .status }}"
store_mode = "Explicit"
unit_type = "Company"
name = "{{ .name }}"

stores {
	key = "acme-store-dusseldorf"
	type_id = "store"
}

stores {
	key = "acme-store-berlin"
	type_id = "store"
}

addresses {
	key = "acme-business-unit-address"
	title = "Acme Business Unit Address"
	salutation = "Mr."
	first_name = "John"
	last_name = "Doe"
	street_name = "Main Street"
	street_number = "1"
	additional_street_info = "Additional Street Info"
	postal_code = "12345"
	city = "Berlin"
	region = "Berlin"
	country = "DE"
	company = "Acme"
	department = "IT"
	building = "Building"
	apartment = "Apartment"
	p_o_box = "P.O. Box"
	phone = "123456789"
	mobile = "987654321"
}

default_shipping_address_id = "acme-business-unit-address"
default_billing_address_id = "acme-business-unit-address"
}`, map[string]any{
		"name":   name,
		"status": status,
	})
}
