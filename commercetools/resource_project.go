package commercetools

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/labd/commercetools-go-sdk/platform"
)

type resourceProjectSettingsType struct{}

// // TODO: A lot of fields are optional in this schema that are not optional in platform. When not set via terraform
// // commercetools simply sets the default values for these fields. This works but can be a little confusing. It is worth
// // considering whether to align the optional/required status of the fields in the provider with that of the API itself
func (r resourceProjectSettingsType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{

		// Description: "The project endpoint provides a limited set of information about settings and configuration of " +
		// 	"the project. Updating the settings is eventually consistent, it may take up to a minute before " +
		// 	"a change becomes fully active.\n\n" +
		// 	"See also the [Project Settings API Documentation](https://docs.commercetools.com/api/projects/project)",
		// Create: resourceProjectCreate,
		// Read:   resourceProjectRead,
		// Update: resourceProjectUpdate,
		// Delete: resourceProjectDelete,
		// Exists: resourceProjectExists,
		// Importer: &schema.ResourceImporter{
		// 	State: schema.ImportStatePassthrough,
		// },
		// SchemaVersion: 1,
		// StateUpgraders: []schema.StateUpgrader{
		// 	{
		// 		Type:    resourceProjectSettingsResourceV0().CoreConfigSchema().ImpliedType(),
		// 		Upgrade: migrateResourceProjectSettingsStateV0toV1,
		// 		Version: 0,
		// 	},
		// },
		Attributes: map[string]tfsdk.Attribute{
			"key": {
				Description: "The unique key of the project",
				Type:        types.StringType,
				Computed:    true,
			},
			"name": {
				Description: "The name of the project",
				Type:        types.StringType,
				Optional:    true,
			},
			"currencies": {
				Description: "A three-digit currency code as per [ISO 4217](https://en.wikipedia.org/wiki/ISO_4217)",
				Type:        types.ListType{ElemType: types.StringType},
				Optional:    true,
			},
			"countries": {
				Description: "A two-digit country code as per [ISO 3166-1 alpha-2](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2)",
				Type:        types.ListType{ElemType: types.StringType},
				Optional:    true,
			},
			"languages": {
				Description: "[IETF Language Tag](https://en.wikipedia.org/wiki/IETF_language_tag)",
				Type:        types.ListType{ElemType: types.StringType},
				Optional:    true,
			},
			"messages": {
				Description: "[Messages Configuration](https://docs.commercetools.com/api/projects/project#messages-configuration)",
				Optional:    true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"enabled": {
						Description: "When true the creation of messages on the Messages Query HTTP API is enabled",
						Type:        types.BoolType,
						Required:    true,
					},
				}),
			},
			"external_oauth": {
				Description: "[External OAUTH](https://docs.commercetools.com/api/projects/project#externaloauth)",
				Type:        types.ListType{ElemType: types.StringType},
				Optional:    true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"url": {
						Type:     types.StringType,
						Required: true,
					},
					"authorization_header": {
						Description: "Partially hidden on retrieval",
						Type:        types.StringType,
						Required:    true,
					},
				}),
			},
			"carts": {
				Description: "[Carts Configuration](https://docs.commercetools.com/api/projects/project#carts-configuration)",
				Type:        types.ListType{ElemType: types.StringType},
				Optional:    true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"country_tax_rate_fallback_enabled": {
						Description: "Indicates if country - no state tax rate fallback should be used when a " +
							"shipping address state is not explicitly covered in the rates lists of all tax " +
							"categories of a cart line items",
						Type:     types.BoolType,
						Required: true,
					},
					"delete_days_after_last_modification": {
						Description: "Number - Optional The default value for the " +
							"deleteDaysAfterLastModification parameter of the CartDraft. Initially set to 90 for " +
							"projects created after December 2019.",
						Type:     types.Int64Type,
						Optional: true,
					},
				}),
			},
			"shipping_rate_input_type": {
				Description: "Three ways to dynamically select a ShippingRatePriceTier exist. The CartValue type uses " +
					"the sum of all line item prices, whereas CartClassification and CartScore use the " +
					"shippingRateInput field on the cart to select a tier",
				Type:     types.StringType,
				Optional: true,
			},
			"shipping_rate_cart_classification_value": {
				Description: "If shipping_rate_input_type is set to CartClassification these values are used to create " +
					"tiers\n. Only a key defined inside the values array can be used to create a tier, or to set a value " +
					"for the shippingRateInput on the cart. The keys are checked for uniqueness and the request is " +
					"rejected if keys are not unique",
				Type:     types.ListType{ElemType: types.StringType},
				Optional: true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"key": {
						Type:     types.StringType,
						Required: true,
					},
					"label": {
						Type:     TypeLocalizedString(),
						Optional: true,
					},
				}, tfsdk.ListNestedAttributesOptions{}),
			},
			"version": {
				Type:     types.Int64Type,
				Computed: true,
			},
		},
	}, nil
}

