package recurrence_policy

import (
	"context"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	sdk_resource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/customvalidator"
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
	resp.TypeName = req.ProviderTypeName + "_recurrence_policy"
}

// Schema defines the schema for the resource.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A Recurrence Policy is the schedule Recurring Orders repeat on: every so many " +
			"days, weeks or months, or on a day of the month.\n\n" +
			"See also the [Recurrence Policies API documentation](https://docs.commercetools.com/api/projects/recurrence-policies).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Platform-generated unique identifier of the RecurrencePolicy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.Int64Attribute{
				Description: "Current version of the RecurrencePolicy.",
				Computed:    true,
			},
			"key": schema.StringAttribute{
				Description: "User-defined unique identifier of the RecurrencePolicy.",
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
				Description: "Name of the RecurrencePolicy as localized string.",
				Optional:    true,
			},
			"description": schema.MapAttribute{
				CustomType:  customtypes.NewLocalizedStringType(),
				Description: "Description of the RecurrencePolicy as localized string.",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"schedule": schema.ListNestedBlock{
				MarkdownDescription: "When orders repeat: `standard` every `value` `interval_unit`s, or " +
					"`dayOfMonth` on `day` of every month.",
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeBetween(1, 1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "`standard` or `dayOfMonth`.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(Standard, DayOfMonth),
								customvalidator.DependencyValidator(
									Standard,
									path.MatchRelative().AtParent().AtName("value"),
									path.MatchRelative().AtParent().AtName("interval_unit"),
								),
								customvalidator.DependencyValidator(
									DayOfMonth,
									path.MatchRelative().AtParent().AtName("day"),
								),
							},
						},
						"value": schema.Int64Attribute{
							MarkdownDescription: "How many `interval_unit`s lie between orders, for `standard`.",
							Optional:            true,
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
						},
						"interval_unit": schema.StringAttribute{
							MarkdownDescription: "`Days`, `Weeks` or `Months`, for `standard`.",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("Days", "Weeks", "Months"),
							},
						},
						"day": schema.Int64Attribute{
							MarkdownDescription: "Day of the month orders are created on, from 1 to 31, for " +
								"`dayOfMonth`. In a shorter month it is the month's last day.",
							Optional: true,
							Validators: []validator.Int64{
								int64validator.Between(1, 31),
							},
						},
					},
				},
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
	var plan RecurrencePolicy
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	draft := toDraft(&plan)
	var res *platform.RecurrencePolicy
	err := sdk_resource.RetryContext(ctx, 20*time.Second, func() *sdk_resource.RetryError {
		var err error
		res, err = r.client.RecurrencePolicies().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating recurrence policy", err.Error())
		return
	}

	current := fromNative(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RecurrencePolicy
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.RecurrencePolicies().WithId(state.ID.ValueString()).Get().Execute(ctx)
	if err != nil {
		if utils.IsResourceNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading recurrence policy", err.Error())
		return
	}

	current := fromNative(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state RecurrencePolicy
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updates := toUpdateActions(&state, &plan)
	var res *platform.RecurrencePolicy
	err := sdk_resource.RetryContext(ctx, 5*time.Second, func() *sdk_resource.RetryError {
		var err error
		res, err = r.client.RecurrencePolicies().WithId(state.ID.ValueString()).Post(updates).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating recurrence policy", err.Error())
		return
	}

	current := fromNative(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RecurrencePolicy
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := sdk_resource.RetryContext(ctx, 5*time.Second, func() *sdk_resource.RetryError {
		_, err := r.client.RecurrencePolicies().
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
		resp.Diagnostics.AddError("Error deleting recurrence policy", err.Error())
	}
}

// ImportState imports a recurrence policy by its id.
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
