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

var (
	accessKey = "some-access-key"
	secretKey = "some-secret-key"
	region    = "eu-west-1"
)

func TestAccSubscription_basic(t *testing.T) {
	name := "foobar"
	key := "my-key"
	resourceName := fmt.Sprintf("commercetools_subscription.%s", name)
	queueUrl := "https://sqs.eu-west-1.amazonaws.com/0000000001/some-queue"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionConfigBasic(name, key, queueUrl),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					resource.TestCheckResourceAttr(resourceName, "destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination.0.type", "SQS"),
					resource.TestCheckResourceAttr(resourceName, "destination.0.queue_url", queueUrl),
					resource.TestCheckResourceAttr(resourceName, "destination.0.access_key", accessKey),
					resource.TestCheckResourceAttr(resourceName, "destination.0.access_secret", secretKey),
					resource.TestCheckResourceAttr(resourceName, "destination.0.region", region),
					resource.TestCheckResourceAttr(resourceName, "changes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "changes.0.resource_type_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "format.0.type", "Platform"),
				),
			},
			{
				ResourceName: resourceName,
				Config:       testAccSubscriptionConfigSQSFailure(name, key),
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					conn, err := acctest.GetClient()
					if err != nil {
						return "", err
					}

					response, err := conn.Subscriptions().WithKey(key).Get().Execute(context.Background())
					if err != nil {
						return "", err
					}
					return response.ID, nil
				},
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"destination.0.access_secret",
				},
			},
		},
	})
}

func testAccSubscriptionConfigBasic(identifier, key, queueURL string) string {
	return utils.HCLTemplate(`
		resource "commercetools_subscription" "{{ .identifier }}" {
			key = "{{ .key }}"

			destination {
				type          = "SQS"
				queue_url     = "{{ .queueURL }}"
				access_key    = "{{ .accessKey }}"
				access_secret = "{{ .secretKey }}"
				region        = "{{ .region }}"
			}

			format {
				type = "Platform"
			}

			changes {
				resource_type_ids = ["customer", "product"]
			}
		}
		`,
		map[string]any{
			"identifier": identifier,
			"key":        key,
			"queueURL":   queueURL,
			"accessKey":  accessKey,
			"secretKey":  secretKey,
			"region":     region,
		})
}

func TestAccSubscription_sqs_failure(t *testing.T) {
	rName := "foobar"
	key := fmt.Sprintf("commercetools-acc-%s", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccSubscriptionConfigSQSFailure("subscription", key),
				ExpectError: regexp.MustCompile(".*A test message could not be delivered to this destination: SQS.*"),
			},
		},
	})
}

func testAccSubscriptionConfigSQSFailure(identifier, key string) string {
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
