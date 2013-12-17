package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"generic"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testnet "testhelpers/net"
	"testing"
)

var singleAppResponse = testnet.TestResponse{
	Status: http.StatusOK,
	Body: `
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
        "memory": 128,
        "instances": 1,
        "state": "STOPPED",
        "stack": {
			"metadata": {
				  "guid": "app1-route-guid"
				},
			"entity": {
			  "name": "awesome-stacks-ahoy"
			  }
        },
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

var findAppRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method:   "GET",
	Path:     "/v2/spaces/my-space-guid/apps?q=name%3AApp1&inline-relations-depth=1",
	Response: singleAppResponse,
})

func TestFindByName(t *testing.T) {
	ts, handler, repo := createAppRepo(t, []testnet.TestRequest{findAppRequest})
	defer ts.Close()

	app, apiResponse := repo.Read("App1")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, app.Name, "App1")
	assert.Equal(t, app.Guid, "app1-guid")
	assert.Equal(t, app.Memory, uint64(128))
	assert.Equal(t, app.InstanceCount, 1)
	assert.Equal(t, app.EnvironmentVars, map[string]string{"foo": "bar", "baz": "boom"})
	assert.Equal(t, app.Routes[0].Host, "app1")
	assert.Equal(t, app.Routes[0].Domain.Name, "cfapps.io")
	assert.Equal(t, app.Stack.Name, "awesome-stacks-ahoy")
}

func TestFindByNameWhenAppIsNotFound(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(findAppRequest)
	request.Response = testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`}

	ts, handler, repo := createAppRepo(t, []testnet.TestRequest{request})
	defer ts.Close()

	_, apiResponse := repo.Read("App1")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsError())
	assert.True(t, apiResponse.IsNotFound())
}

func TestSetEnv(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     "/v2/apps/app1-guid",
		Matcher:  testnet.RequestBodyMatcher(`{"environment_json":{"DATABASE_URL":"mysql://example.com/my-db"}}`),
		Response: testnet.TestResponse{Status: http.StatusCreated},
	})

	ts, handler, repo := createAppRepo(t, []testnet.TestRequest{request})
	defer ts.Close()

	envParams := generic.NewEmptyMap()
	envParams.Set("DATABASE_URL", "mysql://example.com/my-db")

	params := cf.NewEmptyAppParams()
	params.Set("env_vars", envParams)

	_, apiResponse := repo.Update("app1-guid", params)

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
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

var createApplicationRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method:  "POST",
	Path:    "/v2/apps",
	Matcher: testnet.RequestBodyMatcher(`{"name":"my-cool-app","instances":3,"buildpack":"buildpack-url","memory":2048,"space_guid":"some-space-guid","stack_guid":"some-stack-guid","command":"some-command"}`),
	Response: testnet.TestResponse{
		Status: http.StatusCreated,
		Body:   createApplicationResponse},
})

func TestCreateApplication(t *testing.T) {
	ts, handler, repo := createAppRepo(t, []testnet.TestRequest{createApplicationRequest})
	defer ts.Close()

	params := cf.NewEmptyAppParams()
	params.Set("name", "my-cool-app")
	params.Set("buildpack", "buildpack-url")
	params.Set("space_guid", "some-space-guid")
	params.Set("stack_guid", "some-stack-guid")
	params.Set("command", "some-command")
	params.Set("memory", 2048)
	params.Set("instances", 3)

	createdApp, apiResponse := repo.Create(params)

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
	app := cf.Application{}
	app.Name = "my-cool-app"
	app.Guid = "my-cool-app-guid"
	assert.Equal(t, createdApp, app)
}

func TestCreateApplicationWithoutBuildpackStackOrCommand(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "POST",
		Path:     "/v2/apps",
		Matcher:  testnet.RequestBodyMatcher(`{"name":"my-cool-app","instances":1,"buildpack":null,"memory":128,"space_guid":"some-space-guid","stack_guid":null,"command":null}`),
		Response: testnet.TestResponse{Status: http.StatusCreated, Body: createApplicationResponse},
	})

	ts, handler, repo := createAppRepo(t, []testnet.TestRequest{request})
	defer ts.Close()

	params := cf.NewEmptyAppParams()
	params.Set("name", "my-cool-app")
	params.Set("space_guid", "some-space-guid")
	params.Set("memory", 128)
	params.Set("instances", 1)

	_, apiResponse := repo.Create(params)
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestCreateRejectsInproperNames(t *testing.T) {
	baseRequest := testnet.TestRequest{
		Method:   "POST",
		Path:     "/v2/apps",
		Response: testnet.TestResponse{Status: http.StatusCreated, Body: "{}"},
	}

	requests := []testnet.TestRequest{
		baseRequest,
		baseRequest,
	}

	ts, _, repo := createAppRepo(t, requests)
	defer ts.Close()

	params := cf.NewEmptyAppParams()
	params.Set("name", "name with space")
	params.Set("space_guid", "some-space-guid")

	createdApp, apiResponse := repo.Create(params)
	assert.Equal(t, createdApp, cf.Application{})
	assert.Contains(t, apiResponse.Message, "App name is invalid")

	params = cf.NewEmptyAppParams()
	params.Set("name", "name-with-inv@lid-chars!")
	params.Set("space_guid", "some-space-guid")
	_, apiResponse = repo.Create(params)
	assert.True(t, apiResponse.IsNotSuccessful())

	params = cf.NewEmptyAppParams()
	params.Set("name", "Valid-Name")
	params.Set("space_guid", "some-space-guid")
	_, apiResponse = repo.Create(params)
	assert.True(t, apiResponse.IsSuccessful())

	params = cf.NewEmptyAppParams()
	params.Set("name", "name_with_numbers_2")
	params.Set("space_guid", "some-space-guid")
	_, apiResponse = repo.Create(params)
	assert.True(t, apiResponse.IsSuccessful())
}

