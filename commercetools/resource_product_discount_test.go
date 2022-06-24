package commercetools

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"
)

func TestAccProductDiscountCreate_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProductDiscountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProductDiscountConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "key", "standard",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "name.en", "standard name",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "description.en", "Standard description",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "sort_order", "0.95",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "predicate", "1=1",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "valid_from", "2018-01-02T15:04:05Z",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "valid_until", "2019-01-02T15:04:05Z",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "value.0.type", "relative",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "value.0.permyriad", "1000",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "is_active", "true",
					),
					func(s *terraform.State) error {
						res, err := testGetProductDiscount(s, "commercetools_product_discount.standard")
						if err != nil {
							return err
						}

						key := "standard"

						validFrom, _ := expandTime("2018-01-02T15:04:05Z")
						validUntil, _ := expandTime("2019-01-02T15:04:05Z")

						expected := &platform.ProductDiscount{
							ID:             res.ID,
							Version:        res.Version,
							CreatedAt:      res.CreatedAt,
							CreatedBy:      res.CreatedBy,
							LastModifiedAt: res.LastModifiedAt,
							LastModifiedBy: res.LastModifiedBy,

							References: res.References, // TODO

							ValidFrom:  &validFrom,
							ValidUntil: &validUntil,

							Key: &key,
							Name: platform.LocalizedString{
								"en": "standard name",
							},
							Description: &platform.LocalizedString{
								"en": "Standard description",
							},
							IsActive:  true,
							Predicate: "1=1",
							SortOrder: "0.95",
							Value: platform.ProductDiscountValueRelative{
								Permyriad: 1000,
							},
						}

						assert.EqualValues(t, expected, res)

						return nil
					},
				),
			},
			{
				Config: testAccProductDiscountUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "key", "standard_new",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "name.en", "standard name new",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "description.en", "Standard description new",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "sort_order", "0.8",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "predicate", "1=1",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "valid_from", "2017-01-02T15:04:05Z",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "valid_until", "2018-01-02T15:04:05Z",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "value.0.type", "relative",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "value.0.permyriad", "856",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "is_active", "false",
					),
					func(s *terraform.State) error {
						res, err := testGetProductDiscount(s, "commercetools_product_discount.standard")
						if err != nil {
							return err
						}

						key := "standard_new"

						validFrom, _ := expandTime("2017-01-02T15:04:05Z")
						validUntil, _ := expandTime("2018-01-02T15:04:05Z")

						expected := &platform.ProductDiscount{
							ID:             res.ID,
							Version:        res.Version,
							CreatedAt:      res.CreatedAt,
							CreatedBy:      res.CreatedBy,
							LastModifiedAt: res.LastModifiedAt,
							LastModifiedBy: res.LastModifiedBy,

							References: res.References, // TODO

							ValidFrom:  &validFrom,
							ValidUntil: &validUntil,

							Key: &key,
							Name: platform.LocalizedString{
								"en": "standard name new",
							},
							Description: &platform.LocalizedString{
								"en": "Standard description new",
							},
							IsActive:  false,
							Predicate: "1=1",
							SortOrder: "0.8",
							Value: platform.ProductDiscountValueRelative{
								Permyriad: 856,
							},
						}

						assert.EqualValues(t, expected, res)

						return nil
					},
				),
			},
			{
				Config: testAccProductDiscountRemoveProperties(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "key", "standard_new",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "name.en", "standard name",
					),
					resource.TestCheckNoResourceAttr(
						"commercetools_product_discount.standard", "description.en",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "sort_order", "0.8",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "predicate", "1=1",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "valid_from", "",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "valid_until", "",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "value.0.type", "absolute",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "value.0.permyriad", "0",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "value.0.money.0.cent_amount", "42",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "value.0.money.0.currency_code", "EUR",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "is_active", "true",
					),
					func(s *terraform.State) error {
						res, err := testGetProductDiscount(s, "commercetools_product_discount.standard")
						if err != nil {
							return err
						}

						key := "standard_new"

						expected := &platform.ProductDiscount{
							ID:             res.ID,
							Version:        res.Version,
							CreatedAt:      res.CreatedAt,
							CreatedBy:      res.CreatedBy,
							LastModifiedAt: res.LastModifiedAt,
							LastModifiedBy: res.LastModifiedBy,

							References: res.References, // TODO

							Key: &key,
							Name: platform.LocalizedString{
								"en": "standard name",
							},
							IsActive:  true,
							Predicate: "1=1",
							SortOrder: "0.8",
							Value: platform.ProductDiscountValueAbsolute{
								Money: []platform.CentPrecisionMoney{
									{
										CurrencyCode:   "EUR",
										CentAmount:     42,
										FractionDigits: 2,
									},
								},
							},
						}

						assert.EqualValues(t, expected, res)

						return nil
					},
				),
			},
		},
	})
}

func testAccProductDiscountConfig() string {
	return `
	resource "commercetools_product_discount" "standard" {
		key = "standard"
		name = {
		  en = "standard name"
		}
		description = {
			en = "Standard description"
		  }
		predicate              = "1=1"
		sort_order             = "0.95"
		is_active              = true
		valid_from             = "2018-01-02T15:04:05Z"
		valid_until            = "2019-01-02T15:04:05Z"

		value {
			type      = "relative"
			permyriad = 1000
		}
	  }
	  `
}

func testAccProductDiscountUpdate() string {
	return `
	resource "commercetools_product_discount" "standard" {
		key = "standard_new"
		name = {
		  en = "standard name new"
		}
		description = {
			en = "Standard description new"
		  }
		sort_order             = "0.8"
		predicate              = "1=1"
		valid_from             = "2017-01-02T15:04:05Z"
		valid_until            = "2018-01-02T15:04:05Z"

		value {
			type      = "relative"
			permyriad = 856
		}

		is_active = false
	  }
	  `
}

func testAccProductDiscountRemoveProperties() string {
	return `
	resource "commercetools_product_discount" "standard" {
		key = "standard_new"
		name = {
		  en = "standard name"
		}
		sort_order             = "0.8"
		predicate              = "1=1"
		value {
			type          = "absolute"
			money {
				currency_code = "EUR"
				cent_amount   = 42
			}
		}
	  }
	  `
}

func testAccCheckProductDiscountDestroy(s *terraform.State) error {
	client := getClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_product_discount" {
			continue
		}
		response, err := client.ProductDiscounts().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("product discount (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		if newErr := checkApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}

func testGetProductDiscount(s *terraform.State, identifier string) (*platform.ProductDiscount, error) {
	rs, ok := s.RootModule().Resources[identifier]
	if !ok {
		return nil, fmt.Errorf("ProductDiscount not found")
	}

	client := getClient(testAccProvider.Meta())
	result, err := client.ProductDiscounts().WithId(rs.Primary.ID).Get().Execute(context.Background())
	if err != nil {
		return nil, err
	}
	return result, nil
}
