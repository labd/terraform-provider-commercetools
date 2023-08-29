package utils

import (
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsResourceNotFoundError(t *testing.T) {
	var cases = []struct {
		err      error
		expected bool
	}{
		{platform.ErrNotFound, true},
		{platform.ResourceNotFoundError{}, true},
		{platform.ErrorResponse{StatusCode: 404}, true},
		{platform.ErrorResponse{StatusCode: 500}, false},
		{platform.GenericRequestError{StatusCode: 404}, true},
		{platform.GenericRequestError{StatusCode: 500}, false},
	}

	for _, tt := range cases {
		assert.Equal(t, tt.expected, IsResourceNotFoundError(tt.err))
	}
}
