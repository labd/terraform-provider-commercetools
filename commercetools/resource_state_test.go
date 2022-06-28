package commercetools

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
						"commercetools_state.acctest-t1", "transitions.#", "1",
					),
				),
			},
			{
				Config: testAccTransitionsConfig(t, "null"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(
						"commercetools_state.acctest-transitions", "transitions",
					),
				),
			},
			{
				Config: testAccTransitionsConfig(t, "[]"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_state.acctest-transitions", "transitions.#", "0",
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
	resource "commercetools_state" "acctest-t1" {
		depends_on = [commercetools_state.acctest_t2]
		key = "state-a"
		type = "ReviewState"
		name = {
			en = "State #1"
		}
		transitions = [commercetools_state.acctest_t2.id]
	}

	resource "commercetools_state" "acctest_t2" {
		key = "%[1]s"
		type = "ReviewState"
		name = {
			en = "State #2"
		}
		transitions = []
	}
	`, transition)
}

func testAccTransitionsConfig(t *testing.T, transitions string) string {
	return fmt.Sprintf(`
	resource "commercetools_state" "acctest-transitions" {
		key = "state-c"
		type = "ReviewState"
		name = {
			en = "State C"
		}
		transitions = %s
	}
	`, transitions)
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
