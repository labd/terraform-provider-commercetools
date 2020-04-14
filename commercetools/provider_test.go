package commercetools

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/labd/commercetools-go-sdk/commercetools"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProviderFactories func(providers *[]*schema.Provider) map[string]terraform.ResourceProviderFactory
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
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

func TestProviderConfig(t *testing.T) {
	resourceDataMap := map[string]interface{}{
		"client_id":     "test-client-id",
		"client_secret": "test-client-secret",
		"project_key":   "test-project-key",
		"scopes":        "view_project_settings:test-project-key",
		"api_url":       "https://api.europe-west1.gcp.commercetools.com",
		"token_url":     "https://auth.europe-west1.gcp.commercetools.com",
		"mc_api_url":    "https://mc-api.europe-west1.gcp.commercetools.com",
	}
	err := testAccProvider.Configure(terraform.NewResourceConfigRaw(resourceDataMap))
	if err != nil {
		t.Fatal(err)
	}

	meta := testAccProvider.Meta()
	assert.IsType(t, getClient(meta), (*commercetools.Client)(nil))
	assert.IsType(t, getGraphQLClient(meta), (*graphql.Client)(nil))
	assert.Equal(t, getProjectKey(meta), "test-project-key")
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	requiredEnvs := []string{
		"CTP_CLIENT_ID",
		"CTP_CLIENT_SECRET",
		"CTP_PROJECT_KEY",
		"CTP_SCOPES",
		"CTP_API_URL",
		"CTP_AUTH_URL",
		"CTP_MC_API_URL",
	}
	for _, val := range requiredEnvs {
		if os.Getenv(val) == "" {
			t.Fatalf("%v must be set for acceptance tests", val)
		}
	}

	err := testAccProvider.Configure(terraform.NewResourceConfigRaw(nil))
	if err != nil {
		t.Fatal(err)
	}
}
