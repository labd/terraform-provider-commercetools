package commercetools

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccState_createAndUpdateWithID(t *testing.T) {
	name := "test state"
	key := "test-state"
	resourceName := "commercetools_state.acctest-state"

	newName := "new test state name"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStateConfig(t, name, key, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name.en", name),
					resource.TestCheckResourceAttr(resourceName, "key", key),
				),
			},
			{
				Config: testAccStateConfig(t, newName, key, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name.en", newName),
					resource.TestCheckResourceAttr(resourceName, "key", key),
				),
			},
		},
	})
}

func testAccStateConfig(t *testing.T, name string, key string, addRole bool) string {
	return hclTemplate(`
		resource "commercetools_state" "acctest-state" {
			key = "{{ .key }}"
			type = "ReviewState"
			name = {
				en = "{{ .name }}"
				nl = "{{ .name }}"
			}

			{{ if .addRole }}
			roles = ["ReviewIncludedInStatistics"]
			{{ end }}
		}
		`,
		map[string]any{
			"key":     key,
			"name":    name,
			"addRole": addRole,
		})
}

func testAccCheckStateDestroy(s *terraform.State) error {
	client := getClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_state" {
			continue
		}
		response, err := client.States().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("state (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		if newErr := checkApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}
