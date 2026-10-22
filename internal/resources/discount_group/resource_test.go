package discount_group_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func TestAccDiscountGroupCreate_basic(t *testing.T) {
	resourceName := "commercetools_discount_group.acctest"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDiscountGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDiscountGroupConfig("summer-sale", "Summer sale", "0.6", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "summer-sale"),
					resource.TestCheckResourceAttr(resourceName, "name.en", "Summer sale"),
					resource.TestCheckResourceAttr(resourceName, "sort_order", "0.6"),
					resource.TestCheckResourceAttr(resourceName, "is_active", "true"),
				),
			},
			{
				Config: testAccDiscountGroupConfig("summer-sale-2026", "Summer sale 2026", "0.65", "is_active = false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "summer-sale-2026"),
					resource.TestCheckResourceAttr(resourceName, "name.en", "Summer sale 2026"),
					resource.TestCheckResourceAttr(resourceName, "sort_order", "0.65"),
					resource.TestCheckResourceAttr(resourceName, "is_active", "false"),
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

func testAccCheckDiscountGroupDestroy(s *terraform.State) error {
	client, err := acctest.GetClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_discount_group" {
			continue
		}
		response, err := client.DiscountGroups().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("discount group (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		if newErr := acctest.CheckApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}

func testAccDiscountGroupConfig(key, name, sortOrder, extra string) string {
	return utils.HCLTemplate(`
		resource "commercetools_discount_group" "acctest" {
			key        = "{{ .key }}"
			name       = {
				"en" = "{{ .name }}"
			}
			sort_order = "{{ .sortOrder }}"
			{{ .extra }}
		}`, map[string]any{
		"key":       key,
		"name":      name,
		"sortOrder": sortOrder,
		"extra":     extra,
	})
}

func TestAccCartDiscountInDiscountGroup(t *testing.T) {
	resourceName := "commercetools_cart_discount.grouped"
	config := func(group string) string {
		return `
resource "commercetools_discount_group" "summer" {
  key        = "summer"
  sort_order = "0.6"
}

resource "commercetools_discount_group" "winter" {
  key        = "winter"
  sort_order = "0.7"
}

resource "commercetools_cart_discount" "grouped" {
  name = {
    en = "Ten percent off"
  }
  predicate         = "1 = 1"
  discount_group_id = ` + group + `
  target {
    type      = "lineItems"
    predicate = "1 = 1"
  }
  value {
    type      = "relative"
    permyriad = 1000
  }
}
`
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDiscountGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: config("commercetools_discount_group.summer.id"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "discount_group_id", "commercetools_discount_group.summer", "id"),
					// A discount in a group takes the group's sort order.
					resource.TestCheckResourceAttr(resourceName, "sort_order", "0.6"),
				),
			},
			{
				Config: config("commercetools_discount_group.winter.id"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "discount_group_id", "commercetools_discount_group.winter", "id"),
					resource.TestCheckResourceAttr(resourceName, "sort_order", "0.7"),
				),
			},
		},
	})
}
