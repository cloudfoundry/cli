package api_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"

	. "github.com/cloudfoundry/cli/cf/api"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("app security group api", func() {
	var (
		testServer  *httptest.Server
		testHandler *testnet.TestHandler
		configRepo  configuration.ReadWriter
		repo        SecurityGroupRepo
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		gateway := net.NewCloudControllerGateway((configRepo), time.Now)
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
				Path:   "/v2/app_security_groups",
				// FIXME: this matcher depend on the order of the key/value pairs in the map
				Matcher: testnet.RequestBodyMatcher(`{
					"name": "mygroup",
					"rules": [{"my-house": "my-rules"}],
					"space_guids": ["myspace"]
				}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})
			setupTestServer(req)

			err := repo.Create(
				"mygroup",
				[]map[string]string{{"my-house": "my-rules"}},
				[]string{"myspace"},
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(testHandler).To(testnet.HaveAllRequestsCalled())
		})
	})

	Describe(".Read", func() {
		It("returns the app security group with the given name", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/app_security_groups?q=name:the-name&inline-relations-depth=1",
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
            "spaces": [
               {
               	  "metadata":{
               	  	"guid": "my-space-guid"
               	  },
                  "entity": {
                     "name": "my-space"
                  }
               }
            ]
         }
      }
   ]
}
					`,
				},
			}))

			group, err := repo.Read("the-name")

			Expect(err).ToNot(HaveOccurred())
			Expect(group).To(Equal(models.ApplicationSecurityGroup{
				ApplicationSecurityGroupFields: models.ApplicationSecurityGroupFields{
					Name:  "the-name",
					Guid:  "the-group-guid",
					Rules: []map[string]string{{"key": "value"}},
				},
				Spaces: []models.SpaceFields{{Guid: "my-space-guid", Name: "my-space"}},
			}))
		})

		It("returns a ModelNotFound error if the security group cannot be found", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/app_security_groups?q=name:the-name&inline-relations-depth=1",
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
		It("deletes the application security group", func() {
			appSecurityGroupGuid := "the-security-group-guid"
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "DELETE",
				Path:   "/v2/app_security_groups/" + appSecurityGroupGuid,
				Response: testnet.TestResponse{
					Status: http.StatusNoContent,
				},
			}))

			err := repo.Delete(appSecurityGroupGuid)

			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe(".FindAll", func() {
		It("returns all the application security groups", func() {
			setupTestServer(
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "GET",
					Path:   "/v2/app_security_groups?inline-relations-depth=1",
					Response: testnet.TestResponse{
						Status: http.StatusOK,
						Body:   firstListItem(),
					},
				}),
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "GET",
					Path:   "/v2/app_security_groups?inline-relations-depth=1&page=2",
					Response: testnet.TestResponse{
						Status: http.StatusOK,
						Body:   secondListItem(),
					},
				}),
			)

			groups, err := repo.FindAll()

			Expect(err).ToNot(HaveOccurred())
			Expect(groups[0]).To(Equal(models.ApplicationSecurityGroup{
				ApplicationSecurityGroupFields: models.ApplicationSecurityGroupFields{
					Name:  "name-71",
					Guid:  "cd186158-b356-474d-9861-724f34f48502",
					Rules: []map[string]string{{"protocol": "udp"}},
				},
				Spaces: []models.SpaceFields{},
			}))
			Expect(groups[1]).To(Equal(models.ApplicationSecurityGroup{
				ApplicationSecurityGroupFields: models.ApplicationSecurityGroupFields{
					Name:  "name-72",
					Guid:  "d3374b62-7eac-4823-afbd-460d2bf44c67",
					Rules: []map[string]string{{"destination": "198.41.191.47/1"}},
				},
				Spaces: []models.SpaceFields{{Guid: "my-space-guid", Name: "my-space"}},
			}))
		})
	})
})

func firstListItem() string {
	return `{
  "next_url": "/v2/app_security_groups?inline-relations-depth=1&page=2",
  "resources": [
    {
      "metadata": {
        "guid": "cd186158-b356-474d-9861-724f34f48502",
        "url": "/v2/app_security_groups/cd186158-b356-474d-9861-724f34f48502",
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
        "spaces_url": "/v2/app_security_groups/cd186158-b356-474d-9861-724f34f48502/spaces"
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
        "url": "/v2/app_security_groups/d3374b62-7eac-4823-afbd-460d2bf44c67",
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
                     "name": "my-space"
                  }
               }
            ],
        "spaces_url": "/v2/app_security_groups/d3374b62-7eac-4823-afbd-460d2bf44c67/spaces"
      }
    }
  ]
}`
}
