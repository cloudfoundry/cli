package api_test

import (
	"cf"
	. "cf/api"
	"cf/configuration"
	"cf/net"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testhelpers"
	"testing"
)

var singleAppResponse = testhelpers.TestResponse{Status: http.StatusOK, Body: `
{
  "resources": [
    {
      "metadata": {
        "guid": "app1-guid"
      },
      "entity": {
        "name": "App1",
        "environment_json": {
      		"foo": "bar",
      		"baz": "boom"
    	},
        "memory": 256,
        "instances": 1,
        "state": "STOPPED",
        "routes": [
      	  {
      	    "metadata": {
      	      "guid": "app1-route-guid"
      	    },
      	    "entity": {
      	      "host": "app1",
      	      "domain": {
      	      	"metadata": {
      	      	  "guid": "domain1-guid"
      	      	},
      	      	"entity": {
      	      	  "name": "cfapps.io"
      	      	}
      	      }
      	    }
      	  }
        ]
      }
    }
  ]
}`}

var findAppEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/spaces/my-space-guid/apps?q=name%3AApp1&inline-relations-depth=1",
	nil,
	singleAppResponse,
)

var appSummaryResponse = testhelpers.TestResponse{Status: http.StatusOK, Body: `
{
  "guid": "app1-guid",
  "name": "App1",
  "routes": [
    {
      "guid": "route-1-guid",
      "host": "app1",
      "domain": {
        "guid": "domain-1-guid",
        "name": "cfapps.io"
      }
    }
  ],
  "running_instances": 1,
  "memory": 128,
  "instances": 1
}`}

var appSummaryEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/apps/app1-guid/summary",
	nil,
	appSummaryResponse,
)

var singleAppEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	if strings.Contains(request.URL.Path, "summary") {
		appSummaryEndpoint(writer, request)
		return
	}

	findAppEndpoint(writer, request)
	return
}

func TestFindByName(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(singleAppEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Name: "my-space", Guid: "my-space-guid"},
	}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerApplicationRepository(config, gateway)

	app, apiStatus := repo.FindByName("App1")
	assert.False(t, apiStatus.IsNotSuccessful())
	assert.Equal(t, app.Name, "App1")
	assert.Equal(t, app.Guid, "app1-guid")
	assert.Equal(t, app.Memory, uint64(128))
	assert.Equal(t, app.Instances, 1)
	assert.Equal(t, app.EnvironmentVars, map[string]string{"foo": "bar", "baz": "boom"})

	assert.Equal(t, len(app.Urls), 1)
	assert.Equal(t, app.Urls[0], "app1.cfapps.io")
}

var appNotFoundResponse = testhelpers.TestResponse{Status: http.StatusOK, Body: `
{
  "resources": []
}`}

var appNotFoundEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/spaces/my-space-guid/apps?q=name%3AApp1&inline-relations-depth=1",
	nil,
	appNotFoundResponse,
)

func TestFindByNameWhenAppIsNotFound(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(appNotFoundEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Name: "my-space", Guid: "my-space-guid"},
	}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerApplicationRepository(config, gateway)

	_, apiStatus := repo.FindByName("App1")
	assert.False(t, apiStatus.IsError())
	assert.True(t, apiStatus.IsNotFound())
}

var setEnvEndpoint = testhelpers.CreateEndpoint(
	"PUT",
	"/v2/apps/app1-guid",
	testhelpers.RequestBodyMatcher(`{"environment_json":{"DATABASE_URL":"mysql://example.com/my-db"}}`),
	testhelpers.TestResponse{Status: http.StatusCreated},
)

