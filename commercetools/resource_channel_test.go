package commercetools

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/labd/commercetools-go-sdk/commercetools"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccChannelCreate_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "roles.0", "Primary",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "key", "standard-key",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "custom.0.type_key", "channel-test",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "custom.0.field.0.key", "carrier",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "custom.0.field.0.value", "\"example\"",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "custom.0.field.1.key", "meal",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "custom.0.field.1.value", "{\"en-GB\":\"lunch\"}",
					),
				),
			},
			{
				Config: testAccChannelUpdateConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "roles.0", "Primary",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "key", "standard-key",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "custom.0.type_key", "channel-test",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "custom.0.field.0.key", "carrier",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "custom.0.field.0.value", "\"dhl\"",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "custom.0.field.1.key", "meal",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "custom.0.field.1.value", "{\"de-DE\":\"Mittag\",\"en-GB\":\"lunch\"}",
					),
				),
			},
		},
	})
}

func TestAccChannelCreate_updateCustom(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfigWithoutCustom(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "roles.0", "Primary",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "key", "standard-key",
					),
				),
			},
			{
				Config: testAccChannelUpdateConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "roles.0", "Primary",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "key", "standard-key",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "custom.0.type_key", "channel-test",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "custom.0.field.0.key", "carrier",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "custom.0.field.0.value", "\"dhl\"",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "custom.0.field.1.key", "meal",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "custom.0.field.1.value", "{\"de-DE\":\"Mittag\",\"en-GB\":\"lunch\"}",
					),
				),
			},
		},
	})
}

func testAccChannelConfig() string {
	return `

resource "commercetools_type" "channel_test" {
	key = "channel-test"

	resource_type_ids = ["channel"]

	field {
		name = "carrier"
		label = {
			en = "Skype name"
			nl = "Skype naam"
		}
		type {
			name = "String"
		}
	}

	field {
		name = "meal"
		label = {
			en = "Skype name"
			nl = "Skype naam"
		}
		type {
			name = "LocalizedString"
		}
	}

	name = {
		en = "Contact info"
		nl = "Contact informatie"
	}
	description = {
		en = "All things related communication"
		nl = "Alle communicatie-gerelateerde zaken"
	}
}

resource "commercetools_channel" "standard" {
	 depends_on = [
		commercetools_type.channel_test,
	  ]
	roles = ["Primary"]
	key  = "standard-key"
 	custom {
		type_key = "channel-test"
		field {
		  key = "carrier"
		  value = jsonencode("example")
		}

		field {
		  key = "meal"
		  value = jsonencode({
			"en-GB": "lunch",
		  })
		}
	}
}
`
}

func testAccChannelConfigWithoutCustom() string {
	return `

resource "commercetools_type" "channel_test" {
	key = "channel-test"

	resource_type_ids = ["channel"]

	field {
		name = "carrier"
		label = {
			en = "Skype name"
			nl = "Skype naam"
		}
		type {
			name = "String"
		}
	}

	field {
		name = "meal"
		label = {
			en = "Skype name"
			nl = "Skype naam"
		}
		type {
			name = "LocalizedString"
		}
	}

	name = {
		en = "Contact info"
		nl = "Contact informatie"
	}
	description = {
		en = "All things related communication"
		nl = "Alle communicatie-gerelateerde zaken"
	}
}

resource "commercetools_channel" "standard" {
	 depends_on = [
		commercetools_type.channel_test,
	  ]
	roles = ["Primary"]
	key  = "standard-key"
}
`
}

func testAccChannelUpdateConfig() string {
	return `

resource "commercetools_type" "channel_test" {
	key = "channel-test"

	resource_type_ids = ["channel"]

	field {
		name = "carrier"
		label = {
			en = "Skype name"
			nl = "Skype naam"
		}
		type {
			name = "String"
		}
	}

	field {
		name = "meal"
		label = {
			en = "Skype name"
			nl = "Skype naam"
		}
		type {
			name = "LocalizedString"
		}
	}

	name = {
		en = "Contact info"
		nl = "Contact informatie"
	}
	description = {
		en = "All things related communication"
		nl = "Alle communicatie-gerelateerde zaken"
	}
}

resource "commercetools_channel" "standard" {
	 depends_on = [
		commercetools_type.channel_test,
	  ]
	roles = ["Primary"]
	key  = "standard-key"
 	custom {
		type_key = "channel-test"

		field {
		  key = "carrier"
		  value = jsonencode("dhl")
		}

		field {
		  key = "meal"
		  value = jsonencode({
			"en-GB": "lunch",
			"de-DE": "Mittag",
		  })
		}
	}
}
`
}

func testAccCheckChannelDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*commercetools.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_channel" {
			continue
		}

		response, err := conn.ChannelGetWithID(context.Background(), rs.Primary.ID)
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("channel (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		// If we don't get a was not found error, return the actual error. Otherwise resource is destroyed
		if !strings.Contains(err.Error(), "was not found") && !strings.Contains(err.Error(), "Not Found (404)") {
			return err
		}
	}

	typeErr := testAccCheckTypesDestroy(s)

	if typeErr != nil {
		return typeErr
	}

	return nil
}
