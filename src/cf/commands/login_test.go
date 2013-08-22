package commands_test

import (
	"cf"
	. "cf/commands"
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

func TestSuccessfullyLoggingIn(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(successfulLoginEndpoint))
	defer ts.Close()

	config := logout(t, ts.URL)

	ui := new(testhelpers.FakeUI)
	ui.Inputs = []string{"foo@example.com", "bar"}

	Login(nil, ui, &testhelpers.FakeOrgRepository{})

	assert.Contains(t, ui.Outputs[0], config.Target)
	assert.Contains(t, ui.Outputs[2], "OK")
	assert.Contains(t, ui.Prompts[0], "Email")
	assert.Contains(t, ui.Prompts[1], "Password")

	config, err := configuration.Load()
	assert.NoError(t, err)
	assert.Equal(t, config.AccessToken, "BEARER my_access_token")
}

func TestLoggingInWithTwoOrgsAskUserToChooseOrg(t *testing.T) {
	loginServer := httptest.NewTLSServer(http.HandlerFunc(successfulLoginEndpoint))
	defer loginServer.Close()

	config := logout(t, loginServer.URL)

	ui := new(testhelpers.FakeUI)
	ui.Inputs = []string{"foo@example.com", "bar", "2"}

	orgs := []cf.Organization{
		cf.Organization{"FirstOrg", "org-1-guid"},
		cf.Organization{"SecondOrg", "org-2-guid"},
	}
	Login(nil, ui, &testhelpers.FakeOrgRepository{Organizations: orgs})

	assert.Contains(t, ui.Outputs[0], config.Target)

	assert.Contains(t, ui.Prompts[0], "Email")
	assert.Contains(t, ui.Prompts[1], "Password")
	assert.Contains(t, ui.Outputs[2], "OK")

	assert.Contains(t, ui.Outputs[3], "FirstOrg")
	assert.Contains(t, ui.Outputs[4], "SecondOrg")

	assert.Contains(t, ui.Prompts[2], "Organization")
	assert.Contains(t, ui.Outputs[5], "SecondOrg")

	config, err := configuration.Load()
	assert.NoError(t, err)
	assert.Equal(t, orgs[1], config.Organization)
}

func TestWhenUserPicksInvalidOrgNumber(t *testing.T) {
	loginServer := httptest.NewTLSServer(http.HandlerFunc(successfulLoginEndpoint))
	defer loginServer.Close()

	config := logout(t, loginServer.URL)

	orgs := []cf.Organization{
		cf.Organization{"Org1", "org-1-guid"},
		cf.Organization{"Org2", "org-2-guid"},
	}

	ui := new(testhelpers.FakeUI)
	ui.Inputs = []string{"foo@example.com", "bar", "3", "2"}

	Login(nil, ui, &testhelpers.FakeOrgRepository{Organizations: orgs})

	assert.Contains(t, ui.Prompts[2], "Organization")
	assert.Contains(t, ui.Outputs[5], "FAILED")
	assert.Contains(t, ui.Prompts[3], "Organization")
	assert.Contains(t, ui.Outputs[9], "Targeting org")

	config, err := configuration.Load()
	assert.NoError(t, err)
	assert.Equal(t, orgs[1], config.Organization)
}

var unsuccessfulLoginEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusBadRequest)
}

func TestUnsuccessfullyLoggingIn(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(unsuccessfulLoginEndpoint))
	defer ts.Close()

	config := logout(t, ts.URL)

	ui := new(testhelpers.FakeUI)
	ui.Inputs = []string{
		"foo@example.com",
		"bar",
		"bar",
		"bar",
		"bar",
	}

	Login(nil, ui, &testhelpers.FakeOrgRepository{})

	assert.Contains(t, ui.Outputs[0], config.Target)
	assert.Equal(t, ui.Outputs[1], "Authenticating...")
	assert.Equal(t, ui.Outputs[2], "FAILED")
	assert.Equal(t, ui.Outputs[5], "Authenticating...")
	assert.Equal(t, ui.Outputs[6], "FAILED")
	assert.Equal(t, ui.Outputs[9], "Authenticating...")
	assert.Equal(t, ui.Outputs[10], "FAILED")

	config, err := configuration.Load()
	assert.NoError(t, err)
	assert.Equal(t, config.AccessToken, "")
}

func logout(t *testing.T, url string) (config *configuration.Configuration) {
	configuration.Delete()
	config, err := configuration.Load()
	assert.NoError(t, err)
	config.AuthorizationEndpoint = url
	config.AccessToken = ""
	err = config.Save()
	assert.NoError(t, err)
	return
}
