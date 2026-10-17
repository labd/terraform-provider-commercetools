package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegionalURL(t *testing.T) {
	assert.Equal(t,
		"https://extensions.integration-layer.europe-west1.gcp.commercetools.com",
		regionalURL("https://api.europe-west1.gcp.commercetools.com", "extensions.integration-layer"))
	assert.Equal(t,
		"https://connect.us-central1.gcp.commercetools.com",
		regionalURL("https://api.us-central1.gcp.commercetools.com/", "connect"))
	// A host that is not a commercetools region, such as the mock server's.
	assert.Equal(t, "", regionalURL("http://localhost:8989", "connect"))
}
