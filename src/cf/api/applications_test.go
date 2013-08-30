package api_test

import (
	"bytes"
	"cf"
	. "cf/api"
	"cf/configuration"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testhelpers"
	"testing"
)

var multipleAppsResponse = testhelpers.TestResponse{Status: http.StatusOK, Body: `
{
  "resources": [
    {
      "metadata": {
        "guid": "app1-guid"
      },
      "entity": {
        "name": "App1",
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
    },
    {
      "metadata": {
        "guid": "app2-guid"
      },
      "entity": {
        "name": "App2",
        "memory": 512,
        "instances": 2,
        "state": "STARTED",
        "routes": [
      	  {
      	    "metadata": {
      	      "guid": "app2-route-guid"
      	    },
      	    "entity": {
      	      "host": "app2",
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

var multipleAppsEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/spaces/my-space-guid/apps?inline-relations-depth=2",
	nil,
	multipleAppsResponse,
)

func TestApplicationsFindAll(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleAppsEndpoint))
	defer ts.Close()

	repo := CloudControllerApplicationRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Name: "my-space", Guid: "my-space-guid"},
	}

	apps, err := repo.FindAll(config)
	assert.NoError(t, err)
	assert.Equal(t, len(apps), 2)

	app := apps[0]
	assert.Equal(t, app.Name, "App1")
	assert.Equal(t, app.Guid, "app1-guid")
	assert.Equal(t, app.State, "stopped")
	assert.Equal(t, app.Instances, 1)
	assert.Equal(t, app.Memory, 256)
	assert.Equal(t, len(app.Urls), 1)
	assert.Equal(t, app.Urls[0], "app1.cfapps.io")

	app = apps[1]
	assert.Equal(t, app.Guid, "app2-guid")

}

func TestFindByName(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleAppsEndpoint))
	defer ts.Close()

	repo := CloudControllerApplicationRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Name: "my-space", Guid: "my-space-guid"},
	}

	app, err := repo.FindByName(config, "App1")
	assert.NoError(t, err)
	assert.Equal(t, app.Name, "App1")
	assert.Equal(t, app.Guid, "app1-guid")

	app, err = repo.FindByName(config, "app1")
	assert.NoError(t, err)
	assert.Equal(t, app.Guid, "app1-guid")

	app, err = repo.FindByName(config, "app that does not exist")
	assert.Error(t, err)
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

	repo := CloudControllerApplicationRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}

	app := cf.Application{Guid: "app1-guid", Name: "App1"}

	err := repo.SetEnv(config, app, "DATABASE_URL", "mysql://example.com/my-db")

	assert.NoError(t, err)
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
	testhelpers.RequestBodyMatcher(`{"space_guid":"my-space-guid","name":"my-cool-app","instances":3,"buildpack":"buildpack-url","command":null,"memory":2048,"stack_guid":null}`),
	testhelpers.TestResponse{Status: http.StatusCreated, Body: createApplicationResponse},
)

var alwaysSuccessfulEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintln(writer, "{}")
}

func TestCreateApplication(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(createApplicationEndpoint))
	defer ts.Close()

	repo := CloudControllerApplicationRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
	}

	newApp := cf.Application{
		Name:         "my-cool-app",
		Instances:    3,
		Memory:       2048,
		BuildpackUrl: "buildpack-url",
	}

	createdApp, err := repo.Create(config, newApp)
	assert.NoError(t, err)

	assert.Equal(t, createdApp, cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"})
}

var createApplicationWithoutBuildpackEndpoint = testhelpers.CreateEndpoint(
	"POST",
	"/v2/apps",
	testhelpers.RequestBodyMatcher(`{"space_guid":"my-space-guid","name":"my-cool-app","instances":1,"buildpack":null,"command":null,"memory":128,"stack_guid":null}`),
	testhelpers.TestResponse{Status: http.StatusCreated, Body: createApplicationResponse},
)

func TestCreateApplicationWithoutBuildpack(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(createApplicationWithoutBuildpackEndpoint))
	defer ts.Close()

	repo := CloudControllerApplicationRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
	}

	newApp := cf.Application{
		Name:         "my-cool-app",
		Memory:       128,
		Instances:    1,
		BuildpackUrl: "",
	}

	_, err := repo.Create(config, newApp)
	assert.NoError(t, err)
}

func TestCreateRejectsInproperNames(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(alwaysSuccessfulEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{Target: ts.URL}
	repo := CloudControllerApplicationRepository{}

	createdApp, err := repo.Create(config, cf.Application{Name: "name with space"})
	assert.Equal(t, createdApp, cf.Application{})
	assert.Contains(t, err.Error(), "Application name is invalid")

	_, err = repo.Create(config, cf.Application{Name: "name-with-inv@lid-chars!"})
	assert.Error(t, err)

	_, err = repo.Create(config, cf.Application{Name: "Valid-Name"})
	assert.NoError(t, err)

	_, err = repo.Create(config, cf.Application{Name: "name_with_numbers_2"})
	assert.NoError(t, err)
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

	repo := CloudControllerApplicationRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}

	err := repo.Delete(config, app)
	assert.NoError(t, err)
}

var uploadBodyMatcher = func(request *http.Request) bool {
	bodyBytes, err := ioutil.ReadAll(request.Body)

	if err != nil {
		return false
	}

	bodyString := string(bodyBytes)
	zipAttachmentContentDispositionMatches := strings.Contains(bodyString, `Content-Disposition: form-data; name="application"; filename="application.zip"`)
	zipAttachmentContentTypeMatches := strings.Contains(bodyString, `Content-Type: application/zip`)
	zipAttachmentContentTransferEncodingMatches := strings.Contains(bodyString, `Content-Transfer-Encoding: binary`)
	zipAttachmentContentLengthPresent := strings.Contains(bodyString, `Content-Length:`)
	zipAttachmentContentPresent := strings.Contains(bodyString, `hello world!`)

	resourcesContentDispositionMatches := strings.Contains(bodyString, `Content-Disposition: form-data; name="resources"`)

	return zipAttachmentContentDispositionMatches &&
		zipAttachmentContentTypeMatches &&
		zipAttachmentContentTransferEncodingMatches &&
		zipAttachmentContentLengthPresent &&
		zipAttachmentContentPresent &&
		resourcesContentDispositionMatches
}

var uploadApplicationEndpoint = testhelpers.CreateEndpoint(
	"PUT",
	"/v2/apps/my-cool-app-guid/bits",
	uploadBodyMatcher,
	testhelpers.TestResponse{Status: http.StatusCreated},
)

func TestUploadApplication(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(uploadApplicationEndpoint))
	defer ts.Close()

	repo := CloudControllerApplicationRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}
	zipBuffer := bytes.NewBufferString("hello world!")

	err := repo.Upload(config, app, zipBuffer)
	assert.NoError(t, err)
}

var startApplicationEndpoint = testhelpers.CreateEndpoint(
	"PUT",
	"/v2/apps/my-cool-app-guid",
	testhelpers.RequestBodyMatcher(`{"console":true,"state":"STARTED"}`),
	testhelpers.TestResponse{Status: http.StatusCreated, Body: `
{
  "metadata": {
    "guid": "my-cool-app-guid",
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

	repo := CloudControllerApplicationRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}

	err := repo.Start(config, app)
	assert.NoError(t, err)
}

var stopApplicationEndpoint = testhelpers.CreateEndpoint(
	"PUT",
	"/v2/apps/my-cool-app-guid",
	testhelpers.RequestBodyMatcher(`{"console":true,"state":"STOPPED"}`),
	testhelpers.TestResponse{Status: http.StatusCreated, Body: `
{
  "metadata": {
    "guid": "my-cool-app-guid",
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

	repo := CloudControllerApplicationRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}

	err := repo.Stop(config, app)
	assert.NoError(t, err)
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

	repo := CloudControllerApplicationRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}

	instances, code, err := repo.GetInstances(config, app)
	assert.NoError(t, err)
	assert.Equal(t, code, 0)
	assert.Equal(t, len(instances), 2)
	assert.Equal(t, instances[0].State, "running")
	assert.Equal(t, instances[1].State, "starting")
}
