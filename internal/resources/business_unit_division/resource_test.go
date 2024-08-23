package business_unit_division_test

import (
	"context"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/internal/resources/business_unit_division"
	"regexp"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func TestDivisionSchemaImplementation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schemaRequest := fwresource.SchemaRequest{}
	schemaResponse := &fwresource.SchemaResponse{}

	business_unit_division.NewDivisionResource().Schema(ctx, schemaRequest, schemaResponse)

	if schemaResponse.Diagnostics.HasError() {
		t.Fatalf("Schema method diagnostics: %+v", schemaResponse.Diagnostics)
	}

	// schema validation
	diagnostics := schemaResponse.Schema.ValidateImplementation(ctx)

	if diagnostics.HasError() {
		t.Fatalf("Schema validation diagnostics: %+v", diagnostics)
	}
}

func TestBusinessUnitResource_Division(t *testing.T) {
	r := "commercetools_business_unit_division.acme_division"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testBusinessUnitDivisionDestroy,
		Steps: []resource.TestStep{
			{
				Config: businessUnitDivisionTFResourceDef(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "key", "acme-division"),
					resource.TestCheckResourceAttr(r, "name", "Acme Company Business Unit"),
					resource.TestCheckResourceAttr(r, "contact_email", "acme@example.com"),
					resource.TestCheckResourceAttr(r, "status", "Active"),
					resource.TestCheckResourceAttr(r, "address.#", "1"),
				),
			},
			{
				Config:      businessUnitDivisionTFResourceDef(withBusinessDivisionKey("acme-division-updated")),
				ExpectError: regexp.MustCompile(`key is immutable`),
			},
			{
				Config: businessUnitDivisionTFResourceDef(withBusinessUnitDivisionName("Acme Business Unit - Updated")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "name", "Acme Business Unit - Updated"),
				),
			},
			{
				Config: businessUnitDivisionTFResourceDef(withBusinessUnitDivisionContactEmail("acme-update@example.com")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "contact_email", "acme-update@example.com"),
				),
			},
			{
				Config: businessUnitDivisionTFResourceDef(withBusinessUnitDivisionStatus(platform.BusinessUnitConfigurationStatusInactive)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "status", string(platform.BusinessUnitConfigurationStatusInactive)),
				),
			},
			{
				Config: businessUnitDivisionTFResourceDef(withBusinessUnitDivisionAssociateMode(platform.BusinessUnitAssociateModeExplicit)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "associate_mode", string(platform.BusinessUnitAssociateModeExplicit)),
				),
			},
			{
				Config: businessUnitDivisionTFResourceDef(withBusinessUnitDivisionStoreMode(platform.BusinessUnitStoreModeExplicit)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "store_mode", string(platform.BusinessUnitStoreModeExplicit)),
				),
			},
			{
				Config: businessUnitDivisionTFResourceDef(withBusinessUnitDivisionApprovalRuleMode(platform.BusinessUnitApprovalRuleModeExplicit)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "approval_rule_mode", string(platform.BusinessUnitApprovalRuleModeExplicit)),
				),
			},
		},
	})
}

func testBusinessUnitDivisionDestroy(_ *terraform.State) error {
	return nil
}

type option func(map[string]interface{})

func withBusinessDivisionKey(key string) option {
	return func(data map[string]interface{}) {
		data["key"] = key
	}
}

func withBusinessUnitDivisionName(name string) option {
	return func(data map[string]interface{}) {
		data["name"] = name
	}
}

func withBusinessUnitDivisionStatus(status platform.BusinessUnitConfigurationStatus) option {
	return func(data map[string]interface{}) {
		data["status"] = status
	}
}

func withBusinessUnitDivisionContactEmail(email string) option {
	return func(data map[string]interface{}) {
		data["contact_email"] = email
	}
}

func withBusinessUnitDivisionStoreMode(storeMode platform.BusinessUnitStoreMode) option {
	return func(data map[string]interface{}) {
		data["store_mode"] = storeMode
	}
}

func withBusinessUnitDivisionAssociateMode(associateMode platform.BusinessUnitAssociateMode) option {
	return func(data map[string]interface{}) {
		data["associate_mode"] = associateMode
	}
}

func withBusinessUnitDivisionApprovalRuleMode(approvalRuleMode platform.BusinessUnitApprovalRuleMode) option {
	return func(data map[string]interface{}) {
		data["approval_rule_mode"] = approvalRuleMode
	}
}

func businessUnitDivisionTFResourceDef(options ...option) string {
	data := map[string]interface{}{
		"key":                "acme-division",
		"status":             platform.BusinessUnitConfigurationStatusActive,
		"contact_email":      "acme@example.com",
		"name":               "Acme Company Business Unit",
		"store_mode":         platform.BusinessUnitStoreModeFromParent,
		"associate_mode":     platform.BusinessUnitAssociateModeExplicitAndFromParent,
		"approval_rule_mode": platform.BusinessUnitApprovalRuleModeExplicitAndFromParent,
	}

	for _, option := range options {
		option(data)
	}

	return utils.HCLTemplate(`
	resource "commercetools_business_unit_company" "acme_company" {
		key              = "acme-company"
		name             = "Acme Company"
		status           = "Active"
	}
	
	resource "commercetools_business_unit_division" "acme_division" {
		key                = "{{ .key }}"
		name               = "{{ .name }}"
		status             = "{{ .status }}"
		contact_email      = "{{ .contact_email }}"
		store_mode         = "{{ .store_mode }}"
		associate_mode     = "{{ .associate_mode }}"
		approval_rule_mode = "{{ .approval_rule_mode }}"
	
		parent_unit {
			key = commercetools_business_unit_company.acme_company.key
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
		shipping_address_keys = ["acme-business-unit-address"]
		billing_address_keys = ["acme-business-unit-address"]
		default_shipping_address_key     = "acme-business-unit-address"
		default_billing_address_key      = "acme-business-unit-address"
	}
	`, data)
}
