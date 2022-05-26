package commercetools

import (
	"encoding/json"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/labd/commercetools-go-sdk/platform"
)

func processRemoteError(err error) *resource.RetryError {
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case platform.ErrorResponse:
		return resource.NonRetryableError(e)

	case platform.GenericRequestError:
		{
			data := map[string]any{}
			if err := json.Unmarshal(e.Content, &data); err == nil {
				if val, ok := data["message"].(string); ok {
					return resource.NonRetryableError(errors.New(val))
				}
			}
			return resource.NonRetryableError(e)
		}
	}

	return resource.RetryableError(err)
}
