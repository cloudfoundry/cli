package api_test

import (
	"cf"
	. "cf/api"
	"cf/configuration"
	"cf/net"
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

	config := &configuration.Configuration{
		AccessToken:  "BEARER my_access_token",
		Target:       ts.URL,
		Organization: cf.Organization{Guid: "some-org-guid"},
	}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerSpaceRepository(config, gateway)
	spaces, apiResponse := repo.FindAll()

	assert.False(t, apiResponse.IsNotSuccessful())
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

	config := &configuration.Configuration{
		AccessToken:  "BEARER incorrect_access_token",
		Target:       ts.URL,
		Organization: cf.Organization{Guid: "some-org-guid"},
	}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerSpaceRepository(config, gateway)

	var (
		spaces      []cf.Space
		apiResponse net.ApiResponse
	)

	// Capture output so debugging info does not show up in test
	// output
	testhelpers.CaptureOutput(func() {
		spaces, apiResponse = repo.FindAll()
	})

	assert.True(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, 0, len(spaces))
}

var findSpaceByNameResponse = testhelpers.TestResponse{Status: http.StatusOK, Body: `
{
  "total_results": 1,
  "total_pages": 1,
  "prev_url": null,
  "next_url": null,
  "resources": [
    {
      "metadata": {
        "guid": "space1-guid"
      },
      "entity": {
        "name": "Space1",
        "organization_guid": "org1-guid",
        "organization": {
          "metadata": {
            "guid": "org1-guid"
          },
          "entity": {
            "name": "Org1"
          }
        },
        "apps": [
          {
            "metadata": {
              "guid": "app1-guid"
            },
            "entity": {
              "name": "app1"
            }
          },
          {
            "metadata": {
              "guid": "app2-guid"
            },
            "entity": {
              "name": "app2"
            }
          }
        ],
        "domains": [
          {
            "metadata": {
              "guid": "domain1-guid"
            },
            "entity": {
              "name": "domain1"
            }
          }
        ],
        "service_instances": [
          {
			"metadata": {
              "guid": "service1-guid"
            },
            "entity": {
              "name": "service1"
            }
          }
        ]
      }
    }
  ]
}`}

var findSpaceByNameEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/organizations/org-guid/spaces?q=name%3Aspace1&inline-relations-depth=1",
	nil,
	findSpaceByNameResponse,
)

func TestSpacesFindByName(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(findSpaceByNameEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken:  "BEARER my_access_token",
		Target:       ts.URL,
		Organization: cf.Organization{Guid: "org-guid"},
	}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerSpaceRepository(config, gateway)
	existingOrg := cf.Organization{Guid: "org1-guid", Name: "Org1"}
	apps := []cf.Application{
		cf.Application{Name: "app1", Guid: "app1-guid"},
		cf.Application{Name: "app2", Guid: "app2-guid"},
	}
	domains := []cf.Domain{
		cf.Domain{Name: "domain1", Guid: "domain1-guid"},
	}
	services := []cf.ServiceInstance{
		cf.ServiceInstance{Name: "service1", Guid: "service1-guid"},
	}

	space, apiResponse := repo.FindByName("Space1")
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, space.Name, "Space1")
	assert.Equal(t, space.Guid, "space1-guid")

	assert.Equal(t, space.Organization, existingOrg)
	assert.Equal(t, space.Applications, apps)
	assert.Equal(t, space.Domains, domains)
	assert.Equal(t, space.ServiceInstances, services)

	space, apiResponse = repo.FindByName("space1")
	assert.False(t, apiResponse.IsNotSuccessful())
}

var didNotFindSpaceByNameResponse = testhelpers.TestResponse{Status: http.StatusOK, Body: `
{
  "total_results": 0,
  "total_pages": 0,
  "prev_url": null,
  "next_url": null,
  "resources": [

  ]
}`}

var didNotFindSpaceByNameEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/organizations/org-guid/spaces?q=name%3Aspace1&inline-relations-depth=1",
	nil,
	didNotFindSpaceByNameResponse,
)

func TestSpacesDidNotFindByName(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(didNotFindSpaceByNameEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken:  "BEARER my_access_token",
		Target:       ts.URL,
		Organization: cf.Organization{Guid: "org-guid"},
	}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerSpaceRepository(config, gateway)

	_, apiResponse := repo.FindByName("space1")
	assert.False(t, apiResponse.IsError())
	assert.True(t, apiResponse.IsNotFound())
}

