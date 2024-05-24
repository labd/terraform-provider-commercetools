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
	_ resource.Resource                = &companyResource{}
	_ resource.ResourceWithConfigure   = &companyResource{}
	_ resource.ResourceWithImportState = &companyResource{}
)

type companyResource struct {
	client *platform.ByProjectKeyRequestBuilder
}

// Schema implements resource.Resource.
func (b *companyResource) Schema(_ context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schema.Schema{
		Description: "Business Unit type to represent the top level of a business. Contains specific fields and values that differentiate a Company from the generic BusinessUnit.\n\n" +
			"See also the [Business Unit API Documentation](https://docs.commercetools.com/api/projects/business-units",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the Company.",
				Computed:    true,
			},
			"version": schema.Int64Attribute{
				Description: "The current version of the Company.",
				Computed:    true,
			},
			"key": schema.StringAttribute{
				Description: "User-defined unique identifier for the Company.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the Company.",
				Required:    true,
			},
			"status": schema.StringAttribute{
				Description: "The status of the Company.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(BusinessUnitActive, BusinessUnitInactive),
				},
			},
			"contact_email": schema.StringAttribute{
				Description: "The email address of the Company.",
				Optional:    true,
			},
			"shipping_address_ids": schema.ListAttribute{
				Description: "List of the shipping addresses used by the Company.",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"billing_address_ids": schema.ListAttribute{
				Description: "List of the billing addresses used by the Company.",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"default_shipping_address_id": schema.StringAttribute{
				Description: "ID of the default shipping Address.",
				Optional:    true,
			},
			"default_billing_address_id": schema.StringAttribute{
				Description: "ID of the default billing Address.",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"associate": schema.ListNestedBlock{
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
			"address": schema.ListNestedBlock{
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
							Optional:    true,
						},
						"external_id": schema.StringAttribute{
							Description: "ID for the contact used in an external system",
							Optional:    true,
						},
						"country": schema.StringAttribute{
							Description: "Name of the country",
							Required:    true,
						},
						"title": schema.StringAttribute{
							Description: "Title of the contact, for example Dr., Prof.",
							Optional:    true,
						},
						"salutation": schema.StringAttribute{
							Description: "Salutation of the contact, for example Ms., Mr.",
							Optional:    true,
						},
						"first_name": schema.StringAttribute{
							Description: "First name of the contact",
							Optional:    true,
						},
						"last_name": schema.StringAttribute{
							Description: "Last name of the contact",
							Optional:    true,
						},
						"street_name": schema.StringAttribute{
							Description: "Name of the street",
							Optional:    true,
						},
						"street_number": schema.StringAttribute{
							Description: "Street number",
							Optional:    true,
						},
						"additional_street_info": schema.StringAttribute{
							Description: "Further information on the street address",
							Optional:    true,
						},
						"postal_code": schema.StringAttribute{
							Description: "Postal code",
							Optional:    true,
						},
						"city": schema.StringAttribute{
							Description: "Name of the city",
							Optional:    true,
						},
						"region": schema.StringAttribute{
							Description: "Name of the region",
							Optional:    true,
						},
						"state": schema.StringAttribute{
							Description: "Name of the state",
							Optional:    true,
						},
						"company": schema.StringAttribute{
							Description: "Name of the company",
							Optional:    true,
						},
						"department": schema.StringAttribute{
							Description: "Name of the department",
							Optional:    true,
						},
						"building": schema.StringAttribute{
							Description: "Name or number of the building",
							Optional:    true,
						},
						"apartment": schema.StringAttribute{
							Description: "Name or number of the apartment",
							Optional:    true,
						},
						"po_box": schema.StringAttribute{
							Description: "Post office box number",
							Optional:    true,
						},
						"phone": schema.StringAttribute{
							Description: "Phone number",
							Optional:    true,
						},
						"mobile": schema.StringAttribute{
							Description: "Mobile phone number",
							Optional:    true,
						},
						"email": schema.StringAttribute{
							Description: "Email address",
							Optional:    true,
						},
						"fax": schema.StringAttribute{
							Description: "Fax number",
							Optional:    true,
						},
						"additional_address_info": schema.StringAttribute{
							Description: "Further information on the Address",
							Optional:    true,
						},
					},
				},
			},
			"store": schema.ListNestedBlock{
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

	draft := plan.draft()

	var bu platform.BusinessUnit
	err := retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
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

	current := NewCompanyFromNative(bu)

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

// Read implements resource.Resource.
func (b *companyResource) Read(ctx context.Context, req resource.ReadRequest, res *resource.ReadResponse) {
	var state Company
	diags := req.State.Get(ctx, &state)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	company, err := b.client.BusinessUnits().WithId(state.ID.ValueString()).Get().Execute(ctx)
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

	current := NewCompanyFromNative(company)

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

	input := state.updateActions(plan)
	var company *platform.BusinessUnit

	err := retry.RetryContext(ctx, 5*time.Second, func() *retry.RetryError {
		var err error
		company, err = b.client.BusinessUnits().
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

	current := NewCompanyFromNative(company)
	diags = res.State.Set(ctx, &current)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}
}

func NewCompanyResource() resource.Resource {
	return &companyResource{}
}

var (
	_ resource.Resource                = &divisionResource{}
	_ resource.ResourceWithConfigure   = &divisionResource{}
	_ resource.ResourceWithImportState = &divisionResource{}
)

type divisionResource struct {
	client *platform.ByProjectKeyRequestBuilder
}

// Schema implements resource.Resource.
func (b *divisionResource) Schema(_ context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schema.Schema{
		Description: "Business Unit type to represent the top level of a business. Contains specific fields and values that differentiate a Company from the generic BusinessUnit.\n\n" +
			"See also the [Business Unit API Documentation](https://docs.commercetools.com/api/projects/business-units",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the Company.",
				Computed:    true,
			},
			"version": schema.Int64Attribute{
				Description: "The current version of the Company.",
				Computed:    true,
			},
			"key": schema.StringAttribute{
				Description: "User-defined unique identifier for the Company.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the Company.",
				Required:    true,
			},
			"status": schema.StringAttribute{
				Description: "The status of the Company.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(BusinessUnitActive, BusinessUnitInactive),
				},
			},
			"associate_mode": schema.StringAttribute{
				Description: "The association mode of the Company.",
				Required:    true,
			},
			"contact_email": schema.StringAttribute{
				Description: "The email address of the Company.",
				Optional:    true,
			},
			"shipping_address_ids": schema.ListAttribute{
				Description: "List of the shipping addresses used by the Company.",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"billing_address_ids": schema.ListAttribute{
				Description: "List of the billing addresses used by the Company.",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"default_shipping_address_id": schema.StringAttribute{
				Description: "ID of the default shipping Address.",
				Optional:    true,
			},
			"default_billing_address_id": schema.StringAttribute{
				Description: "ID of the default billing Address.",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"associate": schema.ListNestedBlock{
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
			"address": schema.ListNestedBlock{
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
							Optional:    true,
						},
						"external_id": schema.StringAttribute{
							Description: "ID for the contact used in an external system",
							Optional:    true,
						},
						"country": schema.StringAttribute{
							Description: "Name of the country",
							Required:    true,
						},
						"title": schema.StringAttribute{
							Description: "Title of the contact, for example Dr., Prof.",
							Optional:    true,
						},
						"salutation": schema.StringAttribute{
							Description: "Salutation of the contact, for example Ms., Mr.",
							Optional:    true,
						},
						"first_name": schema.StringAttribute{
							Description: "First name of the contact",
							Optional:    true,
						},
						"last_name": schema.StringAttribute{
							Description: "Last name of the contact",
							Optional:    true,
						},
						"street_name": schema.StringAttribute{
							Description: "Name of the street",
							Optional:    true,
						},
						"street_number": schema.StringAttribute{
							Description: "Street number",
							Optional:    true,
						},
						"additional_street_info": schema.StringAttribute{
							Description: "Further information on the street address",
							Optional:    true,
						},
						"postal_code": schema.StringAttribute{
							Description: "Postal code",
							Optional:    true,
						},
						"city": schema.StringAttribute{
							Description: "Name of the city",
							Optional:    true,
						},
						"region": schema.StringAttribute{
							Description: "Name of the region",
							Optional:    true,
						},
						"state": schema.StringAttribute{
							Description: "Name of the state",
							Optional:    true,
						},
						"company": schema.StringAttribute{
							Description: "Name of the company",
							Optional:    true,
						},
						"department": schema.StringAttribute{
							Description: "Name of the department",
							Optional:    true,
						},
						"building": schema.StringAttribute{
							Description: "Name or number of the building",
							Optional:    true,
						},
						"apartment": schema.StringAttribute{
							Description: "Name or number of the apartment",
							Optional:    true,
						},
						"po_box": schema.StringAttribute{
							Description: "Post office box number",
							Optional:    true,
						},
						"phone": schema.StringAttribute{
							Description: "Phone number",
							Optional:    true,
						},
						"mobile": schema.StringAttribute{
							Description: "Mobile phone number",
							Optional:    true,
						},
						"email": schema.StringAttribute{
							Description: "Email address",
							Optional:    true,
						},
						"fax": schema.StringAttribute{
							Description: "Fax number",
							Optional:    true,
						},
						"additional_address_info": schema.StringAttribute{
							Description: "Further information on the Address",
							Optional:    true,
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

	draft := plan.draft()

	var bu platform.BusinessUnit
	err := retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
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

	current := NewDivisionFromNative(bu)

	diags = res.State.Set(ctx, &current)
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

	division, err := b.client.BusinessUnits().WithId(state.ID.ValueString()).Get().Execute(ctx)
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

	current := NewDivisionFromNative(division)

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

	input := state.updateActions(plan)
	var division *platform.BusinessUnit

	err := retry.RetryContext(ctx, 5*time.Second, func() *retry.RetryError {
		var err error
		division, err = b.client.BusinessUnits().
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

	current := NewDivisionFromNative(division)
	diags = res.State.Set(ctx, &current)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}
}

// NewDivisionResource creates a new resource for the Division type.
func NewDivisionResource() resource.Resource {
	return &divisionResource{}
}
