package insights_configuration

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk_resource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/labd/commercetools-go-sdk/insights"

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
	client *insights.ByProjectKeyRequestBuilder
	mutex  *utils.MutexKV
}

// Metadata returns the resource type name.
func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_insights_configuration"
}

// Schema defines the schema for the resource.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Platform Insights exports logs and metrics from your Project to a supported " +
			"Application Performance Monitoring (APM) provider.\n\n" +
			"A Project has at most one Platform Insights configuration, so only a single resource of this type " +
			"should be defined per Project.\n\n" +
			"See also the [Platform Insights API documentation](https://docs.commercetools.com/api/platform-insights).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The key of the Project the configuration belongs to.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_key": schema.StringAttribute{
				Description: "User-defined unique identifier of the Project.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"active": schema.BoolAttribute{
				Description: "Whether the Platform Insights configuration is active.",
				Optional:    true,
				Computed:    true,
			},
			"last_modified_at": schema.StringAttribute{
				Description: "Date and time (UTC) the Platform Insights configuration was last updated.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"apm_provider": schema.ListNestedBlock{
				MarkdownDescription: "An APM provider to export events to. At most five providers can be " +
					"configured, and each provider `type` can be used only once.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(5),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "The APM provider type. One of `NewRelic`, `DynatraceSaaS`, " +
								"`DynatraceActiveGate`, `Datadog`, `Otlp` or `OtlpHttp`.",
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf(
									NewRelic,
									DynatraceSaaS,
									DynatraceActiveGate,
									Datadog,
									Otlp,
									OtlpHttp,
								),
								customvalidator.DependencyValidator(
									NewRelic,
									path.MatchRelative().AtParent().AtName("region"),
								),
								customvalidator.DependencyValidator(
									DynatraceSaaS,
									path.MatchRelative().AtParent().AtName("environment_id"),
								),
								customvalidator.DependencyValidator(
									DynatraceActiveGate,
									path.MatchRelative().AtParent().AtName("active_gate_domain"),
									path.MatchRelative().AtParent().AtName("environment_id"),
								),
								customvalidator.DependencyValidator(
									Datadog,
									path.MatchRelative().AtParent().AtName("site"),
								),
								customvalidator.DependencyValidator(
									Otlp,
									path.MatchRelative().AtParent().AtName("endpoint"),
								),
								customvalidator.DependencyValidator(
									OtlpHttp,
									path.MatchRelative().AtParent().AtName("endpoint"),
								),
							},
						},
						"active": schema.BoolAttribute{
							Description: "Whether this provider configuration is active.",
							Optional:    true,
							Computed:    true,
						},
						"valid": schema.BoolAttribute{
							Description: "Whether it is possible to send metrics to the APM provider.",
							Computed:    true,
						},
						"event_types": schema.ListAttribute{
							MarkdownDescription: "Event types to export to this provider. Valid values are " +
								"`Logs` and `Metrics`.",
							ElementType: types.StringType,
							Required:    true,
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
								listvalidator.ValueStringsAre(
									stringvalidator.OneOf("Logs", "Metrics"),
								),
							},
						},
						"region": schema.StringAttribute{
							MarkdownDescription: "Region in which your NewRelic account is based. One of `eu` " +
								"or `us`. Only applies to the `NewRelic` provider.",
							Optional: true,
							Validators: []validator.String{
								stringvalidator.OneOf("eu", "us"),
							},
						},
						"site": schema.StringAttribute{
							MarkdownDescription: "Website on which your Datadog account is hosted. One of " +
								"`US1`, `US3`, `US5`, `EU1`, `US1-FED` or `AP1`. Only applies to the `Datadog` " +
								"provider.",
							Optional: true,
							Validators: []validator.String{
								stringvalidator.OneOf("US1", "US3", "US5", "EU1", "US1-FED", "AP1"),
							},
						},
						"environment_id": schema.StringAttribute{
							MarkdownDescription: "Unique user-defined environment identifier of your Dynatrace " +
								"account. Only applies to the `DynatraceSaaS` and `DynatraceActiveGate` providers.",
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"active_gate_domain": schema.StringAttribute{
							MarkdownDescription: "Your ActiveGate domain. Only applies to the " +
								"`DynatraceActiveGate` provider.",
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"active_gate_port": schema.Int64Attribute{
							MarkdownDescription: "Your ActiveGate domain port. Only applies to the " +
								"`DynatraceActiveGate` provider.",
							Optional: true,
						},
						"endpoint": schema.StringAttribute{
							MarkdownDescription: "Destination for your OpenTelemetry Protocol (OTLP) provider. " +
								"Only applies to the `Otlp` and `OtlpHttp` providers.",
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"api_key": schema.StringAttribute{
							MarkdownDescription: "API key used to authenticate against the APM provider. Applies " +
								"to the `NewRelic`, `DynatraceSaaS`, `DynatraceActiveGate` and `Datadog` " +
								"providers.\n\n" +
								"The API never returns this value, so it cannot be read back from the Platform. " +
								"Terraform keeps the configured value in state and is unable to detect that it " +
								"was changed outside of Terraform.",
							Optional:  true,
							Sensitive: true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"header": schema.ListNestedBlock{
							MarkdownDescription: "Header sent when connecting to your OTLP provider. Only " +
								"applies to the `Otlp` and `OtlpHttp` providers.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "Name of the header.",
										Required:    true,
									},
									"value": schema.StringAttribute{
										MarkdownDescription: "Value of the header.\n\n" +
											"The API only returns header names, so this value cannot be read " +
											"back from the Platform. Terraform keeps the configured value in " +
											"state and is unable to detect that it was changed outside of " +
											"Terraform.",
										Required:  true,
										Sensitive: true,
									},
								},
							},
						},
					},
				},
			},
			"additional_attribute": schema.ListNestedBlock{
				MarkdownDescription: "An attribute added to the exported events. At most ten attributes can be " +
					"configured, and each `name` can be used only once.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(10),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"event_type": schema.StringAttribute{
							MarkdownDescription: "Event type the attribute is added to. Valid values are " +
								"`Logs` and `Metrics`.",
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf("Logs", "Metrics"),
							},
						},
						"name": schema.StringAttribute{
							Description: "Name of the attribute.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 30),
							},
						},
						"value": schema.StringAttribute{
							Description: "Value of the attribute.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 100),
							},
						},
					},
				},
			},
		},
	}
}

