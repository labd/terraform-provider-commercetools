package product_selection

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

var (
	_ resource.Resource                = &productSelectionResource{}
	_ resource.ResourceWithConfigure   = &productSelectionResource{}
	_ resource.ResourceWithImportState = &productSelectionResource{}
)

type productSelectionResource struct {
	client *platform.ByProjectKeyRequestBuilder
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &productSelectionResource{}
}

// Schema implements resource.Resource.
func (*productSelectionResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Product Selections can be used to manage individual assortments for different sales channels." +
			"See also the [Product Selections API Documentation](https://docs.commercetools.com/api/projects/product-selections)",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the ProductSelection.",
				Computed:    true,
			},
			"version": schema.Int64Attribute{
				Description: "Current version of the ProductSelection.",
				Computed:    true,
			},
			"key": schema.StringAttribute{
				Description: "User-defined unique identifier of the ProductSelection.",
				Optional:    true,
			},
			"mode": schema.StringAttribute{
				Description: "Specifies in which way the Products are assigned to the ProductSelection." +
					"Currently, the only way of doing this is to specify each Product individually, either by including or excluding them explicitly." +
					"Default: Individual",
				Default:  stringdefault.StaticString(string(platform.ProductSelectionModeIndividual)),
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(platform.ProductSelectionModeIndividual),
						string(platform.ProductSelectionModeIndividualExclusion),
					),
				},
			},
			"name": schema.MapAttribute{
				CustomType:          customtypes.NewLocalizedStringType(),
				Description:         "Name of the ProductSelection.",
				MarkdownDescription: "Name of the ProductSelection.",
				Required:            true,
			},
		},
	}
}

// Metadata implements resource.Resource.
func (*productSelectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_product_selection"
}

// Create implements resource.Resource.
func (r *productSelectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProductSelection
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	draft := plan.draft()

	var productSelection *platform.ProductSelection
	err := retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
		var err error
		productSelection, err = r.client.ProductSelections().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating product selection",
			err.Error(),
		)
		return
	}

	current := NewProductSelectionFromNative(productSelection)

	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete implements resource.Resource.
func (r *productSelectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get the current state.
	var state ProductSelection
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := retry.RetryContext(
		ctx,
		5*time.Second,
		func() *retry.RetryError {
			_, err := r.client.ProductSelections().
				WithId(state.ID.ValueString()).
				Delete().
				Version(int(state.Version.ValueInt64())).
				Execute(ctx)

			return utils.ProcessRemoteError(err)
		})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting product selection",
			"Could not delete product selection, unexpected error: "+err.Error(),
		)
		return
	}
}

// Read implements resource.Resource.
func (r *productSelectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state.
	var state ProductSelection
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read remote product selection and check for errors.
	productSelection, err := r.client.ProductSelections().WithId(state.ID.ValueString()).Get().Execute(ctx)
	if err != nil {
		if errors.Is(err, platform.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading product selection",
			"Could not retrieve the product selection, unexpected error: "+err.Error(),
		)
		return
	}

	// Transform the remote platform product selection to the
	// tf schema matching representation.
	current := NewProductSelectionFromNative(productSelection)

	// Set current data as state.
	diags = resp.State.Set(ctx, &current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update implements resource.Resource.
func (r *productSelectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProductSelection
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ProductSelection
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := state.updateActions(plan)
	var productSelection *platform.ProductSelection
	err := retry.RetryContext(ctx, 5*time.Second, func() *retry.RetryError {
		var err error
		productSelection, err = r.client.ProductSelections().
			WithId(state.ID.ValueString()).
			Post(input).
			Execute(ctx)

		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating product selection",
			"Could not update product selection, unexpected error: "+err.Error(),
		)
		return
	}

	current := NewProductSelectionFromNative(productSelection)

	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure implements resource.ResourceWithConfigure.
func (r *productSelectionResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	data := req.ProviderData.(*utils.ProviderData)
	r.client = data.Client
}

// ImportState implements resource.ResourceWithImportState.
func (*productSelectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
