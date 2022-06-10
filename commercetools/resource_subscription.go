package commercetools

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
)

const (
	// Destinations
	subSQS                  = "SQS"
	subSNS                  = "SNS"
	subEventBridge          = "event_bridge"
	subEventBridgeAlias     = "EventBridge"
	subAzureEventGrid       = "azure_eventgrid"
	subAzureEventGridAlias  = "EventGrid"
	subAzureServiceBus      = "azure_servicebus"
	subAzureServiceBusAlias = "AzureServiceBus"
	subGooglePubSub         = "google_pubsub"
	subGooglePubSubAlias    = "GoogleCloudPubSub"

	// Formats
	cloudEvents = "cloud_events"
	fmtPlatform = "platform"
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
	subEventBridge: {
		"region",
		"account_id",
	},
	subAzureEventGrid: {
		"uri",
		"access_key",
	},
	subAzureServiceBus: {
		"connection_string",
	},
	subGooglePubSub: {
		"project_id",
		"topic",
	},
}

var destinationFieldAliases = map[string]string{
	subEventBridgeAlias:     subEventBridge,
	subAzureEventGridAlias:  subAzureEventGrid,
	subAzureServiceBusAlias: subAzureServiceBus,
	subGooglePubSubAlias:    subGooglePubSub,
}

var formatFields = map[string][]string{
	cloudEvents: {
		"cloud_events_version",
	},
	fmtPlatform: {},
}

func resourceSubscription() *schema.Resource {
	return &schema.Resource{
		Description: "Subscriptions allow you to be notified of new messages or changes via a Message Queue of your " +
			"choice. Subscriptions are used to trigger an asynchronous background process in response to an event on " +
			"the commercetools platform. Common use cases include sending an Order Confirmation Email, charging a " +
			"Credit Card after the delivery has been made, or synchronizing customer accounts to a Customer " +
			"Relationship Management (CRM) system.\n\n" +
			"See also the [Subscriptions API Documentation](https://docs.commercetools.com/api/projects/subscriptions)",
		CreateContext: resourceSubscriptionCreate,
		ReadContext:   resourceSubscriptionRead,
		UpdateContext: resourceSubscriptionUpdate,
		DeleteContext: resourceSubscriptionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceSubscriptionResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: migrateSubscriptionStateV0toV1,
				Version: 0,
			},
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "User-specific unique identifier for the subscription",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"destination": {
				Description: "The Message Queue into which the notifications are to be sent" +
					"See also the [Destination API Docs](https://docs.commercetools.com/api/projects/subscriptions#destination)",
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: func(d interface{}, v string) ([]string, []error) {
								allowed := []string{
									subSQS,
									subSNS,
									subEventBridge,
									subEventBridgeAlias,
									subAzureEventGrid,
									subAzureEventGridAlias,
									subAzureServiceBus,
									subAzureServiceBusAlias,
									subGooglePubSub,
									subGooglePubSubAlias,
								}

								if !stringInSlice(d.(string), allowed) {
									return []string{}, []error{
										fmt.Errorf("invalid destination type %s. Accepted are %s",
											d.(string), strings.Join(allowed, ", "),
										),
									}
								}

								return []string{}, []error{}
							},
						},
						"topic_arn": {
							Description:      "For AWS SNS",
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressIfNotDestinationType(subSNS),
						},
						"queue_url": {
							Description:      "For AWS SQS",
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressIfNotDestinationType(subSQS),
						},
						"region": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressIfNotDestinationType(subSQS, subEventBridge, subEventBridgeAlias),
						},
						"account_id": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressIfNotDestinationType(subEventBridge, subEventBridgeAlias),
						},
						"access_key": {
							Description:      "For AWS SNS / SQS / Azure Event Grid",
							Type:             schema.TypeString,
							Optional:         true,
							Sensitive:        true,
							DiffSuppressFunc: suppressIfNotDestinationType(subSQS, subSNS, subAzureEventGrid, subAzureEventGridAlias),
						},
						"access_secret": {
							Description:      "For AWS SNS / SQS",
							Type:             schema.TypeString,
							Optional:         true,
							Sensitive:        true,
							DiffSuppressFunc: suppressIfNotDestinationType(subSQS, subSNS),
						},
						"uri": {
							Description:      "For Azure Event Grid",
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressIfNotDestinationType(subAzureEventGrid, subAzureEventGridAlias),
						},
						"connection_string": {
							Description:      "For Azure Service Bus",
							Type:             schema.TypeString,
							Optional:         true,
							Sensitive:        true,
							ForceNew:         true,
							DiffSuppressFunc: suppressIfNotDestinationType(subAzureServiceBus, subAzureServiceBusAlias),
						},
						"project_id": {
							Description:      "For Google Pub Sub",
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressIfNotDestinationType(subGooglePubSub, subGooglePubSubAlias),
						},
						"topic": {
							Description:      "For Google Pub Sub",
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressIfNotDestinationType(subGooglePubSub, subGooglePubSubAlias),
						},
					},
				},
			},
			"format": {
				Description: "The [format](https://docs.commercetools.com/api/projects/subscriptions#format) " +
					"in which the payload is delivered",
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if k == "format.#" && old == "1" && new == "0" {
						fmt := d.Get("format.0.type").(string)
						if strings.ToLower(fmt) == "platform" {
							return true
						}
					}
					if k == "format.0.type" && strings.ToLower(old) == "platform" && new == "" {
						return true
					}

					return false
				},
				DefaultFunc: func() (interface{}, error) {
					return []map[string]string{{
						"type": "Platform",
					}}, nil
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"cloud_events_version": {
							Description:      "For CloudEvents",
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressIfNotFormatType(cloudEvents),
						},
					},
				},
			},
			"changes": {
				Description: "The change notifications subscribed to",
				Type:        schema.TypeList,
				Optional:    true,
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

func resourceSubscriptionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)
	var subscription *platform.Subscription

	if err := validateSubscriptionDestination(d); err != nil {
		return diag.FromErr(err)
	}
	if err := validateFormat(d); err != nil {
		return diag.FromErr(err)
	}

	messages := unmarshallSubscriptionMessages(d)
	changes := unmarshallSubscriptionChanges(d)
	destination, err := unmarshallSubscriptionDestination(d)
	if err != nil {
		return diag.FromErr(err)
	}
	format, err := unmarshallSubscriptionFormat(d)
	if err != nil {
		return diag.FromErr(err)
	}

	draft := platform.SubscriptionDraft{
		Destination: destination,
		Format:      &format,
		Messages:    messages,
		Changes:     changes,
	}

	key := stringRef(d.Get("key"))
	if *key != "" {
		draft.Key = key
	}

	err = resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		var err error
		subscription, err = client.Subscriptions().Post(draft).Execute(ctx)
		return processRemoteError(err)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(subscription.ID)
	d.Set("version", subscription.Version)

	return resourceSubscriptionRead(ctx, d, m)
}

func resourceSubscriptionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Print("[DEBUG] Reading subscriptions from commercetools")
	client := getClient(m)

	subscription, err := client.Subscriptions().WithId(d.Id()).Get().Execute(ctx)
	if err != nil {
		if IsResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if subscription == nil {
		log.Print("[DEBUG] No subscriptions found")
		d.SetId("")
	} else {
		log.Print("[DEBUG] Found following subscriptions:")
		log.Print(stringFormatObject(subscription))

		d.Set("version", subscription.Version)
		d.Set("key", subscription.Key)
		d.Set("destination", marshallSubscriptionDestination(subscription.Destination, d))
		d.Set("format", marshallSubscriptionFormat(subscription.Format))
		d.Set("message", marshallSubscriptionMessages(subscription.Messages))
		d.Set("changes", marshallSubscriptionChanges(subscription.Changes))
	}
	return nil
}

func resourceSubscriptionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)

	if err := validateSubscriptionDestination(d); err != nil {
		return diag.FromErr(err)
	}
	if err := validateFormat(d); err != nil {
		return diag.FromErr(err)
	}

	input := platform.SubscriptionUpdate{
		Version: d.Get("version").(int),
		Actions: []platform.SubscriptionUpdateAction{},
	}

	if d.HasChange("destination") {
		destination, err := unmarshallSubscriptionDestination(d)
		if err != nil {
			return diag.FromErr(err)
		}

		input.Actions = append(
			input.Actions,
			&platform.SubscriptionChangeDestinationAction{Destination: destination})
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&platform.SubscriptionSetKeyAction{Key: &newKey})
	}

	if d.HasChange("message") {
		messages := unmarshallSubscriptionMessages(d)
		input.Actions = append(
			input.Actions,
			&platform.SubscriptionSetMessagesAction{Messages: messages})
	}

	if d.HasChange("changes") {
		changes := unmarshallSubscriptionChanges(d)
		input.Actions = append(
			input.Actions,
			&platform.SubscriptionSetChangesAction{Changes: changes})
	}

	err := resource.RetryContext(ctx, 5*time.Second, func() *resource.RetryError {
		_, err := client.Subscriptions().WithId(d.Id()).Post(input).Execute(ctx)
		return processRemoteError(err)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSubscriptionRead(ctx, d, m)
}

func resourceSubscriptionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)
	version := d.Get("version").(int)
	err := resource.RetryContext(ctx, 5*time.Second, func() *resource.RetryError {
		_, err := client.Subscriptions().WithId(d.Id()).Delete().Version(version).Execute(ctx)
		return processRemoteError(err)
	})
	return diag.FromErr(err)
}

