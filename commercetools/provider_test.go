package commercetools

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProviderFactories func(providers *[]*schema.Provider) map[string]terraform.ResourceProviderFactory
var testAccProvider terraform.ResourceProvider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]terraform.ResourceProvider{
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
