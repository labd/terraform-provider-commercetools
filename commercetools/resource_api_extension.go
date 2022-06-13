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

func resourceAPIExtension() *schema.Resource {
	return &schema.Resource{
		Description: "Create a new API extension to extend the bevahiour of an API with business logic. " +
			"Note that API extensions affect the performance of the API it is extending. If it fails, the whole API " +
			"call fails \n\n" +
			"Also see the [API Extension API Documentation](https://docs.commercetools.com/api/projects/api-extensions)",
		CreateContext: resourceAPIExtensionCreate,
		ReadContext:   resourceAPIExtensionRead,
		UpdateContext: resourceAPIExtensionUpdate,
		DeleteContext: resourceAPIExtensionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
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
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
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
						"condition": {
							Description: "Valid predicate that controls the conditions under which the API Extension is called.",
							Type:        schema.TypeString,
							Optional:    true,
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
		errs = append(errs, fmt.Errorf("%q not a valid value for %q, valid options are: http, awslambda", val, key))
	}
	return
}

func validateExtensionDestination(draft platform.ExtensionDraft) error {

	switch t := draft.Destination.(type) {
	case platform.AWSLambdaDestination:
		if t.Arn == "" {
			return fmt.Errorf("arn is required when using AWSLambda as destination")
		}
	}
	return nil
}

func resourceAPIExtensionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)

	triggers := unmarshallExtensionTriggers(d)
	destination, err := unmarshallExtensionDestination(d)
	if err != nil {
		return diag.FromErr(err)
	}

	draft := platform.ExtensionDraft{
		Destination: destination,
		Triggers:    triggers,
	}

	timeoutInMs := d.Get("timeout_in_ms")
	if timeoutInMs != 0 {
		draft.TimeoutInMs = intRef(timeoutInMs)
	}

	key := stringRef(d.Get("key"))
	if *key != "" {
		draft.Key = key
	}

	if err := validateExtensionDestination(draft); err != nil {
		return diag.FromErr(err)
	}

	var extension *platform.Extension
	err = resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		var err error
		extension, err = client.Extensions().Post(draft).Execute(ctx)
		return processRemoteError(err)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if extension == nil {
		return diag.Errorf("Error creating extension")
	}

	d.SetId(extension.ID)
	d.Set("version", extension.Version)

	return resourceAPIExtensionRead(ctx, d, m)
}

func resourceAPIExtensionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Print("[DEBUG] Reading extensions from commercetools")
	client := getClient(m)

	extension, err := client.Extensions().WithId(d.Id()).Get().Execute(ctx)
	if err != nil {
		if IsResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if extension == nil {
		log.Print("[DEBUG] No extensions found")
		d.SetId("")
	} else {
		log.Print("[DEBUG] Found following extensions:")
		log.Print(stringFormatObject(extension))

		d.Set("version", extension.Version)
		d.Set("key", extension.Key)
		d.Set("destination", marshallExtensionDestination(extension.Destination, d))
		d.Set("trigger", marshallExtensionTriggers(extension.Triggers))
		d.Set("timeout_in_ms", extension.TimeoutInMs)
	}
	return nil
}

func resourceAPIExtensionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
			return diag.FromErr(err)
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

	err := resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		_, err := client.Extensions().WithId(d.Id()).Post(input).Execute(ctx)
		return processRemoteError(err)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceAPIExtensionRead(ctx, d, m)
}

func resourceAPIExtensionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)
	version := d.Get("version").(int)
	_, err := client.Extensions().WithId(d.Id()).Delete().Version(version).Execute(ctx)
	if err != nil {
		return diag.FromErr(err)
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

		return platform.HttpDestination{
			Url:            input["url"].(string),
			Authentication: &auth,
		}, nil
	case "awslambda":
		return platform.AWSLambdaDestination{
			Arn:          input["arn"].(string),
			AccessKey:    input["access_key"].(string),
			AccessSecret: input["access_secret"].(string),
		}, nil
	default:
		return nil, fmt.Errorf("extension type %s not implemented", input["type"])
	}
}

func unmarshallExtensionDestinationAuthentication(destInput map[string]interface{}) (platform.HttpDestinationAuthentication, error) {
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
			"in the destination only one of the auth values should be definied: %q", authKeys)
	}

	if val, ok := isNotEmpty(destInput, "authorization_header"); ok {
		return &platform.AuthorizationHeaderAuthentication{
			HeaderValue: val.(string),
		}, nil
	}
	if val, ok := isNotEmpty(destInput, "azure_authentication"); ok {
		return &platform.AzureFunctionsAuthentication{
			Key: val.(string),
		}, nil
	}

	return nil, nil
}

func marshallExtensionDestination(dst platform.Destination, d *schema.ResourceData) []map[string]string {

	switch v := dst.(type) {
	case platform.HttpDestination:
		switch a := v.Authentication.(type) {
		case platform.AuthorizationHeaderAuthentication:
			return []map[string]string{{
				"type":                 "HTTP",
				"url":                  v.Url,
				"authorization_header": a.HeaderValue,
			}}
		case platform.AzureFunctionsAuthentication:
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

	case platform.AWSLambdaDestination:
		accessSecret := ""

		// If we already have a state we get the accessSecret from the state.
		// Normally we have a state, the one exception is when we are importing
		// the resource.
		if !d.GetRawState().GetAttr("version").IsNull() {
			// Read the access secret from the current resource data
			c, _ := unmarshallExtensionDestination(d)
			switch current := c.(type) {
			case platform.AWSLambdaDestination:
				accessSecret = current.AccessSecret
			}
		}

		return []map[string]string{{
			"type":          "awslambda",
			"access_key":    v.AccessKey,
			"access_secret": accessSecret,
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
			"condition":        t.Condition,
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

		var condition *string
		if val, ok := i["condition"].(string); ok {
			condition = stringRef(val)
		}

		result = append(result, platform.ExtensionTrigger{
			ResourceTypeId: typeId,
			Actions:        actions,
			Condition:      condition,
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
