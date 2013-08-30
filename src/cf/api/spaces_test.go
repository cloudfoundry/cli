package api_test

import (
	"cf"
	. "cf/api"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testhelpers"
	"testing"
)

var multipleSpacesResponse = testhelpers.TestResponse{Status: http.StatusOK, Body: `
{
  "resources": [
    {
      "metadata": {
        "guid": "acceptance-space-guid"
      },
      "entity": {
        "name": "acceptance"
      }
    },
    {
      "metadata": {
        "guid": "staging-space-guid"
      },
      "entity": {
        "name": "staging"
      }
    }
  ]
}`}

var multipleSpacesEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/organizations/some-org-guid/spaces",
	nil,
	multipleSpacesResponse,
)

func TestSpacesFindAll(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleSpacesEndpoint))
	defer ts.Close()

	repo := CloudControllerSpaceRepository{}
	config := &configuration.Configuration{
		AccessToken:  "BEARER my_access_token",
		Target:       ts.URL,
		Organization: cf.Organization{Guid: "some-org-guid"},
	}
	spaces, err := repo.FindAll(config)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(spaces))

	firstSpace := spaces[0]
	assert.Equal(t, firstSpace.Name, "acceptance")
	assert.Equal(t, firstSpace.Guid, "acceptance-space-guid")

	secondSpace := spaces[1]
	assert.Equal(t, secondSpace.Name, "staging")
	assert.Equal(t, secondSpace.Guid, "staging-space-guid")
}

func TestSpacesFindAllWithIncorrectToken(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleSpacesEndpoint))
	defer ts.Close()

	repo := CloudControllerSpaceRepository{}

	config := &configuration.Configuration{
		AccessToken:  "BEARER incorrect_access_token",
		Target:       ts.URL,
		Organization: cf.Organization{Guid: "some-org-guid"},
	}

	var (
		spaces []cf.Space
		err    error
	)

	// Capture output so debugging info does not show up in test
	// output
	testhelpers.CaptureOutput(func() {
		spaces, err = repo.FindAll(config)
	})

	assert.Error(t, err)
	assert.Equal(t, 0, len(spaces))
}

func TestSpacesFindByName(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleSpacesEndpoint))
	defer ts.Close()

	repo := CloudControllerSpaceRepository{}
	config := &configuration.Configuration{
		AccessToken:  "BEARER my_access_token",
		Target:       ts.URL,
		Organization: cf.Organization{Guid: "some-org-guid"},
	}
	existingOrg := cf.Space{Guid: "staging-space-guid", Name: "staging"}

	org, err := repo.FindByName(config, "staging")
	assert.NoError(t, err)
	assert.Equal(t, org, existingOrg)

	org, err = repo.FindByName(config, "Staging")
	assert.NoError(t, err)
	assert.Equal(t, org, existingOrg)

	org, err = repo.FindByName(config, "space that does not exist")
	assert.Error(t, err)
}

var spaceSummaryResponse = testhelpers.TestResponse{Status: http.StatusOK, Body: `
{
  "guid": "my-space-guid",
  "name":"development",
  "apps":[{
    "guid":"app-1-guid",
    "urls":["app1.cfapps.io"],
    "routes":[
      {
        "guid":"route-1-guid",
        "host":"app1",
        "domain":{
          "guid":"domain-1-guid",
          "name":"cfapps.io"
        }
      }
    ],
    "running_instances":1,
    "name":"app1",
    "memory":128,
    "instances":1,
    "state":"STARTED"
  },{
    "guid":"app-2-guid",
    "urls":["app2.cfapps.io", "foo.cfapps.io"],
    "routes":[
      {
        "guid":"route-2-guid",
        "host":"app2",
        "domain":{
          "guid":"domain-1-guid",
          "name":"cfapps.io"
        }
      }
    ],
    "running_instances":1,
    "name":"app2",
    "memory":512,
    "instances":3,
    "state":"STARTED"
  }]
}`}

var spaceSummaryEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/spaces/my-space-guid/summary",
	nil,
	spaceSummaryResponse,
)

func TestGetSummary(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(spaceSummaryEndpoint))
	defer ts.Close()

	repo := CloudControllerSpaceRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
	}

	space, err := repo.GetSummary(config)
	assert.NoError(t, err)
	assert.Equal(t, "my-space-guid", space.Guid)
	assert.Equal(t, "development", space.Name)
	assert.Equal(t, 2, len(space.Applications))

	app1 := space.Applications[0]
	assert.Equal(t, app1.Name, "app1")
	assert.Equal(t, app1.Guid, "app-1-guid")
	assert.Equal(t, len(app1.Urls), 1)
	assert.Equal(t, app1.Urls[0], "app1.cfapps.io")

	assert.Equal(t, app1.State, "started")
	assert.Equal(t, app1.Instances, 1)
	assert.Equal(t, app1.RunningInstances, 1)
	assert.Equal(t, app1.Memory, 128)

	app2 := space.Applications[1]
	assert.Equal(t, app2.Name, "app2")
	assert.Equal(t, app2.Guid, "app-2-guid")
	assert.Equal(t, len(app2.Urls), 2)
	assert.Equal(t, app2.Urls[0], "app2.cfapps.io")
	assert.Equal(t, app2.Urls[1], "foo.cfapps.io")

	assert.Equal(t, app2.State, "started")
	assert.Equal(t, app2.Instances, 3)
	assert.Equal(t, app2.RunningInstances, 1)
	assert.Equal(t, app2.Memory, 512)
}
