package discount_group

import (
	"context"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	sdk_resource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client *platform.ByProjectKeyRequestBuilder
}

// Metadata returns the resource type name.
func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_discount_group"
}

// Schema defines the schema for the resource.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A Discount Group combines Cart Discounts that are applied as one: the group has one " +
			"sort order, and only its best Cart Discount applies to a Cart. Cart Discounts join a group through " +
			"`discount_group_id` on `commercetools_cart_discount`.\n\n" +
			"See also the [Discount Groups API documentation](https://docs.commercetools.com/api/projects/discount-groups).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Platform-generated unique identifier of the DiscountGroup.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.Int64Attribute{
				Description: "Current version of the DiscountGroup.",
				Computed:    true,
			},
			"key": schema.StringAttribute{
				Description: "User-defined unique identifier of the DiscountGroup.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 256),
					stringvalidator.RegexMatches(
						regexp.MustCompile("^[A-Za-z0-9_-]+$"),
						"Key must match pattern ^[A-Za-z0-9_-]+$"),
				},
			},
			"name": schema.MapAttribute{
				CustomType:  customtypes.NewLocalizedStringType(),
				Description: "Name of the DiscountGroup as localized string.",
				Optional:    true,
			},
			"description": schema.MapAttribute{
				CustomType:  customtypes.NewLocalizedStringType(),
				Description: "Description of the DiscountGroup as localized string.",
				Optional:    true,
			},
			"sort_order": schema.StringAttribute{
				MarkdownDescription: "A decimal between 0 and 1, such as `\"0.5\"`, that orders the group among " +
					"other groups and Cart Discounts; a higher value applies first. It must be unique among Cart " +
					"Discounts and Discount Groups.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^0?\.\d+$`),
						"must be a decimal between 0 and 1, such as \"0.5\""),
				},
			},
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the group's Cart Discounts are considered at all. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*utils.ProviderData)
	if !ok {
		return
	}
	r.client = data.Client
}

// Create creates the resource and sets the initial Terraform state.
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DiscountGroup
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	draft := toDraft(&plan)
	var res *platform.DiscountGroup
	err := sdk_resource.RetryContext(ctx, 20*time.Second, func() *sdk_resource.RetryError {
		var err error
		res, err = r.client.DiscountGroups().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating discount group", err.Error())
		return
	}

	current := fromNative(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DiscountGroup
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.DiscountGroups().WithId(state.ID.ValueString()).Get().Execute(ctx)
	if err != nil {
		if utils.IsResourceNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading discount group", err.Error())
		return
	}

	current := fromNative(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state DiscountGroup
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updates := toUpdateActions(&state, &plan)
	var res *platform.DiscountGroup
	err := sdk_resource.RetryContext(ctx, 5*time.Second, func() *sdk_resource.RetryError {
		var err error
		res, err = r.client.DiscountGroups().WithId(state.ID.ValueString()).Post(updates).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating discount group", err.Error())
		return
	}

	current := fromNative(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DiscountGroup
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := sdk_resource.RetryContext(ctx, 5*time.Second, func() *sdk_resource.RetryError {
		_, err := r.client.DiscountGroups().
			WithId(state.ID.ValueString()).
			Delete().
			Version(int(state.Version.ValueInt64())).
			Execute(ctx)
		if utils.IsResourceNotFoundError(err) {
			return nil
		}
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting discount group", err.Error())
	}
}

// ImportState imports a discount group by its id.
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
