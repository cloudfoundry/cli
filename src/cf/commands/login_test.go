package commands

import (
	"cf/configuration"
	"encoding/base64"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testhelpers"
	"testing"
)

var successfulLoginEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	contentTypeMatches := request.Header.Get("content-type") == "application/x-www-form-urlencoded"
	acceptHeaderMatches := request.Header.Get("accept") == "application/json"
	methodMatches := request.Method == "POST"
	pathMatches := request.URL.Path == "/oauth/token"
	encodedAuth := base64.StdEncoding.EncodeToString([]byte("cf:"))
	basicAuthMatches := request.Header.Get("authorization") == "Basic "+encodedAuth

	err := request.ParseForm()

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	usernameMatches := request.Form.Get("username") == "foo@example.com"
	passwordMatches := request.Form.Get("password") == "bar"
	grantTypeMatches := request.Form.Get("grant_type") == "password"
	scopeMatches := request.Form.Get("scope") == ""
	bodyMatches := usernameMatches && passwordMatches && grantTypeMatches && scopeMatches

	if !(contentTypeMatches && acceptHeaderMatches && methodMatches && pathMatches && bodyMatches && basicAuthMatches) {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonResponse := `
{
  "access_token": "my_access_token",
  "token_type": "BEARER",
  "scope": "openid",
  "expires_in": 98765
} `
	fmt.Fprintln(writer, jsonResponse)
}

var unsuccessfulLoginEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusBadRequest)
}

func TestSuccessfullyLoggingIn(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(successfulLoginEndpoint))
	defer ts.Close()

	config, err := configuration.Load()
	assert.NoError(t, err)
	config.AuthorizationEndpoint = ts.URL
	err = config.Save()
	assert.NoError(t, err)

	ui := new(testhelpers.FakeUI)
	ui.Inputs = []string{"foo@example.com", "bar"}

	Login(nil, ui)

	assert.Contains(t, ui.Outputs[0], config.Target)
	assert.Contains(t, ui.Outputs[2], "OK")
	assert.Equal(t, ui.Prompts[0], "Email>")
	assert.Equal(t, ui.Prompts[1], "Password>")
}

func TestUnsuccessfullyLoggingIn(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(unsuccessfulLoginEndpoint))
	defer ts.Close()

	config, err := configuration.Load()
	assert.NoError(t, err)
	config.AuthorizationEndpoint = ts.URL
	err = config.Save()
	assert.NoError(t, err)

	ui := new(testhelpers.FakeUI)
	ui.Inputs = []string{"foo@example.com", "bar"}

	Login(nil, ui)

	assert.Contains(t, ui.Outputs[0], config.Target)
	assert.Contains(t, ui.Outputs[2], "FAILED")
	assert.Equal(t, ui.Prompts[0], "Email>")
	assert.Equal(t, ui.Prompts[1], "Password>")
}
