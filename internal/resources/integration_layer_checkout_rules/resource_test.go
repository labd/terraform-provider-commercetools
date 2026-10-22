package integration_layer_checkout_rules_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/labd/commercetools-go-sdk/integrationlayer"
	"github.com/labd/terraform-provider-commercetools/internal/acctest"
)

func TestAccIntegrationLayerCheckoutRules(t *testing.T) {
	name := "commercetools_integration_layer_checkout_rules.rules"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: acctest.ProviderConfig(projectKey) + `
resource "commercetools_integration_layer_checkout_rules" "rules" {
  rule {
    countries       = ["de"]
    application_key = "eu-checkout"
    mode            = "COMPLETE"
  }
}
`,
				ExpectError: regexp.MustCompile(`uppercase ISO 3166-1 alpha-2 code`),
			},
			{
				Config: acctest.ProviderConfig(projectKey) + `
resource "commercetools_integration_layer_checkout_rules" "rules" {
  rule {
    application_key = "catch-all"
    mode            = "PAYMENT_ONLY"
  }
  rule {
    countries       = ["DE"]
    application_key = "de-checkout"
    mode            = "COMPLETE"
  }
}
`,
				ExpectError: regexp.MustCompile(`Rule 1 has neither countries nor stores`),
			},
			{
				Config: acctest.ProviderConfig(projectKey) + `
resource "commercetools_integration_layer_checkout_rules" "rules" {
  rule {
    countries       = ["DE", "AT"]
    application_key = "eu-checkout"
    mode            = "COMPLETE"
  }
  rule {
    stores          = ["vip-store"]
    application_key = "vip-checkout"
    mode            = "PAYMENT_ONLY"
  }
  rule {
    application_key = "default-checkout"
    mode            = "PAYMENT_ONLY"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "id", projectKey),
					resource.TestCheckResourceAttr(name, "rule.#", "3"),
					resource.TestCheckNoResourceAttr(name, "rule.2.countries"),
					func(_ *terraform.State) error {
						assert.Equal(t, &integrationlayer.CheckoutConfig{Rules: []integrationlayer.CheckoutRule{
							{Countries: []string{"DE", "AT"}, ApplicationKey: "eu-checkout", Mode: "COMPLETE"},
							{Stores: []string{"vip-store"}, ApplicationKey: "vip-checkout", Mode: "PAYMENT_ONLY"},
							{ApplicationKey: "default-checkout", Mode: "PAYMENT_ONLY"},
						}}, checkout(t))
						return nil
					},
				),
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateId:     projectKey,
				ImportStateVerify: true,
			},
		},
	})
}

// projectKey is the Project of this test alone.
const projectKey = "acc-cil-checkout-rules"

func client(t *testing.T) *integrationlayer.ByProjectKeyRequestBuilder {
	c, err := acctest.GetIntegrationLayerClient(projectKey)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func checkout(t *testing.T) *integrationlayer.CheckoutConfig {
	view, err := client(t).Main().Settings().Get().Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return view.Project.Checkout
}
