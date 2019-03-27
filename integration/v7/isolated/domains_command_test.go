package isolated

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("domains command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("domains", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say(`\s+domains - List domains in the target org`))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`\s+cf domains`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say(`\s+create-route, router-groups, routes`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("user is not logged in", func() {
		BeforeEach(func() {
			helpers.LogoutCF()
		})

		It("displays an error message and fails", func() {
			session := helpers.CF("domains")
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' to log in."))
			Eventually(session).Should(Exit(1))
		})
	})

	When("a random argument is passed to the command", func() {
		It("displays an error message and fails", func() {
			session := helpers.CF("domains", "random-arg")
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say(`Incorrect Usage: unexpected argument "random-arg"`))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say(`\s+domains - List domains in the target org`))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say(`\s+cf domains`))
			Eventually(session).Should(Exit(1))
		})
	})

	When("logged in as admin", func() {
		When("no org is targeted", func() {
			BeforeEach(func() {
				helpers.LoginCF()
			})

			It("displays an error message and fails", func() {
				session := helpers.CF("domains")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`No org targeted, use 'cf target -o ORG' to target an org.`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("an org is targeted", func() {
			var (
				orgName       string
				sharedDomain1 helpers.Domain
				sharedDomain2 helpers.Domain
			)

			BeforeEach(func() {
				orgName = helpers.NewOrgName()
				spaceName := helpers.NewSpaceName()
				helpers.SetupCF(orgName, spaceName)

				sharedDomain1 = helpers.NewDomain(orgName, helpers.NewDomainName())
				sharedDomain1.CreateShared()

				sharedDomain2 = helpers.NewDomain(orgName, helpers.NewDomainName())
				sharedDomain2.CreateShared()
			})

			When("the targeted org has shared domains", func() {
				It("displays the shared domains and denotes that they are shared", func() {
					session := helpers.CF("domains")

					Eventually(session).Should(Say(`Getting domains in org %s as admin\.\.\.`, orgName))
					Eventually(session).Should(Say(`name\s+status\s+type\s+details`))
					Eventually(session).Should(Say(`%s\s+shared\s+`, sharedDomain1.Name))
					Eventually(session).Should(Say(`%s\s+shared\s+`, sharedDomain2.Name))
					Eventually(session).Should(Exit(0))
				})

				When("the shared domain is internal", func() {
					var internalDomainName string

					BeforeEach(func() {
						helpers.SkipIfVersionLessThan(ccversion.MinVersionInternalDomainV2)
						internalDomainName = helpers.NewDomainName()
						internalDomain := helpers.NewDomain(orgName, internalDomainName)
						internalDomain.CreateInternal()
					})

					It("displays the internal flag on the shared domain", func() {
						session := helpers.CF("domains")

						Eventually(session).Should(Say(`Getting domains in org %s as admin`, orgName))
						Eventually(session).Should(Say(`name\s+status\s+type\s+details`))
						Eventually(session).Should(Say(`%s\s+shared\s+internal`, internalDomainName))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			When("the targeted org has a private domain", func() {
				var privateDomain1, privateDomain2 helpers.Domain

				BeforeEach(func() {
					privateDomain1 = helpers.NewDomain(orgName, helpers.NewDomainName())
					privateDomain1.Create()

					privateDomain2 = helpers.NewDomain(orgName, helpers.NewDomainName())
					privateDomain2.Create()
				})

				It("displays the private domains", func() {
					session := helpers.CF("domains")

					Eventually(session).Should(Say(`Getting domains in org %s as admin`, orgName))
					Eventually(session).Should(Say(`name\s+status\s+type\s+details`))
					Eventually(session).Should(Say(`%s\s+owned\s+`, privateDomain1.Name))
					Eventually(session).Should(Say(`%s\s+owned\s+`, privateDomain2.Name))
					Eventually(session).Should(Exit(0))
				})

				When("targeting a different org", func() {
					var (
						newOrgName     string
						privateDomain3 helpers.Domain
					)

					BeforeEach(func() {
						newOrgName = helpers.NewOrgName()
						newSpaceName := helpers.NewSpaceName()
						helpers.SetupCF(newOrgName, newSpaceName)

						privateDomain3 = helpers.NewDomain(newOrgName, helpers.NewDomainName())
						privateDomain3.Create()
					})

					It("should not display the private domains of other orgs", func() {
						session := helpers.CF("domains")

						Eventually(session).Should(Say(`Getting domains in org %s as admin`, newOrgName))
						Eventually(session).Should(Say(`name\s+status\s+type\s+details`))

						Consistently(session).ShouldNot(Say(`%s`, privateDomain1.Name))
						Consistently(session).ShouldNot(Say(`%s`, privateDomain2.Name))

						Eventually(session).Should(Say(`%s\s+shared\s+`, sharedDomain1.Name))
						Eventually(session).Should(Say(`%s\s+shared\s+`, sharedDomain2.Name))
						Eventually(session).Should(Say(`%s\s+owned\s+`, privateDomain3.Name))
						Eventually(session).Should(Exit(0))
					})
				})

				When("logged in as a user that cannot see private domains", func() {
					var username string

					BeforeEach(func() {
						username = helpers.SwitchToOrgRole(orgName, "BillingManager")
						helpers.TargetOrg(orgName)
					})

					It("only prints the shared domains", func() {
						session := helpers.CF("domains")

						Eventually(session).Should(Say(`Getting domains in org %s as %s`, orgName, username))
						Eventually(session).Should(Say(`name\s+status\s+type\s+details`))
						Eventually(session).Should(Say(`%s\s+shared\s+`, sharedDomain1.Name))
						Eventually(session).Should(Say(`%s\s+shared\s+`, sharedDomain2.Name))

						Consistently(session).ShouldNot(Say(privateDomain1.Name))
						Consistently(session).ShouldNot(Say(privateDomain2.Name))

						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})
