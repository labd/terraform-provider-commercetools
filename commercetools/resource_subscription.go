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
						},
						"queue_url": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"access_key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"access_secret": {
							Type:     schema.TypeString,
							Optional: false,
						},
						"region": {
							Type:     schema.TypeString,
							Optional: false,
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
	svc := getSubscriptionService(m)
	var subscription *subscriptions.Subscription

	messages := resourceSubscriptionGetMessages(d)
	changes := resourceSubscriptionGetChanges(d)
	destination, err := resourceSubscriptionGetDestination(d)
	if err != nil {
		return err
	}

	draft := &subscriptions.SubscriptionDraft{
		Key:         d.Get("key").(string),
		Destination: destination,
		Messages:    messages,
		Changes:     changes,
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
	svc := getSubscriptionService(m)

	subscription, err := svc.GetByID(d.Id())

	if err != nil {
		if ctErr, ok := err.(commercetools.Error); ok {
			if ctErr.Code() == commercetools.ErrResourceNotFound {
				d.SetId("")
				return nil
			}
		}
		return err
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
		d.Set("changes", subscription.Changes)
	}
	return nil
}

func resourceSubscriptionUpdate(d *schema.ResourceData, m interface{}) error {
	svc := getSubscriptionService(m)

	input := &subscriptions.UpdateInput{
		ID:      d.Id(),
		Version: d.Get("version").(int),
		Actions: commercetools.UpdateActions{},
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&subscriptions.SetKey{Key: newKey})
	}

	if d.HasChange("message") {
		messages := resourceSubscriptionGetMessages(d)
		input.Actions = append(
			input.Actions,
			&subscriptions.SetMessages{Messages: messages})
	}

	if d.HasChange("changes") {
		changes := resourceSubscriptionGetChanges(d)
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
	svc := getSubscriptionService(m)
	version := d.Get("version").(int)
	_, err := svc.DeleteByID(d.Id(), version)
	if err != nil {
		return err
	}

	return nil
}

func getSubscriptionService(m interface{}) *subscriptions.Service {
	client := m.(*commercetools.Client)
	svc := subscriptions.New(client)
	return svc
}

func resourceSubscriptionGetDestination(d *schema.ResourceData) (subscriptions.Destination, error) {
	input := d.Get("destination").(map[string]interface{})

	switch input["type"] {
	case "SQS":
		return subscriptions.DestinationAWSSQS{
			QueueURL:     input["queue_url"].(string),
			AccessKey:    input["access_key"].(string),
			AccessSecret: input["access_secret"].(string),
			Region:       input["region"].(string),
		}, nil
	case "azure_servicebus":
		return subscriptions.DestinationAzureServiceBus{
			ConnectionString: input["connection_string"].(string),
		}, nil
	case "google_pubsub":
		return subscriptions.DestinationGooglePubSub{
			ProjectID: input["project_id"].(string),
			Topic:     input["topic"].(string),
		}, nil
	default:
		return nil, fmt.Errorf("Destination type %s not implemented", input["type"])
	}
}

func resourceSubscriptionGetChanges(d *schema.ResourceData) []subscriptions.ChangeSubscription {
	input := d.Get("changes").([]interface{})
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

func resourceSubscriptionGetMessages(d *schema.ResourceData) []subscriptions.MessageSubscription {
	input := d.Get("message").([]interface{})
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
