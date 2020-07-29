package commercetools

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

func resourceStore() *schema.Resource {
	return &schema.Resource{
		Create: resourceStoreCreate,
		Read:   resourceStoreRead,
		Update: resourceStoreUpdate,
		Delete: resourceStoreDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     TypeLocalizedString,
				Optional: true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"languages": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"distribution_channels": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceStoreCreate(d *schema.ResourceData, m interface{}) error {
	name := commercetools.LocalizedString(
		expandStringMap(d.Get("name").(map[string]interface{})))
	dcIdentifiers := expandDistributionChannels(d.Get("distribution_channels"))

	draft := &commercetools.StoreDraft{
		Key:                  d.Get("key").(string),
		Name:                 &name,
		Languages:            expandStringArray(d.Get("languages").([]interface{})),
		DistributionChannels: dcIdentifiers,
	}

	client := getClient(m)

	var store *commercetools.Store

	err := resource.Retry(20*time.Second, func() *resource.RetryError {
		var err error
		store, err = client.StoreCreate(context.Background(), draft)

		if err != nil {
			return handleCommercetoolsError(err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	d.SetId(store.ID)
	d.Set("version", store.Version)
	return resourceStoreRead(d, m)
}

func resourceStoreRead(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)

	store, err := client.StoreGetWithID(context.Background(), d.Id(), commercetools.WithReferenceExpansion("distributionChannels[*]"))

	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.SetId(store.ID)
	d.Set("key", store.Key)
	d.Set("name", *store.Name)
	d.Set("version", store.Version)
	if store.Languages != nil {
		d.Set("languages", store.Languages)
	}

	log.Printf("[DEBUG] Store read, distributionChannels: %+v", store.DistributionChannels)
	if store.DistributionChannels != nil {
		channelKeys, err := flattenDistributionChannels(store.DistributionChannels)
		if err != nil {
			return err
		}
		log.Printf("[DEBUG] Setting channel keys to: %+v", channelKeys)
		d.Set("distribution_channels", channelKeys)
	}
	return nil
}

func resourceStoreUpdate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)

	input := &commercetools.StoreUpdateWithIDInput{
		ID:      d.Id(),
		Version: d.Get("version").(int),
		Actions: []commercetools.StoreUpdateAction{},
	}

	if d.HasChange("name") {
		newName := commercetools.LocalizedString(
			expandStringMap(d.Get("name").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&commercetools.StoreSetNameAction{Name: &newName})
	}

	if d.HasChange("languages") {
		languages := expandStringArray(d.Get("languages").([]interface{}))

		input.Actions = append(
			input.Actions,
			&commercetools.StoreSetLanguagesAction{Languages: languages})
	}

	if d.HasChange("distribution_channels") {
		dcIdentifiers := expandDistributionChannels(d.Get("distribution_channels"))

		log.Printf("[DEBUG] distributionChannels change, new identifiers: %v", dcIdentifiers)

		// set action replaces current values
		input.Actions = append(
			input.Actions,
			&commercetools.StoresSetDistributionChannelsAction{
				DistributionChannels: dcIdentifiers,
			},
		)
	}

	_, err := client.StoreUpdateWithID(context.Background(), input)
	if err != nil {
		return err
	}

	return resourceStoreRead(d, m)
}

func resourceStoreDelete(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	version := d.Get("version").(int)
	_, err := client.StoreDeleteWithID(context.Background(), d.Id(), version)
	if err != nil {
		return err
	}

	return nil
}

func convertChannelKeysToIdentifiers(channelKeys []string) []commercetools.ChannelResourceIdentifier {
	identifiers := make([]commercetools.ChannelResourceIdentifier, 0)
	for i := 0; i < len(channelKeys); i++ {
		channelIdentifier := commercetools.ChannelResourceIdentifier{
			Key: channelKeys[i],
		}
		identifiers = append(identifiers, channelIdentifier)
	}

	log.Printf("[DEBUG] Converted keys: %v", identifiers)
	return identifiers
}

func expandDistributionChannels(distributionChannelData interface{}) []commercetools.ChannelResourceIdentifier {
	log.Printf("[DEBUG] Expanding distribution channels: %v", distributionChannelData)
	distributionChannelKeys := expandStringArray(distributionChannelData.([]interface{}))
	log.Printf("[DEBUG] Expanding distribution channels, got keys: %v", distributionChannelKeys)
	return convertChannelKeysToIdentifiers(distributionChannelKeys)
}

func flattenDistributionChannels(distributionChannels []commercetools.ChannelReference) ([]string, error) {
	log.Printf("[DEBUG] flattening: %+v", distributionChannels)
	channelKeys := make([]string, 0)
	for i := 0; i < len(distributionChannels); i++ {

		log.Printf("[DEBUG] flattening checking channel: %s", stringFormatObject(distributionChannels[i]))
		log.Printf("[DEBUG] flattening checking channel obj: %s", stringFormatObject(distributionChannels[i].Obj))
		if distributionChannels[i].Obj == nil {
			return nil, errors.New("failed to expand channel objects")
		}
		channelKeys = append(channelKeys, distributionChannels[i].Obj.Key)
	}
	log.Printf("[DEBUG] flattening final keys: %v", channelKeys)
	return channelKeys, nil
}