type ProjectSettings struct {
	Key           types.String                 `tfsdk:"key"`
	Name          types.String                 `tfsdk:"name"`
	Version       types.Int64                  `tfsdk:"version"`
	Currencies    []types.String               `tfsdk:"currencies"`
	Countries     []types.String               `tfsdk:"countries"`
	Languages     []types.String               `tfsdk:"languages"`
	Message       ProjectSettingsMessage       `tfsdk:"messages"`
	ExternalOAuth ProjectSettingsExternalOAuth `tfsdk:"external_oauth"`
}

type ProjectSettingsMessage struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

type ProjectSettingsExternalOAuth struct {
	URL                 types.String `tfsdk:"url"`
	AuthorizationHeader types.String `tfsdk:"authorization_header"`
}

// New resource instance
func (r resourceProjectSettingsType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceProjectSettings{
		p: *(p.(*provider)),
	}, nil
}

type resourceProjectSettings struct {
	p provider
}

// func resourceProjectExists(d *schema.ResourceData, m interface{}) (bool, error) {
// 	client := getClient(m)

// 	_, err := client.Get().Execute(context.Background())
// 	if err != nil {
// 		if ctErr, ok := err.(platform.ErrorResponse); ok {
// 			if ctErr.StatusCode == 404 {
// 				return false, nil
// 			}
// 		}
// 		return false, err
// 	}
// 	return true, nil
// }

func (r resourceProjectSettings) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan ProjectSettings
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.p.client.Get().Execute(context.Background())

	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				return
			}
		}

		resp.Diagnostics.AddError("Unable to retrieve project settings", err.Error())
		return
	}

	err = r.projectUpdate(&plan, project)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	return resourceProjectRead(d, m)
}

func (r resourceProjectSettings) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	log.Print("[DEBUG] Reading projects from commercetools")
	var state ProjectSettings
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.p.client.Get().Execute(context.Background())

	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				return
			}
		}
		resp.Diagnostics.AddError("Unable to retrieve project settings", err.Error())
	}

	log.Print("[DEBUG] Found the following project:")
	log.Print(stringFormatObject(project))

	state = ProjectSettings{
		Key:        types.String{Value: project.Key},
		Version:    types.Int64{Value: int64(project.Version)},
		Name:       types.String{Value: project.Name},
		Countries:  transformStringSlice(project.Countries),
		Currencies: transformStringSlice(project.Currencies),
		Languages:  transformStringSlice(project.Languages),
		ExternalOAuth: ProjectSettingsExternalOAuth{
			URL:                 types.String{Value: project.ExternalOAuth.Url},
			AuthorizationHeader: types.String{Value: project.ExternalOAuth.AuthorizationHeader},
		},
		Message: ProjectSettingsMessage{
			Enabled: types.Bool{Value: project.Messages.Enabled},
		},
	}

	// d.Set("shipping_rate_input_type", marshallProjectShippingRateInputType(project.ShippingRateInputType))
	// d.Set("carts", marshallProjectCarts(project.Carts))

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceProjectSettings) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {

	var plan ProjectSettings
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ProjectSettings
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.p.client.Get().Execute(context.Background())
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve project settings", err.Error())
	}

	r.projectUpdate(&plan, project)

	// Set state
	// diags = resp.State.Set(ctx, result)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// func resourceProjectDelete(d *schema.ResourceData, m interface{}) error {
	// 	d.SetId("")
	// 	return nil
}

