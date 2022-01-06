package commercetools

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
)

func resourceAPIExtension() *schema.Resource {
	return &schema.Resource{
		Description: "Create a new API extension to extend the bevahiour of an API with business logic. " +
			"Note that API extensions affect the performance of the API it is extending. If it fails, the whole API " +
			"call fails \n\n" +
			"Also see the [API Extension API Documentation](https://docs.commercetools.com/api/projects/api-extensions)",
		Create:        resourceAPIExtensionCreate,
		Read:          resourceAPIExtensionRead,
		Update:        resourceAPIExtensionUpdate,
		Delete:        resourceAPIExtensionDelete,
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceAPIExtensionResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: migrateAPIExtensionStateV0toV1,
				Version: 0,
			},
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "User-specific unique identifier for the extension",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"destination": {
				Description: "[Destination](https://docs.commercetools.com/api/projects/api-extensions#destination) " +
					"Details where the extension can be reached",
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateDestinationType,
						},
						// HTTP specific fields
						"url": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"azure_authentication": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"authorization_header": {
							Type:     schema.TypeString,
							Optional: true,
						},

						// AWSLambda specific fields
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

func validateDestinationType(val interface{}, key string) (warns []string, errs []error) {
	var v = strings.ToLower(val.(string))

	switch v {
	case
		"http",
		"awslambda":
		return
	default:
		errs = append(errs, fmt.Errorf("%q not a valid value for %q", val, key))
	}
	return
}

func resourceAPIExtensionCreate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	var extension *platform.Extension

	triggers := unmarshallExtensionTriggers(d)
	destination, err := unmarshallExtensionDestination(d)
	if err != nil {
		return err
	}

	draft := platform.ExtensionDraft{
		Key:         stringRef(d.Get("key")),
		Destination: destination,
		Triggers:    triggers,
		TimeoutInMs: intRef(d.Get("timeout_in_ms")),
	}

	err = resource.Retry(20*time.Second, func() *resource.RetryError {
		var err error

		extension, err = client.Extensions().Post(draft).Execute(context.Background())
		if err != nil {
			return handleCommercetoolsError(err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	if extension == nil {
		return fmt.Errorf("Error creating extension")
	}

	d.SetId(extension.ID)
	d.Set("version", extension.Version)

	return resourceAPIExtensionRead(d, m)
}

func resourceAPIExtensionRead(d *schema.ResourceData, m interface{}) error {
	log.Print("[DEBUG] Reading extensions from commercetools")
	client := getClient(m)

	extension, err := client.Extensions().WithId(d.Id()).Get().Execute(context.Background())

	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if extension == nil {
		log.Print("[DEBUG] No extensions found")
		d.SetId("")
	} else {
		log.Print("[DEBUG] Found following extensions:")
		log.Print(stringFormatObject(extension))

		d.Set("version", extension.Version)
		d.Set("key", extension.Key)
		d.Set("destination", marshallExtensionDestination(extension.Destination))
		d.Set("trigger", marshallExtensionTriggers(extension.Triggers))
		d.Set("timeout_in_ms", extension.TimeoutInMs)
	}
	return nil
}

func resourceAPIExtensionUpdate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)

	input := platform.ExtensionUpdate{
		Version: d.Get("version").(int),
		Actions: []platform.ExtensionUpdateAction{},
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&platform.ExtensionSetKeyAction{Key: &newKey})
	}

	if d.HasChange("trigger") {
		triggers := unmarshallExtensionTriggers(d)
		input.Actions = append(
			input.Actions,
			&platform.ExtensionChangeTriggersAction{Triggers: triggers})
	}

	if d.HasChange("destination") {
		destination, err := unmarshallExtensionDestination(d)
		if err != nil {
			return err
		}
		input.Actions = append(
			input.Actions,
			&platform.ExtensionChangeDestinationAction{Destination: destination})
	}

	if d.HasChange("timeout_in_ms") {
		newTimeout := d.Get("timeout_in_ms").(int)
		input.Actions = append(
			input.Actions,
			&platform.ExtensionSetTimeoutInMsAction{TimeoutInMs: &newTimeout})
	}

	_, err := client.Extensions().WithId(d.Id()).Post(input).Execute(context.Background())
	if err != nil {
		return err
	}

	return resourceAPIExtensionRead(d, m)
}

func resourceAPIExtensionDelete(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	version := d.Get("version").(int)
	_, err := client.Extensions().WithId(d.Id()).Delete().WithQueryParams(platform.ByProjectKeyExtensionsByIDRequestMethodDeleteInput{
		Version: version,
	}).Execute(context.Background())
	if err != nil {
		return err
	}
	return nil
}

//
// Helper methods
//

func unmarshallExtensionDestination(d *schema.ResourceData) (platform.Destination, error) {
	input, err := elementFromList(d, "destination")
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(input["type"].(string)) {
	case "http":
		auth, err := unmarshallExtensionDestinationAuthentication(input)
		if err != nil {
			return nil, err
		}

		return platform.ExtensionHttpDestination{
			Url:            input["url"].(string),
			Authentication: &auth,
		}, nil
	case "awslambda":
		return platform.ExtensionAWSLambdaDestination{
			Arn:          input["arn"].(string),
			AccessKey:    input["access_key"].(string),
			AccessSecret: input["access_secret"].(string),
		}, nil
	default:
		return nil, fmt.Errorf("Extension type %s not implemented", input["type"])
	}
}

func unmarshallExtensionDestinationAuthentication(destInput map[string]interface{}) (platform.ExtensionHttpDestinationAuthentication, error) {
	authKeys := [2]string{"authorization_header", "azure_authentication"}
	count := 0
	for _, key := range authKeys {
		if value, ok := destInput[key]; ok {
			if value != "" {
				count++
			}
		}
	}
	if count > 1 {
		return nil, fmt.Errorf(
			"In the destination only one of the auth values should be definied: %q", authKeys)
	}

	if val, ok := isNotEmpty(destInput, "authorization_header"); ok {
		return &platform.ExtensionAuthorizationHeaderAuthentication{
			HeaderValue: val.(string),
		}, nil
	}
	if val, ok := isNotEmpty(destInput, "azure_authentication"); ok {
		return &platform.ExtensionAzureFunctionsAuthentication{
			Key: val.(string),
		}, nil
	}

	return nil, nil
}

func marshallExtensionDestination(d platform.Destination) []map[string]string {
	switch v := d.(type) {
	case platform.ExtensionHttpDestination:
		switch a := v.Authentication.(type) {
		case platform.ExtensionAuthorizationHeaderAuthentication:
			return []map[string]string{{
				"type":                 "HTTP",
				"url":                  v.Url,
				"authorization_header": a.HeaderValue,
			}}
		case platform.ExtensionAzureFunctionsAuthentication:
			return []map[string]string{{
				"type":                 "HTTP",
				"url":                  v.Url,
				"azure_authentication": a.Key,
			}}
		}
		return []map[string]string{{
			"type": "HTTP",
			"url":  v.Url,
		}}

	case platform.ExtensionAWSLambdaDestination:
		return []map[string]string{{
			"type":          "awslambda",
			"access_key":    v.AccessKey,
			"access_secret": v.AccessSecret,
			"arn":           v.Arn,
		}}

	}
	return []map[string]string{}
}

func marshallExtensionTriggers(triggers []platform.ExtensionTrigger) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(triggers))

	for _, t := range triggers {
		result = append(result, map[string]interface{}{
			"resource_type_id": t.ResourceTypeId,
			"actions":          t.Actions,
		})
	}

	return result
}

func unmarshallExtensionTriggers(d *schema.ResourceData) []platform.ExtensionTrigger {
	input := d.Get("trigger").([]interface{})
	var result []platform.ExtensionTrigger

	for _, raw := range input {
		i := raw.(map[string]interface{})
		var typeId platform.ExtensionResourceTypeId

		switch i["resource_type_id"].(string) {
		case "cart":
			typeId = platform.ExtensionResourceTypeIdCart
		case "order":
			typeId = platform.ExtensionResourceTypeIdOrder
		case "payment":
			typeId = platform.ExtensionResourceTypeIdPayment
		case "customer":
			typeId = platform.ExtensionResourceTypeIdCustomer
		}

		rawActions := i["actions"].([]interface{})
		actions := make([]platform.ExtensionAction, 0, len(rawActions))
		for _, item := range rawActions {
			actions = append(actions, platform.ExtensionAction(item.(string)))
		}

		result = append(result, platform.ExtensionTrigger{
			ResourceTypeId: typeId,
			Actions:        actions,
		})
	}
	return result
}

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

func migrateAPIExtensionStateV0toV1(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	transformToList(rawState, "destination")
	return rawState, nil
}
