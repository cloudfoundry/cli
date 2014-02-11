package api_test

import (
	. "cf/api"
	"cf/models"
	"cf/net"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestFindByName", func() {
			ts, handler, repo := createAppRepo(mr.T(), []testnet.TestRequest{findAppRequest})
			defer ts.Close()

			app, apiResponse := repo.Read("My App")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
			assert.Equal(mr.T(), app.Name, "My App")
			assert.Equal(mr.T(), app.Guid, "app1-guid")
			assert.Equal(mr.T(), app.Memory, uint64(128))
			assert.Equal(mr.T(), app.InstanceCount, 1)
			assert.Equal(mr.T(), app.EnvironmentVars, map[string]string{"foo": "bar", "baz": "boom"})
			assert.Equal(mr.T(), app.Routes[0].Host, "app1")
			assert.Equal(mr.T(), app.Routes[0].Domain.Name, "cfapps.io")
			assert.Equal(mr.T(), app.Stack.Name, "awesome-stacks-ahoy")
		})

		It("TestFindByNameWhenAppIsNotFound", func() {

			request := testapi.NewCloudControllerTestRequest(findAppRequest)
			request.Response = testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`}

			ts, handler, repo := createAppRepo(mr.T(), []testnet.TestRequest{request})
			defer ts.Close()

			_, apiResponse := repo.Read("My App")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsError())
			assert.True(mr.T(), apiResponse.IsNotFound())
		})

		It("TestSetEnv", func() {
			request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/apps/app1-guid",
				Matcher:  testnet.RequestBodyMatcher(`{"environment_json":{"DATABASE_URL":"mysql://example.com/my-db"}}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			ts, handler, repo := createAppRepo(mr.T(), []testnet.TestRequest{request})
			defer ts.Close()

			envParams := map[string]string{"DATABASE_URL": "mysql://example.com/my-db"}
			params := models.AppParams{EnvironmentVars: &envParams}

			_, apiResponse := repo.Update("app1-guid", params)

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
		})

		It("TestCreateApplication", func() {
			ts, handler, repo := createAppRepo(mr.T(), []testnet.TestRequest{createApplicationRequest})
			defer ts.Close()

			params := defaultAppParams()
			createdApp, apiResponse := repo.Create(params)

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())

			app := models.Application{}
			app.Name = "my-cool-app"
			app.Guid = "my-cool-app-guid"
			assert.Equal(mr.T(), createdApp, app)
		})

		It("TestCreateApplicationWithoutBuildpackStackOrCommand", func() {
			request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "POST",
				Path:     "/v2/apps",
				Matcher:  testnet.RequestBodyMatcher(`{"name":"my-cool-app","instances":3,"memory":2048,"space_guid":"some-space-guid"}`),
				Response: testnet.TestResponse{Status: http.StatusCreated, Body: createApplicationResponse},
			})

			ts, handler, repo := createAppRepo(mr.T(), []testnet.TestRequest{request})
			defer ts.Close()

			params := defaultAppParams()
			params.BuildpackUrl = nil
			params.StackGuid = nil
			params.Command = nil

			_, apiResponse := repo.Create(params)
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
		})

		It("TestUpdateApplication", func() {
			ts, handler, repo := createAppRepo(mr.T(), []testnet.TestRequest{updateApplicationRequest})
			defer ts.Close()

			app := models.Application{}
			app.Guid = "my-app-guid"
			app.Name = "my-cool-app"
			app.BuildpackUrl = "buildpack-url"
			app.Command = "some-command"
			app.Memory = 2048
			app.InstanceCount = 3
			app.Stack.Guid = "some-stack-guid"
			app.SpaceGuid = "some-space-guid"
			app.State = "started"

			updatedApp, apiResponse := repo.Update(app.Guid, app.ToParams())

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())
			assert.Equal(mr.T(), updatedApp.Name, "my-cool-app")
			assert.Equal(mr.T(), updatedApp.Guid, "my-cool-app-guid")
		})

		It("TestUpdateApplicationSetCommandToNull", func() {
			request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/apps/my-app-guid",
				Matcher:  testnet.RequestBodyMatcher(`{"command":""}`),
				Response: testnet.TestResponse{Status: http.StatusOK, Body: updateApplicationResponse},
			})

			ts, handler, repo := createAppRepo(mr.T(), []testnet.TestRequest{request})
			defer ts.Close()

			emptyString := ""
			app := models.AppParams{Command: &emptyString}

			_, apiResponse := repo.Update("my-app-guid", app)
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
		})

		It("TestDeleteApplication", func() {
			deleteApplicationRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "DELETE",
				Path:     "/v2/apps/my-cool-app-guid?recursive=true",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: ""},
			})

			ts, handler, repo := createAppRepo(mr.T(), []testnet.TestRequest{deleteApplicationRequest})
			defer ts.Close()

			apiResponse := repo.Delete("my-cool-app-guid")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
		})
	})
}

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
        "name": "My App",
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
	Path:     "/v2/spaces/my-space-guid/apps?q=name%3AMy+App&inline-relations-depth=1",
	Response: singleAppResponse,
})

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
	Method: "POST",
	Path:   "/v2/apps",
	Matcher: testnet.RequestBodyMatcher(`{
	"name":"my-cool-app",
	"instances":3,
	"buildpack":"buildpack-url",
	"memory":2048,
	"space_guid":"some-space-guid",
	"stack_guid":"some-stack-guid",
	"command":"some-command"
	}`),
	Response: testnet.TestResponse{
		Status: http.StatusCreated,
		Body:   createApplicationResponse},
})

func defaultAppParams() models.AppParams {
	name := "my-cool-app"
	buildpackUrl := "buildpack-url"
	spaceGuid := "some-space-guid"
	stackGuid := "some-stack-guid"
	command := "some-command"
	memory := uint64(2048)
	instanceCount := 3

	return models.AppParams{
		Name:          &name,
		BuildpackUrl:  &buildpackUrl,
		SpaceGuid:     &spaceGuid,
		StackGuid:     &stackGuid,
		Command:       &command,
		Memory:        &memory,
		InstanceCount: &instanceCount,
	}
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
	Matcher: testnet.RequestBodyMatcher(`{"name":"my-cool-app","instances":3,"buildpack":"buildpack-url","memory":2048,"space_guid":"some-space-guid","state":"STARTED","stack_guid":"some-stack-guid","command":"some-command"}`),
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body:   updateApplicationResponse},
})

func createAppRepo(t mr.TestingT, requests []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo ApplicationRepository) {
	ts, handler = testnet.NewTLSServer(t, requests)
	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerApplicationRepository(configRepo, gateway)
	return
}
