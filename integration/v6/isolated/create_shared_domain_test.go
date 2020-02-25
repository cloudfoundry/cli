package isolated

import (
	"fmt"
	"regexp"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("create-shared-domain command", func() {
	Context("Help", func() {
		It("displays the help information", func() {
			session := helpers.CF("create-shared-domain", "--help")
			Eventually(session).Should(Say("NAME:\n"))
			Eventually(session).Should(Say(regexp.QuoteMeta("create-shared-domain - Create a domain that can be used by all orgs (admin-only)")))
			Eventually(session).Should(Say("USAGE:\n"))
			Eventually(session).Should(Say(regexp.QuoteMeta("cf create-shared-domain DOMAIN [--router-group ROUTER_GROUP | --internal]")))
			Eventually(session).Should(Say("OPTIONS:\n"))
			Eventually(session).Should(Say(`--router-group\s+Routes for this domain will be configured only on the specified router group`))
			Eventually(session).Should(Say(`--internal\s+Applications that use internal routes communicate directly on the container network`))
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

	When("user is logged in as a privileged user", func() {

		var username string

		BeforeEach(func() {
			username, _ = helpers.GetCredentials()
		})

		When("No optional flags are specified", func() {
			When("domain name is valid", func() {
				It("should create the shared domain", func() {
					session := helpers.CF("create-shared-domain", domainName)

					Eventually(session).Should(Say("Creating shared domain %s as %s...", domainName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("domains")
					Eventually(session).Should(Say(`%s\s+shared`, domainName))
					Eventually(session).Should(Exit(0))
				})
			})

			When("domain name is invalid", func() {
				BeforeEach(func() {
					domainName = "invalid-domain-name%*$$#)*" + helpers.RandomName()
				})

				It("should fail and return an error", func() {
					session := helpers.CF("create-shared-domain", domainName)

					Eventually(session).Should(Say("Creating shared domain %s as %s...", regexp.QuoteMeta(domainName), username))
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
					Eventually(session).Should(Say("Creating shared domain %s as %s...", domainName, username))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("The domain name is taken: %s", domainName))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the --internal flag is specified", func() {
			When("the CC API version is less than the minimum version specified", func() {
				var server *Server

				BeforeEach(func() {
					server = helpers.StartAndTargetMockServerWithAPIVersions(ccversion.MinSupportedV2ClientVersion, ccversion.MinSupportedV3ClientVersion)
				})

				AfterEach(func() {
					server.Close()
				})

				XIt("fails with error message that the minimum version is not met", func() {
					session := helpers.CF("create-shared-domain", domainName, "--internal", "-v")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say(`Option '--internal' requires CF API version 2\.115\.0 or higher\. Your target is %s`, ccversion.MinSupportedV2ClientVersion))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the CC API version meets the minimum version requirement", func() {
				When("things work as expected", func() {
					It("creates a domain with internal flag", func() {
						session := helpers.CF("create-shared-domain", domainName, "--internal")

						Eventually(session).Should(Say("Creating shared domain %s as %s...", domainName, username))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))

						session = helpers.CF("domains")

						var sharedDomainResponse struct {
							Resources []struct {
								Entity struct {
									Internal bool   `json:"internal"`
									Name     string `json:"name"`
								}
							}
						}

						helpers.Curl(&sharedDomainResponse, "/v2/shared_domains?q=name:%s", domainName)
						Expect(sharedDomainResponse.Resources).To(HaveLen(1))
						isInternal := sharedDomainResponse.Resources[0].Entity.Internal
						Expect(isInternal).To(BeTrue())
					})
				})

				When("both --internal and --router-group flags are specified", func() {
					It("returns an argument error", func() {
						session := helpers.CF("create-shared-domain", domainName, "--router-group", "my-router-group", "--internal")
						Eventually(session.Err).Should(Say("Incorrect Usage: The following arguments cannot be used together: --router-group, --internal"))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session).Should(Exit(1))
					})
				})
			})
		})

		When("With the --router-group flag", func() {
			var routerGroupName string

			BeforeEach(func() {
				helpers.SkipIfNoRoutingAPI()
			})

			When("router-group exists", func() {
				BeforeEach(func() {
					routerGroupName = helpers.FindOrCreateTCPRouterGroup(GinkgoParallelNode())
				})

				It("should create a new shared domain", func() {
					session := helpers.CF("create-shared-domain", domainName, "--router-group", routerGroupName)

					Eventually(session).Should(Say("Creating shared domain %s as %s...", domainName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("domains")
					Eventually(session).Should(Say(`%s\s+shared`, domainName))

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

	When("user is not logged in as a privileged user", func() {
		var (
			username        string
			password        string
			routerGroupName string
		)

		BeforeEach(func() {
			helpers.LoginCF()
			username, password = helpers.CreateUser()
			helpers.LogoutCF()
			helpers.LoginAs(username, password)
		})

		It("should not be able to create shared domain", func() {
			session := helpers.CF("create-shared-domain", domainName)
			Eventually(session).Should(Say(fmt.Sprintf("Creating shared domain %s as %s...", domainName, username)))
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say("You are not authorized to perform the requested action"))
			Eventually(session).Should(Exit(1))
		})

		When("with --internal flag", func() {
			It("should fail and return an unauthorized message", func() {
				session := helpers.CF("create-shared-domain", domainName, "--internal")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("You are not authorized to perform the requested action"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("with --router-group flag", func() {
			BeforeEach(func() {
				helpers.SkipIfNoRoutingAPI()
			})

			When("router-group exists", func() {
				BeforeEach(func() {
					routerGroupName = helpers.FindOrCreateTCPRouterGroup(GinkgoParallelNode())
				})

				It("should fail and return an unauthorized message", func() {
					session := helpers.CF("create-shared-domain", domainName, "--router-group", routerGroupName)
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).ShouldNot(Say("Error Code: 401"))
					Eventually(session.Err).Should(Say("You are not authorized to perform the requested action"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("router-group does not exists", func() {
				BeforeEach(func() {
					routerGroupName = "invalid-router-group"
				})

				It("should fail and return an unauthorized message", func() {
					session := helpers.CF("create-shared-domain", domainName, "--router-group", routerGroupName)
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).ShouldNot(Say("Error Code: 401"))
					Eventually(session.Err).Should(Say("You are not authorized to perform the requested action"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

	})
})
