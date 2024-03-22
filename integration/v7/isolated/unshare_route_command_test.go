package isolated

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("unshare route command", func() {
	Context("Help", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")

			fmt.Println(session)

			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("unshare-route", "ROUTES", "Unshare an existing route from a space"))
		})

		It("displays the help information", func() {
			session := helpers.CF("unshare-route", "--help")
			Eventually(session).Should(Say(`NAME:`))
			Eventually(session).Should(Say("unshare-route - Unshare an existing route from a space"))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`USAGE:`))
			Eventually(session).Should(Say(`Unshare an existing route from a space:`))
			Eventually(session).Should(Say(`cf unshare-route DOMAIN \[--hostname HOSTNAME\] \[--path PATH\] -s OTHER_SPACE \[-o OTHER_ORG\]`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`EXAMPLES:`))
			Eventually(session).Should(Say(`cf unshare-route example.com --hostname myHost --path foo -s TargetSpace -o TargetOrg        # myhost.example.com/foo`))
			Eventually(session).Should(Say(`cf unshare-route example.com --hostname myHost -s TargetSpace                                # myhost.example.com`))
			Eventually(session).Should(Say(`cf unshare-route example.com --hostname myHost -s TargetSpace -o TargetOrg                   # myhost.example.com`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`OPTIONS:`))
			Eventually(session).Should(Say(`--hostname, -n\s+Hostname for the HTTP route \(required for shared domains\)`))
			Eventually(session).Should(Say(`--path\s+Path for the HTTP route`))
			Eventually(session).Should(Say(`-o\s+The org of the destination space \(Default: targeted org\)`))
			Eventually(session).Should(Say(`-s\s+The space to be unshared \(Default: targeted space\)`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`SEE ALSO:`))
			Eventually(session).Should(Say(`share-route, delete-route, map-route, routes, unmap-route`))

			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, false, ReadOnlyOrg, "unshare-route", "some-domain", "-s SOME_SPACE")
		})
	})

	When("the environment is set up correctly", func() {
		var (
			userName  string
			orgName   string
			spaceName string
		)

		BeforeEach(func() {
			helpers.SkipIfVersionLessThan(ccversion.MinVersionHTTP2RoutingV3)
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()

			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the domain extists", func() {
			var (
				domainName      string
				targetSpaceName string
			)

			BeforeEach(func() {
				domainName = helpers.NewDomainName()
			})

			When("the route exists", func() {
				var (
					domain   helpers.Domain
					hostname string
				)
				When("the target space is shared in the targedted space", func() {
					BeforeEach(func() {
						domain = helpers.NewDomain(orgName, domainName)
						hostname = "panera-bread"
						targetSpaceName = helpers.NewSpaceName()
						helpers.CreateSpace(targetSpaceName)
						domain.Create()
						Eventually(helpers.CF("create-route", domain.Name, "--hostname", hostname)).Should(Exit(0))
						Eventually(helpers.CF("share-route", domain.Name, "--hostname", hostname, "-s", targetSpaceName)).Should(Exit(0))
					})

					AfterEach(func() {
						domain.Delete()
					})

					It("unshares the route to the destination space", func() {
						session := helpers.CF("unshare-route", domainName, "--hostname", hostname, "-s", targetSpaceName)
						Eventually(session).Should(Say(`Unsharing route %s.%s from space %s as %s`, hostname, domainName, targetSpaceName, userName))
						Eventually(session).Should(Say(`OK`))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the target organization does not exist", func() {
					var targetOrgName string
					BeforeEach(func() {
						domain = helpers.NewDomain(orgName, domainName)
						hostname = "panera-bread"
						targetSpaceName = helpers.NewSpaceName()
						targetOrgName = helpers.NewOrgName()
						domain.Create()
						Eventually(helpers.CF("create-route", domain.Name, "--hostname", hostname)).Should(Exit(0))
					})

					It("exists with 1 and an error message", func() {
						session := helpers.CF("unshare-route", domainName, "--hostname", hostname, "-o", targetOrgName, "-s", targetSpaceName)
						Eventually(session).Should(Say("Can not unshare route:"))
						Eventually(session).Should(Say(`FAILED`))
						Eventually(session).Should(Exit(1))
					})
				})

				When("the target space exists in another existing org", func() {
					var targetOrgName string
					BeforeEach(func() {
						domain = helpers.NewDomain(orgName, domainName)
						hostname = "menchies-icecream"
						targetOrgName = helpers.NewOrgName()
						targetSpaceName = helpers.NewSpaceName()
						helpers.CreateOrgAndSpace(targetOrgName, targetSpaceName)
						helpers.SetupCF(orgName, spaceName)
						domain.Create()
						Eventually(helpers.CF("create-route", domain.Name, "--hostname", hostname)).Should(Exit(0))
					})

					AfterEach(func() {
						domain.Delete()
					})

					It("unshared the route from the intended space", func() {
						session := helpers.CF("unshare-route", domainName, "--hostname", hostname, "-o", targetOrgName, "-s", targetSpaceName)
						Eventually(session).Should(Say(`Unsharing route %s.%s from space %s as %s`, hostname, domainName, targetSpaceName, userName))
						Eventually(session).Should(Say(`OK`))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the space does not exist", func() {
					var destinationSpaceName string
					BeforeEach(func() {
						domain = helpers.NewDomain(orgName, domainName)
						hostname = "menchies-icecream"
						destinationSpaceName = "doesNotExistSpace"
						domain.Create()
						Eventually(helpers.CF("create-route", domain.Name, "--hostname", hostname)).Should(Exit(0))
					})

					It("exists with 1 with an error", func() {
						session := helpers.CF("unshare-route", domainName, "--hostname", hostname, "-s", destinationSpaceName)
						Eventually(session).Should(Say("Can not unshare route:"))
						Eventually(session).Should(Say(`FAILED`))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			When("the route does not exist", func() {
				var (
					domain   helpers.Domain
					hostname string
				)

				When("the target space exists", func() {
					BeforeEach(func() {
						domain = helpers.NewDomain(orgName, domainName)
						hostname = "panera-bread"
						targetSpaceName = helpers.NewSpaceName()
						helpers.CreateSpace(targetSpaceName)
						domain.Create()
					})

					It("exits with 1 with an error message", func() {
						session := helpers.CF("unshare-route", domainName, "--hostname", hostname, "-s", targetSpaceName)
						Eventually(session).Should(Say("Can not unshare route:"))
						Eventually(session).Should(Say(`FAILED`))
						Eventually(session).Should(Exit(1))
					})
				})
			})
		})
	})
})
