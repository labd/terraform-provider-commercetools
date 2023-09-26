package commercetools

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func resourceStore() *schema.Resource {
	return &schema.Resource{
		Description: "Stores can be used to model, for example, physical retail locations, brand stores, " +
			"or country-specific stores.\n\n" +
			"See also the [Stores API Documentation](https://docs.commercetools.com/api/projects/stores)",
		CreateContext: resourceStoreCreate,
		ReadContext:   resourceStoreRead,
		UpdateContext: resourceStoreUpdate,
		DeleteContext: resourceStoreDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "User-specific unique identifier for the store. The key is mandatory and immutable. " +
					"It is used to reference the store",
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Description:      "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
				Type:             TypeLocalizedString,
				ValidateDiagFunc: validateLocalizedStringKey,
				Optional:         true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"languages": {
				Description: "[IETF Language Tag](https://en.wikipedia.org/wiki/IETF_language_tag)",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"countries": {
				Description: "A two-digit country code as per " +
					"[ISO 3166-1 alpha-2](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2)",
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"distribution_channels": {
				Description: "Set of ResourceIdentifier to a Channel with ProductDistribution",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"supply_channels": {
				Description: "Set of ResourceIdentifier of Channels with InventorySupply",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"product_selection": {
				Description: "Controls availability of Products for this Store via Product Selections",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"active": {
							Description: "If true, all Products assigned to this Product Selection are part of the Store's assortment",
							Type:        schema.TypeBool,
							Required:    true,
						},
						"product_selection_id": {
							Description: "Resource Identifier of a ProductSelection",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
			"custom": CustomFieldSchema(),
		},
	}
}

func resourceStoreCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	name := expandLocalizedString(d.Get("name"))
	dcIdentifiers := expandStoreChannels(d.Get("distribution_channels"))
	scIdentifiers := expandStoreChannels(d.Get("supply_channels"))
	psIdentifiers := expandProductSelections(d.Get("product_selection").(*schema.Set))

	custom, err := CreateCustomFieldDraft(ctx, client, d)
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	draft := platform.StoreDraft{
		Key:                  d.Get("key").(string),
		Name:                 &name,
		Languages:            expandStringArray(d.Get("languages").([]any)),
		Countries:            expandStoreCountries(d.Get("countries").(*schema.Set)),
		DistributionChannels: dcIdentifiers,
		SupplyChannels:       scIdentifiers,
		ProductSelections:    psIdentifiers,
		Custom:               custom,
	}

	var store *platform.Store
	err = resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		var err error
		store, err = client.Stores().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(store.ID)
	_ = d.Set("version", store.Version)
	return resourceStoreRead(ctx, d, m)
}

func resourceStoreRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	store, err := client.Stores().
		WithId(d.Id()).
		Get().
		Expand([]string{"distributionChannels[*]", "supplyChannels[*]", "productSelections[*].productSelection"}).
		Execute(ctx)

	if err != nil {
		if utils.IsResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(store.ID)
	_ = d.Set("key", store.Key)
	if store.Name != nil {
		_ = d.Set("name", *store.Name)
	} else {
		_ = d.Set("name", nil)
	}
	_ = d.Set("version", store.Version)
	if store.Languages != nil {
		_ = d.Set("languages", store.Languages)
	}
	if store.Countries != nil {
		countries, err := flattenCountries(store.Countries)
		if err != nil {
			return diag.FromErr(err)
		}
		_ = d.Set("countries", countries)
	}
	if store.DistributionChannels != nil {
		channelKeys, err := flattenStoreChannels(store.DistributionChannels)
		if err != nil {
			return diag.FromErr(err)
		}
		_ = d.Set("distribution_channels", channelKeys)
	}

	if store.SupplyChannels != nil {
		channelKeys, err := flattenStoreChannels(store.SupplyChannels)
		if err != nil {
			return diag.FromErr(err)
		}
		_ = d.Set("supply_channels", channelKeys)
	}

	if store.ProductSelections != nil {
		selections, err := flattenProductSelections(store.ProductSelections)
		if err != nil {
			return diag.FromErr(err)
		}
		_ = d.Set("product_selection", selections)
	}

	_ = d.Set("custom", flattenCustomFields(store.Custom))
	return nil
}

func resourceStoreUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	input := platform.StoreUpdate{
		Version: d.Get("version").(int),
		Actions: []platform.StoreUpdateAction{},
	}

	if d.HasChange("name") {
		newName := expandLocalizedString(d.Get("name"))
		input.Actions = append(
			input.Actions,
			&platform.StoreSetNameAction{Name: &newName})
	}

	if d.HasChange("languages") {
		languages := expandStringArray(d.Get("languages").([]any))

		input.Actions = append(
			input.Actions,
			&platform.StoreSetLanguagesAction{Languages: languages})
	}

	if d.HasChange("countries") {
		countries := expandStoreCountries(d.Get("countries").(*schema.Set))

		input.Actions = append(
			input.Actions,
			&platform.StoreSetCountriesAction{Countries: countries})
	}

	if d.HasChange("distribution_channels") {
		dcIdentifiers := expandStoreChannels(d.Get("distribution_channels"))

		// set action replaces current values
		input.Actions = append(
			input.Actions,
			&platform.StoreSetDistributionChannelsAction{
				DistributionChannels: dcIdentifiers,
			},
		)
	}

	if d.HasChange("supply_channels") {
		scIdentifiers := expandStoreChannels(d.Get("supply_channels"))
		// set action replaces current values
		input.Actions = append(
			input.Actions,
			&platform.StoreSetSupplyChannelsAction{
				SupplyChannels: scIdentifiers,
			},
		)
	}

	if d.HasChange("product_selection") {
		old, new := d.GetChange("product_selection")

		oldProductSelections := expandProductSelections(old.(*schema.Set))
		newProductSelections := expandProductSelections(new.(*schema.Set))

		for i, productSelection := range oldProductSelections {
			if !productSelectionInSlice(productSelection, newProductSelections) {
				input.Actions = append(
					input.Actions,
					&platform.StoreRemoveProductSelectionAction{ProductSelection: oldProductSelections[i].ProductSelection})
			}
		}
		for i, location := range newProductSelections {
			if !productSelectionInSlice(location, oldProductSelections) {
				input.Actions = append(
					input.Actions,
					&platform.StoreAddProductSelectionAction{
						ProductSelection: newProductSelections[i].ProductSelection,
						Active:           newProductSelections[i].Active,
					})
			}
		}
	}

	if d.HasChange("custom") {

		actions, err := CustomFieldUpdateActions[platform.StoreSetCustomTypeAction, platform.StoreSetCustomFieldAction](ctx, client, d)
		if err != nil {
			// Workaround invalid state to be written, see
			// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
			d.Partial(true)
			return diag.FromErr(err)
		}
		for i := range actions {
			input.Actions = append(input.Actions, actions[i].(platform.StoreUpdateAction))
		}
	}

	err := resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		_, err := client.Stores().WithId(d.Id()).Post(input).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	return resourceStoreRead(ctx, d, m)
}

func resourceStoreDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	version := d.Get("version").(int)

	err := resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		_, err := client.Stores().WithId(d.Id()).Delete().Version(version).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	return diag.FromErr(err)
}

func convertChannelKeysToIdentifiers(channelKeys []string) []platform.ChannelResourceIdentifier {
	identifiers := make([]platform.ChannelResourceIdentifier, 0)
	for i := 0; i < len(channelKeys); i++ {
		channelIdentifier := platform.ChannelResourceIdentifier{
			Key: &channelKeys[i],
		}
		identifiers = append(identifiers, channelIdentifier)
	}
	return identifiers
}

func expandStoreChannels(channelData any) []platform.ChannelResourceIdentifier {
	channelKeys := expandStringArray(channelData.([]any))
	return convertChannelKeysToIdentifiers(channelKeys)
}

func expandProductSelections(input *schema.Set) []platform.ProductSelectionSettingDraft {
	inputSlice := input.List()
	result := make([]platform.ProductSelectionSettingDraft, len(inputSlice))

	for i := range inputSlice {
		raw := inputSlice[i].(map[string]any)
		active, ok := raw["active"].(bool)
		if !ok {
			active = false
		}

		var productSelectionRef *string
		if productSelection, ok := raw["product_selection_id"].(string); ok && productSelection != "" {
			productSelectionRef = &productSelection
		}

		result[i] = platform.ProductSelectionSettingDraft{
			Active: utils.BoolRef(active),
			ProductSelection: platform.ProductSelectionResourceIdentifier{
				ID: productSelectionRef,
			},
		}
	}

	return result
}

func convertCountryCodesToStoreCountries(countryCodes []string) []platform.StoreCountry {
	storeCountries := make([]platform.StoreCountry, 0)
	for i := 0; i < len(countryCodes); i++ {
		storeCountry := platform.StoreCountry{
			Code: countryCodes[i],
		}
		storeCountries = append(storeCountries, storeCountry)
	}
	return storeCountries
}

func expandStoreCountries(countryData any) []platform.StoreCountry {
	countryCodes := expandStringArray(countryData.(*schema.Set).List())
	return convertCountryCodesToStoreCountries(countryCodes)
}

func flattenStoreChannels(channels []platform.ChannelReference) ([]string, error) {
	channelKeys := make([]string, 0)
	for i := 0; i < len(channels); i++ {
		if channels[i].Obj == nil {
			return nil, errors.New("failed to expand channel objects")
		}
		channelKeys = append(channelKeys, channels[i].Obj.Key)
	}
	return channelKeys, nil
}

func flattenProductSelections(selections []platform.ProductSelectionSetting) ([]map[string]any, error) {
	result := make([]map[string]any, len(selections))

	for i := range selections {
		result[i] = map[string]any{
			"active":               selections[i].Active,
			"product_selection_id": selections[i].ProductSelection.ID,
		}
	}

	return result, nil
}

func flattenCountries(countries []platform.StoreCountry) ([]string, error) {
	countryCodes := make([]string, 0)
	for _, country := range countries {
		countryCodes = append(countryCodes, country.Code)
	}
	return countryCodes, nil
}

func productSelectionInSlice(needle platform.ProductSelectionSettingDraft, haystack []platform.ProductSelectionSettingDraft) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