func TestSetEnv(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(setEnvEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerApplicationRepository(config, gateway)

	app := cf.Application{Guid: "app1-guid", Name: "App1"}

	apiStatus := repo.SetEnv(app, map[string]string{"DATABASE_URL": "mysql://example.com/my-db"})

	assert.False(t, apiStatus.IsNotSuccessful())
}

var createApplicationResponse = `
{
    "metadata": {
        "guid": "my-cool-app-guid"
    },
    "entity": {
        "name": "my-cool-app"
    }
}`

var createApplicationEndpoint = testhelpers.CreateEndpoint(
	"POST",
	"/v2/apps",
	testhelpers.RequestBodyMatcher(`{"space_guid":"my-space-guid","name":"my-cool-app","instances":3,"buildpack":"buildpack-url","command":null,"memory":2048,"stack_guid":"some-stack-guid","command":"some-command"}`),
	testhelpers.TestResponse{Status: http.StatusCreated, Body: createApplicationResponse},
)

var alwaysSuccessfulEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintln(writer, "{}")
}

func TestCreateApplication(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(createApplicationEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
	}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerApplicationRepository(config, gateway)

	newApp := cf.Application{
		Name:         "my-cool-app",
		Instances:    3,
		Memory:       2048,
		BuildpackUrl: "buildpack-url",
		Stack:        cf.Stack{Guid: "some-stack-guid"},
		Command:      "some-command",
	}

	createdApp, apiStatus := repo.Create(newApp)
	assert.False(t, apiStatus.IsNotSuccessful())

	assert.Equal(t, createdApp, cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"})
}

var createApplicationWithoutBuildpackOrStackEndpoint = testhelpers.CreateEndpoint(
	"POST",
	"/v2/apps",
	testhelpers.RequestBodyMatcher(`{"space_guid":"my-space-guid","name":"my-cool-app","instances":1,"buildpack":null,"command":null,"memory":128,"stack_guid":null,"command":null}`),
	testhelpers.TestResponse{Status: http.StatusCreated, Body: createApplicationResponse},
)

func TestCreateApplicationWithoutBuildpackStackOrCommand(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(createApplicationWithoutBuildpackOrStackEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
	}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerApplicationRepository(config, gateway)

	newApp := cf.Application{
		Name:         "my-cool-app",
		Memory:       128,
		Instances:    1,
		BuildpackUrl: "",
		Stack:        cf.Stack{},
	}

	_, apiStatus := repo.Create(newApp)
	assert.False(t, apiStatus.IsNotSuccessful())
}

func TestCreateRejectsInproperNames(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(alwaysSuccessfulEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{Target: ts.URL}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerApplicationRepository(config, gateway)

	createdApp, apiStatus := repo.Create(cf.Application{Name: "name with space"})
	assert.Equal(t, createdApp, cf.Application{})
	assert.Contains(t, apiStatus.Message, "App name is invalid")

	_, apiStatus = repo.Create(cf.Application{Name: "name-with-inv@lid-chars!"})
	assert.True(t, apiStatus.IsNotSuccessful())

	_, apiStatus = repo.Create(cf.Application{Name: "Valid-Name"})
	assert.False(t, apiStatus.IsNotSuccessful())

	_, apiStatus = repo.Create(cf.Application{Name: "name_with_numbers_2"})
	assert.False(t, apiStatus.IsNotSuccessful())
}

var deleteApplicationEndpoint = testhelpers.CreateEndpoint(
	"DELETE",
	"/v2/apps/my-cool-app-guid?recursive=true",
	nil,
	testhelpers.TestResponse{Status: http.StatusOK, Body: ""},
)

func TestDeleteApplication(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(deleteApplicationEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerApplicationRepository(config, gateway)

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}

	apiStatus := repo.Delete(app)
	assert.False(t, apiStatus.IsNotSuccessful())
}

var renameAppEndpoint = testhelpers.CreateEndpoint(
	"PUT",
	"/v2/apps/my-app-guid",
	testhelpers.RequestBodyMatcher(`{"name":"my-new-app"}`),
	testhelpers.TestResponse{Status: http.StatusCreated},
)

func TestRename(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(renameAppEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: ts.URL}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerApplicationRepository(config, gateway)

	org := cf.Application{Guid: "my-app-guid"}
	apiStatus := repo.Rename(org, "my-new-app")
	assert.False(t, apiStatus.IsNotSuccessful())
}

func testScale(t *testing.T, app cf.Application, expectedBody string) {
	scaleEndpoint := testhelpers.CreateEndpoint(
		"PUT",
		"/v2/apps/my-app-guid",
		testhelpers.RequestBodyMatcher(expectedBody),
		testhelpers.TestResponse{Status: http.StatusCreated},
	)

	ts := httptest.NewTLSServer(http.HandlerFunc(scaleEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: ts.URL}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerApplicationRepository(config, gateway)

	apiStatus := repo.Scale(app)
	assert.False(t, apiStatus.IsNotSuccessful())
}

func TestScaleAll(t *testing.T) {
	app := cf.Application{
		Guid:      "my-app-guid",
		DiskQuota: 1024,
		Instances: 5,
		Memory:    512,
	}
	testScale(t, app, `{"disk_quota":1024,"instances":5,"memory":512}`)
}

func TestScaleApplicationDiskQuota(t *testing.T) {
	app := cf.Application{
		Guid:      "my-app-guid",
		DiskQuota: 1024,
	}
	testScale(t, app, `{"disk_quota":1024}`)
}

func TestScaleApplicationInstances(t *testing.T) {
	app := cf.Application{
		Guid:      "my-app-guid",
		Instances: 5,
	}
	testScale(t, app, `{"instances":5}`)
}

func TestScaleApplicationMemory(t *testing.T) {
	app := cf.Application{
		Guid:   "my-app-guid",
		Memory: 512,
	}
	testScale(t, app, `{"memory":512}`)
}

var startApplicationEndpoint = testhelpers.CreateEndpoint(
	"PUT",
	"/v2/apps/my-cool-app-guid",
	testhelpers.RequestBodyMatcher(`{"console":true,"state":"STARTED"}`),
	testhelpers.TestResponse{Status: http.StatusCreated, Body: `
{
  "metadata": {
    "guid": "my-updated-app-guid"
  },
  "entity": {
    "name": "cli1",
    "state": "STARTED"
  }
}`},
)

func TestStartApplication(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(startApplicationEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerApplicationRepository(config, gateway)

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}

	updatedApp, apiStatus := repo.Start(app)
	assert.False(t, apiStatus.IsNotSuccessful())
	assert.Equal(t, "cli1", updatedApp.Name)
	assert.Equal(t, "started", updatedApp.State)
	assert.Equal(t, "my-updated-app-guid", updatedApp.Guid)
}

var stopApplicationEndpoint = testhelpers.CreateEndpoint(
	"PUT",
	"/v2/apps/my-cool-app-guid",
	testhelpers.RequestBodyMatcher(`{"console":true,"state":"STOPPED"}`),
	testhelpers.TestResponse{Status: http.StatusCreated, Body: `
{
  "metadata": {
    "guid": "my-updated-app-guid"
  },
  "entity": {
    "name": "cli1",
    "state": "STOPPED"
  }
}`},
)

func TestStopApplication(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(stopApplicationEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerApplicationRepository(config, gateway)

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}

	updatedApp, apiStatus := repo.Stop(app)
	assert.False(t, apiStatus.IsNotSuccessful())
	assert.Equal(t, "cli1", updatedApp.Name)
	assert.Equal(t, "stopped", updatedApp.State)
	assert.Equal(t, "my-updated-app-guid", updatedApp.Guid)
}

var successfulGetInstancesEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/apps/my-cool-app-guid/instances",
	nil,
	testhelpers.TestResponse{Status: http.StatusCreated, Body: `
{
  "1": {
    "state": "STARTING"
  },
  "0": {
    "state": "RUNNING"
  }
}`},
)

func TestGetInstances(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(successfulGetInstancesEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerApplicationRepository(config, gateway)

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}

	instances, apiStatus := repo.GetInstances(app)
	assert.False(t, apiStatus.IsNotSuccessful())
	assert.Equal(t, len(instances), 2)
	assert.Equal(t, instances[0].State, "running")
	assert.Equal(t, instances[1].State, "starting")
}
