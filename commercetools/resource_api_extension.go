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

	"github.com/labd/terraform-provider-commercetools/internal/utils"
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

func validateDestinationType(val any, key string) (warns []string, errs []error) {
	var v = strings.ToLower(val.(string))

	switch v {
	case
		"googlecloudfunction",
		"http",
		"awslambda":
		return
	default:
		errs = append(errs, fmt.Errorf("%q not a valid value for %q, valid options are: googlecloudfunction, http, awslambda", val, key))
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

func resourceAPIExtensionCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	triggers := expandExtensionTriggers(d)
	destination, err := expandExtensionDestination(d)
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
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
		return utils.ProcessRemoteError(err)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if extension == nil {
		return diag.Errorf("Error creating extension")
	}

	d.SetId(extension.ID)
	_ = d.Set("version", extension.Version)

	return resourceAPIExtensionRead(ctx, d, m)
}

func resourceAPIExtensionRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	extension, err := client.Extensions().WithId(d.Id()).Get().Execute(ctx)
	if err != nil {
		if utils.IsResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("version", extension.Version)
	_ = d.Set("key", extension.Key)
	_ = d.Set("destination", flattenExtensionDestination(extension.Destination, d))
	_ = d.Set("trigger", flattenExtensionTriggers(extension.Triggers))
	_ = d.Set("timeout_in_ms", extension.TimeoutInMs)
	return nil
}

func resourceAPIExtensionUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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
		triggers := expandExtensionTriggers(d)
		input.Actions = append(
			input.Actions,
			&platform.ExtensionChangeTriggersAction{Triggers: triggers})
	}

	if d.HasChange("destination") {
		destination, err := expandExtensionDestination(d)
		if err != nil {
			// Workaround invalid state to be written, see
			// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
			d.Partial(true)
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
		return utils.ProcessRemoteError(err)
	})

	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	return resourceAPIExtensionRead(ctx, d, m)
}

func resourceAPIExtensionDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

func expandExtensionDestination(d *schema.ResourceData) (platform.Destination, error) {
	input, err := elementFromList(d, "destination")
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(input["type"].(string)) {
	case "googlecloudfunction":
		return platform.GoogleCloudFunctionDestination{
			Url: input["url"].(string),
		}, nil
	case "http":
		auth, err := expandExtensionDestinationAuthentication(input)
		if err != nil {
			return nil, err
		}

		return platform.HttpDestination{
			Url:            input["url"].(string),
			Authentication: auth,
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

func expandExtensionDestinationAuthentication(destInput map[string]any) (platform.HttpDestinationAuthentication, error) {
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
		return platform.AuthorizationHeaderAuthentication{
			HeaderValue: val.(string),
		}, nil
	}
	if val, ok := isNotEmpty(destInput, "azure_authentication"); ok {
		return platform.AzureFunctionsAuthentication{
			Key: val.(string),
		}, nil
	}

	return nil, nil
}

// flattenExtensionDestination flattens the destination returned by
// commercetools to write it in the state file.
func flattenExtensionDestination(dst platform.Destination, d *schema.ResourceData) []map[string]string {
	// Special handling is required here since the destination contains a secret
	// value which is returned as a masked value by the commercetools API.  This
	// means we need to extract the value from the current raw state file.
	// However when importing a resource we don't have the value so we need to
	// handle that scenario as well.
	isExisting := true
	rawState := d.GetRawState()
	if !rawState.IsNull() {
		isExisting = !rawState.AsValueMap()["version"].IsNull()
	}

	var current platform.Destination
	if isExisting {
		current, _ = expandExtensionDestination(d)
	}

	// A destination is either GoogleCloudFunction, HTTP or AWSLambda
	switch d := dst.(type) {

	case platform.GoogleCloudFunctionDestination:
		return []map[string]string{{
			"type": "GoogleCloudFunction",
			"url":  d.Url,
		}}

	// For the HTTP Destination there are two specific authentication types:
	// AuthorizationHeader and AzureFunctions.
	case platform.HttpDestination:
		switch d.Authentication.(type) {

		case platform.AuthorizationHeaderAuthentication:

			// The headerValue value is masked when retrieved from commercetools,
			// so use the value from the state file instead (if it exists)
			secretValue := ""
			if current != nil {
				if c, ok := current.(platform.HttpDestination); ok {
					if auth, ok := c.Authentication.(platform.AuthorizationHeaderAuthentication); ok {
						secretValue = auth.HeaderValue
					}
				}
			}

			return []map[string]string{{
				"type":                 "HTTP",
				"url":                  d.Url,
				"authorization_header": secretValue,
			}}

		case platform.AzureFunctionsAuthentication:
			// The headerValue value is masked when retrieved from commercetools,
			// so use the value from the state file instead (if it exists)
			secretValue := ""
			if current != nil {
				if c, ok := current.(platform.HttpDestination); ok {
					if auth, ok := c.Authentication.(platform.AzureFunctionsAuthentication); ok {
						secretValue = auth.Key
					}
				}
			}
			return []map[string]string{{
				"type":                 "HTTP",
				"url":                  d.Url,
				"azure_authentication": secretValue,
			}}

		default:
			log.Println("Unexpected authentication type")
			return []map[string]string{{
				"type": "HTTP",
				"url":  d.Url,
			}}
		}

	case platform.AWSLambdaDestination:

		// The accessSecret value is masked when retrieved from commercetools,
		// so use the value from the state file instead (if it exists)
		secretValue := ""
		if current != nil {
			if c, ok := current.(platform.AWSLambdaDestination); ok {
				secretValue = c.AccessSecret
			}
		}

		return []map[string]string{{
			"type":          "awslambda",
			"access_key":    d.AccessKey,
			"access_secret": secretValue,
			"arn":           d.Arn,
		}}

	default:
		return []map[string]string{}
	}
}

func flattenExtensionTriggers(triggers []platform.ExtensionTrigger) []map[string]any {
	result := make([]map[string]any, 0, len(triggers))

	for _, t := range triggers {
		result = append(result, map[string]any{
			"resource_type_id": t.ResourceTypeId,
			"actions":          t.Actions,
			"condition":        nilIfEmpty(t.Condition),
		})
	}

	return result
}

func expandExtensionTriggers(d *schema.ResourceData) []platform.ExtensionTrigger {
	input := d.Get("trigger").([]any)
	var result []platform.ExtensionTrigger

	for _, raw := range input {
		i := raw.(map[string]any)
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
		case "quote-request":
			typeId = platform.ExtensionResourceTypeIdQuoteRequest
		case "staged-quote":
			typeId = platform.ExtensionResourceTypeIdStagedQuote
		case "quote":
			typeId = platform.ExtensionResourceTypeIdQuote
		case "business-unit":
			typeId = platform.ExtensionResourceTypeIdBusinessUnit
		}

		rawActions := i["actions"].([]any)
		actions := make([]platform.ExtensionAction, 0, len(rawActions))
		for _, item := range rawActions {
			actions = append(actions, platform.ExtensionAction(item.(string)))
		}

		var condition *string
		if val, ok := i["condition"].(string); ok {
			condition = nilIfEmpty(stringRef(val))
		}

		result = append(result, platform.ExtensionTrigger{
			ResourceTypeId: typeId,
			Actions:        actions,
			Condition:      condition,
		})
	}
	return result
}
