package custom_type

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &StateSource{}
	_ datasource.DataSourceWithConfigure = &StateSource{}
)

// NewDataSource is a helper function to simplify the data source implementation.
func NewDataSource() datasource.DataSource {
	return &StateSource{}
}

// StateSource is the data source implementation.
type StateSource struct {
	client *platform.ByProjectKeyRequestBuilder
	mutex  *utils.MutexKV
}

// StateModel maps the data source schema data.
type StateModel struct {
	ID  types.String `tfsdk:"id"`
	Key types.String `tfsdk:"key"`
}

// Metadata returns the data source type name.
func (d *StateSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_state"
}

// Schema defines the schema for the data source.
func (d *StateSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches state information for the given key. " +
			"This is an easy way to import the id of an existing state for a given key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the state",
				Computed:    true,
			},
			"key": schema.StringAttribute{
				Description: "Key of the state",
				Required:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *StateSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data := req.ProviderData.(*utils.ProviderData)
	d.client = data.Client
	d.mutex = data.Mutex
}

// Read refreshes the Terraform state with the latest data.
func (d *StateSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state StateModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resource, err := d.client.States().WithKey(state.Key.ValueString()).Get().Execute(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read state",
			err.Error(),
		)
		return
	}

	state.ID = types.StringValue(resource.ID)

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
