package commercetools

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/mutexkv"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/labd/commercetools-go-sdk/commercetools"
	"github.com/machinebox/graphql"
	"golang.org/x/oauth2/clientcredentials"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CTP_CLIENT_ID", nil),
				Description: "The OAuth Client ID for a commercetools platform project. https://docs.commercetools.com/http-api-authorization",
				Sensitive:   true,
			},
			"client_secret": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CTP_CLIENT_SECRET", nil),
				Description: "The OAuth Client Secret for a commercetools platform project. https://docs.commercetools.com/http-api-authorization",
				Sensitive:   true,
			},
			"project_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CTP_PROJECT_KEY", nil),
				Description: "The project key of commercetools platform project. https://docs.commercetools.com/getting-started",
				Sensitive:   true,
			},
			"scopes": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CTP_SCOPES", nil),
				Description: "A list as string of OAuth scopes assigned to a project key, to access resources in a commercetools platform project. https://docs.commercetools.com/http-api-authorization",
			},
			"api_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CTP_API_URL", nil),
				Description: "The API URL of the commercetools platform. https://docs.commercetools.com/http-api",
			},
			"token_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CTP_AUTH_URL", nil),
				Description: "The authentication URL of the commercetools platform. https://docs.commercetools.com/http-api-authorization",
			},
			"mc_api_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CTP_MC_API_URL", nil),
				Description: "The API URL of the Merchant Center. https://docs.commercetools.com/custom-applications/main-concepts/api-gateway#hostnames",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"commercetools_api_client":         resourceAPIClient(),
			"commercetools_api_extension":      resourceAPIExtension(),
			"commercetools_channel":            resourceChannel(),
			"commercetools_custom_application": resourceCustomApplication(),
			"commercetools_product_type":       resourceProductType(),
			"commercetools_project_settings":   resourceProjectSettings(),
			"commercetools_shipping_method":    resourceShippingMethod(),
			"commercetools_shipping_zone":      resourceShippingZone(),
			"commercetools_shipping_zone_rate": resourceShippingZoneRate(),
			"commercetools_state":              resourceState(),
			"commercetools_store":              resourceStore(),
			"commercetools_subscription":       resourceSubscription(),
			"commercetools_tax_category":       resourceTaxCategory(),
			"commercetools_tax_category_rate":  resourceTaxCategoryRate(),
			"commercetools_type":               resourceType(),
		},
		ConfigureFunc: providerConfigure,
	}
}

// TerraformContext holds the HTTP and GraphQL clients to be used by the Terraform resources.
// We recommend to use the utility functions `getClient` and `getGraphQLClient`
// to get the necessary client object.
type TerraformContext struct {
	HTTPClient    *commercetools.Client
	GraphQLClient *graphql.Client
	ProjectKey    string
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	clientID := d.Get("client_id").(string)
	clientSecret := d.Get("client_secret").(string)
	projectKey := d.Get("project_key").(string)
	scopesRaw := d.Get("scopes").(string)
	apiURL := d.Get("api_url").(string)
	authURL := d.Get("token_url").(string)
	mcAPIURL := d.Get("mc_api_url").(string)

	oauthScopes := strings.Split(scopesRaw, " ")

	oauth2Config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       oauthScopes,
		TokenURL:     fmt.Sprintf("%s/oauth/token", authURL),
	}
	httpClient := oauth2Config.Client(context.TODO())

	client := commercetools.New(&commercetools.Config{
		ProjectKey:   projectKey,
		URL:          apiURL,
		HTTPClient:   httpClient,
		LibraryName:  "terraform-provider-commercetools",
		ContactURL:   "https://labdigital.nl",
		ContactEmail: "opensource@labdigital.nl",
	})

	graphqlClient := graphql.NewClient(fmt.Sprintf("%s/graphql", mcAPIURL), graphql.WithHTTPClient(httpClient))

	meta := TerraformContext{
		HTTPClient:    client,
		GraphQLClient: graphqlClient,
		ProjectKey:    projectKey,
	}

	return meta, nil
}

// This is a global MutexKV for use within this plugin.
var ctMutexKV = mutexkv.NewMutexKV()
