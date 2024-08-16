package business_unit_division

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/labd/terraform-provider-commercetools/internal/sharedtypes"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	_ resource.Resource                = &divisionResource{}
	_ resource.ResourceWithConfigure   = &divisionResource{}
	_ resource.ResourceWithImportState = &divisionResource{}
)

type divisionResource struct {
	client *platform.ByProjectKeyRequestBuilder
}

// NewDivisionResource creates a new resource for the Division type.
func NewDivisionResource() resource.Resource {
	return &divisionResource{}
}

// Schema implements resource.Resource.
func (b *divisionResource) Schema(_ context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schema.Schema{
		MarkdownDescription: "business unit type to represent the top level of a business. Contains specific fields and values that differentiate a Division from the generic BusinessUnit.\n\n" +
			"See also the [business unit API Documentation](https://docs.commercetools.com/api/projects/business-units",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the division.",
				Computed:            true,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "The current version of the division.",
				Computed:            true,
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "User-defined unique key for the division. Must be unique within the project. " +
					"Updating this value is not supported.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 256),
					stringvalidator.RegexMatches(
						regexp.MustCompile("^[A-Za-z0-9_-]+$"),
						"Key must match pattern ^[A-Za-z0-9_-]+$",
					),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Indicates whether the business unit can be edited and used in [Orders](https://docs.commercetools.com/api/projects/orders). Defaults to `Active`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(string(platform.BusinessUnitStatusActive)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(platform.BusinessUnitStatusActive),
						string(platform.BusinessUnitStatusInactive),
					),
				},
			},
			"store_mode": schema.StringAttribute{
				MarkdownDescription: "Defines whether the Stores of the business unit are set directly on the business unit or are inherited from a parent. Defaults to `FromParent`",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(string(platform.BusinessUnitStoreModeFromParent)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(platform.BusinessUnitStoreModeFromParent),
						string(platform.BusinessUnitStoreModeExplicit),
					),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the division.",
				Required:            true,
			},
			"contact_email": schema.StringAttribute{
				MarkdownDescription: "The email address of the division.",
				Optional:            true,
			},
			"associate_mode": schema.StringAttribute{
				MarkdownDescription: "Determines whether the business unit can inherit Associates from a parent. Defaults to `ExplicitAndFromParent`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(string(platform.BusinessUnitAssociateModeExplicitAndFromParent)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(platform.BusinessUnitAssociateModeExplicitAndFromParent),
						string(platform.BusinessUnitAssociateModeExplicit),
					),
				},
			},
			"approval_rule_mode": schema.StringAttribute{
				MarkdownDescription: "Determines whether the business unit can inherit Approval Rules from a parent. Defaults to `ExplicitAndFromParent`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(string(platform.BusinessUnitApprovalRuleModeExplicitAndFromParent)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(platform.BusinessUnitApprovalRuleModeExplicit),
						string(platform.BusinessUnitApprovalRuleModeExplicitAndFromParent),
					),
				},
			},
			"shipping_address_keys": schema.ListAttribute{
				MarkdownDescription: "List of the shipping addresses used by the division.",
				Optional:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"default_shipping_address_key": schema.StringAttribute{
				MarkdownDescription: "Key of the default shipping Address.",
				Optional:            true,
			},
			"billing_address_keys": schema.ListAttribute{
				MarkdownDescription: "List of the billing addresses used by the division.",
				Optional:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"default_billing_address_key": schema.StringAttribute{
				MarkdownDescription: "Key of the default billing Address.",
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"store":   sharedtypes.StoreKeyReferenceBlockSchema,
			"address": sharedtypes.AddressBlockSchema,
			"parent_unit": schema.SingleNestedBlock{
				MarkdownDescription: "Reference to a parent business unit by its key or id. One of either is required.",
				Validators: []validator.Object{
					objectvalidator.AtLeastOneOf(
						path.MatchRelative().AtName("key"),
						path.MatchRelative().AtName("id"),
					),
				},
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						MarkdownDescription: "User-defined unique identifier of the business unit",
						Optional:            true,
					},
					"key": schema.StringAttribute{
						MarkdownDescription: "User-defined unique key of the business unit",
						Optional:            true,
					},
				},
			},
		},
	}
}