var updateApplicationResponse = `
{
    "metadata": {
        "guid": "my-cool-app-guid"
    },
    "entity": {
        "name": "my-cool-app"
    }
}`

var updateApplicationRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method:  "PUT",
	Path:    "/v2/apps/my-app-guid?inline-relations-depth=1",
	Matcher: testnet.RequestBodyMatcher(`{"name":"my-cool-app","instances":3,"buildpack":"buildpack-url","memory":2048,"space_guid":"some-space-guid","stack_guid":"some-stack-guid","command":"some-command"}`),
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body:   updateApplicationResponse},
})

func TestUpdateApplication(t *testing.T) {
	ts, handler, repo := createAppRepo(t, []testnet.TestRequest{updateApplicationRequest})
	defer ts.Close()

	app := cf.Application{}
	app.Guid = "my-app-guid"
	app.Name = "my-cool-app"
	app.BuildpackUrl = "buildpack-url"
	app.Command = "some-command"
	app.Memory = 2048
	app.InstanceCount = 3
	app.Stack.Guid = "some-stack-guid"

	updatedApp, apiResponse := repo.Update(app.Guid, app.ToParams())

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, updatedApp.Name, "my-cool-app")
	assert.Equal(t, updatedApp.Guid, "my-cool-app-guid")
}

func TestUpdateApplicationWithoutBuildpackStackOrCommand(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     "/v2/apps/my-app-guid",
		Matcher:  testnet.RequestBodyMatcher(`{"name":"my-cool-app","instances":1,"buildpack":null,"memory":128,"space_guid":"some-space-guid","stack_guid":null,"command":null}`),
		Response: testnet.TestResponse{Status: http.StatusOK, Body: updateApplicationResponse},
	})

	ts, handler, repo := createAppRepo(t, []testnet.TestRequest{request})
	defer ts.Close()

	app := cf.Application{}
	app.Guid = "my-app-guid"
	app.Name = "my-cool-app"
	app.Memory = 128
	app.InstanceCount = 1

	_, apiResponse := repo.Update(app.Guid, app.ToParams())
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestUpdateApplicationSetCommandToNull(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     "/v2/apps/my-app-guid",
		Matcher:  testnet.RequestBodyMatcher(`{"name":"my-cool-app","instances":0,"buildpack":null,"memory":0,"space_guid":"some-space-guid","stack_guid":null,"command":""}`),
		Response: testnet.TestResponse{Status: http.StatusOK, Body: updateApplicationResponse},
	})

	ts, handler, repo := createAppRepo(t, []testnet.TestRequest{request})
	defer ts.Close()

	app := cf.Application{}
	app.Guid = "my-app-guid"
	app.Name = "my-cool-app"
	app.Command = "null"

	_, apiResponse := repo.Update(app.Guid, app.ToParams())
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestUpdateRejectsInproperNames(t *testing.T) {
	baseRequest := testnet.TestRequest{
		Method:   "PUT",
		Path:     "/v2/apps/my-app-guid",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: "{}"},
	}

	requests := []testnet.TestRequest{
		baseRequest,
		baseRequest,
	}

	ts, _, repo := createAppRepo(t, requests)
	defer ts.Close()

	app := cf.Application{}
	app.Guid = "my-app-guid"
	app.Name = "name with space"

	createdApp, apiResponse := repo.Update(app.Guid, app.ToParams())
	assert.Equal(t, createdApp, cf.Application{})
	assert.Contains(t, apiResponse.Message, "App name is invalid")

	app.Name = "name-with-inv@lid-chars!"
	_, apiResponse = repo.Update(app.Guid, app.ToParams())
	assert.True(t, apiResponse.IsNotSuccessful())

	app.Name = "Valid-Name"
	_, apiResponse = repo.Update(app.Guid, app.ToParams())
	assert.True(t, apiResponse.IsSuccessful())

	app.Name = "name_with_numbers_2"
	_, apiResponse = repo.Update(app.Guid, app.ToParams())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestDeleteApplication(t *testing.T) {
	deleteApplicationRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "DELETE",
		Path:     "/v2/apps/my-cool-app-guid?recursive=true",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: ""},
	})

	ts, handler, repo := createAppRepo(t, []testnet.TestRequest{deleteApplicationRequest})
	defer ts.Close()

	apiResponse := repo.Delete("my-cool-app-guid")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestRename(t *testing.T) {
	renameApplicationRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     "/v2/apps/my-app-guid",
		Matcher:  testnet.RequestBodyMatcher(`{"name":"my-new-app"}`),
		Response: testnet.TestResponse{Status: http.StatusCreated},
	})

	ts, handler, repo := createAppRepo(t, []testnet.TestRequest{renameApplicationRequest})
	defer ts.Close()

	params := cf.NewEmptyAppParams()
	params.Set("name", "my-new-app")
	_, apiResponse := repo.Update("my-app-guid", params)

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func testScale(t *testing.T, app cf.Application, expectedBody string) {
	scaleApplicationRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     "/v2/apps/my-app-guid",
		Matcher:  testnet.RequestBodyMatcher(expectedBody),
		Response: testnet.TestResponse{Status: http.StatusCreated},
	})

	ts, handler, repo := createAppRepo(t, []testnet.TestRequest{scaleApplicationRequest})
	defer ts.Close()

	_, apiResponse := repo.Update(app.Guid, app.ToParams())

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestScaleAll(t *testing.T) {
	app := cf.Application{}
	app.Guid = "my-app-guid"
	app.DiskQuota = 1024
	app.InstanceCount = 5
	app.Memory = 512

	testScale(t, app, `{"disk_quota":1024,"instances":5,"memory":512}`)
}

