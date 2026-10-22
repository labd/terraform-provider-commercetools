package mcp_server

import (
	"context"
	"encoding/json"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	resp.TypeName = req.ProviderTypeName + "_mcp_server"
}

// Schema defines the schema for the resource.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	stringList := func(description string) schema.ListAttribute {
		return schema.ListAttribute{Description: description, ElementType: types.StringType, Optional: true}
	}
	matcher := func(description string) schema.ListNestedBlock {
		return schema.ListNestedBlock{
			MarkdownDescription: description,
			Validators:          []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"field_paths": stringList("Exact dotted paths, such as `customer.password`. Case-sensitive."),
					"field_names": stringList("Exact field names at any depth, such as `password`. Case-sensitive."),
					"field_name_contains": stringList("Parts of a field name at any depth, such as `secret`. " +
						"Case-insensitive."),
				},
			},
		}
	}
	jsonValue := func(description string) schema.StringAttribute {
		return schema.StringAttribute{
			MarkdownDescription: description + " Write it with `jsonencode()`.",
			Optional:            true,
			Validators:          []validator.String{jsonValidator{}},
		}
	}
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}

	resp.Schema = schema.Schema{
		MarkdownDescription: "An MCP Server exposes a set of commercetools tools to AI agents through the " +
			"Model Context Protocol, on a URL commercetools assigns. Which tools it offers, how they are " +
			"described to the agent, and which fields of their output it filters are configured here.\n\n" +
			"See also the [MCP Servers API documentation](https://docs.commercetools.com/api/projects/mcp-servers).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Platform-generated unique identifier of the MCP Server.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.Int64Attribute{
				Description: "Current version of the MCP Server.",
				Computed:    true,
			},
			"key": schema.StringAttribute{
				Description:   "User-defined unique identifier of the MCP Server. Changing it replaces the server.",
				Required:      true,
				PlanModifiers: replace,
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 256),
					stringvalidator.RegexMatches(
						regexp.MustCompile("^[A-Za-z0-9_-]+$"),
						"Key must match pattern ^[A-Za-z0-9_-]+$"),
				},
			},
			"name": schema.MapAttribute{
				CustomType:  customtypes.NewLocalizedStringType(),
				Description: "Name of the MCP Server as localized string.",
				Optional:    true,
			},
			"description": schema.MapAttribute{
				CustomType:  customtypes.NewLocalizedStringType(),
				Description: "Description of the MCP Server as localized string.",
				Optional:    true,
			},
			"authentication_mode": schema.StringAttribute{
				MarkdownDescription: "How an agent authenticates: `ClientCredentials` with an API client, or " +
					"`CommercetoolsIdentity` as a commercetools user. Defaults to `ClientCredentials`.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("ClientCredentials"),
				Validators: []validator.String{
					stringvalidator.OneOf("ClientCredentials", "CommercetoolsIdentity"),
				},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "`Enabled` or `Disabled`. Defaults to `Disabled`, so a new server takes " +
					"no calls until it is enabled.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("Disabled"),
				Validators: []validator.String{
					stringvalidator.OneOf("Enabled", "Disabled"),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of MCP Server. `CommerceMCP`, the only type, is the default.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(CommerceMCP),
				PlanModifiers:       replace,
				Validators: []validator.String{
					stringvalidator.OneOf(CommerceMCP),
				},
			},
			"major_version": schema.StringAttribute{
				MarkdownDescription: "Major version of the server type, such as `v1` or `v4`. Changing it " +
					"replaces the server.",
				Required:      true,
				PlanModifiers: replace,
			},
			"url": schema.StringAttribute{
				Description: "URL agents connect to, assigned by commercetools.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tools": schema.SetAttribute{
				MarkdownDescription: "Tools the server offers, such as `read_products` or `all`. The `types` " +
					"endpoint of the MCP Servers API lists them per type and major version.",
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"tool_output_formatting": schema.StringAttribute{
				MarkdownDescription: "`JsonToTabular` turns JSON tool output into a table, which takes fewer " +
					"tokens. Left out, output stays JSON.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf("JsonToTabular"),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"tool_customization": schema.ListNestedBlock{
				MarkdownDescription: "How a tool is described to the agent, and what its parameters accept. " +
					"Each tool has at most one.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"tool": schema.StringAttribute{
							Description: "Tool the customization applies to.",
							Required:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the tool shown to the agent, in place of its own.",
							Optional:    true,
						},
					},
					Blocks: map[string]schema.Block{
						"parameter": schema.ListNestedBlock{
							MarkdownDescription: "An override of one of the tool's parameters.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "Name of the parameter in the tool's input schema.",
										Required:    true,
									},
									"description": schema.StringAttribute{
										Description: "Description of the parameter shown to the agent.",
										Optional:    true,
									},
									"const_json": jsonValue("A fixed value for the parameter. Conflicts with " +
										"`default_json` and `allowed_types_json`."),
									"default_json": jsonValue("The value used when the agent leaves the " +
										"parameter out."),
									"allowed_types_json": jsonValue("A JSON array of the values the parameter " +
										"accepts."),
								},
							},
						},
					},
				},
			},
			"json_output_filtering": schema.ListNestedBlock{
				MarkdownDescription: "Fields to redact from or drop out of tool output before the agent sees it.",
				Validators:          []validator.List{listvalidator.SizeAtMost(1)},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"field_paths_to_keep": stringList("Dotted paths kept even when a matcher selects them."),
						"redacted_with": schema.StringAttribute{
							Description: "Text a redacted value is replaced with.",
							Optional:    true,
						},
					},
					Blocks: map[string]schema.Block{
						"redact": matcher("Fields whose value is replaced with `redacted_with`."),
						"drop":   matcher("Fields left out of the output."),
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
	var plan McpServer
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	draft, err := plan.draft()
	if err != nil {
		resp.Diagnostics.AddError("Error creating MCP server", err.Error())
		return
	}
	var res *platform.McpServer
	err = sdk_resource.RetryContext(ctx, 20*time.Second, func() *sdk_resource.RetryError {
		var err error
		res, err = r.client.McpServers().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating MCP server", err.Error())
		return
	}

	current := NewMcpServerFromNative(res, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state McpServer
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.McpServers().WithId(state.ID.ValueString()).Get().Execute(ctx)
	if err != nil {
		if utils.IsResourceNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading MCP server", err.Error())
		return
	}

	current := NewMcpServerFromNative(res, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state McpServer
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updates, err := state.updateActions(plan)
	if err != nil {
		resp.Diagnostics.AddError("Error updating MCP server", err.Error())
		return
	}
	var res *platform.McpServer
	err = sdk_resource.RetryContext(ctx, 5*time.Second, func() *sdk_resource.RetryError {
		var err error
		res, err = r.client.McpServers().WithId(state.ID.ValueString()).Post(updates).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating MCP server", err.Error())
		return
	}

	current := NewMcpServerFromNative(res, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state McpServer
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := sdk_resource.RetryContext(ctx, 5*time.Second, func() *sdk_resource.RetryError {
		_, err := r.client.McpServers().
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
		resp.Diagnostics.AddError("Error deleting MCP server", err.Error())
	}
}

// ImportState imports a MCP server by its id.
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// jsonValidator accepts a string that is valid JSON, as jsonencode() writes.
type jsonValidator struct{}

func (jsonValidator) Description(_ context.Context) string {
	return "must be valid JSON, as jsonencode() writes"
}

func (v jsonValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v jsonValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if !json.Valid([]byte(req.ConfigValue.ValueString())) {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid JSON", v.Description(ctx))
	}
}
