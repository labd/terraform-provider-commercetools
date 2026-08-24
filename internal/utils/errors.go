package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/labd/commercetools-go-sdk/insights"
	"github.com/labd/commercetools-go-sdk/platform"
)

func ProcessRemoteError(err error) *resource.RetryError {
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case platform.ErrorResponse:
		if err := extractDetailedError(e); err != nil {
			return resource.NonRetryableError(err)
		}
		return resource.NonRetryableError(e)

	case platform.GenericRequestError:
		{
			if err := extractRawDetailedError(e.Content); err != nil {
				return resource.NonRetryableError(err)
			}
			return resource.NonRetryableError(e)
		}

	case insights.ErrorResponse:
		return resource.NonRetryableError(extractDetailedErrorFrom(e.Errors, e))

	case insights.GenericRequestError:
		{
			if err := extractRawDetailedError(e.Content); err != nil {
				return resource.NonRetryableError(err)
			}
			return resource.NonRetryableError(e)
		}
	}

	return resource.RetryableError(err)
}

func extractDetailedError(e platform.ErrorResponse) error {
	return extractDetailedErrorFrom(e.Errors, e)
}

// extractDetailedErrorFrom digs the detailedErrorMessage out of a list of error
// objects. The error objects of the various commercetools APIs are distinct
// (generated) types that share the same shape, so they are inspected
// reflectively. Returns fallback when no detailed message is present.
func extractDetailedErrorFrom[T any](errs []T, fallback error) error {
	for i := range errs {
		metaValue := reflect.ValueOf(errs[i])
		message := metaValue.FieldByName("Message")
		values := metaValue.FieldByName("ExtraValues")

		if message.IsValid() && values.IsValid() {
			data, ok := values.Interface().(map[string]any)
			if !ok {
				continue
			}
			if msg, ok := data["detailedErrorMessage"]; ok {
				return fmt.Errorf("%s %s", message.String(), msg)
			}
		}
	}
	return fallback
}

func extractRawDetailedError(content []byte) error {
	data := map[string]any{}
	if err := json.Unmarshal(content, &data); err != nil {
		return nil
	}

	// Iterate over the errors. This is a list of objects containing the
	// code, message and detailedErrorMessage values.
	if val, ok := data["errors"].([]any); ok {
		for i := range val {
			if error, ok := val[i].(map[string]any); ok {
				var message string

				if detail, ok := error["message"].(string); ok {
					message = detail
				}

				if detail, ok := error["detailedErrorMessage"].(string); ok {
					if message != "" {
						return fmt.Errorf("%s %s", message, detail)
					}
					return errors.New(detail)
				}
			}
		}
	}

	// Fallback to the generic error message
	if val, ok := data["message"].(string); ok {
		return errors.New(val)
	}

	return nil
}

// IsResourceNotFoundError returns true if commercetools returned a 404 error
func IsResourceNotFoundError(err error) bool {
	//Occasionally the SDK returns a sentinel value instead of the parsed error response for 404.
	//This is a workaround to handle that case.
	if errors.Is(err, platform.ErrNotFound) || errors.Is(err, insights.ErrNotFound) {
		return true
	}

	switch e := err.(type) {
	case platform.ResourceNotFoundError:
		return true

	case platform.ErrorResponse:
		return e.StatusCode == 404

	case platform.GenericRequestError:
		return e.StatusCode == 404

	case insights.ErrorResponse:
		return e.StatusCode == 404

	case insights.GenericRequestError:
		return e.StatusCode == 404
	}
	return false
}
