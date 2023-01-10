package state_transition

import (
	"context"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk_resource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_                 resource.Resource                = &stateTransitionResource{}
	_                 resource.ResourceWithConfigure   = &stateTransitionResource{}
	_                 resource.ResourceWithImportState = &stateTransitionResource{}
	globalUniqueStore map[string]bool
	globalUniqueMutex sync.Mutex
)

func init() {
	globalUniqueStore = make(map[string]bool, 0)
}

// NewOrderResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &stateTransitionResource{}
}

// orderResource is the resource implementation.
type stateTransitionResource struct {
	client *platform.ByProjectKeyRequestBuilder
	mutex  *utils.MutexKV
}

// Metadata returns the data source type name.
func (r *stateTransitionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_state_transitions"
}

// Schema defines the schema for the data source.
func (r *stateTransitionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Transitions are a way to describe possible transformations of the current state to other " +
			"states of the same type (for example: Initial -> Shipped). When performing a transitionState update " +
			"action and transitions is set, the currently referenced state must have a transition to the new state.\n" +
			"If transitions is an empty list, it means the current state is a final state and no further " +
			"transitions are allowed.\nIf transitions is not set, the validation is turned off. When " +
			"performing a transitionState update action, any other state of the same type can be transitioned to.\n\n" +
			"Note: Only one resource can be created for each state",
		Attributes: map[string]schema.Attribute{
			"from": schema.StringAttribute{
				Required: true,
			},
			"to": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (r *stateTransitionResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data := req.ProviderData.(*utils.ProviderData)
	r.client = data.Client
	r.mutex = data.Mutex
}

// Create creates the resource and sets the initial Terraform state.
func (r *stateTransitionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan StateTransition
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID := plan.From.ValueString()

	// Use a mutex since the state resource can modify the same resource in
	// commercetools
	r.mutex.Lock(resourceID)
	defer r.mutex.Unlock(resourceID)

	res, err := r.client.States().WithId(resourceID).Get().Execute(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading state",
			"Could not retrieve state, unexpected error: "+err.Error(),
		)
		return
	}
	current := NewStateTransitionFromNative(res)
	current.Version = types.Int64Value(int64(res.Version))

	input := current.updateActions(plan)
	err = sdk_resource.RetryContext(ctx, 5*time.Second, func() *sdk_resource.RetryError {
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

	current = NewStateTransitionFromNative(res)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *stateTransitionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state StateTransition
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID := state.From.ValueString()

	res, err := r.client.States().WithId(resourceID).Get().Execute(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading state",
			"Could not retrieve state, unexpected error: "+err.Error(),
		)
		return
	}

	current := NewStateTransitionFromNative(res)

	// Set refreshed state
	diags = resp.State.Set(ctx, &current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *stateTransitionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan StateTransition
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var state StateTransition
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID := plan.From.ValueString()

	// Use a mutex since the state resource can modify the same resource in
	// commercetools
	r.mutex.Lock(resourceID)
	defer r.mutex.Unlock(resourceID)

	res, err := r.client.States().WithId(resourceID).Get().Execute(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading state",
			"Could not retrieve state, unexpected error: "+err.Error(),
		)
		return
	}
	current := NewStateTransitionFromNative(res)
	current.Version = types.Int64Value(int64(res.Version))

	input := current.updateActions(plan)
	err = sdk_resource.RetryContext(ctx, 5*time.Second, func() *sdk_resource.RetryError {
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

	current = NewStateTransitionFromNative(res)

	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *stateTransitionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state StateTransition
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID := state.From.ValueString()

	// Use a mutex since the state resource can modify the same resource in
	// commercetools
	r.mutex.Lock(resourceID)
	defer r.mutex.Unlock(resourceID)

	res, err := r.client.States().WithId(resourceID).Get().Execute(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading state",
			"Could not retrieve state, unexpected error: "+err.Error(),
		)
		return
	}
	current := NewStateTransitionFromNative(res)
	current.Version = types.Int64Value(int64(res.Version))

	// Create new plan with empty `To` to generate an update action to remove
	// all state transitions from this resource
	plan := StateTransition{
		From:    current.From,
		To:      []types.String{},
		Version: current.Version,
	}

	input := current.updateActions(plan)
	err = sdk_resource.RetryContext(ctx, 5*time.Second, func() *sdk_resource.RetryError {
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
}

func (r *stateTransitionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
