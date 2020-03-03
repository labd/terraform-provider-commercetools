package commercetools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateLookup(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"name":  "name1",
			"value": "Value 1",
		},
		map[string]interface{}{
			"name":  "name2",
			"value": "Value 2",
		},
	}
	result := createLookup(input, "name")
	if _, ok := result["name1"]; !ok {
		t.Error("Could not lookup name1")
	}
	if _, ok := result["name2"]; !ok {
		t.Error("Could not lookup name1")
	}
}

func TestSlugify(t *testing.T) {
	assert.Equal(t, slugify("foo-bar"), "foo-bar")
	assert.Equal(t, slugify("foo_bar"), "foo-bar")
	assert.Equal(t, slugify("foo/bar"), "foo-bar")
}
