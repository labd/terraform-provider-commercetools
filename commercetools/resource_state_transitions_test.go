package commercetools

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccStateTransitions_create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStateTransitionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStateTransitionsOneTransitionConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"commercetools_state_transitions.test_transition", "from",
					),
					resource.TestCheckResourceAttr(
						"commercetools_state_transitions.test_transition", "to.#", "1",
					),
					// TODO: Could we write a check to assert the set actually contains the state ids?
				),
			},
			{
				Config: testAccStateTransitionsTwoTransitionsConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"commercetools_state_transitions.test_transition", "from",
					),
					resource.TestCheckResourceAttr(
						"commercetools_state_transitions.test_transition", "to.#", "2",
					),
					// TODO: Could we write a check to assert the set actually contains the state ids?
				),
			},
			{
				ResourceName:      "commercetools_state_transitions.test_transition",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStateTransitionsZeroTransitionsConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"commercetools_state_transitions.test_transition", "from",
					),
					resource.TestCheckResourceAttr(
						"commercetools_state_transitions.test_transition", "to.#", "0",
					),
				),
			},
			{
				Config: testAccStateTransitionsComplexConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"commercetools_state_transitions.test_state_a", "from",
					),
					resource.TestCheckResourceAttr(
						"commercetools_state_transitions.test_state_a", "to.#", "1",
					),
					resource.TestCheckResourceAttrSet(
						"commercetools_state_transitions.test_state_b", "from",
					),
					resource.TestCheckResourceAttr(
						"commercetools_state_transitions.test_state_b", "to.#", "1",
					),
					resource.TestCheckResourceAttrSet(
						"commercetools_state_transitions.test_state_c", "from",
					),
					resource.TestCheckResourceAttr(
						"commercetools_state_transitions.test_state_c", "to.#", "0",
					),
				),
			},
		},
	})
}

func testAccStateTransitionsStateFixtures() string {
	return fmt.Sprintf(`
resource "commercetools_state" "test_state_a" {
	key = "state-a"
	type = "ReviewState"
	name = {
		en = "State A"
	}
}

resource "commercetools_state" "test_state_b" {
	key = "state-b"
	type = "ReviewState"
	name = {
		en = "State B"
	}
}

resource "commercetools_state" "test_state_c" {
	key = "state-c"
	type = "ReviewState"
	name = {
		en = "State C"
	}
}
`)
}

func testAccStateTransitionsZeroTransitionsConfig() string {
	stateConfig := testAccStateTransitionsStateFixtures()
	return fmt.Sprintf(`%s

resource "commercetools_state_transitions" "test_transition" {
	from = commercetools_state.test_state_a.id
	to   = []
}
`, stateConfig)
}

func testAccStateTransitionsOneTransitionConfig() string {
	stateConfig := testAccStateTransitionsStateFixtures()
	return fmt.Sprintf(`%s

resource "commercetools_state_transitions" "test_transition" {
	from = commercetools_state.test_state_a.id
	to   = [commercetools_state.test_state_b.id]
}
`, stateConfig)
}

func testAccStateTransitionsTwoTransitionsConfig() string {
	stateConfig := testAccStateTransitionsStateFixtures()
	return fmt.Sprintf(`%s

resource "commercetools_state_transitions" "test_transition" {
	from = commercetools_state.test_state_a.id
	to   = [
		commercetools_state.test_state_b.id,
		commercetools_state.test_state_c.id
	]
}
`, stateConfig)
}

func testAccStateTransitionsComplexConfig() string {
	stateConfig := testAccStateTransitionsStateFixtures()
	return fmt.Sprintf(`%s

resource "commercetools_state_transitions" "test_state_a" {
	from = commercetools_state.test_state_a.id
	to   = [
		commercetools_state.test_state_b.id
	]
}

resource "commercetools_state_transitions" "test_state_b" {
	from = commercetools_state.test_state_b.id
	to   = [
		commercetools_state.test_state_c.id
	]
}

resource "commercetools_state_transitions" "test_state_c" {
	from = commercetools_state.test_state_c.id
	to   = []
}
`, stateConfig)
}

func testAccCheckStateTransitionsDestroy(s *terraform.State) error {
	// TODO: Implement me
	return nil
}
