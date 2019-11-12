package isolated

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("domains command", func() {
	Describe("help", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("domains", "DOMAINS", "List domains in the target org"))
		})

		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("domains", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say(`\s+domains - List domains in the target org`))

				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`\s+cf domains`))

				Eventually(session).Should(Say("EXAMPLES:"))
				Eventually(session).Should(Say(`\s+cf domains`))
				Eventually(session).Should(Say(`\s+cf domains --labels 'environment in \(production,staging\),tier in \(backend\)'`))
				Eventually(session).Should(Say(`\s+cf domains --labels 'env=dev,!chargeback-code,tier in \(backend,worker\)'`))

				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`--labels\s+Selector to filter domains by labels`))

				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say(`\s+create-private-domain, create-route, create-shared-domain, routes, set-label`))
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
			Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' or 'cf login --sso' to log in."))
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
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say(`--labels\s+Selector to filter domains by labels`))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say(`\s+create-private-domain, create-route, create-shared-domain, routes`))
			Eventually(session).Should(Exit(1))
		})
	})

	When("logged in", func() {
		var userName string

		BeforeEach(func() {
			userName, _ = helpers.GetCredentials()
		})

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
				spaceName     string
				sharedDomain1 helpers.Domain
				sharedDomain2 helpers.Domain
			)

			BeforeEach(func() {
				orgName = helpers.NewOrgName()
				spaceName = helpers.NewSpaceName()
				helpers.SetupCF(orgName, spaceName)

				sharedDomain1 = helpers.NewDomain(orgName, helpers.NewDomainName("a"))
				sharedDomain1.CreateShared()

				sharedDomain2 = helpers.NewDomain(orgName, helpers.NewDomainName("b"))
				sharedDomain2.CreateShared()
				Eventually(helpers.CF("set-label", "domain", sharedDomain1.Name, "keyfor1=valuefor1")).Should(Exit(0))
			})
			AfterEach(func() {
				Eventually(helpers.CF("delete-shared-domain", sharedDomain1.Name, "-f")).Should(Exit(0))
				Eventually(helpers.CF("delete-shared-domain", sharedDomain2.Name, "-f")).Should(Exit(0))
				Eventually(helpers.CF("delete-space", spaceName, "-f")).Should(Exit(0))
				Eventually(helpers.CF("delete-org", orgName, "-f")).Should(Exit(0))
			})

			When("the targeted org has shared domains", func() {

				It("displays the shared domains and denotes that they are shared", func() {
					session := helpers.CF("domains")
					Eventually(session).Should(Exit(0))

					Expect(session).Should(Say(`Getting domains in org %s as %s\.\.\.`, orgName, userName))
					Expect(session).Should(Say(`name\s+availability\s+internal`))
					Expect(session).Should(Say(`%s\s+shared\s+`, sharedDomain1.Name))
					Expect(session).Should(Say(`%s\s+shared\s+`, sharedDomain2.Name))
				})

				It("displays the shared domains and denotes that they are shared for matching labels only", func() {
					session := helpers.CF("domains", "--labels", "keyfor1=valuefor1")
					Eventually(session).Should(Exit(0))

					Expect(session).Should(Say(`Getting domains in org %s as %s\.\.\.`, orgName, userName))
					Expect(session).Should(Say(`name\s+availability\s+internal`))
					Expect(session).Should(Say(`%s\s+shared\s+`, sharedDomain1.Name))
					Expect(session).ShouldNot(Say(`%s\s+shared\s+`, sharedDomain2.Name))
				})

				When("the shared domain is internal", func() {
					var internalDomainName string

					BeforeEach(func() {
						helpers.SkipIfVersionLessThan(ccversion.MinVersionInternalDomainV2)
						internalDomainName = helpers.NewDomainName()
						internalDomain := helpers.NewDomain(orgName, internalDomainName)
						internalDomain.CreateInternal()
					})
					AfterEach(func() {
						Eventually(helpers.CF("delete-shared-domain", internalDomainName, "-f")).Should(Exit(0))
					})

					It("displays the internal flag on the shared domain", func() {
						session := helpers.CF("domains")
						Eventually(session).Should(Exit(0))

						Expect(session).Should(Say(`Getting domains in org %s as %s`, orgName, userName))
						Expect(session).Should(Say(`name\s+availability\s+internal`))
						Expect(session).Should(Say(`%s\s+shared\s+true`, internalDomainName))
					})

					It("displays the shared domains and denotes that they are shared for matching labels only", func() {
						session := helpers.CF("domains", "--labels", "keyfor1=valuefor1")
						Eventually(session).Should(Exit(0))

						Expect(session).Should(Say(`Getting domains in org %s as %s\.\.\.`, orgName, userName))
						Expect(session).Should(Say(`name\s+availability\s+internal`))
						Expect(session).Should(Say(`%s\s+shared\s+`, sharedDomain1.Name))
						Expect(session).ShouldNot(Say(`%s\s+shared\s+`, sharedDomain2.Name))
						Expect(session).ShouldNot(Say(internalDomainName))

					})
				})
			})

			When("the targeted org has a private domain", func() {
				var privateDomain1, privateDomain2 helpers.Domain

				BeforeEach(func() {
					privateDomain1 = helpers.NewDomain(orgName, helpers.NewDomainName("a"))
					privateDomain1.Create()

					privateDomain2 = helpers.NewDomain(orgName, helpers.NewDomainName("b"))
					privateDomain2.Create()

					Eventually(helpers.CF("set-label", "domain", privateDomain2.Name, "keyfor2=valuefor2")).Should(Exit(0))
				})
				AfterEach(func() {
					Eventually(helpers.CF("delete-private-domain", privateDomain1.Name, "-f")).Should(Exit(0))
					Eventually(helpers.CF("delete-private-domain", privateDomain2.Name, "-f")).Should(Exit(0))
				})

				It("displays the private domains", func() {
					session := helpers.CF("domains")

					Eventually(session).Should(Exit(0))
					Expect(session).Should(Say(`Getting domains in org %s as %s`, orgName, userName))
					Expect(session).Should(Say(`name\s+availability\s+internal`))
					Expect(session).Should(Say(`%s\s+private\s+`, privateDomain1.Name))
					Expect(session).Should(Say(`%s\s+private\s+`, privateDomain2.Name))
				})

				It("filters private domains by label", func() {
					session := helpers.CF("domains", "--labels", "keyfor2=valuefor2")

					Eventually(session).Should(Exit(0))
					Expect(session).Should(Say(`Getting domains in org %s as %s`, orgName, userName))
					Expect(session).Should(Say(`name\s+availability\s+internal`))
					Expect(session).ShouldNot(Say(`%s\s+private\s+`, privateDomain1.Name))
					Expect(session).Should(Say(`%s\s+private\s+`, privateDomain2.Name))
				})

				When("targeting a different org", func() {
					var (
						newOrgName     string
						newSpaceName   string
						privateDomain3 helpers.Domain
					)

					BeforeEach(func() {
						newOrgName = helpers.NewOrgName()
						newSpaceName = helpers.NewSpaceName()
						helpers.SetupCF(newOrgName, newSpaceName)

						privateDomain3 = helpers.NewDomain(newOrgName, helpers.NewDomainName("c"))
						privateDomain3.Create()
					})
					AfterEach(func() {
						Eventually(helpers.CF("delete-private-domain", privateDomain3.Name, "-f")).Should(Exit(0))
						Eventually(helpers.CF("delete-space", newSpaceName, "-f")).Should(Exit(0))
						Eventually(helpers.CF("delete-org", newOrgName, "-f")).Should(Exit(0))
						// Outer after-eaches require the initial org/space to be targetted
						helpers.TargetOrgAndSpace(orgName, spaceName)
					})

					It("should not display the private domains of other orgs", func() {
						session := helpers.CF("domains")

						Eventually(session).Should(Say(`Getting domains in org %s as %s`, newOrgName, userName))
						Eventually(session).Should(Say(`name\s+availability\s+internal`))

						Consistently(session).ShouldNot(Say(`%s`, privateDomain1.Name))
						Consistently(session).ShouldNot(Say(`%s`, privateDomain2.Name))

						Eventually(session).Should(Say(`%s\s+shared\s+`, sharedDomain1.Name))
						Eventually(session).Should(Say(`%s\s+shared\s+`, sharedDomain2.Name))
						Eventually(session).Should(Say(`%s\s+private\s+`, privateDomain3.Name))
						Eventually(session).Should(Exit(0))
					})
				})

				When("logged in as a user that cannot see private domains", func() {
					var userName string

					BeforeEach(func() {
						userName = helpers.SwitchToOrgRole(orgName, "BillingManager")
						helpers.TargetOrg(orgName)
					})
					JustAfterEach(func() {
						// waiting AfterEach blocks are designed to run in as admin
						helpers.LoginCF()
						helpers.TargetOrg(orgName)
					})

					It("only prints the shared domains", func() {
						session := helpers.CF("domains")

						Eventually(session).Should(Say(`Getting domains in org %s as %s`, orgName, userName))
						Eventually(session).Should(Say(`name\s+availability\s+internal`))
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
