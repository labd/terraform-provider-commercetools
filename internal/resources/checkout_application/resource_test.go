package checkout_application_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/commercetools-go-sdk/checkout"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
)

// application reads the application with the key from the Checkout API.
func application(t *testing.T, key string) *checkout.Application {
	client, err := acctest.GetCheckoutClient()
	if err != nil {
		t.Fatal(err)
	}
	a, err := client.Applications().WithKey(key).Get().Execute(context.Background())
	if err != nil {
		return nil
	}
	return a
}

func sorted(values []string) []string {
	out := slices.Clone(values)
	slices.Sort(out)
	return out
}

func config(name, origins, extra string) string {
	return fmt.Sprintf(`
resource "commercetools_checkout_application" "shop" {
  key                           = "acc-checkout-shop"
  name                          = %q
  mode                          = "CompleteFlow"
  status                        = "Active"
  countries                     = ["NL", "BE"]
  allowed_origins               = [%s]
  logo_url                      = "https://shop.example.com/logo.svg"
  payment_return_url            = "https://shop.example.com/checkout/return"
  active_payment_component_type = "Component"
  allow_discounts               = true

  description = {
    en = "Checkout for the shop"
  }

  %s
}
`, name, origins, extra)
}

const terms = `
  agreement {
    name = "terms"
    type = "MandatoryCheckbox"
    text = {
      en = "I agree to the terms."
    }
  }
`

func TestAccCheckoutApplication(t *testing.T) {
	name := "commercetools_checkout_application.shop"
	key := "acc-checkout-shop"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: func(_ *terraform.State) error {
			if application(t, key) != nil {
				return errors.New("the application still exists after destroy")
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: `
resource "commercetools_checkout_application" "shop" {
  key                           = "acc-checkout-shop"
  name                          = "Shop"
  mode                          = "CompleteFlow"
  countries                     = ["NL"]
  active_payment_component_type = "Component"
}
`,
				ExpectError: regexp.MustCompile(`A CompleteFlow application needs a logo_url`),
			},
			{
				Config: `
resource "commercetools_checkout_application" "shop" {
  key                           = "acc-checkout-shop"
  name                          = "Shop"
  mode                          = "PaymentOnly"
  countries                     = ["NL"]
  active_payment_component_type = "Component"
  allow_discounts               = true
}
`,
				ExpectError: regexp.MustCompile(`Only a CompleteFlow application takes allow_discounts`),
			},
			{
				Config: config("Shop", `"https://shop.example.com"`, terms),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(name, "id"),
					resource.TestCheckResourceAttr(name, "status", "Active"),
					resource.TestCheckResourceAttr(name, "allow_all_origins", "false"),
					resource.TestCheckResourceAttr(name, "description.en", "Checkout for the shop"),
					resource.TestCheckResourceAttr(name, "agreement.0.status", "Active"),
					func(_ *terraform.State) error {
						a := application(t, key)
						assert.Equal(t, []string{"BE", "NL"}, sorted(a.Countries))
						assert.Equal(t, []string{"https://shop.example.com"}, a.AllowedOrigins.Origins)
						assert.Equal(t, checkout.PaymentComponentTypeComponent, a.PaymentsConfiguration.ActivePaymentComponentType)
						assert.Len(t, a.Agreements, 1)
						assert.Equal(t, "I agree to the terms.", a.Agreements[0].Text["en"])
						return nil
					},
				),
			},
			{
				// The agreements are set again only when they change.
				Config: config("Shop NL",
					`"https://shop.example.com", "https://www.shop.example.com"`, terms+`
  agreement {
    name   = "newsletter"
    type   = "Text"
    status = "Inactive"
    text = {
      en = "We send a newsletter."
    }
  }
`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "Shop NL"),
					resource.TestCheckResourceAttr(name, "allowed_origins.#", "2"),
					resource.TestCheckResourceAttr(name, "agreement.#", "2"),
					resource.TestCheckResourceAttr(name, "agreement.1.name", "newsletter"),
					func(_ *terraform.State) error {
						a := application(t, key)
						assert.Equal(t, checkout.ApplicationAgreementStatusInactive, a.Agreements[1].Status)
						return nil
					},
				),
			},
			{
				// Removing the description and the agreements clears them.
				Config: `
resource "commercetools_checkout_application" "shop" {
  key                           = "acc-checkout-shop"
  name                          = "Shop NL"
  mode                          = "CompleteFlow"
  status                        = "Active"
  countries                     = ["NL", "BE"]
  allow_all_origins             = true
  logo_url                      = "https://shop.example.com/logo.svg"
  payment_return_url            = "https://shop.example.com/checkout/return"
  active_payment_component_type = "DropIn"
  allow_discounts               = false
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(name, "description.%"),
					resource.TestCheckResourceAttr(name, "agreement.#", "0"),
					resource.TestCheckResourceAttr(name, "allow_all_origins", "true"),
					func(_ *terraform.State) error {
						a := application(t, key)
						assert.Nil(t, a.Description)
						assert.Empty(t, a.Agreements)
						assert.Equal(t, checkout.PaymentComponentTypeDropIn, a.PaymentsConfiguration.ActivePaymentComponentType)
						return nil
					},
				),
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateId:     key,
				ImportStateVerify: true,
			},
		},
	})
}
