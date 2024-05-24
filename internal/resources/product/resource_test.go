package product_test

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/terraform-provider-commercetools/internal/acctest"
)

var templateData = map[string]string{
	"identifier":                 "test_product",
	"key":                        "test-product-key",
	"name":                       "test-product-name",
	"slug":                       "test-product-slug",
	"description":                "Test product description",
	"metaTitle":                  "meta-title",
	"metaDescription":            "meta-description",
	"metaKeywords":               "meta-keywords",
	"addVariant1":                "false",
	"addVariant2":                "false",
	"setTaxCategory":             "true",
	"taxCategoryRef":             "external_shipping_tax",
	"masterVariant":              "master-variant-key",
	"addPrice":                   "false",
	"addPriceValue":              "1000",
	"addToCategory":              "false",
	"published":                  "false",
	"stateName":                  "product_state_for_sale",
	"masterVariantNameAttrValue": "Test product basic variant",
}

func TestAccProductResource_Create(t *testing.T) {
	testData := copyMap(templateData)
	resourceName := "commercetools_product." + testData["identifier"]

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: getResourceConfig(testData),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkProductReference(
						testData["identifier"], "tax_category_id", "commercetools_tax_category", testData["taxCategoryRef"]),
					resource.TestCheckResourceAttr(resourceName, "key", testData["key"]),
					resource.TestCheckResourceAttr(resourceName, "name.en-GB", testData["name"]),
					resource.TestCheckResourceAttr(resourceName, "slug.en-GB", testData["slug"]),
					resource.TestCheckResourceAttr(resourceName, "description.en-GB", "Test product description"),
					resource.TestCheckResourceAttr(resourceName, "meta_title.en-GB", "meta-title"),
					resource.TestCheckResourceAttr(resourceName, "meta_description.en-GB", "meta-description"),
					resource.TestCheckResourceAttr(resourceName, "meta_keywords.en-GB", "meta-keywords"),
					resource.TestCheckResourceAttr(resourceName, "publish", "false"),
					resource.TestCheckResourceAttr(resourceName, "master_variant.key", "master-variant-key"),
					resource.TestCheckResourceAttr(resourceName, "master_variant.sku", "100000"),
					resource.TestCheckResourceAttr(resourceName, "master_variant.attribute.0.name", "name"),
					resource.TestCheckResourceAttr(resourceName, "master_variant.attribute.0.value", "{\"en-GB\":\"Test product basic variant\"}"),
					resource.TestCheckResourceAttr(resourceName, "master_variant.attribute.1.name", "description"),
					resource.TestCheckResourceAttr(resourceName, "master_variant.attribute.1.value", "{\"en-GB\":\"Test product basic variant description\"}"),
					resource.TestCheckResourceAttr(resourceName, "master_variant.price.0.key", "base_price_eur"),
					resource.TestCheckResourceAttr(resourceName, "master_variant.price.0.value.cent_amount", "1000000"),
					resource.TestCheckResourceAttr(resourceName, "master_variant.price.0.value.currency_code", "EUR"),
					resource.TestCheckResourceAttr(resourceName, "master_variant.price.1.key", "base_price_gbr"),
					resource.TestCheckResourceAttr(resourceName, "master_variant.price.1.value.cent_amount", "872795"),
					resource.TestCheckResourceAttr(resourceName, "master_variant.price.1.value.currency_code", "GBP"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.key", "variant-1-key"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.sku", "100001"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.attribute.0.name", "name"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.attribute.0.value", "{\"en-GB\":\"Test product variant one\"}"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.attribute.1.name", "description"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.attribute.1.value", "{\"en-GB\":\"Test product variant one description\"}"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.price.0.key", "base_price_eur"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.price.0.value.cent_amount", "1010000"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.price.0.value.currency_code", "EUR"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.price.1.key", "base_price_gbr"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.price.1.value.cent_amount", "880299"),
					resource.TestCheckResourceAttr(resourceName, "variant.0.price.1.value.currency_code", "GBP"),
				),
			},
		},
	})
}

