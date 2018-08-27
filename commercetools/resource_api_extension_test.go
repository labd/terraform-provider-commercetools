package commercetools

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/labd/commercetools-go-sdk/commercetools"
	"github.com/labd/commercetools-go-sdk/service/extensions"
	"github.com/stretchr/testify/assert"
)

func TestAPIExtensionGetAuthentication(t *testing.T) {
	var input map[string]interface{}
	input = map[string]interface{}{
		"authorization_header": "12345",
		"azure_authentication": "AzureKey",
	}

	auth, err := resourceAPIExtensionGetAuthentication(input)
	assert.Nil(t, auth)
	assert.NotNil(t, err)

	input = map[string]interface{}{
		"authorization_header": "12345",
	}

	auth, err = resourceAPIExtensionGetAuthentication(input)
	httpAuth, ok := auth.(*extensions.DestinationAuthenticationAuth)
	assert.True(t, ok)
	assert.Equal(t, "12345", httpAuth.HeaderValue)
	assert.NotNil(t, auth)
	assert.Nil(t, err)
}

func TestAccAPIExtension_basic(t *testing.T) {
	name := fmt.Sprintf("extension_%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAPIExtensionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIExtsensionConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccAPIExtensionExists(name),
				),
			},
		},
	})
}

func testAccAPIExtsensionConfig(name string) string {
	return fmt.Sprintf(`
resource "commercetools_api_extension" "%s" {
	key = "terraform-acctest-extension"

  destination {
    type                 = "HTTP"
    url                  = "https://example.com"
    authorization_header = "Basic 12345"
  }

  trigger {
    resource_type_id = "customer"
    actions = ["Create", "Update"]
  }
}
`, name)
}

func testAccAPIExtensionExists(n string) resource.TestCheckFunc {
	name := fmt.Sprintf("commercetools_api_extension.%s", n)
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Extension ID is set")
		}

		svc := getExtensionService(testAccProvider.Meta().(*commercetools.Client))
		result, err := svc.GetByID(rs.Primary.ID)
		if err != nil {
			return err
		}
		if result == nil {
			return fmt.Errorf("Extension not found")
		}

		return nil
	}
}

func testAccCheckAPIExtensionDestroy(s *terraform.State) error {
	// TODO: Implement
	return nil
}
