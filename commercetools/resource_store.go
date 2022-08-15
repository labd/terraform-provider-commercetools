package commercetools

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
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
			"custom": CustomFieldSchema(),
		},
	}
}

func resourceStoreCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	name := expandLocalizedString(d.Get("name"))
	dcIdentifiers := expandStoreChannels(d.Get("distribution_channels"))
	scIdentifiers := expandStoreChannels(d.Get("supply_channels"))

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
		DistributionChannels: dcIdentifiers,
		SupplyChannels:       scIdentifiers,
		Custom:               custom,
	}

	var store *platform.Store
	err = resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		var err error
		store, err = client.Stores().Post(draft).Execute(ctx)
		return processRemoteError(err)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(store.ID)
	d.Set("version", store.Version)
	return resourceStoreRead(ctx, d, m)
}

func resourceStoreRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	store, err := client.Stores().
		WithId(d.Id()).
		Get().
		Expand([]string{"distributionChannels[*]", "supplyChannels[*]"}).
		Execute(ctx)

	if err != nil {
		if IsResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(store.ID)
	d.Set("key", store.Key)
	if store.Name != nil {
		d.Set("name", *store.Name)
	} else {
		d.Set("name", nil)
	}
	d.Set("version", store.Version)
	if store.Languages != nil {
		d.Set("languages", store.Languages)
	}

	if store.DistributionChannels != nil {
		channelKeys, err := flattenStoreChannels(store.DistributionChannels)
		if err != nil {
			return diag.FromErr(err)
		}
		d.Set("distribution_channels", channelKeys)
	}

	if store.SupplyChannels != nil {
		channelKeys, err := flattenStoreChannels(store.SupplyChannels)
		if err != nil {
			return diag.FromErr(err)
		}
		d.Set("supply_channels", channelKeys)
	}
	d.Set("custom", flattenCustomFields(store.Custom))
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
		return processRemoteError(err)
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
		return processRemoteError(err)
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
