package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/checkout"
	"github.com/labd/commercetools-go-sdk/connect"
	"github.com/labd/commercetools-go-sdk/ctutils"
	"github.com/labd/commercetools-go-sdk/insights"
	"github.com/labd/commercetools-go-sdk/integrationlayer"
	"github.com/labd/commercetools-go-sdk/platform"
	"golang.org/x/oauth2/clientcredentials"

	datasourcestate "github.com/labd/terraform-provider-commercetools/internal/datasource/state"
	datasourcetype "github.com/labd/terraform-provider-commercetools/internal/datasource/type"
	"github.com/labd/terraform-provider-commercetools/internal/resources/associate_role"
	"github.com/labd/terraform-provider-commercetools/internal/resources/attribute_group"
	"github.com/labd/terraform-provider-commercetools/internal/resources/business_unit_company"
	"github.com/labd/terraform-provider-commercetools/internal/resources/business_unit_division"
	"github.com/labd/terraform-provider-commercetools/internal/resources/checkout_application"
	"github.com/labd/terraform-provider-commercetools/internal/resources/checkout_payment_integration"
	"github.com/labd/terraform-provider-commercetools/internal/resources/connect_connector"
	"github.com/labd/terraform-provider-commercetools/internal/resources/connect_deployment"
	"github.com/labd/terraform-provider-commercetools/internal/resources/discount_group"
	"github.com/labd/terraform-provider-commercetools/internal/resources/insights_configuration"
	"github.com/labd/terraform-provider-commercetools/internal/resources/integration_layer_allowlist"
	"github.com/labd/terraform-provider-commercetools/internal/resources/integration_layer_checkout_rules"
	"github.com/labd/terraform-provider-commercetools/internal/resources/integration_layer_config_entry"
	"github.com/labd/terraform-provider-commercetools/internal/resources/integration_layer_field_visibility"
	"github.com/labd/terraform-provider-commercetools/internal/resources/integration_layer_idp_login"
	"github.com/labd/terraform-provider-commercetools/internal/resources/integration_layer_session_defaults"
	"github.com/labd/terraform-provider-commercetools/internal/resources/mcp_server"
	"github.com/labd/terraform-provider-commercetools/internal/resources/product_selection"
	"github.com/labd/terraform-provider-commercetools/internal/resources/project"
	"github.com/labd/terraform-provider-commercetools/internal/resources/recurrence_policy"
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

	IntegrationLayerURL types.String `tfsdk:"integration_layer_url"`
	ConnectURL          types.String `tfsdk:"connect_url"`
	CheckoutURL         types.String `tfsdk:"checkout_url"`
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
				MarkdownDescription: "The OAuth Client ID for a commercetools platform project. https://docs.commercetools.com/api/authorization",
				Sensitive:           true,
			},
			"client_secret": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The OAuth Client Secret for a commercetools platform project. https://docs.commercetools.com/api/authorization",
				Sensitive:           true,
			},
			"project_key": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The project key of commercetools platform project. https://docs.commercetools.com/getting-started",
				Sensitive:           true,
			},
			"scopes": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A list as string of OAuth scopes assigned to a project key, to access resources in a commercetools platform project. https://docs.commercetools.com/api/authorization",
			},
			"api_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The API URL of the commercetools platform. https://docs.commercetools.com/api/general-concepts#hosts",
			},
			"token_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The authentication URL of the commercetools platform. https://docs.commercetools.com/api/authorization",
			},
			"connect_url": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "The URL of the commercetools Connect API. Defaults to the `connect` host of " +
					"the region in `api_url`. Only the `commercetools_connect_*` resources use it.",
			},
			"checkout_url": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "The URL of the commercetools Checkout API. Defaults to the `checkout` host of " +
					"the region in `api_url`. Only the `commercetools_checkout_*` resources use it.",
			},
			"integration_layer_url": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "The URL of the Commerce Integration Layer management API. Defaults to the " +
					"`extensions.integration-layer` host of the region in `api_url`. Only the " +
					"`commercetools_integration_layer_*` resources use it.",
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

	// Platform Insights is served by a separate API and therefore needs its own
	// generated client, sharing the same credentials and HTTP client.
	insightsClient, err := insights.NewClient(&insights.ClientConfig{
		URL:         apiURL,
		Credentials: oauth2Config,
		UserAgent:   fmt.Sprintf("terraform-provider-commercetools/%s", p.version),
		HTTPClient:  httpClient,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Unable to create commercetools insights client:\n\n"+err.Error(),
		)
		return
	}

	extra, err := newRegionalClients(config, apiURL, projectKey, oauth2Config, p.version)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create client", err.Error())
		return
	}

	data := &utils.ProviderData{
		Client:                 client.WithProjectKey(projectKey),
		InsightsClient:         insightsClient.WithProjectKey(projectKey),
		IntegrationLayerClient: extra.integrationLayer,
		ConnectClient:          extra.connect,
		ConnectorsClient:       extra.connectors,
		CheckoutClient:         extra.checkout,
		Mutex:                  utils.NewMutexKV(),
		ProjectKey:             projectKey,
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
		discount_group.NewResource,
		recurrence_policy.NewResource,
		mcp_server.NewResource,
		associate_role.NewResource,
		product_selection.NewResource,
		insights_configuration.NewResource,
		checkout_application.NewResource,
		checkout_payment_integration.NewResource,
		connect_connector.NewResource,
		connect_deployment.NewResource,
		integration_layer_session_defaults.NewResource,
		integration_layer_config_entry.NewResource,
		integration_layer_allowlist.NewResource,
		integration_layer_checkout_rules.NewResource,
		integration_layer_field_visibility.NewResource,
		integration_layer_idp_login.NewResource,
		business_unit_company.NewCompanyResource,
		business_unit_division.NewDivisionResource,
	}
}

