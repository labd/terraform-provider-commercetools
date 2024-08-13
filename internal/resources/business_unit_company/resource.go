package business_unit_company

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/labd/terraform-provider-commercetools/internal/sharedtypes"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

var (
	_ resource.Resource                = &companyResource{}
	_ resource.ResourceWithConfigure   = &companyResource{}
	_ resource.ResourceWithImportState = &companyResource{}
)

type companyResource struct {
	client *platform.ByProjectKeyRequestBuilder
}

func NewCompanyResource() resource.Resource {
	return &companyResource{}
}

// Schema implements resource.Resource.
func (b *companyResource) Schema(_ context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schema.Schema{
		MarkdownDescription: "Business Unit type to represent the top level of a business. Contains specific fields and values that differentiate a Company from the generic BusinessUnit.\n\n" +
			"See also the [Business Unit API Documentation](https://docs.commercetools.com/api/projects/business-units",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the Company.",
				Computed:            true,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "The current version of the Company.",
				Computed:            true,
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "User-defined unique identifier for the Company.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 256),
					stringvalidator.RegexMatches(
						regexp.MustCompile("^[A-Za-z0-9_-]+$"),
						"Key must match pattern ^[A-Za-z0-9_-]+$",
					),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the Company.",
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
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Company.",
				Required:            true,
			},
			"contact_email": schema.StringAttribute{
				MarkdownDescription: "The email address of the Company.",
				Optional:            true,
			},
			"shipping_address_keys": schema.SetAttribute{
				MarkdownDescription: "Indexes of entries in addresses to set as shipping addresses. The shippingAddressIds of the [Customer](https://docs.commercetools.com/api/projects/customers) will be replaced by these addresses.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"default_shipping_address_key": schema.StringAttribute{
				MarkdownDescription: "Index of the entry in addresses to set as the default shipping address.",
				Optional:            true,
			},
			"billing_address_keys": schema.SetAttribute{
				MarkdownDescription: "Indexes of entries in addresses to set as billing addresses. The billingAddressIds of the [Customer](https://docs.commercetools.com/api/projects/customers) will be replaced by these addresses.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"default_billing_address_key": schema.StringAttribute{
				MarkdownDescription: "Index of the entry in addresses to set as the default billing address.",
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"store":   sharedtypes.StoreKeyReferenceBlockSchema,
			"address": sharedtypes.AddressBlockSchema,
		},
	}
}

// Metadata implements resource.Resource.
func (b *companyResource) Metadata(_ context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_business_unit_company"
}

// ImportState implements resource.ResourceWithImportState.
func (b *companyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, res *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, res)
}

// Configure implements resource.ResourceWithConfigure.
func (b *companyResource) Configure(ctx context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
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
func (b *companyResource) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	var plan Company
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

	current, err := NewCompanyFromNative(bu)
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

// Read implements resource.Resource.
func (b *companyResource) Read(ctx context.Context, req resource.ReadRequest, res *resource.ReadResponse) {
	var state Company
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

	current, err := NewCompanyFromNative(bu)
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
func (b *companyResource) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	var plan Company
	diags := req.Plan.Get(ctx, &plan)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	var state Company
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

	current, err := NewCompanyFromNative(bu)
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

// Delete implements resource.Resource.
func (b *companyResource) Delete(ctx context.Context, req resource.DeleteRequest, res *resource.DeleteResponse) {
	var state Company

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
