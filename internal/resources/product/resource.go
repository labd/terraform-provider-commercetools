package product

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"time"

	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

// productResource is the resource implementation.
type productResource struct {
	client *platform.ByProjectKeyRequestBuilder
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &productResource{}
}

// The Schema definition for the Attribute
var productVariantAttributeSchema = map[string]schema.Attribute{
	"name": schema.StringAttribute{
		Description:         "Name of the Attribute.",
		MarkdownDescription: "Name of the Attribute.",
		Required:            true,
	},
	"value": schema.StringAttribute{
		Description:         "The AttributeType determines the format of the Attribute value to be provided.",
		MarkdownDescription: "The AttributeType determines the format of the Attribute value to be provided.",
		Required:            true,
	},
}

// The Schema definition for the Product Price
var productPriceDraftSchema = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"key": schema.StringAttribute{
		Description: "User-defined identifier for the Price. It must be unique per ProductVariant." +
			"MinLength: 2, MaxLength: 256, Pattern: ^[A-Za-z0-9_-]+$",
		MarkdownDescription: "User-defined identifier for the Price. It must be unique per [ProductVariant](https://docs.commercetools.com/api/projects/products#ctp:api:type:ProductVariant)." +
			"`MinLength: 2` `MaxLength: 256` `Pattern: ^[A-Za-z0-9_-]+$`",
		Optional: true,
		Validators: []validator.String{
			stringvalidator.LengthBetween(2, 256),
			stringvalidator.RegexMatches(
				regexp.MustCompile("^[A-Za-z0-9_-]+$"),
				"Key must match pattern ^[A-Za-z0-9_-]+$"),
		},
	},
}

// The Schema definition for the Money
var moneySchema = map[string]schema.Attribute{
	"cent_amount": schema.Int64Attribute{
		Description: "Amount in the smallest indivisible unit of a currency, such as:" +
			"Cents for EUR and USD, pence for GBP, or centime for CHF (5 CHF is specified as 500)." +
			"The value in the major unit for currencies without minor units, like JPY (5 JPY is specified as 5).",
		MarkdownDescription: "Amount in the smallest indivisible unit of a currency, such as:" +
			"- Cents for EUR and USD, pence for GBP, or centime for CHF (5 CHF is specified as `500`)." +
			"- The value in the major unit for currencies without minor units, like JPY (5 JPY is specified as `5`).",
		Required: true,
	},
	"currency_code": schema.StringAttribute{
		Description:         "Currency code compliant to ISO 4217.",
		MarkdownDescription: "Currency code compliant to [ISO 4217](https://en.wikipedia.org/wiki/ISO_4217).",
		Required:            true,
		Validators: []validator.String{
			stringvalidator.RegexMatches(
				regexp.MustCompile("^[A-Z]{3}$"),
				"Currency Code must match pattern ^[A-Z]{3}$"),
		},
	},
}

// The Schema definition for the ProductVariant
var productVariantSchema = map[string]schema.Attribute{
	"id": schema.Int64Attribute{
		Computed: true,
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.UseStateForUnknown(),
		},
	},
	"key": schema.StringAttribute{
		Description:         "User-defined unique identifier for the ProductVariant.",
		MarkdownDescription: "User-defined unique identifier for the ProductVariant.",
		Optional:            true,
	},
	"sku": schema.StringAttribute{
		Description:         "User-defined unique SKU of the Product Variant.",
		MarkdownDescription: "User-defined unique SKU of the Product Variant.",
		Optional:            true,
	},
}

