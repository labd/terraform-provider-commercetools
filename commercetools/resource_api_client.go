package commercetools

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
)

func resourceAPIClient() *schema.Resource {
	return &schema.Resource{
		Description: "Create a new API client. Note that Commercetools might return slightly different scopes, " +
			"resulting in a new API client being created everytime Terraform is run. In this case, " +
			"fix your scopes accordingly to match what is returned by Commercetools.\n\n" +
			"Also see the [API client HTTP API documentation](https://docs.commercetools.com//http-api-projects-api-clients).",
		CreateContext: resourceAPIClientCreate,
		ReadContext:   resourceAPIClientRead,
		DeleteContext: resourceAPIClientDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the API client",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"scope": {
				Description: "A list of the [OAuth scopes](https://docs.commercetools.com/http-api-authorization.html#scopes)",
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
				ForceNew:    true,
			},
			"secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceAPIClientCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	scopes := d.Get("scope").(*schema.Set).List()

	scopeParts := make([]string, 0)
	for i := 0; i < len(scopes); i++ {
		scopeParts = append(scopeParts, scopes[i].(string))
	}

	draft := platform.ApiClientDraft{
		Name:  name,
		Scope: strings.Join(scopeParts, " "),
	}

	client := getClient(m)

	var apiClient *platform.ApiClient

	err := resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		var err error

		apiClient, err = client.ApiClients().Post(draft).Execute(ctx)
		if err != nil {
			return handleCommercetoolsError(err)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(apiClient.ID)
	d.Set("secret", apiClient.Secret)

	return resourceAPIClientRead(ctx, d, m)
}

func resourceAPIClientRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)
	apiClient, err := client.ApiClients().WithId(d.Id()).Get().Execute(ctx)

	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	d.SetId(apiClient.ID)
	d.Set("name", apiClient.Name)
	scopes := strings.Split(apiClient.Scope, " ")
	sort.Strings(scopes)
	d.Set("scope", scopes)
	return nil
}

func resourceAPIClientDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)

	_, err := client.ApiClients().WithId(d.Id()).Delete().Execute(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
