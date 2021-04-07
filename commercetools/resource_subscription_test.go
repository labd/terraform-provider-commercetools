package commercetools

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/labd/commercetools-go-sdk/commercetools"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestValidateDestination(t *testing.T) {
	validDestinations := []map[string]interface{}{
		{
			"type":          "SQS",
			"queue_url":     "<queue_url>",
			"access_key":    "<access_key>",
			"access_secret": "<access_secret>",
			"region":        "<region>",
		},
		{
			"type":       "azure_eventgrid",
			"uri":        "<uri>",
			"access_key": "<access_key>",
		},
		{
			"type":              "azure_servicebus",
			"connection_string": "<connection_string>",
		},
		{
			"type":       "google_pubsub",
			"project_id": "<project_id>",
			"topic":      "<topic>",
		},
	}
	for _, validDestination := range validDestinations {
		_, errs := validateDestination(validDestination, "destination")
		if len(errs) > 0 {
			t.Error("Expected no validation errors, but got ", errs)
		}
	}
	invalidDestinations := []map[string]interface{}{
		{
			"type": "SQS1",
		},
		{
			"type":          "SQS",
			"access_key":    "<access_key>",
			"access_secret": "<access_secret>",
			"region":        "<region>",
		},
		{
			"type": "azure_servicebus",
		},
		{
			"type":  "google_pubsub",
			"topic": "<topic>",
		},
	}
	for _, validDestination := range invalidDestinations {
		_, errs := validateDestination(validDestination, "destination")
		if len(errs) == 0 {
			t.Error("Expected validation errors, but none was reported")
		}
	}
}

func TestAccSubscription_basic(t *testing.T) {
	rName := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccSubscriptionConfig(rName),
				ExpectError: regexp.MustCompile(".*A test message could not be delivered to this destination: SQS.*"),
			},
		},
	})
}

func testAccSubscriptionConfig(rName string) string {
	queueURL := "https://sqs.eu-west-1.amazonaws.com/0000000000/some-queue"
	accessKey := "some-access-key"
	secretKey := "some-secret-key"

	return fmt.Sprintf(`
resource "commercetools_subscription" "subscription_%[1]s" {
	key = "commercetools-acc-%[1]s"

	destination = {
		type          = "SQS"
		queue_url     = "%[2]s"
		access_key    = "%[3]s"
		access_secret = "%[4]s"
		region        = "eu-west-1"
	}

	changes {
		resource_type_ids = ["customer"]
	}

	message {
		resource_type_id = "product"

		types = ["ProductPublished", "ProductCreated"]
	}
}
`, rName, queueURL, accessKey, secretKey)
}

func testAccCheckSubscriptionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*commercetools.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_subscription" {
			continue
		}
		response, err := conn.SubscriptionGetWithID(context.Background(), rs.Primary.ID)
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("subscription (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		// If we don't get a was not found error, return the actual error. Otherwise resource is destroyed
		if !strings.Contains(err.Error(), "was not found") && !strings.Contains(err.Error(), "Not Found (404)") {
			return err
		}
	}
	return nil
}
