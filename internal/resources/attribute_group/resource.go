package attribute_group

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"regexp"
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
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client *platform.ByProjectKeyRequestBuilder
	mutex  *utils.MutexKV
}

// Metadata returns the data source type name.
func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_attribute_group"
}

// Schema defines the schema for the data source.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Attribute Groups allow you to define a set of referenced Attribute Definitions for the purpose of " +
			"giving one or more dedicated teams access to edit Attribute values in Product Variants for those Attributes " +
			"in the Merchant Center. Depending on the use case, editing permission can be granted to all Attributes, to " +
			"Attributes that are ungrouped, or none.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Platform-generated unique identifier of the AttributeGroup.",
				Computed:    true,
			},
			"version": schema.Int64Attribute{
				Description: "Current version of the AttributeGroup.",
				Computed:    true,
			},
			"key": schema.StringAttribute{
				Description: "User-defined unique identifier of the AttributeGroup.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 256),
					stringvalidator.RegexMatches(
						regexp.MustCompile("^[A-Za-z0-9_-]+$"),
						"Key must match pattern ^[A-Za-z0-9_-]+$"),
				},
			},
			"name": schema.MapAttribute{
				CustomType:          customtypes.NewLocalizedStringType(),
				Description:         "Name of the State as localized string.",
				MarkdownDescription: "Name of the State as localized string.",
				Required:            true,
			},
			"description": schema.MapAttribute{
				CustomType:          customtypes.NewLocalizedStringType(),
				Description:         "Description of the State as localized string.",
				MarkdownDescription: "Description of the State as localized string.",
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"attribute": schema.ListNestedBlock{
				MarkdownDescription: "Attributes with unique values.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Description: "The Attribute's name as given in its AttributeDefinition.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(2, 256),
								stringvalidator.RegexMatches(
									regexp.MustCompile("^[A-Za-z0-9_-]+$"),
									"Key must match pattern ^[A-Za-z0-9_-]+$"),
							},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data := req.ProviderData.(*utils.ProviderData)
	r.client = data.Client
	r.mutex = data.Mutex
}

// Create creates the resource and sets the initial Terraform state.
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from attributeGroup
	var attributeGroup AttributeGroup
	diags := req.Plan.Get(ctx, &attributeGroup)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	draft := toDraft(&attributeGroup)
	var res *platform.AttributeGroup
	err := sdk_resource.RetryContext(ctx, 20*time.Second, func() *sdk_resource.RetryError {
		var err error
		res, err = r.client.AttributeGroups().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating subscription",
			err.Error(),
		)
		return
	}

	current := fromNative(res)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var current AttributeGroup
	diags := req.State.Get(ctx, &current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.AttributeGroups().WithId(current.ID.ValueString()).Get().Execute(ctx)
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

	current = fromNative(res)

	// Set refreshed state
	diags = resp.State.Set(ctx, &current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan AttributeGroup
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var state AttributeGroup
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updates := toUpdateActions(&state, &plan)

	var res *platform.AttributeGroup
	err := sdk_resource.RetryContext(ctx, 5*time.Second, func() *sdk_resource.RetryError {
		var err error
		res, err = r.client.AttributeGroups().WithId(state.ID.ValueString()).Post(updates).Execute(ctx)

		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating project", err.Error())
		return
	}
	result := fromNative(res)

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state AttributeGroup
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.AttributeGroups().WithId(state.ID.ValueString()).Get().Execute(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading state",
			"Could not retrieve state, unexpected error: "+err.Error(),
		)
		return
	}
	current := fromNative(res)
	current.Version = types.Int64Value(int64(res.Version))

	err = sdk_resource.RetryContext(ctx, 5*time.Second, func() *sdk_resource.RetryError {
		var err error
		res, err = r.client.AttributeGroups().
			WithId(state.ID.ValueString()).
			Delete().
			Version(int(current.Version.ValueInt64())).
			Execute(ctx)
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

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
