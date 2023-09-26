package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/ctutils"
	"github.com/labd/commercetools-go-sdk/platform"
	"golang.org/x/oauth2/clientcredentials"

	datasourcestate "github.com/labd/terraform-provider-commercetools/internal/datasource/state"
	datasourcetype "github.com/labd/terraform-provider-commercetools/internal/datasource/type"
	"github.com/labd/terraform-provider-commercetools/internal/resources/associate_role"
	"github.com/labd/terraform-provider-commercetools/internal/resources/attribute_group"
	"github.com/labd/terraform-provider-commercetools/internal/resources/product_selection"
	"github.com/labd/terraform-provider-commercetools/internal/resources/project"
	"github.com/labd/terraform-provider-commercetools/internal/resources/state"
	"github.com/labd/terraform-provider-commercetools/internal/resources/state_transition"
	"github.com/labd/terraform-provider-commercetools/internal/resources/subscription"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &ctProvider{}
)

func New(version string) provider.Provider {
	return &ctProvider{
		version: version,
	}
}

type ctProvider struct {
	version string
}

// Provider schema struct
type ctProviderModel struct {
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	ProjectKey   types.String `tfsdk:"project_key"`
	Scopes       types.String `tfsdk:"scopes"`
	ApiURL       types.String `tfsdk:"api_url"`
	TokenURL     types.String `tfsdk:"token_url"`
}

// Metadata returns the provider type name.
func (p *ctProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "commercetools"
}

// Provider returns a terraform.ResourceProvider.
func (p *ctProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The OAuth Client ID for a commercetools platform project. https://docs.commercetools.com/http-api-authorization",
				Sensitive:           true,
			},
			"client_secret": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The OAuth Client Secret for a commercetools platform project. https://docs.commercetools.com/http-api-authorization",
				Sensitive:           true,
			},
			"project_key": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The project key of commercetools platform project. https://docs.commercetools.com/getting-started",
				Sensitive:           true,
			},
			"scopes": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A list as string of OAuth scopes assigned to a project key, to access resources in a commercetools platform project. https://docs.commercetools.com/http-api-authorization",
			},
			"api_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The API URL of the commercetools platform. https://docs.commercetools.com/http-api",
			},
			"token_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The authentication URL of the commercetools platform. https://docs.commercetools.com/http-api-authorization",
			},
		},
	}
}

func (p *ctProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ctProviderModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var clientID string
	if config.ClientID.IsUnknown() || config.ClientID.IsNull() {
		clientID = os.Getenv("CTP_CLIENT_ID")
	} else {
		clientID = config.ClientID.ValueString()
	}

	var clientSecret string
	if config.ClientSecret.IsUnknown() || config.ClientSecret.IsNull() {
		clientSecret = os.Getenv("CTP_CLIENT_SECRET")
	} else {
		clientSecret = config.ClientSecret.ValueString()
	}

	var projectKey string
	if config.ProjectKey.IsUnknown() || config.ProjectKey.IsNull() {
		projectKey = os.Getenv("CTP_PROJECT_KEY")
	} else {
		projectKey = config.ProjectKey.ValueString()
	}

	var scopesRaw string
	if config.Scopes.IsUnknown() || config.Scopes.IsNull() {
		scopesRaw = os.Getenv("CTP_SCOPES")
	} else {
		scopesRaw = config.Scopes.ValueString()
	}

	var apiURL string
	if config.ApiURL.IsUnknown() || config.ApiURL.IsNull() {
		apiURL = os.Getenv("CTP_API_URL")
	} else {
		apiURL = config.ApiURL.ValueString()
	}

	var authURL string
	if config.TokenURL.IsUnknown() || config.TokenURL.IsNull() {
		authURL = os.Getenv("CTP_AUTH_URL")
	} else {
		authURL = config.TokenURL.ValueString()
	}

	oauthScopes := strings.Split(scopesRaw, " ")
	oauth2Config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       oauthScopes,
		TokenURL:     fmt.Sprintf("%s/oauth/token", authURL),
	}

	httpClient := &http.Client{
		Transport: ctutils.DebugTransport,
	}

	client, err := platform.NewClient(&platform.ClientConfig{
		URL:         apiURL,
		Credentials: oauth2Config,
		UserAgent:   fmt.Sprintf("terraform-provider-commercetools/%s", p.version),
		HTTPClient:  httpClient,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Unable to create commercetools client:\n\n"+err.Error(),
		)
		return
	}

	data := &utils.ProviderData{
		Client: client.WithProjectKey(projectKey),
		Mutex:  utils.NewMutexKV(),
	}
	resp.DataSourceData = data
	resp.ResourceData = data
}

// DataSources defines the data sources implemented in the provider.
func (p *ctProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasourcetype.NewDataSource,
		datasourcestate.NewDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *ctProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		subscription.NewSubscriptionResource,
		project.NewResource,
		state.NewResource,
		state_transition.NewResource,
		attribute_group.NewResource,
		associate_role.NewResource,
		product_selection.NewResource,
	}
}
