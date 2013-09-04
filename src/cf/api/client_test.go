package api_test

import (
	. "cf/api"
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

	_, err = PerformRequest(request)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "The host is taken: test1")
}

func TestPerformRequestForBodyOutputsErrorFromServer(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(failingRequest))
	defer ts.Close()

	request, err := NewAuthorizedRequest("GET", ts.URL, "TOKEN", nil)
	assert.NoError(t, err)

	resource := new(Resource)
	_, err = PerformRequestAndParseResponse(request, resource)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "The host is taken: test1")
}

func TestPerformRequestReturnsErrorCode(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(failingRequest))
	defer ts.Close()

	request, err := NewAuthorizedRequest("GET", ts.URL, "TOKEN", nil)
	assert.NoError(t, err)

	resource := new(Resource)
	errorCode, err := PerformRequestAndParseResponse(request, resource)

	assert.Equal(t, errorCode, 210003)
	assert.Error(t, err)
}

func TestSanitizingRequest(t *testing.T) {
	request := `
REQUEST:
GET /v2/organizations HTTP/1.1
Host: api.run.pivotal.io
Accept: application/json
Authorization: bearer eyJhbGciOiJSUzI1NiJ9.eyJqdGkiOiI3NDRkNWQ1My0xODkxLTQzZjktYjNiMy1mMTQxNDZkYzQ4ZmUiLCJzdWIiOiIzM2U3ZmVkNy1iMWMyLTRjMjAtOTU0My0yMTBiMjc2ODM1MDgiLCJzY29wZSI6WyJjbG91ZF9jb250cm9sbGVyLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwib3BlbmlkIiwicGFzc3dvcmQud3JpdGUiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImdyYW50X3R5cGUiOiJwYXNzd29yZCIsInVzZXJfaWQiOiIzM2U3ZmVkNy1iMWMyLTRjMjAtOTU0My0yMTBiMjc2ODM1MDgiLCJ1c2VyX25hbWUiOiJtZ2VoYXJkK2NsaUBwaXZvdGFsbGFicy5jb20iLCJlbWFpbCI6Im1nZWhhcmQrY2xpQHBpdm90YWxsYWJzLmNvbSIsImlhdCI6MTM3ODI0NzgxNiwiZXhwIjoxMzc4MjkxMDE2LCJpc3MiOiJodHRwczovL3VhYS5ydW4ucGl2b3RhbC5pby9vYXV0aC90b2tlbiIsImF1ZCI6WyJvcGVuaWQiLCJjbG91ZF9jb250cm9sbGVyIiwicGFzc3dvcmQiXX0.LL_QLO0SztGRENmU-9KA2WouOyPkKVENGQoUtjqrGR-UIekXMClH6fmKELzHtB69z3n9x7_jYJbvv32D-dX1J7p1CMWIDLOzXUnIUDK7cU5Q2yuYszf4v5anKiJtrKWU0_Pg87cQTZ_lWXAhdsi-bhLVR_pITxehfz7DKChjC8gh-FiuDvH5qHxxPqYHUl9jPso5OQ0y0fqZpLt8Yq23DKWaFAZehLnrhFltdQ_jSLy1QAYYZVD_HpQDf9NozKXruIvXhyIuwGj99QmUs3LSyNWecy822VqOoBtPYS6CLegMuWWlO64TJNrnZuh5YsOuW8SudJONx2wwEqARysJIHw
This is the body. Please don't get rid of me even though I contain Authorization: and some other text
	`

	expected := `
REQUEST:
GET /v2/organizations HTTP/1.1
Host: api.run.pivotal.io
Accept: application/json
Authorization: [PRIVATE DATA HIDDEN]
This is the body. Please don't get rid of me even though I contain Authorization: and some other text
	`

	assert.Equal(t, SanitizeRequest(request), expected)
}
