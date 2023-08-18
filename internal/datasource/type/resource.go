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
	_ datasource.DataSource              = &CustomTypeSource{}
	_ datasource.DataSourceWithConfigure = &CustomTypeSource{}
)

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &CustomTypeSource{}
}

// CustomTypeSource is the data source implementation.
type CustomTypeSource struct {
	client *platform.ByProjectKeyRequestBuilder
	mutex  *utils.MutexKV
}

// CustomTypeSourceModel maps the data source schema data.
type CustomTypeSourceModel struct {
	ID  types.String `tfsdk:"id"`
	Key types.String `tfsdk:"key"`
}

// Metadata returns the data source type name.
func (d *CustomTypeSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_type"
}

// Schema defines the schema for the data source.
func (d *CustomTypeSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches type information",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the custom type",
				Computed:    true,
			},
			"key": schema.StringAttribute{
				Description: "Key of the custom type",
				Required:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *CustomTypeSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data := req.ProviderData.(*utils.ProviderData)
	d.client = data.Client
	d.mutex = data.Mutex
}

// Read refreshes the Terraform state with the latest data.
func (d *CustomTypeSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state CustomTypeSourceModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resource, err := d.client.Types().WithKey(state.Key.ValueString()).Get().Execute(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Type",
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
