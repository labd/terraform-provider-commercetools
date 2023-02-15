package customtypes

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"
)

func TestLocalizedString(t *testing.T) {
	val := NewLocalizedStringValue(map[string]attr.Value{
		"nl": types.StringValue("foobar"),
	})

	result := val.ValueLocalizedString()
	expected := platform.LocalizedString{
		"nl": "foobar",
	}
	assert.Equal(t, expected, result)
}

func TestLocalizedStringUnknown(t *testing.T) {
	val := NewLocalizedStringNull()

	result := val.ValueLocalizedString()
	expected := platform.LocalizedString(nil)
	assert.Equal(t, expected, result)
}
