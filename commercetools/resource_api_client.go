package commercetools

import (
	"sort"
	"strings"

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
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
	scopes := d.Get("scope").(*schema.Set).List()

	scopeParts := make([]string, 0)
	for i := 0; i < len(scopes); i++ {
		scopeParts = append(scopeParts, scopes[i].(string))
	}

	draft := &commercetools.APIClientDraft{
		Name:  name,
		Scope: strings.Join(scopeParts, " "),
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
	scopes := strings.Split(apiClient.Scope, " ")
	sort.Strings(scopes)
	d.Set("scope", scopes)
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
