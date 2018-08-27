package commercetools

import (
	"testing"

	"github.com/labd/commercetools-go-sdk/service/extensions"
	"github.com/stretchr/testify/assert"
)

func TestAPIExtensionGetAuthentication(t *testing.T) {
	var input map[string]interface{}
	input = map[string]interface{}{
		"authorization_header": "12345",
		"azure_authentication": "AzureKey",
	}

	auth, err := resourceAPIExtensionGetAuthentication(input)
	assert.Nil(t, auth)
	assert.NotNil(t, err)

	input = map[string]interface{}{
		"authorization_header": "12345",
	}

	auth, err = resourceAPIExtensionGetAuthentication(input)
	httpAuth, ok := auth.(*extensions.DestinationAuthenticationAuth)
	assert.True(t, ok)
	assert.Equal(t, "12345", httpAuth.HeaderValue)
	assert.NotNil(t, auth)
	assert.Nil(t, err)
}
