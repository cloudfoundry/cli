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

var singleAppResponse = testhelpers.TestResponse{Status: http.StatusOK, Body: `
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
	client := NewApiClient(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerApplicationRepository(config, client)

	app, err := repo.FindByName("App1")
	assert.NoError(t, err)
	assert.Equal(t, app.Name, "App1")
	assert.Equal(t, app.Guid, "app1-guid")
	assert.Equal(t, app.Memory, 128)
	assert.Equal(t, app.Instances, 1)

	assert.Equal(t, len(app.Urls), 1)
	assert.Equal(t, app.Urls[0], "app1.cfapps.io")

	app, err = repo.FindByName("app that does not exist")
	assert.Error(t, err)
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
	client := NewApiClient(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerApplicationRepository(config, client)

	_, err := repo.FindByName("App1")
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

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	client := NewApiClient(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerApplicationRepository(config, client)

	app := cf.Application{Guid: "app1-guid", Name: "App1"}

	err := repo.SetEnv(app, "DATABASE_URL", "mysql://example.com/my-db")

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
	testhelpers.RequestBodyMatcher(`{"space_guid":"my-space-guid","name":"my-cool-app","instances":3,"buildpack":"buildpack-url","command":null,"memory":2048,"stack_guid":"some-stack-guid"}`),
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
	client := NewApiClient(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerApplicationRepository(config, client)

	newApp := cf.Application{
		Name:         "my-cool-app",
		Instances:    3,
		Memory:       2048,
		BuildpackUrl: "buildpack-url",
		Stack:        cf.Stack{Guid: "some-stack-guid"},
	}

	createdApp, err := repo.Create(newApp)
	assert.NoError(t, err)

	assert.Equal(t, createdApp, cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"})
}

var createApplicationWithoutBuildpackOrStackEndpoint = testhelpers.CreateEndpoint(
	"POST",
	"/v2/apps",
	testhelpers.RequestBodyMatcher(`{"space_guid":"my-space-guid","name":"my-cool-app","instances":1,"buildpack":null,"command":null,"memory":128,"stack_guid":null}`),
	testhelpers.TestResponse{Status: http.StatusCreated, Body: createApplicationResponse},
)

func TestCreateApplicationWithoutBuildpackOrStack(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(createApplicationWithoutBuildpackOrStackEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
	}
	client := NewApiClient(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerApplicationRepository(config, client)

	newApp := cf.Application{
		Name:         "my-cool-app",
		Memory:       128,
		Instances:    1,
		BuildpackUrl: "",
		Stack:        cf.Stack{},
	}

	_, err := repo.Create(newApp)
	assert.NoError(t, err)
}

func TestCreateRejectsInproperNames(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(alwaysSuccessfulEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{Target: ts.URL}
	client := NewApiClient(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerApplicationRepository(config, client)

	createdApp, err := repo.Create(cf.Application{Name: "name with space"})
	assert.Equal(t, createdApp, cf.Application{})
	assert.Contains(t, err.Error(), "Application name is invalid")

	_, err = repo.Create(cf.Application{Name: "name-with-inv@lid-chars!"})
	assert.Error(t, err)

	_, err = repo.Create(cf.Application{Name: "Valid-Name"})
	assert.NoError(t, err)

	_, err = repo.Create(cf.Application{Name: "name_with_numbers_2"})
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

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	client := NewApiClient(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerApplicationRepository(config, client)

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}

	err := repo.Delete(app)
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

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	client := NewApiClient(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerApplicationRepository(config, client)

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}
	zipBuffer := bytes.NewBufferString("hello world!")

	err := repo.Upload(app, zipBuffer)
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

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	client := NewApiClient(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerApplicationRepository(config, client)

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}

	err := repo.Start(app)
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

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	client := NewApiClient(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerApplicationRepository(config, client)

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}

	err := repo.Stop(app)
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

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	client := NewApiClient(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerApplicationRepository(config, client)

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}

	instances, err := repo.GetInstances(app)
	assert.NoError(t, err)
	assert.Equal(t, len(instances), 2)
	assert.Equal(t, instances[0].State, "running")
	assert.Equal(t, instances[1].State, "starting")
}
