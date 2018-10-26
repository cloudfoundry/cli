package isolated

import (
	"fmt"
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-shared-domain command", func() {
	Context("Help", func() {
		It("displays the help information", func() {
			session := helpers.CF("create-shared-domain", "--help")
			Eventually(session).Should(Say("NAME:\n"))
			Eventually(session).Should(Say(regexp.QuoteMeta("create-shared-domain - Create a domain that can be used by all orgs (admin-only)")))
			Eventually(session).Should(Say("USAGE:\n"))
			Eventually(session).Should(Say(regexp.QuoteMeta("cf create-shared-domain DOMAIN [--router-group ROUTER_GROUP]")))
			Eventually(session).Should(Say("OPTIONS:\n"))
			Eventually(session).Should(Say("--router-group\\s+Routes for this domain will be configured only on the specified router group"))
			Eventually(session).Should(Say("SEE ALSO:\n"))
			Eventually(session).Should(Say("create-domain, domains, router-groups"))
			Eventually(session).Should(Exit(0))
		})
	})

	var (
		orgName    string
		spaceName  string
		domainName string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		helpers.SetupCF(orgName, spaceName)
		domainName = helpers.NewDomainName()
	})

	When("user is logged in as admin", func() {
		When("No optional flags are specified", func() {
			When("domain name is valid", func() {
				It("should create the shared domain", func() {
					session := helpers.CF("create-shared-domain", domainName)

					Eventually(session).Should(Say("Creating shared domain %s as admin...", domainName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("domains")
					Eventually(session).Should(Say("%s\\s+shared", domainName))
				})
			})

			When("domain name is invalid", func() {
				BeforeEach(func() {
					domainName = "invalid-domain-name%*$$#)*" + helpers.RandomName()
				})

				It("should fail and return an error", func() {
					session := helpers.CF("create-shared-domain", domainName)

					Eventually(session).Should(Say("Creating shared domain %s as admin...", regexp.QuoteMeta(domainName)))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say(regexp.QuoteMeta("The domain is invalid: name can contain multiple subdomains, each having only alphanumeric characters and hyphens of up to 63 characters, see RFC 1035.")))
					Eventually(session).Should(Exit(1))
				})
			})

			When("domain name is already taken", func() {
				BeforeEach(func() {
					session := helpers.CF("create-shared-domain", domainName)
					Eventually(session).Should(Exit(0))
				})

				It("should fail and return an error", func() {
					session := helpers.CF("create-shared-domain", domainName)
					Eventually(session).Should(Say("Creating shared domain %s as admin...", domainName))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("The domain name is taken: %s", domainName))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("With the --router-group flag", func() {
			var routerGroupName string

			BeforeEach(func() {
				var response struct {
					RoutingEndpoint string `json:"routing_endpoint"`
				}
				helpers.Curl(&response, "/v2/info")

				// TODO: #161159794 remove this skip and check a nicer error message when available
				if response.RoutingEndpoint == "" {
					Skip("Test requires routing endpoint on /v2/info")
				}
			})

			When("router-group exists", func() {
				BeforeEach(func() {
					routerGroupName = helpers.FindOrCreateTCPRouterGroup(GinkgoParallelNode())
				})

				It("should create a new shared domain", func() {
					session := helpers.CF("create-shared-domain", domainName, "--router-group", routerGroupName)

					Eventually(session).Should(Say("Creating shared domain %s as admin...", domainName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("domains")
					Eventually(session).Should(Say("%s\\s+shared", domainName))

					var sharedDomainResponse struct {
						Resources []struct {
							Entity struct {
								RouterGroupGUID string `json:"router_group_guid"`
							}
						}
					}

					helpers.Curl(&sharedDomainResponse, "/v2/shared_domains?q=name:%s", domainName)
					Expect(sharedDomainResponse.Resources).To(HaveLen(1))
					currentRouterGroupGUID := sharedDomainResponse.Resources[0].Entity.RouterGroupGUID

					var routerGroupListResponse []struct{ GUID string }

					helpers.Curl(&routerGroupListResponse, "/routing/v1/router_groups?name=%s", routerGroupName)
					Expect(routerGroupListResponse).To(HaveLen(1))
					expectedRouterGroupGUID := routerGroupListResponse[0].GUID
					Expect(currentRouterGroupGUID).Should(Equal(expectedRouterGroupGUID))
				})
			})

			When("router-group does not exist", func() {
				BeforeEach(func() {
					routerGroupName = "not-a-real-router-group"
					session := helpers.CF("router-groups")
					Consistently(session).ShouldNot(Say(routerGroupName))
					Eventually(session).Should(Exit(0))
				})

				It("should fail and return an error", func() {
					session := helpers.CF("create-shared-domain", domainName, "--router-group", routerGroupName)
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Router group not-a-real-router-group not found"))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})

	When("user is not logged in as admin", func() {
		var (
			username string
			password string
		)

		BeforeEach(func() {
			helpers.LoginCF()
			username, password = helpers.CreateUser()
			helpers.LoginAs(username, password)
		})

		It("should not be able to create shared domain", func() {
			session := helpers.CF("create-shared-domain", domainName)
			Eventually(session).Should(Say(fmt.Sprintf("Creating shared domain %s as %s...", domainName, username)))
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say("You are not authorized to perform the requested action"))
			Eventually(session).Should(Exit(1))
		})
	})
})
