package applications_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testnet "code.cloudfoundry.org/cli/util/testhelpers/net"

	. "code.cloudfoundry.org/cli/cf/api/applications"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("ApplicationsRepository", func() {
	Describe("finding apps by name", func() {
		It("returns the app when it is found", func() {
			ts, handler, repo := createAppRepo([]testnet.TestRequest{findAppRequest})
			defer ts.Close()

			app, apiErr := repo.Read("My App")
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
			Expect(app.Name).To(Equal("My App"))
			Expect(app.GUID).To(Equal("app1-guid"))
			Expect(app.Memory).To(Equal(int64(128)))
			Expect(app.DiskQuota).To(Equal(int64(512)))
			Expect(app.InstanceCount).To(Equal(1))
			Expect(app.EnvironmentVars).To(Equal(map[string]interface{}{"foo": "bar", "baz": "boom"}))
			Expect(app.Routes[0].Host).To(Equal("app1"))
			Expect(app.Routes[0].Domain.Name).To(Equal("cfapps.io"))
			Expect(app.Stack.Name).To(Equal("awesome-stacks-ahoy"))
		})

		It("returns a failure response when the app is not found", func() {
			request := apifakes.NewCloudControllerTestRequest(findAppRequest)
			request.Response = testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`}

			ts, handler, repo := createAppRepo([]testnet.TestRequest{request})
			defer ts.Close()

			_, apiErr := repo.Read("My App")
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr.(*errors.ModelNotFoundError)).NotTo(BeNil())
		})
	})

	Describe(".GetApp", func() {
		It("returns an application using the given app guid", func() {
			request := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/apps/app-guid",
				Response: appModelResponse,
			})
			ts, handler, repo := createAppRepo([]testnet.TestRequest{request})
			defer ts.Close()
			app, err := repo.GetApp("app-guid")

			Expect(err).ToNot(HaveOccurred())
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(app.Name).To(Equal("My App"))
		})
	})

	Describe(".ReadFromSpace", func() {
		It("returns an application using the given space guid", func() {
			request := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/spaces/another-space-guid/apps?q=name%3AMy+App&inline-relations-depth=1",
				Response: singleAppResponse,
			})
			ts, handler, repo := createAppRepo([]testnet.TestRequest{request})
			defer ts.Close()
			app, err := repo.ReadFromSpace("My App", "another-space-guid")

			Expect(err).ToNot(HaveOccurred())
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(app.Name).To(Equal("My App"))
		})
	})

	Describe("Create", func() {
		var (
			ccServer  *ghttp.Server
			repo      CloudControllerRepository
			appParams models.AppParams
		)

		BeforeEach(func() {
			ccServer = ghttp.NewServer()
			configRepo := testconfig.NewRepositoryWithDefaults()
			configRepo.SetAPIEndpoint(ccServer.URL())
			gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
			repo = NewCloudControllerRepository(configRepo, gateway)

			name := "my-cool-app"
			buildpackURL := "buildpack-url"
			spaceGUID := "some-space-guid"
			stackGUID := "some-stack-guid"
			command := "some-command"
			memory := int64(2048)
			diskQuota := int64(512)
			instanceCount := 3

			appParams = models.AppParams{
				Name:          &name,
				BuildpackURL:  &buildpackURL,
				SpaceGUID:     &spaceGUID,
				StackGUID:     &stackGUID,
				Command:       &command,
				Memory:        &memory,
				DiskQuota:     &diskQuota,
				InstanceCount: &instanceCount,
			}

			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v2/apps"),
					ghttp.VerifyJSON(`{
						"name":"my-cool-app",
						"instances":3,
						"buildpack":"buildpack-url",
						"memory":2048,
						"disk_quota": 512,
						"space_guid":"some-space-guid",
						"stack_guid":"some-stack-guid",
						"command":"some-command"
					}`),
				),
			)
		})

		AfterEach(func() {
			ccServer.Close()
		})

		It("tries to create the app", func() {
			repo.Create(appParams)
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})

		Context("when the create succeeds", func() {
			BeforeEach(func() {
				h := ccServer.GetHandler(0)
				ccServer.SetHandler(0,
					ghttp.CombineHandlers(
						h,
						ghttp.RespondWith(http.StatusCreated, `{
							"metadata": {
									"guid": "my-cool-app-guid"
							},
							"entity": {
									"name": "my-cool-app"
							}
					}`),
					),
				)
			})

			It("returns the application", func() {
				createdApp, err := repo.Create(appParams)
				Expect(err).NotTo(HaveOccurred())

				app := models.Application{}
				app.Name = "my-cool-app"
				app.GUID = "my-cool-app-guid"
				Expect(createdApp).To(Equal(app))
			})
		})

		Context("when the create fails", func() {
			BeforeEach(func() {
				h := ccServer.GetHandler(0)
				ccServer.SetHandler(0,
					ghttp.CombineHandlers(
						h,
						ghttp.RespondWith(http.StatusInternalServerError, ""),
					),
				)
			})

			It("returns an error", func() {
				_, err := repo.Create(appParams)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("reading environment for an app", func() {
		Context("when the response can be parsed as json", func() {
			var (
				testServer *httptest.Server
				userEnv    *models.Environment
				err        error
				handler    *testnet.TestHandler
				repo       Repository
			)

			AfterEach(func() {
				testServer.Close()
			})

			Context("when there are system provided env vars", func() {
				BeforeEach(func() {

					var appEnvRequest = apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
						Method: "GET",
						Path:   "/v2/apps/some-cool-app-guid/env",
						Response: testnet.TestResponse{
							Status: http.StatusOK,
							Body: `
{
	 "staging_env_json": {
     "STAGING_ENV": "staging_value",
		 "staging": true,
		 "number": 42
   },
   "running_env_json": {
     "RUNNING_ENV": "running_value",
		 "running": false,
		 "number": 37
   },
   "environment_json": {
     "key": "value",
		 "number": 123,
		 "bool": true
   },
   "system_env_json": {
     "VCAP_SERVICES": {
				"system_hash": {
          "system_key": "system_value"
        }
     }
   }
}
`,
						}})

					testServer, handler, repo = createAppRepo([]testnet.TestRequest{appEnvRequest})
					userEnv, err = repo.ReadEnv("some-cool-app-guid")
					Expect(err).ToNot(HaveOccurred())
					Expect(handler).To(HaveAllRequestsCalled())
				})

				It("returns the user environment, vcap services, running/staging env variables", func() {
					Expect(userEnv.Environment["key"]).To(Equal("value"))
					Expect(userEnv.Environment["number"]).To(Equal(float64(123)))
					Expect(userEnv.Environment["bool"]).To(BeTrue())
					Expect(userEnv.Running["RUNNING_ENV"]).To(Equal("running_value"))
					Expect(userEnv.Running["running"]).To(BeFalse())
					Expect(userEnv.Running["number"]).To(Equal(float64(37)))
					Expect(userEnv.Staging["STAGING_ENV"]).To(Equal("staging_value"))
					Expect(userEnv.Staging["staging"]).To(BeTrue())
					Expect(userEnv.Staging["number"]).To(Equal(float64(42)))

					vcapServices := userEnv.System["VCAP_SERVICES"]
					data, err := json.Marshal(vcapServices)

					Expect(err).ToNot(HaveOccurred())
					Expect(string(data)).To(ContainSubstring("\"system_key\":\"system_value\""))
				})

			})

			Context("when there are no environment variables", func() {
				BeforeEach(func() {
					var emptyEnvRequest = apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
						Method: "GET",
						Path:   "/v2/apps/some-cool-app-guid/env",
						Response: testnet.TestResponse{
							Status: http.StatusOK,
							Body:   `{"system_env_json": {"VCAP_SERVICES": {} }}`,
						}})

					testServer, handler, repo = createAppRepo([]testnet.TestRequest{emptyEnvRequest})
					userEnv, err = repo.ReadEnv("some-cool-app-guid")
					Expect(err).ToNot(HaveOccurred())
					Expect(handler).To(HaveAllRequestsCalled())
				})

				It("returns an empty string", func() {
					Expect(len(userEnv.Environment)).To(Equal(0))
					Expect(len(userEnv.System["VCAP_SERVICES"].(map[string]interface{}))).To(Equal(0))
				})
			})
		})
	})

	Describe("restaging applications", func() {
		It("POSTs to the right URL", func() {
			appRestageRequest := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "POST",
				Path:   "/v2/apps/some-cool-app-guid/restage",
				Response: testnet.TestResponse{
					Status: http.StatusOK,
					Body:   "",
				},
			})

			ts, handler, repo := createAppRepo([]testnet.TestRequest{appRestageRequest})
			defer ts.Close()

			repo.CreateRestageRequest("some-cool-app-guid")
			Expect(handler).To(HaveAllRequestsCalled())
		})
	})

	Describe("updating applications", func() {
		It("makes the right request", func() {
			ts, handler, repo := createAppRepo([]testnet.TestRequest{updateApplicationRequest})
			defer ts.Close()

			app := models.Application{}
			app.GUID = "my-app-guid"
			app.Name = "my-cool-app"
			app.BuildpackURL = "buildpack-url"
			app.Command = "some-command"
			app.HealthCheckType = "none"
			app.Memory = 2048
			app.InstanceCount = 3
			app.Stack = &models.Stack{GUID: "some-stack-guid"}
			app.SpaceGUID = "some-space-guid"
			app.State = "started"
			app.DiskQuota = 512
			Expect(app.EnvironmentVars).To(BeNil())

			updatedApp, apiErr := repo.Update(app.GUID, app.ToParams())

			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
			Expect(updatedApp.Command).To(Equal("some-command"))
			Expect(updatedApp.DetectedStartCommand).To(Equal("detected command"))
			Expect(updatedApp.Name).To(Equal("my-cool-app"))
			Expect(updatedApp.GUID).To(Equal("my-cool-app-guid"))
		})

		It("sets environment variables", func() {
			request := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/apps/app1-guid",
				Matcher:  testnet.RequestBodyMatcher(`{"environment_json":{"DATABASE_URL":"mysql://example.com/my-db"}}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			ts, handler, repo := createAppRepo([]testnet.TestRequest{request})
			defer ts.Close()

			envParams := map[string]interface{}{"DATABASE_URL": "mysql://example.com/my-db"}
			params := models.AppParams{EnvironmentVars: &envParams}

			_, apiErr := repo.Update("app1-guid", params)

			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})

		It("can remove environment variables", func() {
			request := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/apps/app1-guid",
				Matcher:  testnet.RequestBodyMatcher(`{"environment_json":{}}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			ts, handler, repo := createAppRepo([]testnet.TestRequest{request})
			defer ts.Close()

			envParams := map[string]interface{}{}
			params := models.AppParams{EnvironmentVars: &envParams}

			_, apiErr := repo.Update("app1-guid", params)

			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})
	})

	It("deletes applications", func() {
		deleteApplicationRequest := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "DELETE",
			Path:     "/v2/apps/my-cool-app-guid?recursive=true",
			Response: testnet.TestResponse{Status: http.StatusOK, Body: ""},
		})

		ts, handler, repo := createAppRepo([]testnet.TestRequest{deleteApplicationRequest})
		defer ts.Close()

		apiErr := repo.Delete("my-cool-app-guid")
		Expect(handler).To(HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})
})

var appModelResponse = testnet.TestResponse{
	Status: http.StatusOK,
	Body: `
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
        "disk_quota": 512,
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
`}

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
        "disk_quota": 512,
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

var findAppRequest = apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
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

var updateApplicationResponse = `
{
    "metadata": {
        "guid": "my-cool-app-guid"
    },
    "entity": {
        "name": "my-cool-app",
				"command": "some-command",
				"detected_start_command": "detected command"
    }
}`

var updateApplicationRequest = apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "PUT",
	Path:   "/v2/apps/my-app-guid?inline-relations-depth=1",
	Matcher: testnet.RequestBodyMatcher(`{
		"name":"my-cool-app",
		"instances":3,
		"buildpack":"buildpack-url",
		"docker_image":"",
		"memory":2048,
		"health_check_type":"none",
		"health_check_http_endpoint":"",
		"disk_quota":512,
		"space_guid":"some-space-guid",
		"state":"STARTED",
		"stack_guid":"some-stack-guid",
		"command":"some-command"
	}`),
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body:   updateApplicationResponse},
})

func createAppRepo(requests []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo Repository) {
	ts, handler = testnet.NewServer(requests)
	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetAPIEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
	repo = NewCloudControllerRepository(configRepo, gateway)
	return
}
