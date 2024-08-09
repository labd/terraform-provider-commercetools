package product

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewProductFromNative(t *testing.T) {
	productData := platform.ProductData{
		Name: platform.LocalizedString{"en-US": "Test product name"},
		Categories: []platform.CategoryReference{
			{
				ID: "00000000-0000-0000-0000-000000000000",
			},
		},
		Slug:            platform.LocalizedString{"en-US": "Test product slug"},
		Description:     &platform.LocalizedString{"en-US": "Test product description"},
		MetaTitle:       &platform.LocalizedString{"en-US": "Test product Meta Title"},
		MetaDescription: &platform.LocalizedString{"en-US": "Test product Meta Description"},
		MetaKeywords:    &platform.LocalizedString{"en-US": "Test product Meta Keywords"},
		MasterVariant: platform.ProductVariant{
			ID:  1,
			Sku: utils.StringRef("1001"),
			Key: utils.StringRef("variant-1"),
			Prices: []platform.Price{
				{
					ID:  "00000000-0000-0000-0000-000000000001",
					Key: utils.StringRef("price-1"),
					Value: platform.CentPrecisionMoney{
						CentAmount:   1000,
						CurrencyCode: "USD",
					},
				},
			},
			Attributes: []platform.Attribute{
				{
					Name:  "color",
					Value: platform.LocalizedString{"en-US": "Red"},
				},
			},
		},
		Variants: []platform.ProductVariant{
			{
				ID:  2,
				Sku: utils.StringRef("1002"),
				Key: utils.StringRef("variant-2"),
				Prices: []platform.Price{
					{
						ID:  "00000000-0000-0000-0000-000000000002",
						Key: utils.StringRef("price-1"),
						Value: platform.CentPrecisionMoney{
							CentAmount:   1000,
							CurrencyCode: "USD",
						},
					},
				},
				Attributes: []platform.Attribute{
					{
						Name:  "color",
						Value: platform.LocalizedString{"en-US": "Green"},
					},
				},
			},
		},
	}

	native := platform.Product{
		ID:      "00000000-0000-0000-0000-000000000003",
		Key:     utils.StringRef("test-product"),
		Version: 10,
		ProductType: platform.ProductTypeReference{
			ID: "00000000-0000-0000-0000-000000000004",
		},
		MasterData: platform.ProductCatalogData{
			Published: true,
			Current:   productData,
			Staged:    productData,
		},
		TaxCategory: &platform.TaxCategoryReference{
			ID: "00000000-0000-0000-0000-000000000005",
		},
		State: &platform.StateReference{
			ID: "00000000-0000-0000-0000-000000000006",
		},
	}
	productDraft := NewProductFromNative(&native)

	assert.Equal(t, Product{
		ID:            types.StringValue("00000000-0000-0000-0000-000000000003"),
		Key:           types.StringValue("test-product"),
		Version:       types.Int64Value(10),
		ProductTypeId: types.StringValue("00000000-0000-0000-0000-000000000004"),
		Name: customtypes.NewLocalizedStringValue(map[string]attr.Value{
			"en-US": types.StringValue("Test product name"),
		}),
		Slug: customtypes.NewLocalizedStringValue(map[string]attr.Value{
			"en-US": types.StringValue("Test product slug"),
		}),
		Description: customtypes.NewLocalizedStringValue(map[string]attr.Value{
			"en-US": types.StringValue("Test product description"),
		}),
		Categories: []types.String{
			types.StringValue("00000000-0000-0000-0000-000000000000"),
		},
		MetaTitle: customtypes.NewLocalizedStringValue(map[string]attr.Value{
			"en-US": types.StringValue("Test product Meta Title"),
		}),
		MetaDescription: customtypes.NewLocalizedStringValue(map[string]attr.Value{
			"en-US": types.StringValue("Test product Meta Description"),
		}),
		MetaKeywords: customtypes.NewLocalizedStringValue(map[string]attr.Value{
			"en-US": types.StringValue("Test product Meta Keywords"),
		}),
		MasterVariant: ProductVariant{
			ID:  types.Int64Value(1),
			Sku: types.StringValue("1001"),
			Key: types.StringValue("variant-1"),
			Prices: []Price{
				{
					ID:  types.StringValue("00000000-0000-0000-0000-000000000001"),
					Key: types.StringValue("price-1"),
					Value: Money{
						CentAmount:   types.Int64Value(1000),
						CurrencyCode: types.StringValue("USD"),
					},
				},
			},
			Attributes: []Attribute{
				{
					Name:  types.StringValue("color"),
					Value: types.StringValue("{\"en-US\":\"Red\"}"),
				},
			},
		},
		Variants: []ProductVariant{
			{
				ID:  types.Int64Value(2),
				Sku: types.StringValue("1002"),
				Key: types.StringValue("variant-2"),
				Prices: []Price{
					{
						ID:  types.StringValue("00000000-0000-0000-0000-000000000002"),
						Key: types.StringValue("price-1"),
						Value: Money{
							CentAmount:   types.Int64Value(1000),
							CurrencyCode: types.StringValue("USD"),
						},
					},
				},
				Attributes: []Attribute{
					{
						Name:  types.StringValue("color"),
						Value: types.StringValue("{\"en-US\":\"Green\"}"),
					},
				},
			},
		},
		TaxCategoryId: types.StringValue("00000000-0000-0000-0000-000000000005"),
		StateId:       types.StringValue("00000000-0000-0000-0000-000000000006"),
		Publish:       types.BoolValue(true),
	}, productDraft)

}

