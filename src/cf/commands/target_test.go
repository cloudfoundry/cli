package commands

import (
	"cf/configuration"
	term "cf/terminal"
	"flag"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testhelpers"
	"testing"
)

var validInfoEndpoint = func(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/v2/info" {
		w.WriteHeader(http.StatusNotFound)
		return

	}

	infoResponse := `
{
  "name": "vcap",
  "build": "2222",
  "support": "http://support.cloudfoundry.com",
  "version": 2,
  "description": "Cloud Foundry sponsored by Pivotal",
  "authorization_endpoint": "https://login.example.com",
  "api_version": "42.0.0"
} `
	fmt.Fprintln(w, infoResponse)
}

var notFoundEndpoint = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	return
}

var invalidJsonResponseEndpoint = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `I am not valid`)
}

func newContext(args []string) *cli.Context {
	flagSet := new(flag.FlagSet)
	flagSet.Parse(args)
	globalSet := new(flag.FlagSet)

	return cli.NewContext(cli.NewApp(), flagSet, globalSet)
}

func TestTargetDefaults(t *testing.T) {
	configuration.Delete()
	context := newContext([]string{})

	out := testhelpers.CaptureOutput(func() {
		Target(context, new(term.TerminalUI))
	})

	assert.Contains(t, out, "https://api.run.pivotal.io")
}

func TestTargetWhenUrlIsValidInfoEndpoint(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(validInfoEndpoint))
	defer ts.Close()

	URL, err := url.Parse(ts.URL)
	assert.NoError(t, err)

	context := newContext([]string{URL.Host})
	out := testhelpers.CaptureOutput(func() {
		Target(context, new(term.TerminalUI))
	})

	assert.Contains(t, out, "https://"+URL.Host)
	assert.Contains(t, out, "42.0.0")

	context = newContext([]string{})
	out = testhelpers.CaptureOutput(func() {
		Target(context, new(term.TerminalUI))
	})

	assert.Contains(t, out, "https://"+URL.Host)
	assert.Contains(t, out, "42.0.0")

	savedConfig, err := configuration.Load()

	assert.NoError(t, err)
	assert.Equal(t, savedConfig.AuthorizationEndpoint, "https://login.example.com")
}

func TestTargetWhenEndpointReturns404(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(notFoundEndpoint))
	defer ts.Close()

	URL, err := url.Parse(ts.URL)
	assert.NoError(t, err)

	context := newContext([]string{URL.Host})
	out := testhelpers.CaptureOutput(func() {
		Target(context, new(term.TerminalUI))
	})

	assert.Contains(t, out, "https://"+URL.Host)
	assert.Contains(t, out, "FAILED")
	assert.Contains(t, out, "Target refused connection.")
}

func TestTargetWhenEndpointReturnsInvalidJson(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(invalidJsonResponseEndpoint))
	defer ts.Close()

	URL, err := url.Parse(ts.URL)
	assert.NoError(t, err)

	context := newContext([]string{URL.Host})
	out := testhelpers.CaptureOutput(func() {
		Target(context, new(term.TerminalUI))
	})

	assert.Contains(t, out, "FAILED")
	assert.Contains(t, out, "Invalid JSON response from server.")
}
