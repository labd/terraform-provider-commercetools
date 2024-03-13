package tax_category_v2

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
	"time"
)

var (
	_ resource.Resource                = &taxCategoryV2Resource{}
	_ resource.ResourceWithConfigure   = &taxCategoryV2Resource{}
	_ resource.ResourceWithImportState = &taxCategoryV2Resource{}
)

type taxCategoryV2Resource struct {
	client *platform.ByProjectKeyRequestBuilder
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &taxCategoryV2Resource{}
}

// Schema implements resource.Resource.
func (*taxCategoryV2Resource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Tax Categories define how products are to be taxed in different countries.\n\n" +
			"See also the [Tax Category API Documentation](https://docs.commercetools.com/api/projects/taxCategories).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the TaxCategory.",
				Computed:    true,
			},
			"version": schema.Int64Attribute{
				Description: "Current version of the TaxCategory.",
				Computed:    true,
			},
			"key": schema.StringAttribute{
				Description: "User-defined unique identifier of the TaxCategory",
				Optional:    true,
				//TODO: add key validation
			},
			"name": schema.StringAttribute{
				Description: "Name of the TaxCategory.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the TaxCategory.",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"tax_rate": schema.ListNestedBlock{
				MarkdownDescription: "Attributes with unique values.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Present if the TaxRate is part of a TaxCategory. Absent for external TaxRates in LineItem, CustomLineItem, and ShippingInfo.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "User-defined unique identifier of the TaxCategory",
							Optional:    true,
							//TODO: add key validation
						},
						"name": schema.StringAttribute{
							Description: "Name of the TaxRate.",
							Required:    true,
						},
						"amount": schema.Float64Attribute{
							Description: "Number Percentage in the range of [0..1]. The sum of the amounts of all subRates, " +
								"if there are any",
							Required: true,
							//ValidateFunc: validateTaxRateAmount, //TODO add ValidateTaxRateAmount
						},
						"included_in_price": schema.BoolAttribute{
							Description: "If true, tax is included in Embedded Prices or Standalone Prices, and the taxedPrice is present on LineItems. " +
								"In this case, the totalNet price on TaxedPrice includes the TaxRate.",
							Required: true,
						},
						"country": schema.StringAttribute{
							Description: "A two-digit country code as per [ISO 3166-1 alpha-2](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2)",
							Required:    true,
						},
						"state": schema.StringAttribute{
							Description: "The state in the country",
							Optional:    true,
						},
					},
					Blocks: map[string]schema.Block{
						"sub_rate": schema.ListNestedBlock{
							Description: "For countries (for example the US) where the total tax is a combination of multiple " +
								"taxes (for example state and local taxes)",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "Name of the SubRate.",
										Required:    true,
									},
									"amount": schema.Float64Attribute{
										Description: "Number Percentage in the range of [0..1]. The sum of the amounts of all subRates, " +
											"if there are any",
										Required: true,
										//ValidateFunc: validateTaxRateAmount, //TODO add ValidateTaxRateAmount
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (*taxCategoryV2Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tax_category_v2"
}

func (r *taxCategoryV2Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TaxCategory
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var draft platform.TaxCategoryDraft
	draft = plan.draft()
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}

	var taxCategory *platform.TaxCategory
	err := retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
		var err error
		taxCategory, err = r.client.TaxCategories().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating tax category",
			err.Error(),
		)
		return
	}

	current, err := TaxCategoryFromNative(taxCategory)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating tax category from native",
			err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *taxCategoryV2Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get the current state.
	var state TaxCategory
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := retry.RetryContext(
		ctx,
		5*time.Second,
		func() *retry.RetryError {
			_, err := r.client.TaxCategories().
				WithId(state.ID.ValueString()).
				Delete().
				Version(int(state.Version.ValueInt64())).
				Execute(ctx)

			return utils.ProcessRemoteError(err)
		})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting tax category",
			"Could not delete tax category, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *taxCategoryV2Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state.
	var state TaxCategory
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read remote tax category and check for errors.
	taxCategory, err := r.client.TaxCategories().WithId(state.ID.ValueString()).Get().Execute(ctx)
	if err != nil {
		if errors.Is(err, platform.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading tax category",
			"Could not retrieve the tax category, unexpected error: "+err.Error(),
		)
		return
	}

	// Transform the remote platform tax category to the
	// tf schema matching representation.
	current, err := TaxCategoryFromNative(taxCategory)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating tax category from native",
			err.Error(),
		)
		return
	}

	// Set current data as state.
	diags = resp.State.Set(ctx, &current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *taxCategoryV2Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TaxCategory
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state TaxCategory
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := state.updateActions(plan)
	var taxCategory *platform.TaxCategory
	err := retry.RetryContext(ctx, 5*time.Second, func() *retry.RetryError {
		var err error
		taxCategory, err = r.client.TaxCategories().
			WithId(state.ID.ValueString()).
			Post(input).
			Execute(ctx)

		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating tax category",
			"Could not update tax category, unexpected error: "+err.Error(),
		)
		return
	}

	current, err := TaxCategoryFromNative(taxCategory)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating tax category from native",
			err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *taxCategoryV2Resource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	data := req.ProviderData.(*utils.ProviderData)
	r.client = data.Client
}

func (*taxCategoryV2Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
