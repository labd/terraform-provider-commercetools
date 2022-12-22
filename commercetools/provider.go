package commercetools

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/ctutils"
	"github.com/labd/commercetools-go-sdk/platform"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown
}

// Provider returns a terraform.ResourceProvider.
func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
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
			},
			ResourcesMap: map[string]*schema.Resource{
				"commercetools_api_client":         resourceAPIClient(),
				"commercetools_api_extension":      resourceAPIExtension(),
				"commercetools_cart_discount":      resourceCartDiscount(),
				"commercetools_channel":            resourceChannel(),
				"commercetools_custom_object":      resourceCustomObject(),
				"commercetools_customer_group":     resourceCustomerGroup(),
				"commercetools_discount_code":      resourceDiscountCode(),
				"commercetools_product_type":       resourceProductType(),
				"commercetools_project_settings":   resourceProjectSettings(),
				"commercetools_shipping_method":    resourceShippingMethod(),
				"commercetools_shipping_zone_rate": resourceShippingZoneRate(),
				"commercetools_shipping_zone":      resourceShippingZone(),
				"commercetools_state":              resourceState(),
				"commercetools_state_transitions":  resourceStateTransitions(),
				"commercetools_store":              resourceStore(),
				"commercetools_tax_category_rate":  resourceTaxCategoryRate(),
				"commercetools_tax_category":       resourceTaxCategory(),
				"commercetools_category":           resourceCategory(),
				"commercetools_type":               resourceType(),
				"commercetools_product_discount":   resourceProductDiscount(),

				// Following items are moved to new terraform-plugin-framework
				// "commercetools_subscription":       resourceSubscription(),
			},
		}
		p.ConfigureContextFunc = providerConfigure(version, p)
		return p
	}
}

func providerConfigure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (any, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
		clientID := d.Get("client_id").(string)
		clientSecret := d.Get("client_secret").(string)
		projectKey := d.Get("project_key").(string)
		scopesRaw := d.Get("scopes").(string)
		apiURL := d.Get("api_url").(string)

		tokenURL, err := url.Parse(d.Get("token_url").(string))
		if err != nil {
			return nil, diag.FromErr(err)

		}
		tokenURL = tokenURL.ResolveReference(&url.URL{Path: "oauth/token"})

		oauth2Config := &clientcredentials.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       strings.Split(scopesRaw, " "),
			TokenURL:     tokenURL.String(),
		}

		httpClient := &http.Client{
			Transport: ctutils.DebugTransport,
		}

		client, err := platform.NewClient(&platform.ClientConfig{
			URL:         apiURL,
			Credentials: oauth2Config,
			UserAgent:   fmt.Sprintf("terraform-provider-commercetools/%s", version),
			HTTPClient:  httpClient,
		})

		if err != nil {
			return nil, diag.FromErr(err)
		}
		return client.WithProjectKey(projectKey), nil
	}
}

// This is a global MutexKV for use within this plugin.
var ctMutexKV = utils.NewMutexKV()
