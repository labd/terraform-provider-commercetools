package commercetools

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

const (
	subSQS             = "SQS"
	subSNS             = "SNS"
	subAzureServiceBus = "azure_servicebus"
	subGooglePubSub    = "google_pubsub"
)

var destinationFields = map[string][]string{
	subSQS: {
		"queue_url",
		"access_key",
		"access_secret",
		"region",
	},
	subSNS: {
		"topic_arn",
		"access_key",
		"access_secret",
	},
	subAzureServiceBus: {
		"connection_string",
	},
	subGooglePubSub: {
		"project_id",
		"topic",
	},
}

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
				Type:         schema.TypeMap,
				Optional:     true,
				ValidateFunc: validateDestination,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},

						// AWS SNS
						"topic_arn": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressIfNotDestinationType(subSNS),
						},

						// AWS SQS
						"queue_url": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressIfNotDestinationType(subSQS),
						},
						"region": {
							Type:             schema.TypeString,
							Optional:         false,
							DiffSuppressFunc: suppressIfNotDestinationType(subSQS),
						},

						// AWS SNS / SQS
						"access_key": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressIfNotDestinationType(subSQS, subSNS),
						},
						"access_secret": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							DiffSuppressFunc: suppressIfNotDestinationType(subSQS, subSNS),
						},

						// Azure Service Bus
						"connection_string": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							DiffSuppressFunc: suppressIfNotDestinationType(subAzureServiceBus),
						},

						// Google Pub Sub
						"project_id": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressIfNotDestinationType(subGooglePubSub),
						},
						"topic": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressIfNotDestinationType(subGooglePubSub),
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
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceSubscriptionCreate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	var subscription *commercetools.Subscription

	messages := resourceSubscriptionGetMessages(d)
	changes := resourceSubscriptionGetChanges(d)
	destination, err := resourceSubscriptionGetDestination(d)
	if err != nil {
		return err
	}

	draft := &commercetools.SubscriptionDraft{
		Key:         d.Get("key").(string),
		Destination: destination,
		Messages:    messages,
		Changes:     changes,
	}

	err = resource.Retry(20*time.Second, func() *resource.RetryError {
		var err error

		subscription, err = client.SubscriptionCreate(draft)
		if err != nil {
			return resource.RetryableError(err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	if subscription == nil {
		return fmt.Errorf("Error creating subscription")
	}

	d.SetId(subscription.ID)
	d.Set("version", subscription.Version)

	return resourceSubscriptionRead(d, m)
}

func resourceSubscriptionRead(d *schema.ResourceData, m interface{}) error {
	log.Print("[DEBUG] Reading subscriptions from commercetools")
	client := getClient(m)

	subscription, err := client.SubscriptionGetWithID(d.Id())

	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
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
	client := getClient(m)

	input := &commercetools.SubscriptionUpdateWithIDInput{
		ID:      d.Id(),
		Version: d.Get("version").(int),
		Actions: []commercetools.SubscriptionUpdateAction{},
	}

	if d.HasChange("destination") {
		destination, err := resourceSubscriptionGetDestination(d)
		if err != nil {
			return err
		}

		input.Actions = append(
			input.Actions,
			&commercetools.SubscriptionChangeDestinationAction{Destination: destination})
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.SubscriptionSetKeyAction{Key: newKey})
	}

	if d.HasChange("message") {
		messages := resourceSubscriptionGetMessages(d)
		input.Actions = append(
			input.Actions,
			&commercetools.SubscriptionSetMessagesAction{Messages: messages})
	}

	if d.HasChange("changes") {
		changes := resourceSubscriptionGetChanges(d)
		input.Actions = append(
			input.Actions,
			&commercetools.SubscriptionSetChangesAction{Changes: changes})
	}

	_, err := client.SubscriptionUpdateWithID(input)
	if err != nil {
		return err
	}

	return resourceSubscriptionRead(d, m)
}

func resourceSubscriptionDelete(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	version := d.Get("version").(int)
	_, err := client.SubscriptionDeleteWithID(d.Id(), version)
	if err != nil {
		return err
	}

	return nil
}

func resourceSubscriptionGetDestination(d *schema.ResourceData) (commercetools.Destination, error) {
	input := d.Get("destination").(map[string]interface{})

	switch input["type"] {
	case subSNS:
		return commercetools.SnsDestination{
			TopicArn:     input["topic_arn"].(string),
			AccessKey:    input["access_key"].(string),
			AccessSecret: input["access_secret"].(string),
		}, nil
	case subSQS:
		return commercetools.SqsDestination{
			QueueURL:     input["queue_url"].(string),
			AccessKey:    input["access_key"].(string),
			AccessSecret: input["access_secret"].(string),
			Region:       input["region"].(string),
		}, nil
	case subAzureServiceBus:
		return commercetools.AzureServiceBusDestination{
			ConnectionString: input["connection_string"].(string),
		}, nil
	case subGooglePubSub:
		return commercetools.GoogleCloudPubSubDestination{
			ProjectID: input["project_id"].(string),
			Topic:     input["topic"].(string),
		}, nil
	default:
		return nil, fmt.Errorf("Destination type %s not implemented", input["type"])
	}
}

func resourceSubscriptionGetChanges(d *schema.ResourceData) []commercetools.ChangeSubscription {
	input := d.Get("changes").([]interface{})
	var result []commercetools.ChangeSubscription

	for _, raw := range input {
		i := raw.(map[string]interface{})
		rawTypeIds := expandStringArray(i["resource_type_ids"].([]interface{}))

		for _, item := range rawTypeIds {
			result = append(result, commercetools.ChangeSubscription{
				ResourceTypeID: item,
			})
		}
	}

	return result
}

func resourceSubscriptionGetMessages(d *schema.ResourceData) []commercetools.MessageSubscription {
	input := d.Get("message").([]interface{})
	var messageObjects []commercetools.MessageSubscription
	for _, raw := range input {
		i := raw.(map[string]interface{})
		messageObjects = append(messageObjects, commercetools.MessageSubscription{
			ResourceTypeID: i["resource_type_id"].(string),
			Types:          expandStringArray(i["types"].([]interface{})),
		})
	}

	return messageObjects
}

func validateDestination(val interface{}, key string) (warns []string, errs []error) {
	valueAsMap := val.(map[string]interface{})

	destinationType, ok := valueAsMap["type"]

	if !ok {
		errs = append(errs, fmt.Errorf("Property 'type' missing"))
		return warns, errs
	}

	destinationTypeAsString, ok := destinationType.(string)
	if !ok {
		errs = append(errs, fmt.Errorf("Property 'type' has wrong type"))
		return warns, errs
	}
	fields, ok := destinationFields[destinationTypeAsString]
	if !ok {
		errs = append(errs, fmt.Errorf("Property 'type' has invalid value '%v'", destinationTypeAsString))
		return warns, errs
	}

	for _, field := range fields {
		value, ok := valueAsMap[field].(string)
		if !ok {
			errs = append(errs, fmt.Errorf("Required property '%v' missing", field))
		} else if len(value) == 0 {
			errs = append(errs, fmt.Errorf("Required property '%v' is empty", field))
		}
	}

	return warns, errs
}

func suppressIfNotDestinationType(t ...string) schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		input := d.Get("destination").(map[string]interface{})
		for _, val := range t {
			if val == input["type"] {
				return false
			}
		}
		return true
	}
}
