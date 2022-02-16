package commercetools

import (
	"context"
	"fmt"
	"testing"

	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
						"commercetools_channel.standard", "custom.0.field.0.name", "carrier",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "custom.0.field.0.value", "\"example\"",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "custom.0.field.1.name", "meal",
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
						"commercetools_channel.standard", "custom.0.field.1.name", "carrier",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "custom.0.field.1.value", "\"dhl\"",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "custom.0.field.0.name", "meal",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "custom.0.field.0.value", "{\"de-DE\":\"Mittag\",\"en-GB\":\"lunch\"}",
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
						"commercetools_channel.standard", "custom.0.field.1.name", "carrier",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "custom.0.field.1.value", "\"dhl\"",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "custom.0.field.0.name", "meal",
					),
					resource.TestCheckResourceAttr(
						"commercetools_channel.standard", "custom.0.field.0.value", "{\"de-DE\":\"Mittag\",\"en-GB\":\"lunch\"}",
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
		  name = "carrier"
		  value = jsonencode("example")
		}

		field {
		  name = "meal"
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
		  name = "meal"
		  value = jsonencode({
			"en-GB": "lunch",
			"de-DE": "Mittag",
		  })
		}

		field {
		  name = "carrier"
		  value = jsonencode("dhl")
		}
	}
}
`
}

func testAccCheckChannelDestroy(s *terraform.State) error {
	client := getClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_channel" {
			continue
		}

		response, err := client.Channels().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("channel (%s) still exists", rs.Primary.ID)
			}
			return nil
		}

		// If we don't get 404 error we return the error, otherwise the ressource was destroyed
		if requestError, ok := err.(platform.GenericRequestError); !ok || requestError.StatusCode != 404 {
			return err
		}
	}

	typeErr := testAccCheckTypesDestroy(s)

	if typeErr != nil {
		return typeErr
	}

	return nil
}
