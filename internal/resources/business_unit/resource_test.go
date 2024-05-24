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

func TestCompanySchemaImplementation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schemaRequest := fwresource.SchemaRequest{}
	schemaResponse := &fwresource.SchemaResponse{}

	business_unit.NewCompanyResource().Schema(ctx, schemaRequest, schemaResponse)

	if schemaResponse.Diagnostics.HasError() {
		t.Fatalf("Schema method diagnostics: %+v", schemaResponse.Diagnostics)
	}

	// schema validation
	diagnostics := schemaResponse.Schema.ValidateImplementation(ctx)

	if diagnostics.HasError() {
		t.Fatalf("Schema validation diagnostics: %+v", diagnostics)
	}
}

func TestDivisionSchemaImplementation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schemaRequest := fwresource.SchemaRequest{}
	schemaResponse := &fwresource.SchemaResponse{}

	business_unit.NewDivisionResource().Schema(ctx, schemaRequest, schemaResponse)

	if schemaResponse.Diagnostics.HasError() {
		t.Fatalf("Schema method diagnostics: %+v", schemaResponse.Diagnostics)
	}

	// schema validation
	diagnostics := schemaResponse.Schema.ValidateImplementation(ctx)

	if diagnostics.HasError() {
		t.Fatalf("Schema validation diagnostics: %+v", diagnostics)
	}
}

func TestBusinessUnitResource_Company(t *testing.T) {
	r := "commercetools_business_unit_company.acme_company"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testBusinessUnitDestroy,
		Steps: []resource.TestStep{
			{
				Config: businessUnitTFResourceDef("", "", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "key", "acme-company"),
					resource.TestCheckResourceAttr(r, "name", "Acme Company Business Unit"),
					resource.TestCheckResourceAttr(r, "status", "Active"),
					resource.TestCheckResourceAttr(r, "stores.#", "2"),
					resource.TestCheckResourceAttr(r, "stores.0.key", "acme-usa"),
					resource.TestCheckResourceAttr(r, "stores.1.key", "acme-germany"),
					resource.TestCheckResourceAttr(r, "addresses.#", "1"),
				),
			},
			{
				Config: businessUnitTFResourceDef("Acme Business Unit - Updated", "Inactive", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "key", "acme-company"),
					resource.TestCheckResourceAttr(r, "status", "Inactive"),
					resource.TestCheckResourceAttr(r, "stores.#", "2"),
					resource.TestCheckResourceAttr(r, "stores.0.key", "acme-usa"),
					resource.TestCheckResourceAttr(r, "stores.1.key", "acme-germany"),
					resource.TestCheckResourceAttr(r, "addresses.#", "1"),
				),
			},
		},
	})
}

func testBusinessUnitDestroy(_ *terraform.State) error {
	return nil
}

func businessUnitTFResourceDef(name, status, email string) string {
	if status == "" {
		status = "Active"
	}

	if email == "" {
		email = "acme@example.com"
	}

	if name == "" {
		name = "Acme Company Business Unit"
	}

	return utils.HCLTemplate(`resource "commercetools_business_unit_company" "acme_company" {
    key              = "acme-company"
    name             = {{ .name }}
    status           = {{ .status }}
    contact_email    = {{ .email}}

    store {
        key = "acme-usa"
        type_id = "store"
    }

    store {
        key = "acme-germany"
        type_id = "store"
    }

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
    default_shipping_address_id     = "acme-business-unit-address"
    default_billing_address_id      = "acme-business-unit-address"
}`, map[string]any{
		"name":   name,
		"status": status,
		"email":  email,
	})
}