func unmarshallSubscriptionDestination(d *schema.ResourceData) (platform.Destination, error) {
	dst, err := elementFromList(d, "destination")
	if err != nil {
		return nil, err
	}
	if dst == nil {
		return nil, fmt.Errorf("destination is missing")
	}

	switch dst["type"] {
	case subSNS:
		return platform.SnsDestination{
			TopicArn:     dst["topic_arn"].(string),
			AccessKey:    dst["access_key"].(string),
			AccessSecret: dst["access_secret"].(string),
		}, nil
	case subSQS:
		return platform.SqsDestination{
			QueueUrl:     dst["queue_url"].(string),
			AccessKey:    dst["access_key"].(string),
			AccessSecret: dst["access_secret"].(string),
			Region:       dst["region"].(string),
		}, nil
	case subAzureEventGrid, subAzureEventGridAlias:
		return platform.AzureEventGridDestination{
			Uri:       dst["uri"].(string),
			AccessKey: dst["access_key"].(string),
		}, nil
	case subAzureServiceBus, subAzureServiceBusAlias:
		return platform.AzureServiceBusDestination{
			ConnectionString: dst["connection_string"].(string),
		}, nil
	case subGooglePubSub, subGooglePubSubAlias:
		return platform.GoogleCloudPubSubDestination{
			ProjectId: dst["project_id"].(string),
			Topic:     dst["topic"].(string),
		}, nil
	case subEventBridge, subEventBridgeAlias:
		return platform.EventBridgeDestination{
			Region:    dst["region"].(string),
			AccountId: dst["account_id"].(string),
		}, nil
	default:
		return nil, fmt.Errorf("destination type %s not implemented", dst["type"])
	}
}

func marshallSubscriptionDestination(dst platform.Destination, d *schema.ResourceData) []map[string]string {

	// Read the access secret from the current resource data
	c, _ := unmarshallSubscriptionDestination(d)
	accessSecret := ""
	switch current := c.(type) {
	case platform.SnsDestination:
		accessSecret = current.AccessSecret
	case platform.SqsDestination:
		accessSecret = current.AccessSecret
	}

	switch v := dst.(type) {
	case platform.SnsDestination:
		d.Get("destination")
		return []map[string]string{{
			"type":          subSNS,
			"topic_arn":     v.TopicArn,
			"access_key":    v.AccessKey,
			"access_secret": accessSecret,
		}}
	case platform.SqsDestination:
		return []map[string]string{{
			"type":          subSQS,
			"queue_url":     v.QueueUrl,
			"access_key":    v.AccessKey,
			"access_secret": accessSecret,
			"region":        v.Region,
		}}
	case platform.AzureEventGridDestination:
		return []map[string]string{{
			"type":       subAzureEventGrid,
			"uri":        v.Uri,
			"access_key": v.AccessKey,
		}}
	case platform.AzureServiceBusDestination:
		return []map[string]string{{
			"type":              subAzureServiceBus,
			"connection_string": v.ConnectionString,
		}}
	case platform.GoogleCloudPubSubDestination:
		return []map[string]string{{
			"type":       subGooglePubSub,
			"project_id": v.ProjectId,
			"topic":      v.Topic,
		}}
	case platform.EventBridgeDestination:
		return []map[string]string{{
			"type":       subEventBridge,
			"region":     v.Region,
			"account_id": v.AccountId,
		}}
	}
	return []map[string]string{}
}

func marshallSubscriptionFormat(f platform.DeliveryFormat) []map[string]string {
	switch v := f.(type) {
	case platform.PlatformFormat:
		return []map[string]string{{
			"type": "Platform",
		}}
	case platform.CloudEventsFormat:
		return []map[string]string{{
			"type":                 "CloudEvents",
			"cloud_events_version": v.CloudEventsVersion,
		}}
	}
	return []map[string]string{}
}

