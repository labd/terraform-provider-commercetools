package custom_type_test

import (
	"context"
	"fmt"
	"testing"

	acctest "github.com/labd/terraform-provider-commercetools/internal/acctest"
	"github.com/labd/terraform-provider-commercetools/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAccState(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigLoadState(),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						client, err := acctest.GetClient()
						if err != nil {
							return nil
						}
						result, err := client.States().WithKey("test").Get().Execute(context.Background())
						if err != nil {
							return nil
						}
						assert.NotNil(t, result)
						assert.Equal(t, result.Key, "test")
						return nil
					},
				),
			},
		},
	})
}

func testAccCheckDestroy(s *terraform.State) error {
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

func testAccConfigLoadState() string {
	return utils.HCLTemplate(`
	resource "commercetools_state" "test" {
	  key  = "test"
	  type = "ReviewState"
	  name = {
		en = "Unreviewed"
	  }
	  description = {
		en = "Not reviewed yet"
	  }
	  initial = true
	}
	
	data "commercetools_state" "test"  {
		key = "test"
	
		depends_on = [
    		commercetools_state.test
  		]
	}
	`, map[string]any{})
}
