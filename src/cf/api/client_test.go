package api

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewAuthorizedRequest(t *testing.T) {
	request, err := NewAuthorizedRequest("GET", "https://example.com/v2/apps", "BEARER my-access-token", nil)

	assert.NoError(t, err)
	assert.Equal(t, request.Header.Get("Authorization"), "BEARER my-access-token")
	assert.Equal(t, request.Header.Get("accept"), "application/json")
}

var failingRequest = func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusBadRequest)
	jsonResponse := `
	{
	  "code": 210003,
	  "description": "The host is taken: test1"
	}`
	fmt.Fprintln(writer, jsonResponse)
}

func TestPerformRequestOutputsErrorFromServer(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(failingRequest))
	defer ts.Close()

	request, err := NewAuthorizedRequest("GET", ts.URL, "TOKEN", nil)
	assert.NoError(t, err)

	err = PerformRequest(request)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "The host is taken: test1")
}

func TestPerformRequestForBodyOutputsErrorFromServer(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(failingRequest))
	defer ts.Close()

	request, err := NewAuthorizedRequest("GET", ts.URL, "TOKEN", nil)
	assert.NoError(t, err)

	resource := new(Resource)
	err = PerformRequestForBody(request, resource)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "The host is taken: test1")
}