var spaceSummaryResponse = testhelpers.TestResponse{Status: http.StatusOK, Body: `
{
  "guid": "my-space-guid",
  "name":"development",
  "apps":[
    {
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
      "state":"STARTED",
      "service_names":[
      	"my-service-instance"
      ]
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
      "state":"STARTED",
      "service_names":[
      	"my-service-instance"
      ]
    }
  ],
  "services": [
    {
      "guid": "my-service-instance-guid",
      "name": "my-service-instance",
      "bound_app_count": 2,
      "service_plan": {
        "guid": "service-plan-guid",
        "name": "spark",
        "service": {
          "guid": "service-offering-guid",
          "label": "cleardb",
          "provider": "cleardb-provider",
          "version": "n/a"
        }
      }
    }
  ]
}`}

var spaceSummaryEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/spaces/my-space-guid/summary",
	nil,
	spaceSummaryResponse,
)

func TestSpacesGetSummary(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(spaceSummaryEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
	}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerSpaceRepository(config, gateway)

	space, apiResponse := repo.GetSummary()
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, "my-space-guid", space.Guid)
	assert.Equal(t, "development", space.Name)
	assert.Equal(t, 2, len(space.Applications))
	assert.Equal(t, 1, len(space.ServiceInstances))

	app1 := space.Applications[0]
	assert.Equal(t, app1.Name, "app1")
	assert.Equal(t, app1.Guid, "app-1-guid")
	assert.Equal(t, len(app1.Urls), 1)
	assert.Equal(t, app1.Urls[0], "app1.cfapps.io")

	assert.Equal(t, app1.State, "started")
	assert.Equal(t, app1.Instances, 1)
	assert.Equal(t, app1.RunningInstances, 1)
	assert.Equal(t, app1.Memory, uint64(128))

	app2 := space.Applications[1]
	assert.Equal(t, app2.Name, "app2")
	assert.Equal(t, app2.Guid, "app-2-guid")
	assert.Equal(t, len(app2.Urls), 2)
	assert.Equal(t, app2.Urls[0], "app2.cfapps.io")
	assert.Equal(t, app2.Urls[1], "foo.cfapps.io")

	assert.Equal(t, app2.State, "started")
	assert.Equal(t, app2.Instances, 3)
	assert.Equal(t, app2.RunningInstances, 1)
	assert.Equal(t, app2.Memory, uint64(512))

	instance1 := space.ServiceInstances[0]
	assert.Equal(t, instance1.Name, "my-service-instance")
	assert.Equal(t, instance1.ServicePlan.Name, "spark")
	assert.Equal(t, instance1.ServicePlan.ServiceOffering.Label, "cleardb")
	assert.Equal(t, instance1.ServicePlan.ServiceOffering.Provider, "cleardb-provider")
	assert.Equal(t, instance1.ServicePlan.ServiceOffering.Version, "n/a")
	assert.Equal(t, len(instance1.ApplicationNames), 2)
	assert.Equal(t, instance1.ApplicationNames[0], "app1")
	assert.Equal(t, instance1.ApplicationNames[1], "app2")
}

var createSpaceEndpoint = testhelpers.CreateEndpoint(
	"POST",
	"/v2/spaces",
	testhelpers.RequestBodyMatcher(`{"name":"space-name","organization_guid":"org-guid"}`),
	testhelpers.TestResponse{Status: http.StatusCreated},
)

func TestCreateSpace(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(createSpaceEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken:  "BEARER my_access_token",
		Target:       ts.URL,
		Organization: cf.Organization{Guid: "org-guid"},
	}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerSpaceRepository(config, gateway)

	apiResponse := repo.Create("space-name")
	assert.False(t, apiResponse.IsNotSuccessful())
}

var renameSpaceEndpoint = testhelpers.CreateEndpoint(
	"PUT",
	"/v2/spaces/my-space-guid",
	testhelpers.RequestBodyMatcher(`{"name":"new-space-name"}`),
	testhelpers.TestResponse{Status: http.StatusCreated},
)

func TestRenameSpace(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(renameSpaceEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerSpaceRepository(config, gateway)

	space := cf.Space{Guid: "my-space-guid"}
	apiResponse := repo.Rename(space, "new-space-name")
	assert.False(t, apiResponse.IsNotSuccessful())
}

var deleteSpaceEndpoint = testhelpers.CreateEndpoint(
	"DELETE",
	"/v2/spaces/my-space-guid?recursive=true",
	nil,
	testhelpers.TestResponse{Status: http.StatusOK},
)

func TestDeleteSpace(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(deleteSpaceEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerSpaceRepository(config, gateway)

	space := cf.Space{Guid: "my-space-guid"}
	apiResponse := repo.Delete(space)
	assert.False(t, apiResponse.IsNotSuccessful())
}