func (r resourceProjectSettings) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	// Simply remove resource from state since we cannot delete a project
	resp.State.RemoveResource(ctx)
}

// Import resource
func (r resourceProjectSettings) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Save the import identifier in the key attribute
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("key"), req, resp)
}

func (r resourceProjectSettings) projectUpdate(plan *ProjectSettings, obj *platform.Project) error {
	// input := platform.ProjectUpdate{
	// 	Version: obj.Version,
	// 	Actions: []platform.ProjectUpdateAction{},
	// }

	// if d.HasChange("name") {
	// 	input.Actions = append(input.Actions, &platform.ProjectChangeNameAction{Name: d.Get("name").(string)})
	// }

	// if d.HasChange("currencies") {
	// 	newCurrencies := []string{}
	// 	for _, item := range getStringSlice(d, "currencies") {
	// 		newCurrencies = append(newCurrencies, item)
	// 	}

	// 	input.Actions = append(
	// 		input.Actions,
	// 		&platform.ProjectChangeCurrenciesAction{Currencies: newCurrencies})
	// }

	// if d.HasChange("countries") {
	// 	newCountries := []string{}
	// 	for _, item := range getStringSlice(d, "countries") {
	// 		newCountries = append(newCountries, item)
	// 	}

	// 	input.Actions = append(
	// 		input.Actions,
	// 		&platform.ProjectChangeCountriesAction{Countries: newCountries})
	// }

	// if d.HasChange("languages") {
	// 	newLanguages := []string{}
	// 	for _, item := range getStringSlice(d, "languages") {
	// 		newLanguages = append(newLanguages, item)
	// 	}
	// 	input.Actions = append(
	// 		input.Actions,
	// 		&platform.ProjectChangeLanguagesAction{Languages: newLanguages})
	// }

	// if d.HasChange("messages") {
	// 	messages, err := elementFromList(d, "messages")
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if messages["enabled"] != nil {
	// 		input.Actions = append(
	// 			input.Actions,
	// 			&platform.ProjectChangeMessagesEnabledAction{MessagesEnabled: messages["enabled"].(bool)})
	// 	} else {
	// 		// To commercetools this field is not optional, so when deleting we revert to the default: false:
	// 		input.Actions = append(
	// 			input.Actions,
	// 			&platform.ProjectChangeMessagesEnabledAction{MessagesEnabled: false})
	// 	}

	// }

	// if d.HasChange("shipping_rate_input_type") || d.HasChange("shipping_rate_cart_classification_value") {
	// 	newShippingRateInputType, err := getShippingRateInputType(d)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	input.Actions = append(
	// 		input.Actions,
	// 		&platform.ProjectSetShippingRateInputTypeAction{ShippingRateInputType: &newShippingRateInputType})
	// }

	// if d.HasChange("external_oauth") {
	// 	externalOAuth, err := elementFromList(d, "external_oauth")
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if externalOAuth["url"] != nil && externalOAuth["authorization_header"] != nil {
	// 		newExternalOAuth := platform.ExternalOAuth{
	// 			Url:                 externalOAuth["url"].(string),
	// 			AuthorizationHeader: externalOAuth["authorization_header"].(string),
	// 		}
	// 		input.Actions = append(
	// 			input.Actions,
	// 			&platform.ProjectSetExternalOAuthAction{ExternalOAuth: &newExternalOAuth})
	// 	} else {
	// 		input.Actions = append(input.Actions, &platform.ProjectSetExternalOAuthAction{ExternalOAuth: nil})
	// 	}
	// }

	// if d.HasChange("carts") {
	// 	carts, err := elementFromList(d, "carts")
	// 	if err != nil {
	// 		return err
	// 	}
	// 	fallbackEnabled := false
	// 	if carts["country_tax_rate_fallback_enabled"] != nil {
	// 		fallbackEnabled = carts["country_tax_rate_fallback_enabled"].(bool)
	// 	}

	// 	var deleteDaysAfterLastModification *int
	// 	if carts["delete_days_after_last_modification"] != nil {
	// 		val := carts["delete_days_after_last_modification"].(int)
	// 		deleteDaysAfterLastModification = &val
	// 	}

	// 	input.Actions = append(
	// 		input.Actions,
	// 		&platform.ProjectChangeCartsConfiguration{
	// 			CartsConfiguration: &platform.CartsConfiguration{
	// 				CountryTaxRateFallbackEnabled:   boolRef(fallbackEnabled),
	// 				DeleteDaysAfterLastModification: deleteDaysAfterLastModification,
	// 			},
	// 		})
	// }

	// _, err := client.Post(input).Execute(context.Background())
	// return err
	return nil
}

// func getStringSlice(d *schema.ResourceData, field string) []string {
// 	input := d.Get(field).([]interface{})
// 	var currencyObjects []string
// 	for _, raw := range input {
// 		currencyObjects = append(currencyObjects, raw.(string))
// 	}

// 	return currencyObjects
// }

// func getShippingRateInputType(d *schema.ResourceData) (platform.ShippingRateInputType, error) {
// 	switch d.Get("shipping_rate_input_type").(string) {
// 	case "CartValue":
// 		return platform.CartValueType{}, nil
// 	case "CartScore":
// 		return platform.CartScoreType{}, nil
// 	case "CartClassification":
// 		values, err := getCartClassificationValues(d)
// 		if err != nil {
// 			return "", fmt.Errorf("invalid cart classification value: %v, %w", values, err)
// 		}
// 		return platform.CartClassificationType{Values: values}, nil
// 	default:
// 		return "", fmt.Errorf("shipping rate input type %s not implemented", d.Get("shipping_rate_input_type").(string))
// 	}
// }

// func getCartClassificationValues(d *schema.ResourceData) ([]platform.CustomFieldLocalizedEnumValue, error) {
// 	var values []platform.CustomFieldLocalizedEnumValue
// 	data := d.Get("shipping_rate_cart_classification_value").([]interface{})
// 	for _, item := range data {
// 		itemMap := item.(map[string]interface{})
// 		label := platform.LocalizedString(expandStringMap(itemMap["label"].(map[string]interface{})))
// 		values = append(values, platform.CustomFieldLocalizedEnumValue{
// 			Label: label,
// 			Key:   itemMap["key"].(string),
// 		})
// 	}
// 	return values, nil
// }

// func marshallProjectCarts(val platform.CartsConfiguration) []map[string]interface{} {
// 	return []map[string]interface{}{
// 		{
// 			"country_tax_rate_fallback_enabled":   val.CountryTaxRateFallbackEnabled,
// 			"delete_days_after_last_modification": val.DeleteDaysAfterLastModification,
// 		},
// 	}
// }

// func marshallProjectExternalOAuth(val *platform.ExternalOAuth) []map[string]interface{} {
// 	if val == nil {
// 		return []map[string]interface{}{}
// 	}
// 	return []map[string]interface{}{
// 		{
// 			"url":                  val.Url,
// 			"authorization_header": val.AuthorizationHeader,
// 		},
// 	}
// }

// func marshallProjectShippingRateInputType(val platform.ShippingRateInputType) string {
// 	switch val.(type) {
// 	case platform.CartScoreType:
// 		return "CartScore"
// 	case platform.CartValueType:
// 		return "CartValue"
// 	case platform.CartClassificationType:
// 		return "CartClassification"
// 	}
// 	return ""
// }

// func marshallProjectMessages(val platform.MessageConfiguration) []map[string]interface{} {
// 	return []map[string]interface{}{
// 		{
// 			"enabled": val.Enabled,
// 		},
// 	}
// }

// func resourceProjectSettingsResourceV0() *schema.Resource {
// 	return &schema.Resource{
// 		Schema: map[string]*schema.Schema{
// 			"key": {
// 				Description: "The unique key of the project",
// 				Type:        types.StringType,
// 				Computed:    true,
// 			},
// 			"name": {
// 				Description: "The name of the project",
// 				Type:        types.StringType,
// 				Optional:    true,
// 			},
// 			"currencies": {
// 				Description: "A three-digit currency code as per [ISO 4217](https://en.wikipedia.org/wiki/ISO_4217)",
// 				Type:        types.ListType{ElemType: types.StringType},
// 				Optional:    true,
// 				Elem:        &schema.Schema{Type: types.StringType},
// 			},
// 			"countries": {
// 				Description: "A two-digit country code as per [ISO 3166-1 alpha-2](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2)",
// 				Type:        types.ListType{ElemType: types.StringType},
// 				Optional:    true,
// 				Elem:        &schema.Schema{Type: types.StringType},
// 			},
// 			"languages": {
// 				Description: "[IETF Language Tag](https://en.wikipedia.org/wiki/IETF_language_tag)",
// 				Type:        types.ListType{ElemType: types.StringType},
// 				Optional:    true,
// 				Elem:        &schema.Schema{Type: types.StringType},
// 			},
// 			"messages": {
// 				Description: "[Messages Configuration](https://docs.commercetools.com/api/projects/project#messages-configuration)",
// 				Type:        schema.TypeMap,
// 				Optional:    true,
// 				Elem: &schema.Schema{
// 					Type: types.BoolType,
// 				},
// 			},
// 			"external_oauth": {
// 				Description: "[External OAUTH](https://docs.commercetools.com/api/projects/project#externaloauth)",
// 				Type:        schema.TypeMap,
// 				Optional:    true,
// 				Elem: &schema.Schema{
// 					Type: types.StringType,
// 				},
// 			},
// 			"carts": {
// 				Description: "[Carts Configuration](https://docs.commercetools.com/api/projects/project#carts-configuration)",
// 				Type:        schema.TypeMap,
// 				Optional:    true,
// 				Elem: &schema.Schema{
// 					Type: types.StringType,
// 				},
// 			},
// 			"shipping_rate_input_type": {
// 				Description: "Three ways to dynamically select a ShippingRatePriceTier exist. The CartValue type uses " +
// 					"the sum of all line item prices, whereas CartClassification and CartScore use the " +
// 					"shippingRateInput field on the cart to select a tier",
// 				Type:     types.StringType,
// 				Optional: true,
// 			},
// 			"shipping_rate_cart_classification_value": {
// 				Description: "If shipping_rate_input_type is set to CartClassification these values are used to create " +
// 					"tiers\n. Only a key defined inside the values array can be used to create a tier, or to set a value " +
// 					"for the shippingRateInput on the cart. The keys are checked for uniqueness and the request is " +
// 					"rejected if keys are not unique",
// 				Type:     types.ListType{ElemType: types.StringType},
// 				Optional: true,
// 				Elem: &schema.Resource{
// 					Schema: map[string]*schema.Schema{
// 						"key": {
// 							Type:     types.StringType,
// 							Required: true,
// 						},
// 						"label": {
// 							Type:     TypeLocalizedString,
// 							Optional: true,
// 						},
// 					},
// 				},
// 			},
// 			"version": {
// 				Type:     types.Int64Type,
// 				Computed: true,
// 			},
// 		},
// 	}
// }

// func migrateResourceProjectSettingsStateV0toV1(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
// 	transformToList(rawState, "messages")
// 	transformToList(rawState, "external_oauth")
// 	transformToList(rawState, "carts")
// 	return rawState, nil
// }

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