// Metadata implements resource.Resource.
func (b *divisionResource) Metadata(_ context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_business_unit_division"
}

// ImportState implements resource.ResourceWithImportState.
func (b *divisionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, res *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, res)
}

// Configure implements resource.ResourceWithConfigure.
func (b *divisionResource) Configure(ctx context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	data, ok := req.ProviderData.(*utils.ProviderData)
	if !ok {
		return
	}

	b.client = data.Client
}

// Create implements resource.Resource.
func (b *divisionResource) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	var plan Division
	diags := req.Plan.Get(ctx, &plan)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	draft, err := plan.draft()
	if err != nil {
		res.Diagnostics.AddError(
			"Error creating business unit",
			"Could not create business unit, unexpected error: "+err.Error(),
		)
		return

	}

	var bu *platform.BusinessUnit
	err = retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
		var err error
		bu, err = b.client.BusinessUnits().Post(draft).Execute(ctx)

		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		res.Diagnostics.AddError(
			"Error creating business unit",
			"Could not create business unit, unexpected error: "+err.Error(),
		)
		return
	}

	current, err := NewDivisionFromNative(bu)
	if err != nil {
		res.Diagnostics.AddError(
			"Error mapping business unit",
			"Could not create business unit, unexpected error: "+err.Error(),
		)
		return
	}

	diags = res.State.Set(ctx, current)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}
}

// Delete implements resource.Resource.
func (b *divisionResource) Delete(ctx context.Context, req resource.DeleteRequest, res *resource.DeleteResponse) {
	var state Division

	diags := req.State.Get(ctx, &state)
	res.Diagnostics.Append(diags...)

	if res.Diagnostics.HasError() {
		return
	}

	err := retry.RetryContext(
		ctx,
		5*time.Second,
		func() *retry.RetryError {
			_, err := b.client.BusinessUnits().
				WithId(state.ID.ValueString()).
				Delete().
				Version(int(state.Version.ValueInt64())).
				Execute(ctx)

			return utils.ProcessRemoteError(err)
		},
	)
	if err != nil {
		res.Diagnostics.AddError(
			"Error deleting business unit",
			"Could not delete business unit, unexpected error: "+err.Error(),
		)
		return
	}
}

// Read implements resource.Resource.
func (b *divisionResource) Read(ctx context.Context, req resource.ReadRequest, res *resource.ReadResponse) {
	var state Division
	diags := req.State.Get(ctx, &state)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	bu, err := b.client.BusinessUnits().WithId(state.ID.ValueString()).Get().Execute(ctx)
	if err != nil {
		if errors.Is(err, platform.ErrNotFound) {
			res.State.RemoveResource(ctx)
			return
		}

		res.Diagnostics.AddError(
			"Error reading business unit",
			"Could not retrieve the business unit, unexpected error: "+err.Error(),
		)
		return
	}

	current, err := NewDivisionFromNative(bu)
	if err != nil {
		res.Diagnostics.AddError(
			"Error mapping business unit",
			"Could not create business unit, unexpected error: "+err.Error(),
		)
		return
	}

	diags = res.State.Set(ctx, current)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}
}

// Update implements resource.Resource.
func (b *divisionResource) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	var plan Division
	diags := req.Plan.Get(ctx, &plan)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	var state Division
	diags = req.State.Get(ctx, &state)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	input, err := state.updateActions(plan)
	if err != nil {
		res.Diagnostics.AddError(
			"Error updating business unit",
			"Could not update business unit, unexpected error: "+err.Error(),
		)
		return

	}
	var bu *platform.BusinessUnit

	err = retry.RetryContext(ctx, 5*time.Second, func() *retry.RetryError {
		var err error
		bu, err = b.client.BusinessUnits().
			WithId(state.ID.ValueString()).
			Post(input).
			Execute(ctx)

		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		res.Diagnostics.AddError(
			"Error updating business unit",
			"Could not update business unit, unexpected error: "+err.Error(),
		)
		return
	}

	current, err := NewDivisionFromNative(bu)
	if err != nil {
		res.Diagnostics.AddError(
			"Error mapping business unit",
			"Could not create business unit, unexpected error: "+err.Error(),
		)
		return
	}

	diags = res.State.Set(ctx, &current)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}
}
