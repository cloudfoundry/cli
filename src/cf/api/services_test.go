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

var multipleOfferingsResponse = testhelpers.TestResponse{Status: http.StatusOK, Body: `
{
  "resources": [
    {
      "metadata": {
        "guid": "offering-1-guid"
      },
      "entity": {
        "label": "Offering 1",
        "service_plans": [
        	{
        		"metadata": {"guid": "offering-1-plan-1-guid"},
        		"entity": {"name": "Offering 1 Plan 1"}
        	},
        	{
        		"metadata": {"guid": "offering-1-plan-2-guid"},
        		"entity": {"name": "Offering 1 Plan 2"}
        	}
        ]
      }
    },
    {
      "metadata": {
        "guid": "offering-2-guid"
      },
      "entity": {
        "label": "Offering 2",
        "service_plans": [
        	{
        		"metadata": {"guid": "offering-2-plan-1-guid"},
        		"entity": {"name": "Offering 2 Plan 1"}
        	}
        ]
      }
    }
  ]
}`}

var multipleOfferingsEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/services?inline-relations-depth=1",
	nil,
	multipleOfferingsResponse,
)

func TestGetServiceOfferings(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleOfferingsEndpoint))
	defer ts.Close()

	repo := CloudControllerServiceRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	offerings, err := repo.GetServiceOfferings(config)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(offerings))

	firstOffering := offerings[0]
	assert.Equal(t, firstOffering.Label, "Offering 1")
	assert.Equal(t, firstOffering.Guid, "offering-1-guid")
	assert.Equal(t, len(firstOffering.Plans), 2)

	plan := firstOffering.Plans[0]
	assert.Equal(t, plan.Name, "Offering 1 Plan 1")
	assert.Equal(t, plan.Guid, "offering-1-plan-1-guid")

	secondOffering := offerings[1]
	assert.Equal(t, secondOffering.Label, "Offering 2")
	assert.Equal(t, secondOffering.Guid, "offering-2-guid")
	assert.Equal(t, len(secondOffering.Plans), 1)
}

var createServiceInstanceEndpoint = testhelpers.CreateEndpoint(
	"POST",
	"/v2/service_instances",
	testhelpers.RequestBodyMatcher(`{"name":"instance-name","service_plan_guid":"plan-guid","space_guid":"space-guid"}`),
	testhelpers.TestResponse{Status: http.StatusCreated},
)

func TestCreateServiceInstance(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(createServiceInstanceEndpoint))
	defer ts.Close()

	repo := CloudControllerServiceRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "space-guid"},
	}

	err := repo.CreateServiceInstance(config, "instance-name", cf.ServicePlan{Guid: "plan-guid"})
	assert.NoError(t, err)
}