func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data := req.ProviderData.(*utils.ProviderData)
	r.client = data.InsightsClient
	r.mutex = data.Mutex
}

// Create creates the Platform Insights configuration and sets the initial state.
//
// The API exposes a create-or-replace (PUT) endpoint alongside an update-action
// (POST) endpoint. PUT expresses the whole desired configuration in one call,
// which matches Terraform's declarative model, so it backs both Create and
// Update here and the update actions are not used.
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan InsightsConfiguration
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var res *insights.ProjectConfiguration
	err := sdk_resource.RetryContext(ctx, 20*time.Second, func() *sdk_resource.RetryError {
		var err error
		res, err = r.client.InsightsConfiguration().Put(plan.draft()).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating insights configuration",
			err.Error(),
		)
		return
	}

	current := NewConfigurationFromNative(res)
	current.alignWith(plan)

	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state InsightsConfiguration
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.InsightsConfiguration().Get().Execute(ctx)
	if err != nil {
		if utils.IsResourceNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading insights configuration",
			err.Error(),
		)
		return
	}

	current := NewConfigurationFromNative(res)
	current.alignWith(state)

	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
}

// Update replaces the Platform Insights configuration with the planned one.
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan InsightsConfiguration
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var res *insights.ProjectConfiguration
	err := sdk_resource.RetryContext(ctx, 20*time.Second, func() *sdk_resource.RetryError {
		var err error
		res, err = r.client.InsightsConfiguration().Put(plan.draft()).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating insights configuration",
			err.Error(),
		)
		return
	}

	current := NewConfigurationFromNative(res)
	current.alignWith(plan)

	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
}

// Delete removes the Platform Insights configuration.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state InsightsConfiguration
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := sdk_resource.RetryContext(ctx, 20*time.Second, func() *sdk_resource.RetryError {
		_, err := r.client.InsightsConfiguration().Delete().Execute(ctx)
		if utils.IsResourceNotFoundError(err) {
			return nil
		}
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting insights configuration",
			err.Error(),
		)
	}
}

// ImportState imports the Platform Insights configuration of the Project.
//
// Provider credentials cannot be imported: the API does not return them, so
// api_key and OTLP header values are empty after an import and have to be
// applied once to bring the Platform in line with the configuration.
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
