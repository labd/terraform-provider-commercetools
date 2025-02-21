package associate_role

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/labd/terraform-provider-commercetools/commercetools"
	"github.com/labd/terraform-provider-commercetools/internal/sharedtypes"
	"regexp"
	"sort"
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
		MarkdownDescription: "Associate Roles provide a way to group granular Permissions and assign " +
			"them to Associates within a Business Unit.\n\n" +
			"See also the [Associate Role API Documentation](https://docs.commercetools.com/api/projects/associate-roles)",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the associate role.",
				Computed:            true,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "Current version of the associate role.",
				Computed:            true,
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "User-defined unique identifier of the associate role.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 256),
					stringvalidator.RegexMatches(
						regexp.MustCompile("^[A-Za-z0-9_-]+$"),
						"Key must match pattern ^[A-Za-z0-9_-]+$",
					),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the associate role.",
				Optional:            true,
			},
			"buyer_assignable": schema.BoolAttribute{
				MarkdownDescription: "Whether the associate role can be assigned to an associate by a buyer. If false, " +
					"the associate role can only be assigned using the general endpoint. Defaults to true.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"permissions": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				MarkdownDescription: "List of permissions for the associate role. See the [Associate Role API " +
					"Documentation](https://docs.commercetools.com/api/projects/associate-roles#ctp:api:type:Permission) " +
					"for more information.",
				Validators: []validator.List{
					listvalidator.UniqueValues(),
					listvalidator.ValueStringsAre(
						stringvalidator.OneOf(
							string(platform.PermissionAddChildUnits),
							string(platform.PermissionUpdateAssociates),
							string(platform.PermissionUpdateBusinessUnitDetails),
							string(platform.PermissionUpdateParentUnit),
							string(platform.PermissionViewMyCarts),
							string(platform.PermissionViewOthersCarts),
							string(platform.PermissionUpdateMyCarts),
							string(platform.PermissionUpdateOthersCarts),
							string(platform.PermissionCreateMyCarts),
							string(platform.PermissionCreateOthersCarts),
							string(platform.PermissionDeleteMyCarts),
							string(platform.PermissionDeleteOthersCarts),
							string(platform.PermissionViewMyOrders),
							string(platform.PermissionViewOthersOrders),
							string(platform.PermissionUpdateMyOrders),
							string(platform.PermissionUpdateOthersOrders),
							string(platform.PermissionCreateMyOrdersFromMyCarts),
							string(platform.PermissionCreateMyOrdersFromMyQuotes),
							string(platform.PermissionCreateOrdersFromOthersCarts),
							string(platform.PermissionCreateOrdersFromOthersQuotes),
							string(platform.PermissionViewMyQuotes),
							string(platform.PermissionViewOthersQuotes),
							string(platform.PermissionAcceptMyQuotes),
							string(platform.PermissionAcceptOthersQuotes),
							string(platform.PermissionDeclineMyQuotes),
							string(platform.PermissionDeclineOthersQuotes),
							string(platform.PermissionRenegotiateMyQuotes),
							string(platform.PermissionRenegotiateOthersQuotes),
							string(platform.PermissionReassignMyQuotes),
							string(platform.PermissionReassignOthersQuotes),
							string(platform.PermissionViewMyQuoteRequests),
							string(platform.PermissionViewOthersQuoteRequests),
							string(platform.PermissionUpdateMyQuoteRequests),
							string(platform.PermissionUpdateOthersQuoteRequests),
							string(platform.PermissionCreateMyQuoteRequestsFromMyCarts),
							string(platform.PermissionCreateQuoteRequestsFromOthersCarts),
							string(platform.PermissionCreateApprovalRules),
							string(platform.PermissionUpdateApprovalRules),
							string(platform.PermissionUpdateApprovalFlows),
						),
					),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"custom": sharedtypes.CustomSchema,
		},
	}
}

// Metadata implements resource.Resource.
func (*associateRoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_associate_role"
}

func resortPermissions(permissions, plan []platform.Permission) []platform.Permission {

	// Build a map which maps the planned permissions to its index from the array
	indexMap := make(map[platform.Permission]int)
	for idx, p := range plan {
		indexMap[p] = idx
	}

	// Build a map of the current permissions to the planned index
	targetMap := make(map[platform.Permission]int)
	for _, p := range permissions {
		idx, ok := indexMap[p]
		if ok {
			targetMap[p] = idx
		}
	}

	// Sort the target permission list by the index from the map
	targetList := make([]platform.Permission, 0, len(targetMap))
	for key := range targetMap {
		targetList = append(targetList, key)
	}
	sort.SliceStable(targetList, func(i, j int) bool {
		return targetMap[targetList[i]] < targetMap[targetList[j]]
	})

	return targetList
}

// Create implements resource.Resource.
func (r *associateRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AssociateRole
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var customType *platform.Type
	var err error
	if plan.Custom.IsSet() {
		customType, err = commercetools.GetTypeResource(ctx, commercetools.CreateTypeFetcher(r.client), *plan.Custom.TypeID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error getting custom type",
				"Could not get custom type, unexpected error: "+err.Error(),
			)
			return
		}
	}

	draft, err := plan.draft(customType)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating associate role",
			"Could not create associate role, unexpected error: "+err.Error(),
		)
		return
	}

	var associateRole *platform.AssociateRole
	err = retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
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

	associateRole.Permissions = resortPermissions(associateRole.Permissions, draft.Permissions)

	current, err := NewAssociateRoleFromNative(associateRole)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating associate role",
			"Could not create associate role, unexpected error: "+err.Error(),
		)
		return
	}

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
		if utils.IsResourceNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading associate role",
			"Could not retrieve the associate role, unexpected error: "+err.Error(),
		)
		return
	}

	var customType *platform.Type
	if state.Custom.IsSet() {
		customType, err = commercetools.GetTypeResource(ctx, commercetools.CreateTypeFetcher(r.client), *state.Custom.TypeID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error getting custom type",
				"Could not get custom type, unexpected error: "+err.Error(),
			)
			return
		}
	}

	sDraft, err := state.draft(customType)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading associate role",
			"Could not retrieve the associate role, unexpected error: "+err.Error(),
		)
		return
	}
	associateRole.Permissions = resortPermissions(associateRole.Permissions, sDraft.Permissions)

	// Transform the remote platform associate role to the
	// tf schema matching representation.
	current, err := NewAssociateRoleFromNative(associateRole)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading associate role",
			"Could not retrieve the associate role, unexpected error: "+err.Error(),
		)
		return
	}

	// Set current data as state.
	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update implements resource.Resource.
func (r *associateRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var err error
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

	var customType *platform.Type
	if plan.Custom.IsSet() {
		customType, err = commercetools.GetTypeResource(ctx, commercetools.CreateTypeFetcher(r.client), *plan.Custom.TypeID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error getting custom type",
				"Could not get custom type, unexpected error: "+err.Error(),
			)
			return
		}
	}

	input, err := state.updateActions(customType, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating associate role",
			"Could not update associate role, unexpected error: "+err.Error(),
		)
		return
	}

	var associateRole *platform.AssociateRole
	err = retry.RetryContext(ctx, 5*time.Second, func() *retry.RetryError {
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

	sDraft, err := plan.draft(customType)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading associate role",
			"Could not retrieve the associate role, unexpected error: "+err.Error(),
		)
		return
	}
	associateRole.Permissions = resortPermissions(associateRole.Permissions, sDraft.Permissions)

	current, err := NewAssociateRoleFromNative(associateRole)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating associate role",
			"Could not update associate role, unexpected error: "+err.Error(),
		)
		return
	}

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
