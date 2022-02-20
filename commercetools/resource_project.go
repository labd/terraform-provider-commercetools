package commercetools

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
)

// TODO: A lot of fields are optional in this schema that are not optional in platform. When not set via terraform
// commercetools simply sets the default values for these fields. This works but can be a little confusing. It is worth
// considering whether to align the optional/required status of the fields in the provider with that of the API itself
func resourceProjectSettings() *schema.Resource {
	return &schema.Resource{
		Description: "The project endpoint provides a limited set of information about settings and configuration of " +
			"the project. Updating the settings is eventually consistent, it may take up to a minute before " +
			"a change becomes fully active.\n\n" +
			"See also the [Project Settings API Documentation](https://docs.commercetools.com/api/projects/project)",
		Create: resourceProjectCreate,
		Read:   resourceProjectRead,
		Update: resourceProjectUpdate,
		Delete: resourceProjectDelete,
		Exists: resourceProjectExists,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceProjectSettingsResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: migrateResourceProjectSettingsStateV0toV1,
				Version: 0,
			},
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "The unique key of the project",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "The name of the project",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"currencies": {
				Description: "A three-digit currency code as per [ISO 4217](https://en.wikipedia.org/wiki/ISO_4217)",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"countries": {
				Description: "A two-digit country code as per [ISO 3166-1 alpha-2](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2)",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"languages": {
				Description: "[IETF Language Tag](https://en.wikipedia.org/wiki/IETF_language_tag)",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"messages": {
				Description: "[Messages Configuration](https://docs.commercetools.com/api/projects/project#messages-configuration)",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Description: "When true the creation of messages on the Messages Query HTTP API is enabled",
							Type:        schema.TypeBool,
							Required:    true,
						},
					},
				},
			},
			"enable_search_index_products": {
				Description: "Enable the Search Indexing of products",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"enable_search_index_orders": {
				Description: "Enable the Search Indexing of orders",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"external_oauth": {
				Description: "[External OAUTH](https://docs.commercetools.com/api/projects/project#externaloauth)",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"authorization_header": {
							Description: "Partially hidden on retrieval",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
			"carts": {
				Description: "[Carts Configuration](https://docs.commercetools.com/api/projects/project#carts-configuration)",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"country_tax_rate_fallback_enabled": {
							Description: "Indicates if country - no state tax rate fallback should be used when a " +
								"shipping address state is not explicitly covered in the rates lists of all tax " +
								"categories of a cart line items",
							Type:     schema.TypeBool,
							Required: true,
						},
						"delete_days_after_last_modification": {
							Description: "Number - Optional The default value for the " +
								"deleteDaysAfterLastModification parameter of the CartDraft. Initially set to 90 for " +
								"projects created after December 2019.",
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"shipping_rate_input_type": {
				Description: "Three ways to dynamically select a ShippingRatePriceTier exist. The CartValue type uses " +
					"the sum of all line item prices, whereas CartClassification and CartScore use the " +
					"shippingRateInput field on the cart to select a tier",
				Type:     schema.TypeString,
				Optional: true,
			},
			"shipping_rate_cart_classification_value": {
				Description: "If shipping_rate_input_type is set to CartClassification these values are used to create " +
					"tiers\n. Only a key defined inside the values array can be used to create a tier, or to set a value " +
					"for the shippingRateInput on the cart. The keys are checked for uniqueness and the request is " +
					"rejected if keys are not unique",
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"label": {
							Type:     TypeLocalizedString,
							Optional: true,
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

func resourceProjectExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := getClient(m)

	_, err := client.Get().Execute(context.Background())
	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func resourceProjectCreate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	project, err := client.Get().Execute(context.Background())

	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				return nil
			}
		}
		return err
	}

	err = projectUpdate(d, client, project.Version)
	if err != nil {
		return err
	}
	return resourceProjectRead(d, m)
}

func resourceProjectRead(d *schema.ResourceData, m interface{}) error {
	log.Print("[DEBUG] Reading projects from commercetools")
	client := getClient(m)

	project, err := client.Get().Execute(context.Background())

	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				return nil
			}
		}
		return err
	}

	log.Print("[DEBUG] Found the following project:")
	log.Print(stringFormatObject(project))

	d.SetId(project.Key)
	d.Set("version", project.Version)
	d.Set("name", project.Name)
	d.Set("currencies", project.Currencies)
	d.Set("countries", project.Countries)
	d.Set("languages", project.Languages)
	d.Set("shipping_rate_input_type", marshallProjectShippingRateInputType(project.ShippingRateInputType))
	d.Set("enable_search_index_products", marshallProjectSearchIndexProducts(project.SearchIndexing))
	d.Set("enable_search_index_orders", marshallProjectSearchIndexOrders(project.SearchIndexing))
	d.Set("external_oauth", marshallProjectExternalOAuth(project.ExternalOAuth))
	d.Set("carts", marshallProjectCarts(project.Carts))
	d.Set("messages", marshallProjectMessages(project.Messages))
	return nil
}

func resourceProjectUpdate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	version := d.Get("version").(int)
	err := projectUpdate(d, client, version)
	if err != nil {
		return err
	}
	return resourceProjectRead(d, m)
}

func resourceProjectDelete(d *schema.ResourceData, m interface{}) error {
	d.SetId("")
	return nil
}

func projectUpdate(d *schema.ResourceData, client *platform.ByProjectKeyRequestBuilder, version int) error {
	input := platform.ProjectUpdate{
		Version: version,
		Actions: []platform.ProjectUpdateAction{},
	}

	if d.HasChange("name") {
		input.Actions = append(input.Actions, &platform.ProjectChangeNameAction{Name: d.Get("name").(string)})
	}

	if d.HasChange("currencies") {
		newCurrencies := []string{}
		for _, item := range getStringSlice(d, "currencies") {
			newCurrencies = append(newCurrencies, item)
		}

		input.Actions = append(
			input.Actions,
			&platform.ProjectChangeCurrenciesAction{Currencies: newCurrencies})
	}

	if d.HasChange("countries") {
		newCountries := []string{}
		for _, item := range getStringSlice(d, "countries") {
			newCountries = append(newCountries, item)
		}

		input.Actions = append(
			input.Actions,
			&platform.ProjectChangeCountriesAction{Countries: newCountries})
	}

	if d.HasChange("languages") {
		newLanguages := []string{}
		for _, item := range getStringSlice(d, "languages") {
			newLanguages = append(newLanguages, item)
		}
		input.Actions = append(
			input.Actions,
			&platform.ProjectChangeLanguagesAction{Languages: newLanguages})
	}

	if d.HasChange("messages") {
		messages, err := elementFromList(d, "messages")
		if err != nil {
			return err
		}
		if messages["enabled"] != nil {
			input.Actions = append(
				input.Actions,
				&platform.ProjectChangeMessagesEnabledAction{MessagesEnabled: messages["enabled"].(bool)})
		} else {
			// To commercetools this field is not optional, so when deleting we revert to the default: false:
			input.Actions = append(
				input.Actions,
				&platform.ProjectChangeMessagesEnabledAction{MessagesEnabled: false})
		}

	}

	if d.HasChange("shipping_rate_input_type") || d.HasChange("shipping_rate_cart_classification_value") {
		newShippingRateInputType, err := getShippingRateInputType(d)
		if err != nil {
			return err
		}
		input.Actions = append(
			input.Actions,
			&platform.ProjectSetShippingRateInputTypeAction{ShippingRateInputType: &newShippingRateInputType})
	}

	if d.HasChange("external_oauth") {
		externalOAuth, err := elementFromList(d, "external_oauth")
		if err != nil {
			return err
		}
		if externalOAuth["url"] != nil && externalOAuth["authorization_header"] != nil {
			newExternalOAuth := platform.ExternalOAuth{
				Url:                 externalOAuth["url"].(string),
				AuthorizationHeader: externalOAuth["authorization_header"].(string),
			}
			input.Actions = append(
				input.Actions,
				&platform.ProjectSetExternalOAuthAction{ExternalOAuth: &newExternalOAuth})
		} else {
			input.Actions = append(input.Actions, &platform.ProjectSetExternalOAuthAction{ExternalOAuth: nil})
		}
	}

	if d.HasChange("enable_search_index_products") {
		value := d.Get("enable_search_index_products").(bool)
		action := platform.ProjectChangeProductSearchIndexingEnabledAction{
			Enabled: value,
		}
		input.Actions = append(input.Actions, action)
	}

	if d.HasChange("enable_search_index_orders") {
		value := d.Get("enable_search_index_orders")

		status := platform.OrderSearchStatusDeactivated
		if value.(bool) {
			status = platform.OrderSearchStatusActivated
		}
		action := platform.ProjectChangeOrderSearchStatusAction{
			Status: status,
		}
		input.Actions = append(input.Actions, action)
	}

	if d.HasChange("carts") {
		carts, err := elementFromList(d, "carts")
		if err != nil {
			return err
		}
		fallbackEnabled := false
		if carts["country_tax_rate_fallback_enabled"] != nil {
			fallbackEnabled = carts["country_tax_rate_fallback_enabled"].(bool)
		}

		var deleteDaysAfterLastModification *int
		if carts["delete_days_after_last_modification"] != nil {
			val := carts["delete_days_after_last_modification"].(int)
			deleteDaysAfterLastModification = &val
		}

		input.Actions = append(
			input.Actions,
			&platform.ProjectChangeCartsConfigurationAction{
				CartsConfiguration: &platform.CartsConfiguration{
					CountryTaxRateFallbackEnabled:   boolRef(fallbackEnabled),
					DeleteDaysAfterLastModification: deleteDaysAfterLastModification,
				},
			})
	}

	_, err := client.Post(input).Execute(context.Background())
	return err
}

func getStringSlice(d *schema.ResourceData, field string) []string {
	input := d.Get(field).([]interface{})
	var currencyObjects []string
	for _, raw := range input {
		currencyObjects = append(currencyObjects, raw.(string))
	}

	return currencyObjects
}

func getShippingRateInputType(d *schema.ResourceData) (platform.ShippingRateInputType, error) {
	switch d.Get("shipping_rate_input_type").(string) {
	case "CartValue":
		return platform.CartValueType{}, nil
	case "CartScore":
		return platform.CartScoreType{}, nil
	case "CartClassification":
		values, err := getCartClassificationValues(d)
		if err != nil {
			return "", fmt.Errorf("invalid cart classification value: %v, %w", values, err)
		}
		return platform.CartClassificationType{Values: values}, nil
	default:
		return "", fmt.Errorf("shipping rate input type %s not implemented", d.Get("shipping_rate_input_type").(string))
	}
}

func getCartClassificationValues(d *schema.ResourceData) ([]platform.CustomFieldLocalizedEnumValue, error) {
	var values []platform.CustomFieldLocalizedEnumValue
	data := d.Get("shipping_rate_cart_classification_value").([]interface{})
	for _, item := range data {
		itemMap := item.(map[string]interface{})
		label := platform.LocalizedString(expandStringMap(itemMap["label"].(map[string]interface{})))
		values = append(values, platform.CustomFieldLocalizedEnumValue{
			Label: label,
			Key:   itemMap["key"].(string),
		})
	}
	return values, nil
}

func marshallProjectCarts(val platform.CartsConfiguration) []map[string]interface{} {
	if *val.CountryTaxRateFallbackEnabled == false && val.DeleteDaysAfterLastModification == nil {
		return []map[string]interface{}{}
	}

	result := []map[string]interface{}{
		{
			"country_tax_rate_fallback_enabled":   val.CountryTaxRateFallbackEnabled,
			"delete_days_after_last_modification": val.DeleteDaysAfterLastModification,
		},
	}
	return result
}

func marshallProjectExternalOAuth(val *platform.ExternalOAuth) []map[string]interface{} {
	if val == nil {
		return []map[string]interface{}{}
	}
	return []map[string]interface{}{
		{
			"url":                  val.Url,
			"authorization_header": val.AuthorizationHeader,
		},
	}
}

func marshallProjectSearchIndexProducts(val *platform.SearchIndexingConfiguration) bool {
	if val == nil {
		return false
	}

	if val.Products != nil && val.Products.Status != nil {
		return *val.Products.Status != platform.SearchIndexingConfigurationStatusDeactivated
	}
	return false
}

func marshallProjectSearchIndexOrders(val *platform.SearchIndexingConfiguration) bool {
	if val == nil {
		return false
	}

	if val.Orders != nil && val.Orders.Status != nil {
		return *val.Orders.Status != platform.SearchIndexingConfigurationStatusDeactivated
	}
	return false
}

func marshallProjectShippingRateInputType(val platform.ShippingRateInputType) string {
	switch val.(type) {
	case platform.CartScoreType:
		return "CartScore"
	case platform.CartValueType:
		return "CartValue"
	case platform.CartClassificationType:
		return "CartClassification"
	}
	return ""
}

func marshallProjectMessages(val platform.MessagesConfiguration) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"enabled": val.Enabled,
		},
	}
}

func resourceProjectSettingsResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "The unique key of the project",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "The name of the project",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"currencies": {
				Description: "A three-digit currency code as per [ISO 4217](https://en.wikipedia.org/wiki/ISO_4217)",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"countries": {
				Description: "A two-digit country code as per [ISO 3166-1 alpha-2](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2)",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"languages": {
				Description: "[IETF Language Tag](https://en.wikipedia.org/wiki/IETF_language_tag)",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"messages": {
				Description: "[Messages Configuration](https://docs.commercetools.com/api/projects/project#messages-configuration)",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeBool,
				},
			},
			"external_oauth": {
				Description: "[External OAUTH](https://docs.commercetools.com/api/projects/project#externaloauth)",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"carts": {
				Description: "[Carts Configuration](https://docs.commercetools.com/api/projects/project#carts-configuration)",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"shipping_rate_input_type": {
				Description: "Three ways to dynamically select a ShippingRatePriceTier exist. The CartValue type uses " +
					"the sum of all line item prices, whereas CartClassification and CartScore use the " +
					"shippingRateInput field on the cart to select a tier",
				Type:     schema.TypeString,
				Optional: true,
			},
			"shipping_rate_cart_classification_value": {
				Description: "If shipping_rate_input_type is set to CartClassification these values are used to create " +
					"tiers\n. Only a key defined inside the values array can be used to create a tier, or to set a value " +
					"for the shippingRateInput on the cart. The keys are checked for uniqueness and the request is " +
					"rejected if keys are not unique",
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"label": {
							Type:     TypeLocalizedString,
							Optional: true,
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

func migrateResourceProjectSettingsStateV0toV1(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	transformToList(rawState, "messages")
	transformToList(rawState, "external_oauth")
	transformToList(rawState, "carts")
	return rawState, nil
}
