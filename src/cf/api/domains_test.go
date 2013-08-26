package api

import (
	"cf"
	"cf/configuration"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var multipleDomainsEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	acceptHeaderMatches := request.Header.Get("accept") == "application/json"
	methodMatches := request.Method == "GET"
	pathMatches := request.URL.Path == "/v2/spaces/my-space-guid/domains"
	authMatches := request.Header.Get("authorization") == "BEARER my_access_token"

	if !(acceptHeaderMatches && methodMatches && pathMatches && authMatches) {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonResponse := `
{
  "total_results": 2,
  "total_pages": 1,
  "prev_url": null,
  "next_url": null,
  "resources": [
    {
      "metadata": {
        "guid": "domain1-guid"
      },
      "entity": {
        "name": "domain1.cf-app.com"
      }
    },
    {
      "metadata": {
        "guid": "domain2-guid"
      },
      "entity": {
        "name": "domain2.cf-app.com"
      }
    }
  ]
}`
	fmt.Fprintln(writer, jsonResponse)
}

func TestFindAll(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleDomainsEndpoint))
	defer ts.Close()

	repo := CloudControllerDomainRepository{}

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
	}

	domains, err := repo.FindAll(config)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(domains))

	first := domains[0]
	assert.Equal(t, first.Name, "domain1.cf-app.com")
	assert.Equal(t, first.Guid, "domain1-guid")
	second := domains[1]
	assert.Equal(t, second.Name, "domain2.cf-app.com")
	assert.Equal(t, second.Guid, "domain2-guid")
}
