package commercetools

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSubscriptionResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "User-specific unique identifier for the subscription",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"destination": {
				Description: "The Message Queue into which the notifications are to be sent" +
					"See also the [Destination API Docs](https://docs.commercetools.com/api/projects/subscriptions#destination)",
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{},
				},
			},
			"format": {
				Description: "The [format](https://docs.commercetools.com/api/projects/subscriptions#format) " +
					"in which the payload is delivered",
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"changes": {
				Description: "The change notifications subscribed to",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_type_ids": {
							Description: "[Resource Type ID](https://docs.commercetools.com/api/projects/subscriptions#changesubscription)",
							Type:        schema.TypeList,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"message": {
				Description: "The messages subscribed to",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_type_id": {
							Description: "[Resource Type ID](https://docs.commercetools.com/api/projects/subscriptions#changesubscription)",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"types": {
							Description: "types must contain valid message types for this resource, for example for " +
								"resource type product the message type ProductPublished is valid. If no types of " +
								"messages are given, the subscription is valid for all messages of this resource",
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func migrateSubscriptionStateV0toV1(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	transformToList(rawState, "destination")
	transformToList(rawState, "format")
	return rawState, nil
}
