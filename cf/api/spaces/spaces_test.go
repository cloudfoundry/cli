package spaces_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testnet "code.cloudfoundry.org/cli/util/testhelpers/net"

	. "code.cloudfoundry.org/cli/cf/api/spaces"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Space Repository", func() {
	Describe("ListSpaces", func() {
		var (
			ccServer *ghttp.Server
			repo     CloudControllerSpaceRepository
		)

		BeforeEach(func() {
			ccServer = ghttp.NewServer()
			configRepo := testconfig.NewRepositoryWithDefaults()
			configRepo.SetAPIEndpoint(ccServer.URL())
			gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
			repo = NewCloudControllerSpaceRepository(configRepo, gateway)
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/organizations/my-org-guid/spaces", "order-by=name&inline-relations-depth=1"),
					ghttp.VerifyHeader(http.Header{
						"accept": []string{"application/json"},
					}),
					ghttp.RespondWith(http.StatusOK, `{
						"total_results": 3,
						"total_pages": 2,
						"prev_url": null,
						"next_url": "/v2/organizations/my-org-guid/spaces?order-by=name&page=2&inline-relations-depth=1",
						"resources": [
							{
								"metadata": { "guid": "space3-guid" },
								"entity": {
								  "name": "Alpha",
								  "allow_ssh": true,
		              "security_groups": [
		                {
		                  "metadata": { "guid": "4302b3b4-4afc-4f12-ae6d-ed1bb815551f" },
		                  "entity": { "name": "imma-security-group" }
		                }
		              ]
		            }
							},
							{
								"metadata": { "guid": "space2-guid" },
								"entity": { "name": "Beta" }
							}
						]
					}`),
				),
			)

			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/organizations/my-org-guid/spaces", "order-by=name&page=2&inline-relations-depth=1"),
					ghttp.VerifyHeader(http.Header{
						"accept": []string{"application/json"},
					}),
					ghttp.RespondWith(http.StatusOK, `{
						"total_results": 3,
						"total_pages": 2,
						"prev_url": null,
						"next_url": null,
						"resources": [
							{
								"metadata": { "guid": "space1-guid" },
								"entity": { "name": "Gamma" }
							}
						]
					}`),
				),
			)
		})

		AfterEach(func() {
			ccServer.Close()
		})

		It("lists all the spaces", func() {
			spaces := []models.Space{}
			apiErr := repo.ListSpaces(func(space models.Space) bool {
				spaces = append(spaces, space)
				return true
			})

			Expect(apiErr).NotTo(HaveOccurred())
			Expect(len(spaces)).To(Equal(3))
			Expect(spaces[0].GUID).To(Equal("space3-guid"))
			Expect(spaces[0].AllowSSH).To(BeTrue())
			Expect(spaces[0].SecurityGroups[0].Name).To(Equal("imma-security-group"))
			Expect(spaces[0].Name).To(Equal("Alpha"))

			Expect(spaces[1].GUID).To(Equal("space2-guid"))
			Expect(spaces[1].Name).To(Equal("Beta"))

			Expect(spaces[2].GUID).To(Equal("space1-guid"))
			Expect(spaces[2].Name).To(Equal("Gamma"))
		})
	})

	Describe("ListSpacesFromOrg", func() {
		var (
			ccServer *ghttp.Server
			repo     CloudControllerSpaceRepository
		)

		BeforeEach(func() {
			ccServer = ghttp.NewServer()
			configRepo := testconfig.NewRepositoryWithDefaults()
			configRepo.SetAPIEndpoint(ccServer.URL())
			configRepo.SetOrganizationFields(models.OrganizationFields{})
			gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
			repo = NewCloudControllerSpaceRepository(configRepo, gateway)
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/organizations/my-org-guid/spaces", "order-by=name&inline-relations-depth=1"),
					ghttp.VerifyHeader(http.Header{
						"accept": []string{"application/json"},
					}),
					ghttp.RespondWith(http.StatusOK, `{
						"total_results": 3,
						"total_pages": 2,
						"prev_url": null,
						"next_url": "/v2/organizations/my-org-guid/spaces?order-by=name&page=2&inline-relations-depth=1",
						"resources": [
							{
								"metadata": { "guid": "space3-guid" },
								"entity": {
								  "name": "Alpha",
								  "allow_ssh": true,
		              "security_groups": [
		                {
		                  "metadata": { "guid": "4302b3b4-4afc-4f12-ae6d-ed1bb815551f" },
		                  "entity": { "name": "imma-security-group" }
		                }
		              ]
		            }
							},
							{
								"metadata": { "guid": "space2-guid" },
								"entity": { "name": "Beta" }
							}
						]
					}`),
				),
			)

			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/organizations/my-org-guid/spaces", "order-by=name&page=2&inline-relations-depth=1"),
					ghttp.VerifyHeader(http.Header{
						"accept": []string{"application/json"},
					}),
					ghttp.RespondWith(http.StatusOK, `{
						"total_results": 3,
						"total_pages": 2,
						"prev_url": null,
						"next_url": null,
						"resources": [
							{
								"metadata": { "guid": "space1-guid" },
								"entity": { "name": "Gamma" }
							}
						]
					}`),
				),
			)
		})

		AfterEach(func() {
			ccServer.Close()
		})

		It("lists all the spaces", func() {
			spaces := []models.Space{}
			apiErr := repo.ListSpacesFromOrg("my-org-guid", func(space models.Space) bool {
				spaces = append(spaces, space)
				return true
			})

			Expect(apiErr).NotTo(HaveOccurred())
			Expect(len(spaces)).To(Equal(3))
			Expect(spaces[0].GUID).To(Equal("space3-guid"))
			Expect(spaces[0].AllowSSH).To(BeTrue())
			Expect(spaces[0].SecurityGroups[0].Name).To(Equal("imma-security-group"))
			Expect(spaces[0].Name).To(Equal("Alpha"))

			Expect(spaces[1].GUID).To(Equal("space2-guid"))
			Expect(spaces[1].Name).To(Equal("Beta"))

			Expect(spaces[2].GUID).To(Equal("space1-guid"))
			Expect(spaces[2].Name).To(Equal("Gamma"))
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
		request := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
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
		Expect(space.GUID).To(Equal("space-guid"))
		Expect(space.SpaceQuotaGUID).To(Equal(""))
	})

	It("creates spaces with a space-quota", func() {
		request := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
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
		Expect(space.GUID).To(Equal("space-guid"))
		Expect(space.SpaceQuotaGUID).To(Equal("space-quota-guid"))
	})

	It("sets allow_ssh field", func() {
		request := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "PUT",
			Path:     "/v2/spaces/my-space-guid",
			Matcher:  testnet.RequestBodyMatcher(`{"allow_ssh":true}`),
			Response: testnet.TestResponse{Status: http.StatusCreated},
		})

		ts, handler, repo := createSpacesRepo(request)
		defer ts.Close()

		apiErr := repo.SetAllowSSH("my-space-guid", true)
		Expect(handler).To(HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})

	It("renames spaces", func() {
		request := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
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
		request := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
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

func testSpacesFindByNameWithOrg(orgGUID string, findByName func(SpaceRepository, string) (models.Space, error)) {
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
	request := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     fmt.Sprintf("/v2/organizations/%s/spaces?q=name%%3Aspace1", orgGUID),
		Response: findSpaceByNameResponse,
	})

	ts, handler, repo := createSpacesRepo(request)
	defer ts.Close()

	space, apiErr := findByName(repo, "Space1")
	Expect(handler).To(HaveAllRequestsCalled())
	Expect(apiErr).NotTo(HaveOccurred())
	Expect(space.Name).To(Equal("Space1"))
	Expect(space.GUID).To(Equal("space1-guid"))

	Expect(space.Organization.GUID).To(Equal("org1-guid"))

	Expect(len(space.Applications)).To(Equal(2))
	Expect(space.Applications[0].GUID).To(Equal("app1-guid"))
	Expect(space.Applications[1].GUID).To(Equal("app2-guid"))

	Expect(len(space.Domains)).To(Equal(1))
	Expect(space.Domains[0].GUID).To(Equal("domain1-guid"))

	Expect(len(space.ServiceInstances)).To(Equal(1))
	Expect(space.ServiceInstances[0].GUID).To(Equal("service1-guid"))

	Expect(apiErr).NotTo(HaveOccurred())
	return
}

func testSpacesDidNotFindByNameWithOrg(orgGUID string, findByName func(SpaceRepository, string) (models.Space, error)) {
	request := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   fmt.Sprintf("/v2/organizations/%s/spaces?q=name%%3Aspace1", orgGUID),
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
	configRepo.SetAPIEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
	repo = NewCloudControllerSpaceRepository(configRepo, gateway)
	return
}
