package commercetools

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAPIExtension() *schema.Resource {
	return &schema.Resource{
		Create: resourceAPIExtensionCreate,
		Read:   resourceAPIExtensionRead,
		Update: resourceAPIExtensionUpdate,
		Delete: resourceAPIExtensionDelete,

		Schema: map[string]*schema.Schema{

			"destination": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"access_key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"access_secret": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"trigger": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_type_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"actions": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func resourceAPIExtensionCreate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceAPIExtensionRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceAPIExtensionUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceAPIExtensionDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