func TestScaleApplicationDiskQuota(t *testing.T) {
	app := cf.Application{}
	app.Guid = "my-app-guid"
	app.DiskQuota = 1024

	testScale(t, app, `{"disk_quota":1024}`)
}

func TestScaleApplicationInstances(t *testing.T) {
	app := cf.Application{}
	app.Guid = "my-app-guid"
	app.InstanceCount = 5

	testScale(t, app, `{"instances":5}`)
}

func TestScaleApplicationMemory(t *testing.T) {
	app := cf.Application{}
	app.Guid = "my-app-guid"
	app.Memory = 512

	testScale(t, app, `{"memory":512}`)
}

func TestStartApplication(t *testing.T) {
	startApplicationRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:  "PUT",
		Path:    "/v2/apps/my-cool-app-guid?inline-relations-depth=1",
		Matcher: testnet.RequestBodyMatcher(`{"state":"STARTED"}`),
		Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
{
  "metadata": {
    "guid": "my-updated-app-guid"
  },
  "entity": {
    "name": "cli1",
    "state": "STARTED"
  }
}`},
	})

	ts, handler, repo := createAppRepo(t, []testnet.TestRequest{startApplicationRequest})
	defer ts.Close()

	params := cf.NewEmptyAppParams()
	params.Set("state", "STARTED")
	updatedApp, apiResponse := repo.Update("my-cool-app-guid", params)

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, "cli1", updatedApp.Name)
	assert.Equal(t, "started", updatedApp.State)
	assert.Equal(t, "my-updated-app-guid", updatedApp.Guid)
}

func TestStopApplication(t *testing.T) {
	stopApplicationRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:  "PUT",
		Path:    "/v2/apps/my-cool-app-guid?inline-relations-depth=1",
		Matcher: testnet.RequestBodyMatcher(`{"state":"STOPPED"}`),
		Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
{
  "metadata": {
    "guid": "my-updated-app-guid"
  },
  "entity": {
    "name": "cli1",
    "state": "STOPPED"
  }
}`},
	})

	ts, handler, repo := createAppRepo(t, []testnet.TestRequest{stopApplicationRequest})
	defer ts.Close()

	params := cf.NewEmptyAppParams()
	params.Set("state", "STOPPED")
	updatedApp, apiResponse := repo.Update("my-cool-app-guid", params)

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, "cli1", updatedApp.Name)
	assert.Equal(t, "stopped", updatedApp.State)
	assert.Equal(t, "my-updated-app-guid", updatedApp.Guid)
}

func createAppRepo(t *testing.T, requests []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo ApplicationRepository) {
	ts, handler = testnet.NewTLSServer(t, requests)
	space := cf.SpaceFields{}
	space.Name = "my-space"
	space.Guid = "my-space-guid"
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		SpaceFields: space,
	}
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerApplicationRepository(config, gateway)
	return
}
