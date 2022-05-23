package commercetools

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/labd/commercetools-go-sdk/platform"
)

func processRemoteError(err error) diag.Diagnostics {
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case platform.GenericRequestError:
		{
			data := map[string]any{}
			if err := json.Unmarshal(e.Content, &data); err == nil {
				if val, ok := data["message"].(string); ok {
					return diag.Errorf(val)
				}
			}
			return diag.FromErr(e)
		}
	default:
		return diag.FromErr(err)
	}
}