// regionalURL derives the host of another commercetools API from the platform
// API URL, so https://api.europe-west1.gcp.commercetools.com becomes
// https://connect.europe-west1.gcp.commercetools.com for "connect". It returns
// an empty string for a URL that does not follow that pattern.
func regionalURL(apiURL, service string) string {
	u, err := url.Parse(apiURL)
	if err != nil {
		return ""
	}
	region, ok := strings.CutPrefix(u.Hostname(), "api.")
	if !ok || !strings.HasSuffix(region, ".commercetools.com") {
		return ""
	}
	return fmt.Sprintf("%s://%s.%s", u.Scheme, service, region)
}

type regionalClients struct {
	integrationLayer *integrationlayer.ByProjectKeyRequestBuilder
	connect          *connect.ByProjectKeyRequestBuilder
	connectors       *connect.ConnectorsRequestBuilder
	checkout         *checkout.ByProjectKeyRequestBuilder
}

// newRegionalClients builds the clients for the APIs served by hosts of their
// own: the Commerce Integration Layer and Connect. Each URL comes from the
// provider block, then the environment, then the region of the platform API
// URL. Without an integration layer URL there is no client, and its resources
// say so.
func newRegionalClients(
	config ctProviderModel,
	apiURL, projectKey string,
	credentials *clientcredentials.Config,
	version string,
) (regionalClients, error) {
	userAgent := fmt.Sprintf("terraform-provider-commercetools/%s", version)
	var clients regionalClients

	integrationLayerURL := configuredURL(config.IntegrationLayerURL, "CTP_INTEGRATION_LAYER_URL")
	if integrationLayerURL == "" {
		integrationLayerURL = regionalURL(apiURL, "extensions.integration-layer")
	}
	if integrationLayerURL != "" {
		integrationLayer, err := integrationlayer.NewClient(&integrationlayer.ClientConfig{
			URL:         integrationLayerURL,
			Credentials: credentials,
			UserAgent:   userAgent,
			// Its own client, since NewClient wraps the transport it is given.
			HTTPClient: &http.Client{Transport: ctutils.DebugTransport},
		})
		if err != nil {
			return clients, fmt.Errorf("unable to create commercetools integration layer client: %w", err)
		}
		clients.integrationLayer = integrationLayer.WithProjectKey(projectKey)
	}

	connectURL := configuredURL(config.ConnectURL, "CTP_CONNECT_URL")
	if connectURL == "" {
		connectURL = regionalURL(apiURL, "connect")
	}
	if connectURL == "" {
		connectURL = apiURL
	}
	connectClient, err := connect.NewClient(&connect.ClientConfig{
		URL:         connectURL,
		Credentials: credentials,
		UserAgent:   userAgent,
		HTTPClient:  &http.Client{Transport: ctutils.DebugTransport},
	})
	if err != nil {
		return clients, fmt.Errorf("unable to create commercetools connect client: %w", err)
	}
	clients.connect = connectClient.WithProjectKey(projectKey)
	clients.connectors = connectClient.Connectors()

	checkoutURL := configuredURL(config.CheckoutURL, "CTP_CHECKOUT_URL")
	if checkoutURL == "" {
		checkoutURL = regionalURL(apiURL, "checkout")
	}
	if checkoutURL == "" {
		checkoutURL = apiURL
	}
	checkoutClient, err := checkout.NewClient(&checkout.ClientConfig{
		URL:         checkoutURL,
		Credentials: credentials,
		UserAgent:   userAgent,
		HTTPClient:  &http.Client{Transport: ctutils.DebugTransport},
	})
	if err != nil {
		return clients, fmt.Errorf("unable to create commercetools checkout client: %w", err)
	}
	clients.checkout = checkoutClient.WithProjectKey(projectKey)
	return clients, nil
}

// configuredURL is the provider block's value, or else the environment's.
func configuredURL(value types.String, env string) string {
	if !value.IsUnknown() && !value.IsNull() {
		return value.ValueString()
	}
	return os.Getenv(env)
}