// Schema defines the schema for the data source.
func (r *productResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 1,
		Description: "An abstract sellable good with a set of Attributes defined by a Product Type. Products " +
			"themselves are not sellable. Instead, they act as a parent structure for Product Variants. Each Product " +
			"must have at least one Product Variant, which is called the Master Variant. A single Product " +
			"representation contains the current and the staged representation of its product data.",
		MarkdownDescription: "An abstract sellable good with a set of Attributes defined by a Product Type. Products " +
			"themselves are not sellable. Instead, they act as a parent structure for Product Variants. Each Product " +
			"must have at least one Product Variant, which is called the Master Variant. A single Product " +
			"representation contains the *current* and the *staged* representation of its product data.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.Int64Attribute{
				Computed: true,
			},
			"key": schema.StringAttribute{
				Description:         "User-defined unique identifier of the Product.",
				MarkdownDescription: "User-defined unique identifier of the Product.",
				Optional:            true,
			},
			"product_type_id": schema.StringAttribute{
				Description:         "The Product Type defining the Attributes for the Product. Cannot be changed later.",
				MarkdownDescription: "The Product Type defining the Attributes for the Product. Cannot be changed later.",
				Required:            true,
			},
			"name": schema.MapAttribute{
				CustomType:          customtypes.NewLocalizedStringType(),
				Description:         "Name of the Product.",
				MarkdownDescription: "Name of the Product.",
				Required:            true,
			},
			"slug": schema.MapAttribute{
				CustomType: customtypes.NewLocalizedStringType(),
				Description: "User-defined identifier used in a deep-link URL for the Product. It must be unique " +
					"across a Project, but a Product can have the same slug in different Locales. It must match the " +
					"pattern [a-zA-Z0-9_-]{2,256}.",
				MarkdownDescription: "User-defined identifier used in a deep-link URL for the Product. It must be " +
					"unique across a Project, but a Product can have the same slug in different " +
					"[Locales](https://docs.commercetools.com/api/types#ctp:api:type:Locale). It must match the " +
					"pattern `[a-zA-Z0-9_-]{2,256}`.",
				Required: true,
			},
			"description": schema.MapAttribute{
				CustomType:          customtypes.NewLocalizedStringType(),
				Description:         "Description of the Product.",
				MarkdownDescription: "Description of the Product.",
				Optional:            true,
			},
			"categories": schema.ListAttribute{
				ElementType:         types.StringType,
				Description:         "Categories assigned to the Product.",
				MarkdownDescription: "Categories assigned to the Product.",
				Optional:            true,
				Validators:          []validator.List{listvalidator.SizeAtLeast(1)},
			},
			"meta_title": schema.MapAttribute{
				CustomType:          customtypes.NewLocalizedStringType(),
				Description:         "Title of the Product displayed in search results.",
				MarkdownDescription: "Title of the Product displayed in search results.",
				Optional:            true,
			},
			"meta_description": schema.MapAttribute{
				CustomType:          customtypes.NewLocalizedStringType(),
				Description:         "Description of the Product displayed in search results.",
				MarkdownDescription: "Description of the Product displayed in search results.",
				Optional:            true,
			},
			"meta_keywords": schema.MapAttribute{
				CustomType:          customtypes.NewLocalizedStringType(),
				Description:         "Keywords that give additional information about the Product to search engines.",
				MarkdownDescription: "Keywords that give additional information about the Product to search engines.",
				Optional:            true,
			},
			"tax_category_id": schema.StringAttribute{
				Description:         "The Tax Category to be assigned to the Product.",
				MarkdownDescription: "The Tax Category to be assigned to the Product.",
				Optional:            true,
			},
			"state_id": schema.StringAttribute{
				Description:         "State to be assigned to the Product.",
				MarkdownDescription: "State to be assigned to the Product.",
				Optional:            true,
			},
			"publish": schema.BoolAttribute{
				Description:         "If true, the Product is published immediately to the current projection. Default: false",
				MarkdownDescription: "If `true`, the Product is published immediately to the current projection. Default: `false`",
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"master_variant": schema.SingleNestedBlock{
				Attributes:          productVariantSchema,
				Description:         "The Product Variant to be the Master Variant for the Product. Required if variants are provided also.",
				MarkdownDescription: "The Product Variant to be the Master Variant for the Product. Required if `variants` are provided also.",
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
				Blocks: map[string]schema.Block{
					"attribute": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: productVariantAttributeSchema,
						},
					},
					"price": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: productPriceDraftSchema,
							Blocks: map[string]schema.Block{
								"value": schema.SingleNestedBlock{
									Attributes: moneySchema,
								},
							},
						},
					},
				},
			},
			"variant": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: productVariantSchema,
					Blocks: map[string]schema.Block{
						"attribute": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: productVariantAttributeSchema,
							},
						},
						"price": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: productPriceDraftSchema,
								Blocks: map[string]schema.Block{
									"value": schema.SingleNestedBlock{
										Attributes: moneySchema,
									},
								},
							},
						},
					},
				},
				Description:         "The additional Product Variants for the Product.",
				MarkdownDescription: "The additional Product Variants for the Product.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.AlsoRequires(
						path.MatchRelative().AtParent().AtName("master_variant"),
					),
				},
			},
		},
	}
}

// Metadata implements resource.Resource.
func (*productResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_product"
}

// Create implements resource.Resource.
func (r *productResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Product
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	draft := plan.draft()

	var product *platform.Product
	err := retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
		var err error
		product, err = r.client.Products().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating product",
			err.Error(),
		)
		return
	}

	current := NewProductFromNative(product)

	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete implements resource.Resource.
func (r *productResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get the current state.
	var state Product
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := retry.RetryContext(
		ctx,
		5*time.Second,
		func() *retry.RetryError {
			_, err := r.client.Products().
				WithId(state.ID.ValueString()).
				Delete().
				Version(int(state.Version.ValueInt64())).
				Execute(ctx)

			return utils.ProcessRemoteError(err)
		})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting product",
			"Could not delete product, unexpected error: "+err.Error(),
		)
		return
	}
}

