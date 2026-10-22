package recurrence_policy_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func TestAccRecurrencePolicyCreate_basic(t *testing.T) {
	resourceName := "commercetools_recurrence_policy.acctest"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecurrencePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRecurrencePolicyConfig("monthly", `
					type          = "standard"
					value         = 1
					interval_unit = "Months"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "monthly"),
					resource.TestCheckResourceAttr(resourceName, "name.en", "Replenishment"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.type", "standard"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.value", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.interval_unit", "Months"),
				),
			},
			{
				Config: testAccRecurrencePolicyConfig("first-of-month", `
					type = "dayOfMonth"
					day  = 1`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "first-of-month"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.type", "dayOfMonth"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.day", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "schedule.0.value"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckRecurrencePolicyDestroy(s *terraform.State) error {
	client, err := acctest.GetClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_recurrence_policy" {
			continue
		}
		response, err := client.RecurrencePolicies().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("recurrence policy (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		if newErr := acctest.CheckApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}

func testAccRecurrencePolicyConfig(key, schedule string) string {
	return utils.HCLTemplate(`
		resource "commercetools_recurrence_policy" "acctest" {
			key  = "{{ .key }}"
			name = {
				"en" = "Replenishment"
			}
			schedule {
				{{ .schedule }}
			}
		}`, map[string]any{
		"key":      key,
		"schedule": schedule,
	})
}
