package commercetools

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = New("snapshot")()
	testAccProviders = map[string]*schema.Provider{
		"commercetools": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	provider := New("snapshot")()
	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	requiredEnvs := []string{
		"CTP_CLIENT_ID",
		"CTP_CLIENT_SECRET",
		"CTP_PROJECT_KEY",
		"CTP_SCOPES",
		"CTP_API_URL",
		"CTP_AUTH_URL",
	}
	for _, val := range requiredEnvs {
		if os.Getenv(val) == "" {
			t.Fatalf("%v must be set for acceptance tests", val)
		}
	}

	cfg := map[string]any{
		"client_id":     "dummy-client-id",
		"client_secret": "dummy-client-secret",
		"project_key":   "terraform-provider-commercetools",
	}

	err := testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(cfg))
	if err != nil {
		t.Fatal(err)
	}
}
