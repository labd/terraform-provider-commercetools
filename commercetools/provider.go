package commercetools

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/mutexkv"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/labd/commercetools-go-sdk/commercetools"
	"golang.org/x/oauth2/clientcredentials"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"CTP_CLIENT_ID",
				}, nil),
				Description: "CommercesTools Client ID",
			},
			"client_secret": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"CTP_CLIENT_SECRET",
				}, nil),
				Description: "CommercesTools Client Secret",
			},
			"project_key": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"CTP_PROJECT_KEY",
				}, nil),
				Description: "CommercesTools Project key",
			},
			"token_url": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"CTP_AUTH_URL",
				}, "https://auth.sphere.io"),
				Description: "CommercesTools Token URL",
			},
			"api_url": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"CTP_API_URL",
				}, "https://api.sphere.io"),
				Description: "CommercesTools API URL",
			},
			"scopes": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"CTP_SCOPES",
				}, nil),
				Description: "CommercesTools Scopes",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"commercetools_api_client":         resourceAPIClient(),
			"commercetools_api_extension":      resourceAPIExtension(),
			"commercetools_subscription":       resourceSubscription(),
			"commercetools_project_settings":   resourceProjectSettings(),
			"commercetools_type":               resourceType(),
			"commercetools_channel":            resourceChannel(),
			"commercetools_product_discount":   resourceProductDiscount(),
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
	projectKey := d.Get("project_key").(string)

	scopesRaw := d.Get("scopes").(string)
	var scopes []string
	if scopesRaw == "" {
		scopes = []string{fmt.Sprintf("manage_project:%s", projectKey)}
	} else {
		scopes = strings.Split(scopesRaw, " ")
	}

	oauth2Config := &clientcredentials.Config{
		ClientID:     d.Get("client_id").(string),
		ClientSecret: d.Get("client_secret").(string),
		Scopes:       scopes,
		TokenURL:     fmt.Sprintf("%s/oauth/token", d.Get("token_url").(string)),
	}
	httpClient := oauth2Config.Client(context.TODO())

	client := commercetools.New(&commercetools.Config{
		ProjectKey:   projectKey,
		URL:          d.Get("api_url").(string),
		HTTPClient:   httpClient,
		LibraryName:  "terraform-provider-commercetools",
		ContactURL:   "https://labdigital.nl",
		ContactEmail: "opensource@labdigital.nl",
	})

	return client, nil
}

// This is a global MutexKV for use within this plugin.
var ctMutexKV = mutexkv.NewMutexKV()