func TestAccProductResource_Update(t *testing.T) {
	testData := copyMap(templateData)
	resourceName := "commercetools_product." + testData["identifier"]

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: getResourceConfig(testData),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", testData["key"]),
				),
			},
			{
				// Test setKey action
				PreConfig: func() { fmt.Println(" - Test setKey action") },
				Config:    getUpdatedResourceConfig(testData, "key", "new-key-value"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "new-key-value"),
				),
			},
			{
				// Test changeName action
				PreConfig: func() { fmt.Println(" - Test changeName action") },
				Config:    getUpdatedResourceConfig(testData, "name", "new-test-product-name"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name.en-GB", "new-test-product-name"),
				),
			},
			{
				// Test changeSlug action
				PreConfig: func() { fmt.Println(" - Test changeSlug action") },
				Config:    getUpdatedResourceConfig(testData, "slug", "new-test-product-slug"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "slug.en-GB", "new-test-product-slug"),
				),
			},
			{
				// Test setDescription action
				PreConfig: func() { fmt.Println(" - Test setDescription action") },
				Config:    getUpdatedResourceConfig(testData, "description", "New Test product description"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description.en-GB", "New Test product description"),
				),
			},
			{
				// Test setMetaTitle action
				PreConfig: func() { fmt.Println(" - Test setMetaTitle action") },
				Config:    getUpdatedResourceConfig(testData, "metaTitle", "new-meta-title"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "meta_title.en-GB", "new-meta-title"),
				),
			},
			{
				// Test setMetaDescription action
				PreConfig: func() { fmt.Println(" - Test setMetaDescription action") },
				Config:    getUpdatedResourceConfig(testData, "metaDescription", "new-meta-description"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "meta_description.en-GB", "new-meta-description"),
				),
			},
			{
				// Test setMetaKeywords action
				PreConfig: func() { fmt.Println(" - Test setMetaKeywords action") },
				Config:    getUpdatedResourceConfig(testData, "metaKeywords", "new-meta-keywords"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "meta_keywords.en-GB", "new-meta-keywords"),
				),
			},
			{
				// Test addVariant action 1
				PreConfig: func() { fmt.Println(" - Test addVariant action 1") },
				Config:    getUpdatedResourceConfig(testData, "addVariant1", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "variant.1.key", "variant-2-key"),
				),
			},
			{
				// Test addVariant action 2
				PreConfig: func() { fmt.Println(" - Test addVariant action 2") },
				Config:    getUpdatedResourceConfig(testData, "addVariant2", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "variant.2.key", "variant-3-key"),
				),
			},
			{
				// Test removeVariant action
				PreConfig: func() { fmt.Println(" - Test removeVariant action") },
				Config:    getUpdatedResourceConfig(testData, "addVariant1", "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "variant.1.key", "variant-3-key"),
				),
			},
			{
				// Test setTaxCategory action
				PreConfig: func() { fmt.Println(" - Test setTaxCategory action") },
				Config:    getUpdatedResourceConfig(testData, "taxCategoryRef", "vat_tax"),
				Check: resource.ComposeTestCheckFunc(
					checkProductReference(
						testData["identifier"], "tax_category_id", "commercetools_tax_category", "vat_tax"),
				),
			},
			{
				// Test addPrice action
				PreConfig: func() { fmt.Println(" - Test addPrice action") },
				Config:    getUpdatedResourceConfig(testData, "addPrice", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "master_variant.price.2.key", "base_price_usd"),
				),
			},
			// #TBD Mock server bug, can't identify the issue for the moment
			// Test passes successfully on the real environment
			// {
			// 	// Test changePrice action
			// 	PreConfig: func() { fmt.Println(" - Test changePrice action") },
			// 	Config:    getUpdatedResourceConfig(testData, "addPriceValue", "9999"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr(resourceName, "master_variant.price.2.value.cent_amount", "9999"),
			// 	),
			// },
			{
				// Test removePrice action
				PreConfig: func() { fmt.Println(" - Test removePrice action") },
				Config:    getUpdatedResourceConfig(testData, "addPrice", "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "master_variant.price.2.key"),
				),
			},
			{
				// Test addToCategory action
				PreConfig: func() { fmt.Println(" - Test addToCategory action") },
				Config:    getUpdatedResourceConfig(testData, "addToCategory", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "categories.1"),
				),
			},
			{
				// Test removeFromCategory action
				PreConfig: func() { fmt.Println(" - Test removeFromCategory action") },
				Config:    getUpdatedResourceConfig(testData, "addToCategory", "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "categories.1"),
				),
			},
			{
				// Test publish action
				PreConfig: func() { fmt.Println(" - Test publish action") },
				Config:    getUpdatedResourceConfig(testData, "published", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "publish", "true"),
				),
			},
			{
				// Test unpublish action
				PreConfig: func() { fmt.Println(" - Test unpublish action") },
				Config:    getUpdatedResourceConfig(testData, "published", "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "publish", "false"),
				),
			},
			{
				// Test transitionState action
				PreConfig: func() { fmt.Println(" - Test transitionState action") },
				Config:    getUpdatedResourceConfig(testData, "stateName", "product_out_of_stock"),
				Check: resource.ComposeTestCheckFunc(
					checkProductReference(
						testData["identifier"], "state_id", "commercetools_state", "product_out_of_stock"),
				),
			},
			{
				// Test setAttribute action
				PreConfig: func() { fmt.Println(" - Test setAttribute action") },
				Config:    getUpdatedResourceConfig(testData, "masterVariantNameAttrValue", "New name"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "master_variant.attribute.0.value", "{\"en-GB\":\"New name\"}"),
				),
			},
			{
				// Test changeMasterVariant action
				PreConfig: func() { fmt.Println(" - Test changeMasterVariant action") },
				Config:    getUpdatedResourceConfig(testData, "masterVariant", "variant-1-key"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "master_variant.sku", "100001"),
				),
			},
		},
	})
}

