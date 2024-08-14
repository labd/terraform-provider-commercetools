package product

import (
	"encoding/json"
	"reflect"

	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

// Product represents the main schema data.
type Product struct {
	ID              types.String                     `tfsdk:"id"`
	Key             types.String                     `tfsdk:"key"`
	Version         types.Int64                      `tfsdk:"version"`
	ProductTypeId   types.String                     `tfsdk:"product_type_id"`
	Name            customtypes.LocalizedStringValue `tfsdk:"name"`
	Slug            customtypes.LocalizedStringValue `tfsdk:"slug"`
	Description     customtypes.LocalizedStringValue `tfsdk:"description"`
	Categories      []types.String                   `tfsdk:"categories"`
	MetaTitle       customtypes.LocalizedStringValue `tfsdk:"meta_title"`
	MetaDescription customtypes.LocalizedStringValue `tfsdk:"meta_description"`
	MetaKeywords    customtypes.LocalizedStringValue `tfsdk:"meta_keywords"`
	MasterVariant   ProductVariant                   `tfsdk:"master_variant"`
	Variants        []ProductVariant                 `tfsdk:"variant"`
	TaxCategoryId   types.String                     `tfsdk:"tax_category_id"`
	StateId         types.String                     `tfsdk:"state_id"`
	Publish         types.Bool                       `tfsdk:"publish"`
}

type ProductVariant struct {
	ID         types.Int64  `tfsdk:"id"`
	Key        types.String `tfsdk:"key"`
	Sku        types.String `tfsdk:"sku"`
	Attributes []Attribute  `tfsdk:"attribute"`
	Prices     []Price      `tfsdk:"price"`
}

type Attribute struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type Price struct {
	ID    types.String `tfsdk:"id"`
	Key   types.String `tfsdk:"key"`
	Value Money        `tfsdk:"value"`
}

type Money struct {
	CentAmount   types.Int64  `tfsdk:"cent_amount"`
	CurrencyCode types.String `tfsdk:"currency_code"`
}

func NewProductFromNative(p *platform.Product) Product {
	product := Product{
		ID:              types.StringValue(p.ID),
		Key:             utils.FromOptionalString(p.Key),
		Version:         types.Int64Value(int64(p.Version)),
		ProductTypeId:   types.StringValue(p.ProductType.ID),
		Name:            utils.FromLocalizedString(p.MasterData.Staged.Name),
		Slug:            utils.FromLocalizedString(p.MasterData.Staged.Slug),
		Description:     utils.FromOptionalLocalizedString(p.MasterData.Staged.Description),
		MetaTitle:       utils.FromOptionalLocalizedString(p.MasterData.Staged.MetaTitle),
		MetaDescription: utils.FromOptionalLocalizedString(p.MasterData.Staged.MetaDescription),
		MetaKeywords:    utils.FromOptionalLocalizedString(p.MasterData.Staged.MetaKeywords),
		MasterVariant:   NewProductVariantFromNative(p.MasterData.Staged.MasterVariant),
		Variants: pie.SortUsing(pie.Map(p.MasterData.Staged.Variants, NewProductVariantFromNative), func(a, b ProductVariant) bool {
			return a.ID.ValueInt64() < b.ID.ValueInt64()
		}),
		Publish: utils.FromOptionalBool(&p.MasterData.Published),
	}

	// Add product categories
	if len(p.MasterData.Staged.Categories) > 0 {
		product.Categories = pie.Map(p.MasterData.Staged.Categories, func(category platform.CategoryReference) types.String {
			return types.StringValue(category.ID)
		})
	}

	// Add Tax Category Id if defined
	if p.TaxCategory != nil {
		product.TaxCategoryId = types.StringValue(p.TaxCategory.ID)
	}

	// Add State Id if defined
	if p.State != nil {
		product.StateId = types.StringValue(p.State.ID)
	}

	return product
}

func NewProductVariantFromNative(p platform.ProductVariant) ProductVariant {
	return ProductVariant{
		ID:  types.Int64Value(int64(p.ID)),
		Key: utils.FromOptionalString(p.Key),
		Sku: utils.FromOptionalString(p.Sku),
		Attributes: pie.Map(p.Attributes, func(attribute platform.Attribute) Attribute {
			return Attribute{
				Name:  types.StringValue(attribute.Name),
				Value: types.StringValue(marshalAttributeValue(attribute)),
			}
		}),
		Prices: pie.Map(p.Prices, func(price platform.Price) Price {
			return Price{
				ID:  types.StringValue(price.ID),
				Key: utils.FromOptionalString(price.Key),
				Value: Money{
					CentAmount:   types.Int64Value(int64(price.Value.(platform.CentPrecisionMoney).CentAmount)),
					CurrencyCode: types.StringValue(price.Value.(platform.CentPrecisionMoney).CurrencyCode),
				},
			}
		}),
	}
}

func NewProductVariantFromNativeRef(p platform.ProductVariant) *ProductVariant {
	ref := NewProductVariantFromNative(p)
	return &ref
}

func (p Product) draft() platform.ProductDraft {
	productDraft := platform.ProductDraft{
		Key: p.Key.ValueStringPointer(),
		ProductType: platform.ProductTypeResourceIdentifier{
			ID: p.ProductTypeId.ValueStringPointer(),
		},
		Name:            p.Name.ValueLocalizedString(),
		Slug:            p.Slug.ValueLocalizedString(),
		Description:     p.Description.ValueLocalizedStringRef(),
		MetaTitle:       p.MetaTitle.ValueLocalizedStringRef(),
		MetaDescription: p.MetaDescription.ValueLocalizedStringRef(),
		MetaKeywords:    p.MetaKeywords.ValueLocalizedStringRef(),
		MasterVariant:   NewProductVariantDraftRef(p.MasterVariant),
		Variants:        pie.Map(p.Variants, NewProductVariantDraft),
		Categories: pie.Map(p.Categories, func(categoryId basetypes.StringValue) platform.CategoryResourceIdentifier {
			return platform.CategoryResourceIdentifier{
				ID: categoryId.ValueStringPointer(),
			}
		}),
		Publish: p.Publish.ValueBoolPointer(),
	}

	if !p.TaxCategoryId.IsNull() {
		productDraft.TaxCategory = &platform.TaxCategoryResourceIdentifier{
			ID: p.TaxCategoryId.ValueStringPointer(),
		}
	}

	if !p.StateId.IsNull() {
		productDraft.State = &platform.StateResourceIdentifier{
			ID: p.StateId.ValueStringPointer(),
		}
	}

	return productDraft
}

func NewProductVariantDraft(p ProductVariant) platform.ProductVariantDraft {
	return platform.ProductVariantDraft{
		Key:        p.Key.ValueStringPointer(),
		Sku:        p.Sku.ValueStringPointer(),
		Attributes: pie.Map(p.Attributes, NewProductVariantAttribute),
		Prices:     pie.Map(p.Prices, NewProductPriceDraft),
	}
}

func NewProductVariantDraftRef(p ProductVariant) *platform.ProductVariantDraft {
	productVariantDraft := NewProductVariantDraft(p)
	return &productVariantDraft
}

func NewProductPriceDraft(p Price) platform.PriceDraft {
	return platform.PriceDraft{
		Key: p.Key.ValueStringPointer(),
		Value: platform.Money{
			CentAmount:   int(p.Value.CentAmount.ValueInt64()),
			CurrencyCode: p.Value.CurrencyCode.ValueString(),
		},
	}
}

func NewProductVariantAttribute(a Attribute) platform.Attribute {
	return platform.Attribute{
		Name:  a.Name.ValueString(),
		Value: unmarshalAttributeValue(a.Value.ValueString()),
	}
}

func (p Product) updateActions(plan Product) platform.ProductUpdate {
	result := platform.ProductUpdate{
		Version: int(p.Version.ValueInt64()),
		Actions: []platform.ProductUpdateAction{},
	}

	// setKey
	if p.Key != plan.Key {
		result.Actions = append(result.Actions, platform.ProductSetKeyAction{
			Key: plan.Key.ValueStringPointer(),
		})
	}

	// changeName
	if !reflect.DeepEqual(p.Name, plan.Name) {
		result.Actions = append(result.Actions, platform.ProductChangeNameAction{
			Name:   plan.Name.ValueLocalizedString(),
			Staged: utils.BoolRef(false),
		})
	}

	// changeSlug
	if !reflect.DeepEqual(p.Slug, plan.Slug) {
		result.Actions = append(result.Actions, platform.ProductChangeSlugAction{
			Slug:   plan.Slug.ValueLocalizedString(),
			Staged: utils.BoolRef(false),
		})
	}

	// setDescription
	if !reflect.DeepEqual(p.Description, plan.Description) {
		result.Actions = append(result.Actions, platform.ProductSetDescriptionAction{
			Description: plan.Description.ValueLocalizedStringRef(),
			Staged:      utils.BoolRef(false),
		})
	}

	// setMetaTitle
	if !reflect.DeepEqual(p.MetaTitle, plan.MetaTitle) {
		result.Actions = append(result.Actions, platform.ProductSetMetaTitleAction{
			MetaTitle: plan.MetaTitle.ValueLocalizedStringRef(),
			Staged:    utils.BoolRef(false),
		})
	}

	// setMetaDescription
	if !reflect.DeepEqual(p.MetaDescription, plan.MetaDescription) {
		result.Actions = append(result.Actions, platform.ProductSetMetaDescriptionAction{
			MetaDescription: plan.MetaDescription.ValueLocalizedStringRef(),
			Staged:          utils.BoolRef(false),
		})
	}

	// setMetaKeywords
	if !reflect.DeepEqual(p.MetaKeywords, plan.MetaKeywords) {
		result.Actions = append(result.Actions, platform.ProductSetMetaKeywordsAction{
			MetaKeywords: plan.MetaKeywords.ValueLocalizedStringRef(),
			Staged:       utils.BoolRef(false),
		})
	}

	// publish
	if !p.Publish.ValueBool() && plan.Publish.ValueBool() {
		all := platform.ProductPublishScopeAll
		result.Actions = append(result.Actions, platform.ProductPublishAction{
			Scope: &all,
		})
	}

	// unpublish
	if p.Publish.ValueBool() && !plan.Publish.ValueBool() {
		result.Actions = append(result.Actions, platform.ProductUnpublishAction{})
	}

	// setTaxCategory
	if !p.TaxCategoryId.Equal(plan.TaxCategoryId) {
		if plan.TaxCategoryId.IsNull() {
			result.Actions = append(result.Actions, platform.ProductSetTaxCategoryAction{
				TaxCategory: nil,
			})
		} else {
			result.Actions = append(result.Actions, platform.ProductSetTaxCategoryAction{
				TaxCategory: &platform.TaxCategoryResourceIdentifier{
					ID: plan.TaxCategoryId.ValueStringPointer(),
				},
			})
		}
	}

	// transitionState
	if !p.StateId.Equal(plan.StateId) {
		result.Actions = append(result.Actions, platform.ProductTransitionStateAction{
			State: &platform.StateResourceIdentifier{
				ID: plan.StateId.ValueStringPointer(),
			},
			Force: utils.BoolRef(false),
		})
	}

	// # Category Actions
	categoriesDiffAdd, categoriesDiffRemove := pie.Diff(p.Categories, plan.Categories)
	// addToCategory
	for _, categoryId := range categoriesDiffAdd {
		result.Actions = append(result.Actions, platform.ProductAddToCategoryAction{
			Category: platform.CategoryResourceIdentifier{
				ID: categoryId.ValueStringPointer(),
			},
			Staged: utils.BoolRef(false),
		})
	}

	// removeFromCategory
	for _, categoryId := range categoriesDiffRemove {
		result.Actions = append(result.Actions, platform.ProductRemoveFromCategoryAction{
			Category: platform.CategoryResourceIdentifier{
				ID: categoryId.ValueStringPointer(),
			},
			Staged: utils.BoolRef(false),
		})
	}

	// # ProductVariants Actions
	currentVariants := append(p.Variants, p.MasterVariant)
	planVariants := append(plan.Variants, plan.MasterVariant)
	variantsAdded, _, variantsRemoved := compare(currentVariants, planVariants, "Sku")

	// addVariant
	for _, productVariant := range variantsAdded {
		result.Actions = append(result.Actions, platform.ProductAddVariantAction{
			Sku:        productVariant.Sku.ValueStringPointer(),
			Key:        productVariant.Key.ValueStringPointer(),
			Attributes: pie.Map(productVariant.Attributes, NewProductVariantAttribute),
			Prices:     pie.Map(productVariant.Prices, NewProductPriceDraft),
			Staged:     utils.BoolRef(false),
		})
	}

	// changeMasterVariant
	if !p.MasterVariant.ID.Equal(plan.MasterVariant.ID) {
		result.Actions = append(result.Actions, platform.ProductChangeMasterVariantAction{
			Sku:    plan.MasterVariant.Sku.ValueStringPointer(),
			Staged: utils.BoolRef(false),
		})
	}

	// removeVariant
	for _, productVariant := range variantsRemoved {
		result.Actions = append(result.Actions, platform.ProductRemoveVariantAction{
			Sku:    productVariant.Sku.ValueStringPointer(),
			Staged: utils.BoolRef(false),
		})
	}

	// # Compare Product Variant attributes and prices
	for _, currentVariant := range currentVariants {
		sku := currentVariant.Sku.ValueString()
		planVariantRef := getProductVariantBySku(planVariants, sku)
		if planVariantRef != nil {
			// Process Attributes
			attributesAdded, attributesModified, attributesRemoved := compare(currentVariant.Attributes, planVariantRef.Attributes, "Name")
			for _, attribute := range append(attributesAdded, attributesModified...) {
				// setAttribute
				result.Actions = append(result.Actions, platform.ProductSetAttributeAction{
					Sku:    &sku,
					Name:   attribute.Name.ValueString(),
					Staged: utils.BoolRef(false),
					Value:  unmarshalAttributeValue(attribute.Value.ValueString()),
				})
			}
			for _, attribute := range attributesRemoved {
				// removeAttribute
				result.Actions = append(result.Actions, platform.ProductSetAttributeAction{
					Sku:    &sku,
					Name:   attribute.Name.ValueString(),
					Staged: utils.BoolRef(false),
				})
			}

			// #Process Prices
			pricesAdded, pricesModified, pricesRemoved := compare(currentVariant.Prices, planVariantRef.Prices, "Key")

			// addPrice
			for _, price := range pricesAdded {
				result.Actions = append(result.Actions, platform.ProductAddPriceAction{
					Sku:    &sku,
					Price:  NewProductPriceDraft(price),
					Staged: utils.BoolRef(false),
				})
			}

			// changePrice
			for _, price := range pricesModified {
				result.Actions = append(result.Actions, platform.ProductChangePriceAction{
					PriceId: price.ID.ValueString(),
					Price:   NewProductPriceDraft(price),
					Staged:  utils.BoolRef(false),
				})
			}

			// removePrice
			for _, price := range pricesRemoved {
				result.Actions = append(result.Actions, platform.ProductRemovePriceAction{
					Sku:     &sku,
					PriceId: price.ID.ValueString(),
					Staged:  utils.BoolRef(false),
				})
			}
		}
	}

	return result
}

func unmarshalAttributeValue(value string) any {
	var data any
	json.Unmarshal([]byte(value), &data)
	return data
}

func marshalAttributeValue(o platform.Attribute) string {
	val, err := json.Marshal(o.Value)
	if err != nil {
		panic(err)
	}
	return string(val)
}

type ProductVariantComparable interface {
	ProductVariant | Attribute | Price
}

func compare[T ProductVariantComparable](current, planned []T, id string) (added, modified, removed []T) {
	currentSet := make(map[string]T, len(current))
	plannedSet := make(map[string]T, len(planned))

	for _, c := range current {
		key := reflect.ValueOf(c).FieldByName(id).Interface().(basetypes.StringValue).ValueString()
		currentSet[key] = c
	}

	for _, c := range planned {
		key := reflect.ValueOf(c).FieldByName(id).Interface().(basetypes.StringValue).ValueString()
		plannedSet[key] = c
	}

	// Find added/modified items
	for _, c := range planned {
		key := reflect.ValueOf(c).FieldByName(id).Interface().(basetypes.StringValue).ValueString()
		if cc, exists := currentSet[key]; !exists {
			added = append(added, c)
		} else {
			if !reflect.DeepEqual(c, cc) {
				modified = append(modified, c)
			}
		}
	}

	// Find removed items
	for _, c := range current {
		key := reflect.ValueOf(c).FieldByName(id).Interface().(basetypes.StringValue).ValueString()
		if _, exists := plannedSet[key]; !exists {
			removed = append(removed, c)
		}
	}

	return
}

func getProductVariantBySku(l []ProductVariant, sku string) *ProductVariant {
	for _, p := range l {
		if p.Sku.ValueString() == sku {
			return &p
		}
	}
	return nil
}
