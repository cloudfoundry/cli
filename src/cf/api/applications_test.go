package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
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

	app, apiResponse := repo.FindByName("App1")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, app.Name, "App1")
	assert.Equal(t, app.Guid, "app1-guid")
	assert.Equal(t, app.Memory, uint64(128))
	assert.Equal(t, app.Instances, 1)
	assert.Equal(t, app.EnvironmentVars, map[string]string{"foo": "bar", "baz": "boom"})
	assert.Equal(t, app.Routes[0].Host, "app1")
	assert.Equal(t, app.Routes[0].Domain.Name, "cfapps.io")
}

func TestFindByNameWhenAppIsNotFound(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(findAppRequest)
	request.Response = testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`}

	ts, handler, repo := createAppRepo(t, []testnet.TestRequest{request})
	defer ts.Close()

	_, apiResponse := repo.FindByName("App1")
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

	app := cf.Application{Guid: "app1-guid", Name: "App1"}

	apiResponse := repo.SetEnv(app, map[string]string{"DATABASE_URL": "mysql://example.com/my-db"})

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
	Matcher: testnet.RequestBodyMatcher(`{"space_guid":"my-space-guid","name":"my-cool-app","instances":3,"buildpack":"buildpack-url","command":null,"memory":2048,"stack_guid":"some-stack-guid","command":"some-command"}`),
	Response: testnet.TestResponse{
		Status: http.StatusCreated,
		Body:   createApplicationResponse},
})

func TestCreateApplication(t *testing.T) {
	ts, handler, repo := createAppRepo(t, []testnet.TestRequest{createApplicationRequest})
	defer ts.Close()

	newApp := cf.Application{
		Name:         "my-cool-app",
		Instances:    3,
		Memory:       2048,
		BuildpackUrl: "buildpack-url",
		Stack:        cf.Stack{Guid: "some-stack-guid"},
		Command:      "some-command",
	}

	createdApp, apiResponse := repo.Create(newApp)
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())

	assert.Equal(t, createdApp, cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"})
}

func TestCreateApplicationWithoutBuildpackStackOrCommand(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "POST",
		Path:     "/v2/apps",
		Matcher:  testnet.RequestBodyMatcher(`{"space_guid":"my-space-guid","name":"my-cool-app","instances":1,"buildpack":null,"command":null,"memory":128,"stack_guid":null,"command":null}`),
		Response: testnet.TestResponse{Status: http.StatusCreated, Body: createApplicationResponse},
	})

	ts, handler, repo := createAppRepo(t, []testnet.TestRequest{request})
	defer ts.Close()

	newApp := cf.Application{
		Name:         "my-cool-app",
		Memory:       128,
		Instances:    1,
		BuildpackUrl: "",
		Stack:        cf.Stack{},
	}

	_, apiResponse := repo.Create(newApp)
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

	createdApp, apiResponse := repo.Create(cf.Application{Name: "name with space"})
	assert.Equal(t, createdApp, cf.Application{})
	assert.Contains(t, apiResponse.Message, "App name is invalid")

	_, apiResponse = repo.Create(cf.Application{Name: "name-with-inv@lid-chars!"})
	assert.True(t, apiResponse.IsNotSuccessful())

	_, apiResponse = repo.Create(cf.Application{Name: "Valid-Name"})
	assert.True(t, apiResponse.IsSuccessful())

	_, apiResponse = repo.Create(cf.Application{Name: "name_with_numbers_2"})
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

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}

	apiResponse := repo.Delete(app)

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

	org := cf.Application{Guid: "my-app-guid"}
	apiResponse := repo.Rename(org, "my-new-app")

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

	apiResponse := repo.Scale(app)

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
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

func TestStartApplication(t *testing.T) {
	startApplicationRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:  "PUT",
		Path:    "/v2/apps/my-cool-app-guid",
		Matcher: testnet.RequestBodyMatcher(`{"console":true,"state":"STARTED"}`),
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

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}
	updatedApp, apiResponse := repo.Start(app)

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, "cli1", updatedApp.Name)
	assert.Equal(t, "started", updatedApp.State)
	assert.Equal(t, "my-updated-app-guid", updatedApp.Guid)
}

func TestStopApplication(t *testing.T) {
	stopApplicationRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:  "PUT",
		Path:    "/v2/apps/my-cool-app-guid",
		Matcher: testnet.RequestBodyMatcher(`{"console":true,"state":"STOPPED"}`),
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

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}
	updatedApp, apiResponse := repo.Stop(app)

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, "cli1", updatedApp.Name)
	assert.Equal(t, "stopped", updatedApp.State)
	assert.Equal(t, "my-updated-app-guid", updatedApp.Guid)
}

func TestGetInstances(t *testing.T) {
	getInstancesRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/apps/my-cool-app-guid/instances",
		Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
{
  "1": {
    "state": "STARTING"
  },
  "0": {
    "state": "RUNNING"
  }
}`},
	})

	ts, handler, repo := createAppRepo(t, []testnet.TestRequest{getInstancesRequest})
	defer ts.Close()

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}
	instances, apiResponse := repo.GetInstances(app)

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, len(instances), 2)
	assert.Equal(t, instances[0].State, "running")
	assert.Equal(t, instances[1].State, "starting")
}

func createAppRepo(t *testing.T, requests []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo ApplicationRepository) {
	ts, handler = testnet.NewTLSServer(t, requests)

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Name: "my-space", Guid: "my-space-guid"},
	}
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerApplicationRepository(config, gateway)
	return
}
