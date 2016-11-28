package stacks_test

import (
	"net/http"
	"time"

	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"

	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"

	"github.com/onsi/gomega/ghttp"

	. "code.cloudfoundry.org/cli/cf/api/stacks"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StacksRepo", func() {
	var (
		testServer *ghttp.Server
		configRepo coreconfig.ReadWriter
		repo       StackRepository
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		configRepo.SetAccessToken("BEARER my_access_token")

		gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
		repo = NewCloudControllerStackRepository(configRepo, gateway)
	})

	BeforeEach(func() {
		testServer = ghttp.NewServer()
		configRepo.SetAPIEndpoint(testServer.URL())
	})

	AfterEach(func() {
		if testServer != nil {
			testServer.Close()
		}
	})

	Describe("FindByName", func() {
		Context("when a stack exists", func() {
			BeforeEach(func() {
				testServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/stacks", "q=name%3Alinux"),
						ghttp.RespondWith(http.StatusOK, `{
							"resources": [
								{
									"metadata": { "guid": "custom-linux-guid" },
									"entity": { "name": "custom-linux" }
								}
							]
						}`),
					),
				)
			})

			It("tries to find the stack", func() {
				repo.FindByName("linux")
				Expect(testServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns the stack", func() {
				stack, err := repo.FindByName("linux")
				Expect(err).NotTo(HaveOccurred())
				Expect(stack).To(Equal(models.Stack{
					Name: "custom-linux",
					GUID: "custom-linux-guid",
				}))
			})
		})

		Context("when the stack cannot be found", func() {
			BeforeEach(func() {
				testServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/stacks", "q=name%3Alinux"),
						ghttp.RespondWith(http.StatusOK, `{
							"resources": []
						}`),
					),
				)
			})

			It("tries to find the stack", func() {
				repo.FindByName("linux")
				Expect(testServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns an error", func() {
				_, err := repo.FindByName("linux")
				Expect(err).To(BeAssignableToTypeOf(&errors.ModelNotFoundError{}))
			})
		})
	})

	Describe("FindAll", func() {
		BeforeEach(func() {
			testServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/stacks"),
					ghttp.RespondWith(http.StatusOK, `{
							"next_url": "/v2/stacks?page=2",
							"resources": [
								{
									"metadata": {
										"guid": "stack-guid-1",
										"url": "/v2/stacks/stack-guid-1",
										"created_at": "2013-08-31 01:32:40 +0000",
										"updated_at": "2013-08-31 01:32:40 +0000"
									},
									"entity": {
										"name": "lucid64",
										"description": "Ubuntu 10.04"
									}
								}
							]
						}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/stacks"),
					ghttp.RespondWith(http.StatusOK, `{
							"resources": [
								{
									"metadata": {
										"guid": "stack-guid-2",
										"url": "/v2/stacks/stack-guid-2",
										"created_at": "2013-08-31 01:32:40 +0000",
										"updated_at": "2013-08-31 01:32:40 +0000"
									},
									"entity": {
										"name": "lucid64custom",
										"description": "Fake Ubuntu 10.04"
									}
								}
							]
						}`),
				),
			)
		})

		It("tries to find all stacks", func() {
			repo.FindAll()
			Expect(testServer.ReceivedRequests()).To(HaveLen(2))
		})

		It("returns the stacks it found", func() {
			stacks, err := repo.FindAll()
			Expect(err).NotTo(HaveOccurred())
			Expect(stacks).To(ConsistOf([]models.Stack{
				{
					GUID:        "stack-guid-1",
					Name:        "lucid64",
					Description: "Ubuntu 10.04",
				},
				{
					GUID:        "stack-guid-2",
					Name:        "lucid64custom",
					Description: "Fake Ubuntu 10.04",
				},
			}))
		})
	})

	Describe("FindByGUID", func() {
		Context("when a stack with that GUID can be found", func() {
			BeforeEach(func() {
				testServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/stacks/the-stack-guid"),
						ghttp.RespondWith(http.StatusOK, `{
						  "metadata": {
						    "guid": "the-stack-guid",
						    "url": "/v2/stacks/the-stack-guid",
						    "created_at": "2016-01-26T22:20:04Z",
						    "updated_at": null
						  },
						  "entity": {
						    "name": "the-stack-name",
						    "description": "the-stack-description"
						  }
						}`),
					),
				)
			})

			It("tries to find the stack", func() {
				repo.FindByGUID("the-stack-guid")
				Expect(testServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns a stack", func() {
				stack, err := repo.FindByGUID("the-stack-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(stack).To(Equal(models.Stack{
					GUID:        "the-stack-guid",
					Name:        "the-stack-name",
					Description: "the-stack-description",
				}))
			})
		})

		Context("when a stack with that GUID cannot be found", func() {
			BeforeEach(func() {
				testServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/stacks/the-stack-guid"),
						ghttp.RespondWith(http.StatusNotFound, `{
							"code": 250003,
							"description": "The stack could not be found: ea541500-a1bd-4ac2-8ab1-f38ed3a483ad",
							"error_code": "CF-StackNotFound"
						}`),
					),
				)
			})

			It("returns an error", func() {
				_, err := repo.FindByGUID("the-stack-guid")
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(&errors.HTTPNotFoundError{}))
			})
		})
	})

	Context("when finding the stack results in an error", func() {
		BeforeEach(func() {
			testServer.Close()
			testServer = nil
		})

		It("returns an error", func() {
			_, err := repo.FindByGUID("the-stack-guid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Error retrieving stacks: "))
		})
	})
})
