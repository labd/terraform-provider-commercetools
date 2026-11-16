package commercetools

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
		CustomizeDiff: func(_ context.Context, d *schema.ResourceDiff, _ any) error {
			return validateProductSelectionsConfig(d.GetRawConfig())
		},
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
							Description: "ID of the ProductSelection. Exactly one of `product_selection_id` or `product_selection_key` must be set",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"product_selection_key": {
							Description: "Key of the ProductSelection. Exactly one of `product_selection_id` or `product_selection_key` must be set",
							Type:        schema.TypeString,
							Optional:    true,
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
	psIdentifiers, err := expandProductSelections(d.Get("product_selection").(*schema.Set))
	if err != nil {
		return diag.FromErr(err)
	}

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
	err = retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
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
		selections, err := flattenProductSelections(
			ctx,
			d.Get("product_selection").(*schema.Set),
			store.ProductSelections,
			func(ctx context.Context, id string) (string, error) {
				return fetchProductSelectionKey(ctx, client, id)
			},
		)
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
		o, n := d.GetChange("product_selection")

		oldProductSelections, err := expandProductSelections(o.(*schema.Set))
		if err != nil {
			// Keep the prior state, nothing was sent to the API, see
			// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
			d.Partial(true)
			return diag.FromErr(err)
		}
		newProductSelections, err := expandProductSelections(n.(*schema.Set))
		if err != nil {
			d.Partial(true)
			return diag.FromErr(err)
		}

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

	err := retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
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

	err := retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
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

func expandProductSelections(input *schema.Set) ([]platform.ProductSelectionSettingDraft, error) {
	inputSlice := input.List()
	result := make([]platform.ProductSelectionSettingDraft, len(inputSlice))

	for i := range inputSlice {
		raw := inputSlice[i].(map[string]any)
		active, ok := raw["active"].(bool)
		if !ok {
			active = false
		}

		var productSelectionID, productSelectionKey *string
		if id, ok := raw["product_selection_id"].(string); ok && id != "" {
			productSelectionID = &id
		}
		if key, ok := raw["product_selection_key"].(string); ok && key != "" {
			productSelectionKey = &key
		}
		if (productSelectionID == nil) == (productSelectionKey == nil) {
			return nil, errors.New("product_selection: exactly one of product_selection_id or product_selection_key must be set")
		}

		result[i] = platform.ProductSelectionSettingDraft{
			Active: utils.BoolRef(active),
			ProductSelection: platform.ProductSelectionResourceIdentifier{
				ID:  productSelectionID,
				Key: productSelectionKey,
			},
		}
	}

	return result, nil
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

// productSelectionKeyResolver returns the key of the product selection with
// the given ID.
type productSelectionKeyResolver func(ctx context.Context, id string) (string, error)

// fetchProductSelectionKey looks up the key of a product selection by ID. It is
// used when the store response contains a product selection reference that was
// not expanded (commercetools allows incomplete reference expansion).
func fetchProductSelectionKey(ctx context.Context, client *platform.ByProjectKeyRequestBuilder, id string) (string, error) {
	var ps *platform.ProductSelection
	err := retry.RetryContext(ctx, 20*time.Second, func() *retry.RetryError {
		var err error
		ps, err = client.ProductSelections().WithId(id).Get().Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		return "", err
	}
	if ps.Key == nil {
		return "", nil
	}
	return *ps.Key, nil
}

// flattenProductSelections maps the API response back to state. A selection
// is written back by key when the prior state/config referenced it by key,
// otherwise by id, so the practitioner's chosen identifier does not drift.
// If a reference was not expanded and key-based identifiers are in use, the
// key is resolved with resolveKey; failures are returned instead of silently
// switching to the id.
func flattenProductSelections(
	ctx context.Context,
	prior *schema.Set,
	selections []platform.ProductSelectionSetting,
	resolveKey productSelectionKeyResolver,
) ([]map[string]any, error) {
	usedKeys := map[string]struct{}{}
	usedIDs := map[string]struct{}{}
	for _, v := range prior.List() {
		raw, ok := v.(map[string]any)
		if !ok {
			continue
		}
		if key, ok := raw["product_selection_key"].(string); ok && key != "" {
			usedKeys[key] = struct{}{}
		}
		if id, ok := raw["product_selection_id"].(string); ok && id != "" {
			usedIDs[id] = struct{}{}
		}
	}

	result := make([]map[string]any, len(selections))
	for i := range selections {
		item := map[string]any{"active": selections[i].Active}
		ref := selections[i].ProductSelection

		var key string
		switch {
		case ref.Obj != nil:
			if ref.Obj.Key != nil {
				key = *ref.Obj.Key
			}
		case len(usedKeys) > 0 && !isUsedByID(usedIDs, ref.ID):
			// Only look the key up for references that were not configured by
			// id; selections configured by id are preserved as they are.
			var err error
			key, err = resolveKey(ctx, ref.ID)
			if err != nil {
				return nil, fmt.Errorf("resolving key of product selection %q: %w", ref.ID, err)
			}
		}

		if _, useKey := usedKeys[key]; useKey && key != "" {
			item["product_selection_key"] = key
		} else {
			item["product_selection_id"] = ref.ID
		}
		result[i] = item
	}

	return result, nil
}

func isUsedByID(usedIDs map[string]struct{}, id string) bool {
	_, ok := usedIDs[id]
	return ok
}

// validateProductSelectionsConfig checks at plan time that every
// product_selection block sets exactly one of product_selection_id and
// product_selection_key. Unknown values (for example references to other
// resources) are not treated as unset; those blocks are only rejected when
// they are certainly invalid, the apply-time check in expandProductSelections
// remains as a backstop.
func validateProductSelectionsConfig(config cty.Value) error {
	if config.IsNull() || !config.IsKnown() || !config.Type().IsObjectType() ||
		!config.Type().HasAttribute("product_selection") {
		return nil
	}
	blocks := config.GetAttr("product_selection")
	if blocks.IsNull() || !blocks.IsKnown() || !blocks.CanIterateElements() {
		return nil
	}

	for it := blocks.ElementIterator(); it.Next(); {
		_, block := it.Element()
		if block.IsNull() || !block.IsKnown() || !block.Type().IsObjectType() {
			continue
		}

		set, unknown := 0, 0
		for _, name := range []string{"product_selection_id", "product_selection_key"} {
			if !block.Type().HasAttribute(name) {
				continue
			}
			v := block.GetAttr(name)
			switch {
			case !v.IsKnown():
				unknown++
			case !v.IsNull() && v.Type() == cty.String && v.AsString() != "":
				set++
			}
		}

		if set > 1 || (set == 0 && unknown == 0) {
			return errors.New("product_selection: exactly one of product_selection_id or product_selection_key must be set")
		}
	}
	return nil
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
		if reflect.DeepEqual(item, needle) {
			return true
		}
	}
	return false
}
