package commercetools

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

// TODO: A lot of fields are optional in this schema that are not optional in commercetools. When not set via terraform
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
			State: schema.ImportStatePassthrough,
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
				Type:        schema.TypeMap,
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
			"external_oauth": {
				Description: "[External OAUTH](https://docs.commercetools.com/api/projects/project#externaloauth)",
				Type:        schema.TypeMap,
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
				Type:        schema.TypeMap,
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
					}},
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

	_, err := client.ProjectGet()
	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
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
	project, err := client.ProjectGet()

	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
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

	project, err := client.ProjectGet()

	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
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
	d.Set("external_oauth", project.ExternalOAuth)
	d.Set("carts", project.Carts)
	log.Print("[DEBUG] Logging messages enabled")
	log.Print(stringFormatObject(project.Messages))
	d.Set("messages", project.Messages)
	log.Print(stringFormatObject(d))
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

func projectUpdate(d *schema.ResourceData, client *commercetools.Client, version int) error {
	input := &commercetools.ProjectUpdateInput{
		Version: version,
		Actions: []commercetools.ProjectUpdateAction{},
	}

	if d.HasChange("name") {
		input.Actions = append(input.Actions, &commercetools.ProjectChangeNameAction{Name: d.Get("name").(string)})
	}

	if d.HasChange("currencies") {
		newCurrencies := []commercetools.CurrencyCode{}
		for _, item := range getStringSlice(d, "currencies") {
			newCurrencies = append(newCurrencies, commercetools.CurrencyCode(item))
		}

		input.Actions = append(
			input.Actions,
			&commercetools.ProjectChangeCurrenciesAction{Currencies: newCurrencies})
	}

	if d.HasChange("countries") {
		newCountries := []commercetools.CountryCode{}
		for _, item := range getStringSlice(d, "countries") {
			newCountries = append(newCountries, commercetools.CountryCode(item))
		}

		input.Actions = append(
			input.Actions,
			&commercetools.ProjectChangeCountriesAction{Countries: newCountries})
	}

	if d.HasChange("languages") {
		newLanguages := []commercetools.Locale{}
		for _, item := range getStringSlice(d, "languages") {
			newLanguages = append(newLanguages, commercetools.Locale(item))
		}
		input.Actions = append(
			input.Actions,
			&commercetools.ProjectChangeLanguagesAction{Languages: newLanguages})
	}

	if d.HasChange("messages") {
		messages := d.Get("messages").(map[string]interface{})
		if messages["enabled"] != nil {
			// boolean value is somehow interface{} | string so we have to convert it
			var enabled bool
			switch messages["enabled"] {
			case "true":
				enabled = true
			case "false":
				enabled = false
			default:
				return fmt.Errorf("invalid value for messages[\"enabled\"]: %t", messages["enabled"])
			}

			input.Actions = append(
				input.Actions,
				&commercetools.ProjectChangeMessagesEnabledAction{MessagesEnabled: enabled})
		} else {
			// To commercetools this field is not optional, so when deleting we revert to the default: false:
			input.Actions = append(
				input.Actions,
				&commercetools.ProjectChangeMessagesEnabledAction{MessagesEnabled: false})
		}

	}

	if d.HasChange("external_oauth") {
		externalOAuth := d.Get("external_oauth").(map[string]interface{})
		if externalOAuth["url"] != nil && externalOAuth["authorization_header"] != nil {
			newExternalOAuth := commercetools.ExternalOAuth{
				URL:                 externalOAuth["url"].(string),
				AuthorizationHeader: externalOAuth["authorization_header"].(string),
			}
			input.Actions = append(
				input.Actions,
				&commercetools.ProjectSetExternalOAuthAction{ExternalOAuth: &newExternalOAuth})
		} else {
			input.Actions = append(input.Actions, &commercetools.ProjectSetExternalOAuthAction{ExternalOAuth: nil})
		}
	}

	if d.HasChange("carts") {
		carts := d.Get("carts").(map[string]interface{})
		var fallbackEnabled bool = false
		if carts["country_tax_rate_fallback_enabled"] != nil {
			// boolean value is somehow interface{} | string so we have to convert it
			switch carts["country_tax_rate_fallback_enabled"] {
			case "true":
				fallbackEnabled = true
			case "false":
			default:
				return fmt.Errorf("invalid value for carts[\"country_tax_rate_fallback_enabled\"]: %s", carts["country_tax_rate_fallback_enabled"])
			}
		}

		input.Actions = append(
			input.Actions,
			&commercetools.ProjectChangeCountryTaxRateFallbackEnabledAction{
				CountryTaxRateFallbackEnabled: fallbackEnabled,
			})

		var deleteDaysAfterLastModification int
		if carts["delete_days_after_last_modification"] != nil {
			// int value is somehow interface{} | string so we have to convert the value
			var err error
			deleteDaysAfterLastModification, err = strconv.Atoi(carts["delete_days_after_last_modification"].(string))
			if err != nil {
				return fmt.Errorf("invalid value for carts[\"delete_days_after_last_modification\"]: %s", carts["delete_days_after_last_modification"])
			}
		}

		input.Actions = append(
			input.Actions,
			&commercetools.ProjectChangeCartsConfiguration{
				CartsConfiguration: &commercetools.CartsConfiguration{
					DeleteDaysAfterLastModification: deleteDaysAfterLastModification,
				},
			})
	}

	_, err := client.ProjectUpdate(input)
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
