package project

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk_resource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/customvalidator"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &ProjectResource{}
	_ resource.ResourceWithConfigure   = &ProjectResource{}
	_ resource.ResourceWithImportState = &ProjectResource{}
)

// NewOrderResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &ProjectResource{}
}

// orderResource is the resource implementation.
type ProjectResource struct {
	client *platform.ByProjectKeyRequestBuilder
}

// Metadata returns the data source type name.
func (r *ProjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_settings"
}

// Schema defines the schema for the data source.
func (r *ProjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The project endpoint provides a limited set of information about settings and configuration of " +
			"the project. Updating the settings is eventually consistent, it may take up to a minute before " +
			"a change becomes fully active.\n\n" +
			"See also the [Project Settings API Documentation](https://docs.commercetools.com/api/projects/project)",
		Version: 1,
		Attributes: map[string]schema.Attribute{
			// The ID is only here to make testing framework happy.
			"id": schema.StringAttribute{
				Description: "The unique key of the project",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Description: "The unique key of the project",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the project",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"currencies": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "A three-digit currency code as per [ISO 4217](https://en.wikipedia.org/wiki/ISO_4217)",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"countries": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "A two-digit country code as per [ISO 3166-1 alpha-2](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2)",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"languages": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "[IETF Language Tag](https://en.wikipedia.org/wiki/IETF_language_tag)",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"enable_search_index_products": schema.BoolAttribute{
				Description: "Enable the Search Indexing of products",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"enable_search_index_orders": schema.BoolAttribute{
				Description: "Enable the Search Indexing of orders",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"shipping_rate_input_type": schema.StringAttribute{
				Description: "Three ways to dynamically select a ShippingRatePriceTier exist. The CartValue type uses " +
					"the sum of all line item prices, whereas CartClassification and CartScore use the " +
					"shippingRateInput field on the cart to select a tier",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"CartClassification",
						"CartScore",
						"CartValue",
					),
					customvalidator.DependencyValidator(
						"CartClassification",
						path.MatchRoot("shipping_rate_cart_classification_value").AtAnyListIndex(),
					),
				},
			},
			"version": schema.Int64Attribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"carts": schema.ListNestedBlock{
				MarkdownDescription: "[Carts Configuration](https://docs.commercetools.com/api/projects/project#carts-configuration)",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"country_tax_rate_fallback_enabled": schema.BoolAttribute{
							Description: "Indicates if country - no state tax rate fallback should be used when a " +
								"shipping address state is not explicitly covered in the rates lists of all tax " +
								"categories of a cart line items",
							Optional: true,
						},
						"delete_days_after_last_modification": schema.Int64Attribute{
							Description: "Number - Optional The default value for the " +
								"deleteDaysAfterLastModification parameter of the CartDraft. Initially set to 90 for " +
								"projects created after December 2019.",
							Optional: true,
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"messages": schema.ListNestedBlock{
				Description: "The change notifications subscribed to",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"enabled": schema.BoolAttribute{
							Description: "When true the creation of messages on the Messages Query HTTP API is enabled",
							Optional:    true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"delete_days_after_creation": schema.Int64Attribute{
							Description: "Specifies the number of days each Message should be available via the Messages Query API",
							Optional:    true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"external_oauth": schema.ListNestedBlock{
				MarkdownDescription: "[External OAUTH](https://docs.commercetools.com/api/projects/project#externaloauth)",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"url": schema.StringAttribute{
							Optional:  true,
							Sensitive: true,
							Validators: []validator.String{
								stringvalidator.AlsoRequires(
									path.MatchRelative().AtParent().AtName("authorization_header"),
								),
							},
						},
						"authorization_header": schema.StringAttribute{
							Description: "Partially hidden on retrieval",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.AlsoRequires(
									path.MatchRelative().AtParent().AtName("url"),
								),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"shipping_rate_cart_classification_value": schema.ListNestedBlock{
				Description: "If shipping_rate_input_type is set to CartClassification these values are used to create " +
					"tiers\n. Only a key defined inside the values array can be used to create a tier, or to set a value " +
					"for the shippingRateInput on the cart. The keys are checked for uniqueness and the request is " +
					"rejected if keys are not unique",
				Validators: []validator.List{
					customvalidator.RequireValueValidator(
						"CartClassification",
						path.MatchRoot("shipping_rate_input_type"),
					),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							MarkdownDescription: "[Resource Type ID](https://docs.commercetools.com/api/projects/Projects#changeProject)",
							Required:            true,
						},
						"label": customtypes.LocalizedString(customtypes.LocalizedStringOpts{
							Optional: true,
						}),
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (r *ProjectResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data := req.ProviderData.(utils.ProviderData)
	r.client = data.Client
}

func (p *ProjectResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			StateUpgrader: upgradeStateV0,
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan Project
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.client.Get().Execute(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading project",
			"Could not retrieve project, unexpected error: "+err.Error(),
		)
		return
	}
	current := NewProjectFromNative(project)

	input := current.updateActions(plan)
	var res *platform.Project
	err = sdk_resource.RetryContext(ctx, 5*time.Second, func() *sdk_resource.RetryError {
		var err error
		res, err = r.client.Post(input).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating project", err.Error())
		return
	}

	result := NewProjectFromNative(res)
	result.setStateData(plan)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state Project
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.Get().Execute(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading project",
			"Could not retrieve project, unexpected error: "+err.Error(),
		)
		return
	}
	current := NewProjectFromNative(res)
	current.setStateData(state)

	// Set refreshed state
	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan Project
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var state Project
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := state.updateActions(plan)

	var res *platform.Project
	err := sdk_resource.RetryContext(ctx, 5*time.Second, func() *sdk_resource.RetryError {
		var err error
		res, err = r.client.Post(input).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating project", err.Error())
		return
	}
	result := NewProjectFromNative(res)
	result.setStateData(plan)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state Project
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
