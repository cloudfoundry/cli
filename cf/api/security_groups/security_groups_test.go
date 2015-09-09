package security_groups_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/api/security_groups"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("app security group api", func() {
	var (
		testServer  *httptest.Server
		testHandler *testnet.TestHandler
		configRepo  core_config.ReadWriter
		repo        SecurityGroupRepo
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		gateway := net.NewCloudControllerGateway(configRepo, time.Now, &testterm.FakeUI{})
		repo = NewSecurityGroupRepo(configRepo, gateway)
	})

	AfterEach(func() {
		testServer.Close()
	})

	setupTestServer := func(reqs ...testnet.TestRequest) {
		testServer, testHandler = testnet.NewServer(reqs)
		configRepo.SetApiEndpoint(testServer.URL)
	}

	Describe(".Create", func() {
		It("can create an app security group, given some attributes", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "POST",
				Path:   "/v2/security_groups",
				// FIXME: this matcher depend on the order of the key/value pairs in the map
				Matcher: testnet.RequestBodyMatcher(`{
					"name": "mygroup",
					"rules": [{"my-house": "my-rules"}]
				}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})
			setupTestServer(req)

			err := repo.Create(
				"mygroup",
				[]map[string]interface{}{{"my-house": "my-rules"}},
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(testHandler).To(HaveAllRequestsCalled())
		})
	})

	Describe(".Read", func() {
		It("returns the app security group with the given name", func() {

			req1 := testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/security_groups?q=name:the-name",
				Response: testnet.TestResponse{
					Status: http.StatusOK,
					Body: `
{
   "resources": [
      {
         "metadata": {
            "guid": "the-group-guid"
         },
         "entity": {
            "name": "the-name",
            "rules": [{"key": "value"}],
						"spaces_url": "/v2/security_groups/guid-id/spaces"
         }
      }
   ]
}
					`,
				},
			}

			req2 := testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/security_groups/guid-id/spaces?inline-relations-depth=1",
				Response: testnet.TestResponse{
					Status: http.StatusOK,
					Body: `
{
   "resources": [
      {
				"metadata":{
					"guid": "my-space-guid"
				},
				"entity": {
					 "name": "my-space",
					 "organization": {
							"metadata": {
								 "guid": "my-org-guid"
							},
							"entity": {
								 "name": "my-org"
							}
					 }
				}
      },
      {
				"metadata":{
					"guid": "my-space-guid2"
				},
				"entity": {
					 "name": "my-space2",
					 "organization": {
							"metadata": {
								 "guid": "my-org-guid2"
							},
							"entity": {
								 "name": "my-org2"
							}
					 }
				}
      }
   ]
}
					`,
				},
			}

			setupTestServer(testapi.NewCloudControllerTestRequest(req1), testapi.NewCloudControllerTestRequest(req2))

			group, err := repo.Read("the-name")

			Expect(err).ToNot(HaveOccurred())
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(group).To(Equal(models.SecurityGroup{
				SecurityGroupFields: models.SecurityGroupFields{
					Name:     "the-name",
					Guid:     "the-group-guid",
					SpaceUrl: "/v2/security_groups/guid-id/spaces",
					Rules:    []map[string]interface{}{{"key": "value"}},
				},
				Spaces: []models.Space{
					{
						SpaceFields:  models.SpaceFields{Guid: "my-space-guid", Name: "my-space"},
						Organization: models.OrganizationFields{Guid: "my-org-guid", Name: "my-org"},
					},
					{
						SpaceFields:  models.SpaceFields{Guid: "my-space-guid2", Name: "my-space2"},
						Organization: models.OrganizationFields{Guid: "my-org-guid2", Name: "my-org2"},
					},
				},
			}))
		})

		It("returns a ModelNotFound error if the security group cannot be found", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/security_groups?q=name:the-name",
				Response: testnet.TestResponse{
					Status: http.StatusOK,
					Body:   `{"resources": []}`,
				},
			}))

			_, err := repo.Read("the-name")

			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(errors.NewModelNotFoundError("model-type", "description")))
		})
	})

	Describe(".Delete", func() {
		It("deletes the security group", func() {
			securityGroupGuid := "the-security-group-guid"
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "DELETE",
				Path:   "/v2/security_groups/" + securityGroupGuid,
				Response: testnet.TestResponse{
					Status: http.StatusNoContent,
				},
			}))

			err := repo.Delete(securityGroupGuid)

			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe(".FindAll", func() {
		It("returns all the security groups", func() {
			setupTestServer(
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "GET",
					Path:   "/v2/security_groups",
					Response: testnet.TestResponse{
						Status: http.StatusOK,
						Body:   firstListItem(),
					},
				}),
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "GET",
					Path:   "/v2/security_groups?page=2",
					Response: testnet.TestResponse{
						Status: http.StatusOK,
						Body:   secondListItem(),
					},
				}),
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "GET",
					Path:   "/v2/security_groups/cd186158-b356-474d-9861-724f34f48502/spaces?inline-relations-depth=1",
					Response: testnet.TestResponse{
						Status: http.StatusOK,
						Body:   spacesItems(),
					},
				}),
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "GET",
					Path:   "/v2/security_groups/d3374b62-7eac-4823-afbd-460d2bf44c67/spaces?inline-relations-depth=1",
					Response: testnet.TestResponse{
						Status: http.StatusOK,
						Body:   spacesItems(),
					},
				}),
			)

			groups, err := repo.FindAll()

			Expect(err).ToNot(HaveOccurred())
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(groups[0]).To(Equal(models.SecurityGroup{
				SecurityGroupFields: models.SecurityGroupFields{
					Name:     "name-71",
					Guid:     "cd186158-b356-474d-9861-724f34f48502",
					Rules:    []map[string]interface{}{{"protocol": "udp"}},
					SpaceUrl: "/v2/security_groups/cd186158-b356-474d-9861-724f34f48502/spaces",
				},
				Spaces: []models.Space{
					{
						SpaceFields:  models.SpaceFields{Guid: "my-space-guid", Name: "my-space"},
						Organization: models.OrganizationFields{Guid: "my-org-guid", Name: "my-org"},
					},
				},
			}))
			Expect(groups[1]).To(Equal(models.SecurityGroup{
				SecurityGroupFields: models.SecurityGroupFields{
					Name:     "name-72",
					Guid:     "d3374b62-7eac-4823-afbd-460d2bf44c67",
					Rules:    []map[string]interface{}{{"destination": "198.41.191.47/1"}},
					SpaceUrl: "/v2/security_groups/d3374b62-7eac-4823-afbd-460d2bf44c67/spaces",
				},
				Spaces: []models.Space{
					{
						SpaceFields:  models.SpaceFields{Guid: "my-space-guid", Name: "my-space"},
						Organization: models.OrganizationFields{Guid: "my-org-guid", Name: "my-org"},
					},
				},
			}))
		})
	})
})

func firstListItem() string {
	return `{
  "next_url": "/v2/security_groups?inline-relations-depth=2&page=2",
  "resources": [
    {
      "metadata": {
        "guid": "cd186158-b356-474d-9861-724f34f48502",
        "url": "/v2/security_groups/cd186158-b356-474d-9861-724f34f48502",
        "created_at": "2014-06-23T22:55:30+00:00",
        "updated_at": null
      },
      "entity": {
        "name": "name-71",
        "rules": [
          {
            "protocol": "udp"
          }
        ],
        "spaces_url": "/v2/security_groups/cd186158-b356-474d-9861-724f34f48502/spaces"
      }
    }
  ]
}`
}

func secondListItem() string {
	return `{
  "next_url": null,
  "resources": [
    {
      "metadata": {
        "guid": "d3374b62-7eac-4823-afbd-460d2bf44c67",
        "url": "/v2/security_groups/d3374b62-7eac-4823-afbd-460d2bf44c67",
        "created_at": "2014-06-23T22:55:30+00:00",
        "updated_at": null
      },
      "entity": {
        "name": "name-72",
        "rules": [
          {
            "destination": "198.41.191.47/1"
          }
        ],
        "spaces": [
               {
               	  "metadata":{
               	  	"guid": "my-space-guid"
               	  },
                  "entity": {
                     "name": "my-space",
                     "organization": {
                        "metadata": {
                           "guid": "my-org-guid"
                        },
                        "entity": {
                           "name": "my-org"
                        }
                     }
                  }
               }
            ],
        "spaces_url": "/v2/security_groups/d3374b62-7eac-4823-afbd-460d2bf44c67/spaces"
      }
    }
  ]
}`
}

func spacesItems() string {
	return `{
   "resources": [
      {
				"metadata":{
					"guid": "my-space-guid"
				},
				"entity": {
					 "name": "my-space",
					 "organization": {
							"metadata": {
								 "guid": "my-org-guid"
							},
							"entity": {
								 "name": "my-org"
							}
					 }
				}
      }
   ]
}`
}
