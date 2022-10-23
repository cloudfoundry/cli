package isolated

import (
	"regexp"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-orphaned-routes command", func() {
	Context("Help", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("delete-orphaned-routes", "ROUTES", "Delete all orphaned routes in the currently targeted space (i.e. those that are not mapped to an app or service instance)"))
		})

		It("displays the help information", func() {
			session := helpers.CF("delete-orphaned-routes", "--help")
			Eventually(session).Should(Say(`NAME:`))
			Eventually(session).Should(Say(`delete-orphaned-routes - Delete all orphaned routes in the currently targeted space \(i\.e\. those that are not mapped to an app or service instance\)`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`USAGE:`))
			Eventually(session).Should(Say(`cf delete-orphaned-routes \[-f\]`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`OPTIONS:`))
			Eventually(session).Should(Say(`-f\s+Force deletion without confirmation`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`SEE ALSO:`))
			Eventually(session).Should(Say(`delete-routes, routes`))

			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "delete-orphaned-routes")
		})
	})

	When("the environment is set up correctly", func() {
		var (
			buffer     *Buffer
			orgName    string
			spaceName  string
			domainName string
			appName    string
			hostName   string
			userName   string
		)

		BeforeEach(func() {
			buffer = NewBuffer()
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			appName = helpers.NewAppName()
			hostName = helpers.NewHostName()
			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
			domainName = helpers.DefaultSharedDomain()

			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName, "--no-start")).Should(Exit(0))
			})

			Eventually(helpers.CF("create-route", domainName, "--hostname", hostName)).Should(Exit(0))
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the -f flag is not given", func() {
			When("the user enters 'y'", func() {
				BeforeEach(func() {
					_, err := buffer.Write([]byte("y\n"))
					Expect(err).ToNot(HaveOccurred())
				})
				It("it asks for confirmation and deletes the domain", func() {
					session := helpers.CFWithStdin(buffer, "delete-orphaned-routes")
					Eventually(session).Should(Say(`Really delete orphaned routes?`))
					Eventually(session).Should(Say(regexp.QuoteMeta(`Deleting orphaned routes as %s...`), userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					Expect(string(session.Out.Contents())).NotTo(ContainSubstring("Unable to delete"))
					session = helpers.CF("routes")
					Consistently(session).ShouldNot(Say(`%s\s+%s`, hostName, domainName))
					Eventually(session).Should(Say(`%s\s+%s`, appName, domainName))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the user enters 'n'", func() {
				BeforeEach(func() {
					_, err := buffer.Write([]byte("n\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("it asks for confirmation and does not delete the domain", func() {
					session := helpers.CFWithStdin(buffer, "delete-orphaned-routes")
					Eventually(session).Should(Say(`Really delete orphaned routes?`))
					Eventually(session).Should(Say(`Routes have not been deleted`))
					Consistently(session).ShouldNot(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the user's input is invalid", func() {
				BeforeEach(func() {
					_, err := buffer.Write([]byte("abc\n"))
					Expect(err).ToNot(HaveOccurred())
					_, err = buffer.Write([]byte("n\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("it asks for confirmation and does not delete the domain", func() {
					session := helpers.CFWithStdin(buffer, "delete-orphaned-routes")
					Eventually(session).Should(Say(`Really delete orphaned routes?`))
					Eventually(session).Should(Say(`invalid input \(not y, n, yes, or no\)`))
					Eventually(session).Should(Say(`Really delete orphaned routes?`))
					Eventually(session).Should(Say(`Routes have not been deleted`))
					Consistently(session).ShouldNot(Say("OK"))
					Eventually(session).Should(Exit(0))
				})

			})
		})

		When("the -f flag is given", func() {
			It("deletes the orphaned routes", func() {
				session := helpers.CF("delete-orphaned-routes", "-f")
				Eventually(session).Should(Say(`Deleting orphaned routes as %s\.\.\.`, userName))
				Eventually(session).Should(Say(`OK`))
				Eventually(session).Should(Exit(0))

				Expect(string(session.Out.Contents())).NotTo(ContainSubstring("Unable to delete"))

				session = helpers.CF("routes")
				Consistently(session).ShouldNot(Say(`%s\s+%s`, hostName, domainName))
				Eventually(session).Should(Say(`%s\s+%s`, appName, domainName))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
