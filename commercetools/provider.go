package commercetools

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/mutexkv"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/labd/commercetools-go-sdk/commercetools"
	"golang.org/x/oauth2/clientcredentials"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CTP_CLIENT_ID", nil),
				Description: "The OAuth Client ID for a commercetools platform project. https://docs.commercetools.com/http-api-authorization",
				Sensitive:   true,
			},
			"client_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CTP_CLIENT_SECRET", nil),
				Description: "The OAuth Client Secret for a commercetools platform project. https://docs.commercetools.com/http-api-authorization",
				Sensitive:   true,
			},
			"project_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CTP_PROJECT_KEY", nil),
				Description: "The project key of commercetools platform project. https://docs.commercetools.com/getting-started",
				Sensitive:   true,
			},
			"scopes": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CTP_SCOPES", nil),
				Description: "A list as string of OAuth scopes assigned to a project key, to access resources in a commercetools platform project. https://docs.commercetools.com/http-api-authorization",
			},
			"api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CTP_API_URL", nil),
				Description: "The API URL of the commercetools platform. https://docs.commercetools.com/http-api",
				Deprecated:  "Use the region and cloud_provider fields, to let the provider construct the correct hostname.",
			},
			"token_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CTP_AUTH_URL", nil),
				Description: "The authentication URL of the commercetools platform. https://docs.commercetools.com/http-api-authorization",
				Deprecated:  "Use the region and cloud_provider fields, to let the provider construct the correct hostname.",
			},
			"region": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CTP_REGION", nil),
				Description: "The region where the commercetools platform runs, for example 'europe-west1', 'us-central1', etc. https://docs.commercetools.com/http-api.html#regions",
			},
			"cloud_provider": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CTP_CLOUD_PROVIDER", nil),
				Description: "The cloud provider where the commercetools platform runs: 'gcp', 'aws'. https://docs.commercetools.com/http-api.html#regions",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"commercetools_api_client":         resourceAPIClient(),
			"commercetools_api_extension":      resourceAPIExtension(),
			"commercetools_subscription":       resourceSubscription(),
			"commercetools_project_settings":   resourceProjectSettings(),
			"commercetools_type":               resourceType(),
			"commercetools_channel":            resourceChannel(),
			"commercetools_product_type":       resourceProductType(),
			"commercetools_shipping_method":    resourceShippingMethod(),
			"commercetools_shipping_zone_rate": resourceShippingZoneRate(),
			"commercetools_store":              resourceStore(),
			"commercetools_tax_category":       resourceTaxCategory(),
			"commercetools_tax_category_rate":  resourceTaxCategoryRate(),
			"commercetools_shipping_zone":      resourceShippingZone(),
			"commercetools_state":              resourceState(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	clientID := d.Get("client_id").(string)
	clientSecret := d.Get("client_secret").(string)
	projectKey := d.Get("project_key").(string)
	scopesRaw := d.Get("scopes").(string)

	oauthScopes := strings.Split(scopesRaw, " ")
	apiURL := getAPIURL(d)
	authURL := getAuthURL(d)

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

	return client, nil
}

// This is a global MutexKV for use within this plugin.
var ctMutexKV = mutexkv.NewMutexKV()

func getAPIURL(d *schema.ResourceData) string {
	testAPIURL := os.Getenv("CTP_API_URL")
	if testAPIURL != "" {
		return testAPIURL
	}
	return fmt.Sprintf("https://api.%s", getBaseHostname(d))
}

func getAuthURL(d *schema.ResourceData) string {
	testAuthURL := os.Getenv("CTP_AUTH_URL")
	if testAuthURL != "" {
		return testAuthURL
	}
	return fmt.Sprintf("https://auth.%s", getBaseHostname(d))
}

func getBaseHostname(d *schema.ResourceData) string {
	region := d.Get("region").(string)
	cloudProvider := d.Get("cloud_provider").(string)
	return fmt.Sprintf("%s.%s.commercetools.com", region, cloudProvider)
}
