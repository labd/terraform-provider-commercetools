package commercetools

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/labd/commercetools-go-sdk/commercetools"
	"github.com/stretchr/testify/assert"
)

func TestAPIExtensionGetDestination(t *testing.T) {
	resourceDataMap := map[string]interface{}{
		"id":             "2845b936-e407-4f29-957b-f8deb0fcba97",
		"version":        1,
		"createdAt":      "2018-12-03T16:13:03.969Z",
		"lastModifiedAt": "2018-12-04T09:06:59.491Z",
		"destination": map[string]interface{}{
			"type":          "AWSLambda",
			"arn":           "arn:aws:lambda:eu-west-1:111111111:function:api_extensions",
			"access_key":    "ABCSDF123123123",
			"access_secret": "****abc/",
		},
		"timeout_in_ms": 1,
		"key":           "create-order",
	}

	d := schema.TestResourceDataRaw(t, resourceAPIExtension().Schema, resourceDataMap)
	destination, _ := resourceAPIExtensionGetDestination(d)
	lambdaDestination, ok := destination.(commercetools.ExtensionAWSLambdaDestination)

	assert.True(t, ok)
	assert.Equal(t, lambdaDestination.Arn, "arn:aws:lambda:eu-west-1:111111111:function:api_extensions")
	assert.Equal(t, lambdaDestination.AccessKey, "ABCSDF123123123")
	assert.Equal(t, lambdaDestination.AccessSecret, "****abc/")
}

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
	httpAuth, ok := auth.(*commercetools.ExtensionAuthorizationHeaderAuthentication)
	assert.True(t, ok)
	assert.Equal(t, "12345", httpAuth.HeaderValue)
	assert.NotNil(t, auth)
	assert.Nil(t, err)
}

func TestAccAPIExtension_basic(t *testing.T) {
	name := fmt.Sprintf("extension_%s", acctest.RandString(5))
	timeoutInMs := acctest.RandIntRange(200, 1800)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAPIExtensionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIExtensionConfig(name, timeoutInMs),
				Check: resource.ComposeTestCheckFunc(
					testAccAPIExtensionExists("ext"),
					resource.TestCheckResourceAttr(
						"commercetools_api_extension.ext", "key", name),
					resource.TestCheckResourceAttr(
						"commercetools_api_extension.ext", "timeout_in_ms", strconv.FormatInt(int64(timeoutInMs), 10)),
				),
			},
		},
	})
}

func testAccAPIExtensionConfig(name string, timeoutInMs int) string {
	return fmt.Sprintf(`
resource "commercetools_api_extension" "ext" {
  key = "%s"
  timeout_in_ms = %d

  destination = {
    type                 = "HTTP"
    url                  = "https://example.com"
    authorization_header = "Basic 12345"
  }

  trigger {
    resource_type_id = ["customer", "product"]
    actions = ["Create", "Update"]
  }
}
`, name, timeoutInMs)
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
		client := getClient(testAccProvider.Meta())
		result, err := client.ExtensionGetWithID(context.Background(), rs.Primary.ID)
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
