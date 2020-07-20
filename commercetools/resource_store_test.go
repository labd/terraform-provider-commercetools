package commercetools

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccStore_createAndUpdateWithID(t *testing.T) {

	name := "test method"
	key := "test-method"
	languages := []string{"en-US"}

	newName := "new test method"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStoreConfig(name, key),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_store.standard", "name.en", name,
					),
					resource.TestCheckResourceAttr(
						"commercetools_store.standard", "key", key,
					),
				),
			},
			{
				Config: testAccStoreConfig(newName, key),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_store.standard", "name.en", newName,
					),
					resource.TestCheckResourceAttr(
						"commercetools_store.standard", "key", key,
					),
				),
			},
			{
				Config: testAccStoreConfigWithLanguages(name, key, languages),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_store.standard", "languages.#", "1",
					),
					resource.TestCheckResourceAttr(
						"commercetools_store.standard", "languages.0", "en-US",
					),
				),
			},
			{
				Config: testAccNewStoreConfigWithLanguages(name, key, languages),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_store.standard", "languages.#", "1",
					),
					resource.TestCheckResourceAttr(
						"commercetools_store.standard", "languages.0", "en-US",
					),
				),
			},
		},
	})
}

func TestAccStore_createAndUpdateDistributionLanguages(t *testing.T) {
	name := "test dl"
	key := "test-dl"
	languages := []string{"en-US"}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNewStoreConfigWithChannels(name, key, languages),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_store.test", "distribution_channels.#", "1",
					),
					resource.TestCheckResourceAttr(
						"commercetools_store.test", "distribution_channels.0", "TEST",
					),
				),
			},
		},
	})
}

func testAccStoreConfig(name string, key string) string {
	return fmt.Sprintf(`
	resource "commercetools_store" "standard" {
		name = {
			en = "%[1]s"
			nl = "%[1]s"
		}
		key = "%[2]s"
	}`, name, key)
}

func testAccStoreConfigWithLanguages(name string, key string, languages []string) string {
	return fmt.Sprintf(`
	resource "commercetools_store" "standard" {
		name = {
			en = "%[1]s"
			nl = "%[1]s"
		}
		key = "%[2]s"
		languages = %[3]q
	}`, name, key, languages)
}

func testAccNewStoreConfigWithLanguages(name string, key string, languages []string) string {
	return fmt.Sprintf(`
	resource "commercetools_store" "standard" {
		name = {
			en = "%[1]s"
			nl = "%[1]s"
		}
		key = "%[2]s"
		languages = %[3]q
	}`, name, key, languages)
}

func testAccNewStoreConfigWithChannels(name string, key string, languages []string) string {
	return fmt.Sprintf(`
	resource "commercetools_channel" "test_channel" {
		key = "TEST"
		roles = ["ProductDistribution"]
	}

	resource "commercetools_store" "test" {
		name = {
			en = "%[1]s"
			nl = "%[1]s"
		}
		key = "%[2]s"
		languages = %[3]q
		distribution_channels = [commercetools_channel.test_channel.key]
	}

	`, name, key, languages)
}

func testAccCheckStoreDestroy(s *terraform.State) error {
	return nil
}
