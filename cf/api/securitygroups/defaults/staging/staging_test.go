package staging_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	. "code.cloudfoundry.org/cli/cf/api/securitygroups/defaults/staging"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	testnet "code.cloudfoundry.org/cli/util/testhelpers/net"

	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StagingSecurityGroupsRepo", func() {
	var (
		testServer  *httptest.Server
		testHandler *testnet.TestHandler
		configRepo  coreconfig.ReadWriter
		repo        SecurityGroupsRepo
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
		repo = NewSecurityGroupsRepo(configRepo, gateway)
	})

	AfterEach(func() {
		testServer.Close()
	})

	setupTestServer := func(reqs ...testnet.TestRequest) {
		testServer, testHandler = testnet.NewServer(reqs)
		configRepo.SetAPIEndpoint(testServer.URL)
	}

	Describe("BindToStagingSet", func() {
		It("makes a correct request", func() {
			setupTestServer(
				apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "PUT",
					Path:   "/v2/config/staging_security_groups/a-real-guid",
					Response: testnet.TestResponse{
						Status: http.StatusCreated,
						Body:   bindStagingResponse,
					},
				}),
			)

			err := repo.BindToStagingSet("a-real-guid")

			Expect(err).ToNot(HaveOccurred())
			Expect(testHandler).To(HaveAllRequestsCalled())
		})
	})

	Describe(".List", func() {
		It("returns a list of security groups that are the defaults for staging", func() {
			setupTestServer(
				apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "GET",
					Path:   "/v2/config/staging_security_groups",
					Response: testnet.TestResponse{
						Status: http.StatusOK,
						Body:   firstStagingListItem,
					},
				}),
				apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
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
			Expect(defaults).To(ConsistOf([]models.SecurityGroupFields{
				{
					Name:     "name-71",
					GUID:     "cd186158-b356-474d-9861-724f34f48502",
					SpaceURL: "/v2/security_groups/d3374b62-7eac-4823-afbd-460d2bf44c67/spaces",
					Rules: []map[string]interface{}{{
						"protocol": "udp",
					}},
				},
				{
					Name:     "name-72",
					GUID:     "d3374b62-7eac-4823-afbd-460d2bf44c67",
					SpaceURL: "/v2/security_groups/d3374b62-7eac-4823-afbd-460d2bf44c67/spaces",
					Rules: []map[string]interface{}{{
						"destination": "198.41.191.47/1",
					}},
				},
			}))
		})
	})

	Describe("UnbindFromStagingSet", func() {
		It("makes a correct request", func() {
			testServer, testHandler = testnet.NewServer([]testnet.TestRequest{
				apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "DELETE",
					Path:   "/v2/config/staging_security_groups/my-guid",
					Response: testnet.TestResponse{
						Status: http.StatusNoContent,
					},
				}),
			})

			configRepo.SetAPIEndpoint(testServer.URL)
			err := repo.UnbindFromStagingSet("my-guid")

			Expect(err).ToNot(HaveOccurred())
			Expect(testHandler).To(HaveAllRequestsCalled())
		})
	})
})

var bindStagingResponse string = `{
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
