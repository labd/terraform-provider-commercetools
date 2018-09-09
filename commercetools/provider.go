package commercetools

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/labd/commercetools-go-sdk/commercetools"
	"golang.org/x/oauth2/clientcredentials"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"client_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"CTP_CLIENT_ID",
				}, nil),
				Description: "CommercesTools Client ID",
			},
			"client_secret": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"CTP_CLIENT_SECRET",
				}, nil),
				Description: "CommercesTools Client Secret",
			},
			"project_key": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"CTP_PROJECT_KEY",
				}, nil),
				Description: "CommercesTools Project key",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"commercetools_api_extension": resourceAPIExtension(),
			"commercetools_subscription":  resourceSubscription(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	projectKey := d.Get("project_key").(string)

	oauth2Config := &clientcredentials.Config{
		ClientID:     d.Get("client_id").(string),
		ClientSecret: d.Get("client_secret").(string),
		Scopes:       []string{fmt.Sprintf("manage_project:%s", projectKey)},
		TokenURL:     "https://auth.sphere.io/oauth/token",
	}
	httpClient := oauth2Config.Client(context.TODO())

	client := commercetools.New(&commercetools.Config{
		ProjectKey: projectKey,
		URL:        "https://api.sphere.io",
		HTTPClient: httpClient,
	})

	return client, nil
}