// Read implements resource.Resource.
func (r *productResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state.
	var state Product
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read remote product and check for errors.
	product, err := r.client.Products().WithId(state.ID.ValueString()).Get().Execute(ctx)
	if err != nil {
		if errors.Is(err, platform.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading product",
			"Could not retrieve the product, unexpected error: "+err.Error(),
		)
		return
	}

	// Transform the remote platform product to the
	// tf schema matching representation.
	current := NewProductFromNative(product)

	// Set current data as state.
	diags = resp.State.Set(ctx, &current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update implements resource.Resource.
func (r *productResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Product
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state Product
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := state.updateActions(plan)
	var product *platform.Product
	err := retry.RetryContext(ctx, 5*time.Second, func() *retry.RetryError {
		var err error
		product, err = r.client.Products().
			WithId(state.ID.ValueString()).
			Post(input).
			Execute(ctx)

		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating product",
			"Could not update product, unexpected error: "+err.Error(),
		)
		return
	}

	current := NewProductFromNative(product)

	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (r *productResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data := req.ProviderData.(*utils.ProviderData)
	r.client = data.Client
}

// ImportState implements resource.ResourceWithImportState.
func (r *productResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	product, err := r.client.Products().WithId(req.ID).Get().Execute(ctx)
	if err != nil {
		if errors.Is(err, platform.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading product",
			"Could not retrieve product, unexpected error: "+err.Error(),
		)
		return
	}

	current := NewProductFromNative(product)

	// Set refreshed state
	diags := resp.State.Set(ctx, &current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

var _ resource.ResourceWithModifyPlan = &productResource{}

func (r productResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() || req.Config.Raw.IsNull() {
		return
	}

	var state Product
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config Product
	diags = req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Master variant change
	if !state.MasterVariant.Sku.Equal(config.MasterVariant.Sku) {
		// Check if new Master persits within the Variants
		newMasterIndex := pie.FindFirstUsing(state.Variants, func(pv ProductVariant) bool { return pv.Sku.Equal(config.MasterVariant.Sku) })
		if newMasterIndex == -1 {
			resp.Diagnostics.AddError(
				"Master Variant must be within the Variants.",
				fmt.Sprintf("Variant with sku %s not found on product %s", config.MasterVariant.Sku, state.ID),
			)
			return
		} else {
			currentVariants := append(state.Variants, state.MasterVariant)
			newMaster := config.MasterVariant
			currentMasterVariant := getVariantBySku(currentVariants, config.MasterVariant.Sku.ValueString())
			newMaster.ID = currentMasterVariant.ID
			updatedPrices := []Price{}
			for _, p := range newMaster.Prices {
				currentPrice := getPriceByKey(currentMasterVariant.Prices, p.Key.ValueString())
				if currentPrice != nil {
					p.ID = currentPrice.ID
				} else {
					p.ID = types.StringUnknown()
				}
				updatedPrices = append(updatedPrices, p)
			}
			newMaster.Prices = updatedPrices
			diags = resp.Plan.SetAttribute(ctx, path.Root("master_variant"), newMaster)
			resp.Diagnostics.Append(diags...)
		}
	}

	// Variants change
	if !reflect.DeepEqual(state.Variants, config.Variants) {
		currentVariants := append(state.Variants, state.MasterVariant)
		newVariants := []ProductVariant{}
		for _, v := range config.Variants {
			currentVariant := getVariantBySku(currentVariants, v.Sku.ValueString())
			if currentVariant != nil {
				v.ID = currentVariant.ID
				updatedPrices := []Price{}
				for _, p := range v.Prices {
					currentPrice := getPriceByKey(currentVariant.Prices, p.Key.ValueString())
					if currentPrice != nil {
						p.ID = currentPrice.ID
					} else {
						p.ID = types.StringUnknown()
					}
					updatedPrices = append(updatedPrices, p)
				}
				v.Prices = updatedPrices
			} else {
				v.ID = types.Int64Unknown()
				updatedPrices := []Price{}
				for _, p := range v.Prices {
					p.ID = types.StringUnknown()
					updatedPrices = append(updatedPrices, p)
				}
				v.Prices = updatedPrices
			}
			newVariants = append(newVariants, v)
		}

		diags = resp.Plan.SetAttribute(ctx, path.Root("variant"),
			pie.SortUsing(newVariants, func(a, b ProductVariant) bool {
				if a.ID.IsUnknown() {
					return false
				}
				if b.ID.IsUnknown() {
					return true
				}
				return a.ID.ValueInt64() < b.ID.ValueInt64()
			}))
		resp.Diagnostics.Append(diags...)
	}
}

func getVariantBySku(variants []ProductVariant, sku string) *ProductVariant {
	for _, v := range variants {
		if v.Sku.ValueString() == sku {
			return &v
		}
	}
	return nil
}

func getPriceByKey(prices []Price, key string) *Price {
	for _, p := range prices {
		if p.Key.ValueString() == key {
			return &p
		}
	}
	return nil
}
