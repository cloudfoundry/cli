package api_test

import (
	. "github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("Buildpacks repo", func() {
	var (
		ts      *httptest.Server
		handler *testnet.TestHandler
		config  configuration.ReadWriter
		repo    BuildpackRepository
	)

	BeforeEach(func() {
		config = testconfig.NewRepositoryWithDefaults()
		gateway := net.NewCloudControllerGateway(config)
		repo = NewCloudControllerBuildpackRepository(config, gateway)
	})

	AfterEach(func() {
		ts.Close()
	})

	var setupTestServer = func(requests ...testnet.TestRequest) {
		ts, handler = testnet.NewServer(requests)
		config.SetApiEndpoint(ts.URL)
	}

	It("lists buildpacks", func() {
		setupTestServer(
			testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/buildpacks",
				Response: testnet.TestResponse{
					Status: http.StatusOK,
					Body: `{
					  "next_url": "/v2/buildpacks?page=2",
					  "resources": [
						{
						  "metadata": {
							"guid": "buildpack1-guid"
						  },
						  "entity": {
							"name": "Buildpack1",
							"position" : 1,
							"filename" : "firstbp.zip"
						  }
						}
					  ]
					}`}}),
			testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/buildpacks?page=2",
				Response: testnet.TestResponse{
					Status: http.StatusOK,
					Body: `{
					  "resources": [
						{
						  "metadata": {
							"guid": "buildpack2-guid"
						  },
						  "entity": {
							"name": "Buildpack2",
							"position" : 2
						  }
						}
					  ]
					}`},
			}))

		buildpacks := []models.Buildpack{}
		err := repo.ListBuildpacks(func(b models.Buildpack) bool {
			buildpacks = append(buildpacks, b)
			return true
		})

		one := 1
		two := 2
		Expect(buildpacks).To(Equal([]models.Buildpack{
			{
				Guid:     "buildpack1-guid",
				Name:     "Buildpack1",
				Position: &one,
				Filename: "firstbp.zip",
			},
			{
				Guid:     "buildpack2-guid",
				Name:     "Buildpack2",
				Position: &two,
			},
		}))
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("finding buildpacks by name", func() {
		It("returns the buildpack with that name", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/buildpacks?q=name%3ABuildpack1",
				Response: testnet.TestResponse{
					Status: http.StatusOK,
					Body: `{"resources": [
					  {
						  "metadata": {
							  "guid": "buildpack1-guid"
						  },
						  "entity": {
							  "name": "Buildpack1",
							  "position": 10
						  }
					  }
					  ]
				  }`}}))

			buildpack, apiErr := repo.FindByName("Buildpack1")

			Expect(handler).To(testnet.HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())

			Expect(buildpack.Name).To(Equal("Buildpack1"))
			Expect(buildpack.Guid).To(Equal("buildpack1-guid"))
			Expect(*buildpack.Position).To(Equal(10))
		})

		It("returns a ModelNotFoundError when the buildpack is not found", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/buildpacks?q=name%3ABuildpack1",
				Response: testnet.TestResponse{
					Status: http.StatusOK,
					Body:   `{"resources": []}`,
				},
			}))

			_, apiErr := repo.FindByName("Buildpack1")
			Expect(handler).To(testnet.HaveAllRequestsCalled())
			Expect(apiErr.(*errors.ModelNotFoundError)).NotTo(BeNil())
		})
	})

	Describe("creating buildpacks", func() {
		It("returns an error when the buildpack has an invalid name", func() {
			setupTestServer(testnet.TestRequest{
				Method: "POST",
				Path:   "/v2/buildpacks",
				Response: testnet.TestResponse{
					Status: http.StatusBadRequest,
					Body: `{
					"code":290003,
					"description":"Buildpack is invalid: [\"name name can only contain alphanumeric characters\"]",
					"error_code":"CF-BuildpackInvalid"
				}`,
				}})

			one := 1
			createdBuildpack, apiErr := repo.Create("name with space", &one, nil, nil)
			Expect(apiErr).To(HaveOccurred())
			Expect(createdBuildpack).To(Equal(models.Buildpack{}))
			Expect(apiErr.(errors.HttpError).ErrorCode()).To(Equal("290003"))
			Expect(apiErr.Error()).To(ContainSubstring("Buildpack is invalid"))
		})

		It("sets the position flag when creating a buildpack", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:  "POST",
				Path:    "/v2/buildpacks",
				Matcher: testnet.RequestBodyMatcher(`{"name":"my-cool-buildpack","position":999}`),
				Response: testnet.TestResponse{
					Status: http.StatusCreated,
					Body: `{
					"metadata": {
						"guid": "my-cool-buildpack-guid"
					},
					"entity": {
						"name": "my-cool-buildpack",
						"position":999
					}
				}`},
			}))

			position := 999
			created, apiErr := repo.Create("my-cool-buildpack", &position, nil, nil)

			Expect(handler).To(testnet.HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())

			Expect(created.Guid).NotTo(BeNil())
			Expect("my-cool-buildpack").To(Equal(created.Name))
			Expect(999).To(Equal(*created.Position))
		})

		It("sets the enabled flag when creating a buildpack", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:  "POST",
				Path:    "/v2/buildpacks",
				Matcher: testnet.RequestBodyMatcher(`{"name":"my-cool-buildpack","position":999, "enabled":true}`),
				Response: testnet.TestResponse{
					Status: http.StatusCreated,
					Body: `{
					"metadata": {
						"guid": "my-cool-buildpack-guid"
					},
					"entity": {
						"name": "my-cool-buildpack",
						"position":999,
						"enabled":true
					}
				}`},
			}))

			position := 999
			enabled := true
			created, apiErr := repo.Create("my-cool-buildpack", &position, &enabled, nil)

			Expect(handler).To(testnet.HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())

			Expect(created.Guid).NotTo(BeNil())
			Expect(created.Name).To(Equal("my-cool-buildpack"))
			Expect(999).To(Equal(*created.Position))
		})
	})

	It("deletes buildpacks", func() {
		setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "DELETE",
			Path:   "/v2/buildpacks/my-cool-buildpack-guid",
			Response: testnet.TestResponse{
				Status: http.StatusNoContent,
			}}))

		err := repo.Delete("my-cool-buildpack-guid")

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("updating buildpacks", func() {
		It("updates a buildpack's name, position and enabled flag", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:  "PUT",
				Path:    "/v2/buildpacks/my-cool-buildpack-guid",
				Matcher: testnet.RequestBodyMatcher(`{"name":"my-cool-buildpack","position":555,"enabled":false}`),
				Response: testnet.TestResponse{
					Status: http.StatusCreated,
					Body: `{
						"metadata": {
							"guid": "my-cool-buildpack-guid"
						},
						"entity": {
							"name": "my-cool-buildpack",
							"position":555,
							"enabled":false
						}
					}`},
			}))

			position := 555
			enabled := false
			buildpack := models.Buildpack{
				Name:     "my-cool-buildpack",
				Guid:     "my-cool-buildpack-guid",
				Position: &position,
				Enabled:  &enabled,
			}

			updated, err := repo.Update(buildpack)

			Expect(handler).To(testnet.HaveAllRequestsCalled())
			Expect(err).NotTo(HaveOccurred())
			Expect(updated).To(Equal(buildpack))
		})

		It("sets the locked attribute on the buildpack", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:  "PUT",
				Path:    "/v2/buildpacks/my-cool-buildpack-guid",
				Matcher: testnet.RequestBodyMatcher(`{"name":"my-cool-buildpack","locked":true}`),
				Response: testnet.TestResponse{
					Status: http.StatusCreated,
					Body: `{

						"metadata": {
							"guid": "my-cool-buildpack-guid"
						},
						"entity": {
							"name": "my-cool-buildpack",
							"position":123,
							"locked": true
						}
					}`},
			}))

			locked := true

			buildpack := models.Buildpack{
				Name:   "my-cool-buildpack",
				Guid:   "my-cool-buildpack-guid",
				Locked: &locked,
			}

			updated, err := repo.Update(buildpack)

			Expect(handler).To(testnet.HaveAllRequestsCalled())
			Expect(err).NotTo(HaveOccurred())

			position := 123
			Expect(updated).To(Equal(models.Buildpack{
				Name:     "my-cool-buildpack",
				Guid:     "my-cool-buildpack-guid",
				Position: &position,
				Locked:   &locked,
			}))
		})
	})
})
