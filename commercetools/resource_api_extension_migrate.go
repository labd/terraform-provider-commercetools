package commercetools

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAPIExtensionResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "User-specific unique identifier for the extension",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"destination": {
				Description: "[Destination](https://docs.commercetools.com/api/projects/api-extensions#destination) " +
					"Details where the extension can be reached",
				Type:     schema.TypeSet,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"trigger": {
				Description: "Array of [Trigger](https://docs.commercetools.com/api/projects/api-extensions#trigger) " +
					"Describes what triggers the extension",
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_type_id": {
							Description: "Currently, cart, order, payment, and customer are supported",
							Type:        schema.TypeString,
							Required:    true,
						},
						"actions": {
							Description: "Currently, Create and Update are supported",
							Type:        schema.TypeList,
							Required:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"timeout_in_ms": {
				Description: "Extension timeout in milliseconds",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func migrateAPIExtensionStateV0toV1(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	transformToList(rawState, "destination")
	return rawState, nil
}
