package business_unit

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
	_ resource.Resource                = &businessUnitResource{}
	_ resource.ResourceWithConfigure   = &businessUnitResource{}
	_ resource.ResourceWithImportState = &businessUnitResource{}
)

type businessUnitResource struct {
	client *platform.ByProjectKeyRequestBuilder
}

// Schema implements resource.Resource.
func (r *businessUnitResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Business Units are used to group Associates and Associate Roles.\n\n" +
			"See also the [Business Unit API Documentation](https://docs.commercetools.com/api/projects/business-units",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the Business Unit.",
				Computed:    true,
			},
			"version": schema.Int64Attribute{
				Description: "Current version of the Business Unit.",
				Computed:    true,
			},
			"key": schema.StringAttribute{
				Description: "User-defined unique identifier of the Business Unit.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "User-defined name of the Business Unit.",
				Required:    true,
			},
			"status": schema.StringAttribute{
				Description: "The status of the Business Unit.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(BusinessUnitActive, BusinessUnitInactive),
				},
			},
			"store_mode": schema.StringAttribute{
				Description: "The store mode of the Business Unit.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(StoreModeExplicit, StoreModeFromParent),
				},
			},
			"unit_type": schema.StringAttribute{
				Description: "The type of the Business Unit.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(CompanyType, DivisionType),
				},
			},
			"associate_mode": schema.StringAttribute{
				Description: "The associate mode of the Business Unit.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(ExplicitAssociateMode, ExplicitAndFromParentAssociateMode),
				},
			},
			"contact_email": schema.StringAttribute{
				Description: "Email address of the Business Unit.",
				Required:    false,
			},
			"shipping_address_ids": schema.ListAttribute{
				Description: "List of shipping addresses used by the Business Unit.",
				Required:    false,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"default_shipping_address_id": schema.StringAttribute{
				Description: "ID of the default shipping Address.",
				Required:    false,
			},
			"billing_address_ids": schema.ListAttribute{
				Description: "List of billing addresses used by the Business Unit.",
				Required:    false,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"default_billing_address_id": schema.StringAttribute{
				Description: "ID of the default billing Address.",
				Required:    false,
			},
		},
		Blocks: map[string]schema.Block{
			"associates": schema.ListNestedBlock{
				Description: "Associates that are part of the Business Unit in specific roles",
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"associate_role_assignments": schema.ListNestedBlock{
							Description: "Roles assigned to the Associate within a Business Unit.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"inheritance": schema.StringAttribute{
										Description: "Determines whether the AssociateRoleAssignment can be inherited by child Business Units",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf(
												AssociateRoleInheritanceEnabled,
												AssociateRoleInheritanceDisabled,
											),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"associate_role": schema.ListNestedBlock{
										Description: "Reference to an AssociateRole by its key.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"key": schema.StringAttribute{
													Description: "User-defined unique identifier of the Associate Role",
													Required:    true,
												},
												"type_id": schema.StringAttribute{
													Description: "The type of the Associate Role",
													Required:    true,
													Validators: []validator.String{
														stringvalidator.OneOf(AssociateRoleTypeID),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"addresses": schema.ListNestedBlock{
				Description: "Addresses used by the Business Unit.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Unique identifier of the Address",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "User-defined unique identifier of the Address",
							Required:    false,
						},
						"external_id": schema.StringAttribute{
							Description: "ID for the contact used in an external system",
							Required:    false,
						},
						"country": schema.StringAttribute{
							Description: "Name of the country",
							Required:    true,
						},
						"title": schema.StringAttribute{
							Description: "Title of the contact, for example Dr., Prof.",
							Required:    false,
						},
						"salutation": schema.StringAttribute{
							Description: "Salutation of the contact, for example Ms., Mr.",
							Required:    false,
						},
						"first_name": schema.StringAttribute{
							Description: "First name of the contact",
							Required:    false,
						},
						"last_name": schema.StringAttribute{
							Description: "Last name of the contact",
							Required:    false,
						},
						"street_name": schema.StringAttribute{
							Description: "Name of the street",
							Required:    false,
						},
						"street_number": schema.StringAttribute{
							Description: "Street number",
							Required:    false,
						},
						"additional_street_info": schema.StringAttribute{
							Description: "Further information on the street address",
							Required:    false,
						},
						"postal_code": schema.StringAttribute{
							Description: "Postal code",
							Required:    false,
						},
						"city": schema.StringAttribute{
							Description: "Name of the city",
							Required:    false,
						},
						"region": schema.StringAttribute{
							Description: "Name of the region",
							Required:    false,
						},
						"state": schema.StringAttribute{
							Description: "Name of the state",
							Required:    false,
						},
						"company": schema.StringAttribute{
							Description: "Name of the company",
							Required:    false,
						},
						"department": schema.StringAttribute{
							Description: "Name of the department",
							Required:    false,
						},
						"building": schema.StringAttribute{
							Description: "Name or number of the building",
							Required:    false,
						},
						"apartment": schema.StringAttribute{
							Description: "Name or number of the apartment",
							Required:    false,
						},
						"po_box": schema.StringAttribute{
							Description: "Post office box number",
							Required:    false,
						},
						"phone": schema.StringAttribute{
							Description: "Phone number",
							Required:    false,
						},
						"mobile": schema.StringAttribute{
							Description: "Mobile phone number",
							Required:    false,
						},
						"email": schema.StringAttribute{
							Description: "Email address",
							Required:    false,
						},
						"fax": schema.StringAttribute{
							Description: "Fax number",
							Required:    false,
						},
						"additional_address_info": schema.StringAttribute{
							Description: "Further information on the Address",
							Required:    false,
						},
					},
				},
			},
			"top_level_unit": schema.ListNestedBlock{
				Description: "Reference to a parent Business Unit by its key.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.SizeAtLeast(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Description: "User-defined unique identifier of the Business Unit",
							Required:    true,
						},
						"type_id": schema.StringAttribute{
							Description: "The type of the Business Unit",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf(CompanyType),
							},
						},
					},
				},
			},
			"parent_unit": schema.ListNestedBlock{
				Description: "Reference to a parent Business Unit by its key.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.SizeAtLeast(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Description: "User-defined unique identifier of the Business Unit",
							Required:    true,
						},
						"type_id": schema.StringAttribute{
							Description: "The type of the Business Unit",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf(CompanyType, DivisionType),
							},
						},
					},
				},
			},
			"stores": schema.ListNestedBlock{
				Description: "Stores that are part of the Business Unit.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Description: "User-defined unique identifier of the Store",
							Required:    true,
						},
						"type_id": schema.StringAttribute{
							Description: "The type of the Store",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf(StoreTypeID),
							},
						},
					},
				},
			},
			"inherited_associates": schema.ListNestedBlock{
				Description: "Associates that are inherited from parent Business Unit",
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"inherited_associate_role_assignments": schema.ListNestedBlock{
							Description: "Inherited roles of the Associate within a Business Unit",
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"associate_role": schema.ListNestedBlock{
										Description: "Inherited role the Associate holds within a Business Unit",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"key": schema.StringAttribute{
													Description: "User-defined unique identifier of the Associate Role",
													Required:    true,
												},
												"type_id": schema.StringAttribute{
													Description: "The type of the Associate Role",
													Required:    true,
													Validators: []validator.String{
														stringvalidator.OneOf(AssociateRoleTypeID),
													},
												},
											},
										},
									},
									"source": schema.ListNestedBlock{
										Description: "",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"key": schema.StringAttribute{
													Description: "User-defined unique identifier of the Business Unit",
													Required:    true,
												},
												"type_id": schema.StringAttribute{
													Description: "The type of the Business Unit",
													Required:    true,
													Validators: []validator.String{
														stringvalidator.OneOf(BusinessUnitTypeID),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &businessUnitResource{}
}

// Metadata implements resource.Resource.
func (r *businessUnitResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_business_unit"
}

// Create implements resource.Resource.
func (r *businessUnitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BusinessUnit
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	draft := plan.draft()

	var businessUnit platform.BusinessUnit
	err := retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
		var err error
		businessUnit, err = r.client.BusinessUnits().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating business unit",
			err.Error(),
		)
		return
	}

	current := NewBusinessUnitFromNative(businessUnit)

	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read implements resource.Resource.
func (r *businessUnitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BusinessUnit
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	businessUnit, err := r.client.BusinessUnits().WithId(state.ID.ValueString()).Get().Execute(ctx)
	if err != nil {
		if errors.Is(err, platform.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading business unit",
			"Could not retrieve the business unit, unexpected error: "+err.Error(),
		)
		return
	}

	// Transform the remote platform business unit to the
	// tf schema matching representation.
	current := NewBusinessUnitFromNative(businessUnit)

	// Set current data as state.
	diags = resp.State.Set(ctx, &current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update implements resource.Resource.
func (r *businessUnitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BusinessUnit
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state BusinessUnit
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := state.updateActions(plan)
	var businessUnit platform.BusinessUnit
	err := retry.RetryContext(ctx, 5*time.Second, func() *retry.RetryError {
		var err error
		businessUnit, err = r.client.BusinessUnits().
			WithId(state.ID.ValueString()).
			Post(input).
			Execute(ctx)

		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating business unit",
			"Could not update business unit, unexpected error: "+err.Error(),
		)
		return
	}

	current := NewBusinessUnitFromNative(businessUnit)

	// Set current data as state.
	diags = resp.State.Set(ctx, &current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete implements resource.Resource.
func (r *businessUnitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BusinessUnit
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := retry.RetryContext(
		ctx,
		5*time.Second,
		func() *retry.RetryError {
			_, err := r.client.BusinessUnits().
				WithId(state.ID.ValueString()).
				Delete().
				Version(int(state.Version.ValueInt64())).
				Execute(ctx)

			return utils.ProcessRemoteError(err)
		})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting business unit",
			"Could not delete business unit, unexpected error: "+err.Error(),
		)
		return
	}
}

// Configure implements resource.ResourceWithConfigure.
func (r *businessUnitResource) Configure(ctx context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	data := req.ProviderData.(*utils.ProviderData)
	r.client = data.Client
}

// ImportState implements resource.ResourceWithImportState.
func (*businessUnitResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
