package state_transition_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func TestAccStateTransitions_createAndUpdateWithID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStateTransitionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: strings.Join(
					[]string{
						testAccStateTransitionsNone("acctest-state-1", "state-1"),
						testAccStateTransitionsNone("acctest-state-2", "state-2"),
					},
					"\n\n",
				),
				Check: resource.ComposeTestCheckFunc(),
			},
			{
				Config: strings.Join(
					[]string{
						testAccStateTransitionsNone("acctest-state-1", "state-1"),
						testAccStateTransitionsNone("acctest-state-2", "state-2"),
						testAccStateTransitionsConfig("acctest-transition-1",
							"commercetools_state.acctest-state-1.id",
							[]string{"commercetools_state.acctest-state-2.id"}),
						testAccStateTransitionsConfig("acctest-transition-2",
							"commercetools_state.acctest-state-2.id",
							[]string{"commercetools_state.acctest-state-1.id"}),
					},
					"\n\n",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("commercetools_state_transitions.acctest-transition-1", "to.#", "1"),
					resource.TestCheckResourceAttr("commercetools_state_transitions.acctest-transition-2", "to.#", "1"),
				),
			},
			{
				Config: strings.Join(
					[]string{
						testAccStateTransitionsNone("acctest-state-1", "state-1"),
						testAccStateTransitionsNone("acctest-state-2", "state-2"),
						testAccStateTransitionsConfig("acctest-transition-1",
							"commercetools_state.acctest-state-1.id",
							[]string{}),
					},
					"\n\n",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("commercetools_state_transitions.acctest-transition-1", "to.#", "0"),
				),
			},
		},
	})
}

func testAccStateTransitionsNone(identifier string, key string) string {
	return utils.HCLTemplate(`
		resource "commercetools_state" "{{ .identifier }}" {
			key = "{{ .key }}"
			type = "ReviewState"
			name = {
				en = "State C"
			}
		}`,
		map[string]any{
			"identifier": identifier,
			"key":        key,
		})
}

func testAccStateTransitionsConfig(identifier string, from string, to []string) string {
	return utils.HCLTemplate(`
		resource "commercetools_state_transitions" "{{ .identifier }}" {
			from = {{ .from }}
			to = {{ .to | printf "%s" }}
		}`,
		map[string]any{
			"identifier": identifier,
			"from":       from,
			"to":         to,
		})
}

func testAccCheckStateTransitionsDestroy(s *terraform.State) error {
	client, err := acctest.GetClient()
	if err != nil {
		return err
	}

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
		if newErr := acctest.CheckApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}
