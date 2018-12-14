package commercetools

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

func resourceAPIClient() *schema.Resource {
	return &schema.Resource{
		Create: resourceAPIClientCreate,
		Read:   resourceAPIClientRead,
		Update: resourceAPIClientUpdate,
		Delete: resourceAPIClientDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"scope": {
				Type:     schema.TypeString,
				Required: true,
			},
			"secret": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAPIClientCreate(d *schema.ResourceData, m interface{}) error {
	name := d.Get("name").(string)
	scope := d.Get("scope").(string)

	draft := &commercetools.APIClientDraft{
		Name:  name,
		Scope: scope,
	}

	client := getClient(m)
	apiClient, err := client.APIClientCreate(draft)
	if err != nil {
		return err
	}

	d.SetId(apiClient.ID)
	d.Set("secret", apiClient.Secret)

	return resourceAPIClientRead(d, m)
}

func resourceAPIClientRead(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	apiClient, err := client.APIClientGetByID(d.Id())

	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.SetId(apiClient.ID)
	d.Set("name", apiClient.Name)
	d.Set("scope", apiClient.Scope)
	return nil
}

func resourceAPIClientUpdate(d *schema.ResourceData, m interface{}) error {
	// not supported
	return nil
}

func resourceAPIClientDelete(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)

	_, err := client.APIClientDelete(d.Id())
	if err != nil {
		return err
	}

	return nil
}
