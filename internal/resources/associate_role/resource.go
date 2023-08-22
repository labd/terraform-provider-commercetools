package associate_role

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

var (
	_ resource.Resource                = &associateRoleResource{}
	_ resource.ResourceWithConfigure   = &associateRoleResource{}
	_ resource.ResourceWithImportState = &associateRoleResource{}
)

type associateRoleResource struct {
	client *platform.ByProjectKeyRequestBuilder
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &associateRoleResource{}
}

// Schema implements resource.Resource.
func (*associateRoleResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Associate Roles provide a way to group granular Permissions and assign " +
			"them to Associates within a Business Unit.\n\n" +
			"See also the [Associate Role API Documentation](https://docs.commercetools.com/api/projects/associate-roles)",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the AssociateRole.",
				Computed:    true,
			},
			"version": schema.Int64Attribute{
				Description: "Current version of the AssociateRole.",
				Computed:    true,
			},
			"key": schema.StringAttribute{
				Description: "User-defined unique identifier of the AssociateRole.",
				Required:    true,
			},
			"buyer_assignable": schema.BoolAttribute{
				Description: "Whether the AssociateRole can be assigned to an Associate by a buyer. If false, " +
					"the AssociateRole can only be assigned using the general endpoint.",
				Optional: true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the AssociateRole.",
				Optional:    true,
			},
			"permissions": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of Permissions for the AssociateRole.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Metadata implements resource.Resource.
func (*associateRoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_associate_role"
}

// Create implements resource.Resource.
func (r *associateRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AssociateRole
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	draft := plan.draft()

	var associateRole *platform.AssociateRole
	err := retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
		var err error
		associateRole, err = r.client.AssociateRoles().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating associate role",
			err.Error(),
		)
		return
	}

	current := NewAssociateRoleFromNative(associateRole)

	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete implements resource.Resource.
func (r *associateRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get the current state.
	var state AssociateRole
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := retry.RetryContext(
		ctx,
		5*time.Second,
		func() *retry.RetryError {
			_, err := r.client.AssociateRoles().
				WithId(state.ID.ValueString()).
				Delete().
				Version(int(state.Version.ValueInt64())).
				Execute(ctx)

			return utils.ProcessRemoteError(err)
		})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting associate role",
			"Could not delete associate role, unexpected error: "+err.Error(),
		)
		return
	}
}

// Read implements resource.Resource.
func (r *associateRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state.
	var state AssociateRole
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read remote associate role and check for errors.
	associateRole, err := r.client.AssociateRoles().WithId(state.ID.ValueString()).Get().Execute(ctx)
	if err != nil {
		if errors.Is(err, platform.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading associate role",
			"Could not retrieve the associate role, unexpected error: "+err.Error(),
		)
		return
	}

	// Transform the remote platform associate role to the
	// tf schema matching representation.
	current := NewAssociateRoleFromNative(associateRole)

	// Set current data as state.
	diags = resp.State.Set(ctx, &current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update implements resource.Resource.
func (r *associateRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AssociateRole
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state AssociateRole
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := state.updateActions(plan)
	var associateRole *platform.AssociateRole
	err := retry.RetryContext(ctx, 5*time.Second, func() *retry.RetryError {
		var err error
		associateRole, err = r.client.AssociateRoles().
			WithId(state.ID.ValueString()).
			Post(input).
			Execute(ctx)

		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating associate role",
			"Could not update associate role, unexpected error: "+err.Error(),
		)
		return
	}

	current := NewAssociateRoleFromNative(associateRole)

	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure implements resource.ResourceWithConfigure.
func (r *associateRoleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	data := req.ProviderData.(*utils.ProviderData)
	r.client = data.Client
}

// ImportState implements resource.ResourceWithImportState.
func (*associateRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
