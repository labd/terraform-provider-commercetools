package commercetools

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
	"github.com/labd/commercetools-go-sdk/commercetools/customize"
)

func resourceSubscription() *schema.Resource {
	return &schema.Resource{
		Create: resourceSubscriptionCreate,
		Read:   resourceSubscriptionRead,
		Update: resourceSubscriptionUpdate,
		Delete: resourceSubscriptionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"destination": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"queue_url": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"access_key": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"access_secret": {
							Type:     schema.TypeString,
							Optional: false,
							ForceNew: true,
						},
						"region": {
							Type:     schema.TypeString,
							Optional: false,
							ForceNew: true,
						},
					},
				},
			},
			"changes": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_type_ids": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"message": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_type_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"types": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"version": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceSubscriptionCreate(d *schema.ResourceData, m interface{}) error {
	svc := getCustomizeService(m)
	var subscription *customize.Subscription

	messageInput := d.Get("message").([]interface{})
	messages := resourceSubscriptionMapMessages(messageInput)

	destinationInput := d.Get("destination").(map[string]interface{})
	destination, err := resourceSubscriptionCreateDestination(destinationInput)
	if err != nil {
		return err
	}

	draft := &customize.SubscriptionDraft{
		Key:         d.Get("key").(string),
		Destination: destination,
		Messages:    messages,
	}

	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error

		subscription, err = svc.SubscriptionCreate(draft)
		if err != nil {
			return resource.RetryableError(err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	if subscription == nil {
		log.Fatal("No subscription created?")
	}

	d.SetId(subscription.ID)
	d.Set("version", subscription.Version)

	return resourceSubscriptionRead(d, m)
}

func resourceSubscriptionRead(d *schema.ResourceData, m interface{}) error {
	svc := getCustomizeService(m)

	subscription, err := svc.SubscriptionGetByID(d.Id())

	if err != nil {
		log.Fatalf("Error retrieving subscription: %s", err)
		return nil
	}

	if subscription == nil {
		d.SetId("")
	} else {
		log.Println(subscription)

		d.Set("version", subscription.Version)
		d.Set("key", subscription.Key)
		d.Set("destination", subscription.Destination)
		d.Set("message", subscription.Messages)
	}
	return nil
}

func resourceSubscriptionUpdate(d *schema.ResourceData, m interface{}) error {
	svc := getCustomizeService(m)

	input := &customize.SubscriptionUpdateInput{
		ID:      d.Id(),
		Version: d.Get("version").(int),
		Actions: commercetools.UpdateActions{},
	}

	if d.HasChange("message") {
		messageInput := d.Get("message").([]interface{})
		messages := resourceSubscriptionMapMessages(messageInput)

		input.Actions = append(
			input.Actions,
			&customize.SubscriptionSetMessages{Messages: messages})
	}

	if d.HasChange("changes") {
		changeInput := d.Get("changes").([]interface{})
		changes := resourceSubscriptionMapChanges(changeInput)

		input.Actions = append(
			input.Actions,
			&customize.SubscriptionSetChanges{Changes: changes})
	}

	_, err := svc.SubscriptionUpdate(input)
	if err != nil {
		return err
	}

	return resourceSubscriptionRead(d, m)
}

func resourceSubscriptionDelete(d *schema.ResourceData, m interface{}) error {
	svc := getCustomizeService(m)
	version := d.Get("version").(int)
	svc.SubscriptionDeleteByID(d.Id(), version)

	return nil
}

func getCustomizeService(m interface{}) *customize.Service {
	client := m.(*commercetools.Client)
	svc := customize.New(client)
	return svc
}

func resourceSubscriptionCreateDestination(input map[string]interface{}) (customize.SubscriptionDestination, error) {
	switch input["type"] {
	case "SQS":
		return customize.SubscriptionAWSSQSDestination{
			QueueURL:     input["queue_url"].(string),
			AccessKey:    input["access_key"].(string),
			AccessSecret: input["access_secret"].(string),
			Region:       "eu-west-1",
		}, nil
	default:
		return nil, fmt.Errorf("Destination type %s not implemented", input["type"])
	}
}

func expandStringArray(input []interface{}) []string {
	s := make([]string, len(input))
	for i, v := range input {
		s[i] = fmt.Sprint(v)
	}
	return s
}

func resourceSubscriptionMapChanges(input []interface{}) []customize.ChangeSubscription {
	var result []customize.ChangeSubscription

	for _, raw := range input {
		i := raw.(map[string]interface{})
		rawTypeIds := expandStringArray(i["resource_type_ids"].([]interface{}))

		for _, item := range rawTypeIds {
			result = append(result, customize.ChangeSubscription{
				ResourceTypeID: item,
			})
		}
	}

	return result
}

func resourceSubscriptionMapMessages(input []interface{}) []customize.MessageSubscription {
	var messageObjects []customize.MessageSubscription
	for _, raw := range input {
		i := raw.(map[string]interface{})
		messageObjects = append(messageObjects, customize.MessageSubscription{
			ResourceTypeID: i["resource_type_id"].(string),
			Types:          expandStringArray(i["types"].([]interface{})),
		})
	}

	return messageObjects
}
