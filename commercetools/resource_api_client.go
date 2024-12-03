package commercetools

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func resourceAPIClient() *schema.Resource {
	return &schema.Resource{
		Description: "Create a new API client. Note that Commercetools might return slightly different scopes, " +
			"resulting in a new API client being created everytime Terraform is run. In this case, " +
			"fix your scopes accordingly to match what is returned by Commercetools.\n\n" +
			"Also see the [API client HTTP API documentation](https://docs.commercetools.com/api/projects/api-clients).",
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
				Description: "A list of the [OAuth scopes](https://docs.commercetools.com/api/scopes)",
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
				ForceNew:    true,
			},
			"access_token_validity_seconds": {
				Description: "Expiration time in seconds for each access token obtained by the APIClient. Only present when set with the APIClientDraft. If not present the default value applies.",
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
			},
			"refresh_token_validity_seconds": {
				Description: "Inactivity expiration time in seconds for each refresh token obtained by the APIClient. Only present when set with the APIClientDraft. If not present the default value applies.",
				Type:        schema.TypeInt,
				Optional:    true,
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

func resourceAPIClientCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	scopes := d.Get("scope").(*schema.Set).List()

	scopeParts := make([]string, 0)
	for i := 0; i < len(scopes); i++ {
		scopeParts = append(scopeParts, scopes[i].(string))
	}

	draft := platform.ApiClientDraft{
		Name:  d.Get("name").(string),
		Scope: strings.Join(scopeParts, " "),
	}
	if val := d.Get("access_token_validity_seconds").(int); val != 0 {
		draft.AccessTokenValiditySeconds = &val
	}
	if val := d.Get("refresh_token_validity_seconds").(int); val != 0 {
		draft.RefreshTokenValiditySeconds = &val
	}

	client := getClient(m)

	var apiClient *platform.ApiClient

	err := retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
		var err error
		apiClient, err = client.ApiClients().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(apiClient.ID)
	_ = d.Set("secret", apiClient.Secret)

	return resourceAPIClientRead(ctx, d, m)
}

func resourceAPIClientRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	apiClient, err := client.ApiClients().WithId(d.Id()).Get().Execute(ctx)

	if err != nil {
		if utils.IsResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(apiClient.ID)
	_ = d.Set("name", apiClient.Name)
	scopes := strings.Split(apiClient.Scope, " ")
	sort.Strings(scopes)
	_ = d.Set("scope", scopes)
	_ = d.Set("access_token_validity_seconds", apiClient.AccessTokenValiditySeconds)
	_ = d.Set("refresh_token_validity_seconds", apiClient.RefreshTokenValiditySeconds)
	return nil
}

func resourceAPIClientDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	err := retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
		_, err := client.ApiClients().WithId(d.Id()).Delete().Execute(ctx)
		return utils.ProcessRemoteError(err)
	})

	return diag.FromErr(err)
}
