package subscription_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func TestAccSubscription_basic(t *testing.T) {
	rName := "foobar"
	key := fmt.Sprintf("commercetools-acc-%s", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccSubscriptionConfig("subscription", key),
				ExpectError: regexp.MustCompile(".*A test message could not be delivered to this destination: SQS.*"),
			},
		},
	})
}

func testAccSubscriptionConfig(identifier, key string) string {
	queueURL := "https://sqs.eu-west-1.amazonaws.com/0000000000/some-queue"
	accessKey := "some-access-key"
	secretKey := "some-secret-key"

	return utils.HCLTemplate(`
		resource "commercetools_subscription" "{{ .identifier }}" {
			key = "commercetools-acc-{{ .key }}"

			destination {
				type          = "SQS"
				queue_url     = "{{ .queueURL }}"
				access_key    = "{{ .accessKey }}"
				access_secret = "{{ .secretKey }}"
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
		`,
		map[string]any{
			"identifier": identifier,
			"key":        key,
			"queueURL":   queueURL,
			"accessKey":  accessKey,
			"secretKey":  secretKey,
		})
}

func testAccCheckSubscriptionDestroy(s *terraform.State) error {
	conn, err := acctest.GetClient()
	if err != nil {
		return err
	}
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
		if newErr := acctest.CheckApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}