func unmarshallSubscriptionFormat(d *schema.ResourceData) (platform.DeliveryFormat, error) {
	input := d.Get("format").([]interface{})

	if len(input) == 1 {
		format := input[0].(map[string]interface{})

		switch format["type"] {
		case cloudEvents:
			return platform.CloudEventsFormat{
				CloudEventsVersion: format["cloud_events_version"].(string),
			}, nil
		case fmtPlatform:
			return platform.PlatformFormat{}, nil
		}
	}

	return nil, nil
}

func marshallSubscriptionChanges(m []platform.ChangeSubscription) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)

	for _, raw := range m {
		result = append(result, map[string]interface{}{
			"resource_type_ids": raw.ResourceTypeId,
		})
	}
	return result
}

func unmarshallSubscriptionChanges(d *schema.ResourceData) []platform.ChangeSubscription {
	var result []platform.ChangeSubscription
	input := d.Get("changes").([]interface{})
	if len(input) > 0 {
		for _, raw := range input {
			i := raw.(map[string]interface{})
			rawTypeIds := expandStringArray(i["resource_type_ids"].([]interface{}))

			for _, item := range rawTypeIds {
				result = append(result, platform.ChangeSubscription{
					ResourceTypeId: item,
				})
			}
		}
	}
	return result
}

func marshallSubscriptionMessages(m []platform.MessageSubscription) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(m))
	for _, raw := range m {
		result = append(result, map[string]interface{}{
			"resource_type_id": raw.ResourceTypeId,
			"types":            raw.Types,
		})
	}
	return result
}

func unmarshallSubscriptionMessages(d *schema.ResourceData) []platform.MessageSubscription {
	input := d.Get("message").([]interface{})
	var messageObjects []platform.MessageSubscription
	for _, raw := range input {
		i := raw.(map[string]interface{})
		messageObjects = append(messageObjects, platform.MessageSubscription{
			ResourceTypeId: i["resource_type_id"].(string),
			Types:          expandStringArray(i["types"].([]interface{})),
		})
	}

	return messageObjects
}

func validateSubscriptionDestination(d *schema.ResourceData) error {
	input := d.Get("destination").([]interface{})

	if len(input) != 1 {
		return fmt.Errorf("destination is missing")
	}

	dst := input[0].(map[string]interface{})

	dstType := dst["type"].(string)

	if dstTypeAlias, ok := destinationFieldAliases[dstType]; ok {
		dstType = dstTypeAlias
	}

	requiredFields, ok := destinationFields[dstType]
	if !ok {
		return fmt.Errorf("invalid type for destination: '%v'", dstType)
	}

	for _, field := range requiredFields {
		value, ok := dst[field].(string)
		if !ok {
			return fmt.Errorf("required property '%v' missing", field)
		} else if len(value) == 0 {
			return fmt.Errorf("required property '%v' is empty", field)
		}
	}
	return nil
}

func validateFormat(d *schema.ResourceData) error {
	input := d.Get("format").([]interface{})
	if len(input) < 1 {
		return nil
	}

	dst := input[0].(map[string]interface{})

	dstType := dst["type"].(string)
	requiredFields, ok := formatFields[dstType]
	if !ok {
		return fmt.Errorf("invalid type for format: '%v'", dstType)
	}

	for _, field := range requiredFields {
		value, ok := dst[field].(string)
		if !ok {
			return fmt.Errorf("required property '%v' missing", field)
		} else if len(value) == 0 {
			return fmt.Errorf("required property '%v' is empty", field)
		}
	}
	return nil

}

func suppressFuncForAttribute(attribute string, t ...string) schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		switch input := d.Get(attribute).(type) {
		case []interface{}:
			for _, dest := range input {
				for _, val := range t {
					if val == dest.(map[string]interface{})["type"] {
						return false
					}
				}
			}
		case map[string]interface{}:
			for _, val := range t {
				if val == input["type"] {
					return false
				}
			}
		}
		return true
	}
}

func suppressIfNotDestinationType(t ...string) schema.SchemaDiffSuppressFunc {
	return suppressFuncForAttribute("destination", t...)
}

func suppressIfNotFormatType(t ...string) schema.SchemaDiffSuppressFunc {
	return suppressFuncForAttribute("format", t...)
}

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

func migrateSubscriptionStateV0toV1(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	transformToList(rawState, "destination")
	transformToList(rawState, "format")
	return rawState, nil
}
