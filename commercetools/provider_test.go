package commercetools

import (
	"fmt"
	"os"
	"testing"

	// "github.com/terraform-providers/terraform-provider-aws/aws"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProviderFactories func(providers *[]*schema.Provider) map[string]terraform.ResourceProviderFactory
var testAccProvider *schema.Provider
var testAccAWSProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	// TODO: The following raises an type assertion error
	// Fix the AWS provider in order for the acceptance tests to work
	// testAccAWSProvider = aws.Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		// "aws":           testAccAWSProvider,
		"commercetools": testAccProvider,
	}
	testAccProviderFactories = func(providers *[]*schema.Provider) map[string]terraform.ResourceProviderFactory {
		return map[string]terraform.ResourceProviderFactory{
			"commercetools": func() (terraform.ResourceProvider, error) {
				p := Provider()
				*providers = append(*providers, p.(*schema.Provider))
				return p, nil
			},
		}
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	requiredEnvs := []string{
		"COMMERCETOOLS_CLIENT_ID",
		"COMMERCETOOLS_CLIENT_SECRET",
		"COMMERCETOOLS_PROJECT_KEY",
	}
	for _, val := range requiredEnvs {
		if os.Getenv(val) == "" {
			t.Fatalf("%v must be set for acceptance tests", val)
		}
	}

	err := testAccProvider.Configure(terraform.NewResourceConfig(nil))
	if err != nil {
		t.Fatal(err)
	}
}

// Creates a SQS configuration
func testAccSQSConfig(suffix string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "ct" {
	name = "commercetools-acctest-%[1]s"
}
	
resource "aws_iam_user_policy_attachment" "sqs" {
	user       = "${aws_iam_user.ct.name}"
	policy_arn = "arn:aws:iam::aws:policy/AmazonSQSFullAccess"
}
	
resource "aws_iam_access_key" "ct" {
	user = "${aws_iam_user.ct.name}"
}
	
resource "aws_sqs_queue" "ct_queue" {
	name                      = "commercetools-acctest-queue-%[1]s"
	delay_seconds             = 90
	max_message_size          = 2048
	message_retention_seconds = 86400
	receive_wait_time_seconds = 10
}
`, suffix)
}
