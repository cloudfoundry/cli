package api_test

import (
	"fmt"
	. "github.com/cloudfoundry/cli/cf/api"
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

var _ = Describe("Space Repository", func() {
	It("lists all the spaces", func() {
		firstPageSpacesRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "GET",
			Path:   "/v2/organizations/my-org-guid/spaces",
			Response: testnet.TestResponse{
				Status: http.StatusOK,
				Body: `
				{
					"next_url": "/v2/organizations/my-org-guid/spaces?page=2",
					"resources": [
						{
							"metadata": {
								"guid": "acceptance-space-guid"
							},
							"entity": {
								"name": "acceptance"
							}
						}
					]
				}`}})

		secondPageSpacesRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "GET",
			Path:   "/v2/organizations/my-org-guid/spaces?page=2",
			Response: testnet.TestResponse{
				Status: http.StatusOK,
				Body: `
				{
					"resources": [
						{
							"metadata": {
								"guid": "staging-space-guid"
							},
							"entity": {
								"name": "staging"
							}
						}
					]
				}`}})

		ts, handler, repo := createSpacesRepo(firstPageSpacesRequest, secondPageSpacesRequest)
		defer ts.Close()

		spaces := []models.Space{}
		apiErr := repo.ListSpaces(func(space models.Space) bool {
			spaces = append(spaces, space)
			return true
		})

		Expect(len(spaces)).To(Equal(2))
		Expect(spaces[0].Guid).To(Equal("acceptance-space-guid"))
		Expect(spaces[1].Guid).To(Equal("staging-space-guid"))
		Expect(apiErr).NotTo(HaveOccurred())
		Expect(handler).To(testnet.HaveAllRequestsCalled())
	})

	Describe("finding spaces by name", func() {
		It("returns the space", func() {
			testSpacesFindByNameWithOrg("my-org-guid",
				func(repo SpaceRepository, spaceName string) (models.Space, error) {
					return repo.FindByName(spaceName)
				},
			)
		})

		It("can find spaces in a particular org", func() {
			testSpacesFindByNameWithOrg("another-org-guid",
				func(repo SpaceRepository, spaceName string) (models.Space, error) {
					return repo.FindByNameInOrg(spaceName, "another-org-guid")
				},
			)
		})

		It("returns a 'not found' response when the space doesn't exist", func() {
			testSpacesDidNotFindByNameWithOrg("my-org-guid",
				func(repo SpaceRepository, spaceName string) (models.Space, error) {
					return repo.FindByName(spaceName)
				},
			)
		})

		It("returns a 'not found' response when the space doesn't exist in the given org", func() {
			testSpacesDidNotFindByNameWithOrg("another-org-guid",
				func(repo SpaceRepository, spaceName string) (models.Space, error) {
					return repo.FindByNameInOrg(spaceName, "another-org-guid")
				},
			)
		})
	})

	It("creates spaces", func() {
		request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:  "POST",
			Path:    "/v2/spaces",
			Matcher: testnet.RequestBodyMatcher(`{"name":"space-name","organization_guid":"my-org-guid"}`),
			Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
			{
				"metadata": {
					"guid": "space-guid"
				},
				"entity": {
					"name": "space-name"
				}
			}`},
		})

		ts, handler, repo := createSpacesRepo(request)
		defer ts.Close()

		space, apiErr := repo.Create("space-name", "my-org-guid")
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
		Expect(space.Guid).To(Equal("space-guid"))
	})

	It("renames spaces", func() {
		request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "PUT",
			Path:     "/v2/spaces/my-space-guid",
			Matcher:  testnet.RequestBodyMatcher(`{"name":"new-space-name"}`),
			Response: testnet.TestResponse{Status: http.StatusCreated},
		})

		ts, handler, repo := createSpacesRepo(request)
		defer ts.Close()

		apiErr := repo.Rename("my-space-guid", "new-space-name")
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})

	It("deletes spaces", func() {
		request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "DELETE",
			Path:     "/v2/spaces/my-space-guid?recursive=true",
			Response: testnet.TestResponse{Status: http.StatusOK},
		})

		ts, handler, repo := createSpacesRepo(request)
		defer ts.Close()

		apiErr := repo.Delete("my-space-guid")
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})
})

func testSpacesFindByNameWithOrg(orgGuid string, findByName func(SpaceRepository, string) (models.Space, error)) {
	findSpaceByNameResponse := testnet.TestResponse{
		Status: http.StatusOK,
		Body: `
{
  "resources": [
    {
      "metadata": {
        "guid": "space1-guid"
      },
      "entity": {
        "name": "Space1",
        "organization_guid": "org1-guid",
        "organization": {
          "metadata": {
            "guid": "org1-guid"
          },
          "entity": {
            "name": "Org1"
          }
        },
        "apps": [
          {
            "metadata": {
              "guid": "app1-guid"
            },
            "entity": {
              "name": "app1"
            }
          },
          {
            "metadata": {
              "guid": "app2-guid"
            },
            "entity": {
              "name": "app2"
            }
          }
        ],
        "domains": [
          {
            "metadata": {
              "guid": "domain1-guid"
            },
            "entity": {
              "name": "domain1"
            }
          }
        ],
        "service_instances": [
          {
			"metadata": {
              "guid": "service1-guid"
            },
            "entity": {
              "name": "service1"
            }
          }
        ]
      }
    }
  ]
}`}
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     fmt.Sprintf("/v2/organizations/%s/spaces?q=name%%3Aspace1&inline-relations-depth=1", orgGuid),
		Response: findSpaceByNameResponse,
	})

	ts, handler, repo := createSpacesRepo(request)
	defer ts.Close()

	space, apiErr := findByName(repo, "Space1")
	Expect(handler).To(testnet.HaveAllRequestsCalled())
	Expect(apiErr).NotTo(HaveOccurred())
	Expect(space.Name).To(Equal("Space1"))
	Expect(space.Guid).To(Equal("space1-guid"))

	Expect(space.Organization.Guid).To(Equal("org1-guid"))

	Expect(len(space.Applications)).To(Equal(2))
	Expect(space.Applications[0].Guid).To(Equal("app1-guid"))
	Expect(space.Applications[1].Guid).To(Equal("app2-guid"))

	Expect(len(space.Domains)).To(Equal(1))
	Expect(space.Domains[0].Guid).To(Equal("domain1-guid"))

	Expect(len(space.ServiceInstances)).To(Equal(1))
	Expect(space.ServiceInstances[0].Guid).To(Equal("service1-guid"))

	Expect(apiErr).NotTo(HaveOccurred())
	return
}

func testSpacesDidNotFindByNameWithOrg(orgGuid string, findByName func(SpaceRepository, string) (models.Space, error)) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   fmt.Sprintf("/v2/organizations/%s/spaces?q=name%%3Aspace1&inline-relations-depth=1", orgGuid),
		Response: testnet.TestResponse{
			Status: http.StatusOK,
			Body:   ` { "resources": [ ] }`,
		},
	})

	ts, handler, repo := createSpacesRepo(request)
	defer ts.Close()

	_, apiErr := findByName(repo, "Space1")
	Expect(handler).To(testnet.HaveAllRequestsCalled())

	Expect(apiErr.(*errors.ModelNotFoundError)).NotTo(BeNil())
}

func createSpacesRepo(reqs ...testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo SpaceRepository) {
	ts, handler = testnet.NewServer(reqs)
	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway(configRepo)
	repo = NewCloudControllerSpaceRepository(configRepo, gateway)
	return
}
