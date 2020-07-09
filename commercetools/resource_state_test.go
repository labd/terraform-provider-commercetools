package commercetools

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccState_createAndUpdateWithID(t *testing.T) {
	name := "test state"
	key := "test-state"

	newName := "new test state name"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStateConfig(t, name, key, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_state.acctest-state", "name.en", name,
					),
					resource.TestCheckResourceAttr(
						"commercetools_state.acctest-state", "key", key,
					),
				),
			},
			{
				Config: testAccStateConfig(t, newName, key, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_state.acctest-state", "name.en", newName,
					),
					resource.TestCheckResourceAttr(
						"commercetools_state.acctest-state", "key", key,
					),
				),
			},
		},
	})
}

func testAccStateConfig(t *testing.T, name string, key string, addRole bool) string {
	buf := bytes.Buffer{}
	stateConfig := fmt.Sprintf(`
	resource "commercetools_state" "acctest-state" {
		key = "%[2]s"
		type = "ReviewState"
		name = {
			en = "%[1]s"
			nl = "%[1]s"
		}
	`, name, key)
	buf.WriteString(stateConfig)

	if addRole {
		buf.WriteString("roles = [\"ReviewIncludedInStatistics\"]\n")
	}
	buf.WriteString("}")
	newState := buf.String()

	return newState
}

func testAccCheckStateDestroy(s *terraform.State) error {
	return nil
}
