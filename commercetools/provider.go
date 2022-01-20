package commercetools

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/commercetools-go-sdk/platform"
	"golang.org/x/oauth2/clientcredentials"
)

var stderr = os.Stderr

func New() tfsdk.Provider {
	return &provider{}
}

type provider struct {
	configured bool
	client     *platform.ByProjectKeyRequestBuilder
	mutex      *MutexKV
}

// Provider returns a terraform.ResourceProvider.
func (p *provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"client_id": {
				Type:        types.StringType,
				Required:    true,
				Description: "The OAuth Client ID for a commercetools platform project. https://docs.commercetools.com/http-api-authorization",
				Sensitive:   true,
			},
			"client_secret": {
				Type:        types.StringType,
				Required:    true,
				Description: "The OAuth Client Secret for a commercetools platform project. https://docs.commercetools.com/http-api-authorization",
				Sensitive:   true,
			},
			"project_key": {
				Type:        types.StringType,
				Required:    true,
				Description: "The project key of commercetools platform project. https://docs.commercetools.com/getting-started",
				Sensitive:   true,
			},
			"scopes": {
				Type:        types.StringType,
				Required:    true,
				Description: "A list as string of OAuth scopes assigned to a project key, to access resources in a commercetools platform project. https://docs.commercetools.com/http-api-authorization",
			},
			"api_url": {
				Type:     types.StringType,
				Required: true,
				// DefaultFunc: schema.EnvDefaultFunc("CTP_API_URL", nil),
				Description: "The API URL of the commercetools platform. https://docs.commercetools.com/http-api",
			},
			"token_url": {
				Type:     types.StringType,
				Required: true,
				// DefaultFunc: schema.EnvDefaultFunc("CTP_AUTH_URL", nil),
				Description: "The authentication URL of the commercetools platform. https://docs.commercetools.com/http-api-authorization",
			},
		},
	}, nil
}

// GetResources - Defines provider resources
func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		// "commercetools_api_client":         resourceAPIClient(),
		// "commercetools_api_extension":      resourceAPIExtension(),
		// "commercetools_cart_discount":      resourceCartDiscount(),
		// "commercetools_channel":            resourceChannel(),
		// "commercetools_custom_object":      resourceCustomObject(),
		// "commercetools_customer_group":     resourceCustomerGroup(),
		// "commercetools_discount_code":      resourceDiscountCode(),
		// "commercetools_product_type":       resourceProductType(),
		"commercetools_project_settings": resourceProjectSettingsType{},
		// "commercetools_shipping_method":    resourceShippingMethod(),
		// "commercetools_shipping_zone_rate": resourceShippingZoneRate(),
		// "commercetools_shipping_zone":      resourceShippingZone(),
		// "commercetools_state":              resourceState(),
		// "commercetools_store":              resourceStore(),
		// "commercetools_subscription":       resourceSubscription(),
		// "commercetools_tax_category_rate":  resourceTaxCategoryRate(),
		// "commercetools_tax_category":       resourceTaxCategory(),
		// "commercetools_category":           resourceCategory(),
		// "commercetools_type":               resourceType(),
	}, nil
}

// GetDataSources - Defines provider data sources
func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{}, nil
}

// Provider schema struct
type providerData struct {
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	ProjectKey   types.String `tfsdk:"project_key"`
	Scopes       types.String `tfsdk:"scopes"`
	ApiURL       types.String `tfsdk:"api_url"`
	TokenURL     types.String `tfsdk:"token_url"`
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	var config providerData

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var clientID string
	if config.ClientID.Null {
		clientID = os.Getenv("CTP_CLIENT_ID")
	} else {
		clientID = config.ClientID.Value
	}

	var clientSecret string
	if config.ClientSecret.Null {
		clientSecret = os.Getenv("CTP_CLIENT_SECRET")
	} else {
		clientSecret = config.ClientSecret.Value
	}

	var projectKey string
	if config.ProjectKey.Null {
		projectKey = os.Getenv("CTP_PROJECT_KEY")
	} else {
		projectKey = config.ProjectKey.Value
	}

	var scopesRaw string
	if config.Scopes.Null {
		scopesRaw = os.Getenv("CTP_SCOPES")
	} else {
		scopesRaw = config.Scopes.Value
	}

	var apiURL string
	if config.ApiURL.Null {
		apiURL = os.Getenv("CTP_API_URL")
	} else {
		apiURL = config.ApiURL.Value
	}

	var authURL string
	if config.TokenURL.Null {
		authURL = os.Getenv("CTP_AUTH_URL")
	} else {
		authURL = config.TokenURL.Value
	}

	oauthScopes := strings.Split(scopesRaw, " ")

	oauth2Config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       oauthScopes,
		TokenURL:     fmt.Sprintf("%s/oauth/token", authURL),
	}

	client, err := platform.NewClient(&platform.ClientConfig{
		URL:         apiURL,
		Credentials: oauth2Config,
		UserAgent:   fmt.Sprintf("%s (terraform-provider-commercetools)", platform.GetUserAgent()),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Unable to create commercetools client:\n\n"+err.Error(),
		)
		return
	}

	p.client = client.WithProjectKey(projectKey)
	p.configured = true
	p.mutex = NewMutexKV()
}
