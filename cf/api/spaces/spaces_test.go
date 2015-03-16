package spaces_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/api/spaces"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Space Repository", func() {
	It("lists all the spaces", func() {
		firstPageSpacesRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "GET",
			Path:   "/v2/organizations/my-org-guid/spaces?inline-relations-depth=1",
			Response: testnet.TestResponse{
				Status: http.StatusOK,
				Body: `
				{
					"next_url": "/v2/organizations/my-org-guid/spaces?inline-relations-depth=1&page=2",
					"resources": [
						{
							"metadata": {
								"guid": "acceptance-space-guid"
							},
							"entity": {
								"name": "acceptance",
		            "security_groups": [
		               {
		                  "metadata": {
		                     "guid": "4302b3b4-4afc-4f12-ae6d-ed1bb815551f"
		                  },
		                  "entity": {
		                     "name": "imma-security-group"
		                  }
		               }
                ]
							}
						}
					]
				}`}})

		secondPageSpacesRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "GET",
			Path:   "/v2/organizations/my-org-guid/spaces?inline-relations-depth=1&page=2",
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
								"name": "staging",
		            "security_groups": []
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
		Expect(spaces[0].SecurityGroups[0].Name).To(Equal("imma-security-group"))
		Expect(spaces[1].Guid).To(Equal("staging-space-guid"))
		Expect(apiErr).NotTo(HaveOccurred())
		Expect(handler).To(HaveAllRequestsCalled())
	})

	Describe("finding spaces by role", func() {
		It("returns the space with developer role", func() {

			pageSpacesRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/spaces?q=developer_guid:my-user-guid;q=name:my-space",
				Response: testnet.TestResponse{
					Status: http.StatusOK,
					Body: `
				{
					  "total_results": 1,
					  "total_pages": 1,
					  "prev_url": null,
					  "next_url": null,
					  "resources": [
					    {
					      "metadata": {
					        "guid": "my-org-guid",
					        "url": "/v2/spaces/my-space-guid",
					        "updated_at": null
					      },
					      "entity": {
					        "name": "my-space",
					        "organization_guid": "my-org-guid",
					        "space_quota_definition_guid": null
					      }
					    }
					  ]


				}`}})
			ts, handler, repo := createSpacesRepo(pageSpacesRequest)
			defer ts.Close()

			spaces := []models.Space{}
			apiErr := repo.GetSpaceRole(func(space models.Space) bool {
				spaces = append(spaces, space)
				return true
			}, "developer_guid")
			Expect(len(spaces)).To(Equal(1))
			Expect(spaces[0].Name).To(Equal("my-space"))
			Expect(apiErr).NotTo(HaveOccurred())
			Expect(handler).To(HaveAllRequestsCalled())
		})

		It("returns the space with manager role", func() {

			pageSpacesRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/users/my-user-guid/managed_spaces?q=organization_guid:my-org-guid;q=name:my-space",
				Response: testnet.TestResponse{
					Status: http.StatusOK,
					Body: `
				{
					  "total_results": 1,
					  "total_pages": 1,
					  "prev_url": null,
					  "next_url": null,
					  "resources": [
					    {
					      "metadata": {
					        "guid": "my-org-guid",
					        "url": "/v2/spaces/my-space-guid",
					        "created_at": "2015-03-03T20:46:20Z",
					        "updated_at": null
					      },
					      "entity": {
					        "name": "my-space",
					        "organization_guid": "my-org-guid",
					        "space_quota_definition_guid": null,
					        "organization_url": "/v2/organizations/7d080c6f-6b58-4b71-9079-f72839d9d0d3",
					        "developers_url": "/v2/spaces/2d837a89-60c0-4bd8-a6b9-7dfe7aea8db8/developers"
					        
					      }
					    }
					  ]
				}`}})
			ts, handler, repo := createSpacesRepo(pageSpacesRequest)
			defer ts.Close()

			spaces := []models.Space{}
			apiErr := repo.GetSpaceRole(func(space models.Space) bool {
				spaces = append(spaces, space)
				return true
			}, "managed_spaces")
			Expect(len(spaces)).To(Equal(1))
			Expect(spaces[0].Name).To(Equal("my-space"))
			Expect(apiErr).NotTo(HaveOccurred())
			Expect(handler).To(HaveAllRequestsCalled())
		})

		It("returns the apiErr while fetching space role", func() {

			pageSpacesRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/users/my-user-guid/managed_spaces?q=organization_guid:my-org-guid;q=name:my-space",
				Response: testnet.TestResponse{
					Status: http.StatusOK,
					Body: `
				{ 
					  "code": 10000,
					  "description": "Unknown request",
					  "error_code": "CF-NotFound"


				}`}})
			ts, handler, repo := createSpacesRepo(pageSpacesRequest)
			defer ts.Close()

			spaces := []models.Space{}
			apiErr := repo.GetSpaceRole(func(space models.Space) bool {
				spaces = append(spaces, space)
				return true
			}, "managed_spaces")
			Expect(len(spaces)).To(Equal(0))
			Expect(apiErr).To(HaveOccurred())
			Expect(handler).To(HaveAllRequestsCalled())
		})

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

	It("creates spaces without a space-quota", func() {
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

		space, apiErr := repo.Create("space-name", "my-org-guid", "")
		Expect(handler).To(HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
		Expect(space.Guid).To(Equal("space-guid"))
		Expect(space.SpaceQuotaGuid).To(Equal(""))
	})

	It("creates spaces with a space-quota", func() {
		request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:  "POST",
			Path:    "/v2/spaces",
			Matcher: testnet.RequestBodyMatcher(`{"name":"space-name","organization_guid":"my-org-guid","space_quota_definition_guid":"space-quota-guid"}`),
			Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
			{
				"metadata": {
					"guid": "space-guid"
				},
				"entity": {
					"name": "space-name",
					"space_quota_definition_guid":"space-quota-guid"
				}
			}`},
		})

		ts, handler, repo := createSpacesRepo(request)
		defer ts.Close()

		space, apiErr := repo.Create("space-name", "my-org-guid", "space-quota-guid")
		Expect(handler).To(HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
		Expect(space.Guid).To(Equal("space-guid"))
		Expect(space.SpaceQuotaGuid).To(Equal("space-quota-guid"))
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
		Expect(handler).To(HaveAllRequestsCalled())
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
		Expect(handler).To(HaveAllRequestsCalled())
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
	Expect(handler).To(HaveAllRequestsCalled())
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
	Expect(handler).To(HaveAllRequestsCalled())

	Expect(apiErr.(*errors.ModelNotFoundError)).NotTo(BeNil())
}

func createSpacesRepo(reqs ...testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo SpaceRepository) {
	ts, handler = testnet.NewServer(reqs)
	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway(configRepo, time.Now, &testterm.FakeUI{})
	repo = NewCloudControllerSpaceRepository(configRepo, gateway)
	return
}
