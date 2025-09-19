package project

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	sdkresource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/customvalidator"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &projectResource{}
	_ resource.ResourceWithConfigure   = &projectResource{}
	_ resource.ResourceWithImportState = &projectResource{}
)

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &projectResource{}
}

// projectResource is the resource implementation.
type projectResource struct {
	client *platform.ByProjectKeyRequestBuilder
}

// Metadata returns the data source type name.
func (r *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_settings"
}

// Schema defines the schema for the data source.
func (r *projectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The project endpoint provides a limited set of information about settings and configuration of " +
			"the project. Updating the settings is eventually consistent, it may take up to a minute before " +
			"a change becomes fully active.\n\n" +
			"See also the [Project Settings API Documentation](https://docs.commercetools.com/api/projects/project)",
		Version: 1,
		Attributes: map[string]schema.Attribute{
			// The ID is only here to make testing framework happy.
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique key of the project",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "The unique key of the project",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the project",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"currencies": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "A three-digit currency code as per [ISO 4217](https://en.wikipedia.org/wiki/ISO_4217)",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"countries": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "A two-digit country code as per [ISO 3166-1 alpha-2](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2)",
				Optional:            true,
				Computed:            true,
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
				MarkdownDescription: "Enable the Search Indexing of product projections",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"enable_search_index_product_search": schema.BoolAttribute{
				MarkdownDescription: "Enable the Search Indexing of products",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"enable_search_index_orders": schema.BoolAttribute{
				MarkdownDescription: "Enable the Search Indexing of orders",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"enable_search_index_customers": schema.BoolAttribute{
				MarkdownDescription: "Enable the Search Indexing of customers",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"enable_search_index_business_units": schema.BoolAttribute{
				MarkdownDescription: "Enable the Search Indexing of business  units",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"shipping_rate_input_type": schema.StringAttribute{
				MarkdownDescription: "Three ways to dynamically select a ShippingRatePriceTier exist. The CartValue type uses " +
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
			"messages": schema.ListNestedBlock{
				MarkdownDescription: "The change notifications subscribed to",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"enabled": schema.BoolAttribute{
							MarkdownDescription: "When true the creation of messages on the Messages Query HTTP API is enabled",
							Optional:            true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"delete_days_after_creation": schema.Int64Attribute{
							MarkdownDescription: "Specifies the number of days each Message should be available via the Messages Query API",
							Optional:            true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
							Validators: []validator.Int64{
								int64validator.Between(1, 90),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"carts": schema.ListNestedBlock{
				MarkdownDescription: "[Carts Configuration](https://docs.commercetools.com/api/projects/project#cartsconfiguration)",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"country_tax_rate_fallback_enabled": schema.BoolAttribute{
							MarkdownDescription: "Indicates if country - no state tax rate fallback should be used when a " +
								"shipping address state is not explicitly covered in the rates lists of all tax " +
								"categories of a cart line items",
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
						"delete_days_after_last_modification": schema.Int64Attribute{
							MarkdownDescription: "Number - Optional The default value for the " +
								"deleteDaysAfterLastModification parameter of the CartDraft. Initially set to 90 for " +
								"projects created after December 2019.",
							Optional: true,
							Computed: true,
							Default:  int64default.StaticInt64(90),
							Validators: []validator.Int64{
								int64validator.Between(1, 365250),
							},
						},
						"price_rounding_mode": schema.StringAttribute{
							MarkdownDescription: "Default value for the priceRoundingMode parameter of the CartDraft. " +
								"Indicates how the total prices on LineItems and CustomLineItems are rounded when " +
								"calculated.",
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(string(platform.RoundingModeHalfEven)),
						},
						"tax_rounding_mode": schema.StringAttribute{
							MarkdownDescription: "Default value for the taxRoundingMode parameter of the CartDraft. " +
								"Indicates how monetary values are rounded when calculating taxes for taxedPrice.",
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(string(platform.RoundingModeHalfEven)),
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"shopping_lists": schema.ListNestedBlock{
				MarkdownDescription: "[Shopping List Configuration](https://docs.commercetools.com/api/projects/project#ctp:api:type:ShoppingListsConfiguration)",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"delete_days_after_last_modification": schema.Int64Attribute{
							MarkdownDescription: "Number - Optional The default value for the " +
								"deleteDaysAfterLastModification parameter of the CartDraft. Initially set to 90 for " +
								"projects created after December 2019.",
							Optional: true,
							Computed: true,
							Default:  int64default.StaticInt64(360),
							Validators: []validator.Int64{
								int64validator.Between(1, 365250),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"shipping_rate_cart_classification_value": schema.ListNestedBlock{
				MarkdownDescription: "If shipping_rate_input_type is set to CartClassification these values are used to create " +
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
							Required: true,
						},
						"label": customtypes.LocalizedString(customtypes.LocalizedStringOpts{
							Optional: true,
						}),
					},
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
							MarkdownDescription: "Partially hidden on retrieval",
							Optional:            true,
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
			"business_units": schema.ListNestedBlock{
				MarkdownDescription: "Holds configuration specific to [Business Units](https://docs.commercetools.com/api/projects/business-units#ctp:api:type:BusinessUnit).",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"my_business_unit_status_on_creation": schema.StringAttribute{
							MarkdownDescription: "Status of Business Units created using the My Business Unit endpoint.",
							Computed:            true,
							Optional:            true,
							Default:             stringdefault.StaticString(string(platform.BusinessUnitConfigurationStatusInactive)),
							Validators: []validator.String{
								stringvalidator.OneOf(
									string(platform.BusinessUnitConfigurationStatusActive),
									string(platform.BusinessUnitConfigurationStatusInactive),
								),
							},
						},
						"my_business_unit_associate_role_key_on_creation": schema.StringAttribute{
							MarkdownDescription: "Default Associate Role assigned to the Associate creating a " +
								"Business Unit using the My Business Unit endpoint. Note that this field cannot be " +
								"unset once assigned!",
							Optional: true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (r *projectResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data := req.ProviderData.(*utils.ProviderData)
	r.client = data.Client
}

func (r *projectResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			StateUpgrader: upgradeStateV0,
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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

	input, err := current.updateActions(plan)
	if err != nil {
		resp.Diagnostics.AddError("Error updating project", err.Error())
		return
	}

	var res *platform.Project
	err = sdkresource.RetryContext(ctx, 5*time.Second, func() *sdkresource.RetryError {
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
func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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

	input, err := state.updateActions(plan)
	if err != nil {
		resp.Diagnostics.AddError("Error updating project", err.Error())
		return
	}

	var res *platform.Project
	err = sdkresource.RetryContext(ctx, 5*time.Second, func() *sdkresource.RetryError {
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

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state Project
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
