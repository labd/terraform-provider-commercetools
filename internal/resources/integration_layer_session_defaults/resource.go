package integration_layer_session_defaults

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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

type SessionDefaults struct {
	ID       types.String `tfsdk:"id"`
	Label    types.String `tfsdk:"label"`
	Language types.String `tfsdk:"language"`
	Currency types.String `tfsdk:"currency"`
	Country  types.String `tfsdk:"country"`
	RegionID types.String `tfsdk:"region_id"`
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration_layer_session_defaults"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The language, currency and country a new shopper session on the Commerce " +
			"Integration Layer starts with, and the label the Project is shown with. A storefront can choose " +
			"others for a session; these apply when it does not.\n\n" +
			"They are picked from what the commercetools Project offers, which " +
			"`commercetools_project_settings` defines, and spelled as the Project spells them (`DE`, not " +
			"`de`). The API stores any value and a wrong one only fails once a session uses it, so the " +
			"provider checks them before writing.\n\n" +
			"The Project has to be enrolled in the Commerce Integration Layer already, and the API client needs " +
			"the `manage_project` scope. A Project has one set of session defaults, so define this resource " +
			"once per Project. They cannot be removed: destroying the resource only removes it from the state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The key of the Project the defaults belong to.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"label": schema.StringAttribute{
				Description: "Display name of the Project on the Commerce Integration Layer.",
				Required:    true,
			},
			"language": schema.StringAttribute{
				Description: "Language a session starts with, as a BCP 47 tag the Project offers, such as `en-US`.",
				Required:    true,
			},
			"currency": schema.StringAttribute{
				Description: "Currency a session starts with, as an ISO 4217 code the Project offers, such as `EUR`.",
				Required:    true,
			},
			"country": schema.StringAttribute{
				Description: "Country a session starts with, as an ISO 3166-1 alpha-2 code the Project offers, such as `DE`.",
				Required:    true,
			},
			"region_id": schema.StringAttribute{
				Description: "Region the Commerce Integration Layer serves the Project from. Set by commercetools.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
	var plan SessionDefaults
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.write(ctx, &plan, "creating")...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SessionDefaults
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	view, err := r.client.Main().Settings().Get().Execute(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading integration layer session defaults", err.Error())
		return
	}

	current := SessionDefaults{
		ID:       types.StringValue(r.projectKey),
		Label:    types.StringValue(valueOf(view.Project.Label)),
		Language: types.StringValue(valueOf(view.Project.Language)),
		Currency: types.StringValue(valueOf(view.Project.Currency)),
		Country:  types.StringValue(valueOf(view.Project.Country)),
		RegionID: types.StringValue(view.Platform.RegionId),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SessionDefaults
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.write(ctx, &plan, "updating")...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete only removes the defaults from the state: the API has no way to
// unset them, and all four are required.
func (r *Resource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning(
		"Integration layer session defaults are kept",
		"The Commerce Integration Layer cannot remove a Project's session defaults. They are removed from the "+
			"Terraform state only, and stay in force on the Project.",
	)
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// write checks the planned values against what the Project presents, then
// replaces the defaults, and fills in the computed attributes of plan.
func (r *Resource) write(ctx context.Context, plan *SessionDefaults, verb string) diag.Diagnostics {
	var diags diag.Diagnostics
	summary := fmt.Sprintf("Error %s integration layer session defaults", verb)

	r.mutex.Lock(utils.IntegrationLayerLock)
	defer r.mutex.Unlock(utils.IntegrationLayerLock)

	view, err := r.client.Main().Settings().Get().Execute(ctx)
	if err != nil {
		diags.AddError(summary, err.Error())
		return diags
	}
	if view.Presentment != nil {
		diags.Append(presentable(path.Root("language"), plan.Language.ValueString(), view.Presentment.Languages)...)
		diags.Append(presentable(path.Root("currency"), plan.Currency.ValueString(), view.Presentment.Currencies)...)
		diags.Append(presentable(path.Root("country"), plan.Country.ValueString(), view.Presentment.Countries)...)
		if diags.HasError() {
			return diags
		}
	}

	_, err = r.client.Main().Settings().Put(integrationlayer.ProjectSettings{
		Label:    plan.Label.ValueString(),
		Language: plan.Language.ValueString(),
		Currency: plan.Currency.ValueString(),
		Country:  plan.Country.ValueString(),
	}).Execute(ctx)
	if err != nil {
		diags.AddError(summary, err.Error())
		return diags
	}

	plan.ID = types.StringValue(r.projectKey)
	plan.RegionID = types.StringValue(view.Platform.RegionId)
	return diags
}

// presentable reports a value the Project cannot present. A value that only
// differs in case is reported with the spelling to use, since the API would
// store it as given and the Experience API would then not match it.
func presentable(at path.Path, value string, allowed []string) diag.Diagnostics {
	var diags diag.Diagnostics
	if slices.Contains(allowed, value) {
		return diags
	}
	for _, candidate := range allowed {
		if strings.EqualFold(candidate, value) {
			diags.AddAttributeError(at, "Value spelled differently by the Project",
				fmt.Sprintf("The Project spells %q as %q.", value, candidate))
			return diags
		}
	}
	diags.AddAttributeError(at, "Value not presented by the Project",
		fmt.Sprintf("%q is not one of the Project's values: %s.", value, strings.Join(allowed, ", ")))
	return diags
}

// valueOf reads a setting a Project created before it existed may lack.
func valueOf(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
