package checkout_payment_integration_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/commercetools-go-sdk/checkout"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
)

func checkoutClient(t *testing.T) *checkout.ByProjectKeyRequestBuilder {
	client, err := acctest.GetCheckoutClient()
	if err != nil {
		t.Fatal(err)
	}
	return client
}

// integration reads the payment integration with the key from the Checkout API.
func integration(t *testing.T, key string) *checkout.PaymentIntegration {
	p, err := checkoutClient(t).PaymentIntegrations().WithKey(key).Get().Execute(context.Background())
	if err != nil {
		return nil
	}
	return p
}

const application = `
resource "commercetools_checkout_application" "shop" {
  key                           = "acc-checkout-pi-shop"
  name                          = "Shop"
  mode                          = "PaymentOnly"
  countries                     = ["NL"]
  allowed_origins               = ["https://shop.example.com"]
  active_payment_component_type = "Component"
}
`

func config(name, deployment, extra string) string {
	return application + fmt.Sprintf(`
resource "commercetools_checkout_payment_integration" "card" {
  key                   = "acc-checkout-adyen-card"
  application_id        = commercetools_checkout_application.shop.id
  connect_deployment_id = %q
  name                  = %q
  type                  = "card"
  component_type        = "Component"
  status                = "Active"
  predicate             = "country='NL'"
  sort_priority         = 1

  label = {
    en = "Credit card"
    nl = "Creditcard"
  }

  %s
}
`, deployment, name, extra)
}

func TestAccCheckoutPaymentIntegration(t *testing.T) {
	name := "commercetools_checkout_payment_integration.card"
	key := "acc-checkout-adyen-card"
	var firstID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: func(_ *terraform.State) error {
			_, err := checkoutClient(t).Applications().WithKey("acc-checkout-pi-shop").Get().Execute(context.Background())
			if integration(t, key) != nil || err == nil {
				return errors.New("the payment integration or its application still exists after destroy")
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: config("Adyen card", "deployment-adyen", `
  automated_reversal_predicate = "true"
`),
				ExpectError: regexp.MustCompile(`automated_reversal`),
			},
			{
				Config: config("Adyen card", "deployment-adyen", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(name, "application_id", "commercetools_checkout_application.shop", "id"),
					resource.TestCheckResourceAttr(name, "status", "Active"),
					resource.TestCheckResourceAttr(name, "label.nl", "Creditcard"),
					resource.TestCheckNoResourceAttr(name, "automated_reversal"),
					func(s *terraform.State) error {
						firstID = s.RootModule().Resources[name].Primary.ID
						p := integration(t, key)
						assert.Equal(t, "deployment", p.ConnectorDeployment.TypeId)
						assert.Equal(t, "deployment-adyen", p.ConnectorDeployment.ID)
						assert.Equal(t, 1, p.SortingInfo.Priority)
						return nil
					},
				),
			},
			{
				// Display info and reversal change in place.
				Config: config("Adyen cards", "deployment-adyen", `
  pay_button_text = {
    en = "Pay now"
  }
  automated_reversal           = "Active"
  automated_reversal_predicate = "country='NL'"
`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "Adyen cards"),
					resource.TestCheckResourceAttr(name, "pay_button_text.en", "Pay now"),
					resource.TestCheckResourceAttr(name, "automated_reversal", "Active"),
					func(s *terraform.State) error {
						assert.Equal(t, firstID, s.RootModule().Resources[name].Primary.ID)
						p := integration(t, key)
						assert.Equal(t, "Pay now", (*p.DisplayInfo.PayButtonText)["en"])
						assert.Equal(t, "Creditcard", (*p.DisplayInfo.Label)["nl"])
						assert.Equal(t, "country='NL'", *p.AutomatedReversalConfiguration.Predicate)
						return nil
					},
				),
			},
			{
				// Another connector deployment replaces it.
				Config: config("Adyen cards", "deployment-adyen-2", ""),
				Check: func(s *terraform.State) error {
					assert.NotEqual(t, firstID, s.RootModule().Resources[name].Primary.ID)
					p := integration(t, key)
					assert.Equal(t, checkout.PaymentComponentTypeComponent, p.ComponentType)
					assert.Nil(t, p.AutomatedReversalConfiguration)
					return nil
				},
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
