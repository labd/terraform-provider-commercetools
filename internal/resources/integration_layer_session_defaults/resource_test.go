package integration_layer_session_defaults_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/commercetools-go-sdk/integrationlayer"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
)

// projectKey is the Project of this test alone, whose languages, currencies
// and countries it sets.
const projectKey = "acc-cil-session-defaults"

func TestAccIntegrationLayerSessionDefaults(t *testing.T) {
	name := "commercetools_integration_layer_session_defaults.defaults"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config("Shop", "nl-NL", "EUR", "NL"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "id", projectKey),
					resource.TestCheckResourceAttr(name, "region_id", "europe-west1"),
					func(_ *terraform.State) error {
						project := settings(t).Project
						assert.Equal(t, "Shop", *project.Label)
						assert.Equal(t, "nl-NL", *project.Language)
						assert.Equal(t, "EUR", *project.Currency)
						assert.Equal(t, "NL", *project.Country)
						return nil
					},
				),
			},
			{
				Config: config("Shop", "de-DE", "EUR", "DE"),
				Check: func(_ *terraform.State) error {
					assert.Equal(t, "DE", *settings(t).Project.Country)
					return nil
				},
			},
			{
				Config:      config("Shop", "de-DE", "EUR", "de"),
				ExpectError: regexp.MustCompile(`The Project spells "de" as "DE"`),
			},
			{
				Config:      config("Shop", "de-DE", "GBP", "DE"),
				ExpectError: regexp.MustCompile(`"GBP" is not one of the Project's values: EUR, USD`),
			},
			{
				Config:            config("Shop", "de-DE", "EUR", "DE"),
				ResourceName:      name,
				ImportState:       true,
				ImportStateId:     projectKey,
				ImportStateVerify: true,
			},
		},
	})
}

// config sets the Project's languages, currencies and countries first, which
// the session defaults are checked against.
func config(label, language, currency, country string) string {
	return acctest.ProviderConfig(projectKey) + fmt.Sprintf(`
resource "commercetools_project_settings" "project" {
  name       = "Session defaults"
  countries  = ["NL", "DE"]
  currencies = ["EUR", "USD"]
  languages  = ["nl-NL", "de-DE", "en-US"]
}

resource "commercetools_integration_layer_session_defaults" "defaults" {
  label    = %q
  language = %q
  currency = %q
  country  = %q

  depends_on = [commercetools_project_settings.project]
}
`, label, language, currency, country)
}

func settings(t *testing.T) *settingsView {
	client, err := acctest.GetIntegrationLayerClient(projectKey)
	if err != nil {
		t.Fatal(err)
	}
	view, err := client.Main().Settings().Get().Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return view
}

func TestAccIntegrationLayerURLUnknown(t *testing.T) {
	// api_url is the mock's, from which no region can be derived, and no
	// integration_layer_url is given.
	providerConfig := fmt.Sprintf(`
provider "commercetools" {
  client_id     = %q
  client_secret = %q
  project_key   = %q
  scopes        = "manage_project:%s"
  api_url       = %q
  token_url     = %q
}
`, os.Getenv("CTP_CLIENT_ID"), os.Getenv("CTP_CLIENT_SECRET"), projectKey, projectKey,
		os.Getenv("CTP_API_URL"), os.Getenv("CTP_AUTH_URL"))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.TestAccPreCheck(t)
			if os.Getenv("CTP_INTEGRATION_LAYER_URL") != "" {
				t.Skip("CTP_INTEGRATION_LAYER_URL gives the provider the URL")
			}
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{{
			Config: providerConfig + `
resource "commercetools_integration_layer_session_defaults" "defaults" {
  label    = "Shop"
  language = "nl-NL"
  currency = "EUR"
  country  = "NL"
}
`,
			ExpectError: regexp.MustCompile(`Commerce Integration Layer URL unknown`),
		}},
	})
}

type settingsView = integrationlayer.ProjectSettingsView
