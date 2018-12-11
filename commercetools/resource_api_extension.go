package commercetools

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/labd/commercetools-go-sdk/commercetools"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAPIExtension() *schema.Resource {
	return &schema.Resource{
		Create: resourceAPIExtensionCreate,
		Read:   resourceAPIExtensionRead,
		Update: resourceAPIExtensionUpdate,
		Delete: resourceAPIExtensionDelete,

		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"destination": {
				Type:     schema.TypeMap,
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
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_type_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"actions": {
							Type:     schema.TypeList,
							Required: true,
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

func validateDestinationType(val interface{}, key string) (warns []string, errs []error) {
	var v = strings.ToLower(val.(string))

	switch v {
	case
		"http",
		"awslambda",
		"azurefunctions":
		return
	default:
		errs = append(errs, fmt.Errorf("%q not a valid value for %q", val, key))
	}
	return
}

func resourceAPIExtensionCreate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	var extension *commercetools.Extension

	triggers := resourceAPIExtensionGetTriggers(d)
	destination, err := resourceAPIExtensionGetDestination(d)
	if err != nil {
		return err
	}

	draft := &commercetools.ExtensionDraft{
		Key:         d.Get("key").(string),
		Destination: destination,
		Triggers:    triggers,
	}

	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error

		extension, err = client.Extensions.Create(draft)
		if err != nil {
			log.Print("[DEBUG] Error while creating extension, will try again")
			log.Print(err)
			return resource.RetryableError(err)
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

	extension, err := client.Extensions.GetByID(d.Id())

	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
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
		d.Set("destination", extension.Destination)
		d.Set("triggers", extension.Triggers)
	}
	return nil
}

func resourceAPIExtensionUpdate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)

	input := &commercetools.ExtensionUpdateInput{
		ID:      d.Id(),
		Version: d.Get("version").(int),
		Actions: []commercetools.ExtensionUpdateAction{},
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.ExtensionSetKeyAction{Key: newKey})
	}

	if d.HasChange("triggers") {
		triggers := resourceAPIExtensionGetTriggers(d)
		input.Actions = append(
			input.Actions,
			&commercetools.ExtensionChangeTriggersAction{Triggers: triggers})
	}

	if d.HasChange("destination") {
		destination, err := resourceAPIExtensionGetDestination(d)
		if err != nil {
			return err
		}
		input.Actions = append(
			input.Actions,
			&commercetools.ExtensionChangeDestinationAction{Destination: destination})
	}

	_, err := client.Extensions.Update(input)
	if err != nil {
		return err
	}

	return resourceAPIExtensionRead(d, m)
}

func resourceAPIExtensionDelete(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	version := d.Get("version").(int)
	_, err := client.Extensions.DeleteByID(d.Id(), version)
	if err != nil {
		return err
	}
	return nil
}

//
// Helper methods
//

func resourceAPIExtensionGetDestination(d *schema.ResourceData) (commercetools.Destination, error) {
	input := d.Get("destination").(map[string]interface{})
	switch strings.ToLower(input["type"].(string)) {
	case "http":
		auth, err := resourceAPIExtensionGetAuthentication(input)
		if err != nil {
			return nil, err
		}

		return commercetools.ExtensionHttpDestination{
			URL:            input["url"].(string),
			Authentication: auth,
		}, nil
	case "awslambda":
		return commercetools.ExtensionAWSLambdaDestination{
			Arn:          input["arn"].(string),
			AccessKey:    input["access_key"].(string),
			AccessSecret: input["access_secret"].(string),
		}, nil
	default:
		return nil, fmt.Errorf("Extension type %s not implemented", input["type"])
	}
}

func resourceAPIExtensionGetAuthentication(destInput map[string]interface{}) (commercetools.ExtensionHttpDestinationAuthentication, error) {
	authKeys := [2]string{"authorization_header", "azure_authentication"}
	count := 0
	for _, key := range authKeys {
		if _, ok := destInput[key]; ok {
			count++
		}
	}
	if count > 1 {
		return nil, fmt.Errorf(
			"In the destination only one of the auth values should be definied: %q", authKeys)
	}

	if authVal, ok := destInput["authorization_header"]; ok {
		return &commercetools.ExtensionAuthorizationHeaderAuthentication{
			HeaderValue: authVal.(string),
		}, nil
	}
	if authVal, ok := destInput["azure_authentication"]; ok {
		return &commercetools.ExtensionAzureFunctionsAuthentication{
			Key: authVal.(string),
		}, nil
	}

	return nil, nil
}

func resourceAPIExtensionGetTriggers(d *schema.ResourceData) []commercetools.ExtensionTrigger {
	input := d.Get("trigger").([]interface{})
	var result []commercetools.ExtensionTrigger

	for _, raw := range input {
		i := raw.(map[string]interface{})
		typeID := i["resource_type_id"].(string)

		actions := []commercetools.ExtensionAction{}
		for _, item := range expandStringArray(i["actions"].([]interface{})) {
			actions = append(actions, commercetools.ExtensionAction(item))
		}

		result = append(result, commercetools.ExtensionTrigger{
			ResourceTypeID: commercetools.ExtensionResourceTypeID(typeID),
			Actions:        actions,
		})
	}

	return result
}
