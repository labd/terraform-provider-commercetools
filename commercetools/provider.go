package commercetools

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/labd/commercetools-go-sdk/commercetools"
	"github.com/labd/commercetools-go-sdk/commercetools/credentials"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"client_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"COMMERCETOOLS_CLIENT_ID",
				}, nil),
				Description: "CommercesTools Client ID",
			},
			"client_secret": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"COMMERCETOOLS_CLIENT_SECRET",
				}, nil),
				Description: "CommercesTools Client Secret",
			},
			"project_key": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"COMMERCETOOLS_PROJECT_KEY",
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

	auth := credentials.NewClientCredentialsProvider(
		d.Get("client_id").(string),
		d.Get("client_secret").(string),
		fmt.Sprintf("manage_project:%s", projectKey),
		"https://auth.sphere.io/")

	client, err := commercetools.NewClient(&commercetools.Config{
		ProjectKey:   projectKey,
		ApiURL:       "https://api.sphere.io",
		AuthProvider: auth,
	})

	return client, err
}
