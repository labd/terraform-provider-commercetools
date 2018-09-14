package commercetools

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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
				ExpectError: regexp.MustCompile("A test message could not be delivered to this destination: SQS.*"),
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
	
	destination {
		type          = "SQS"
		queue_url     = "%[2]s"
		access_key    = "%[3]s"
		access_secret = "%[4]s"
		region        = "eu-west-1"
	}
	
	changes {
		resource_type_id = ["customer"]
	}
	
	message {
		resource_type_id = "product"
	
		types = ["ProductPublished", "ProductCreated"]
	}
}
`, rName, queueURL, accessKey, secretKey)
}

func testAccCheckSubscriptionDestroy(s *terraform.State) error {
	// TODO: Implement
	return nil
}
