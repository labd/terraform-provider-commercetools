package commercetools

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
)

func flattenStores(storeKeyReferences []platform.StoreKeyReference) []string {
	var storeKeys []string
	for _, store := range storeKeyReferences {
		storeKeys = append(storeKeys, store.Key)
	}

	return storeKeys
}

func expandStores(storeKeys *schema.Set) []platform.StoreResourceIdentifier {
	storeResourceIdentifiers := make([]platform.StoreResourceIdentifier, 0, storeKeys.Len())
	for _, key := range storeKeys.List() {
		var keyVal = key.(string)
		storeResourceIdentifiers = append(storeResourceIdentifiers, platform.StoreResourceIdentifier{Key: &keyVal})
	}
	return storeResourceIdentifiers
}