func Test_UpdateActions(t *testing.T) {
	productPublishScopeAll := platform.ProductPublishScopeAll
	productVariant1 := ProductVariant{
		ID:  types.Int64Value(1),
		Sku: types.StringValue("1001"),
		Key: types.StringValue("product-variant-1"),
	}
	productVariant2 := ProductVariant{
		ID:  types.Int64Value(2),
		Sku: types.StringValue("1002"),
		Key: types.StringValue("product-variant-2"),
	}
	productVariantWithAttribute1 := ProductVariant{
		ID:  types.Int64Value(1),
		Sku: types.StringValue("1001"),
		Key: types.StringValue("product-variant-1"),
		Attributes: []Attribute{
			{
				Name:  types.StringValue("color"),
				Value: types.StringValue("{\"en-US\":\"Red\"}"),
			},
		},
	}
	productVariantWithAttribute2 := ProductVariant{
		ID:  types.Int64Value(1),
		Sku: types.StringValue("1001"),
		Key: types.StringValue("product-variant-1"),
		Attributes: []Attribute{
			{
				Name:  types.StringValue("color"),
				Value: types.StringValue("{\"en-US\":\"Green\"}"),
			},
		},
	}
	productVariantWithPrice1 := ProductVariant{
		ID:  types.Int64Value(1),
		Sku: types.StringValue("1001"),
		Key: types.StringValue("product-variant-1"),
		Prices: []Price{
			{
				ID:  types.StringValue("price-1-id"),
				Key: types.StringValue("price-1-key"),
				Value: Money{
					CentAmount:   types.Int64Value(1000),
					CurrencyCode: types.StringValue("USD"),
				},
			},
		},
	}
	productVariantWithPrice2 := ProductVariant{
		ID:  types.Int64Value(1),
		Sku: types.StringValue("1001"),
		Key: types.StringValue("product-variant-1"),
		Prices: []Price{
			{
				ID:  types.StringValue("price-1-id"),
				Key: types.StringValue("price-1-key"),
				Value: Money{
					CentAmount:   types.Int64Value(2000),
					CurrencyCode: types.StringValue("USD"),
				},
			},
		},
	}
	testCases := []struct {
		name     string
		state    Product
		plan     Product
		expected platform.ProductUpdate
	}{
		{
			"product setKey",
			Product{
				Key: types.StringValue("product-key-1"),
			},
			Product{
				Key: types.StringValue("product-key-2"),
			},
			platform.ProductUpdate{
				Actions: []platform.ProductUpdateAction{
					platform.ProductSetKeyAction{
						Key: utils.StringRef("product-key-2"),
					},
				},
			},
		},
		{
			"product changeName",
			Product{
				Name: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"en-US": types.StringValue("Product name"),
				}),
			},
			Product{
				Name: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"en-US": types.StringValue("New product name"),
				}),
			},
			platform.ProductUpdate{
				Actions: []platform.ProductUpdateAction{
					platform.ProductChangeNameAction{
						Name:   platform.LocalizedString{"en-US": "New product name"},
						Staged: utils.BoolRef(false),
					},
				},
			},
		},
		{
			"product changeSlug",
			Product{
				Slug: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"en-US": types.StringValue("Product slug"),
				}),
			},
			Product{
				Slug: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"en-US": types.StringValue("New product slug"),
				}),
			},
			platform.ProductUpdate{
				Actions: []platform.ProductUpdateAction{
					platform.ProductChangeSlugAction{
						Slug:   platform.LocalizedString{"en-US": "New product slug"},
						Staged: utils.BoolRef(false),
					},
				},
			},
		},
		{
			"product setDescription",
			Product{
				Description: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"en-US": types.StringValue("Product description"),
				}),
			},
			Product{
				Description: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"en-US": types.StringValue("New product description"),
				}),
			},
			platform.ProductUpdate{
				Actions: []platform.ProductUpdateAction{
					platform.ProductSetDescriptionAction{
						Description: &platform.LocalizedString{"en-US": "New product description"},
						Staged:      utils.BoolRef(false),
					},
				},
			},
		},
		{
			"product setMetaTitle",
			Product{
				MetaTitle: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"en-US": types.StringValue("Product meta title"),
				}),
			},
			Product{
				MetaTitle: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"en-US": types.StringValue("New product meta title"),
				}),
			},
			platform.ProductUpdate{
				Actions: []platform.ProductUpdateAction{
					platform.ProductSetMetaTitleAction{
						MetaTitle: &platform.LocalizedString{"en-US": "New product meta title"},
						Staged:    utils.BoolRef(false),
					},
				},
			},
		},
		{
			"product setMetaDescription",
			Product{
				MetaDescription: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"en-US": types.StringValue("Product meta description"),
				}),
			},
			Product{
				MetaDescription: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"en-US": types.StringValue("New product meta description"),
				}),
			},
			platform.ProductUpdate{
				Actions: []platform.ProductUpdateAction{
					platform.ProductSetMetaDescriptionAction{
						MetaDescription: &platform.LocalizedString{"en-US": "New product meta description"},
						Staged:          utils.BoolRef(false),
					},
				},
			},
		},
		{
			"product setMetaKeywords",
			Product{
				MetaKeywords: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"en-US": types.StringValue("Product meta keywords"),
				}),
			},
			Product{
				MetaKeywords: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"en-US": types.StringValue("New product meta keywords"),
				}),
			},
			platform.ProductUpdate{
				Actions: []platform.ProductUpdateAction{
					platform.ProductSetMetaKeywordsAction{
						MetaKeywords: &platform.LocalizedString{"en-US": "New product meta keywords"},
						Staged:       utils.BoolRef(false),
					},
				},
			},
		},
		{
			"product publish",
			Product{
				Publish: types.BoolValue(false),
			},
			Product{
				Publish: types.BoolValue(true),
			},
			platform.ProductUpdate{
				Actions: []platform.ProductUpdateAction{
					platform.ProductPublishAction{
						Scope: &productPublishScopeAll,
					},
				},
			},
		},
		{
			"product unpublish",
			Product{
				Publish: types.BoolValue(true),
			},
			Product{
				Publish: types.BoolValue(false),
			},
			platform.ProductUpdate{
				Actions: []platform.ProductUpdateAction{
					platform.ProductUnpublishAction{},
				},
			},
		},
		{
			"product setTaxCategory",
			Product{
				TaxCategoryId: types.StringValue("category-id-1"),
			},
			Product{
				TaxCategoryId: types.StringValue("category-id-2"),
			},
			platform.ProductUpdate{
				Actions: []platform.ProductUpdateAction{
					platform.ProductSetTaxCategoryAction{
						TaxCategory: &platform.TaxCategoryResourceIdentifier{
							ID: utils.StringRef("category-id-2"),
						},
					},
				},
			},
		},
		{
			"product transitionState",
			Product{
				StateId: types.StringValue("state-id-1"),
			},
			Product{
				StateId: types.StringValue("state-id-2"),
			},
			platform.ProductUpdate{
				Actions: []platform.ProductUpdateAction{
					platform.ProductTransitionStateAction{
						State: &platform.StateResourceIdentifier{
							ID: utils.StringRef("state-id-2"),
						},
						Force: utils.BoolRef(false),
					},
				},
			},
		},
		{
			"product addToCategory",
			Product{
				Categories: []basetypes.StringValue{},
			},
			Product{
				Categories: []basetypes.StringValue{
					types.StringValue("category-1"),
				},
			},
			platform.ProductUpdate{
				Actions: []platform.ProductUpdateAction{
					platform.ProductAddToCategoryAction{
						Category: platform.CategoryResourceIdentifier{
							ID: utils.StringRef("category-1"),
						},
						Staged: utils.BoolRef(false),
					},
				},
			},
		},
		{
			"product removeFromCategory",
			Product{
				Categories: []basetypes.StringValue{
					types.StringValue("category-1"),
				},
			},
			Product{
				Categories: []basetypes.StringValue{},
			},
			platform.ProductUpdate{
				Actions: []platform.ProductUpdateAction{
					platform.ProductRemoveFromCategoryAction{
						Category: platform.CategoryResourceIdentifier{
							ID: utils.StringRef("category-1"),
						},
						Staged: utils.BoolRef(false),
					},
				},
			},
		},
		{
			"product addVariant",
			Product{
				MasterVariant: productVariant1,
				Variants:      []ProductVariant{},
			},
			Product{
				MasterVariant: productVariant1,
				Variants: []ProductVariant{
					productVariant2,
				},
			},
			platform.ProductUpdate{
				Actions: []platform.ProductUpdateAction{
					platform.ProductAddVariantAction{
						Sku:    productVariant2.Sku.ValueStringPointer(),
						Key:    productVariant2.Key.ValueStringPointer(),
						Staged: utils.BoolRef(false),
					},
				},
			},
		},
		{
			"product changeMasterVariant",
			Product{
				MasterVariant: productVariant1,
				Variants: []ProductVariant{
					productVariant2,
				},
			},
			Product{
				MasterVariant: productVariant2,
				Variants: []ProductVariant{
					productVariant1,
				},
			},
			platform.ProductUpdate{
				Actions: []platform.ProductUpdateAction{
					platform.ProductChangeMasterVariantAction{
						Sku:    productVariant2.Sku.ValueStringPointer(),
						Staged: utils.BoolRef(false),
					},
				},
			},
		},
		{
			"product removeVariant",
			Product{
				MasterVariant: productVariant1,
				Variants: []ProductVariant{
					productVariant2,
				},
			},
			Product{
				MasterVariant: productVariant1,
				Variants:      []ProductVariant{},
			},
			platform.ProductUpdate{
				Actions: []platform.ProductUpdateAction{
					platform.ProductRemoveVariantAction{
						Sku:    productVariant2.Sku.ValueStringPointer(),
						Staged: utils.BoolRef(false),
					},
				},
			},
		},
		{
			"product setAttribute",
			Product{
				MasterVariant: productVariantWithAttribute1,
			},
			Product{
				MasterVariant: productVariantWithAttribute2,
			},
			platform.ProductUpdate{
				Actions: []platform.ProductUpdateAction{
					platform.ProductSetAttributeAction{
						Sku:    productVariantWithAttribute2.Sku.ValueStringPointer(),
						Name:   productVariantWithAttribute2.Attributes[0].Name.ValueString(),
						Value:  unmarshalAttributeValue(productVariantWithAttribute2.Attributes[0].Value.ValueString()),
						Staged: utils.BoolRef(false),
					},
				},
			},
		},
		{
			"product removeAttribute",
			Product{
				MasterVariant: productVariantWithAttribute1,
			},
			Product{
				MasterVariant: productVariant1,
			},
			platform.ProductUpdate{
				Actions: []platform.ProductUpdateAction{
					platform.ProductSetAttributeAction{
						Sku:    productVariantWithAttribute1.Sku.ValueStringPointer(),
						Name:   productVariantWithAttribute1.Attributes[0].Name.ValueString(),
						Staged: utils.BoolRef(false),
					},
				},
			},
		},
		{
			"product addPrice",
			Product{
				MasterVariant: productVariant1,
			},
			Product{
				MasterVariant: productVariantWithPrice1,
			},
			platform.ProductUpdate{
				Actions: []platform.ProductUpdateAction{
					platform.ProductAddPriceAction{
						Sku: productVariant1.Sku.ValueStringPointer(),
						Price: platform.PriceDraft{
							Key: productVariantWithPrice1.Prices[0].Key.ValueStringPointer(),
							Value: platform.Money{
								CentAmount:   int(productVariantWithPrice1.Prices[0].Value.CentAmount.ValueInt64()),
								CurrencyCode: productVariantWithPrice1.Prices[0].Value.CurrencyCode.ValueString(),
							},
						},
						Staged: utils.BoolRef(false),
					},
				},
			},
		},
		{
			"product changePrice",
			Product{
				MasterVariant: productVariantWithPrice1,
			},
			Product{
				MasterVariant: productVariantWithPrice2,
			},
			platform.ProductUpdate{
				Actions: []platform.ProductUpdateAction{
					platform.ProductChangePriceAction{
						PriceId: productVariantWithPrice2.Prices[0].ID.ValueString(),
						Price: platform.PriceDraft{
							Key: productVariantWithPrice2.Prices[0].Key.ValueStringPointer(),
							Value: platform.Money{
								CentAmount:   int(productVariantWithPrice2.Prices[0].Value.CentAmount.ValueInt64()),
								CurrencyCode: productVariantWithPrice2.Prices[0].Value.CurrencyCode.ValueString(),
							},
						},
						Staged: utils.BoolRef(false),
					},
				},
			},
		},
		{
			"product removePrice",
			Product{
				MasterVariant: productVariantWithPrice1,
			},
			Product{
				MasterVariant: productVariant1,
			},
			platform.ProductUpdate{
				Actions: []platform.ProductUpdateAction{
					platform.ProductRemovePriceAction{
						PriceId: productVariantWithPrice1.Prices[0].ID.ValueString(),
						Sku:     productVariantWithPrice1.Sku.ValueStringPointer(),
						Staged:  utils.BoolRef(false),
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.state.updateActions(tc.plan)
			assert.EqualValues(t, tc.expected, result)
		})
	}
}
