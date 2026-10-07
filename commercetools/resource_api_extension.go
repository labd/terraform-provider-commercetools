package commercetools

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func resourceAPIExtension() *schema.Resource {
	return &schema.Resource{
		Description: "Create a new API extension to extend the behaviour of an API with business logic. " +
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
				Description: "Maximum time (in milliseconds) that the Extension can respond within. If no timeout is " +
					"provided, the default value is used for all types of Extensions, including payment Extensions. " +
					"The maximum value is 10000 ms (10 seconds) for payment Extensions and 2000 ms (2 seconds) for " +
					"all other Extensions.",
				Type:     schema.TypeInt,
				Default:  2000,
				Optional: true,
			},
			"dependencies": {
				Description: "Other [API Extensions](https://docs.commercetools.com/api/projects/api-extensions#dependencies) " +
					"that must complete before this Extension is called. The Extension receives the resource state after " +
					"all transitive ancestors' update actions have been applied. A maximum of 5 dependencies is allowed " +
					"and the resulting chain may not be deeper than 3 layers. " +
					"Each entry is either the key or the id of another Extension; an entry formatted as a UUID is " +
					"treated as an id, anything else as a key. Referencing by key is recommended, also for Extensions " +
					"that are not managed by this Terraform configuration.",
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 5,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"expansion_paths": {
				Description: "[Expansion paths](https://docs.commercetools.com/api/general-concepts#expansion-paths) " +
					"used for reference expansion of the payload. Be aware of the limits of this feature and its " +
					"performance impact.",
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"additional_context": {
				Description: "Configures additional information included in the payload sent to the API Extension",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"include_old_resource": {
							Description: "Whether the payload sent to the API Extension should include an " +
								"`oldResource` field with the state of the resource before the update. This only " +
								"applies to `Update` actions; for `Create` actions `oldResource` is never included.",
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
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

	if dependencies := expandExtensionDependencies(d); len(dependencies) > 0 {
		draft.Dependencies = dependencies
	}

	if expansionPaths := expandExtensionExpansionPaths(d); len(expansionPaths) > 0 {
		draft.ExpansionPaths = expansionPaths
	}

	draft.AdditionalContext = expandExtensionAdditionalContext(d)

	if err := validateExtensionDestination(draft); err != nil {
		return diag.FromErr(err)
	}

	var extension *platform.Extension
	err = retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
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
	dependencyIDs := extensionDependencyIDs(extension.Dependencies)
	dependencyKeys, err := lookupExtensionKeys(ctx, client, dependencyIDs)
	if err != nil {
		return diag.FromErr(err)
	}
	_ = d.Set("dependencies", flattenExtensionDependencies(dependencyIDs, dependencyKeys, d))
	_ = d.Set("expansion_paths", extension.ExpansionPaths)
	_ = d.Set("additional_context", flattenExtensionAdditionalContext(extension.AdditionalContext, d))
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

	if d.HasChange("dependencies") {
		// Send an empty (non-nil) list to remove all dependencies.
		dependencies := expandExtensionDependencies(d)
		if dependencies == nil {
			dependencies = []platform.ExtensionResourceIdentifier{}
		}
		input.Actions = append(
			input.Actions,
			&platform.ExtensionSetDependenciesAction{Dependencies: dependencies})
	}

	if d.HasChange("expansion_paths") {
		// Send an empty (non-nil) list to remove all expansion paths.
		expansionPaths := expandExtensionExpansionPaths(d)
		if expansionPaths == nil {
			expansionPaths = []string{}
		}
		input.Actions = append(
			input.Actions,
			&platform.ExtensionSetExpansionPathsAction{ExpansionPaths: expansionPaths})
	}

	if d.HasChange("additional_context") {
		additionalContext := expandExtensionAdditionalContext(d)
		if additionalContext == nil {
			additionalContext = &platform.ExtensionAdditionalContextDraft{}
		}
		input.Actions = append(
			input.Actions,
			&platform.ExtensionSetAdditionalContextAction{AdditionalContext: *additionalContext})
	}

	err := retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
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
	// value which is returned as a masked value by the commercetools API. This means
	// we need to extract the value from the current raw state file. However, when
	// importing a resource we don't have the value, so we need to handle that
	// scenario as well.
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
		case "shopping-list":
			typeId = platform.ExtensionResourceTypeIdShoppingList
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

// extensionIDRegexp matches the UUID format used for commercetools resource
// ids. It is used to tell an id apart from a user-defined key, since a
// dependency can be referenced by either.
var extensionIDRegexp = regexp.MustCompile(
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func isExtensionID(value string) bool {
	return extensionIDRegexp.MatchString(value)
}

func expandExtensionDependencies(d *schema.ResourceData) []platform.ExtensionResourceIdentifier {
	input := d.Get("dependencies").([]any)
	if len(input) == 0 {
		return nil
	}

	result := make([]platform.ExtensionResourceIdentifier, 0, len(input))
	for _, raw := range input {
		value, ok := raw.(string)
		if !ok || value == "" {
			continue
		}

		if isExtensionID(value) {
			result = append(result, platform.ExtensionResourceIdentifier{ID: &value})
		} else {
			result = append(result, platform.ExtensionResourceIdentifier{Key: &value})
		}
	}
	return result
}

func extensionDependencyIDs(dependencies []platform.ExtensionReference) []string {
	result := make([]string, 0, len(dependencies))
	for _, dependency := range dependencies {
		result = append(result, dependency.ID)
	}
	return result
}

// lookupExtensionKeys resolves extension ids to their keys. Dependencies are
// returned as references holding only an id, while the configuration may
// reference them by key. The extension endpoint does not support reference
// expansion, so a single query is used to resolve them. Extensions without a
// key are absent from the result.
func lookupExtensionKeys(ctx context.Context, client *platform.ByProjectKeyRequestBuilder, ids []string) (map[string]string, error) {
	result := make(map[string]string, len(ids))
	if len(ids) == 0 {
		return result, nil
	}

	quoted := make([]string, 0, len(ids))
	for _, id := range ids {
		quoted = append(quoted, strconv.Quote(id))
	}

	response, err := client.Extensions().Get().
		Where([]string{fmt.Sprintf("id in (%s)", strings.Join(quoted, ", "))}).
		Limit(len(ids)).
		Execute(ctx)
	if err != nil {
		return nil, err
	}

	for _, extension := range response.Results {
		if extension.Key != nil && *extension.Key != "" {
			result[extension.ID] = *extension.Key
		}
	}
	return result, nil
}

// flattenExtensionDependencies converts the ids returned by commercetools back
// into the notation used in the configuration. An entry written as an id stays
// an id, everything else is written as a key. Dependencies that are not in the
// state yet (drift, or an imported resource) are written as a key when the
// extension has one, and as an id otherwise.
func flattenExtensionDependencies(ids []string, keys map[string]string, d *schema.ResourceData) []string {
	current := make(map[string]bool)
	for _, raw := range d.Get("dependencies").([]any) {
		if value, ok := raw.(string); ok {
			current[value] = true
		}
	}

	result := make([]string, 0, len(ids))
	for _, id := range ids {
		key := keys[id]
		if current[id] || key == "" {
			result = append(result, id)
			continue
		}
		result = append(result, key)
	}
	return result
}

func expandExtensionExpansionPaths(d *schema.ResourceData) []string {
	input := d.Get("expansion_paths").([]any)
	if len(input) == 0 {
		return nil
	}

	result := make([]string, 0, len(input))
	for _, raw := range input {
		result = append(result, raw.(string))
	}
	return result
}

func expandExtensionAdditionalContext(d *schema.ResourceData) *platform.ExtensionAdditionalContextDraft {
	input, err := elementFromList(d, "additional_context")
	if err != nil || input == nil {
		return nil
	}

	includeOldResource, _ := input["include_old_resource"].(bool)
	return &platform.ExtensionAdditionalContextDraft{
		IncludeOldResource: &includeOldResource,
	}
}

// flattenExtensionAdditionalContext writes the additional context to the state
// file. commercetools always returns the additional context, also when it was
// never configured. To avoid a permanent diff for those resources we only write
// the block when it holds a non-default value or when it is already present in
// the state.
func flattenExtensionAdditionalContext(ac *platform.ExtensionAdditionalContext, d *schema.ResourceData) []map[string]any {
	if ac == nil {
		return []map[string]any{}
	}

	if !ac.IncludeOldResource {
		if current, ok := d.Get("additional_context").([]any); !ok || len(current) == 0 {
			return []map[string]any{}
		}
	}

	return []map[string]any{{
		"include_old_resource": ac.IncludeOldResource,
	}}
}
