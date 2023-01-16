package acctest

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-commercetools/internal/provider"
)

func TestProvider(t *testing.T) {
	p := provider.New("version")

	assert.NotNil(t, p)

}
