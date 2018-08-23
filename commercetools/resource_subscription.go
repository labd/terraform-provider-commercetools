package commercetools

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
	"github.com/labd/commercetools-go-sdk/service/subscriptions"
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
	var subscription *subscriptions.Subscription

	messageInput := d.Get("message").([]interface{})
	messages := resourceSubscriptionMapMessages(messageInput)

	destinationInput := d.Get("destination").(map[string]interface{})
	destination, err := resourceSubscriptionCreateDestination(destinationInput)
	if err != nil {
		return err
	}

	draft := &subscriptions.SubscriptionDraft{
		Key:         d.Get("key").(string),
		Destination: destination,
		Messages:    messages,
	}

	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error

		subscription, err = svc.Create(draft)
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
	log.Print("[DEBUG] Reading subscriptions from commercetools")
	svc := getCustomizeService(m)

	subscription, err := svc.GetByID(d.Id())

	if err != nil {
		log.Fatalf("Error retrieving subscription: %s", err)
		return nil
	}

	if subscription == nil {
		log.Print("[DEBUG] No subscriptions found")
		d.SetId("")
	} else {
		log.Print("[DEBUG] Found following subscriptions:")
		log.Print(stringFormatObject(subscription))

		d.Set("version", subscription.Version)
		d.Set("key", subscription.Key)
		d.Set("destination", subscription.Destination)
		d.Set("message", subscription.Messages)
	}
	return nil
}

func resourceSubscriptionUpdate(d *schema.ResourceData, m interface{}) error {
	svc := getCustomizeService(m)

	input := &subscriptions.UpdateInput{
		ID:      d.Id(),
		Version: d.Get("version").(int),
		Actions: commercetools.UpdateActions{},
	}

	if d.HasChange("message") {
		messageInput := d.Get("message").([]interface{})
		messages := resourceSubscriptionMapMessages(messageInput)

		input.Actions = append(
			input.Actions,
			&subscriptions.SetMessages{Messages: messages})
	}

	if d.HasChange("changes") {
		changeInput := d.Get("changes").([]interface{})
		changes := resourceSubscriptionMapChanges(changeInput)

		input.Actions = append(
			input.Actions,
			&subscriptions.SetChanges{Changes: changes})
	}

	_, err := svc.Update(input)
	if err != nil {
		return err
	}

	return resourceSubscriptionRead(d, m)
}

func resourceSubscriptionDelete(d *schema.ResourceData, m interface{}) error {
	svc := getCustomizeService(m)
	version := d.Get("version").(int)
	svc.DeleteByID(d.Id(), version)

	return nil
}

func getCustomizeService(m interface{}) *subscriptions.Service {
	client := m.(*commercetools.Client)
	svc := subscriptions.New(client)
	return svc
}

func resourceSubscriptionCreateDestination(input map[string]interface{}) (subscriptions.Destination, error) {
	switch input["type"] {
	case "SQS":
		return subscriptions.DestinationAWSSQS{
			QueueURL:     input["queue_url"].(string),
			AccessKey:    input["access_key"].(string),
			AccessSecret: input["access_secret"].(string),
			Region:       input["region"].(string),
		}, nil
	default:
		return nil, fmt.Errorf("Destination type %s not implemented", input["type"])
	}
}

func resourceSubscriptionMapChanges(input []interface{}) []subscriptions.ChangeSubscription {
	var result []subscriptions.ChangeSubscription

	for _, raw := range input {
		i := raw.(map[string]interface{})
		rawTypeIds := expandStringArray(i["resource_type_ids"].([]interface{}))

		for _, item := range rawTypeIds {
			result = append(result, subscriptions.ChangeSubscription{
				ResourceTypeID: item,
			})
		}
	}

	return result
}

func resourceSubscriptionMapMessages(input []interface{}) []subscriptions.MessageSubscription {
	var messageObjects []subscriptions.MessageSubscription
	for _, raw := range input {
		i := raw.(map[string]interface{})
		messageObjects = append(messageObjects, subscriptions.MessageSubscription{
			ResourceTypeID: i["resource_type_id"].(string),
			Types:          expandStringArray(i["types"].([]interface{})),
		})
	}

	return messageObjects
}
