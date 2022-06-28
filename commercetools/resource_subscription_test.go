package commercetools

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestValidateSubscriptionDestination(t *testing.T) {
	resource := resourceSubscription()
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
			"type":       "EventGrid",
			"uri":        "<uri>",
			"access_key": "<access_key>",
		},
		{
			"type":              "azure_servicebus",
			"connection_string": "<connection_string>",
		},
		{
			"type":              "AzureServiceBus",
			"connection_string": "<connection_string>",
		},
		{
			"type":       "google_pubsub",
			"project_id": "<project_id>",
			"topic":      "<topic>",
		},
		{
			"type":       "GoogleCloudPubSub",
			"project_id": "<project_id>",
			"topic":      "<topic>",
		},
		{
			"type":       "event_bridge",
			"region":     "<region>",
			"account_id": "<account_id>",
		},
		{
			"type":       "EventBridge",
			"region":     "<region>",
			"account_id": "<account_id>",
		},
	}
	for _, validDestination := range validDestinations {
		rawData := map[string]interface{}{
			"destination": []interface{}{validDestination},
		}
		data := schema.TestResourceDataRaw(t, resource.Schema, rawData)
		err := validateSubscriptionDestination(data)
		if err != nil {
			t.Error("Expected no validation errors, but got ", err)
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
		{
			"type":  "event_bridge",
			"topic": "<region>",
		},
	}
	for _, validDestination := range invalidDestinations {
		rawData := map[string]interface{}{
			"destination": []interface{}{validDestination},
		}
		data := schema.TestResourceDataRaw(t, resource.Schema, rawData)
		err := validateSubscriptionDestination(data)
		if err == nil {
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

	destination {
		type          = "SQS"
		queue_url     = "%[2]s"
		access_key    = "%[3]s"
		access_secret = "%[4]s"
		region        = "eu-west-1"
	}

	format {
		type = "Platform"
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
	conn := getClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_subscription" {
			continue
		}
		response, err := conn.Subscriptions().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("subscription (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		if newErr := checkApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}
