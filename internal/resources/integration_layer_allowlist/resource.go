package integration_layer_allowlist

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/integrationlayer"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client     *integrationlayer.ByProjectKeyRequestBuilder
	projectKey string
	mutex      *utils.MutexKV
}

type Allowlist struct {
	ID    types.String   `tfsdk:"id"`
	Allow []types.String `tfsdk:"allow"`
	// Deny is a set value rather than a slice, since it is unknown until the
	// first write.
	Deny types.Set `tfsdk:"deny"`
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration_layer_allowlist"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The hosts the Commerce Integration Layer extension may call with `fetch`. A host " +
			"is reachable when it matches an `allow` pattern and no `deny` pattern. A request to any other " +
			"host fails at request time, not when the bundle is pushed.\n\n" +
			"A Project has one allowlist, so define this resource once per Project. It replaces the whole list, " +
			"including hosts added with the commercetools CLI. Destroying it empties the list.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The key of the Project the allowlist belongs to.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"allow": schema.SetAttribute{
				MarkdownDescription: "Host patterns: a host such as `cdn.contentful.com`, or a leading " +
					"wildcard such as `*.algolia.net`. No scheme, port or path. The API stores them in " +
					"lowercase, so they have to be written that way.",
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(regexp.MustCompile(`^(\*\.)?[a-z0-9-]+(\.[a-z0-9-]+)+$`),
							"must be a lowercase host, or a host after a leading `*.`"),
					),
				},
			},
			"deny": schema.SetAttribute{
				Description: "Host patterns commercetools does not allow, whatever `allow` says.",
				ElementType: types.StringType,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*utils.ProviderData)
	if !ok {
		return
	}
	r.client = data.IntegrationLayerClient
	r.projectKey = data.ProjectKey
	r.mutex = data.Mutex
	if r.client == nil {
		resp.Diagnostics.AddError("Commerce Integration Layer URL unknown",
			"Set integration_layer_url on the provider, or CTP_INTEGRATION_LAYER_URL, when api_url "+
				"is not a commercetools.com host the URL can be derived from.")
	}
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Allowlist
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	current, err := r.write(ctx, plan.Allow)
	if err != nil {
		resp.Diagnostics.AddError("Error creating integration layer allowlist", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Allowlist
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	res, err := r.client.Main().Extension().Allowlist().Get().Execute(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading integration layer allowlist", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, NewAllowlistFromNative(r.projectKey, res))...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Allowlist
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	current, err := r.write(ctx, plan.Allow)
	if err != nil {
		resp.Diagnostics.AddError("Error updating integration layer allowlist", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
}

func (r *Resource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	if _, err := r.write(ctx, nil); err != nil {
		resp.Diagnostics.AddError("Error deleting integration layer allowlist", err.Error())
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// write replaces the allow patterns. The answer to the write carries no deny
// patterns, so they are read afterwards.
func (r *Resource) write(ctx context.Context, allow []types.String) (Allowlist, error) {
	r.mutex.Lock(utils.IntegrationLayerLock)
	defer r.mutex.Unlock(utils.IntegrationLayerLock)

	patterns := make([]string, 0, len(allow))
	for _, pattern := range allow {
		patterns = append(patterns, pattern.ValueString())
	}
	if _, err := r.client.Main().Extension().Allowlist().Put(patterns).Execute(ctx); err != nil {
		return Allowlist{}, err
	}
	res, err := r.client.Main().Extension().Allowlist().Get().Execute(ctx)
	if err != nil {
		return Allowlist{}, err
	}
	return NewAllowlistFromNative(r.projectKey, res), nil
}

func NewAllowlistFromNative(projectKey string, a *integrationlayer.Allowlist) Allowlist {
	return Allowlist{
		ID:    types.StringValue(projectKey),
		Allow: values(a.Allow),
		Deny:  types.SetValueMust(types.StringType, toAttr(values(a.Deny))),
	}
}

func toAttr(in []types.String) []attr.Value {
	out := make([]attr.Value, 0, len(in))
	for _, v := range in {
		out = append(out, v)
	}
	return out
}

func values(in []string) []types.String {
	out := make([]types.String, 0, len(in))
	for _, v := range in {
		out = append(out, types.StringValue(v))
	}
	return out
}
