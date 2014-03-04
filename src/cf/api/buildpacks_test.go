package api_test

import (
	. "cf/api"
	"cf/errors"
	"cf/models"
	"cf/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
)

var _ = Describe("Buildpacks repo", func() {
	It("lists buildpacks", func() {
		firstRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
					}`},
		})

		secondRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
		})

		ts, handler, repo := createBuildpackRepo(firstRequest, secondRequest)
		defer ts.Close()

		one := 1
		two := 2
		expectedBuildpacks := []models.Buildpack{
			models.Buildpack{
				Guid:     "buildpack1-guid",
				Name:     "Buildpack1",
				Position: &one,
				Filename: "firstbp.zip",
			},
			models.Buildpack{
				Guid:     "buildpack2-guid",
				Name:     "Buildpack2",
				Position: &two,
			},
		}

		buildpacks := []models.Buildpack{}
		apiErr := repo.ListBuildpacks(func(b models.Buildpack) bool {
			buildpacks = append(buildpacks, b)
			return true
		})

		Expect(buildpacks).To(Equal(expectedBuildpacks))
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})

	It("TestBuildpacksFindByName", func() {
		req := testapi.NewCloudControllerTestRequest(findBuildpackRequest)

		ts, handler, repo := createBuildpackRepo(req)
		defer ts.Close()
		existingBuildpack := models.Buildpack{}
		existingBuildpack.Guid = "buildpack1-guid"
		existingBuildpack.Name = "Buildpack1"

		buildpack, apiErr := repo.FindByName("Buildpack1")

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())

		Expect(buildpack.Name).To(Equal(existingBuildpack.Name))
		Expect(buildpack.Guid).To(Equal(existingBuildpack.Guid))
		Expect(*buildpack.Position).To(Equal(10))
	})

	It("TestFindByNameWhenBuildpackIsNotFound", func() {
		req := testapi.NewCloudControllerTestRequest(findBuildpackRequest)
		req.Response = testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`}

		ts, handler, repo := createBuildpackRepo(req)
		defer ts.Close()

		_, apiErr := repo.FindByName("Buildpack1")
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr.(errors.ModelNotFoundError)).NotTo(BeNil())
	})

	It("TestBuildpackCreateRejectsImproperNames", func() {
		badRequest := testnet.TestRequest{
			Method: "POST",
			Path:   "/v2/buildpacks",
			Response: testnet.TestResponse{
				Status: http.StatusBadRequest,
				Body: `{
				"code":290003,
				"description":"Buildpack is invalid: [\"name name can only contain alphanumeric characters\"]",
				"error_code":"CF-BuildpackInvalid"
			}`,
			}}

		ts, _, repo := createBuildpackRepo(badRequest)
		defer ts.Close()
		one := 1
		createdBuildpack, apiErr := repo.Create("name with space", &one, nil, nil)
		Expect(apiErr).To(HaveOccurred())
		Expect(createdBuildpack).To(Equal(models.Buildpack{}))
		Expect(apiErr.ErrorCode()).To(Equal("290003"))
		Expect(apiErr.Error()).To(ContainSubstring("Buildpack is invalid"))
	})

	It("TestCreateBuildpackWithPosition", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
		})

		ts, handler, repo := createBuildpackRepo(req)
		defer ts.Close()

		position := 999
		created, apiErr := repo.Create("my-cool-buildpack", &position, nil, nil)

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())

		Expect(created.Guid).NotTo(BeNil())
		Expect("my-cool-buildpack").To(Equal(created.Name))
		Expect(999).To(Equal(*created.Position))
	})

	It("TestCreateBuildpackEnabled", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
		})

		ts, handler, repo := createBuildpackRepo(req)
		defer ts.Close()

		position := 999
		enabled := true
		created, apiErr := repo.Create("my-cool-buildpack", &position, &enabled, nil)

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())

		Expect(created.Guid).NotTo(BeNil())
		Expect("my-cool-buildpack").To(Equal(created.Name))
		Expect(999).To(Equal(*created.Position))
	})

	It("TestDeleteBuildpack", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "DELETE",
			Path:   "/v2/buildpacks/my-cool-buildpack-guid",
			Response: testnet.TestResponse{
				Status: http.StatusNoContent,
			}})

		ts, handler, repo := createBuildpackRepo(req)
		defer ts.Close()

		apiErr := repo.Delete("my-cool-buildpack-guid")

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})

	It("TestUpdateBuildpack", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
		})

		ts, handler, repo := createBuildpackRepo(req)
		defer ts.Close()

		position := 555
		enabled := false
		buildpack := models.Buildpack{}
		buildpack.Name = "my-cool-buildpack"
		buildpack.Guid = "my-cool-buildpack-guid"
		buildpack.Position = &position
		buildpack.Enabled = &enabled
		updated, apiErr := repo.Update(buildpack)

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())

		Expect(buildpack).To(Equal(updated))
	})

	It("TestLockBuildpack", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
		})

		ts, handler, repo := createBuildpackRepo(req)
		defer ts.Close()

		position := 123
		locked := true

		buildpack := models.Buildpack{}
		buildpack.Name = "my-cool-buildpack"
		buildpack.Guid = "my-cool-buildpack-guid"
		buildpack.Locked = &locked

		expectedBuildpack := models.Buildpack{}
		expectedBuildpack.Name = "my-cool-buildpack"
		expectedBuildpack.Guid = "my-cool-buildpack-guid"
		expectedBuildpack.Position = &position
		expectedBuildpack.Locked = &locked

		updated, apiErr := repo.Update(buildpack)

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())

		Expect(expectedBuildpack).To(Equal(updated))
	})
})

var singleBuildpackResponse = testnet.TestResponse{
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
	  }`}

var findBuildpackRequest = testnet.TestRequest{
	Method:   "GET",
	Path:     "/v2/buildpacks?q=name%3ABuildpack1",
	Response: singleBuildpackResponse,
}

func createBuildpackRepo(requests ...testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo BuildpackRepository) {
	ts, handler = testnet.NewServer(requests)
	config := testconfig.NewRepositoryWithDefaults()
	config.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerBuildpackRepository(config, gateway)
	return
}
