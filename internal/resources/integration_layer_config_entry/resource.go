package integration_layer_config_entry

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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

type ConfigEntry struct {
	ID          types.String `tfsdk:"id"`
	Key         types.String `tfsdk:"key"`
	Value       types.String `tfsdk:"value"`
	SecretValue types.String `tfsdk:"secret_value"`
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration_layer_config_entry"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	oneOf := stringvalidator.ExactlyOneOf(path.MatchRoot("value"), path.MatchRoot("secret_value"))
	resp.Schema = schema.Schema{
		MarkdownDescription: "A config entry of the Commerce Integration Layer extension, which its resolvers " +
			"read as `ctx.config.<key>`, such as an API key for a service the extension calls. A change takes " +
			"effect on the next request, without pushing the bundle again.\n\n" +
			"Each entry is its own resource, and writing one leaves every other key alone, including keys set " +
			"with the commercetools CLI.\n\n" +
			"Set `secret_value` for a secret: the Commerce Integration Layer stores it encrypted and never " +
			"returns it. Terraform keeps the configured value in state and cannot detect that it was changed " +
			"outside of Terraform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The key of the entry.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Description: "Name of the entry, as the extension reads it.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`),
						"must be a letter or underscore followed by letters, digits and underscores"),
				},
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "The value, readable by anyone who can read the Project's config. Conflicts " +
					"with `secret_value`.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					// Turning a value into a secret, or back, writes the entry anew.
					stringplanmodifier.RequiresReplaceIf(switchesKind, "", ""),
				},
				Validators: []validator.String{oneOf},
			},
			"secret_value": schema.StringAttribute{
				MarkdownDescription: "The value, stored as a secret that is never returned. Conflicts with " +
					"`value`.",
				Optional:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(switchesKind, "", ""),
				},
				Validators: []validator.String{oneOf},
			},
		},
	}
}

// switchesKind is true when an entry goes from a plain value to a secret or
// back: a changed value of the same kind is written in place.
func switchesKind(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	resp.RequiresReplace = req.StateValue.IsNull() != req.PlanValue.IsNull()
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
	var plan ConfigEntry
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.write(ctx, plan.patch()); err != nil {
		resp.Diagnostics.AddError("Error creating integration layer config entry", err.Error())
		return
	}
	plan.ID = plan.Key
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ConfigEntry
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := r.client.Main().Extension().Config().Get().Execute(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading integration layer config entry", err.Error())
		return
	}
	key := state.ID.ValueString()
	for _, entry := range config.Entries {
		if entry.Key != key {
			continue
		}
		current := ConfigEntry{ID: types.StringValue(key), Key: types.StringValue(key)}
		if entry.Secret {
			// The value is not returned, so the configured one is kept.
			current.Value = types.StringNull()
			current.SecretValue = state.SecretValue
			if current.SecretValue.IsNull() {
				// After an import, the value is unknown but set.
				current.SecretValue = types.StringValue("")
			}
		} else {
			current.Value = plainValue(entry.Value)
			current.SecretValue = types.StringNull()
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ConfigEntry
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.write(ctx, plan.patch()); err != nil {
		resp.Diagnostics.AddError("Error updating integration layer config entry", err.Error())
		return
	}
	plan.ID = plan.Key
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ConfigEntry
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.write(ctx, integrationlayer.ConfigEntryPatch{Key: state.ID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting integration layer config entry", err.Error())
	}
}

// ImportState imports an entry by its key. A secret value cannot be imported;
// it stays as stored until secret_value is applied.
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *Resource) write(ctx context.Context, patch integrationlayer.ConfigEntryPatch) error {
	r.mutex.Lock(utils.IntegrationLayerLock)
	defer r.mutex.Unlock(utils.IntegrationLayerLock)
	_, err := r.client.Main().Extension().Config().Patch(integrationlayer.ExtensionConfigPatch{
		Entries: []integrationlayer.ConfigEntryPatch{patch},
	}).Execute(ctx)
	return err
}

func (e ConfigEntry) patch() integrationlayer.ConfigEntryPatch {
	secret := !e.SecretValue.IsNull()
	value := e.Value.ValueString()
	if secret {
		value = e.SecretValue.ValueString()
	}
	return integrationlayer.ConfigEntryPatch{
		Key:    e.Key.ValueString(),
		Value:  value,
		Secret: &secret,
	}
}

// plainValue reads a plain entry's value, which the API returns as a string,
// and as null for a key set without one.
func plainValue(value any) types.String {
	if text, ok := value.(string); ok {
		return types.StringValue(text)
	}
	return types.StringNull()
}
