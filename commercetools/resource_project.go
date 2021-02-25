package commercetools

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

// TODO: A lot of fields are optional in this schema that are not optional in commercetools. When not set via terraform
// commercetools simply sets the default values for these fields. This works but can be a little confusing. It is worth
// considering whether to align the optional/required status of the fields in the provider with that of the API itself
func resourceProjectSettings() *schema.Resource {
	return &schema.Resource{
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
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"currencies": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"countries": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"languages": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"messages": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			"external_oauth": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"authorization_header": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"carts": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"country_tax_rate_fallback_enabled": {
							Type:     schema.TypeBool,
							Required: true,
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
	// d.Set("createdAt", project.CreatedAt)
	// d.Set("trialUntil", project.TrialUntil)
	log.Print("[DEBUG] Logging messages enabled")
	log.Print(stringFormatObject(project.Messages))
	d.Set("messages", project.Messages)
	log.Print(stringFormatObject(d))
	// d.Set("shippingRateInputType", project.ShippingRateInputType)
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
			// boolean value is somehow interface{} | string so....
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
		if carts["country_tax_rate_fallback_enabled"] != nil {
			// boolean value is somehow interface{} | string so....
			var fallbackEnabled bool
			switch carts["country_tax_rate_fallback_enabled"] {
			case "true":
				fallbackEnabled = true
			case "false":
				fallbackEnabled = false
			default:
				return fmt.Errorf("invalid value for carts[\"country_tax_rate_fallback_enabled\"]: %s", carts["country_tax_rate_fallback_enabled"])
			}
			input.Actions = append(
				input.Actions,
				&commercetools.ProjectChangeCountryTaxRateFallbackEnabledAction{
					CountryTaxRateFallbackEnabled: fallbackEnabled,
				})
		} else {
			// To commercetools this field is not optional, so when deleting we revert to the default: false:
			input.Actions = append(
				input.Actions,
				&commercetools.ProjectChangeCountryTaxRateFallbackEnabledAction{
					CountryTaxRateFallbackEnabled: false,
				})
		}

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
