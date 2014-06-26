package api_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"

	. "github.com/cloudfoundry/cli/cf/api"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StagingSecurityGroupsRepo", func() {
	var (
		testServer  *httptest.Server
		testHandler *testnet.TestHandler
		configRepo  configuration.ReadWriter
		repo        StagingSecurityGroupsRepo
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		gateway := net.NewCloudControllerGateway((configRepo), time.Now)
		repo = NewStagingSecurityGroupsRepo(configRepo, gateway)
	})

	AfterEach(func() {
		testServer.Close()
	})

	setupTestServer := func(reqs ...testnet.TestRequest) {
		testServer, testHandler = testnet.NewServer(reqs)
		configRepo.SetApiEndpoint(testServer.URL)
	}

	Describe("AddToDefaultStagingSet", func() {
		It("makes a correct request", func() {
			setupTestServer(
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "PUT",
					Path:   "/v2/config/staging_security_groups/a-real-guid",
					Response: testnet.TestResponse{
						Status: http.StatusCreated,
						Body:   addStagingResponse,
					},
				}),
			)

			err := repo.AddToDefaultStagingSet("a-real-guid")

			Expect(err).ToNot(HaveOccurred())
			Expect(testHandler).To(HaveAllRequestsCalled())
		})
	})

	Describe(".List", func() {
		It("returns a list of security groups that are the defaults for staging", func() {
			setupTestServer(
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "GET",
					Path:   "/v2/config/staging_security_groups",
					Response: testnet.TestResponse{
						Status: http.StatusOK,
						Body:   firstStagingListItem,
					},
				}),
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "GET",
					Path:   "/v2/config/staging_security_groups",
					Response: testnet.TestResponse{
						Status: http.StatusOK,
						Body:   secondStagingListItem,
					},
				}),
			)

			defaults, err := repo.List()

			Expect(err).ToNot(HaveOccurred())
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(defaults).To(Equal([]models.SecurityGroupFields{
				{
					Name: "name-71",
					Guid: "cd186158-b356-474d-9861-724f34f48502",
					Rules: []map[string]string{{
						"protocol": "udp",
					}},
				},
				{
					Name: "name-72",
					Guid: "d3374b62-7eac-4823-afbd-460d2bf44c67",
					Rules: []map[string]string{{
						"destination": "198.41.191.47/1",
					}},
				},
			}))
		})
	})
})

var addStagingResponse string = `{
  "metadata": {
    "guid": "897341eb-ef31-406f-b57b-414f51583a3a",
    "url": "/v2/config/staging_security_groups/897341eb-ef31-406f-b57b-414f51583a3a",
    "created_at": "2014-06-23T21:43:30+00:00",
    "updated_at": "2014-06-23T21:43:30+00:00"
  },
  "entity": {
    "name": "name-904",
    "rules": [
      {
        "protocol": "udp",
        "ports": "8080",
        "destination": "198.41.191.47/1"
      }
    ]
  }
}`

var firstStagingListItem string = `{
  "next_url": "/v2/config/staging_security_groups?page=2",
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
        "spaces_url": "/v2/security_groups/d3374b62-7eac-4823-afbd-460d2bf44c67/spaces"
      }
    }
  ]
}`

var secondStagingListItem string = `{
  "next_url": null,
  "resources": [
    {
      "metadata": {
        "guid": "d3374b62-7eac-4823-afbd-460d2bf44c67",
        "url": "/v2/config/staging_security_groups/d3374b62-7eac-4823-afbd-460d2bf44c67",
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
        "spaces_url": "/v2/security_groups/d3374b62-7eac-4823-afbd-460d2bf44c67/spaces"
      }
    }
  ]
}`
