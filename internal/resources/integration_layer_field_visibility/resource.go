package integration_layer_field_visibility

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

const (
	CustomField = "customField"
	Attribute   = "attribute"
)

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client     *integrationlayer.ByProjectKeyRequestBuilder
	projectKey string
	mutex      *utils.MutexKV
}

type FieldVisibility struct {
	ID           types.String  `tfsdk:"id"`
	HiddenFields []HiddenField `tfsdk:"hidden_field"`
}

type HiddenField struct {
	Space     types.String `tfsdk:"space"`
	TypeKey   types.String `tfsdk:"type_key"`
	FieldName types.String `tfsdk:"field_name"`
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration_layer_field_visibility"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	identifier := []validator.String{stringvalidator.LengthBetween(1, 256)}
	resp.Schema = schema.Schema{
		MarkdownDescription: "The custom fields and product attributes the Commerce Integration Layer leaves out " +
			"of its public Experience API schema.\n\n" +
			"A Project has one set of hidden fields, so define this resource once per Project. The set replaces " +
			"whatever was hidden before, including fields hidden in the Merchant Center. Every change republishes " +
			"the schema. When the republish fails the fields are stored but still public: the apply then fails, " +
			"and the next apply retries it. Destroying the resource makes every field public again.\n\n" +
			"A field can be hidden before it exists, so a rule can be applied ahead of the type that adds it.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The key of the Project the rules belong to.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"hidden_field": schema.SetNestedBlock{
				Description: "A field to hide. At most 500.",
				Validators: []validator.Set{
					setvalidator.SizeAtMost(500),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"space": schema.StringAttribute{
							MarkdownDescription: "`customField` for a field of a Custom Type, `attribute` for an " +
								"attribute of a Product Type.",
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf(CustomField, Attribute),
							},
						},
						"type_key": schema.StringAttribute{
							Description: "Key of the Custom Type or Product Type.",
							Required:    true,
							Validators:  identifier,
						},
						"field_name": schema.StringAttribute{
							Description: "Name of the field or attribute.",
							Required:    true,
							Validators:  identifier,
						},
					},
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
	var plan FieldVisibility
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.write(ctx, plan.draft(), "creating")...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(r.projectKey)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FieldVisibility
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	view, err := r.client.Main().Schema().Visibility().Get().Execute(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading integration layer field visibility", err.Error())
		return
	}
	current := NewFieldVisibilityFromNative(r.projectKey, view)
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FieldVisibility
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.write(ctx, plan.draft(), "updating")...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(r.projectKey)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete makes every field public again.
func (r *Resource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(r.write(ctx, nil, "deleting")...)
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// write replaces the hidden fields. A failed republish is an error, so the
// state keeps its previous value and the next apply writes again: an unchanged
// plan would otherwise never retry it, and the fields would stay public.
func (r *Resource) write(ctx context.Context, hidden []integrationlayer.SourceCoordinate, verb string) diag.Diagnostics {
	var diags diag.Diagnostics
	summary := fmt.Sprintf("Error %s integration layer field visibility", verb)

	r.mutex.Lock(utils.IntegrationLayerLock)
	defer r.mutex.Unlock(utils.IntegrationLayerLock)

	result, err := r.client.Main().Schema().Visibility().Put(nonNil(hidden)).Execute(ctx)
	if err != nil {
		diags.AddError(summary, err.Error())
		return diags
	}
	if result.PublishPending {
		diags.AddError(summary, "The hidden fields are stored, but republishing the schema failed, so they "+
			"are still public. Apply again to retry the republish.")
	}
	return diags
}

func (f FieldVisibility) draft() []integrationlayer.SourceCoordinate {
	out := make([]integrationlayer.SourceCoordinate, 0, len(f.HiddenFields))
	for _, field := range f.HiddenFields {
		out = append(out, integrationlayer.SourceCoordinate{
			Space:     integrationlayer.SourceCoordinateSpace(field.Space.ValueString()),
			TypeKey:   field.TypeKey.ValueString(),
			FieldName: field.FieldName.ValueString(),
		})
	}
	return out
}

// NewFieldVisibilityFromNative takes the stored rules from the annotated view,
// which also lists every field that can be hidden. A row carries a rule when
// it is in force or when the API says why it is not.
func NewFieldVisibilityFromNative(projectKey string, view *integrationlayer.FieldVisibilityView) FieldVisibility {
	hidden := []HiddenField{}
	for _, row := range view.Fields {
		if !row.Active && (row.Reason == nil || *row.Reason == "") {
			continue
		}
		hidden = append(hidden, HiddenField{
			Space:     types.StringValue(string(row.Coordinate.Space)),
			TypeKey:   types.StringValue(row.Coordinate.TypeKey),
			FieldName: types.StringValue(row.Coordinate.FieldName),
		})
	}
	return FieldVisibility{ID: types.StringValue(projectKey), HiddenFields: hidden}
}

// nonNil sends an empty list as `[]`, which unhides every field, rather than `null`.
func nonNil(hidden []integrationlayer.SourceCoordinate) []integrationlayer.SourceCoordinate {
	if hidden == nil {
		return []integrationlayer.SourceCoordinate{}
	}
	return hidden
}
