package associate_role_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func TestAssociateRoleResource_Create(t *testing.T) {
	rn := "commercetools_associate_role.sales_manager_associate_role"

	id := "sales_manager_associate_role"
	key := "sales_manager_europe_region"
	name := "Sales Manager - Europe"

	config := testAssociateRoleConfig(id, name, key)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAssociateRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rn, "name", name),
					resource.TestCheckResourceAttr(rn, "key", key),
					resource.TestCheckResourceAttr(rn, "id", id),
					resource.TestCheckResourceAttr(rn, "permissions.#", "7"),
				),
			},
		},
	})
}

func testAssociateRoleDestroy(s *terraform.State) error {
	return nil
}

func testAssociateRoleConfig(identifier, name, key string) string {
	return utils.HCLTemplate(`
		resource "commercetools_associate_role" "{{ .identifier }}" {
			key = "{{ .key }}"
			buyer_assignable = false
			name = "{{ .name }}"
			permissions = [
				"AddChildUnits",
				"UpdateBusinessUnitDetails",
				"UpdateAssociates",
				"CreateMyCarts",
				"DeleteMyCarts",
				"UpdateMyCarts",
				"ViewMyCarts",
			]
		}
	`, map[string]any{
		"identifier": identifier,
		"name":       name,
		"key":        key,
	})
}
