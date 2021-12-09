package commercetools

import (
	"context"
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
			{
				Config: testAccNewStoreConfigWithoutChannels(name, key, languages),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_store.test", "distribution_channels.#", "0",
					),
				),
			},
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

func testAccNewStoreConfigWithoutChannels(name string, key string, languages []string) string {
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
	}
	`, name, key, languages)
}

func testAccCheckStoreDestroy(s *terraform.State) error {
	client := getClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "commercetools_store":
			{
				response, err := client.Stores().WithId(rs.Primary.ID).Get().Execute(context.Background())
				if err == nil {
					if response != nil && response.ID == rs.Primary.ID {
						return fmt.Errorf("store (%s) still exists", rs.Primary.ID)
					}
					continue
				}

				if newErr := checkApiResult(err); newErr != nil {
					return newErr
				}
			}
		case "commercetools_channel":
			{
				response, err := client.Channels().WithId(rs.Primary.ID).Get().Execute(context.Background())
				if err == nil {
					if response != nil && response.ID == rs.Primary.ID {
						return fmt.Errorf("supply channel (%s) still exists", rs.Primary.ID)
					}
					continue
				}
				if newErr := checkApiResult(err); newErr != nil {
					return newErr
				}
			}
		default:
			continue
		}
	}
	return nil
}
