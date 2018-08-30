package commercetools

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

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
	// TODO: Implement
	return nil
}
