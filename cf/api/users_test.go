package api_test

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"

	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("UserRepository", func() {
	var (
		client api.UserRepository

		config     coreconfig.ReadWriter
		ccServer   *ghttp.Server
		uaaServer  *ghttp.Server
		ccGateway  net.Gateway
		uaaGateway net.Gateway
	)

	BeforeEach(func() {
		ccServer = ghttp.NewServer()
		uaaServer = ghttp.NewServer()

		config = testconfig.NewRepositoryWithDefaults()
		config.SetAPIEndpoint(ccServer.URL())
		config.SetUaaEndpoint(uaaServer.URL())
		ccGateway = net.NewCloudControllerGateway(config, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
		uaaGateway = net.NewUAAGateway(config, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
		client = api.NewCloudControllerUserRepository(config, uaaGateway, ccGateway)
	})

	AfterEach(func() {
		if ccServer != nil {
			ccServer.Close()
		}
		if uaaServer != nil {
			uaaServer.Close()
		}
	})

	Describe("ListUsersInOrgForRole", func() {
		Context("when there are no users in the given org with the given role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/organizations/org-guid/managers"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{"resources":[]}`),
					),
				)
			})

			It("makes a request to CC", func() {
				_, err := client.ListUsersInOrgForRole("org-guid", models.RoleOrgManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns no users", func() {
				users, err := client.ListUsersInOrgForRole("org-guid", models.RoleOrgManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(users)).To(Equal(0))
			})
		})

		Context("when there are users in the given org with the given role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/organizations/org-guid/managers"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
							"resources":[
							{"metadata": {"guid": "user-1-guid"}, "entity": {}}
							]}`),
					),
				)

				uaaServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/Users", fmt.Sprintf("attributes=id,userName&filter=%s", url.QueryEscape(`ID eq "user-1-guid"`))),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
							"resources": [
							{ "id": "user-1-guid", "userName": "Super user 1" }
							]}`),
					),
				)
			})

			It("makes a request to CC", func() {
				_, err := client.ListUsersInOrgForRole("org-guid", models.RoleOrgManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("makes a request to UAA", func() {
				_, err := client.ListUsersInOrgForRole("org-guid", models.RoleOrgManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns the users", func() {
				users, err := client.ListUsersInOrgForRole("org-guid", models.RoleOrgManager)
				Expect(err).NotTo(HaveOccurred())

				Expect(len(users)).To(Equal(1))
				Expect(users[0].GUID).To(Equal("user-1-guid"))
				Expect(users[0].Username).To(Equal("Super user 1"))
			})
		})

		Context("when there are multiple pages of users in the given org with the given role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/organizations/org-guid/managers"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
								"next_url": "/v2/organizations/org-guid/managers?page=2",
								"resources":[
								{"metadata": {"guid": "user-1-guid"}, "entity": {}}
								]}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/organizations/org-guid/managers", "page=2"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
								"resources":[
								{"metadata": {"guid": "user-2-guid"}, "entity": {"username":"user 2 from cc"}},
								{"metadata": {"guid": "user-3-guid"}, "entity": {"username":"user 3 from cc"}}
								]}`),
					),
				)

				uaaServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/Users", fmt.Sprintf("attributes=id,userName&filter=%s", url.QueryEscape(`ID eq "user-1-guid" or ID eq "user-2-guid" or ID eq "user-3-guid"`))),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
								"resources": [
								{ "id": "user-1-guid", "userName": "Super user 1" },
								{ "id": "user-2-guid", "userName": "Super user 2" },
								{ "id": "user-3-guid", "userName": "Super user 3" }
								]
							}`),
					),
				)
			})

			It("makes a request to CC for each page of results", func() {
				_, err := client.ListUsersInOrgForRole("org-guid", models.RoleOrgManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
			})

			It("makes a request to UAA", func() {
				_, err := client.ListUsersInOrgForRole("org-guid", models.RoleOrgManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns all paginated users", func() {
				users, err := client.ListUsersInOrgForRole("org-guid", models.RoleOrgManager)
				Expect(err).NotTo(HaveOccurred())

				Expect(len(users)).To(Equal(3))
				Expect(users[0].GUID).To(Equal("user-1-guid"))
				Expect(users[0].Username).To(Equal("Super user 1"))
				Expect(users[1].GUID).To(Equal("user-2-guid"))
				Expect(users[1].Username).To(Equal("Super user 2"))
				Expect(users[2].GUID).To(Equal("user-3-guid"))
				Expect(users[2].Username).To(Equal("Super user 3"))
			})
		})

		Context("when CC returns an error", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/organizations/org-guid/managers"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusGatewayTimeout, nil),
					),
				)
			})

			It("does not make a request to UAA", func() {
				client.ListUsersInOrgForRole("org-guid", models.RoleOrgManager)
				Expect(uaaServer.ReceivedRequests()).To(BeZero())
			})

			It("returns an error", func() {
				_, err := client.ListUsersInOrgForRole("org-guid", models.RoleOrgManager)
				httpErr, ok := err.(errors.HTTPError)
				Expect(ok).To(BeTrue())
				Expect(httpErr.StatusCode()).To(Equal(http.StatusGatewayTimeout))
			})
		})

		Context("when the UAA endpoint has not been configured", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/organizations/org-guid/managers"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
							"resources":[
							{"metadata": {"guid": "user-1-guid"}, "entity": {}}
							]}`),
					),
				)
				config.SetUaaEndpoint("")
			})

			It("returns an error", func() {
				_, err := client.ListUsersInOrgForRole("org-guid", models.RoleOrgManager)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("ListUsersInOrgForRoleWithNoUAA", func() {
		Context("when there are users in the given org with the given role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/organizations/org-guid/managers"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
							"resources":[
							{"metadata": {"guid": "user-1-guid"}, "entity": {}}
							]}`),
					),
				)
			})

			It("makes a request to CC", func() {
				_, err := client.ListUsersInOrgForRoleWithNoUAA("org-guid", models.RoleOrgManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("does not make a request to UAA", func() {
				_, err := client.ListUsersInOrgForRoleWithNoUAA("org-guid", models.RoleOrgManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(uaaServer.ReceivedRequests()).To(BeZero())
			})

			It("returns the users", func() {
				users, err := client.ListUsersInOrgForRoleWithNoUAA("org-guid", models.RoleOrgManager)
				Expect(err).NotTo(HaveOccurred())

				Expect(len(users)).To(Equal(1))
				Expect(users[0].GUID).To(Equal("user-1-guid"))
				Expect(users[0].Username).To(BeEmpty())
			})
		})

		Context("when there are multiple pages of users in the given org with the given role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/organizations/org-guid/managers"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
								"next_url": "/v2/organizations/org-guid/managers?page=2",
								"resources":[
								{"metadata": {"guid": "user-1-guid"}, "entity": {}}
								]}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/organizations/org-guid/managers", "page=2"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
									"resources":[
									{"metadata": {"guid": "user-2-guid"}, "entity": {"username":"user 2 from cc"}},
									{"metadata": {"guid": "user-3-guid"}, "entity": {"username":"user 3 from cc"}}
									]}`),
					),
				)
			})

			It("makes a request to CC for each page of results", func() {
				_, err := client.ListUsersInOrgForRoleWithNoUAA("org-guid", models.RoleOrgManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
			})

			It("does not make a request to UAA", func() {
				_, err := client.ListUsersInOrgForRoleWithNoUAA("org-guid", models.RoleOrgManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(uaaServer.ReceivedRequests()).To(BeZero())
			})

			It("returns all paginated users", func() {
				users, err := client.ListUsersInOrgForRoleWithNoUAA("org-guid", models.RoleOrgManager)
				Expect(err).NotTo(HaveOccurred())

				Expect(len(users)).To(Equal(3))
				Expect(users[0].GUID).To(Equal("user-1-guid"))
				Expect(users[0].Username).To(BeEmpty())
				Expect(users[1].GUID).To(Equal("user-2-guid"))
				Expect(users[1].Username).To(Equal("user 2 from cc"))
				Expect(users[2].GUID).To(Equal("user-3-guid"))
				Expect(users[2].Username).To(Equal("user 3 from cc"))
			})
		})

		Context("when there are no users in the given org with the given role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/organizations/org-guid/managers"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{"resources":[]}`),
					),
				)
			})

			It("makes a request to CC", func() {
				_, err := client.ListUsersInOrgForRoleWithNoUAA("org-guid", models.RoleOrgManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("does not make a request to UAA", func() {
				_, err := client.ListUsersInOrgForRoleWithNoUAA("org-guid", models.RoleOrgManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(uaaServer.ReceivedRequests()).To(BeZero())
			})

			It("returns no users", func() {
				users, err := client.ListUsersInOrgForRoleWithNoUAA("org-guid", models.RoleOrgManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(users)).To(Equal(0))
			})
		})

		Context("when CC returns an error", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/organizations/org-guid/managers"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusGatewayTimeout, nil),
					),
				)
			})

			It("does not make a request to UAA", func() {
				client.ListUsersInOrgForRoleWithNoUAA("org-guid", models.RoleOrgManager)
				Expect(uaaServer.ReceivedRequests()).To(BeZero())
			})

			It("returns an error", func() {
				_, err := client.ListUsersInOrgForRoleWithNoUAA("org-guid", models.RoleOrgManager)
				httpErr, ok := err.(errors.HTTPError)
				Expect(ok).To(BeTrue())
				Expect(httpErr.StatusCode()).To(Equal(http.StatusGatewayTimeout))
			})
		})
	})

	Describe("ListUsersInSpaceForRole", func() {
		Context("when there are users in the space with the given role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/spaces/space-guid/managers"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
							"resources":[
							{"metadata": {"guid": "user-1-guid"}, "entity": {}}
							]}`),
					),
				)

				uaaServer = ghttp.NewServer()
				config.SetUaaEndpoint(uaaServer.URL())
				uaaServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/Users", fmt.Sprintf("attributes=id,userName&filter=%s", url.QueryEscape(`ID eq "user-1-guid"`))),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
								"resources": [
								{ "id": "user-1-guid", "userName": "Super user 1" }
								]
							}`),
					),
				)
			})

			It("makes a request to CC", func() {
				_, err := client.ListUsersInSpaceForRole("space-guid", models.RoleSpaceManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("makes a request to UAA", func() {
				_, err := client.ListUsersInSpaceForRole("space-guid", models.RoleSpaceManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns the users", func() {
				users, err := client.ListUsersInSpaceForRole("space-guid", models.RoleSpaceManager)
				Expect(err).NotTo(HaveOccurred())

				Expect(len(users)).To(Equal(1))
				Expect(users[0].GUID).To(Equal("user-1-guid"))
				Expect(users[0].Username).To(Equal("Super user 1"))
			})
		})

		Context("when there are no users in the given space with the given role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/spaces/space-guid/managers"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{"resources":[]}`),
					),
				)
			})

			It("makes a request to CC", func() {
				_, err := client.ListUsersInSpaceForRole("space-guid", models.RoleSpaceManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns no users", func() {
				users, err := client.ListUsersInSpaceForRole("space-guid", models.RoleSpaceManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(users)).To(Equal(0))
			})
		})
	})

	Describe("ListUsersInSpaceForRoleWithNoUAA", func() {
		Context("when there are users in the given space with the given role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/spaces/space-guid/managers"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
							"resources":[
							{"metadata": {"guid": "user-1-guid"}, "entity": {}}
							]}`),
					),
				)
			})

			It("makes a request to CC", func() {
				_, err := client.ListUsersInSpaceForRoleWithNoUAA("space-guid", models.RoleSpaceManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("does not make a request to UAA", func() {
				_, err := client.ListUsersInSpaceForRoleWithNoUAA("space-guid", models.RoleSpaceManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(uaaServer.ReceivedRequests()).To(BeZero())
			})

			It("returns the users", func() {
				users, err := client.ListUsersInSpaceForRoleWithNoUAA("space-guid", models.RoleSpaceManager)
				Expect(err).NotTo(HaveOccurred())

				Expect(len(users)).To(Equal(1))
				Expect(users[0].GUID).To(Equal("user-1-guid"))
				Expect(users[0].Username).To(BeEmpty())
			})
		})

		Context("when there are multiple pages of users in the given space with the given role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/spaces/space-guid/managers"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
								"next_url": "/v2/spaces/space-guid/managers?page=2",
								"resources":[
								{"metadata": {"guid": "user-1-guid"}, "entity": {}}
								]}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/spaces/space-guid/managers", "page=2"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
									"resources":[
									{"metadata": {"guid": "user-2-guid"}, "entity": {"username":"user 2 from cc"}},
									{"metadata": {"guid": "user-3-guid"}, "entity": {"username":"user 3 from cc"}}
									]}`),
					),
				)
			})

			It("makes a request to CC for each page of results", func() {
				_, err := client.ListUsersInSpaceForRoleWithNoUAA("space-guid", models.RoleSpaceManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
			})

			It("does not make a request to UAA", func() {
				_, err := client.ListUsersInSpaceForRoleWithNoUAA("space-guid", models.RoleSpaceManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(uaaServer.ReceivedRequests()).To(BeZero())
			})

			It("returns all paginated users", func() {
				users, err := client.ListUsersInSpaceForRoleWithNoUAA("space-guid", models.RoleSpaceManager)
				Expect(err).NotTo(HaveOccurred())

				Expect(len(users)).To(Equal(3))
				Expect(users[0].GUID).To(Equal("user-1-guid"))
				Expect(users[0].Username).To(BeEmpty())
				Expect(users[1].GUID).To(Equal("user-2-guid"))
				Expect(users[1].Username).To(Equal("user 2 from cc"))
				Expect(users[2].GUID).To(Equal("user-3-guid"))
				Expect(users[2].Username).To(Equal("user 3 from cc"))
			})
		})

		Context("when there are no users in the given space with the given role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/spaces/space-guid/managers"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{"resources":[]}`),
					),
				)
			})

			It("makes a request to CC", func() {
				_, err := client.ListUsersInSpaceForRoleWithNoUAA("space-guid", models.RoleSpaceManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("does not make a request to UAA", func() {
				_, err := client.ListUsersInSpaceForRoleWithNoUAA("space-guid", models.RoleSpaceManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(uaaServer.ReceivedRequests()).To(BeZero())
			})

			It("returns no users", func() {
				users, err := client.ListUsersInSpaceForRoleWithNoUAA("space-guid", models.RoleSpaceManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(users)).To(Equal(0))
			})
		})

		Context("when CC returns an error", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/spaces/space-guid/managers"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusGatewayTimeout, nil),
					),
				)
			})

			It("does not make a request to UAA", func() {
				client.ListUsersInSpaceForRoleWithNoUAA("space-guid", models.RoleSpaceManager)
				Expect(uaaServer.ReceivedRequests()).To(BeZero())
			})

			It("returns an error", func() {
				_, err := client.ListUsersInSpaceForRoleWithNoUAA("space-guid", models.RoleSpaceManager)
				httpErr, ok := err.(errors.HTTPError)
				Expect(ok).To(BeTrue())
				Expect(httpErr.StatusCode()).To(Equal(http.StatusGatewayTimeout))
			})
		})
	})

	Describe("FindAllByUsername", func() {
		Context("when the user exists", func() {
			BeforeEach(func() {
				uaaServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/Users", fmt.Sprintf("attributes=id,userName&filter=%s", url.QueryEscape(`userName Eq "damien+user1@pivotallabs.com"`))),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
								"resources": [
								{ "id": "my-guid", "userName": "my-full-username" },
								{ "id": "my-guid-2", "userName": "my-full-username-2" }
								]}`),
					),
				)
			})

			It("makes a request to UAA", func() {
				_, err := client.FindAllByUsername("damien+user1@pivotallabs.com")
				Expect(err).NotTo(HaveOccurred())
				Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns the users", func() {
				users, err := client.FindAllByUsername("damien+user1@pivotallabs.com")
				Expect(err).NotTo(HaveOccurred())
				Expect(users).To(Equal([]models.UserFields{
					{
						Username: "my-full-username",
						GUID:     "my-guid",
					},
					{
						Username: "my-full-username-2",
						GUID:     "my-guid-2",
					},
				}))
			})
		})

		Context("when the user does not exist", func() {
			BeforeEach(func() {
				uaaServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/Users", fmt.Sprintf("attributes=id,userName&filter=%s", url.QueryEscape(`userName Eq "damien+user1@pivotallabs.com"`))),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{"resources": []}`),
					),
				)
			})

			It("makes a request to UAA", func() {
				client.FindByUsername("damien+user1@pivotallabs.com")
				Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns an error", func() {
				_, err := client.FindByUsername("damien+user1@pivotallabs.com")
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(&errors.ModelNotFoundError{}))
			})
		})

		Context("when the cli user is not authorized", func() {
			BeforeEach(func() {
				uaaServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/Users", fmt.Sprintf("attributes=id,userName&filter=%s", url.QueryEscape(`userName Eq "damien+user1@pivotallabs.com"`))),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusForbidden, `{"error":"access_denied","error_description":"Access is denied"}`),
					),
				)
			})

			It("makes a request to UAA", func() {
				client.FindByUsername("damien+user1@pivotallabs.com")
				Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns an error", func() {
				_, err := client.FindByUsername("damien+user1@pivotallabs.com")
				Expect(err).To(BeAssignableToTypeOf(&errors.AccessDeniedError{}))
			})
		})

		Context("when UAA returns a non-403 error", func() {
			BeforeEach(func() {
				uaaServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/Users", fmt.Sprintf("attributes=id,userName&filter=%s", url.QueryEscape(`userName Eq "damien+user1@pivotallabs.com"`))),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusInternalServerError, nil),
					),
				)
			})

			It("makes a request to UAA", func() {
				client.FindByUsername("damien+user1@pivotallabs.com")
				Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns an error", func() {
				_, err := client.FindByUsername("damien+user1@pivotallabs.com")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("FindByUsername", func() {
		Context("when the user exists", func() {
			BeforeEach(func() {
				uaaServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/Users", fmt.Sprintf("attributes=id,userName&filter=%s", url.QueryEscape(`userName Eq "damien+user1@pivotallabs.com"`))),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
								"resources": [
								{ "id": "my-guid", "userName": "my-full-username" }
								]}`),
					),
				)
			})

			It("makes a request to UAA", func() {
				_, err := client.FindByUsername("damien+user1@pivotallabs.com")
				Expect(err).NotTo(HaveOccurred())
				Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns the user", func() {
				user, err := client.FindByUsername("damien+user1@pivotallabs.com")
				Expect(err).NotTo(HaveOccurred())
				Expect(user).To(Equal(models.UserFields{
					Username: "my-full-username",
					GUID:     "my-guid",
				}))
			})
		})

		Context("when the user does not exist", func() {
			BeforeEach(func() {
				uaaServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/Users", fmt.Sprintf("attributes=id,userName&filter=%s", url.QueryEscape(`userName Eq "damien+user1@pivotallabs.com"`))),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{"resources": []}`),
					),
				)
			})

			It("makes a request to UAA", func() {
				client.FindByUsername("damien+user1@pivotallabs.com")
				Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns an error", func() {
				_, err := client.FindByUsername("damien+user1@pivotallabs.com")
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(&errors.ModelNotFoundError{}))
			})
		})

		Context("when the cli user is not authorized", func() {
			BeforeEach(func() {
				uaaServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/Users", fmt.Sprintf("attributes=id,userName&filter=%s", url.QueryEscape(`userName Eq "damien+user1@pivotallabs.com"`))),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusForbidden, `{"error":"access_denied","error_description":"Access is denied"}`),
					),
				)
			})

			It("makes a request to UAA", func() {
				client.FindByUsername("damien+user1@pivotallabs.com")
				Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns an error", func() {
				_, err := client.FindByUsername("damien+user1@pivotallabs.com")
				Expect(err).To(BeAssignableToTypeOf(&errors.AccessDeniedError{}))
			})
		})

		Context("when UAA returns a non-403 error", func() {
			BeforeEach(func() {
				uaaServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/Users", fmt.Sprintf("attributes=id,userName&filter=%s", url.QueryEscape(`userName Eq "damien+user1@pivotallabs.com"`))),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusInternalServerError, nil),
					),
				)
			})

			It("makes a request to UAA", func() {
				client.FindByUsername("damien+user1@pivotallabs.com")
				Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns an error", func() {
				_, err := client.FindByUsername("damien+user1@pivotallabs.com")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Create", func() {
		Context("when the user does not exist", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v2/users"),
						ghttp.VerifyJSON(`{"guid":"my-user-guid"}`),
					),
				)
				uaaServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/Users"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.VerifyJSON(`{
							"userName":"my-user",
							"emails":[{"value":"my-user"}],
							"password":"my-password",
							"name":{
								"givenName":"my-user",
								"familyName":"my-user"}
							}`),
						ghttp.RespondWith(http.StatusOK, `{"id":"my-user-guid"}`),
					),
				)
			})

			It("makes a call to CC", func() {
				err := client.Create("my-user", "my-password")
				Expect(err).ToNot(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("makes a call to UAA", func() {
				err := client.Create("my-user", "my-password")
				Expect(err).ToNot(HaveOccurred())
				Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
			})
		})

		Context("when the user already exists", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v2/users"),
						ghttp.VerifyJSON(`{"guid":"my-user-guid"}`),
					),
				)
				uaaServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/Users"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.VerifyJSON(`{
								"userName":"my-user",
								"emails":[{"value":"my-user"}],
								"password":"my-password",
								"name":{
									"givenName":"my-user",
									"familyName":"my-user"}
								}`),
						ghttp.RespondWith(http.StatusConflict, `
								{
									"message":"Username already in use: my-user",
									"error":"scim_resource_already_exists"
								}`),
					),
				)
			})

			It("does not make a call to CC", func() {
				client.Create("my-user", "my-password")
				Expect(ccServer.ReceivedRequests()).To(BeZero())
			})

			It("makes a call to UAA", func() {
				client.Create("my-user", "my-password")
				Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns an error", func() {
				err := client.Create("my-user", "my-password")
				Expect(err).To(BeAssignableToTypeOf(&errors.ModelAlreadyExistsError{}))
			})
		})

		Context("when UAA returns an error", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v2/users"),
						ghttp.VerifyJSON(`{"guid":"my-user-guid"}`),
					),
				)
				uaaServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/Users"),
						ghttp.RespondWith(http.StatusForbidden, `
								{
									"message":"Access Denied",
									"error":"Forbidden"
								}`),
					),
				)
			})

			It("makes a call to UAA", func() {
				client.Create("my-user", "my-password")
				Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("does not make a call to CC", func() {
				client.Create("my-user", "my-password")
				Expect(ccServer.ReceivedRequests()).To(BeZero())
			})

			It("returns an error", func() {
				err := client.Create("my-user", "my-password")
				Expect(err.Error()).To(ContainSubstring("Forbidden"))
			})
		})
	})

	Describe("Delete", func() {
		Context("when the user is found in CC", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/v2/users/my-user-guid"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
				uaaServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/Users/my-user-guid"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
			})

			It("makes a call to CC", func() {
				err := client.Delete("my-user-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("makes a call to UAA", func() {
				err := client.Delete("my-user-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
			})
		})

		Context("when the user is not found in CC", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/v2/users/my-user-guid"),
						ghttp.RespondWith(http.StatusNotFound, `
						{
							"code": 20003,
							"description": "The user could not be found"
						}`),
					),
				)
				uaaServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/Users/my-user-guid"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
			})

			It("makes a call to CC", func() {
				err := client.Delete("my-user-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("makes a call to UAA", func() {
				err := client.Delete("my-user-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
			})
		})
	})

	Describe("SetOrgRoleByGUID", func() {
		Context("when given the OrgManager role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/managers/user-guid"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/users/user-guid"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
			})

			It("makes two requests to CC", func() {
				err := client.SetOrgRoleByGUID("user-guid", "org-guid", models.RoleOrgManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
			})
		})

		Context("when given the BillingManager role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/billing_managers/user-guid"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/users/user-guid"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
			})

			It("makes two requests to CC", func() {
				err := client.SetOrgRoleByGUID("user-guid", "org-guid", models.RoleBillingManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
			})
		})

		Context("when given the OrgAuditor role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/auditors/user-guid"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/users/user-guid"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
			})

			It("makes two requests to CC", func() {
				err := client.SetOrgRoleByGUID("user-guid", "org-guid", models.RoleOrgAuditor)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
			})
		})

		Context("when given an invalid role", func() {
			It("does not make a request to CC", func() {
				client.SetOrgRoleByGUID("user-guid", "org-guid", 666)
				Expect(ccServer.ReceivedRequests()).To(BeZero())
			})

			It("returns an error", func() {
				err := client.SetOrgRoleByGUID("user-guid", "org-guid", 666)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("UnsetOrgRoleByGUID", func() {
		Context("when given the OrgManager role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/v2/organizations/org-guid/managers/user-guid"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
			})

			It("makes a request to CC", func() {
				err := client.UnsetOrgRoleByGUID("user-guid", "org-guid", models.RoleOrgManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})
		})

		Context("when given the BillingManager role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/v2/organizations/org-guid/billing_managers/user-guid"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
			})

			It("makes a request to CC", func() {
				err := client.UnsetOrgRoleByGUID("user-guid", "org-guid", models.RoleBillingManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})
		})

		Context("when given the OrgAuditor role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/v2/organizations/org-guid/auditors/user-guid"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
			})

			It("makes a request to CC", func() {
				err := client.UnsetOrgRoleByGUID("user-guid", "org-guid", models.RoleOrgAuditor)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})
		})

		Context("when given an invalid role", func() {
			It("does not make a request to CC", func() {
				client.UnsetOrgRoleByGUID("user-guid", "org-guid", 666)
				Expect(ccServer.ReceivedRequests()).To(BeZero())
			})

			It("returns an error", func() {
				err := client.UnsetOrgRoleByGUID("user-guid", "org-guid", 666)
				Expect(err).To(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(BeZero())
			})
		})
	})

	Describe("UnsetOrgRoleByUsername", func() {
		Context("when given the OrgManager role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/v2/organizations/org-guid/managers"),
						ghttp.VerifyJSON(`{"username":"the-user-name"}`),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
			})

			It("makes a request to CC", func() {
				err := client.UnsetOrgRoleByUsername("the-user-name", "org-guid", models.RoleOrgManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})
		})

		Context("when given the BillingManager role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/v2/organizations/org-guid/billing_managers"),
						ghttp.VerifyJSON(`{"username":"the-user-name"}`),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
			})

			It("makes a request to CC", func() {
				err := client.UnsetOrgRoleByUsername("the-user-name", "org-guid", models.RoleBillingManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})
		})

		Context("when given the OrgAuditor role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/v2/organizations/org-guid/auditors"),
						ghttp.VerifyJSON(`{"username":"the-user-name"}`),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
			})

			It("makes a request to CC", func() {
				err := client.UnsetOrgRoleByUsername("the-user-name", "org-guid", models.RoleOrgAuditor)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})
		})

		Context("when given an invalid role", func() {
			It("does not make a request to CC", func() {
				client.UnsetOrgRoleByUsername("user-guid", "org-guid", 666)
				Expect(ccServer.ReceivedRequests()).To(BeZero())
			})

			It("returns an error", func() {
				err := client.UnsetOrgRoleByUsername("user-guid", "org-guid", 666)
				Expect(err).To(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(BeZero())
			})
		})
	})

	Describe("SetOrgRoleByUsername", func() {
		Context("when given the OrgManager role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/managers"),
						ghttp.VerifyJSON(`{"username":"user@example.com"}`),
						ghttp.RespondWith(http.StatusOK, nil),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/users"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
			})

			It("makes two requests to CC", func() {
				err := client.SetOrgRoleByUsername("user@example.com", "org-guid", models.RoleOrgManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
			})
		})

		Context("when given the BillingManager role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/billing_managers"),
						ghttp.VerifyJSON(`{"username":"user@example.com"}`),
						ghttp.RespondWith(http.StatusOK, nil),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/users"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
			})

			It("makes two requests to CC", func() {
				err := client.SetOrgRoleByUsername("user@example.com", "org-guid", models.RoleBillingManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
			})
		})

		Context("when given the OrgAuditor role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/auditors"),
						ghttp.VerifyJSON(`{"username":"user@example.com"}`),
						ghttp.RespondWith(http.StatusOK, nil),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/users"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
			})

			It("makes two requests to CC", func() {
				err := client.SetOrgRoleByUsername("user@example.com", "org-guid", models.RoleOrgAuditor)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
			})
		})

		Context("when given an invalid role", func() {
			It("returns an error", func() {
				err := client.SetOrgRoleByUsername("user@example.com", "org-guid", 666)
				Expect(err).To(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(BeZero())
			})
		})

		Context("when assigning the given role to the user fails", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/auditors"),
						ghttp.VerifyJSON(`{"username":"user@example.com"}`),
						ghttp.RespondWith(http.StatusInternalServerError, nil),
					),
				)
			})

			It("returns an error", func() {
				err := client.SetOrgRoleByUsername("user@example.com", "org-guid", models.RoleOrgAuditor)
				Expect(err).To(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
				Expect(err.Error()).To(ContainSubstring("status code: 500"))
			})
		})

		Context("when assigning the user to the org fails", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/auditors"),
						ghttp.VerifyJSON(`{"username":"user@example.com"}`),
						ghttp.RespondWith(http.StatusOK, nil),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/users"),
						ghttp.VerifyJSON(`{"username":"user@example.com"}`),
						ghttp.RespondWith(http.StatusInternalServerError, nil),
					),
				)
			})

			It("returns an error", func() {
				err := client.SetOrgRoleByUsername("user@example.com", "org-guid", models.RoleOrgAuditor)
				Expect(err).To(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
				Expect(err.Error()).To(ContainSubstring("status code: 500"))
			})
		})
	})

	Describe("SetSpaceRoleByUsername", func() {
		Context("when operator does not have privilege to add user role to org", func() {
			Context("and user is already part of the parent org", func() {
				BeforeEach(func() {
					ccServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/users"),
							ghttp.VerifyJSON(`{"username":"user@example.com"}`),
							ghttp.RespondWith(http.StatusUnauthorized, `{
  "code": 10003,
  "description": "You are not authorized to perform the requested action",
  "error_code": "CF-NotAuthorized"
}`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/v2/spaces/space-guid/managers"),
							ghttp.VerifyJSON(`{"username":"user@example.com"}`),
							ghttp.RespondWith(http.StatusOK, nil),
						),
					)
				})

				It("sets space role and ignores the '10003' error for adding user to parent org", func() {
					err := client.SetSpaceRoleByUsername("user@example.com", "space-guid", "org-guid", models.RoleSpaceManager)
					Expect(err).NotTo(HaveOccurred())
					Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
				})

				Context("when setting space role returns an error", func() {
					It("wraps '1002' error from set space role with friendly custom message", func() {
						ccServer.SetHandler(1, ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/v2/spaces/space-guid/managers"),
							ghttp.VerifyJSON(`{"username":"user@example.com"}`),
							ghttp.RespondWith(http.StatusBadRequest, `{"code":1002, "description":"invalid relation"}`),
						))

						err := client.SetSpaceRoleByUsername("user@example.com", "space-guid", "org-guid", models.RoleSpaceManager)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("1002"))
						Expect(err.Error()).To(ContainSubstring("cannot set space role because user is not part of the org"))
						Expect(err.Error()).ToNot(ContainSubstring("invalid relation"))
						Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
					})

					It("returns any non '1002' error from set space role", func() {
						ccServer.SetHandler(1, ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/v2/spaces/space-guid/managers"),
							ghttp.VerifyJSON(`{"username":"user@example.com"}`),
							ghttp.RespondWith(http.StatusBadRequest, `{"code":1001, "description":"some error"}`),
						))

						err := client.SetSpaceRoleByUsername("user@example.com", "space-guid", "org-guid", models.RoleSpaceManager)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("1001"))
						Expect(err.Error()).To(ContainSubstring("some error"))
						Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
					})
				})
			})
		})

		Context("when operator does have privilege to add user role to parent org", func() {
			Context("when given the SpaceManager role", func() {
				BeforeEach(func() {
					ccServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/users"),
							ghttp.VerifyJSON(`{"username":"user@example.com"}`),
							ghttp.RespondWith(http.StatusOK, nil),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/v2/spaces/space-guid/managers"),
							ghttp.VerifyJSON(`{"username":"user@example.com"}`),
							ghttp.RespondWith(http.StatusOK, nil),
						),
					)
				})

				It("makes two requests to CC", func() {
					err := client.SetSpaceRoleByUsername("user@example.com", "space-guid", "org-guid", models.RoleSpaceManager)
					Expect(err).NotTo(HaveOccurred())
					Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
				})
			})

			Context("when given the SpaceDeveloper role", func() {
				BeforeEach(func() {
					ccServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/users"),
							ghttp.VerifyJSON(`{"username":"user@example.com"}`),
							ghttp.RespondWith(http.StatusOK, nil),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/v2/spaces/space-guid/developers"),
							ghttp.VerifyJSON(`{"username":"user@example.com"}`),
							ghttp.RespondWith(http.StatusOK, nil),
						),
					)
				})

				It("makes two requests to CC", func() {
					err := client.SetSpaceRoleByUsername("user@example.com", "space-guid", "org-guid", models.RoleSpaceDeveloper)
					Expect(err).NotTo(HaveOccurred())
					Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
				})
			})

			Context("when given the SpaceAuditor role", func() {
				BeforeEach(func() {
					ccServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/users"),
							ghttp.VerifyJSON(`{"username":"user@example.com"}`),
							ghttp.RespondWith(http.StatusOK, nil),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/v2/spaces/space-guid/auditors"),
							ghttp.VerifyJSON(`{"username":"user@example.com"}`),
							ghttp.RespondWith(http.StatusOK, nil),
						),
					)
				})

				It("makes two requests to CC", func() {
					err := client.SetSpaceRoleByUsername("user@example.com", "space-guid", "org-guid", models.RoleSpaceAuditor)
					Expect(err).NotTo(HaveOccurred())
					Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
				})
			})

			Context("when given an invalid role", func() {
				It("returns an error", func() {
					err := client.SetSpaceRoleByUsername("user@example.com", "space-guid", "org-guid", 666)
					Expect(err).To(HaveOccurred())
					Expect(ccServer.ReceivedRequests()).To(BeZero())
				})
			})

			Context("when assigning the user to the org fails with non '10003' error", func() {
				BeforeEach(func() {
					ccServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/users"),
							ghttp.VerifyJSON(`{"username":"user@example.com"}`),
							ghttp.RespondWith(http.StatusInternalServerError, nil),
						),
					)
				})

				It("returns an error", func() {
					err := client.SetSpaceRoleByUsername("user@example.com", "space-guid", "org-guid", models.RoleSpaceAuditor)
					Expect(err).To(HaveOccurred())
					Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
					Expect(err.Error()).To(ContainSubstring("status code: 500"))
				})
			})

			Context("when assigning the given role to the user fails", func() {
				BeforeEach(func() {
					ccServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/users"),
							ghttp.VerifyJSON(`{"username":"user@example.com"}`),
							ghttp.RespondWith(http.StatusOK, nil),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/v2/spaces/space-guid/auditors"),
							ghttp.VerifyJSON(`{"username":"user@example.com"}`),
							ghttp.RespondWith(http.StatusInternalServerError, nil),
						),
					)
				})

				It("returns an error", func() {
					err := client.SetSpaceRoleByUsername("user@example.com", "space-guid", "org-guid", models.RoleSpaceAuditor)
					Expect(err).To(HaveOccurred())
					Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
					Expect(err.Error()).To(ContainSubstring("status code: 500"))
				})
			})
		})
	})

	Describe("SetSpaceRoleByGUID", func() {
		Context("when given the SpaceManager role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/users/user-guid"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/spaces/space-guid/managers/user-guid"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
			})

			It("makes two requests to CC", func() {
				err := client.SetSpaceRoleByGUID("user-guid", "space-guid", "org-guid", models.RoleSpaceManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
			})
		})

		Context("when given the SpaceDeveloper role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/users/user-guid"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/spaces/space-guid/developers/user-guid"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
			})

			It("makes two requests to CC", func() {
				err := client.SetSpaceRoleByGUID("user-guid", "space-guid", "org-guid", models.RoleSpaceDeveloper)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
			})
		})

		Context("when given the SpaceAuditor role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/users/user-guid"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/spaces/space-guid/auditors/user-guid"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
			})

			It("makes two requests to CC", func() {
				err := client.SetSpaceRoleByGUID("user-guid", "space-guid", "org-guid", models.RoleSpaceAuditor)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
			})
		})

		Context("when given an invalid role", func() {
			It("returns an error", func() {
				err := client.SetSpaceRoleByGUID("user-guid", "space-guid", "org-guid", 666)
				Expect(err).To(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(BeZero())
			})
		})

		Context("when assigning the user to the org fails", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/users/user-guid"),
						ghttp.RespondWith(http.StatusInternalServerError, nil),
					),
				)
			})

			It("returns an error", func() {
				err := client.SetSpaceRoleByGUID("user-guid", "space-guid", "org-guid", models.RoleSpaceAuditor)
				Expect(err).To(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
				Expect(err.Error()).To(ContainSubstring("status code: 500"))
			})
		})

		Context("when assigning the given role to the user fails", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/organizations/org-guid/users/user-guid"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v2/spaces/space-guid/auditors/user-guid"),
						ghttp.RespondWith(http.StatusInternalServerError, nil),
					),
				)
			})

			It("returns an error", func() {
				err := client.SetSpaceRoleByGUID("user-guid", "space-guid", "org-guid", models.RoleSpaceAuditor)
				Expect(err).To(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
				Expect(err.Error()).To(ContainSubstring("status code: 500"))
			})
		})
	})

	Describe("UnsetSpaceRoleByUsername", func() {
		Context("when given the SpaceManager role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/v2/spaces/space-guid/managers"),
						ghttp.VerifyJSON(`{"username":"user@example.com"}`),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
			})

			It("makes one request to CC", func() {
				err := client.UnsetSpaceRoleByUsername("user@example.com", "space-guid", models.RoleSpaceManager)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})
		})

		Context("when given the SpaceDeveloper role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/v2/spaces/space-guid/developers"),
						ghttp.VerifyJSON(`{"username":"user@example.com"}`),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
			})

			It("makes one request to CC", func() {
				err := client.UnsetSpaceRoleByUsername("user@example.com", "space-guid", models.RoleSpaceDeveloper)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})
		})

		Context("when given the SpaceAuditor role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/v2/spaces/space-guid/auditors"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
			})

			It("makes one requests to CC", func() {
				err := client.UnsetSpaceRoleByUsername("user@example.com", "space-guid", models.RoleSpaceAuditor)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})
		})
	})

	Describe("ListUsersInOrgForRole", func() {
		Context("when there are users in the given org with the given role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/organizations/org-guid/users"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
							"resources":[
							{"metadata": {"guid": "user-1-guid"}, "entity": {}}
							]}`),
					),
				)

				uaaServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/Users", fmt.Sprintf("attributes=id,userName&filter=%s", url.QueryEscape(`ID eq "user-1-guid"`))),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
								"resources": [
								{ "id": "user-1-guid", "userName": "Super user 1" }
								]
							}`),
					),
				)
			})

			It("makes a call to CC", func() {
				_, err := client.ListUsersInOrgForRole("org-guid", models.RoleOrgUser)
				Expect(err).ToNot(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("makes a call to UAA", func() {
				_, err := client.ListUsersInOrgForRole("org-guid", models.RoleOrgUser)
				Expect(err).ToNot(HaveOccurred())
				Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns all users in the given org for the given role", func() {
				users, err := client.ListUsersInOrgForRole("org-guid", models.RoleOrgUser)
				Expect(err).ToNot(HaveOccurred())

				Expect(len(users)).To(Equal(1))
				Expect(users[0].GUID).To(Equal("user-1-guid"))
				Expect(users[0].Username).To(Equal("Super user 1"))
			})
		})

		Context("when there are multiple pages of users in the given org with the given role", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/organizations/org-guid/users"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
								"next_url": "/v2/organizations/org-guid/users?page=2",
								"resources":[
								{"metadata": {"guid": "user-1-guid"}, "entity": {}}
								]}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/organizations/org-guid/users", "page=2"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
									"resources":[
									{"metadata": {"guid": "user-2-guid"}, "entity": {"username":"user 2 from cc"}},
									{"metadata": {"guid": "user-3-guid"}, "entity": {"username":"user 3 from cc"}}
									]}`),
					),
				)

				uaaServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/Users", fmt.Sprintf("attributes=id,userName&filter=%s", url.QueryEscape(`ID eq "user-1-guid" or ID eq "user-2-guid" or ID eq "user-3-guid"`))),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
										"resources": [
										{ "id": "user-1-guid", "userName": "Super user 1" },
										{ "id": "user-2-guid", "userName": "Super user 2" },
										{ "id": "user-3-guid", "userName": "Super user 3" }
										]
									}`),
					),
				)
			})

			It("makes a request to CC for each page of results", func() {
				_, err := client.ListUsersInOrgForRole("org-guid", models.RoleOrgUser)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
			})

			It("makes a request to UAA", func() {
				_, err := client.ListUsersInOrgForRole("org-guid", models.RoleOrgUser)
				Expect(err).NotTo(HaveOccurred())
				Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns all paginated users", func() {
				users, err := client.ListUsersInOrgForRole("org-guid", models.RoleOrgUser)
				Expect(err).NotTo(HaveOccurred())

				Expect(len(users)).To(Equal(3))
				Expect(users[0].GUID).To(Equal("user-1-guid"))
				Expect(users[0].Username).To(Equal("Super user 1"))
				Expect(users[1].GUID).To(Equal("user-2-guid"))
				Expect(users[1].Username).To(Equal("Super user 2"))
				Expect(users[2].GUID).To(Equal("user-3-guid"))
				Expect(users[2].Username).To(Equal("Super user 3"))
			})
		})
	})
})
