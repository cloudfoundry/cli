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

	Describe("FindByUsername", func() {
		Context("when a username has multiple origins", func() {
			BeforeEach(func() {
				uaaServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/Users"),
						ghttp.RespondWith(http.StatusOK, `{
							"resources": [
								{ "id": "user-1-guid", "userName": "some-user" },
								{ "id": "user-2-guid", "userName": "some-user" }
							]}`),
					),
				)
			})

			It("returns an error", func() {
				_, err := client.FindByUsername("some-user")
				Expect(err).To(MatchError("The user exists in multiple origins."))
			})
		})
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
})
