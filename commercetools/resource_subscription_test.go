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
				Config:      testAccSubscriptionConfig(rName, false),
				ExpectError: regexp.MustCompile("A test message could not be delivered to this destination: SQS.*"),
			},
			// {
			// 	Config: testAccSubscriptionConfig(rName, true),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		testAccSubscriptionPresence,
			// 	),
			// },
		},
	})
}

func testAccSubscriptionConfig(rName string, withAWSResources bool) string {
	// TODO: Create AWS resources to test with when withAWSResources is true

	return fmt.Sprintf(`
resource "commercetools_subscription" "subscription_%[1]s" {
	key = "commercetools-acc-%[1]s"
	
	destination {
		type          = "SQS"
		queue_url     = "https://sqs.eu-west-1.amazonaws.com/0000000000/terraform-queue-%[1]s"
		access_key    = "accesskey_%[1]s"
		access_secret = "secret_%[1]s"
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
`, rName)
}

// func testAccSubscriptionPresence(s *terraform.State) error {
// 	return nil
// }

func testAccCheckSubscriptionDestroy(s *terraform.State) error {
	return nil
}
