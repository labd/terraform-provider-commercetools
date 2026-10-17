package integration_layer_allowlist_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/commercetools-go-sdk/integrationlayer"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
)

func TestAccIntegrationLayerAllowlist(t *testing.T) {
	name := "commercetools_integration_layer_allowlist.allowlist"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: func(_ *terraform.State) error {
			if len(allow(t)) != 0 {
				return errors.New("hosts are still allowed after destroy")
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: acctest.ProviderConfig(projectKey) + `
resource "commercetools_integration_layer_allowlist" "allowlist" {
  allow = ["CDN.contentful.com"]
}
`,
				ExpectError: regexp.MustCompile(`must be a lowercase host`),
			},
			{
				Config: acctest.ProviderConfig(projectKey) + `
resource "commercetools_integration_layer_allowlist" "allowlist" {
  allow = ["cdn.contentful.com", "*.algolia.net", "*.algolianet.com"]
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "allow.#", "3"),
					resource.TestCheckTypeSetElemAttr(name, "deny.*", "*.commercetools.app"),
					func(_ *terraform.State) error {
						assert.ElementsMatch(t,
							[]string{"cdn.contentful.com", "*.algolia.net", "*.algolianet.com"}, allow(t))
						return nil
					},
				),
			},
			{
				Config: acctest.ProviderConfig(projectKey) + `
resource "commercetools_integration_layer_allowlist" "allowlist" {
  allow = ["cdn.contentful.com", "api.addressy.com"]
}
`,
				Check: func(_ *terraform.State) error {
					assert.ElementsMatch(t, []string{"cdn.contentful.com", "api.addressy.com"}, allow(t))
					return nil
				},
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
const projectKey = "acc-cil-allowlist"

func client(t *testing.T) *integrationlayer.ByProjectKeyRequestBuilder {
	c, err := acctest.GetIntegrationLayerClient(projectKey)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func allow(t *testing.T) []string {
	list, err := client(t).Main().Extension().Allowlist().Get().Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return list.Allow
}
