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
			{
				Config: testAccSubscriptionConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccSubscriptionPresence,
				),
			},
		},
	})
}

func testAccSubscriptionConfig(rName string, withAWSResources bool) string {
	var config string

	queueURL := "https://sqs.eu-west-1.amazonaws.com/0000000000/some-queue"
	accessKey := "some-access-key"
	secretKey := "some-secret-key"

	if withAWSResources {
		config += testAccSQSConfig(rName)
		queueURL = "${aws_sqs_queue.ct_queue.id}"
		accessKey = "${aws_iam_access_key.ct.id}"
		secretKey = "${aws_iam_access_key.ct.secret}"
	}

	config += fmt.Sprintf(`
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
	return config
}

func testAccSubscriptionPresence(s *terraform.State) error {
	// TODO: Implement
	return nil
}

func testAccCheckSubscriptionDestroy(s *terraform.State) error {
	// TODO: Implement
	return nil
}
