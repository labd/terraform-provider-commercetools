package state

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk_resource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &stateResource{}
	_ resource.ResourceWithConfigure   = &stateResource{}
	_ resource.ResourceWithImportState = &stateResource{}
)

// NewStateResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &stateResource{}
}

// stateResource is the resource implementation.
type stateResource struct {
	client *platform.ByProjectKeyRequestBuilder
	mutex  *utils.MutexKV
}

// Metadata returns the data source type name.
func (r *stateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_state"
}

// Schema defines the schema for the data source.
func (r *stateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The commercetools platform allows you to model states of certain objects, such as orders, line " +
			"items, products, reviews, and payments to define finite state machines reflecting the business " +
			"logic you'd like to implement.\n\n" +
			"See also the [State API Documentation](https://docs.commercetools.com/api/projects/states)",
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"key": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the order.",
				Optional:    true,
			},
			"name": schema.MapAttribute{
				CustomType:          customtypes.NewLocalizedStringType(),
				Description:         "Name of the State as localized string.",
				MarkdownDescription: "Name of the State as localized string.",
				Optional:            true,
			},
			"description": schema.MapAttribute{
				CustomType:          customtypes.NewLocalizedStringType(),
				Description:         "Description of the State as localized string.",
				MarkdownDescription: "Description of the State as localized string.",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				Description:         "Specify to which resource or object type the State is assigned to.",
				MarkdownDescription: "[StateType](https://docs.commercetools.com/api/projects/states#statetype)",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(platform.StateTypeEnumOrderState),
						string(platform.StateTypeEnumLineItemState),
						string(platform.StateTypeEnumProductState),
						string(platform.StateTypeEnumReviewState),
						string(platform.StateTypeEnumPaymentState),
						string(platform.StateTypeEnumQuoteRequestState),
						string(platform.StateTypeEnumStagedQuoteState),
						string(platform.StateTypeEnumQuoteState),
					),
				},
			},
			"initial": schema.BoolAttribute{
				Description: "A state can be declared as an initial state for any state machine. When a workflow " +
					"starts, this first state must be an initial state",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"roles": schema.ListAttribute{
				MarkdownDescription: `[State Role](https://docs.commercetools.com/api/projects/states#staterole)`,
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(
						stringvalidator.OneOf(
							string(platform.StateRoleEnumReviewIncludedInStatistics),
							string(platform.StateRoleEnumReturn),
						),
					),
				},
			},
			"version": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (r *stateResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data := req.ProviderData.(*utils.ProviderData)
	r.client = data.Client
	r.mutex = data.Mutex
}

// Create creates the resource and sets the initial Terraform state.
func (r *stateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan State
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	draft := plan.draft()
	var res *platform.State
	err := sdk_resource.RetryContext(ctx, 20*time.Second, func() *sdk_resource.RetryError {
		var err error
		res, err = r.client.States().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating state",
			err.Error(),
		)
		return
	}

	current := NewStateFromNative(res)
	current.matchDefaults(plan)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *stateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state State
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.States().WithId(state.ID.ValueString()).Get().Execute(ctx)
	if err != nil {
		if errors.Is(err, platform.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading state",
			"Could not retrieve state, unexpected error: "+err.Error(),
		)
		return
	}

	current := NewStateFromNative(res)
	current.matchDefaults(state)

	// Set refreshed state
	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *stateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan State
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var state State
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID := state.ID.ValueString()

	// Use a mutex since the state_transitions resource can modify the same
	// resource in commercetools
	r.mutex.Lock(resourceID)
	defer r.mutex.Unlock(resourceID)

	// Retrieve the last version. This is needed since the state_transition
	// resource can also modify the state version in commercetools
	version, diags := r.GetVersion(ctx, resourceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Version = types.Int64Value(int64(version))
	state.setDefaults()

	input := state.updateActions(plan)
	var res *platform.State
	err := sdk_resource.RetryContext(ctx, 5*time.Second, func() *sdk_resource.RetryError {
		var err error
		res, err = r.client.States().WithId(resourceID).Post(input).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating state",
			"Could not update state, unexpected error: "+err.Error(),
		)
		return
	}

	current := NewStateFromNative(res)
	current.matchDefaults(plan)

	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *stateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state State
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resourceID := state.ID.ValueString()

	// Use a mutex since the state_transitions resource can modify the same
	// resource in commercetools
	r.mutex.Lock(resourceID)
	defer r.mutex.Unlock(resourceID)

	// Retrieve the last version. This is needed since the state_transition
	// resource can also modify the state version in commercetools
	version, diags := r.GetVersion(ctx, resourceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := sdk_resource.RetryContext(ctx, 5*time.Second, func() *sdk_resource.RetryError {
		_, err := r.client.States().WithId(resourceID).Delete().Version(version).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting state",
			"Could not delete state, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *stateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *stateResource) GetVersion(ctx context.Context, resourceID string) (int, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	res, err := r.client.States().WithId(resourceID).Get().Execute(ctx)
	if err != nil {
		diags.AddError(
			"Error reading state",
			"Could not retrieve state, unexpected error: "+err.Error(),
		)
		return 0, diags
	}
	return res.Version, diags
}
