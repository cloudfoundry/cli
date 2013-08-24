package api

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewAuthorizedRequest(t *testing.T) {
	request, err := NewAuthorizedRequest("GET", "https://example.com/v2/apps", "BEARER my-access-token", nil)

	assert.NoError(t, err)
	assert.Equal(t, request.Header.Get("Authorization"), "BEARER my-access-token")
	assert.Equal(t, request.Header.Get("accept"), "application/json")
}
