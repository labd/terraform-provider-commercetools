package integration_layer_idp_login_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/commercetools-go-sdk/integrationlayer"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
)

// projectKey is the Project of this test alone.
const projectKey = "acc-cil-idp-login"

func TestAccIntegrationLayerIdpLogin(t *testing.T) {
	name := "commercetools_integration_layer_idp_login.login"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: func(_ *terraform.State) error {
			if idp(t) != nil {
				return errors.New("the IdP login is still configured after destroy")
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: acctest.ProviderConfig(projectKey) + `
resource "commercetools_integration_layer_idp_login" "login" {
  issuer              = "https://login.example.com"
  token_endpoint      = "https://user:pass@login.example.com/token"
  revocation_endpoint = "https://login.example.com/revoke"
  jwks_uri            = "https://login.example.com/jwks"
  client_id           = "storefront"
  redirect_uri        = "https://shop.example.com/callback"
  match_strategy      = "email"
}
`,
				ExpectError: regexp.MustCompile(`absolute http\(s\) URL without credentials`),
			},
			{
				// The API needs a client secret for the first login.
				Config:      config("storefront", "", ""),
				ExpectError: regexp.MustCompile(`clientSecret is required`),
			},
			{
				Config: config("storefront", `client_secret = "s3cret"`, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "id", projectKey),
					resource.TestCheckResourceAttr(name, "has_client_secret", "true"),
					resource.TestCheckResourceAttr(name, "external_id_claim", "sub"),
					resource.TestCheckResourceAttr(name, "phone_claim", "phone_number"),
					resource.TestCheckNoResourceAttr(name, "authorization_endpoint"),
					func(_ *terraform.State) error {
						assert.Equal(t, integrationlayer.IdpClaimMapping{
							ExternalId: "sub", Email: "email", FirstName: "given_name",
							LastName: "family_name", Phone: "phone_number",
						}, idp(t).Claims)
						return nil
					},
				),
			},
			{
				// Without client_secret the stored one is kept, which the API
				// would otherwise refuse.
				Config: config("storefront-2", "", `
  external_id_claim      = "oid"
  authorization_endpoint = "https://login.example.com/authorize"
`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "client_id", "storefront-2"),
					resource.TestCheckResourceAttr(name, "external_id_claim", "oid"),
					resource.TestCheckResourceAttr(name, "email_claim", "email"),
					func(_ *terraform.State) error {
						login := idp(t)
						assert.True(t, login.HasClientSecret)
						assert.Equal(t, "https://login.example.com/authorize", *login.AuthorizationEndpoint)
						return nil
					},
				),
			},
			{
				ResourceName:            name,
				ImportState:             true,
				ImportStateId:           projectKey,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_secret"},
			},
		},
	})
}

func config(clientID, secret, extra string) string {
	return acctest.ProviderConfig(projectKey) + fmt.Sprintf(`
resource "commercetools_integration_layer_idp_login" "login" {
  issuer              = "https://login.example.com"
  token_endpoint      = "https://login.example.com/token"
  revocation_endpoint = "https://login.example.com/revoke"
  jwks_uri            = "https://login.example.com/jwks"
  client_id           = %q
  redirect_uri        = "https://shop.example.com/callback"
  match_strategy      = "externalId"
  %s
  %s
}
`, clientID, secret, extra)
}

// idp reads the IdP login, or nil when there is none.
func idp(t *testing.T) *integrationlayer.IdpConfig {
	client, err := acctest.GetIntegrationLayerClient(projectKey)
	if err != nil {
		t.Fatal(err)
	}
	login, err := client.Main().Idp().Get().Execute(context.Background())
	if err != nil {
		return nil
	}
	return login
}