func testAccProductDestroy(s *terraform.State) error {
	client, err := acctest.GetClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_product" {
			continue
		}
		response, err := client.Products().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("product (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		if newErr := acctest.CheckApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}

func getResourceConfig(data map[string]string) string {
	// Load templates
	tpl, err := template.ParseGlob("testdata/*")
	if err != nil {
		panic(err)
	}

	var out bytes.Buffer
	err = tpl.ExecuteTemplate(&out, "main", data)
	if err != nil {
		panic(err)
	}

	return out.String()
}

func getUpdatedResourceConfig(data map[string]string, key, value string) string {
	// Update map value
	data[key] = value
	return getResourceConfig(data)
}

func copyMap(srcMap map[string]string) map[string]string {
	newMap := make(map[string]string, len(srcMap))

	for k, v := range srcMap {
		newMap[k] = v
	}

	return newMap
}

func checkProductReference(productName, productRefAttribute, refResourceType, refResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Retrieve product state
		productResourceName := "commercetools_product." + productName
		productResourceState, ok := s.RootModule().Resources[productResourceName]
		if !ok {
			return fmt.Errorf("Product '%v' not found", productResourceName)
		}
		productRefId := productResourceState.Primary.Attributes[productRefAttribute]

		// Retrieve referenced resource
		refId := ""
		if refResourceName != "" { // if empty string is passed, no reference is expected
			// Retrieve referenced resource
			refResourceName := fmt.Sprintf("%s.%s", refResourceType, refResourceName)
			refResourceState, ok := s.RootModule().Resources[refResourceName]
			if !ok {
				return fmt.Errorf("Resource '%s' of type '%s' not found", refResourceName, refResourceType)
			}
			refId = refResourceState.Primary.ID
		}

		if productRefId != refId {
			return fmt.Errorf("Attribute '%s' expected '%s', got '%s'", productRefAttribute, refId, productRefId)
		}

		return nil
	}
}
