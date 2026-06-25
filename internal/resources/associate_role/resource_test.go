package associate_role_test

import (
	"fmt"
	"strings"
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

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAssociateRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAssociateRoleConfig(id, name, key),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rn, "name", name),
					resource.TestCheckResourceAttr(rn, "key", key),
					resource.TestCheckResourceAttr(rn, "permissions.#", "6"),
				),
			},
			{
				Config: testAssociateRoleConfigUpdate(id, "Sales Manager - DACH", key, true, "AddChildUnits", "my-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rn, "name", "Sales Manager - DACH"),
					resource.TestCheckResourceAttr(rn, "key", key),
					resource.TestCheckResourceAttr(rn, "permissions.#", "7"),
					resource.TestCheckResourceAttr(rn, "permissions.6", "AddChildUnits"),
					resource.TestCheckResourceAttr(rn, "buyer_assignable", "true"),
					resource.TestCheckResourceAttrWith(rn, "custom.type_id", acctest.IsValidUUID),
					resource.TestCheckResourceAttr(rn, "custom.fields.my-field", "my-value"),
				),
			},
		},
	})
}

func TestAssociateRoleResource_ImportPermissions(t *testing.T) {
	resourceName := "commercetools_associate_role.sales_manager_associate_role"

	identifier := "sales_manager_associate_role"
	key := "sales_manager_europe_region"
	name := "Sales Manager - Europe"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAssociateRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAssociateRoleConfig(identifier, name, key),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "6"),
				),
			},
			{
				ResourceName:     resourceName,
				ImportState:      true,
				ImportStateCheck: testAssociateRolePermissionsImported,
			},
		},
	})
}

// testAssociateRolePermissionsImported asserts that every permission from the
// configuration survives the import. The check is order-insensitive to
// check that the permissions are not being dropped entirely, regardless of their ordering.
func testAssociateRolePermissionsImported(states []*terraform.InstanceState) error {
	if len(states) != 1 {
		return fmt.Errorf("expected exactly one imported state, got %d", len(states))
	}

	imported := states[0]
	expectedPermissions := []string{
		"UpdateBusinessUnitDetails",
		"UpdateAssociates",
		"CreateMyCarts",
		"DeleteMyCarts",
		"UpdateMyCarts",
		"ViewMyCarts",
	}

	if count := imported.Attributes["permissions.#"]; count != "6" {
		return fmt.Errorf("expected 6 permissions after import, got %q", count)
	}

	importedPermissions := make(map[string]bool)
	for attribute, value := range imported.Attributes {
		if strings.HasPrefix(attribute, "permissions.") && attribute != "permissions.#" {
			importedPermissions[value] = true
		}
	}
	for _, permission := range expectedPermissions {
		if !importedPermissions[permission] {
			return fmt.Errorf("permission %q missing from imported state", permission)
		}
	}

	return nil
}

func testAssociateRoleDestroy(_ *terraform.State) error {
	return nil
}

func testAssociateRoleConfig(identifier, name, key string) string {
	return utils.HCLTemplate(`
		resource "commercetools_associate_role" "{{ .identifier }}" {
			key = "{{ .key }}"
			buyer_assignable = false
			name = "{{ .name }}"
			permissions = [
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

func testAssociateRoleConfigUpdate(identifier, name, key string, buyerAssign bool, permission string, customValue string) string {
	return utils.HCLTemplate(`
		resource "commercetools_type" "my-type-{{ .identifier }}" {
		  key = "my-type"
		  name = {
			en = "My type"
			nl = "Mijn type"
		  }
		
		  resource_type_ids = ["associate-role"]
		
		  field {
			name = "my-field"
			label = {
			  en = "My field"
			  nl = "Mijn veld"
			}
			type {
			  name = "String"
			}
		  }
		}
	
		resource "commercetools_associate_role" "{{ .identifier }}" {
			key = "{{ .key }}"
			buyer_assignable = {{ .buyer_assignable }}
			name = "{{ .name }}"
			permissions = [
				"UpdateBusinessUnitDetails",
				"UpdateAssociates",
				"CreateMyCarts",
				"DeleteMyCarts",
				"UpdateMyCarts",
				"ViewMyCarts",
				"{{ .permission }}",
			]
	
		   custom {
			 type_id = commercetools_type.my-type-{{ .identifier }}.id
			 fields = {
			   my-field = "{{ .custom_value }}"
			 } 
		   }
		}
	`, map[string]any{
		"identifier":       identifier,
		"name":             name,
		"key":              key,
		"buyer_assignable": buyerAssign,
		"permission":       permission,
		"custom_value":     customValue,
	})
}
