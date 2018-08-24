package commercetools

import (
	"testing"

	"github.com/labd/commercetools-go-sdk/service/extensions"
	"github.com/stretchr/testify/assert"
)

func TestExtensionCreateAuthentication(t *testing.T) {
	var input map[string]interface{}
	input = map[string]interface{}{
		"authorization_header": "12345",
		"azure_authentication": "AzureKey",
	}

	auth, err := resourceExtensionCreateAuthentication(input)
	assert.Nil(t, auth)
	assert.NotNil(t, err)

	input = map[string]interface{}{
		"authorization_header": "12345",
	}

	auth, err = resourceExtensionCreateAuthentication(input)
	httpAuth := auth.(*extensions.DestinationAuthenticationAuth)
	assert.Equal(t, "AuthorizationHeader", httpAuth.Type())
	assert.Equal(t, "12345", httpAuth.HeaderValue)
	assert.NotNil(t, auth)
	assert.Nil(t, err)
}
