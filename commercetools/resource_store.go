package commercetools

import (
	"context"
	"errors"
	"log"
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
				Description: "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
				Type:        TypeLocalizedString,
				Optional:    true,
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
		},
	}
}

func resourceStoreCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	name := unmarshallLocalizedString(d.Get("name"))
	dcIdentifiers := expandStoreChannels(d.Get("distribution_channels"))
	scIdentifiers := expandStoreChannels(d.Get("supply_channels"))

	draft := platform.StoreDraft{
		Key:                  d.Get("key").(string),
		Name:                 &name,
		Languages:            expandStringArray(d.Get("languages").([]interface{})),
		DistributionChannels: dcIdentifiers,
		SupplyChannels:       scIdentifiers,
	}

	client := getClient(m)

	var store *platform.Store

	err := resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		var err error
		store, err = client.Stores().Post(draft).Execute(ctx)

		if err != nil {
			return handleCommercetoolsError(err)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(store.ID)
	d.Set("version", store.Version)
	return resourceStoreRead(ctx, d, m)
}

func resourceStoreRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)

	store, err := client.Stores().
		WithId(d.Id()).
		Get().
		Expand([]string{"distributionChannels[*]", "supplyChannels[*]"}).
		Execute(ctx)

	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
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

	log.Printf("[DEBUG] Store read, distributionChannels: %+v", store.DistributionChannels)
	if store.DistributionChannels != nil {
		channelKeys, err := flattenStoreChannels(store.DistributionChannels)
		if err != nil {
			return diag.FromErr(err)
		}
		log.Printf("[DEBUG] Setting channel keys to: %+v", channelKeys)
		d.Set("distribution_channels", channelKeys)
	}
	log.Printf("[DEBUG] Store read, supplyChannels: %+v", store.SupplyChannels)

	if store.SupplyChannels != nil {
		channelKeys, err := flattenStoreChannels(store.SupplyChannels)
		if err != nil {
			return diag.FromErr(err)
		}
		log.Printf("[DEBUG] Setting channel keys to: %+v", channelKeys)
		d.Set("supply_channels", channelKeys)
	}
	return nil
}

func resourceStoreUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)

	input := platform.StoreUpdate{
		Version: d.Get("version").(int),
		Actions: []platform.StoreUpdateAction{},
	}

	if d.HasChange("name") {
		newName := unmarshallLocalizedString(d.Get("name"))
		input.Actions = append(
			input.Actions,
			&platform.StoreSetNameAction{Name: &newName})
	}

	if d.HasChange("languages") {
		languages := expandStringArray(d.Get("languages").([]interface{}))

		input.Actions = append(
			input.Actions,
			&platform.StoreSetLanguagesAction{Languages: languages})
	}

	if d.HasChange("distribution_channels") {
		dcIdentifiers := expandStoreChannels(d.Get("distribution_channels"))

		log.Printf("[DEBUG] distributionChannels change, new identifiers: %v", dcIdentifiers)

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

		log.Printf("[DEBUG] supplyChannels change, new identifiers: %v", scIdentifiers)

		// set action replaces current values
		input.Actions = append(
			input.Actions,
			&platform.StoreSetSupplyChannelsAction{
				SupplyChannels: scIdentifiers,
			},
		)
	}

	_, err := client.Stores().WithId(d.Id()).Post(input).Execute(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceStoreRead(ctx, d, m)
}

func resourceStoreDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)
	version := d.Get("version").(int)

	_, err := client.Stores().WithId(d.Id()).Delete().Version(version).Execute(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func convertChannelKeysToIdentifiers(channelKeys []string) []platform.ChannelResourceIdentifier {
	identifiers := make([]platform.ChannelResourceIdentifier, 0)
	for i := 0; i < len(channelKeys); i++ {
		channelIdentifier := platform.ChannelResourceIdentifier{
			Key: &channelKeys[i],
		}
		identifiers = append(identifiers, channelIdentifier)
	}

	log.Printf("[DEBUG] Converted keys: %v", identifiers)
	return identifiers
}

func expandStoreChannels(channelData interface{}) []platform.ChannelResourceIdentifier {
	log.Printf("[DEBUG] Expanding store channels: %v", channelData)
	channelKeys := expandStringArray(channelData.([]interface{}))
	log.Printf("[DEBUG] Expanding store channels, got keys: %v", channelKeys)
	return convertChannelKeysToIdentifiers(channelKeys)
}

func flattenStoreChannels(channels []platform.ChannelReference) ([]string, error) {
	log.Printf("[DEBUG] flattening: %+v", channels)
	channelKeys := make([]string, 0)
	for i := 0; i < len(channels); i++ {

		log.Printf("[DEBUG] flattening checking channel: %s", stringFormatObject(channels[i]))
		log.Printf("[DEBUG] flattening checking channel obj: %s", stringFormatObject(channels[i].Obj))
		if channels[i].Obj == nil {
			return nil, errors.New("failed to expand channel objects")
		}
		channelKeys = append(channelKeys, channels[i].Obj.Key)
	}
	log.Printf("[DEBUG] flattening final keys: %v", channelKeys)
	return channelKeys, nil
}
