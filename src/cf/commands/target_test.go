package commands

import (
	"testing"
	"github.com/codegangsta/cli"
	"github.com/stretchr/testify/assert"
	"flag"
	"testhelpers"
	"net/http"
	"fmt"
	"net/http/httptest"
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
  "authorization_endpoint": "https://login.run.pivotal.io",
  "api_version": "2.0.0"
} `
	fmt.Fprintln(w, infoResponse)
}

var invalidInfoEndpoint = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	return
}

func TestTargetWithoutArgument(t *testing.T) {
	flagSet := new(flag.FlagSet)
	globalSet := new(flag.FlagSet)
	context := cli.NewContext(cli.NewApp(), flagSet, globalSet)

	out := testhelpers.CaptureOutput(func() {
		Target(context)
	})

	assert.Contains(t, out, "https://api.run.pivotal.io")
}

func TestTargetWhenUrlIsValidInfoEndpoint(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(validInfoEndpoint))
	defer ts.Close()

	args := []string{ts.URL}
	flagSet := new(flag.FlagSet)
	flagSet.Parse(args)

	globalSet := new(flag.FlagSet)
	context := cli.NewContext(cli.NewApp(), flagSet, globalSet)

	out := testhelpers.CaptureOutput(func() {
		Target(context)
	})

	assert.Contains(t, out, ts.URL)
}

func TestTargetWhenUrlIsInValidInfoEndpoint(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(invalidInfoEndpoint))
	defer ts.Close()
	args := []string{ts.URL}
	flagSet := new(flag.FlagSet)
	flagSet.Parse(args)

	globalSet := new(flag.FlagSet)
	context := cli.NewContext(cli.NewApp(), flagSet, globalSet)

	out := testhelpers.CaptureOutput(func() {
		Target(context)
	})

	assert.Contains(t, out, ts.URL)
	assert.Contains(t, out, "FAILED")
	assert.Contains(t, out, "Target refused connection.")
}

