package commercetools

// import (
// 	"fmt"
// 	"testing"

// 	"github.com/labd/commercetools-go-sdk/platform"
// )

// func TestCreateLookup(t *testing.T) {
// 	input := []interface{}{
// 		map[string]interface{}{
// 			"name":  "name1",
// 			"value": "Value 1",
// 		},
// 		map[string]interface{}{
// 			"name":  "name2",
// 			"value": "Value 2",
// 		},
// 	}
// 	result := createLookup(input, "name")
// 	if _, ok := result["name1"]; !ok {
// 		t.Error("Could not lookup name1")
// 	}
// 	if _, ok := result["name2"]; !ok {
// 		t.Error("Could not lookup name1")
// 	}
// }

// func checkApiResult(err error) error {
// 	switch v := err.(type) {
// 	case platform.GenericRequestError:
// 		if v.StatusCode == 404 {
// 			return nil
// 		}
// 		return fmt.Errorf("unhandled error generic error returned (%d)", v.StatusCode)
// 	case platform.ResourceNotFoundError:
// 		return nil
// 	default:
// 		return fmt.Errorf("unexpected result returned")
// 	}
// }
