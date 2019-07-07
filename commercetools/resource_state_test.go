package commercetools

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccState_createAndUpdateWithID(t *testing.T) {

	name := "test state"
	key := "test-state"

	newName := "new test state name"

	transition := "state-b"

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
			{
				Config: testAccTransitionConfig(t, transition),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_state.acctest-state-a", "transitions.#", "1",
					),
					resource.TestCheckResourceAttr(
						"commercetools_state.acctest-state-a", "transitions.0", transition,
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

func testAccTransitionConfig(t *testing.T, transition string) string {
	return fmt.Sprintf(`
	resource "commercetools_state" "acctest-transitions" {
		key = "state-a"
		type = "ReviewState"
		name = {
			en = "State A"
		}
		transitions = ["%s"]
	}
	`, transition)
}

func testAccCheckStateDestroy(s *terraform.State) error {
	return nil
}
